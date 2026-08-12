resource "agyn_agent" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-agent"
  nickname        = "example-agent"
  role            = "assistant"
  description     = "Example agent managed by Terraform."

  # The environment supplies the image and compute the agent runs with.
  environment_id = agyn_environment.build.id
  model          = agyn_model.example.id
  image          = "ghcr.io/agynio/agent-runtime:v1.0.0"
  availability   = "internal"
}
