package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynTool_basic(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-tool")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynToolConfig(resourceName, "Terraform acceptance tool", "Acceptance prompt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_tool.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_tool.test", "description", "Terraform acceptance tool"),
					resource.TestCheckResourceAttr("agyn_tool.test", "type", "send_message"),
					resource.TestCheckResourceAttrSet("agyn_tool.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_tool.test", "config"),
				),
			},
		},
	})
}

func TestAccAgynTool_update(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-tool")
	updatedName := resourceName + "-updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynToolConfig(resourceName, "Terraform acceptance tool", "Acceptance prompt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_tool.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_tool.test", "description", "Terraform acceptance tool"),
					resource.TestCheckResourceAttrSet("agyn_tool.test", "id"),
				),
			},
			{
				Config: testAccAgynToolConfig(updatedName, "Terraform acceptance tool updated", "Updated prompt"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_tool.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_tool.test", "description", "Terraform acceptance tool updated"),
					resource.TestCheckResourceAttrSet("agyn_tool.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynTool_import(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-tool")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynToolConfig(resourceName, "Terraform acceptance tool", "Acceptance prompt"),
			},
			{
				ResourceName:      "agyn_tool.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgynToolConfig(name, description, prompt string) string {
	return fmt.Sprintf(`
%s

resource "agyn_tool" "test" {
  name        = %q
  description = %q
  type        = "send_message"
  config = jsonencode({
    prompt = %q
  })
}
`, testAccProviderConfig(), name, description, prompt)
}
