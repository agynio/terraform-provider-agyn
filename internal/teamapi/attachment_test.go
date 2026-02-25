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

func TestGetAttachment(t *testing.T) {
	attachmentID := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	createdAt := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET request, got %s", r.Method)
		}
		expectedPath := "/attachments/" + attachmentID.String()
		if r.URL.Path != expectedPath {
			t.Fatalf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		payload := map[string]any{
			"id":         attachmentID.String(),
			"kind":       "agent_tool",
			"sourceId":   sourceID.String(),
			"sourceType": "agent",
			"targetId":   targetID.String(),
			"targetType": "tool",
			"createdAt":  createdAt.Format(time.RFC3339Nano),
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

	result, err := client.GetAttachment(context.Background(), attachmentID.String())
	if err != nil {
		t.Fatalf("GetAttachment returned error: %v", err)
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

func TestGetAttachmentNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.GetAttachment(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound, got %v", err)
	}
}
