package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAgynImagePullSecretAttachment_basic(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImagePullSecretAttachmentConfig(agentName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_image_pull_secret_attachment.test", "image_pull_secret_id", "agyn_image_pull_secret.primary", "id"),
					resource.TestCheckResourceAttrPair("agyn_image_pull_secret_attachment.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret_attachment.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynImagePullSecretAttachment_update(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImagePullSecretAttachmentConfig(agentName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_image_pull_secret_attachment.test", "image_pull_secret_id", "agyn_image_pull_secret.primary", "id"),
					resource.TestCheckResourceAttrPair("agyn_image_pull_secret_attachment.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret_attachment.test", "id"),
				),
			},
			{
				Config: testAccAgynImagePullSecretAttachmentConfig(agentName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("agyn_image_pull_secret_attachment.test", "image_pull_secret_id", "agyn_image_pull_secret.secondary", "id"),
					resource.TestCheckResourceAttrPair("agyn_image_pull_secret_attachment.test", "agent_id", "agyn_agent.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_image_pull_secret_attachment.test", "id"),
				),
			},
		},
	})
}

func TestAccAgynImagePullSecretAttachment_import(t *testing.T) {
	agentName := acctest.RandomWithPrefix("tf-acc-agent")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccAgynImagePullSecretAttachmentConfig(agentName, false),
			},
			{
				ResourceName:      "agyn_image_pull_secret_attachment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAgynImagePullSecretAttachmentConfig(agentName string, useSecondary bool) string {
	secretRef := "agyn_image_pull_secret.primary.id"
	if useSecondary {
		secretRef = "agyn_image_pull_secret.secondary.id"
	}

	return fmt.Sprintf(`
%s

%s

resource "agyn_image_pull_secret" "primary" {
	  description = "Terraform acceptance image pull secret"
	  registry    = "registry.example.com"
	  username    = "registry-user"
	  password    = "registry-password"
}

resource "agyn_image_pull_secret" "secondary" {
	  description = "Terraform acceptance image pull secret secondary"
	  registry    = "registry-secondary.example.com"
	  username    = "registry-user-secondary"
	  password    = "registry-password-secondary"
}

resource "agyn_image_pull_secret_attachment" "test" {
	  image_pull_secret_id = %s
	  agent_id             = agyn_agent.test.id
}
`, testAccProviderConfig(), testAccAgynAgentResourceBlock(agentName, "Terraform acceptance agent", "Terraform acceptance role"), secretRef)
}
