package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynPrivateNetwork_basic(t *testing.T) {
	organizationName := acctest.RandomWithPrefix("tf-acc-org")
	networkName := acctest.RandomWithPrefix("tf-acc-network")
	resourceName := acctest.RandomWithPrefix("tf-acc-resource")
	groupName := acctest.RandomWithPrefix("tf-acc-group")
	userSubject := fmt.Sprintf("%s@example.com", acctest.RandomWithPrefix("tf-acc-user"))
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccUserPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynPrivateNetworkConfig(organizationName, networkName, resourceName, groupName, userSubject),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_network.test", "name", networkName),
					resource.TestCheckResourceAttrSet("agyn_network.test", "id"),
					resource.TestCheckResourceAttr("agyn_tunnel_credential.test", "enrollment_jwt_revealed", "true"),
					resource.TestCheckResourceAttrSet("agyn_tunnel_credential.test", "enrollment_jwt"),
					resource.TestCheckResourceAttr("agyn_private_resource.test", "protocol", "tcp"),
					resource.TestCheckResourceAttr("agyn_private_resource.test", "target_ports.0", "5432"),
					resource.TestCheckResourceAttr("agyn_group.test", "source", "platform"),
					resource.TestCheckResourceAttr("agyn_group_membership.test", "member_type", "user"),
					resource.TestCheckResourceAttr("agyn_private_resource_access.test", "principal_type", "group"),
				),
			},
		},
	})
}

func testAccAgynPrivateNetworkConfig(organizationName, networkName, resourceName, groupName, userSubject string) string {
	return fmt.Sprintf(`
%s

resource "agyn_organization" "test" {
	  name = %q
}

resource "agyn_user" "test" {
	  oidc_subject = %q
	  name         = "Terraform Private Networks User"
	  nickname     = %q
	  cluster_role = "none"
}

resource "agyn_network" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  description     = "Terraform acceptance private network"
}

resource "agyn_tunnel_credential" "test" {
	  network_id = agyn_network.test.id
}

resource "agyn_private_resource" "test" {
	  network_id      = agyn_network.test.id
	  name            = %q
	  protocol        = "tcp"
	  target_host     = "postgres.internal"
	  target_ports    = [5432]
	  intercept_host  = "postgres.agyn.internal"
	  intercept_ports = [5432]
}

resource "agyn_group" "test" {
	  organization_id = agyn_organization.test.id
	  name            = %q
	  description     = "Terraform acceptance group"
}

resource "agyn_group_membership" "test" {
	  group_id    = agyn_group.test.id
	  member_type = "user"
	  member_id   = agyn_user.test.identity_id
}

resource "agyn_private_resource_access" "test" {
	  private_resource_id = agyn_private_resource.test.id
	  principal_type      = "group"
	  principal_id        = agyn_group.test.id
}
`, testAccProviderConfig(), organizationName, userSubject, acctest.RandomWithPrefix("tf-acc-nickname"), networkName, resourceName, groupName)
}
