package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/agynio/terraform-provider-agyn/internal/teamapi"
)

func TestAccAgynAttachment_mcpServerWorkspaceConfiguration(t *testing.T) {
	mcpTitle := acctest.RandomWithPrefix("tf-acc-mcp")
	workspaceTitle := acctest.RandomWithPrefix("tf-acc-workspace")
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckAgynAttachmentDestroy(context.Background()),
		Steps: []resource.TestStep{
			{
				Config: testAccAgynAttachmentConfig(mcpTitle, workspaceTitle),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("agyn_attachment.test", "kind", "mcpServer_workspaceConfiguration"),
					resource.TestCheckResourceAttrSet("agyn_attachment.test", "id"),
					resource.TestCheckResourceAttrPair("agyn_attachment.test", "source_id", "agyn_mcp_server.test", "id"),
					resource.TestCheckResourceAttrPair("agyn_attachment.test", "target_id", "agyn_workspace_configuration.test", "id"),
					resource.TestCheckResourceAttrSet("agyn_attachment.test", "source_type"),
					resource.TestCheckResourceAttrSet("agyn_attachment.test", "target_type"),
				),
			},
		},
	})
}

func testAccAgynAttachmentConfig(mcpTitle, workspaceTitle string) string {
	return fmt.Sprintf(`
%s

resource "agyn_mcp_server" "test" {
  title       = %q
  description = "Terraform acceptance MCP server"
  config = jsonencode({
    namespace = %q
    command   = "mcp start --stdio"
  })
}

resource "agyn_workspace_configuration" "test" {
  title       = %q
  description = "Terraform acceptance workspace config"
  config = jsonencode({
    platform   = "auto"
    ttlSeconds = 600
    enableDinD = false
  })
}

resource "agyn_attachment" "test" {
  kind      = "mcpServer_workspaceConfiguration"
  source_id = agyn_mcp_server.test.id
  target_id = agyn_workspace_configuration.test.id
}
`, testAccProviderConfig(), mcpTitle, mcpTitle, workspaceTitle)
}

func testAccCheckAgynAttachmentDestroy(ctx context.Context) func(*terraform.State) error {
	return func(state *terraform.State) error {
		client, err := teamapi.NewClient(teamapi.Config{BaseURL: os.Getenv("AGYN_BASE_URL")})
		if err != nil {
			return err
		}

		for _, rs := range state.RootModule().Resources {
			if rs.Type != "agyn_attachment" {
				continue
			}
			if rs.Primary == nil || rs.Primary.ID == "" {
				continue
			}

			kind := rs.Primary.Attributes["kind"]
			if kind == "mcpServer_workspaceConfiguration" {
				_, err := client.FindGraphEdge(ctx, rs.Primary.ID)
				if err == nil {
					return fmt.Errorf("graph attachment %s still exists", rs.Primary.ID)
				}
				if errors.Is(err, teamapi.ErrGraphEdgeNotFound) {
					continue
				}
				return err
			}

			_, err := client.GetAttachment(ctx, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("attachment %s still exists", rs.Primary.ID)
			}
			if errors.Is(err, teamapi.ErrAttachmentNotFound) {
				continue
			}
			var apiErr *teamapi.APIError
			if errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound {
				continue
			}
			return err
		}
		return nil
	}
}
