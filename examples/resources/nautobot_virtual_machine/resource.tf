// This fetches the cluster with the given name
data "nautobot_cluster" "example" {
  name = "My Cluster Name"
}

// So we can create a virtual machine belonging to it
resource "nautobot_virtual_machine" "new" {
  name            = "My New VM"
  cluster_id      = nautobot_cluster.example.id
  status          = "Active"
  vcpus           = 4
  memory          = 8192
  disk            = 100
  comments        = "This virtual machine was created using Terraform."
}

output "resource_virtual_machine" {
  value = nautobot_virtual_machine.new
}
