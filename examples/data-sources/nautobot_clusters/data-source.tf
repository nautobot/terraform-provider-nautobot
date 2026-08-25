// This fetches all known clusters
data "nautobot_clusters" "example" {}

// And we can filter later on
output "data_clusters_example" {
  value = data.nautobot_clusters.example.clusters[0]
}
