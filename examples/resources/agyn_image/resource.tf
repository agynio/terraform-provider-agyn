# An image is registered once; its versions are discovered from the upstream
# repository, so there is nothing to declare per release.
resource "agyn_image" "devcontainer" {
  organization_id = var.organization_id
  name            = "devcontainer-go"
  description     = "Go devcontainer"
  type            = "workspace"
  repository      = "ghcr.io/agynio/devcontainer-go"
  visibility      = "internal"
}

# A private repository names a credential by reference, so the password lives
# in a secret rather than in this configuration or in the state file.
resource "agyn_secret" "registry" {
  organization_id = var.organization_id
  name            = "internal-tools-registry"
  value           = var.registry_password
}

resource "agyn_image" "private" {
  organization_id = var.organization_id
  name            = "internal-tools"
  type            = "mcp"
  repository      = "ghcr.io/acme/internal-tools"
  username        = "robot"
  secret_id       = agyn_secret.registry.id
  visibility      = "internal"
  tag_filter      = "v*"
}
