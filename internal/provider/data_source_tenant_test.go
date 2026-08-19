package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const tenantDataSourceName = "data.nautobot_tenant.test"

func testAccTenantDataSourceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name = "%[1]s"
}

data "nautobot_tenant" "test" {
  name = nautobot_tenant.test.name
}
`, name)
}

func testAccTenantDataSourceConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
  comments    = "acceptance test comment"
}

data "nautobot_tenant" "test" {
  name = nautobot_tenant.test.name
}
`, name)
}

func TestAccTenantDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tenant-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(tenantDataSourceName, "comments", ""),
					resource.TestCheckResourceAttr(tenantDataSourceName, "tenant_group_id", ""),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantDataSourceName, "id",
						"nautobot_tenant.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tenant-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tenantDataSourceName, "name", name),
					resource.TestCheckResourceAttr(tenantDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(tenantDataSourceName, "comments", "acceptance test comment"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						tenantDataSourceName, "id",
						"nautobot_tenant.test", "id",
					),

					resource.TestCheckResourceAttrSet(tenantDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "notes_url"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(tenantDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccTenantDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-tenant-notfound-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "nautobot_tenant" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Tenant not found`),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
