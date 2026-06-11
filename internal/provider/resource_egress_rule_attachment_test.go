package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynEgressRuleAttachment_basic(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	ruleName := acctest.RandomWithPrefix("tf-acc-egress-rule")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynEgressRuleAttachmentConfig(organizationName, agentName, ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_egress_rule_attachment.test", "organization_id", "agyn_organization.test", "id"),
					resource.TestCheckResourceAttrPair("agyn_egress_rule_attachment.test", "rule_id", "agyn_egress_rule.test", "id"),
					resource.TestCheckResourceAttrPair("agyn_egress_rule_attachment.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_egress_rule_attachment.test", "id"),
				),
			},
		},
	})
}

func testAccAgynEgressRuleAttachmentConfig(organizationName, agentName, ruleName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_agent" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  role            = "Terraform acceptance role"
	  model           = %q
	  image           = %q
	  init_image      = %q
	  availability    = "private"
}

resource "agyn_egress_rule" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  domain_pattern  = "api.example.com"
	  action          = "allow"
}

resource "agyn_egress_rule_attachment" "test" {
	  organization_id = agyn_organization.test.id
	  rule_id         = agyn_egress_rule.test.id
	  agent_id        = agyn_agent.test.id
}
`, testAccProviderConfig(), organizationName, agentName, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"), os.Getenv("AGYN_AGENT_INIT_IMAGE"), ruleName)
}
