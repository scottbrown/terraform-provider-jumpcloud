resource "jumpcloud_system_group_membership" "server1" {
  group_id  = jumpcloud_system_group.production.id
  system_id = "5f0c1234abcdef567890abcd"
}
