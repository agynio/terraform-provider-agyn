# agyn_egress_rule_attachment

Attaches an Agyn egress rule to an agent.

## Example

```hcl
resource "agyn_egress_rule_attachment" "github" {
  rule_id  = agyn_egress_rule.github.id
  agent_id = agyn_agent.main.id
}
```

## Schema

- `rule_id` (String, Required) Egress rule identifier.
- `agent_id` (String, Required) Agent identifier.
- `id` (String, Computed) Attachment identifier.
