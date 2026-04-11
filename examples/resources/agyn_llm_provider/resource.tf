resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_llm_provider" "example" {
  organization_id = agyn_organization.example.id
  endpoint        = "https://api.example.com"
  auth_method     = "bearer"
  token           = "example-token"
}
