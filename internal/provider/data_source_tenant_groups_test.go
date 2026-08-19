package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const tenantGroupsDataSourceName = "data.nautobot_tenant_groups.test"

func testAccTenantGroupsDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant_group" "g1" {
  name = "%[1]s-1"
}

resource "nautobot_tenant_group" "g2" {
  name        = "%[1]s-2"
  description = "g2 created by terraform acceptance test"
}

resource "nautobot_tenant_group" "g3" {
  name        = "%[1]s-3"
  description = "g3 created by terraform acceptance test"
}

data "nautobot_tenant_groups" "test" {
  depends_on = [
    nautobot_tenant_group.g1,
    nautobot_tenant_group.g2,
    nautobot_tenant_group.g3,
  ]
}
`, base)
}

func TestAccTenantGroupsDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-tgs-%d", time.Now().Unix())

	g1 := base + "-1"
	g2 := base + "-2"
	g3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantGroupsDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckTenantGroupsCountAtLeast(tenantGroupsDataSourceName, 3),

					testFindTenantGroupIndexByName(tenantGroupsDataSourceName, g1),
					testFindTenantGroupIndexByName(tenantGroupsDataSourceName, g2),
					testFindTenantGroupIndexByName(tenantGroupsDataSourceName, g3),

					testCheckTenantGroupInListHasAttrs(tenantGroupsDataSourceName, g1, map[string]string{
						"name":        g1,
						"description": "",
					}),
					testCheckTenantGroupInListHasAttrs(tenantGroupsDataSourceName, g2, map[string]string{
						"name":        g2,
						"description": "g2 created by terraform acceptance test",
					}),
					testCheckTenantGroupInListHasAttrs(tenantGroupsDataSourceName, g3, map[string]string{
						"name":        g3,
						"description": "g3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
