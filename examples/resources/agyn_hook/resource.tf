resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_agent" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-agent"
  role            = "assistant"
  model           = "gpt-4o"
  image           = "ghcr.io/agynio/agent-runtime:v1.0.0"
  init_image      = "ghcr.io/agynio/agent-init:v1.0.0"
  description     = "Example agent managed by Terraform."
}

resource "agyn_hook" "example" {
  agent_id    = agyn_agent.example.id
  event       = "conversation.start"
  function    = "handle_start"
  image       = "ghcr.io/agynio/agent-hook:v1.0.0"
  description = "Example hook."
}
