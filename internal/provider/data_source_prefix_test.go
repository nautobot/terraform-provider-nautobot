package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	prefixDataSourceName = "data.nautobot_prefix.test"
)

func testAccPrefixDataSourceConfigByID(prefixCIDR string, baseName string, baseVid int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[2]s-vlan"
  vid    = %[3]d
  status = "%[4]s"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[1]s"
  status      = "%[4]s"
  vlan_id     = nautobot_vlan.v.id
  description = "created by terraform acceptance test"
  tenant_id   = "%[5]s"
}

data "nautobot_prefix" "test" {
  id = nautobot_prefix.test.id
}
`, prefixCIDR, baseName, baseVid, testStatus, "")
}

func testAccPrefixDataSourceConfigByPrefix(prefixCIDR string, baseName string, baseVid int) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[2]s-vlan"
  vid    = %[3]d
  status = "%[4]s"
}

resource "nautobot_prefix" "test" {
  prefix    = "%[1]s"
  status    = "%[4]s"
  vlan_id   = nautobot_vlan.v.id
  tenant_id = "%[5]s"
}

data "nautobot_namespace" "global" {
  name = "Global"
}

data "nautobot_prefix" "test" {
  depends_on   = [nautobot_prefix.test]
  prefix       = "%[1]s"
  namespace_id = data.nautobot_namespace.global.id
}
`, prefixCIDR, baseName, baseVid, testStatus, "")
}

func TestAccPrefixDataSource_byID(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	baseName := fmt.Sprintf("tfacc-ds-prefix-byid-%d", seed)
	baseVid := testAccVLANVid(seed, 3)
	prefixCIDR := testAccPrefixCIDR(seed, 22)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixDataSourceConfigByID(prefixCIDR, baseName, baseVid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "prefix"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "last_updated"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "ip_version"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "notes_url"),

					resource.TestCheckResourceAttr(prefixDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixDataSourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "rir_id", ""),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "namespace_id"),
					resource.TestCheckResourceAttrPair(prefixDataSourceName, "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckNoResourceAttr(prefixDataSourceName, "date_allocated"),
					resource.TestCheckResourceAttr(prefixDataSourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixDataSource_byPrefix(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	baseName := fmt.Sprintf("tfacc-ds-prefix-byprefix-%d", seed)
	baseVid := testAccVLANVid(seed, 4)
	prefixCIDR := testAccPrefixCIDR(seed, 23)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixDataSourceConfigByPrefix(prefixCIDR, baseName, baseVid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(prefixDataSourceName, "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(prefixDataSourceName, "prefix", prefixCIDR),
					resource.TestCheckResourceAttrPair(prefixDataSourceName, "namespace_id", "nautobot_prefix.test", "namespace_id"),

					resource.TestCheckResourceAttr(prefixDataSourceName, "tenant_id", ""),

					resource.TestCheckResourceAttr(prefixDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixDataSourceName, "rir_id", ""),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "namespace_id"),
					resource.TestCheckNoResourceAttr(prefixDataSourceName, "date_allocated"),
					resource.TestCheckResourceAttr(prefixDataSourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "last_updated"),

					resource.TestCheckResourceAttrSet(prefixDataSourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixDataSourceName, "ip_version"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
