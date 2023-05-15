package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGraphQL(t *testing.T) {
	// https://github.com/hashicorp/terraform-plugin-sdk/issues/952
	t.Skip("resource not yet implemented, remove this once you add your own code")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGraphQL,
				Check: resource.ComposeTestCheckFunc(
					//resource.TestCheckResourceAttr("data.nautobot_manufacturers.list.manufacturers", "", ""),
					resource.TestCheckOutput("data_source_graphql", "virtual_machines"),
				),
			},
		},
	})
}

const testAccDataSourceGraphQL = `
provider "nautobot" {
	url = "https://demo.nautobot.com/api/"
	token = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
  }

data "nautobot_graphql" "vms" {
  query = <<EOF
query {
  virtual_machines {
    id
  }
}
EOF
}

output "data_source_graphql" {
  value = data.nautobot_graphql.vms
}
`
