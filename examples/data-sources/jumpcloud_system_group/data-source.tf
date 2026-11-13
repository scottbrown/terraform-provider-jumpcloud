# Look up by name
data "jumpcloud_system_group" "production" {
  name = "Production Servers"
}

# Look up by ID
data "jumpcloud_system_group" "by_id" {
  id = "5f0c1234abcdef567890abcd"
}
