resource "agyn_memory_bucket" "example" {
  title       = "My Memory Bucket"
  description = "An example memory bucket managed by Terraform."
  config = jsonencode({
    provider = "pinecone"
  })
}
