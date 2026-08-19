data "nautobot_namespaces" "all" {}

output "namespaces" {
  value = data.nautobot_namespaces.all.namespaces
}
