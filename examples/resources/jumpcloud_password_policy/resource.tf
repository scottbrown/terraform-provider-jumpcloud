resource "jumpcloud_password_policy" "example" {
  name                             = "Strong Password Policy"
  min_length                       = 12
  max_length                       = 128
  needs_uppercase                  = true
  needs_lowercase                  = true
  needs_numeric                    = true
  needs_symbolic                   = true
  min_changes                      = 4
  password_expiration_in_days      = 90
  lockout_threshold                = 5
  lockout_duration_in_seconds      = 300
  days_before_expiration_to_notify = 14
  precedence                       = 1
}
