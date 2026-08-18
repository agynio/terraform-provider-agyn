package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// The data source reads back an image this config registers, so the test does
// not depend on what the organization's catalog already holds.
func TestAccAgynImageDataSource_byName(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	imageName := acctest.RandomWithPrefix("tf-acc-image")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImageDataSourceConfig(organizationName, imageName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.agyn_image.test", "id", "agyn_image.test", "id"),
					resource.TestCheckResourceAttr("data.agyn_image.test", "name", imageName),
					resource.TestCheckResourceAttr("data.agyn_image.test", "type", "workspace"),
					resource.TestCheckResourceAttr("data.agyn_image.test", "repository", "ghcr.io/agynio/devcontainer"),
					resource.TestCheckResourceAttr("data.agyn_image.test", "visibility", "internal"),
					resource.TestCheckResourceAttrSet("data.agyn_image.test", "versions.#"),
				),
			},
		},
	})
}

func TestAccAgynImageDataSource_missing(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAgynImageDataSourceMissingConfig(organizationName),
				ExpectError: regexp.MustCompile(`no image named "tf-acc-absent" is visible`),
			},
		},
	})
}

func testAccAgynImageDataSourceConfig(organizationName, imageName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_image" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  type            = "workspace"
	  repository      = "ghcr.io/agynio/devcontainer"
	  visibility      = "internal"
}

data "agyn_image" "test" {
	  organization_id = agyn_organization.test.id
	  name            = agyn_image.test.name
	  type            = "workspace"
}
`, testAccProviderConfig(), organizationName, imageName)
}

func testAccAgynImageDataSourceMissingConfig(organizationName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

data "agyn_image" "test" {
	  organization_id = agyn_organization.test.id
	  name            = "tf-acc-absent"
}
`, testAccProviderConfig(), organizationName)
}
