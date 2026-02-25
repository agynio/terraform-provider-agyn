package teamapi

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/agynio/terraform-provider-agyn/internal/teamclient"
)

const defaultTimeout = 30 * time.Second

type Config struct {
	BaseURL    string
	HTTPClient *http.Client
}

type Client struct {
	raw *teamclient.ClientWithResponses
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

	rawClient, err := teamclient.NewClientWithResponses(cfg.BaseURL, teamclient.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create team client: %w", err)
	}

	return &Client{raw: rawClient}, nil
}
