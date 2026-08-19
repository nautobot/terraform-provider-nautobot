data "nautobot_namespace" "global" {
  name = "Global"
}

// Fetch a prefix by its exact CIDR and namespace
data "nautobot_prefix" "example" {
  prefix       = "10.151.0.0/16"
  namespace_id = data.nautobot_namespace.global.id
}

// And finally get the first available IP address from that prefix
resource "nautobot_available_ip_address" "example" {
  prefix_id = data.nautobot_prefix.example.id
  status    = "Active"
  dns_name  = "test-vm.test.com"
}

// This fetches the virtual machine with the given name
data "nautobot_virtual_machine" "example" {
  name = "My VM Name"
}

// So we can use the virtual machine and IP to create an interface
resource "nautobot_vm_interface" "new" {
  name = "eth0"
  virtual_machine_id = nautobot_virtual_machine.example.id
  status = "Active"
  ip_addresses = [
    nautobot_available_ip_address.example.id
  ]
}

output "resource_vm_interface" {
  value = nautobot_vm_interface.new
}
