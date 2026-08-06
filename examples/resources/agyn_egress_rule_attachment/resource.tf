resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_agent" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-agent"
  nickname        = "example-agent"
  role            = "assistant"
  model           = "gpt-4o"
  image           = "ghcr.io/agynio/agent-runtime:v1.0.0"
  availability    = "private"
}

resource "agyn_egress_rule" "example" {
  organization_id = agyn_organization.example.id
  name            = "example-api"
  domain_pattern  = "api.example.com"
  action          = "allow"
}

resource "agyn_egress_rule_attachment" "example" {
  organization_id = agyn_organization.example.id
  rule_id         = agyn_egress_rule.example.id
  agent_id        = agyn_agent.example.id
}
