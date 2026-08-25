data "nautobot_vlan" "example" {
  name = "VOICE - wna01-asw-04"
}

output "data_vlan" {
  value = data.nautobot_vlan.example
}
