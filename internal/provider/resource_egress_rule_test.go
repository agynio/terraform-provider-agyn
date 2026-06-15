package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEgressRuleSchemaValidation(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "agyn" {
  api_url = "https://gateway.example.com"
}

resource "agyn_egress_rule" "bad" {
  organization_id = "org-id"
  name            = "github"
  domain_pattern  = "*.github.com"
  action          = "allow"

  injected_header {
    name      = "Authorization"
    value     = "literal"
    secret_id = "secret-id"
  }
}
`,
				ExpectError: regexp.MustCompile(`one \(and only one\)`),
			},
		},
	})
}

func TestAccEgressRuleRequiresActionOrHeader(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "agyn" {
  api_url = "https://gateway.example.com"
}

resource "agyn_egress_rule" "noop" {
  organization_id = "org-id"
  name            = "github"
  domain_pattern  = "*.github.com"
}
`,
				ExpectError: regexp.MustCompile(`requires an action or at least one injected header`),
			},
		},
	})
}
