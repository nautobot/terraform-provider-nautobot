terraform {
  required_providers {
    nautobot = {
      version = "3.0.2"
      source  = "registry.terraform.io/nautobot/nautobot"
    }
  }
}

provider "nautobot" {
  url                            = "https://my.instance.com/api"
  token                          = "MyAPIToken0000abcdefghijklmnopqrstuvwxyz"
  status_request_timeout_seconds = 10
}
