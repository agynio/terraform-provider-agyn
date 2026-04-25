package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynAppInstallation_basic(t *testing.T) {
	appSlug := acctest.RandomWithPrefix("tf-acc-app")
	appName := fmt.Sprintf("Terraform acceptance app %s", appSlug)
	installationSlug := acctest.RandomWithPrefix("tf-acc-install")
	organizationID := os.Getenv("AGYN_ORGANIZATION_ID")
	configuration := `{"setting":"value"}`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccAppPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAppInstallationConfig(organizationID, appSlug, appName, installationSlug, configuration),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_app_installation.test", "app_id", "agyn_app.test", "id"),
					resource.TestCheckResourceAttr("agyn_app_installation.test", "organization_id", organizationID),
					resource.TestCheckResourceAttr("agyn_app_installation.test", "slug", installationSlug),
					resource.TestCheckResourceAttr("agyn_app_installation.test", "configuration", configuration),
					resource.TestCheckResourceAttrSet("agyn_app_installation.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynAppInstallation_update(t *testing.T) {
	appSlug := acctest.RandomWithPrefix("tf-acc-app")
	appName := fmt.Sprintf("Terraform acceptance app %s", appSlug)
	installationSlug := acctest.RandomWithPrefix("tf-acc-install")
	updatedInstallationSlug := installationSlug + "-updated"
	organizationID := os.Getenv("AGYN_ORGANIZATION_ID")
	configuration := `{"setting":"value"}`
	updatedConfiguration := `{"setting":"updated"}`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccAppPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAppInstallationConfig(organizationID, appSlug, appName, installationSlug, configuration),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_app_installation.test", "slug", installationSlug),
					resource.TestCheckResourceAttr("agyn_app_installation.test", "configuration", configuration),
					resource.TestCheckResourceAttrSet("agyn_app_installation.test", "id"),
				),
			},
			{
				Config: testAccAgynAppInstallationConfig(organizationID, appSlug, appName, updatedInstallationSlug, updatedConfiguration),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_app_installation.test", "slug", updatedInstallationSlug),
					resource.TestCheckResourceAttr("agyn_app_installation.test", "configuration", updatedConfiguration),
					resource.TestCheckResourceAttrSet("agyn_app_installation.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynAppInstallation_import(t *testing.T) {
	appSlug := acctest.RandomWithPrefix("tf-acc-app")
	appName := fmt.Sprintf("Terraform acceptance app %s", appSlug)
	installationSlug := acctest.RandomWithPrefix("tf-acc-install")
	organizationID := os.Getenv("AGYN_ORGANIZATION_ID")
	configuration := `{"setting":"value"}`
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccAppPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAppInstallationConfig(organizationID, appSlug, appName, installationSlug, configuration),
			},
			{
				ResourceName:      "agyn_app_installation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgynAppInstallationConfig(organizationID, appSlug, appName, installationSlug, configuration string) string {
	return fmt.Sprintf(`
%s

resource "agyn_app" "test" {
	  organization_id = %q
	  slug            = %q
	  name            = %q
	  description     = "Terraform acceptance app"
	  icon            = "https://example.com/icon.png"
	  visibility      = "internal"
	  permissions     = ["read"]
}

resource "agyn_app_installation" "test" {
	  app_id          = agyn_app.test.id
	  organization_id = %q
	  slug            = %q
	  configuration   = %q
}
`, testAccProviderConfig(), organizationID, appSlug, appName, organizationID, installationSlug, configuration)
}
