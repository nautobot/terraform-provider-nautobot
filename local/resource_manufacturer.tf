resource "nautobot_manufacturer" "new" {
  description = "Created with Terraform"
  name        = "New Vendor"
}

output "resource_manufacturer" {
  value = nautobot_manufacturer.new
}
