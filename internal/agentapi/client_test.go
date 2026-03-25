package agentapi

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	if client.gateway == nil {
		t.Fatalf("expected gateway client to be initialized")
	}
	if client.appsGateway == nil {
		t.Fatalf("expected apps gateway client to be initialized")
	}
}

func TestAuthTransportAddsAuthorizationHeader(t *testing.T) {
	const token = "test-token"
	var gotHeader string
	transport := &authTransport{
		base: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			gotHeader = req.Header.Get("Authorization")
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
		token: token,
	}

	client := &http.Client{Transport: transport}
	request, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}

	if _, err := client.Do(request); err != nil {
		t.Fatalf("unexpected error performing request: %v", err)
	}

	if gotHeader != "Bearer "+token {
		t.Fatalf("expected Authorization header to be set, got %q", gotHeader)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
