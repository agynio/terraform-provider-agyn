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

# A private repository names a credential. The password is write-only: the
# platform stores it as a secret and never returns it.
resource "agyn_image" "private" {
  organization_id = var.organization_id
  name            = "internal-tools"
  type            = "mcp"
  repository      = "ghcr.io/acme/internal-tools"
  username        = "robot"
  password        = var.registry_password
  visibility      = "internal"
  tag_filter      = "v*"
}
