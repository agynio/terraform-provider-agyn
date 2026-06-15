resource "agyn_egress_rule" "github" {
  organization_id = agyn_organization.example.id
  name            = "github-api"
  domain_pattern  = "*.github.com"
  ports           = "443"
  methods         = "GET,POST"
  path_pattern    = "/repos/**"
  action          = "allow"

  injected_header {
    name      = "Authorization"
    scheme    = "bearer"
    secret_id = agyn_secret.github_token.id
  }
}
