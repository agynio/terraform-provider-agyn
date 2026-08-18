package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynRunnerDataSource_byName(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-runner")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynRunnerDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.agyn_runner.test", "id", "agyn_runner.test", "id"),
					resource.TestCheckResourceAttr("data.agyn_runner.test", "name", name),
					resource.TestCheckResourceAttr("data.agyn_runner.test", "labels.region", "us-east-1"),
					resource.TestCheckResourceAttr("data.agyn_runner.test", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("data.agyn_runner.test", "capabilities.0", "docker"),
				),
			},
		},
	})
}

func TestAccAgynRunnerDataSource_missing(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccRunnerPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAgynRunnerDataSourceMissingConfig(),
				ExpectError: regexp.MustCompile(`no runner named "tf-acc-absent-runner" is visible`),
			},
		},
	})
}

func testAccAgynRunnerDataSourceConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "agyn_runner" "test" {
	  name = %q
	  labels = {
			region = "us-east-1"
	  }
	  capabilities = ["docker"]
}

data "agyn_runner" "test" {
	  name = agyn_runner.test.name
}
`, testAccProviderConfig(), name)
}

func testAccAgynRunnerDataSourceMissingConfig() string {
	return fmt.Sprintf(`
%s

data "agyn_runner" "test" {
	  name = "tf-acc-absent-runner"
}
`, testAccProviderConfig())
}
