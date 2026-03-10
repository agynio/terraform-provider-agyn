resource "agyn_mcp_server" "example" {
  title       = "My MCP Server"
  description = "An example MCP server managed by Terraform."
  namespace   = "example"
  command     = "mcp start --stdio"
  workdir     = "/srv/mcp"

  restart {
    max_attempts = 5
    backoff_ms   = 1000
  }

  env {
    name  = "LOG_LEVEL"
    value = "info"
  }
}
