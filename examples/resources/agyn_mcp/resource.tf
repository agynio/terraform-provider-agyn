resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_agent" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-agent"
  role            = "assistant"
  model           = "gpt-4o"
  image           = "ghcr.io/agynio/agent-runtime:v1.0.0"
  description     = "Example agent managed by Terraform."
}

resource "agyn_mcp" "example" {
  agent_id    = agyn_agent.example.id
  name        = "example-mcp"
  image       = "ghcr.io/agynio/mcp-server:v1.0.0"
  command     = "mcp-server --port 8080"
  description = "Example MCP service."
}
