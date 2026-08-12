# A volume is a definition, not a disk: one disk is provisioned per agent
# instance and per sandbox that runs the environment mounting it.
resource "agyn_volume" "workspace" {
  environment_id = agyn_environment.build.id
  name           = "workspace"
  mount_path     = "/workspace"
  size           = "10Gi"
  ttl            = "168h"
}

# No size makes the volume ephemeral scratch: a fresh empty mount per workload.
resource "agyn_volume" "cache" {
  environment_id = agyn_environment.build.id
  name           = "cache"
  mount_path     = "/home/agent/.cache"
}

# An MCP sidecar mounts volumes of its own. An MCP's shared_volumes names the
# environment's volumes instead, by the name declared here.
resource "agyn_volume" "index" {
  mcp_id        = agyn_mcp.search.id
  name          = "index"
  mount_path    = "/var/lib/index"
  size          = "5Gi"
  storage_class = "fast"
}
