# No VM exists per default
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

output "data_graphql" {
  value = data.nautobot_graphql.nodes
}
output "data_graphql_example" {
  value = jsondecode(data.nautobot_graphql.nodes.data).virtual_machines
}
