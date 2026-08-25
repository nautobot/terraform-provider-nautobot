# No cluster types exist per default
resource "nautobot_cluster" "new" {
  name            = "My New Cluster"
  comments        = "This cluster was created using Terraform."
  cluster_type_id = nautobot_cluster_type.new.id 

  # Optionally add cluster group, tenant, location, etc.
#  cluster_group_id   = "your-cluster-group-id"
#  tenant_id          = data.nautobot_tenant.example.id  # Referencing tenant data source
#  location_id        = "your-location-id"
#  tags_id            = ["tag1", "tag2"]
}

output "resource_cluster_new" {
  value = nautobot_cluster.new
}

resource "nautobot_cluster" "new2" {
  name            = "My New Cluster 2"
  comments        = "This cluster was created using Terraform."
  cluster_type_id = nautobot_cluster_type.new2.id

  # Optionally add cluster group, tenant, location, etc.
#  cluster_group_id   = "your-cluster-group-id"
#  tenant_id          = data.nautobot_tenant.example.id  # Referencing tenant data source
#  location_id        = "your-location-id"
#  tags_id            = ["tag1", "tag2"]
}

output "resource_cluster_new2" {
  value = nautobot_cluster.new2
}
