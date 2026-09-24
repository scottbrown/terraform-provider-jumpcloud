resource "jumpcloud_policy_group_membership" "example" {
  group_id  = jumpcloud_policy_group.compliance.id
  policy_id = jumpcloud_policy.filevault.id
}
