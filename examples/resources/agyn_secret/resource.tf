resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_secret" "local" {
  organization_id = agyn_organization.example.id
  name            = "example-local-secret"
  description     = "Local secret value."
  value           = "local-secret-value"
}

resource "agyn_secret_provider" "vault" {
  organization_id = agyn_organization.example.id
  name            = "example-vault"
  description     = "Example Vault-backed secret provider."
  type            = "vault"
  vault = {
    address = "https://vault.example.com"
    token   = "vault-token"
  }
}

resource "agyn_secret" "remote" {
  organization_id      = agyn_organization.example.id
  name                 = "example-remote-secret"
  description          = "Vault-backed secret reference."
  provider_id          = agyn_secret_provider.vault.id
  provider_secret_name = "secret/platform/keys/api_key"
}
