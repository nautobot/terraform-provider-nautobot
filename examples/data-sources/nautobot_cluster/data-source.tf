data "nautobot_cluster" "example" {
  name = "My Cluster Name"
}

output "data_cluster" {
  value = data.nautobot_cluster.example
}
