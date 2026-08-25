package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	manufacturerDataSourceName = "data.nautobot_manufacturer.test"
)

func testAccManufacturerDataSourceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name = "%[1]s"
}

data "nautobot_manufacturer" "test" {
  name = nautobot_manufacturer.test.name
}
`, name)
}

func testAccManufacturerDataSourceConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
}

data "nautobot_manufacturer" "test" {
  name = nautobot_manufacturer.test.name
}
`, name)
}

func TestAccManufacturerDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-manufacturer-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						manufacturerDataSourceName, "id",
						"nautobot_manufacturer.test", "id",
					),

					resource.TestCheckResourceAttr(manufacturerDataSourceName, "description", ""),

					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-manufacturer-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						manufacturerDataSourceName, "id",
						"nautobot_manufacturer.test", "id",
					),

					resource.TestCheckResourceAttr(manufacturerDataSourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(manufacturerDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-manufacturer-notfound-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "nautobot_manufacturer" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Manufacturer not found`),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
