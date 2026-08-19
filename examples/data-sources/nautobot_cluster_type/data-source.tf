data "nautobot_cluster_type" "example" {
  name = "My Cluster Type Name"
}

output "data_cluster_type" {
  value = data.nautobot_cluster_type.example
}
