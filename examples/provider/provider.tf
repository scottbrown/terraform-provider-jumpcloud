terraform {
  required_providers {
    jumpcloud = {
      source  = "scottbrown/jumpcloud"
      version = "~> 0.1"
    }
  }
}

provider "jumpcloud" {
  # api_key sourced from JUMPCLOUD_API_KEY env var
  # org_id  sourced from JUMPCLOUD_ORG_ID env var
}
