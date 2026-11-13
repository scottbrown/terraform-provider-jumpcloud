resource "jumpcloud_user_group_membership" "example" {
  group_id = jumpcloud_user_group.engineering.id
  user_id  = "64a1b2c3d4e5f6a7b8c9d0e1"
}
