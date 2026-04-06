package provider

import (
	"fmt"
	"os"
)

func testAccAgynAgentResourceBlock(name, description, role string) string {
	return fmt.Sprintf(`
resource "agyn_agent" "test" {
	  name        = %q
	  description = %q
	  role        = %q
	  model       = %q
	  image       = %q
	  init_image  = %q
}
`, name, description, role, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"), os.Getenv("AGYN_AGENT_INIT_IMAGE"))
}
