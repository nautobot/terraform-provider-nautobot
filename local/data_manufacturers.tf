data "nautobot_manufacturers" "all" {
}

variable "manufacturer_name" {
  type    = string
  default = "Cisco"
}

# Only returns the manufacturer
output "data_manufacturers_example" {
  value = {
    for manufacturer in data.nautobot_manufacturers.all.manufacturers :
    manufacturer.id => manufacturer
    if manufacturer.name == var.manufacturer_name
  }
}