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

resource "agyn_image_pull_secret" "example" {
  organization_id = agyn_organization.example.id
  description     = "Example image pull secret."
  registry        = "registry.example.com"
  username        = "registry-user"
  password        = "registry-password"
}

resource "agyn_image_pull_secret_attachment" "example" {
  image_pull_secret_id = agyn_image_pull_secret.example.id
  agent_id             = agyn_agent.example.id
}
