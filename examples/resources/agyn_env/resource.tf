resource "agyn_env" "example" {
  name        = "LOG_LEVEL"
  description = "Example environment variable."
  agent_id    = agyn_agent.example.id
  value       = "info"
}

# A secret is delivered by reference: the value never enters the state file.
resource "agyn_env" "token" {
  name      = "API_TOKEN"
  agent_id  = agyn_agent.example.id
  secret_id = agyn_secret.local.id
}
