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
	APIToken   string
	HTTPClient *http.Client
}

type Client struct {
	gateway              gatewayv1connect.AgentsGatewayClient
	appsGateway          gatewayv1connect.AppsGatewayClient
	runnersGateway       gatewayv1connect.RunnersGatewayClient
	secretsGateway       gatewayv1connect.SecretsGatewayClient
	organizationsGateway gatewayv1connect.OrganizationsGatewayClient
	llmGateway           gatewayv1connect.LLMGatewayClient
	usersGateway         gatewayv1connect.UsersGatewayClient
	egressGateway        gatewayv1connect.EgressRulesGatewayClient
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
	} else {
		clone := *httpClient
		httpClient = &clone
	}

	if cfg.APIToken != "" {
		baseTransport := httpClient.Transport
		if baseTransport == nil {
			baseTransport = http.DefaultTransport
		}
		httpClient.Transport = &authTransport{base: baseTransport, token: cfg.APIToken}
	}

	return &Client{
		gateway:              gatewayv1connect.NewAgentsGatewayClient(httpClient, cfg.BaseURL),
		appsGateway:          gatewayv1connect.NewAppsGatewayClient(httpClient, cfg.BaseURL),
		runnersGateway:       gatewayv1connect.NewRunnersGatewayClient(httpClient, cfg.BaseURL),
		secretsGateway:       gatewayv1connect.NewSecretsGatewayClient(httpClient, cfg.BaseURL),
		organizationsGateway: gatewayv1connect.NewOrganizationsGatewayClient(httpClient, cfg.BaseURL),
		llmGateway:           gatewayv1connect.NewLLMGatewayClient(httpClient, cfg.BaseURL),
		usersGateway:         gatewayv1connect.NewUsersGatewayClient(httpClient, cfg.BaseURL),
		egressGateway:        gatewayv1connect.NewEgressRulesGatewayClient(httpClient, cfg.BaseURL),
	}, nil
}

type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(req)
}
