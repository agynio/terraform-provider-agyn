resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_secret" "local" {
  organization_id = agyn_organization.example.id
  title           = "example-local-secret"
  description     = "Local secret value."
  value           = "local-secret-value"
}

resource "agyn_secret_provider" "vault" {
  organization_id = agyn_organization.example.id
  title           = "example-vault"
  description     = "Example Vault-backed secret provider."
  type            = "vault"
  vault = {
    address = "https://vault.example.com"
    token   = "vault-token"
  }
}

resource "agyn_secret" "remote" {
  organization_id    = agyn_organization.example.id
  title              = "example-remote-secret"
  description        = "Vault-backed secret reference."
  secret_provider_id = agyn_secret_provider.vault.id
  remote_name        = "secret/platform/keys/api_key"
}
