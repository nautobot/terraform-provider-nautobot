data "nautobot_vlans" "example" {
}

output "data_vlans_example" {
  value = data.nautobot_vlans.example.vlans[0]
}
