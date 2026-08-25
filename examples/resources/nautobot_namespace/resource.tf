resource "nautobot_namespace" "example" {
  name        = "Production"
  description = "Production IP address space"
}

output "namespace_id" {
  value = nautobot_namespace.example.id
}
