package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
