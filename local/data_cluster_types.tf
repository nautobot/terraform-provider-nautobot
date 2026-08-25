data "nautobot_cluster_types" "example" {
  depends_on = [nautobot_cluster_type.new]
}

output "data_cluster_types_example" {
  value = data.nautobot_cluster_types.example.cluster_types[0]
}
