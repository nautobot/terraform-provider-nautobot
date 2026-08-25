// This fetches all known tenants
data "nautobot_tenants" "all" {}

output "data_tenants_example" {
  value = data.nautobot_tenants.all.tenants[0]
}
