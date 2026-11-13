resource "jumpcloud_policy" "example" {
  name        = "FileVault Enforcement"
  template_id = "policy-template-id"
  state       = "active"
  notes       = "Enforces FileVault encryption on all macOS devices"
  values = jsonencode({
    enabled = true
  })
}
