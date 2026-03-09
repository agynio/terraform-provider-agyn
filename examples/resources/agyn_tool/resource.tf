resource "agyn_tool" "example" {
  name        = "my-tool"
  description = "An example tool managed by Terraform."
  type        = "http"
  config = jsonencode({
    url    = "https://api.example.com/action"
    method = "POST"
  })
}
