resource "nautobot_vm_primary_ip" "new" {
  depends_on = [nautobot_vm_interface.new]
  virtual_machine_id = nautobot_virtual_machine.new.id
  primary_ip4_id     = nautobot_available_ip_address.example.id
}

output "resource_primary_ip" {
  value = nautobot_vm_primary_ip.new
}
