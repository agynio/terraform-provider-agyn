package teamapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFindAttachmentByID(t *testing.T) {
	attachmentID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	createdAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/attachments" {
			t.Fatalf("expected path /attachments, got %s", r.URL.Path)
		}
		query := r.URL.Query()
		if query.Get("kind") != "agent_tool" {
			t.Fatalf("expected kind agent_tool, got %s", query.Get("kind"))
		}
		if query.Get("sourceId") != sourceID.String() {
			t.Fatalf("expected sourceId %s, got %s", sourceID, query.Get("sourceId"))
		}
		if query.Get("targetId") != targetID.String() {
			t.Fatalf("expected targetId %s, got %s", targetID, query.Get("targetId"))
		}

		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"items": []map[string]any{
				{
					"id":         attachmentID.String(),
					"kind":       "agent_tool",
					"sourceId":   sourceID.String(),
					"sourceType": "agent",
					"targetId":   targetID.String(),
					"targetType": "tool",
					"createdAt":  createdAt.Format(time.RFC3339Nano),
				},
			},
			"page":    1,
			"perPage": 25,
			"total":   1,
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	result, err := client.FindAttachmentByID(context.Background(), attachmentID.String(), "agent_tool", sourceID.String(), targetID.String())
	if err != nil {
		t.Fatalf("FindAttachmentByID returned error: %v", err)
	}

	if result.ID != attachmentID.String() {
		t.Fatalf("expected ID %s, got %s", attachmentID, result.ID)
	}
	if result.Kind != "agent_tool" {
		t.Fatalf("expected kind agent_tool, got %s", result.Kind)
	}
	if result.SourceID != sourceID.String() {
		t.Fatalf("expected source ID %s, got %s", sourceID, result.SourceID)
	}
	if result.TargetID != targetID.String() {
		t.Fatalf("expected target ID %s, got %s", targetID, result.TargetID)
	}
	if result.SourceType != "agent" {
		t.Fatalf("expected source type agent, got %s", result.SourceType)
	}
	if result.TargetType != "tool" {
		t.Fatalf("expected target type tool, got %s", result.TargetType)
	}
	if !result.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created at %s, got %s", createdAt, result.CreatedAt)
	}
}

func TestFindAttachmentByIDNotFound(t *testing.T) {
	attachmentID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/attachments" {
			t.Fatalf("expected path /attachments, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"items":   []map[string]any{},
			"page":    1,
			"perPage": 25,
			"total":   0,
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.FindAttachmentByID(context.Background(), attachmentID.String(), "agent_tool", sourceID.String(), targetID.String())
	if !errors.Is(err, ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound, got %v", err)
	}
}

func TestFindAttachmentByIDNoMatchingID(t *testing.T) {
	attachmentID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	otherID := uuid.New()
	createdAt := time.Date(2024, 6, 2, 12, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/attachments" {
			t.Fatalf("expected path /attachments, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"items": []map[string]any{
				{
					"id":         otherID.String(),
					"kind":       "agent_tool",
					"sourceId":   sourceID.String(),
					"sourceType": "agent",
					"targetId":   targetID.String(),
					"targetType": "tool",
					"createdAt":  createdAt.Format(time.RFC3339Nano),
				},
			},
			"page":    1,
			"perPage": 25,
			"total":   1,
		}
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.FindAttachmentByID(context.Background(), attachmentID.String(), "agent_tool", sourceID.String(), targetID.String())
	if !errors.Is(err, ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound, got %v", err)
	}
}
