data "nautobot_cluster" "example" {
  depends_on = [nautobot_cluster.new]
  name = nautobot_cluster.new.name
}

output "data_cluster" {
  value = data.nautobot_cluster.example
}
