resource "nautobot_cluster_type" "new" {
  name        = "My New Cluster Type"
  description = "This is a cluster type created via Terraform"
}

output "resource_cluster_type_new" {
  value = nautobot_cluster_type.new
}