resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_app" "example" {
  organization_id = agyn_organization.example.id
  slug            = "example-app"
  name            = "Example App"
  visibility      = "internal"
  description     = "Example app managed by Terraform."
  permissions     = ["agents:read"]
}
