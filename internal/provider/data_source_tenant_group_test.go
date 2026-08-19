package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const tenantGroupDataSourceName = "data.nautobot_tenant_group.test"

func testAccTenantGroupDataSourceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name = "%[1]s"
}

data "nautobot_tenant_group" "test" {
  name = nautobot_tenant_group.test.name
}
`, name)
}

func testAccTenantGroupDataSourceConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
}

data "nautobot_tenant_group" "test" {
  name = nautobot_tenant_group.test.name
}
`, name)
}

func TestAccTenantGroupDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tg-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "parent_id", ""),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantGroupDataSourceName, "id",
						"nautobot_tenant_group.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tg-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantGroupDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantGroupDataSourceName, "id",
						"nautobot_tenant_group.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantGroupDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantGroupDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tg-notfound-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "nautobot_tenant_group" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Tenant group not found`),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
