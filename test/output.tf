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

data "nautobot_cluster_type" "example" {
  depends_on = [nautobot_cluster_type.new]
}

output "cluster_type_details" {
  value = data.nautobot_cluster_type.example.cluster_types[0]
}

output "cluster_type_id" {
  value = data.nautobot_cluster_type.example.cluster_types[0].id
}