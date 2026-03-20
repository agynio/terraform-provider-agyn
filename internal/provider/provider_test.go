package provider

import (
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func testAccPreCheck(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance tests skipped unless TF_ACC=1")
	}
	if os.Getenv("AGYN_BASE_URL") == "" {
		t.Skip("AGYN_BASE_URL must be set for acceptance tests")
	}
	if os.Getenv("AGYN_MODEL_ID") == "" {
		t.Skip("AGYN_MODEL_ID must be set for acceptance tests")
	}
	if os.Getenv("AGYN_AGENT_IMAGE") == "" {
		t.Skip("AGYN_AGENT_IMAGE must be set for acceptance tests")
	}
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"agyn": providerserver.NewProtocol6WithError(New("test", "test")()),
}

func TestBuildAgentAPIURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "no trailing slash",
			baseURL: "https://gateway.example.com",
			want:    "https://gateway.example.com" + agentAPIPath + "/",
		},
		{
			name:    "trailing slash",
			baseURL: "https://gateway.example.com/",
			want:    "https://gateway.example.com" + agentAPIPath + "/",
		},
		{
			name:    "already includes path",
			baseURL: "https://gateway.example.com" + agentAPIPath,
			want:    "https://gateway.example.com" + agentAPIPath + agentAPIPath + "/",
		},
		{
			name:    "path with trailing slash",
			baseURL: "https://gateway.example.com" + agentAPIPath + "/",
			want:    "https://gateway.example.com" + agentAPIPath + agentAPIPath + "/",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := buildAgentAPIURL(test.baseURL); got != test.want {
				t.Fatalf("buildAgentAPIURL(%q) = %q, want %q", test.baseURL, got, test.want)
			}
		})
	}
}

func TestBuildAgentAPIURLResolvesOperationPaths(t *testing.T) {
	t.Parallel()

	baseURL := buildAgentAPIURL("https://gateway.example.com")

	serverURL, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse base URL: %v", err)
	}

	tests := []struct {
		name          string
		operationPath string
		want          string
	}{
		{
			name:          "agents",
			operationPath: "/agents",
			want:          "https://gateway.example.com" + agentAPIPath + "/agents",
		},
		{
			name:          "volumes",
			operationPath: "/volumes",
			want:          "https://gateway.example.com" + agentAPIPath + "/volumes",
		},
		{
			name:          "envs",
			operationPath: "/envs",
			want:          "https://gateway.example.com" + agentAPIPath + "/envs",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			operationPath := test.operationPath
			if operationPath[0] == '/' {
				operationPath = "." + operationPath
			}

			queryURL, err := serverURL.Parse(operationPath)
			if err != nil {
				t.Fatalf("parse operation path %q: %v", operationPath, err)
			}

			if got := queryURL.String(); got != test.want {
				t.Fatalf("resolved %q = %q, want %q", operationPath, got, test.want)
			}
		})
	}
}
