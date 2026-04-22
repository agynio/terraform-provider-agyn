package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynSecretProvider_basic(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	providerTitle := acctest.RandomWithPrefix("tf-acc-secret-provider")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccSecretProviderPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretProviderConfig(organizationName, providerTitle, "Terraform acceptance secret provider"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "title", providerTitle),
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "description", "Terraform acceptance secret provider"),
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "type", "vault"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "vault.address"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "vault.token"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynSecretProvider_update(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	providerTitle := acctest.RandomWithPrefix("tf-acc-secret-provider")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccSecretProviderPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretProviderConfig(organizationName, providerTitle, "Terraform acceptance secret provider"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "title", providerTitle),
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "description", "Terraform acceptance secret provider"),
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "type", "vault"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "vault.address"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "vault.token"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "id"),
				),
			},
			{
				Config: testAccAgynSecretProviderConfig(organizationName, providerTitle+"-updated", "Terraform acceptance secret provider updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "title", providerTitle+"-updated"),
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "description", "Terraform acceptance secret provider updated"),
					resource.TestCheckResourceAttr("agyn_secret_provider.test", "type", "vault"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "vault.address"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "vault.token"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret_provider.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynSecretProvider_import(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	providerTitle := acctest.RandomWithPrefix("tf-acc-secret-provider")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccSecretProviderPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretProviderConfig(organizationName, providerTitle, "Terraform acceptance secret provider"),
			},
			{
				ResourceName:            "agyn_secret_provider.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"organization_id", "vault.token"},
			},
		},
	})
}

func testAccSecretProviderPreCheck(t *testing.T) {
	testAccOrganizationPreCheck(t)
	if os.Getenv("AGYN_VAULT_ADDRESS") == "" {
		t.Skip("AGYN_VAULT_ADDRESS must be set for secret provider acceptance tests")
	}
	if os.Getenv("AGYN_VAULT_TOKEN") == "" {
		t.Skip("AGYN_VAULT_TOKEN must be set for secret provider acceptance tests")
	}
}

func testAccAgynSecretProviderConfig(organizationName, title, description string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_secret_provider" "test" {
	  organization_id = agyn_organization.test.id
	  title           = %q
	  description     = %q
	  type            = "vault"
	  vault = {
		  address = %q
		  token   = %q
	  }
}
`, testAccProviderConfig(), organizationName, title, description, os.Getenv("AGYN_VAULT_ADDRESS"), os.Getenv("AGYN_VAULT_TOKEN"))
}
