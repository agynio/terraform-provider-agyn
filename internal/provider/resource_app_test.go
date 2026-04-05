package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynApp_basic(t *testing.T) {
	slug := acctest.RandomWithPrefix("tf-acc-app")
	name := fmt.Sprintf("Terraform acceptance app %s", slug)
	organizationID := os.Getenv("AGYN_ORGANIZATION_ID")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccAppPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAppConfig(slug, name, "Terraform acceptance app", "https://example.com/icon.png", organizationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_app.test", "organization_id", organizationID),
					resource.TestCheckResourceAttr("agyn_app.test", "slug", slug),
					resource.TestCheckResourceAttr("agyn_app.test", "name", name),
					resource.TestCheckResourceAttr("agyn_app.test", "description", "Terraform acceptance app"),
					resource.TestCheckResourceAttr("agyn_app.test", "icon", "https://example.com/icon.png"),
					resource.TestCheckResourceAttr("agyn_app.test", "visibility", "internal"),
					resource.TestCheckResourceAttr("agyn_app.test", "permissions.#", "2"),
					resource.TestCheckResourceAttr("agyn_app.test", "permissions.0", "read"),
					resource.TestCheckResourceAttr("agyn_app.test", "permissions.1", "write"),
					resource.TestCheckResourceAttrSet("agyn_app.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_app.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_app.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynApp_import(t *testing.T) {
	slug := acctest.RandomWithPrefix("tf-acc-app")
	name := fmt.Sprintf("Terraform acceptance app %s", slug)
	organizationID := os.Getenv("AGYN_ORGANIZATION_ID")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccAppPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAppConfig(slug, name, "Terraform acceptance app", "https://example.com/icon.png", organizationID),
			},
			{
				ResourceName:            "agyn_app.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_token"},
			},
		},
	})
}

func testAccAppPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv("AGYN_API_TOKEN") == "" {
		t.Skip("AGYN_API_TOKEN must be set for app acceptance tests")
	}
	if os.Getenv("AGYN_ORGANIZATION_ID") == "" {
		t.Skip("AGYN_ORGANIZATION_ID must be set for app acceptance tests")
	}
}

func testAccAgynAppConfig(slug, name, description, icon, organizationID string) string {
	return fmt.Sprintf(`
%s

resource "agyn_app" "test" {
	  organization_id = %q
	  slug            = %q
	  name            = %q
	  description     = %q
	  icon            = %q
	  visibility      = "internal"
	  permissions     = ["read", "write"]
}
`, testAccProviderConfig(), organizationID, slug, name, description, icon)
}
