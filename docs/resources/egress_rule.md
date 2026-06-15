# agyn_egress_rule

Manages an Agyn egress rule.

## Example

```hcl
resource "agyn_egress_rule" "github" {
  organization_id = agyn_organization.main.id
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
```

## Schema

- `organization_id` (String, Required) Organization identifier.
- `name` (String, Required) Rule name.
- `description` (String, Optional) Description.
- `domain_pattern` (String, Required) Hostname pattern.
- `ports` (String, Optional) Comma-separated ports.
- `methods` (String, Optional) Comma-separated methods.
- `path_pattern` (String, Optional) Request path glob.
- `action` (String, Optional) `allow` or `deny`. Omit only when at least one `injected_header` is configured.
- `injected_header` (Block List, Optional) Headers to inject.
  - `name` (String, Required) Header name.
  - `scheme` (String, Optional) `bearer` or `basic`.
  - `value` (String, Sensitive, Optional) Literal credential.
  - `secret_id` (String, Optional) Organization Secret reference.

Exactly one of `value` and `secret_id` must be set for each `injected_header`.
