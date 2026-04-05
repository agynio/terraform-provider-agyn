resource "agyn_image_pull_secret_attachment" "example" {
  image_pull_secret_id = agyn_image_pull_secret.example.id
  agent_id             = agyn_agent.example.id
}
