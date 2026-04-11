resource "agyn_organization" "example" {
  name = "example-org"
}

resource "agyn_runner" "example" {
  name            = "example-runner"
  organization_id = agyn_organization.example.id
  labels = {
    region = "us-east-1"
  }
}
