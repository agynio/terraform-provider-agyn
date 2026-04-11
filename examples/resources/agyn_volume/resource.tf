resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_volume" "example" {
  organization_id = agyn_organization.example.id
  persistent      = true
  mount_path      = "/data"
  size            = "10Gi"
  description     = "Example persistent volume."
}
