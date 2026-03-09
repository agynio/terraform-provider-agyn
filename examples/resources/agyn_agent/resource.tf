resource "agyn_agent" "example" {
  title       = "My Agent"
  description = "An example agent managed by Terraform."
  config = jsonencode({
    model       = "gpt-4o"
    temperature = 0.7
  })
}
