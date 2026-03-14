package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynMCPServer_basic(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-mcp")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynMCPServerConfig(resourceName, "Terraform acceptance MCP server"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_mcp_server.test", "title", resourceName),
					resource.TestCheckResourceAttr("agyn_mcp_server.test", "description", "Terraform acceptance MCP server"),
					resource.TestCheckResourceAttr("agyn_mcp_server.test", "namespace", resourceName),
					resource.TestCheckResourceAttr("agyn_mcp_server.test", "command", "mcp start --stdio"),
					resource.TestCheckResourceAttrSet("agyn_mcp_server.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynMCPServer_deprecatedConfig(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-mcp")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynMCPServerDeprecatedConfig(resourceName, "Terraform acceptance MCP server"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_mcp_server.test", "title", resourceName),
					resource.TestCheckResourceAttr("agyn_mcp_server.test", "description", "Terraform acceptance MCP server"),
					resource.TestCheckResourceAttrSet("agyn_mcp_server.test", "config"),
					resource.TestCheckResourceAttrSet("agyn_mcp_server.test", "id"),
				),
			},
		},
	})
}

func testAccAgynMCPServerConfig(title, description string) string {
	return fmt.Sprintf(`
%s

resource "agyn_mcp_server" "test" {
  title       = %q
  description = %q
  namespace = %q
  command   = "mcp start --stdio"
}
`, testAccProviderConfig(), title, description, title)
}

func testAccAgynMCPServerDeprecatedConfig(title, description string) string {
	return fmt.Sprintf(`
%s

resource "agyn_mcp_server" "test" {
  title       = %q
  description = %q
  config = jsonencode({
    namespace = %q
    command   = "mcp start --stdio"
  })
}
`, testAccProviderConfig(), title, description, title)
}
