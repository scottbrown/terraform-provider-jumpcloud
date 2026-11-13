resource "jumpcloud_policy_association" "example" {
  policy_id   = jumpcloud_policy.filevault.id
  target_id   = "mac-system-group-id"
  target_type = "system_group"
}
