# No VM's exist per default
data "nautobot_virtual_machine" "example" {
  depends_on = [nautobot_vm_primary_ip.new]
  name = nautobot_virtual_machine.new.name
}

output "data_virtual_machine" {
  value = data.nautobot_virtual_machine.example
}
