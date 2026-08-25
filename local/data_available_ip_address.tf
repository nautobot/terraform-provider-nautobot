data "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
}

output "data_available_ip_address" {
  value = data.nautobot_available_ip_address.example
}
