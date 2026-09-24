resource "jumpcloud_ip_list" "office" {
  name        = "Office Egress"
  description = "Public IP ranges for corporate offices"
  ips = [
    "203.0.113.0/24",
    "198.51.100.42",
  ]
}
