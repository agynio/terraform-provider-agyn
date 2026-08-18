# Runners are infrastructure an environment points at, registered outside the
# config that uses them. Look one up by name rather than pasting its UUID.
data "agyn_runner" "default" {
  organization_id = var.organization_id
  name            = "default"
}
