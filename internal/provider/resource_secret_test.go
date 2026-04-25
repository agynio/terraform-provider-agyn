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
	secretName := acctest.RandomWithPrefix("tf-acc-secret")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretLocalConfig(organizationName, secretName, "Terraform acceptance secret", "secret-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret.test", "name", secretName),
					resource.TestCheckResourceAttr("agyn_secret.test", "description", "Terraform acceptance secret"),
					resource.TestCheckResourceAttr("agyn_secret.test", "value", "secret-value"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_secret.test", "id"),
				),
			},
			{
				Config: testAccAgynSecretLocalConfig(organizationName, secretName+"-updated", "Terraform acceptance secret updated", "secret-value-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_secret.test", "name", secretName+"-updated"),
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
	providerName := acctest.RandomWithPrefix("tf-acc-secret-provider")
	secretName := acctest.RandomWithPrefix("tf-acc-secret")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccSecretRemotePreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynSecretRemoteConfig(organizationName, providerName, secretName, "Terraform acceptance remote secret", os.Getenv("AGYN_VAULT_REMOTE_NAME")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_secret.test", "provider_id", "agyn_secret_provider.test", "id"),
					resource.TestCheckResourceAttr("agyn_secret.test", "name", secretName),
					resource.TestCheckResourceAttr("agyn_secret.test", "description", "Terraform acceptance remote secret"),
					resource.TestCheckResourceAttr("agyn_secret.test", "provider_secret_name", os.Getenv("AGYN_VAULT_REMOTE_NAME")),
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

func testAccAgynSecretLocalConfig(organizationName, name, description, value string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_secret" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  description     = %q
	  value           = %q
}
`, testAccProviderConfig(), organizationName, name, description, value)
}

func testAccAgynSecretRemoteConfig(organizationName, providerName, secretName, description, remoteName string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_secret_provider" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  description     = "Terraform acceptance secret provider"
	  type            = "vault"
	  vault = {
		  address = %q
		  token   = %q
	  }
}

resource "agyn_secret" "test" {
	  organization_id   = agyn_organization.test.id
	  name              = %q
	  description       = %q
	  provider_id       = agyn_secret_provider.test.id
	  provider_secret_name = %q
}
`, testAccProviderConfig(), organizationName, providerName, os.Getenv("AGYN_VAULT_ADDRESS"), os.Getenv("AGYN_VAULT_TOKEN"), secretName, description, remoteName)
}
