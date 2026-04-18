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
				Config: testAccAgynRunnerConfig(name, []string{"docker"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("agyn_runner.test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("agyn_runner.test", "capabilities.0", "docker"),
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
				Config: testAccAgynRunnerConfigWithLabels(name, "test", "infra", []string{"docker"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.%", "2"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.environment", "test"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.team", "infra"),
					resource.TestCheckResourceAttr("agyn_runner.test", "capabilities.#", "1"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
			{
				Config: testAccAgynRunnerConfigWithLabels(updatedName, "prod", "platform", []string{"docker", "gpu"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", updatedName),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.%", "2"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.environment", "prod"),
					resource.TestCheckResourceAttr("agyn_runner.test", "labels.team", "platform"),
					resource.TestCheckResourceAttr("agyn_runner.test", "capabilities.#", "2"),
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
	updatedOrganizationID := organizationID + "-updated"
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerOrgPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynRunnerConfigWithOrganizationID(name, organizationID, []string{"docker"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("agyn_runner.test", "organization_id", organizationID),
					resource.TestCheckResourceAttr("agyn_runner.test", "capabilities.#", "1"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "identity_id"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "service_token"),
					resource.TestCheckResourceAttrSet("agyn_runner.test", "id"),
				),
			},
			{
				Config: testAccAgynRunnerConfigWithOrganizationID(name, updatedOrganizationID, []string{"docker"}),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("agyn_runner.test", plancheck.ResourceActionReplace),
					},
				},
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
				Config: testAccAgynRunnerConfig(name, []string{"docker"}),
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

func testAccAgynRunnerConfig(name string, capabilities []string) string {
	capabilityLine := formatCapabilitiesLine(capabilities, "\t  ")
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name = %q
%s
}
`, testAccProviderConfig(), name, capabilityLine)
}

func testAccAgynRunnerConfigWithLabels(name, environment, team string, capabilities []string) string {
	capabilityLine := formatCapabilitiesLine(capabilities, "\t  ")
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name = %q
	  labels = {
			environment = %q
			team        = %q
	  }
%s
}
`, testAccProviderConfig(), name, environment, team, capabilityLine)
}

func testAccAgynRunnerConfigWithOrganizationID(name, organizationID string, capabilities []string) string {
	capabilityLine := formatCapabilitiesLine(capabilities, "\t  ")
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name            = %q
	  organization_id = %q
%s
}
`, testAccProviderConfig(), name, organizationID, capabilityLine)
}
