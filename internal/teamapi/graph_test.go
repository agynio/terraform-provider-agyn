package teamapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/agynio/terraform-provider-agyn/internal/teamclient"
)

func TestGraphRequestURLStripsBasePath(t *testing.T) {
	baseClient, err := teamclient.NewClient("https://example.com/team/v1", teamclient.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatalf("failed to build base client: %v", err)
	}

	requestURL, err := graphRequestURL(baseClient, "/api/graph")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestURL.Path != "/api/graph" {
		t.Fatalf("expected path /api/graph, got %s", requestURL.Path)
	}
	if requestURL.RawQuery != "" {
		t.Fatalf("expected empty query string, got %q", requestURL.RawQuery)
	}
}

func TestFindGraphEdge(t *testing.T) {
	edge := GraphEdge{
		ID:           "edge-1",
		Source:       "source-1",
		SourceHandle: "source-handle",
		Target:       "target-1",
		TargetHandle: "target-handle",
	}
	graph := Graph{
		Name:    "main",
		Version: 1,
		Nodes: []GraphNode{
			{ID: edge.Source},
			{ID: edge.Target},
		},
		Edges: []GraphEdge{edge},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/graph" {
			t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(graph); err != nil {
			t.Fatalf("failed to encode graph response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.FindGraphEdge(context.Background(), edge.ID)
	if err != nil {
		t.Fatalf("FindGraphEdge returned error: %v", err)
	}
	if result.ID != edge.ID {
		t.Fatalf("expected edge id %s, got %s", edge.ID, result.ID)
	}
	if result.Source != edge.Source || result.Target != edge.Target {
		t.Fatalf("unexpected edge endpoints: %+v", result)
	}

	_, err = client.FindGraphEdge(context.Background(), "missing")
	if !errors.Is(err, ErrGraphEdgeNotFound) {
		t.Fatalf("expected ErrGraphEdgeNotFound, got %v", err)
	}
}

func TestUpsertGraphEdgeRetriesOnConflict(t *testing.T) {
	edge := GraphEdge{
		ID:           "edge-1",
		Source:       "source-1",
		SourceHandle: "source-handle",
		Target:       "target-1",
		TargetHandle: "target-handle",
	}
	graph := Graph{
		Name:    "main",
		Version: 1,
		Nodes: []GraphNode{
			{ID: edge.Source},
			{ID: edge.Target},
		},
	}

	var mu sync.Mutex
	getCount := 0
	postCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			getCount++
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graph); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		case http.MethodPost:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			postCount++
			w.Header().Set("Content-Type", "application/json")
			if postCount == 1 {
				w.WriteHeader(http.StatusConflict)
				if err := json.NewEncoder(w).Encode(map[string]any{
					"title":  "Conflict",
					"detail": "VERSION_CONFLICT",
					"status": http.StatusConflict,
				}); err != nil {
					t.Fatalf("failed to encode conflict response: %v", err)
				}
				return
			}

			graphResponse := graph
			graphResponse.Edges = []GraphEdge{edge}
			if err := json.NewEncoder(w).Encode(graphResponse); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.UpsertGraphEdge(context.Background(), edge); err != nil {
		t.Fatalf("UpsertGraphEdge returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if postCount != 2 {
		t.Fatalf("expected 2 POST attempts, got %d", postCount)
	}
	if getCount != 2 {
		t.Fatalf("expected 2 GET attempts, got %d", getCount)
	}
}

func TestUpsertGraphEdgeWithNodesCreatesNodes(t *testing.T) {
	edge := GraphEdge{
		ID:           "edge-1",
		Source:       "source-1",
		SourceHandle: "source-handle",
		Target:       "target-1",
		TargetHandle: "target-handle",
	}
	sourceHint := GraphNodeHint{ID: edge.Source, Template: "agent"}
	targetHint := GraphNodeHint{ID: edge.Target, Template: "tool"}
	graph := Graph{
		Name:    "main",
		Version: 1,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graph); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		case http.MethodPost:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			var payload GraphUpsertRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode graph request: %v", err)
			}
			if len(payload.Nodes) != 2 {
				t.Fatalf("expected 2 nodes, got %d", len(payload.Nodes))
			}
			sourceNode, ok := findGraphNode(payload.Nodes, sourceHint.ID)
			if !ok {
				t.Fatalf("expected source node %s to be included", sourceHint.ID)
			}
			if sourceNode.Template != sourceHint.Template {
				t.Fatalf("expected source node template %s, got %s", sourceHint.Template, sourceNode.Template)
			}
			targetNode, ok := findGraphNode(payload.Nodes, targetHint.ID)
			if !ok {
				t.Fatalf("expected target node %s to be included", targetHint.ID)
			}
			if targetNode.Template != targetHint.Template {
				t.Fatalf("expected target node template %s, got %s", targetHint.Template, targetNode.Template)
			}
			if len(payload.Edges) != 1 {
				t.Fatalf("expected 1 edge, got %d", len(payload.Edges))
			}
			if payload.Edges[0] != edge {
				t.Fatalf("unexpected edge payload: %+v", payload.Edges[0])
			}

			graphResponse := graph
			graphResponse.Nodes = payload.Nodes
			graphResponse.Edges = payload.Edges
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graphResponse); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.UpsertGraphEdgeWithNodes(context.Background(), edge, sourceHint, targetHint); err != nil {
		t.Fatalf("UpsertGraphEdgeWithNodes returned error: %v", err)
	}
}

func TestUpsertGraphEdgeWithNodesSkipsExistingNodes(t *testing.T) {
	edge := GraphEdge{
		ID:           "edge-1",
		Source:       "source-1",
		SourceHandle: "source-handle",
		Target:       "target-1",
		TargetHandle: "target-handle",
	}
	sourceHint := GraphNodeHint{ID: edge.Source, Template: "agent"}
	targetHint := GraphNodeHint{ID: edge.Target, Template: "tool"}
	graph := Graph{
		Name:    "main",
		Version: 1,
		Nodes: []GraphNode{
			{ID: edge.Source, Template: "existing-source"},
			{ID: edge.Target, Template: "existing-target"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graph); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		case http.MethodPost:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			var payload GraphUpsertRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode graph request: %v", err)
			}
			if len(payload.Nodes) != 2 {
				t.Fatalf("expected 2 nodes, got %d", len(payload.Nodes))
			}
			sourceNode, ok := findGraphNode(payload.Nodes, edge.Source)
			if !ok {
				t.Fatalf("expected source node %s to be included", edge.Source)
			}
			if sourceNode.Template != "existing-source" {
				t.Fatalf("expected existing source node template, got %s", sourceNode.Template)
			}
			targetNode, ok := findGraphNode(payload.Nodes, edge.Target)
			if !ok {
				t.Fatalf("expected target node %s to be included", edge.Target)
			}
			if targetNode.Template != "existing-target" {
				t.Fatalf("expected existing target node template, got %s", targetNode.Template)
			}
			if len(payload.Edges) != 1 {
				t.Fatalf("expected 1 edge, got %d", len(payload.Edges))
			}
			if payload.Edges[0] != edge {
				t.Fatalf("unexpected edge payload: %+v", payload.Edges[0])
			}

			graphResponse := graph
			graphResponse.Edges = payload.Edges
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graphResponse); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.UpsertGraphEdgeWithNodes(context.Background(), edge, sourceHint, targetHint); err != nil {
		t.Fatalf("UpsertGraphEdgeWithNodes returned error: %v", err)
	}
}

func TestUpsertGraphEdgeWithNodesPartialNodeCreation(t *testing.T) {
	edge := GraphEdge{
		ID:           "edge-1",
		Source:       "source-1",
		SourceHandle: "source-handle",
		Target:       "target-1",
		TargetHandle: "target-handle",
	}
	sourceHint := GraphNodeHint{ID: edge.Source, Template: "agent"}
	targetHint := GraphNodeHint{ID: edge.Target, Template: "tool"}
	graph := Graph{
		Name:    "main",
		Version: 1,
		Nodes: []GraphNode{
			{ID: edge.Source, Template: "existing-source"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graph); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		case http.MethodPost:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			var payload GraphUpsertRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("failed to decode graph request: %v", err)
			}
			if len(payload.Nodes) != 2 {
				t.Fatalf("expected 2 nodes, got %d", len(payload.Nodes))
			}
			sourceNode, ok := findGraphNode(payload.Nodes, edge.Source)
			if !ok {
				t.Fatalf("expected source node %s to be included", edge.Source)
			}
			if sourceNode.Template != "existing-source" {
				t.Fatalf("expected existing source node template, got %s", sourceNode.Template)
			}
			targetNode, ok := findGraphNode(payload.Nodes, edge.Target)
			if !ok {
				t.Fatalf("expected target node %s to be included", edge.Target)
			}
			if targetNode.Template != targetHint.Template {
				t.Fatalf("expected target node template %s, got %s", targetHint.Template, targetNode.Template)
			}
			if len(payload.Edges) != 1 {
				t.Fatalf("expected 1 edge, got %d", len(payload.Edges))
			}
			if payload.Edges[0] != edge {
				t.Fatalf("unexpected edge payload: %+v", payload.Edges[0])
			}

			graphResponse := graph
			graphResponse.Nodes = payload.Nodes
			graphResponse.Edges = payload.Edges
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graphResponse); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.UpsertGraphEdgeWithNodes(context.Background(), edge, sourceHint, targetHint); err != nil {
		t.Fatalf("UpsertGraphEdgeWithNodes returned error: %v", err)
	}
}

func TestDeleteGraphEdgeRetriesOnConflict(t *testing.T) {
	edge := GraphEdge{
		ID:           "edge-1",
		Source:       "source-1",
		SourceHandle: "source-handle",
		Target:       "target-1",
		TargetHandle: "target-handle",
	}
	graph := Graph{
		Name:    "main",
		Version: 1,
		Nodes: []GraphNode{
			{ID: edge.Source},
			{ID: edge.Target},
		},
		Edges: []GraphEdge{edge},
	}

	var mu sync.Mutex
	getCount := 0
	postCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			getCount++
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(graph); err != nil {
				t.Fatalf("failed to encode graph response: %v", err)
			}
		case http.MethodPost:
			if r.URL.Path != "/api/graph" {
				t.Fatalf("expected path /api/graph, got %s", r.URL.Path)
			}
			postCount++
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"title":  "Conflict",
				"detail": "VERSION_CONFLICT",
				"status": http.StatusConflict,
			}); err != nil {
				t.Fatalf("failed to encode conflict response: %v", err)
			}
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := client.DeleteGraphEdge(context.Background(), edge.ID); err == nil {
		t.Fatalf("expected conflict error, got nil")
	} else if !isVersionConflict(err) {
		t.Fatalf("expected version conflict error, got %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if postCount != conflictRetryCount {
		t.Fatalf("expected %d POST attempts, got %d", conflictRetryCount, postCount)
	}
	if getCount != conflictRetryCount {
		t.Fatalf("expected %d GET attempts, got %d", conflictRetryCount, getCount)
	}
}

func findGraphNode(nodes []GraphNode, id string) (GraphNode, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
	}
	return GraphNode{}, false
}
