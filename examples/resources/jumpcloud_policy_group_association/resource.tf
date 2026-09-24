resource "jumpcloud_policy_group_association" "example" {
  group_id    = jumpcloud_policy_group.compliance.id
  target_id   = "system-group-id"
  target_type = "system_group"
}
