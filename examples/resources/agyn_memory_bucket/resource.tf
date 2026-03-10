resource "agyn_memory_bucket" "example" {
  title             = "My Memory Bucket"
  description       = "An example memory bucket managed by Terraform."
  scope             = "global"
  collection_prefix = "example"
}
