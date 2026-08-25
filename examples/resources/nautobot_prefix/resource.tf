// This fetches the VLAN with the given name
data "nautobot_vlan" "example" {
  name = "My VLAN Name"
}

// So we can create a prefix belonging to it
resource "nautobot_prefix" "new" {
  prefix  = "10.124.22.0/24"
  status  = "Active"
  vlan_id = data.nautobot_vlan.example.id
}

output "resource_prefix" {
  value = nautobot_prefix.new
}
