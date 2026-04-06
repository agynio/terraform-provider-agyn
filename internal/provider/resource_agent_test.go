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
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(organizationName, resourceName, "Terraform acceptance agent", "Terraform acceptance role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "organization_id"),
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
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(organizationName, resourceName, "Terraform acceptance agent", "Terraform acceptance role"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "model"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "id"),
				),
			},
			{
				Config: testAccAgynAgentConfig(organizationName, updatedName, "Terraform acceptance agent updated", "Terraform acceptance role updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent updated"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role updated"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "organization_id"),
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
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(organizationName, resourceName, "Terraform acceptance agent", "Terraform acceptance role"),
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
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAgynAgentInvalidConfig(organizationName),
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

func testAccAgynAgentConfig(organizationName, title, description, role string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_agent" "test" {
	  organization_id = agyn_organization.test.id
	  name        = %q
	  description = %q
	  role        = %q
	  model       = %q
	  image       = %q
}
`, testAccProviderConfig(), organizationName, title, description, role, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"))
}

func testAccAgynAgentInvalidConfig(organizationName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_agent" "test" {
	  organization_id = agyn_organization.test.id
	  name          = "invalid"
	  role          = "invalid"
	  model         = %q
	  image         = %q
	  configuration = "{invalid"
}
`, testAccProviderConfig(), organizationName, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"))
}
