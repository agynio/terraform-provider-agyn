resource "agyn_mcp_server" "example" {
  title       = "My MCP Server"
  description = "An example MCP server managed by Terraform."
  config = jsonencode({
    url = "https://mcp.example.com"
  })
}
