package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynOrganization_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynOrganizationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_organization.test", "name", name),
					resource.TestCheckResourceAttrSet("agyn_organization.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynOrganization_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-org")
	updatedName := name + "-updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynOrganizationConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_organization.test", "name", name),
					resource.TestCheckResourceAttrSet("agyn_organization.test", "id"),
				),
			},
			{
				Config: testAccAgynOrganizationConfig(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_organization.test", "name", updatedName),
					resource.TestCheckResourceAttrSet("agyn_organization.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynOrganization_import(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynOrganizationConfig(name),
			},
			{
				ResourceName:      "agyn_organization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrganizationPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv("AGYN_API_TOKEN") == "" {
		t.Skip("AGYN_API_TOKEN must be set for organization acceptance tests")
	}
}

func testAccAgynOrganizationConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}
`, testAccProviderConfig(), name)
}
