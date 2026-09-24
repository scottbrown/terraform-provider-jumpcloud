resource "jumpcloud_software_app" "chrome" {
  display_name = "Google Chrome"

  settings = jsonencode({
    packageId = "com.google.chrome"
  })
}
