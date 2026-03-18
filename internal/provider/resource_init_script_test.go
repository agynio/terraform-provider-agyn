package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynInitScript_basic(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynInitScriptConfig(agentName, "Terraform acceptance init script", "echo hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_init_script.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_init_script.test", "script", "echo hello"),
					resource.TestCheckResourceAttr("agyn_init_script.test", "description", "Terraform acceptance init script"),
					resource.TestCheckResourceAttrSet("agyn_init_script.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynInitScript_update(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynInitScriptConfig(agentName, "Terraform acceptance init script", "echo hello"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_init_script.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_init_script.test", "script", "echo hello"),
					resource.TestCheckResourceAttr("agyn_init_script.test", "description", "Terraform acceptance init script"),
					resource.TestCheckResourceAttrSet("agyn_init_script.test", "id"),
				),
			},
			{
				Config: testAccAgynInitScriptConfig(agentName, "Terraform acceptance init script updated", "echo updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_init_script.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_init_script.test", "script", "echo updated"),
					resource.TestCheckResourceAttr("agyn_init_script.test", "description", "Terraform acceptance init script updated"),
					resource.TestCheckResourceAttrSet("agyn_init_script.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynInitScript_import(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynInitScriptConfig(agentName, "Terraform acceptance init script", "echo hello"),
			},
			{
				ResourceName:      "agyn_init_script.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgynInitScriptConfig(agentName, description, script string) string {
	return fmt.Sprintf(`
%s

%s

resource "agyn_init_script" "test" {
	  script      = %q
	  description = %q
	  agent_id    = agyn_agent.test.id
}
`, testAccProviderConfig(), testAccAgynAgentResourceBlock(agentName, "Terraform acceptance agent", "Terraform acceptance role"), script, description)
}
