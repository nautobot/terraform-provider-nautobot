data "nautobot_manufacturer" "example" {
  depends_on = [nautobot_manufacturer.new]
  name = nautobot_manufacturer.new.name
}

output "data_manufacturer" {
  value = data.nautobot_manufacturer.example
}
