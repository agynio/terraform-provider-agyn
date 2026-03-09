resource "agyn_workspace_configuration" "example" {
  title       = "My Workspace"
  description = "An example workspace configuration managed by Terraform."
  config = jsonencode({
    name = "production"
  })
}
