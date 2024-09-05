terraform {
  required_providers {
    nautobot = {
      version = "0.0.1-beta"
      source  = "github.com/nautobot/nautobot"
    }
  }
}

provider "nautobot" {
  url   = "https://demo.nautobot.com/api"
  token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}

resource "nautobot_manufacturer" "new" {
  description = "Created with Terraform"
  name        = "New Vendor"
}

resource "nautobot_cluster_type" "new" {
  name        = "Example Cluster Type"
  description = "This is a cluster type created via Terraform"
}