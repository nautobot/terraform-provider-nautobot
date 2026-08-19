// This fetches all known manufacturers
data "nautobot_manufacturers" "all" {}

// And we can filter later on
output "data_manufacturers_example" {
  value = {
    for manufacturer in data.nautobot_manufacturers.all.manufacturers :
    manufacturer.id => manufacturer
    if manufacturer.name == "My Manufaturer Name"
  }
}