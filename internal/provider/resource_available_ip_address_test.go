package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccAvailableIPAddressConfigMinimal(name string, vid int, cidr string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
}
`, name, vid, cidr, testStatus)
}

func testAccAvailableIPAddressConfigWithDNSName(name string, vid int, cidr string, dnsName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[5]s"
}
`, name, vid, cidr, testStatus, dnsName)
}

func testAccAvailableIPAddressConfigWithStatusAndDNSName(name string, vid int, cidr string, status, dnsName string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "test" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[5]s"
  dns_name  = "%[6]s"
}
`, name, vid, cidr, testStatus, status, dnsName)
}

func testAccAvailableIPAddressConfigParallel(name string, vid int, cidr string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[4]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[4]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]s-parallel-1"
}

resource "nautobot_available_ip_address" "ip2" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]s-parallel-2"
}

resource "nautobot_available_ip_address" "ip3" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[4]s"
  dns_name  = "%[1]s-parallel-3"
}
`, name, vid, cidr, testStatus)
}

func TestAccAvailableIPAddressResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-min-%d", seed)
	vid := testAccVLANVid(seed, 9)
	cidr := testAccPrefixCIDR(seed, 3)

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "address"),
					resource.TestCheckResourceAttrSet(resourceName, "ip_version"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_update(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-upd-%d", seed)
	vid := testAccVLANVid(seed, 10)
	cidr := testAccPrefixCIDR(seed, 4)

	resourceName := "nautobot_available_ip_address.test"
	dnsName1 := fmt.Sprintf("tfacc-ip-%d.example.com", time.Now().Unix())
	dnsName2 := fmt.Sprintf("tfacc-ip-upd-%d.example.com", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithDNSName(name, vid, cidr, dnsName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName1),
					resource.TestCheckResourceAttr(resourceName, "status", testStatus),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithStatusAndDNSName(name, vid, cidr, "Reserved", dnsName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", dnsName2),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testAccAvailableIPAddressConfigWithStatusAndDNSName(name, vid, cidr, "Reserved", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttr(resourceName, "dns_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "Reserved"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_import(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-imp-%d", seed)
	vid := testAccVLANVid(seed, 11)
	cidr := testAccPrefixCIDR(seed, 5)

	resourceName := "nautobot_available_ip_address.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_delete(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-del-%d", seed)
	vid := testAccVLANVid(seed, 12)
	cidr := testAccPrefixCIDR(seed, 6)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigMinimal(name, vid, cidr),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccAvailableIPAddressResource_parallelAllocations(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-availip-par-%d", seed)
	vid := testAccVLANVid(seed, 13)
	cidr := testAccPrefixCIDR(seed, 7)

	resourceName1 := "nautobot_available_ip_address.ip1"
	resourceName2 := "nautobot_available_ip_address.ip2"
	resourceName3 := "nautobot_available_ip_address.ip3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAvailableIPAddressConfigParallel(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "address"),
					resource.TestCheckResourceAttr(resourceName1, "status", testStatus),

					resource.TestCheckResourceAttrPair(resourceName2, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "address"),
					resource.TestCheckResourceAttr(resourceName2, "status", testStatus),

					resource.TestCheckResourceAttrPair(resourceName3, "prefix_id", "nautobot_prefix.p", "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "address"),
					resource.TestCheckResourceAttr(resourceName3, "status", testStatus),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
