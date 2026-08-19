package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	prefixResourceName = "nautobot_prefix.test"
)

func testAccPrefixConfigMinimal(name string, vid int, cidr string) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigFull(name string, vid int, cidr string) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[4]s"
  status      = "%[3]s"
  vlan_id     = nautobot_vlan.v.id
  description = "created by terraform acceptance test"
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigUpdated(name string, vid int, cidr string) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "test" {
  prefix      = "%[4]s"
  status      = "%[3]s"
  vlan_id     = nautobot_vlan.v.id
  description = "updated by terraform acceptance test"
}
`, name, vid, status, cidr)
}

func testAccPrefixConfigParallel(baseName string, baseVid int, seed int64) string {
	status := testStatus
	c1 := testAccPrefixCIDR(seed, 8)
	c2 := testAccPrefixCIDR(seed, 9)
	c3 := testAccPrefixCIDR(seed, 10)

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "p1" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_prefix" "p2" {
  prefix  = "%[5]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_prefix" "p3" {
  prefix  = "%[6]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}
`, baseName, baseVid, status, c1, c2, c3)
}

func TestAccPrefixResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-minimal-%d", seed)
	vid := testAccVLANVid(seed, 14)
	cidr := testAccPrefixCIDR(seed, 11)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "status", testStatus),

					resource.TestCheckResourceAttrSet(prefixResourceName, "id"),
					resource.TestCheckResourceAttrPair(prefixResourceName, "vlan_id", "nautobot_vlan.v", "id"),

					resource.TestCheckResourceAttr(prefixResourceName, "description", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "role_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "parent_id", ""),
					resource.TestCheckResourceAttr(prefixResourceName, "rir_id", ""),
					resource.TestCheckNoResourceAttr(prefixResourceName, "date_allocated"),

					resource.TestCheckResourceAttr(prefixResourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(prefixResourceName, "created"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "network"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "broadcast"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "prefix_length"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "ip_version"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "namespace_id"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "display"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "url"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(prefixResourceName, "notes_url"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_full(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-full-%d", seed)
	vid := testAccVLANVid(seed, 15)
	cidr := testAccPrefixCIDR(seed, 12)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigFull(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "status", testStatus),
					resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttrPair(prefixResourceName, "vlan_id", "nautobot_vlan.v", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-update-%d", seed)
	vid := testAccVLANVid(seed, 16)
	cidr := testAccPrefixCIDR(seed, 13)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigFull(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccPrefixConfigUpdated(name, vid, cidr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccPrefixConfigUpdated(name, vid, cidr),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(prefixResourceName, "prefix", cidr),
					resource.TestCheckResourceAttr(prefixResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(prefixResourceName, "tenant_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_import(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-import-%d", seed)
	vid := testAccVLANVid(seed, 17)
	cidr := testAccPrefixCIDR(seed, 14)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
			},
			{
				ResourceName:      prefixResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_delete(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-prefix-delete-%d", seed)
	vid := testAccVLANVid(seed, 18)
	cidr := testAccPrefixCIDR(seed, 15)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigMinimal(name, vid, cidr),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccPrefixResource_parallel(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	baseName := fmt.Sprintf("tf-acc-prefix-parallel-%d", seed)
	baseVid := testAccVLANVid(seed, 19)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPrefixConfigParallel(baseName, baseVid, seed),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("nautobot_prefix.p1", "id"),
					resource.TestCheckResourceAttrSet("nautobot_prefix.p2", "id"),
					resource.TestCheckResourceAttrSet("nautobot_prefix.p3", "id"),

					resource.TestCheckResourceAttrPair("nautobot_prefix.p1", "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttrPair("nautobot_prefix.p2", "vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttrPair("nautobot_prefix.p3", "vlan_id", "nautobot_vlan.v", "id"),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "status", testStatus),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "status", testStatus),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "status", testStatus),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "description", ""),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "description", ""),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "description", ""),

					resource.TestCheckResourceAttr("nautobot_prefix.p1", "tags_ids.#", "0"),
					resource.TestCheckResourceAttr("nautobot_prefix.p2", "tags_ids.#", "0"),
					resource.TestCheckResourceAttr("nautobot_prefix.p3", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
