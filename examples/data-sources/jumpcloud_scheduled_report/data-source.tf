# Look up by display name
data "jumpcloud_scheduled_report" "compliance" {
  display_name = "Weekly Compliance Report"
}

# Look up by ID
data "jumpcloud_scheduled_report" "by_id" {
  id = "5f0c1234abcdef567890abcd"
}
