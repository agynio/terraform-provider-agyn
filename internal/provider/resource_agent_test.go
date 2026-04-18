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
				Config: testAccAgynAgentConfig(organizationName, resourceName, "Terraform acceptance agent", "Terraform acceptance role", nil, []string{"docker"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent"),
					resource.TestCheckNoResourceAttr("agyn_agent.test", "nickname"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role"),
					resource.TestCheckResourceAttr("agyn_agent.test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("agyn_agent.test", "capabilities.0", "docker"),
					resource.TestCheckResourceAttr("agyn_agent.test", "init_image", os.Getenv("AGYN_AGENT_INIT_IMAGE")),
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
	nickname := "tf-acc-nickname"
	updatedNickname := "tf-acc-nickname-updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(organizationName, resourceName, "Terraform acceptance agent", "Terraform acceptance role", &nickname, []string{"docker"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", resourceName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent"),
					resource.TestCheckResourceAttr("agyn_agent.test", "nickname", nickname),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role"),
					resource.TestCheckResourceAttr("agyn_agent.test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("agyn_agent.test", "capabilities.0", "docker"),
					resource.TestCheckResourceAttr("agyn_agent.test", "init_image", os.Getenv("AGYN_AGENT_INIT_IMAGE")),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "model"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "id"),
				),
			},
			{
				Config: testAccAgynAgentConfig(organizationName, updatedName, "Terraform acceptance agent updated", "Terraform acceptance role updated", &updatedNickname, []string{"docker", "gpu"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent updated"),
					resource.TestCheckResourceAttr("agyn_agent.test", "nickname", updatedNickname),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role updated"),
					resource.TestCheckResourceAttr("agyn_agent.test", "capabilities.#", "2"),
					resource.TestCheckResourceAttr("agyn_agent.test", "init_image", os.Getenv("AGYN_AGENT_INIT_IMAGE")),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "model"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "image"),
					resource.TestCheckResourceAttrSet("agyn_agent.test", "id"),
				),
			},
			{
				Config: testAccAgynAgentConfig(organizationName, updatedName, "Terraform acceptance agent updated", "Terraform acceptance role updated", nil, nil),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_agent.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_agent.test", "description", "Terraform acceptance agent updated"),
					resource.TestCheckNoResourceAttr("agyn_agent.test", "nickname"),
					resource.TestCheckResourceAttr("agyn_agent.test", "role", "Terraform acceptance role updated"),
					resource.TestCheckResourceAttr("agyn_agent.test", "init_image", os.Getenv("AGYN_AGENT_INIT_IMAGE")),
					resource.TestCheckResourceAttr("agyn_agent.test", "capabilities.#", "0"),
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
	nickname := "tf-acc-import-nickname"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAgentConfig(organizationName, resourceName, "Terraform acceptance agent", "Terraform acceptance role", &nickname, []string{"docker"}),
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

func TestAccAgynAgent_invalidNickname(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAgynAgentInvalidNicknameConfig(organizationName),
				ExpectError: regexp.MustCompile("must contain only lowercase letters"),
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

func testAccAgynAgentConfig(organizationName, title, description, role string, nickname *string, capabilities []string) string {
	nicknameLine := ""
	if nickname != nil {
		nicknameLine = fmt.Sprintf("\n\t\tnickname    = %q", *nickname)
	}
	capabilityLine := formatCapabilitiesLine(capabilities, "\t\t")
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
	  init_image  = %q
%s
%s
}
`, testAccProviderConfig(), organizationName, title, description, role, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"), os.Getenv("AGYN_AGENT_INIT_IMAGE"), nicknameLine, capabilityLine)
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
	  init_image    = %q
	  configuration = "{invalid"
}
`, testAccProviderConfig(), organizationName, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"), os.Getenv("AGYN_AGENT_INIT_IMAGE"))
}

func testAccAgynAgentInvalidNicknameConfig(organizationName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_agent" "test" {
	  organization_id = agyn_organization.test.id
	  name          = "invalid"
	  nickname      = "Invalid Nick"
	  role          = "invalid"
	  model         = %q
	  image         = %q
	  init_image    = %q
}
`, testAccProviderConfig(), organizationName, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"), os.Getenv("AGYN_AGENT_INIT_IMAGE"))
}
