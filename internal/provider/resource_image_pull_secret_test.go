package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynImagePullSecret_basic(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImagePullSecretConfig(organizationName, "Terraform acceptance image pull secret", "registry.example.com", "registry-user", "registry-password"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "description", "Terraform acceptance image pull secret"),
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "registry", "registry.example.com"),
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "username", "registry-user"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "password"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynImagePullSecret_update(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImagePullSecretConfig(organizationName, "Terraform acceptance image pull secret", "registry.example.com", "registry-user", "registry-password"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "description", "Terraform acceptance image pull secret"),
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "registry", "registry.example.com"),
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "username", "registry-user"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "password"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "id"),
				),
			},
			{
				Config: testAccAgynImagePullSecretConfig(organizationName, "Terraform acceptance image pull secret updated", "registry-updated.example.com", "registry-user-updated", "registry-password-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "description", "Terraform acceptance image pull secret updated"),
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "registry", "registry-updated.example.com"),
					resource.TestCheckResourceAttr("agyn_image_pull_secret.test", "username", "registry-user-updated"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "organization_id"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "password"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynImagePullSecret_import(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccOrganizationPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImagePullSecretConfig(organizationName, "Terraform acceptance image pull secret", "registry.example.com", "registry-user", "registry-password"),
			},
			{
				ResourceName:            "agyn_image_pull_secret.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"organization_id", "password"},
			},
		},
	})
}

func testAccAgynImagePullSecretConfig(organizationName, description, registry, username, password string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_image_pull_secret" "test" {
	  organization_id = agyn_organization.test.id
	  description = %q
	  registry    = %q
	  username    = %q
	  password    = %q
}
`, testAccProviderConfig(), organizationName, description, registry, username, password)
}
