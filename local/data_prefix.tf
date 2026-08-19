data "nautobot_prefix" "example" {
  depends_on = [data.nautobot_vlan.example]
  vlan_id = data.nautobot_vlan.example.id
}

output "data_prefix" {
  value = data.nautobot_prefix.example
}

###############################

data "nautobot_prefix" "example_parent" {
  depends_on = [data.nautobot_prefix.example]
  id = data.nautobot_prefix.example.parent_id
}

output "data_prefix_parent" {
  value = data.nautobot_prefix.example_parent
}