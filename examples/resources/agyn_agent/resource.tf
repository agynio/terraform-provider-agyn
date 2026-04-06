resource "agyn_agent" "example" {
  title         = "My Agent"
  description   = "An example agent managed by Terraform."
  name          = "my-agent"
  role          = "assistant"
  model         = "gpt-4o"
  system_prompt = "Assist with workspace automation."
  when_busy     = "wait"
  init_image    = "ghcr.io/agynio/agent-init-codex:v1.0.0"
}
