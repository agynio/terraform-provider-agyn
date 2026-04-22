package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynSecret_local(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	secretTitle := acctest.RandomWithPrefix("tf-acc-secret")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretLocalConfig(organizationName, secretTitle, "Terraform acceptance secret", "secret-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret.test", "title", secretTitle),
					resource.TestCheckResourceAttr("agyn_secret.test", "description", "Terraform acceptance secret"),
					resource.TestCheckResourceAttr("agyn_secret.test", "value", "secret-value"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "id"),
				),
			},
			{
				Config: testAccAgynSecretLocalConfig(organizationName, secretTitle+"-updated", "Terraform acceptance secret updated", "secret-value-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret.test", "title", secretTitle+"-updated"),
					resource.TestCheckResourceAttr("agyn_secret.test", "description", "Terraform acceptance secret updated"),
					resource.TestCheckResourceAttr("agyn_secret.test", "value", "secret-value-updated"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynSecret_remote(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	providerTitle := acctest.RandomWithPrefix("tf-acc-secret-provider")
	secretTitle := acctest.RandomWithPrefix("tf-acc-secret")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccSecretRemotePreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretRemoteConfig(organizationName, providerTitle, secretTitle, "Terraform acceptance remote secret", os.Getenv("AGYN_VAULT_REMOTE_NAME")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_secret.test", "secret_provider_id", "agyn_secret_provider.test", "id"),
					resource.TestCheckResourceAttr("agyn_secret.test", "title", secretTitle),
					resource.TestCheckResourceAttr("agyn_secret.test", "description", "Terraform acceptance remote secret"),
					resource.TestCheckResourceAttr("agyn_secret.test", "remote_name", os.Getenv("AGYN_VAULT_REMOTE_NAME")),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "id"),
				),
			},
		},
	})
}

func testAccSecretRemotePreCheck(t *testing.T) {
	testAccSecretProviderPreCheck(t)
	if os.Getenv("AGYN_VAULT_REMOTE_NAME") == "" {
		t.Skip("AGYN_VAULT_REMOTE_NAME must be set for remote secret acceptance tests")
	}
}

func testAccAgynSecretLocalConfig(organizationName, title, description, value string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_secret" "test" {
	  organization_id = agyn_organization.test.id
	  title           = %q
	  description     = %q
	  value           = %q
}
`, testAccProviderConfig(), organizationName, title, description, value)
}

func testAccAgynSecretRemoteConfig(organizationName, providerTitle, secretTitle, description, remoteName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_secret_provider" "test" {
	  organization_id = agyn_organization.test.id
	  title           = %q
	  description     = "Terraform acceptance secret provider"
	  type            = "vault"
	  vault = {
		  address = %q
		  token   = %q
	  }
}

resource "agyn_secret" "test" {
	  organization_id   = agyn_organization.test.id
	  title             = %q
	  description       = %q
	  secret_provider_id = agyn_secret_provider.test.id
	  remote_name       = %q
}
`, testAccProviderConfig(), organizationName, providerTitle, os.Getenv("AGYN_VAULT_ADDRESS"), os.Getenv("AGYN_VAULT_TOKEN"), secretTitle, description, remoteName)
}
