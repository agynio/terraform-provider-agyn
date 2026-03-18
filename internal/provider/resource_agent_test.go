package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynAgent_basic(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(resourceName, "Terraform acceptance agent", "Terraform acceptance role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "model"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynAgent_update(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-agent")
	updatedName := resourceName + "-updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(resourceName, "Terraform acceptance agent", "Terraform acceptance role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "model"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "id"),
				),
			},
			{
				Config: testAccAgynAgentConfig(updatedName, "Terraform acceptance agent updated", "Terraform acceptance role updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent updated"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role updated"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "model"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynAgent_import(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(resourceName, "Terraform acceptance agent", "Terraform acceptance role"),
			},
			{
				ResourceName:      "agyn_agent.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAgynAgent_expectError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAgynAgentInvalidConfig(),
				ExpectError: regexp.MustCompile("Invalid JSON"),
			},
		},
	})
}

func testAccProviderConfig() string {
	return fmt.Sprintf(`
provider "agyn" {
  api_url = %q
}
`, os.Getenv("AGYN_BASE_URL"))
}

func testAccAgynAgentConfig(title, description, role string) string {
	return fmt.Sprintf(`
%s

resource "agyn_agent" "test" {
	  name        = %q
	  description = %q
	  role        = %q
	  model       = %q
	  image       = %q
}
`, testAccProviderConfig(), title, description, role, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"))
}

func testAccAgynAgentInvalidConfig() string {
	return fmt.Sprintf(`
%s

resource "agyn_agent" "test" {
	  name          = "invalid"
	  role          = "invalid"
	  model         = %q
	  image         = %q
	  configuration = "{invalid"
}
`, testAccProviderConfig(), os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"))
}
