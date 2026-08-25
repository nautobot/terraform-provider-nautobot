package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	prefixesDataSourceName = "data.nautobot_prefixes.test"
)

func testAccPrefixesDataSourceConfigBasic(base string, p1 string, p2 string, p3 string) string {

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_prefix" "p1" {
  prefix  = "%[2]s"
  status  = "%[5]s"
}

resource "nautobot_prefix" "p2" {
  prefix  = "%[3]s"
  status  = "%[5]s"
}

resource "nautobot_prefix" "p3" {
  prefix  = "%[4]s"
  status  = "%[5]s"
}

data "nautobot_prefixes" "test" {
  depends_on = [
    nautobot_prefix.p1,
    nautobot_prefix.p2,
    nautobot_prefix.p3,
  ]
}
`, base, p1, p2, p3, testStatus)
}

func testAccPrefixesDataSourceConfigFull(base string, vid int, p1 string, p2 string, p3 string) string {

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[7]d
  status = "%[6]s"
}

# minimal prefix
resource "nautobot_prefix" "p1" {
  prefix = "%[2]s"
  status = "%[6]s"
}

# full prefix A
resource "nautobot_prefix" "p2" {
  prefix      = "%[3]s"
  status      = "%[6]s"
  description = "p2 created by terraform acceptance test"
  tenant_id   = "%[5]s"
  vlan_id     = nautobot_vlan.v.id
}

# full prefix B (different values)
resource "nautobot_prefix" "p3" {
  prefix      = "%[4]s"
  status      = "%[6]s"
  description = "p3 created by terraform acceptance test"
  tenant_id   = "%[5]s"
  vlan_id     = nautobot_vlan.v.id
}

data "nautobot_prefixes" "test" {
  depends_on = [
    nautobot_prefix.p1,
    nautobot_prefix.p2,
    nautobot_prefix.p3,
  ]
}
`, base, p1, p2, p3, "", testStatus, vid)
}

func TestAccPrefixesDataSource_basic(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-prefixes-basic-%d", time.Now().Unix())
	seed := testAccSeedForTest(t)

	p1 := testAccPrefixCIDR(seed, 24)
	p2 := testAccPrefixCIDR(seed, 25)
	p3 := testAccPrefixCIDR(seed, 26)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixesDataSourceConfigBasic(base, p1, p2, p3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckPrefixesCountAtLeast(prefixesDataSourceName, 3),

					testFindPrefixIndexByCIDR(prefixesDataSourceName, p1),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p2),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p3),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p1, map[string]string{
						"prefix":      p1,
						"description": "",
						"tenant_id":   "",
						"role_id":     "",
						"parent_id":   "",
						"rir_id":      "",
						"vlan_id":     "",
						"tags_ids.#":  "0",
					}),
					testCheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p1, "date_allocated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixesDataSource_full(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	base := fmt.Sprintf("tfacc-ds-prefixes-full-%d", seed)
	vid := testAccVLANVid(seed, 5)

	p1 := testAccPrefixCIDR(seed, 26)
	p2 := testAccPrefixCIDR(seed, 27)
	p3 := testAccPrefixCIDR(seed, 28)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixesDataSourceConfigFull(base, vid, p1, p2, p3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckPrefixesCountAtLeast(prefixesDataSourceName, 3),

					testFindPrefixIndexByCIDR(prefixesDataSourceName, p1),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p2),
					testFindPrefixIndexByCIDR(prefixesDataSourceName, p3),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p1, map[string]string{
						"prefix":      p1,
						"description": "",
						"tenant_id":   "",
						"vlan_id":     "",
						"tags_ids.#":  "0",
					}),
					testCheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p1, "date_allocated"),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p2, map[string]string{
						"prefix":      p2,
						"description": "p2 created by terraform acceptance test",
						"tenant_id":   "",
						"tags_ids.#":  "0",
					}),
					testCheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p2, "date_allocated"),
					testCheckPrefixInListAttrEqualsResourceAttr(prefixesDataSourceName, p2, "vlan_id", "nautobot_vlan.v", "id"),

					testCheckPrefixInListHasAttrs(prefixesDataSourceName, p3, map[string]string{
						"prefix":      p3,
						"description": "p3 created by terraform acceptance test",
						"tenant_id":   "",
						"tags_ids.#":  "0",
					}),
					testCheckListItemAttrNull(prefixesDataSourceName, "prefixes", "prefix", p3, "date_allocated"),
					testCheckPrefixInListAttrEqualsResourceAttr(prefixesDataSourceName, p3, "vlan_id", "nautobot_vlan.v", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
