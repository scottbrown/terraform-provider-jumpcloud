# Webhook notification channel
resource "jumpcloud_notification_channel" "prod_webhook" {
  name        = "Production Alerts"
  description = "Webhook for production alert notifications"
  type        = "CHANNEL_TYPE_WEBHOOK"
  enabled     = true

  configuration = jsonencode({
    webhook = {
      url             = "https://hooks.example.com/jumpcloud"
      authType        = "WEBHOOK_AUTH_TYPE_BEARER"
      authToken       = "my-secret-token"
      sslVerification = true
      customHeaders = [
        {
          key   = "X-Source"
          value = "jumpcloud"
        }
      ]
    }
  })
}

# Slack notification channel
resource "jumpcloud_notification_channel" "slack_alerts" {
  name    = "Security Slack"
  type    = "CHANNEL_TYPE_SLACK"
  enabled = true

  configuration = jsonencode({
    slack = {
      slackChannel = [
        {
          slackChannelId     = "C0123456789"
          slackTeamId        = "T0123456789"
          slackWorkspaceName = "mycompany"
        }
      ]
    }
  })
}
