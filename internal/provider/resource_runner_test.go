package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccAgynRunner_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-runner")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynRunnerConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynRunner_update(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-runner")
	updatedName := name + "-updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynRunnerConfigWithLabels(name, "test", "infra"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.%", "2"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.environment", "test"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.team", "infra"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
			{
				Config: testAccAgynRunnerConfigWithLabels(updatedName, "prod", "platform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.%", "2"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.environment", "prod"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.team", "platform"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynRunner_organizationIDRequiresReplace(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-runner")
	organizationID := os.Getenv("AGYN_ORGANIZATION_ID")
	updatedOrganizationID := os.Getenv("AGYN_ORGANIZATION_ID_ALT")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerOrgReplacePreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynRunnerConfigWithOrganizationID(name, organizationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("agyn_runner.test", "organization_id", organizationID),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
			{
				Config: testAccAgynRunnerConfigWithOrganizationID(name, updatedOrganizationID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("agyn_runner.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("agyn_runner.test", "organization_id", updatedOrganizationID),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynRunner_import(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-runner")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynRunnerConfig(name),
			},
			{
				ResourceName:            "agyn_runner.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"service_token"},
			},
		},
	})
}

func testAccRunnerPreCheck(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv("AGYN_API_TOKEN") == "" {
		t.Skip("AGYN_API_TOKEN must be set for runner acceptance tests")
	}
}

func testAccRunnerOrgPreCheck(t *testing.T) {
	testAccRunnerPreCheck(t)
	if os.Getenv("AGYN_ORGANIZATION_ID") == "" {
		t.Skip("AGYN_ORGANIZATION_ID must be set for runner organization tests")
	}
}

func testAccRunnerOrgReplacePreCheck(t *testing.T) {
	testAccRunnerOrgPreCheck(t)
	if os.Getenv("AGYN_ORGANIZATION_ID_ALT") == "" {
		t.Skip("AGYN_ORGANIZATION_ID_ALT must be set for runner organization replace tests")
	}
}

func testAccAgynRunnerConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name = %q
}
`, testAccProviderConfig(), name)
}

func testAccAgynRunnerConfigWithLabels(name, environment, team string) string {
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name = %q
	  labels = {
			environment = %q
			team        = %q
	  }
}
`, testAccProviderConfig(), name, environment, team)
}

func testAccAgynRunnerConfigWithOrganizationID(name, organizationID string) string {
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name            = %q
	  organization_id = %q
}
`, testAccProviderConfig(), name, organizationID)
}
