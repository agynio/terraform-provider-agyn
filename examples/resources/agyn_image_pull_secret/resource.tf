resource "agyn_image_pull_secret" "example" {
  description = "Example image pull secret."
  registry    = "registry.example.com"
  username    = "registry-user"
  password    = "registry-password"
}
