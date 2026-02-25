package teamapi

import (
	"errors"
	"net/http"
	"testing"
)

func TestAPIErrorFormatting(t *testing.T) {
	wrapped := errorFromResponse("create agent", http.StatusBadRequest, []byte(`{"title":"Invalid","detail":"bad config"}`))

	var apiErr *APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatalf("expected to unwrap APIError, got %T", wrapped)
	}

	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("unexpected status: %d", apiErr.Status)
	}
	if apiErr.Title != "Invalid" {
		t.Errorf("unexpected title: %s", apiErr.Title)
	}
	if apiErr.Detail != "bad config" {
		t.Errorf("unexpected detail: %s", apiErr.Detail)
	}

	expected := "Invalid (400): bad config"
	if apiErr.Error() != expected {
		t.Errorf("unexpected error string: %s", apiErr.Error())
	}
}

func TestErrorFromResponseFallsBackToStatusText(t *testing.T) {
	wrapped := errorFromResponse("delete agent", http.StatusNotFound, nil)

	var apiErr *APIError
	if !errors.As(wrapped, &apiErr) {
		t.Fatalf("expected APIError from wrapped error")
	}

	if apiErr.Title != http.StatusText(http.StatusNotFound) {
		t.Errorf("expected title %q, got %q", http.StatusText(http.StatusNotFound), apiErr.Title)
	}
	if apiErr.Detail != http.StatusText(http.StatusNotFound) {
		t.Errorf("expected detail %q, got %q", http.StatusText(http.StatusNotFound), apiErr.Detail)
	}
}
