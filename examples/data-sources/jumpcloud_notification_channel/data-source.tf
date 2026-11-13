# Look up by name
data "jumpcloud_notification_channel" "prod_webhook" {
  name = "Production Alerts"
}

# Look up by ID
data "jumpcloud_notification_channel" "by_id" {
  id = "5f0c1234abcdef567890abcd"
}
