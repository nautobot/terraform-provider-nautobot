data "nautobot_virtual_machine" "example" {
  name = "My VM Name"
}

output "data_virtual_machine" {
  value = data.nautobot_virtual_machine.example
}
