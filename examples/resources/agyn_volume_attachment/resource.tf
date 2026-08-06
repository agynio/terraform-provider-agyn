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

resource "agyn_volume" "example" {
  organization_id = agyn_organization.example.id
  persistent      = true
  mount_path      = "/data"
  size            = "10Gi"
  description     = "Example persistent volume."
}

resource "agyn_volume_attachment" "example" {
  volume_id = agyn_volume.example.id
  agent_id  = agyn_agent.example.id
}
