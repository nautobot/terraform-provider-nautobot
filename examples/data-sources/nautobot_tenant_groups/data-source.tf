// This fetches all known tenant groups
data "nautobot_tenant_groups" "all" {}

output "data_tenant_groups_example" {
  value = data.nautobot_tenant_groups.all.tenant_groups[0]
}
