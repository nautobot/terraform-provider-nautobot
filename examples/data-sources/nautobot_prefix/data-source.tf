data "nautobot_namespace" "global" {
  name = "Global"
}

// Fetch a prefix by its exact CIDR and namespace
data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

output "data_prefix" {
  value = data.nautobot_prefix.example
}

// We can also get the parent Prefix
data "nautobot_prefix" "example_parent" {
  id = data.nautobot_prefix.example.parent_id
}

output "data_prefix_parent" {
  value = data.nautobot_prefix.example_parent
}
