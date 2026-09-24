resource "jumpcloud_user_group" "example" {
  name              = "Engineering"
  description       = "Engineering team members"
  membership_method = "STATIC"
}
