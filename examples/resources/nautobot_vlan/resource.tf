resource "nautobot_vlan" "new" {
  name   = "My New VLAN"
  vid    = 1234
  status = "Active"
}

output "resource_vlan" {
  value = nautobot_vlan.new
}
