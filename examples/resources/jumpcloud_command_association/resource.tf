resource "jumpcloud_command_association" "example" {
  command_id  = jumpcloud_command.security_audit.id
  target_id   = "system-group-id"
  target_type = "system_group"
}
