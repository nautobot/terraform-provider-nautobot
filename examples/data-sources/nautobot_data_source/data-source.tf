data "nautobot_manufacturers" "all" {}

data "nautobot_graphql" "vms" {
  query = <<EOF
query {
  virtual_machines {
    id
  }
}
EOF
}
