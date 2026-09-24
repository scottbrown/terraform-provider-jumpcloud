resource "jumpcloud_command" "example" {
  name         = "Security Posture Check"
  command      = <<-EOT
    #!/bin/bash
    echo "Checking security posture..."
  EOT
  command_type = "linux"
  user         = "root"
  sudo         = true
  shell        = "bash"
  trigger      = "periodic"
  schedule     = "0 9 * * 1"
}
