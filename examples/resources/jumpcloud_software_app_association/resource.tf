resource "jumpcloud_software_app_association" "chrome_to_all_systems" {
  software_app_id = jumpcloud_software_app.chrome.id
  target_type     = "system_group"
  target_id       = "5f0c1234abcdef567890abcd"
}
