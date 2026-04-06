resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_image_pull_secret" "example" {
  organization_id = agyn_organization.example.id
  description = "Example image pull secret."
  registry    = "registry.example.com"
  username    = "registry-user"
  password    = "registry-password"
}
