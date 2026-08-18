# An image the platform seeded, or one this organization registered. The lookup
# sees the organization's own images and every public one, so a demo config can
# name the catalog image it runs instead of carrying its UUID.
data "agyn_image" "workspace" {
  organization_id = var.organization_id
  name            = "devcontainer"
  type            = "workspace"
}

data "agyn_image" "codex" {
  organization_id = var.organization_id
  name            = "codex"
  type            = "agent_runtime"
}

# versions lists the tags the platform discovered, newest first — an environment
# may only name one of them.
resource "agyn_environment" "build" {
  organization_id = var.organization_id
  name            = "build"
  runner_id       = data.agyn_runner.default.id
  availability    = "internal"

  workspace_image_id  = data.agyn_image.workspace.id
  workspace_image_tag = data.agyn_image.workspace.versions[0]

  agent_runtime_image_id  = data.agyn_image.codex.id
  agent_runtime_image_tag = "0.147.0"
}
