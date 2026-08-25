// This fetches all known prefixes
data "nautobot_prefixes" "example" {}

// And we can filter later on
output "data_prefixes_example" {
  value = data.nautobot_prefixes.example.prefixes[0]
}
