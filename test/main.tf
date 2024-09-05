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

resource "nautobot_cluster" "new" {
  name            = "My New Cluster"
  comments        = "This cluster was created using Terraform."
  cluster_type_id = nautobot_cluster_type.new.id 

  # Optionally add cluster group, tenant, location, etc.
#  cluster_group_id   = "your-cluster-group-id"
#  tenant_id          = data.nautobot_tenant.example.id  # Referencing tenant data source
#  location_id        = "your-location-id"
#  tags_id            = ["tag1", "tag2"]

  custom_fields = {
    "custom_field_1" = "value1"
    "custom_field_2" = "value2"
  }
}