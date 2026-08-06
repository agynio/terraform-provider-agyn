# An environment names a runner, a flavor on that runner, and the images a
# workload runs. Agents and sandboxes run in environments.
resource "agyn_environment" "build" {
  organization_id     = var.organization_id
  name                = "build"
  runner_id           = var.runner_id
  flavor              = "ram-2gb"
  workspace_image_id  = agyn_image.devcontainer.id
  workspace_image_tag = "1.2.0"

  # The agent CLI. Omit both to make a workspace-only environment: usable by a
  # sandbox, rejected when creating an agent.
  agent_runtime_image_id  = agyn_image.runtime_codex.id
  agent_runtime_image_tag = "0.146.0"
}
