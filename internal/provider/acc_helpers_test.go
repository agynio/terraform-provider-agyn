package provider

import (
	"fmt"
	"os"
	"strings"
)

func testAccAgynAgentResourceBlock(name, description, role string) string {
	return fmt.Sprintf(`
resource "agyn_agent" "test" {
	  name        = %q
	  description = %q
	  role        = %q
	  model       = %q
	  image       = %q
	  }
`, name, description, role, os.Getenv("AGYN_MODEL_ID"), os.Getenv("AGYN_AGENT_IMAGE"))
}

func formatCapabilitiesLine(capabilities []string, indent string) string {
	if len(capabilities) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(capabilities))
	for _, capability := range capabilities {
		quoted = append(quoted, fmt.Sprintf("%q", capability))
	}
	return fmt.Sprintf("\n%s%s", indent, fmt.Sprintf("capabilities = [%s]", strings.Join(quoted, ", ")))
}
