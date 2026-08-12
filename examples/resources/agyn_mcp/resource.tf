# An MCP server runs as a sidecar to every workload in its environment.
resource "agyn_mcp" "search" {
  environment_id = agyn_environment.build.id
  name           = "search"
  image          = "ghcr.io/agynio/mcp-search:v1.0.0"
  command        = "mcp-search --port 8080"
  description    = "Example MCP service."

  # Environment volumes to mount into the sidecar, at the paths the main
  # container uses.
  shared_volumes = ["workspace"]
}
