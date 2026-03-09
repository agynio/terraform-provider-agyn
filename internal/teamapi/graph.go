package teamapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/agynio/terraform-provider-agyn/internal/teamclient"
)

type Graph struct {
	Name      string          `json:"name"`
	Version   int             `json:"version"`
	UpdatedAt string          `json:"updatedAt"`
	Nodes     []GraphNode     `json:"nodes"`
	Edges     []GraphEdge     `json:"edges"`
	Variables []GraphVariable `json:"variables,omitempty"`
}

type GraphNode struct {
	ID       string             `json:"id"`
	Template string             `json:"template"`
	Config   map[string]any     `json:"config,omitempty"`
	State    map[string]any     `json:"state,omitempty"`
	Position *GraphNodePosition `json:"position,omitempty"`
}

type GraphNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type GraphEdge struct {
	ID           string `json:"id,omitempty"`
	Source       string `json:"source"`
	SourceHandle string `json:"sourceHandle"`
	Target       string `json:"target"`
	TargetHandle string `json:"targetHandle"`
}

type GraphVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GraphUpsertRequest struct {
	Name      string          `json:"name"`
	Version   int             `json:"version"`
	Nodes     []GraphNode     `json:"nodes"`
	Edges     []GraphEdge     `json:"edges"`
	Variables []GraphVariable `json:"variables,omitempty"`
}

var ErrGraphEdgeNotFound = errors.New("graph edge not found")

func (c *Client) GetGraph(ctx context.Context) (*Graph, error) {
	baseClient, ok := c.raw.ClientInterface.(*teamclient.Client)
	if !ok {
		return nil, fmt.Errorf("unsupported client type %T", c.raw.ClientInterface)
	}

	requestURL, err := graphRequestURL(baseClient, "/api/graph")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create graph request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := baseClient.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get graph request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read graph response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse("get graph", resp.StatusCode, body)
	}

	var graph Graph
	if err := json.Unmarshal(body, &graph); err != nil {
		return nil, fmt.Errorf("decode graph response: %w", err)
	}
	return &graph, nil
}

func (c *Client) UpsertGraph(ctx context.Context, input GraphUpsertRequest) (*Graph, error) {
	baseClient, ok := c.raw.ClientInterface.(*teamclient.Client)
	if !ok {
		return nil, fmt.Errorf("unsupported client type %T", c.raw.ClientInterface)
	}

	requestURL, err := graphRequestURL(baseClient, "/api/graph")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal graph payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create graph request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := baseClient.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upsert graph request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read graph response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errorFromResponse("upsert graph", resp.StatusCode, body)
	}

	var graph Graph
	if err := json.Unmarshal(body, &graph); err != nil {
		return nil, fmt.Errorf("decode graph response: %w", err)
	}
	return &graph, nil
}

func (c *Client) FindGraphEdge(ctx context.Context, edgeID string) (*GraphEdge, error) {
	if edgeID == "" {
		return nil, ErrGraphEdgeNotFound
	}

	graph, err := c.GetGraph(ctx)
	if err != nil {
		return nil, err
	}

	for _, edge := range graph.Edges {
		if edge.ID == edgeID {
			return &edge, nil
		}
	}
	return nil, ErrGraphEdgeNotFound
}

func (c *Client) UpsertGraphEdge(ctx context.Context, edge GraphEdge) error {
	if edge.ID == "" {
		return fmt.Errorf("graph edge id is required")
	}

	return withConflictRetryNoResult(ctx, "upsert graph edge", func() error {
		graph, err := c.GetGraph(ctx)
		if err != nil {
			return err
		}

		if !graphHasNode(graph.Nodes, edge.Source) {
			return fmt.Errorf("graph missing source node %s", edge.Source)
		}
		if !graphHasNode(graph.Nodes, edge.Target) {
			return fmt.Errorf("graph missing target node %s", edge.Target)
		}

		if graphEdgeIndex(graph.Edges, edge.ID) != -1 {
			return nil
		}

		graph.Edges = append(graph.Edges, edge)
		_, err = c.UpsertGraph(ctx, graphUpsertRequest(graph))
		return err
	})
}

func (c *Client) DeleteGraphEdge(ctx context.Context, edgeID string) error {
	if edgeID == "" {
		return fmt.Errorf("graph edge id is required")
	}

	return withConflictRetryNoResult(ctx, "delete graph edge", func() error {
		graph, err := c.GetGraph(ctx)
		if err != nil {
			return err
		}

		idx := graphEdgeIndex(graph.Edges, edgeID)
		if idx == -1 {
			return nil
		}
		graph.Edges = append(graph.Edges[:idx], graph.Edges[idx+1:]...)
		_, err = c.UpsertGraph(ctx, graphUpsertRequest(graph))
		return err
	})
}

func graphRequestURL(baseClient *teamclient.Client, path string) (*url.URL, error) {
	serverURL, err := url.Parse(baseClient.Server)
	if err != nil {
		return nil, fmt.Errorf("parse server url: %w", err)
	}
	if path == "" {
		path = "/"
	} else if path[0] != '/' {
		path = "/" + path
	}

	requestURL := &url.URL{
		Scheme: serverURL.Scheme,
		Host:   serverURL.Host,
		User:   serverURL.User,
		Path:   path,
	}
	requestURL.RawQuery = ""
	requestURL.Fragment = ""
	return requestURL, nil
}

func graphUpsertRequest(graph *Graph) GraphUpsertRequest {
	return GraphUpsertRequest{
		Name:      graph.Name,
		Version:   graph.Version,
		Nodes:     graph.Nodes,
		Edges:     graph.Edges,
		Variables: graph.Variables,
	}
}

func graphHasNode(nodes []GraphNode, id string) bool {
	for _, node := range nodes {
		if node.ID == id {
			return true
		}
	}
	return false
}

func graphEdgeIndex(edges []GraphEdge, id string) int {
	for idx, edge := range edges {
		if edge.ID == id {
			return idx
		}
	}
	return -1
}
