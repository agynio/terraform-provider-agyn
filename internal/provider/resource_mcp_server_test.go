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
					resource.TestCheckResourceAttrSet("agyn_mcp_server.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_mcp_server.test", "config"),
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
  config = jsonencode({
    namespace = %q
    command   = "mcp start --stdio"
  })
}
`, testAccProviderConfig(), title, description, title)
}
