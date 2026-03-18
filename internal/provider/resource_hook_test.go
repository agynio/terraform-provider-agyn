package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynHook_basic(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynHookConfig(agentName, "Terraform acceptance hook", "agent.started", "handler"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_hook.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_hook.test", "event", "agent.started"),
					resource.TestCheckResourceAttr("agyn_hook.test", "function", "handler"),
					resource.TestCheckResourceAttr("agyn_hook.test", "description", "Terraform acceptance hook"),
					resource.TestCheckResourceAttrSet("agyn_hook.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_hook.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynHook_update(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynHookConfig(agentName, "Terraform acceptance hook", "agent.started", "handler"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_hook.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_hook.test", "event", "agent.started"),
					resource.TestCheckResourceAttr("agyn_hook.test", "function", "handler"),
					resource.TestCheckResourceAttr("agyn_hook.test", "description", "Terraform acceptance hook"),
					resource.TestCheckResourceAttrSet("agyn_hook.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_hook.test", "id"),
				),
			},
			{
				Config: testAccAgynHookConfig(agentName, "Terraform acceptance hook updated", "agent.started", "handler"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_hook.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_hook.test", "event", "agent.started"),
					resource.TestCheckResourceAttr("agyn_hook.test", "function", "handler"),
					resource.TestCheckResourceAttr("agyn_hook.test", "description", "Terraform acceptance hook updated"),
					resource.TestCheckResourceAttrSet("agyn_hook.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_hook.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynHook_import(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynHookConfig(agentName, "Terraform acceptance hook", "agent.started", "handler"),
			},
			{
				ResourceName:      "agyn_hook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgynHookConfig(agentName, description, event, function string) string {
	return fmt.Sprintf(`
%s

%s

resource "agyn_hook" "test" {
	  agent_id    = agyn_agent.test.id
	  event       = %q
	  function    = %q
	  image       = %q
	  description = %q
}
`, testAccProviderConfig(), testAccAgynAgentResourceBlock(agentName, "Terraform acceptance agent", "Terraform acceptance role"), event, function, os.Getenv("AGYN_AGENT_IMAGE"), description)
}
