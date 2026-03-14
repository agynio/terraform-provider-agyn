package provider

import (
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
}

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"agyn": providerserver.NewProtocol6WithError(New("test", "test")()),
}

func TestBuildTeamAPIURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "no trailing slash",
			baseURL: "https://gateway.example.com",
			want:    "https://gateway.example.com" + teamAPIPath + "/",
		},
		{
			name:    "trailing slash",
			baseURL: "https://gateway.example.com/",
			want:    "https://gateway.example.com" + teamAPIPath + "/",
		},
		{
			name:    "already includes path",
			baseURL: "https://gateway.example.com" + teamAPIPath,
			want:    "https://gateway.example.com" + teamAPIPath + teamAPIPath + "/",
		},
		{
			name:    "path with trailing slash",
			baseURL: "https://gateway.example.com" + teamAPIPath + "/",
			want:    "https://gateway.example.com" + teamAPIPath + teamAPIPath + "/",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := buildTeamAPIURL(test.baseURL); got != test.want {
				t.Fatalf("buildTeamAPIURL(%q) = %q, want %q", test.baseURL, got, test.want)
			}
		})
	}
}
