resource "jumpcloud_approval_flow_settings" "example" {
  admin_approval_email     = true
  expose_approval_progress = true
  resource_request         = true
  user_approval_email      = true
  user_request_email       = true
}
