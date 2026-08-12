resource "agyn_init_script" "example" {
  agent_id    = agyn_agent.example.id
  description = "Initialize agent workspace."
  script      = <<-EOT
    echo "Preparing agent workspace"
  EOT
}
