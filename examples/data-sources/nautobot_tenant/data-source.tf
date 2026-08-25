data "nautobot_tenant" "example" {
  name = "My Tenant Name"
}

output "data_tenant" {
  value = data.nautobot_tenant.example
}
