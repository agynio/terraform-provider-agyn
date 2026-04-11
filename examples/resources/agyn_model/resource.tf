resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_llm_provider" "example" {
  organization_id = agyn_organization.example.id
  endpoint        = "https://api.example.com"
  auth_method     = "bearer"
  token           = "example-token"
}

resource "agyn_model" "example" {
  organization_id = agyn_organization.example.id
  name            = "gpt-4o"
  llm_provider_id = agyn_llm_provider.example.id
  remote_name     = "gpt-4o"
}
