package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynAttachment_expectError(t *testing.T) {
	resourceName := acctest.RandomWithPrefix("tf-acc-attachment")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccAgynAttachmentConfig(resourceName),
				ExpectError: regexp.MustCompile("Create Attachment Failed"),
			},
		},
	})
}

func testAccAgynAttachmentConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "agyn_agent" "attachment" {
  title       = %q
  description = "Terraform acceptance attachment agent"
  config = jsonencode({
    name = %q
    role = "Terraform acceptance role"
  })
}

resource "agyn_tool" "attachment" {
  name        = %q
  description = "Terraform acceptance attachment tool"
  type        = "send_message"
  config = jsonencode({
    prompt = "Attachment prompt"
  })
}

resource "agyn_attachment" "test" {
  kind      = "agent_tool"
  source_id = agyn_agent.attachment.id
  target_id = agyn_tool.attachment.id
}
`, testAccProviderConfig(), name, name, name)
}
