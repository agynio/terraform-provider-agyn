resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_app" "example" {
  organization_id = agyn_organization.example.id
  slug            = "example-app"
  name            = "Example App"
  visibility      = "internal"
}

resource "agyn_app_installation" "example" {
  app_id          = agyn_app.example.id
  organization_id = agyn_organization.example.id
  slug            = "example-installation"
  configuration = jsonencode({
    workspace = "default"
  })
}
