// We can for example fetch all VMs with name and ID
data "nautobot_graphql" "nodes" {
  depends_on = [nautobot_virtual_machine.new]
  query = <<EOF
query {
  virtual_machines {
      name
      id
  }
}
EOF
}

// And output everything
output "data_graphql" {
  value = data.nautobot_graphql.nodes
}

// Or filter the output
output "data_graphql_example" {
  value = jsondecode(data.nautobot_graphql.nodes.data).virtual_machines
}
