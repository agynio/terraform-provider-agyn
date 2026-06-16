resource "agyn_egress_rule_attachment" "github" {
  rule_id  = agyn_egress_rule.github.id
  agent_id = agyn_agent.example.id
}
