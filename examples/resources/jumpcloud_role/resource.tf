# Read-only auditor role scoped to the whole organization.
resource "jumpcloud_role" "auditor" {
  name        = "Security Auditor"
  description = "Read-only access for compliance reviews"

  permissions = [
    { resource_type = "users", action = "readonly" },
    { resource_type = "systems", action = "readonly" },
    { resource_type = "policies", action = "readonly" },
  ]

  scope = {
    type = "all"
  }
}
