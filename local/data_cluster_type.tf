data "nautobot_cluster_type" "example" {
  depends_on = [nautobot_cluster_type.new]
  name = nautobot_cluster_type.new.name
}

output "data_cluster_type" {
  value = data.nautobot_cluster_type.example
}
