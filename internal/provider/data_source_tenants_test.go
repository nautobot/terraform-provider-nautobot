package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const tenantsDataSourceName = "data.nautobot_tenants.test"

func testAccTenantsDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "t1" {
  name = "%[1]s-1"
}

resource "nautobot_tenant" "t2" {
  name        = "%[1]s-2"
  description = "t2 created by terraform acceptance test"
}

resource "nautobot_tenant" "t3" {
  name        = "%[1]s-3"
  description = "t3 created by terraform acceptance test"
  comments    = "t3 comment"
}

data "nautobot_tenants" "test" {
  depends_on = [
    nautobot_tenant.t1,
    nautobot_tenant.t2,
    nautobot_tenant.t3,
  ]
}
`, base)
}

func TestAccTenantsDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-tenants-%d", time.Now().Unix())

	t1 := base + "-1"
	t2 := base + "-2"
	t3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccTenantsDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckTenantsCountAtLeast(tenantsDataSourceName, 3),

					testFindTenantIndexByName(tenantsDataSourceName, t1),
					testFindTenantIndexByName(tenantsDataSourceName, t2),
					testFindTenantIndexByName(tenantsDataSourceName, t3),

					testCheckTenantInListHasAttrs(tenantsDataSourceName, t1, map[string]string{
						"name":        t1,
						"description": "",
						"comments":    "",
					}),
					testCheckTenantInListHasAttrs(tenantsDataSourceName, t2, map[string]string{
						"name":        t2,
						"description": "t2 created by terraform acceptance test",
					}),
					testCheckTenantInListHasAttrs(tenantsDataSourceName, t3, map[string]string{
						"name":        t3,
						"description": "t3 created by terraform acceptance test",
						"comments":    "t3 comment",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
