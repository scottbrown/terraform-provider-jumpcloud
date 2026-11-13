resource "jumpcloud_approval_flow" "example" {
  name          = "Request Production Access"
  description   = "Request temporary access to production systems"
  approval_type = "manual"
  resource_id   = jumpcloud_user_group.prod_access.id
  resource_type = "user_group"
  status        = "active"

  time_based_access  = true
  fixed_duration     = false
  non_admin_approval = false
  slack_enabled      = true

  visible_to = [jumpcloud_user_group.engineering.id]
}
