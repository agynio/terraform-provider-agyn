package agentapi

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestParseUUID(t *testing.T) {
	id := uuid.New().String()
	parsed, err := parseUUID(id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if uuidToString(parsed) != id {
		t.Errorf("expected %s, got %s", id, uuidToString(parsed))
	}

	if _, err = parseUUID("not-a-uuid"); err == nil {
		t.Fatalf("expected error for invalid UUID")
	}
}

func TestDecodePayload(t *testing.T) {
	source := map[string]any{"name": "agent", "value": 1}
	var target struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	if err := decodePayload(source, &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if target.Name != "agent" || target.Value != 1 {
		t.Fatalf("unexpected decoded values: %+v", target)
	}

	bad := make(chan struct{})
	if err := decodePayload(bad, &target); err == nil {
		t.Fatalf("expected error for non-serializable input")
	}

	var invalidTarget struct {
		Count int `json:"count"`
	}
	if err := decodePayload(source, &invalidTarget); err != nil {
		t.Fatalf("unexpected error decoding partial payload: %v", err)
	}
	if invalidTarget.Count != 0 {
		t.Fatalf("expected zero value for missing field, got %d", invalidTarget.Count)
	}

	// ensure original source unchanged
	if _, err := json.Marshal(source); err != nil {
		t.Fatalf("source should remain serializable: %v", err)
	}
}
