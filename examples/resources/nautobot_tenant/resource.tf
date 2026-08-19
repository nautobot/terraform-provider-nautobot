resource "nautobot_tenant" "new" {
  name        = "New Tenant"
  description = "Created with Terraform"
}

output "resource_tenant" {
  value = nautobot_tenant.new
}
