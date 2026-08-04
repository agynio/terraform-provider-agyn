resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_agent" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-agent"
  nickname        = "example-agent"
  role            = "assistant"
  model           = "gpt-4o"
  image           = "ghcr.io/agynio/agent-runtime:v1.0.0"
  init_image      = "ghcr.io/agynio/agent-init:v1.0.0"
  description     = "Example agent managed by Terraform."
  availability    = "internal"
}
