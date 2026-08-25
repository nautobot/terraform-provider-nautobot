data "nautobot_namespace" "global" {
  name = "Global"
}

// Fetch a prefix by its exact CIDR and namespace
data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

// And finally get the first available IP address from that prefix
data "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
}

output "data_available_ip_address" {
  value = data.nautobot_available_ip_address.example
}
