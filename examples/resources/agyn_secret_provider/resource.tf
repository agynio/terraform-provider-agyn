resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_secret_provider" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-vault"
  description     = "Example Vault-backed secret provider."
  type            = "vault"
  vault = {
    address = "https://vault.example.com"
    token   = "vault-token"
  }
}
