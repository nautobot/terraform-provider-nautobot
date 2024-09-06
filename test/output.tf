data "nautobot_manufacturers" "all" {
  depends_on = [nautobot_manufacturer.new]
}

variable "manufacturer_name" {
  type    = string
  default = "New Vendor"
}

# Only returns te created manufacturer
output "data_source_example" {
  value = {
    for manufacturer in data.nautobot_manufacturers.all.manufacturers :
    manufacturer.id => manufacturer
    if manufacturer.name == var.manufacturer_name
  }
}
data "nautobot_graphql" "nodes" {
  query = <<EOF
query {
  virtual_machines {
      name
      id
  }
  devices {
    name
    id
  }
}
EOF
}

output "data_source_graphql" {
  value = data.nautobot_graphql.nodes
}
output "data_source_graphql_vm" {
  value = jsondecode(data.nautobot_graphql.nodes.data).virtual_machines
}

data "nautobot_cluster_types" "example" {
  depends_on = [nautobot_cluster_type.new]
}

output "cluster_types_details" {
  value = data.nautobot_cluster_types.example.cluster_types[0]
}

output "cluster_types_id" {
  value = data.nautobot_cluster_types.example.cluster_types[0].id
}

data "nautobot_clusters" "example" {
  depends_on = [nautobot_cluster.new]
}

output "clusters_details" {
  value = data.nautobot_clusters.example.clusters[0]
}

output "clusters_id" {
  value = data.nautobot_clusters.example.clusters[0].id
}

data "nautobot_virtual_machines" "example" {
  depends_on = [nautobot_virtual_machine.new]
}

output "vms_details" {
  value = data.nautobot_virtual_machines.example.virtual_machines[0]
}

output "vms_id" {
  value = data.nautobot_virtual_machines.example.virtual_machines[0].id
}

data "nautobot_manufacturer" "example" {
  depends_on = [nautobot_manufacturer.new]
  name = nautobot_manufacturer.new.name
}

output "manufacturer_details" {
  value = data.nautobot_manufacturer.example
}

output "manufacturer_id" {
  value = data.nautobot_manufacturer.example.id
}

data "nautobot_cluster_type" "example" {
  depends_on = [nautobot_cluster_type.new]
  name = nautobot_cluster_type.new.name
}

output "cluster_type_details" {
  value = data.nautobot_cluster_type.example
}

output "cluster_type_id" {
  value = data.nautobot_cluster_type.example.id
}

data "nautobot_cluster" "example" {
  depends_on = [nautobot_cluster.new]
  name = nautobot_cluster.new.name
}

output "cluster_details" {
  value = data.nautobot_cluster.example
}

output "cluster_id" {
  value = data.nautobot_cluster.example.id
}

data "nautobot_virtual_machine" "example" {
  depends_on = [nautobot_virtual_machine.new]
  name = nautobot_virtual_machine.new.name
}

output "vm_details" {
  value = data.nautobot_virtual_machine.example
}

output "vm_id" {
  value = data.nautobot_virtual_machine.example.id
}