# Static system group
resource "jumpcloud_system_group" "production" {
  name        = "Production Servers"
  description = "All production systems"
}

# Dynamic system group (auto-populated based on OS)
resource "jumpcloud_system_group" "mac_devices" {
  name              = "Mac Devices"
  description       = "Automatically includes all macOS systems"
  membership_method = "DYNAMIC_AUTOMATED"

  member_query = jsonencode({
    queryType = "FilterQuery"
    filters = [
      {
        field    = "os"
        operator = "eq"
        value    = "Mac OS X"
      }
    ]
  })
}
