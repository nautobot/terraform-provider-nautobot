data "nautobot_tenant_group" "example" {
  name = "My Tenant Group Name"
}

output "data_tenant_group" {
  value = data.nautobot_tenant_group.example
}
