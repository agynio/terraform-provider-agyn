package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynEgressRule_basic(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	ruleName := acctest.RandomWithPrefix("tf-acc-egress-rule")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynEgressRuleConfig(organizationName, ruleName, "allow"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "name", ruleName),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "domain_pattern", "api.example.com"),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "ports.#", "1"),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "ports.0", "443"),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "methods.#", "1"),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "methods.0", "GET"),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "action", "allow"),
					resource.TestCheckResourceAttrSet("agyn_egress_rule.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_egress_rule.test", "id"),
				),
			},
			{
				Config: testAccAgynEgressRuleConfig(organizationName, ruleName+"-updated", "deny"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "name", ruleName+"-updated"),
					resource.TestCheckResourceAttr("agyn_egress_rule.test", "action", "deny"),
					resource.TestCheckResourceAttrSet("agyn_egress_rule.test", "id"),
				),
			},
		},
	})
}

func testAccAgynEgressRuleConfig(organizationName, ruleName, action string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_secret" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  value           = "secret-value"
}

resource "agyn_egress_rule" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  description     = "Terraform acceptance egress rule"
	  domain_pattern  = "api.example.com"
	  ports           = [443]
	  methods         = ["get"]
	  path_pattern    = "/v1/*"
	  action          = %q

	  header {
		name      = "Authorization"
		scheme    = "bearer"
		secret_id = agyn_secret.test.id
	  }
}
`, testAccProviderConfig(), organizationName, ruleName+"-secret", ruleName, action)
}
