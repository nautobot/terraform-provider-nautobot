// This fetches all known VLANs
data "nautobot_vlans" "example" {}

// And we can filter later on
output "data_vlans_example" {
  value = data.nautobot_vlans.example.vlans[0]
}
