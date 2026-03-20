package agentapi

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/agynio/terraform-provider-agyn/gen/agynio/api/gateway/v1/gatewayv1connect"
)

const defaultTimeout = 30 * time.Second

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Client struct {
	gateway gatewayv1connect.AgentsGatewayClient
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	return &Client{gateway: gatewayv1connect.NewAgentsGatewayClient(httpClient, cfg.BaseURL)}, nil
}
