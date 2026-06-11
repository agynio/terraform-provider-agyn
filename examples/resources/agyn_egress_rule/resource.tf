resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_secret" "api_token" {
  organization_id = agyn_organization.example.id
  name            = "example-api-token"
  value           = "token-value"
}

resource "agyn_egress_rule" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-api"
  description     = "Allow API calls with injected authentication."
  domain_pattern  = "api.example.com"
  ports           = [443]
  methods         = ["GET", "POST"]
  path_pattern    = "/v1/*"
  action          = "allow"

  header {
    name      = "Authorization"
    scheme    = "bearer"
    secret_id = agyn_secret.api_token.id
  }
}
