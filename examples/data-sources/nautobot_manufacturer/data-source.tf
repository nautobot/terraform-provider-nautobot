data "nautobot_manufacturer" "example" {
  name = "My Manufacturer Name"
}

output "data_manufacturer" {
  value = data.nautobot_manufacturer.example
}
