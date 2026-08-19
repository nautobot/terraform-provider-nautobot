data "nautobot_prefixes" "example" {
}

output "data_prefixes_example" {
  value = data.nautobot_prefixes.example.prefixes[0]
}
