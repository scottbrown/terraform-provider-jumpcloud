resource "jumpcloud_scheduled_report" "weekly_compliance" {
  display_name       = "Weekly Compliance Report"
  description        = "Automated weekly compliance summary"
  schedule_frequency = "weekly"
  schedule_time      = "09:00 AM"
  time_zone          = "America/New_York"
  days               = ["MONDAY"]
  is_active          = true
  export_type        = "csv"
  report_id          = "5f0c1234abcdef567890abcd"
  report_type        = "report_view"

  recipients {
    channel_type = "email"
    value        = ["compliance@example.com", "security@example.com"]
  }

  recipients {
    channel_type = "slack"
    value        = ["#compliance-reports"]
  }
}
