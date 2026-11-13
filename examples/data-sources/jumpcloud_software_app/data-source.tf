# Look up by display name
data "jumpcloud_software_app" "chrome" {
  display_name = "Google Chrome"
}

# Look up by ID
data "jumpcloud_software_app" "by_id" {
  id = "5f0c1234abcdef567890abcd"
}
