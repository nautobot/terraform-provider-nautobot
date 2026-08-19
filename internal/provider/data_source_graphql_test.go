package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	graphQLDataSourceName = "data.nautobot_graphql.test"
)

func testAccGraphQLDataSourceConfigManufacturers() string {
	return testAccProviderConfig() + `
resource "nautobot_manufacturer" "test" {
  name = "Cisco"
}

data "nautobot_graphql" "test" {
  depends_on = [nautobot_manufacturer.test]
  query = <<-GQL
query {
  manufacturers {
    name
  }
}
GQL
}
`
}

func testCheckGraphQLDataContains(dsAddr, substr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		got := rs.Primary.Attributes["data"]
		if got == "" {
			return fmt.Errorf("%s: data is empty", dsAddr)
		}
		if !strings.Contains(got, substr) {
			return fmt.Errorf("%s: expected data to contain %q, got %q", dsAddr, substr, got)
		}
		return nil
	}
}

func TestAccGraphQLDataSource_manufacturersContainsCisco(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccGraphQLDataSourceConfigManufacturers(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(graphQLDataSourceName, "query"),
					resource.TestCheckResourceAttrSet(graphQLDataSourceName, "data"),
					testCheckGraphQLDataContains(graphQLDataSourceName, "Cisco"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
