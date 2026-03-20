package agentapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agynio/terraform-provider-agyn/internal/agentclient"
)

func TestNewClientRequiresBaseURL(t *testing.T) {
	if _, err := NewClient(Config{}); err == nil {
		t.Fatalf("expected error when base URL is empty")
	}

	if _, err := NewClient(Config{BaseURL: "http://%zz"}); err == nil {
		t.Fatalf("expected error for invalid base URL")
	}
}

func TestNewClientUsesProvidedHTTPClient(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(server.Close)

	customClient := &http.Client{Timeout: 42 * time.Second}

	client, err := NewClient(Config{BaseURL: server.URL, HTTPClient: customClient})
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}

	underlying, ok := client.raw.ClientInterface.(*agentclient.Client)
	if !ok {
		t.Fatalf("expected underlying client type %T", &agentclient.Client{})
	}

	if underlying.Client != customClient {
		t.Fatalf("expected custom HTTP client to be preserved")
	}
}
