# Grant a user group access to an application.
resource "jumpcloud_application_association" "finance_team" {
  application_id = jumpcloud_application.finance.id
  target_type    = "user_group"
  target_id      = jumpcloud_user_group.finance.id
}

# Grant a single user access to an application.
resource "jumpcloud_application_association" "contractor" {
  application_id = jumpcloud_application.finance.id
  target_type    = "user"
  target_id      = "64a1b2c3d4e5f6a7b8c9d0e1"
}
