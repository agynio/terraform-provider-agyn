resource "agyn_workspace_configuration" "example" {
  title          = "My Workspace"
  description    = "An example workspace configuration managed by Terraform."
  image          = "ghcr.io/agynio/workspace:latest"
  platform       = "auto"
  enable_dind    = false
  ttl_seconds    = 600
  cpu_limit      = "2"
  memory_limit   = "4Gi"
  initial_script = "echo 'workspace ready'"

  env {
    name  = "APP_ENV"
    value = "production"
  }

  volumes {
    enabled    = true
    mount_path = "/workspace"
  }

  nix = jsonencode({
    packages = ["git", "curl"]
  })
}
