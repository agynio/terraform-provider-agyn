package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynWorkspaceConfiguration_basic(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-workspace")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynWorkspaceConfigurationConfig(resourceName, "Terraform acceptance workspace config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_workspace_configuration.test", "title", resourceName),
					resource.TestCheckResourceAttr("agyn_workspace_configuration.test", "description", "Terraform acceptance workspace config"),
					resource.TestCheckResourceAttrSet("agyn_workspace_configuration.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_workspace_configuration.test", "config"),
				),
			},
		},
	})
}

func testAccAgynWorkspaceConfigurationConfig(title, description string) string {
	return fmt.Sprintf(`
%s

resource "agyn_workspace_configuration" "test" {
  title       = %q
  description = %q
  config = jsonencode({
    platform     = "auto"
    ttlSeconds   = 600
    enableDinD   = false
  })
}
`, testAccProviderConfig(), title, description)
}
