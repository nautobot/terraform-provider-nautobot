resource "nautobot_tenant_group" "new" {
  name        = "New Tenant Group"
  description = "Created with Terraform"
}

output "resource_tenant_group" {
  value = nautobot_tenant_group.new
}
