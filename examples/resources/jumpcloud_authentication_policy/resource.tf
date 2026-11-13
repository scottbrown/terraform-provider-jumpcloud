# Require MFA for user portal logins from outside the office IP list.
resource "jumpcloud_authentication_policy" "portal_mfa" {
  name     = "User Portal MFA Outside Office"
  type     = "user_portal"
  disabled = false

  conditions = {
    ip_address_not_in = [jumpcloud_ip_list.office.id]
  }

  effect = {
    action              = "allow"
    mfa_required        = true
    mfa_allowed_factors = ["totp", "webauthn", "push"]
  }
}

# Deny access to an application from untrusted devices.
resource "jumpcloud_authentication_policy" "app_trusted_devices" {
  name = "Finance App Trusted Devices Only"
  type = "resource"

  conditions = {
    device_trusted = false
  }

  effect = {
    action = "deny"
  }

  targets = [
    {
      type = "application"
      id   = jumpcloud_application.finance.id
    },
  ]
}
