data "nautobot_clusters" "example" {
  depends_on = [nautobot_cluster.new]
}

output "data_clusters_example" {
  value = data.nautobot_clusters.example.clusters[0]
}
