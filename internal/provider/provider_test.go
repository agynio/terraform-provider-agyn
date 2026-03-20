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
