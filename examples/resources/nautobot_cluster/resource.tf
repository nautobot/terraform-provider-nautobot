// This fetches the cluster type with the given name
data "nautobot_cluster_type" "example" {
  name = "My Cluster Type Name"
}

// So we can create a cluster of it's type
resource "nautobot_cluster" "new" {
  name            = "My New Cluster"
  comments        = "This cluster was created using Terraform."
  cluster_type_id = nautobot_cluster_type.example
}

output "resource_cluster_new" {
  value = nautobot_cluster.new
}
