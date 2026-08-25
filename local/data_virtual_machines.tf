# No VM's exist per default
data "nautobot_virtual_machines" "example" {
  depends_on = [nautobot_vm_primary_ip.new]
}

output "data_virtual_machines_example" {
  value = data.nautobot_virtual_machines.example.virtual_machines[0]
}
