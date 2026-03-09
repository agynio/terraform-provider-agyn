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

func TestGraphRequestURLResolvesRelativePath(t *testing.T) {
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
