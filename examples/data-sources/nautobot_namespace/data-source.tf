data "nautobot_namespace" "global" {
  name = "Global"
}

output "global_namespace_id" {
  value = data.nautobot_namespace.global.id
}
