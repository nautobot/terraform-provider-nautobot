// This fetches all known cluster types
data "nautobot_cluster_types" "example" {}

// And we can filter later on
output "data_cluster_types_example" {
  value = data.nautobot_cluster_types.example.cluster_types[0]
}
