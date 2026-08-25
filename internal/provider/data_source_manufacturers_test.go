package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	manufacturersDataSourceName = "data.nautobot_manufacturers.test"
)

func testAccManufacturersDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "m1" {
  name = "%[1]s-1"
}

resource "nautobot_manufacturer" "m2" {
  name        = "%[1]s-2"
  description = "m2 created by terraform acceptance test"
}

resource "nautobot_manufacturer" "m3" {
  name        = "%[1]s-3"
  description = "m3 created by terraform acceptance test"
}

data "nautobot_manufacturers" "test" {
  depends_on = [
    nautobot_manufacturer.m1,
    nautobot_manufacturer.m2,
    nautobot_manufacturer.m3,
  ]
}
`, base)
}

func TestAccManufacturersDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-manufacturers-%d", time.Now().Unix())

	m1 := base + "-1"
	m2 := base + "-2"
	m3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturersDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckManufacturersCountAtLeast(manufacturersDataSourceName, 3),

					testFindManufacturerIndexByName(manufacturersDataSourceName, m1),
					testFindManufacturerIndexByName(manufacturersDataSourceName, m2),
					testFindManufacturerIndexByName(manufacturersDataSourceName, m3),

					testCheckManufacturerInListHasAttrs(manufacturersDataSourceName, m1, map[string]string{
						"name":        m1,
						"description": "",
					}),
					testCheckManufacturerInListHasAttrs(manufacturersDataSourceName, m2, map[string]string{
						"name":        m2,
						"description": "m2 created by terraform acceptance test",
					}),
					testCheckManufacturerInListHasAttrs(manufacturersDataSourceName, m3, map[string]string{
						"name":        m3,
						"description": "m3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
