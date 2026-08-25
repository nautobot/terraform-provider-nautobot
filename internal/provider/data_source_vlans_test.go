package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	VlansDataSourceName = "data.nautobot_vlans.test"
)

func testAccVLANsDataSourceConfigBasic(base string, baseVid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "vlan1" {
  name   = "%[1]s-1"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_vlan" "vlan2" {
  name   = "%[1]s-2"
  vid    = %[2]d + 1
  status = "%[3]s"
}

resource "nautobot_vlan" "vlan3" {
  name   = "%[1]s-3"
  vid    = %[2]d + 2
  status = "%[3]s"
}

data "nautobot_vlans" "test" {
  depends_on = [
    nautobot_vlan.vlan1,
    nautobot_vlan.vlan2,
    nautobot_vlan.vlan3,
  ]
}
`, base, baseVid, status)
}

func TestAccVLANsDataSource_basic(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	base := fmt.Sprintf("tfacc-ds-vlans-basic-%d", seed)
	baseVid := testAccVLANVid(seed, 8)

	v1 := base + "-1"
	v2 := base + "-2"
	v3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANsDataSourceConfigBasic(base, baseVid),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckVLANsCountAtLeast(VlansDataSourceName, 3),

					testFindVLANIndexByName(VlansDataSourceName, v1),
					testFindVLANIndexByName(VlansDataSourceName, v2),
					testFindVLANIndexByName(VlansDataSourceName, v3),

					testCheckVLANInListHasAttrs(VlansDataSourceName, v1, map[string]string{
						"name":          v1,
						"status":        testStatus,
						"tenant_id":     "",
						"role_id":       "",
						"vlan_group_id": "",
						"prefix_count":  "0",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
