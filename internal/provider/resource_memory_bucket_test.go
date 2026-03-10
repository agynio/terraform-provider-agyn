package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynMemoryBucket_basic(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-memory")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynMemoryBucketConfig(resourceName, "Terraform acceptance memory bucket"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_memory_bucket.test", "title", resourceName),
					resource.TestCheckResourceAttr("agyn_memory_bucket.test", "description", "Terraform acceptance memory bucket"),
					resource.TestCheckResourceAttr("agyn_memory_bucket.test", "scope", "global"),
					resource.TestCheckResourceAttr("agyn_memory_bucket.test", "collection_prefix", resourceName),
					resource.TestCheckResourceAttrSet("agyn_memory_bucket.test", "id"),
				),
			},
		},
	})
}

func testAccAgynMemoryBucketConfig(title, description string) string {
	return fmt.Sprintf(`
%s

resource "agyn_memory_bucket" "test" {
  title       = %q
  description = %q
  scope             = "global"
  collection_prefix = %q
}
`, testAccProviderConfig(), title, description, title)
}
