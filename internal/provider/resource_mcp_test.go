package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynMcp_basic(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynMcpConfig(agentName, "Terraform acceptance MCP", "mcp start"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_mcp.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_mcp.test", "command", "mcp start"),
					resource.TestCheckResourceAttr("agyn_mcp.test", "description", "Terraform acceptance MCP"),
					resource.TestCheckResourceAttrSet("agyn_mcp.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_mcp.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynMcp_update(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynMcpConfig(agentName, "Terraform acceptance MCP", "mcp start"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_mcp.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_mcp.test", "command", "mcp start"),
					resource.TestCheckResourceAttr("agyn_mcp.test", "description", "Terraform acceptance MCP"),
					resource.TestCheckResourceAttrSet("agyn_mcp.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_mcp.test", "id"),
				),
			},
			{
				Config: testAccAgynMcpConfig(agentName, "Terraform acceptance MCP updated", "mcp start --updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_mcp.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttr("agyn_mcp.test", "command", "mcp start --updated"),
					resource.TestCheckResourceAttr("agyn_mcp.test", "description", "Terraform acceptance MCP updated"),
					resource.TestCheckResourceAttrSet("agyn_mcp.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_mcp.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynMcp_import(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynMcpConfig(agentName, "Terraform acceptance MCP", "mcp start"),
			},
			{
				ResourceName:      "agyn_mcp.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgynMcpConfig(agentName, description, command string) string {
	return fmt.Sprintf(`
%s

%s

resource "agyn_mcp" "test" {
	  agent_id    = agyn_agent.test.id
	  image       = %q
	  command     = %q
	  description = %q
}
`, testAccProviderConfig(), testAccAgynAgentResourceBlock(agentName, "Terraform acceptance agent", "Terraform acceptance role"), os.Getenv("AGYN_AGENT_IMAGE"), command, description)
}
