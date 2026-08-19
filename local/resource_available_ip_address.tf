resource "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
  status    = "Active"
  dns_name  = "test-vm.test.com"
}

output "resource_available_ip_address" {
  value = nautobot_available_ip_address.example
}
