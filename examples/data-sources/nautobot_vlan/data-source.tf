data "nautobot_vlan" "example" {
  name = "My VLAN Name"
}

output "data_vlan" {
  value = data.nautobot_vlan.example
}
