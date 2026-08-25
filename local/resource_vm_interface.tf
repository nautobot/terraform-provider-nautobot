# No VM exist per default
resource "nautobot_vm_interface" "new" {
  name = "eth0"
  virtual_machine_id = nautobot_virtual_machine.new.id
  status = "Active"
  ip_addresses = [
    nautobot_available_ip_address.example.id
  ]
}

output "resource_vm_interface" {
  value = nautobot_vm_interface.new
}
