// This fetches all known VMs
data "nautobot_virtual_machines" "example" {}

// And we can filter later on
output "data_virtual_machines_example" {
  value = data.nautobot_virtual_machines.example.virtual_machines[0]
}
