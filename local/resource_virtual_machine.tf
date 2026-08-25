resource "nautobot_virtual_machine" "new" {
  name            = "Example VM"
  cluster_id      = nautobot_cluster.new.id
  status          = "Active"
  vcpus           = 4
  memory          = 8192
  disk            = 100
  comments        = "This virtual machine was created using Terraform."
}

output "resource_virtual_machine" {
  value = nautobot_virtual_machine.new
}
