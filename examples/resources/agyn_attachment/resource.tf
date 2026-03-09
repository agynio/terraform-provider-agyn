resource "agyn_attachment" "example" {
  kind      = "tool_agent"
  source_id = agyn_tool.example.id
  target_id = agyn_agent.example.id
}
