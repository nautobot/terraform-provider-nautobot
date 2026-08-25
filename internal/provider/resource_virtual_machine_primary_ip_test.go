package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vmPrimaryIPResourceName = "nautobot_vm_primary_ip.test"
)

func testAccVMPrimaryIPConfigIPv4Single(name string, vid int, cidr string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%[1]s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[3]s"
}

resource "nautobot_available_ip_address" "ip4" {
  prefix_id = nautobot_prefix.p.id
  status    = "Active"
}

resource "nautobot_vm_interface" "if0" {
  name               = "%[1]s-if0"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  ip_addresses = [nautobot_available_ip_address.ip4.id]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4.id

  depends_on = [nautobot_vm_interface.if0]
}
`, name, vid, status, cidr, testStatus)
}

func testAccVMPrimaryIPConfigIPv4UpdateA(name string, vid int, cidr string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%[1]s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[3]s"
}

resource "nautobot_available_ip_address" "ip4a" {
  prefix_id = nautobot_prefix.p.id
  status    = "Active"
}

resource "nautobot_available_ip_address" "ip4b" {
  prefix_id = nautobot_prefix.p.id
  status    = "Active"
}

resource "nautobot_vm_interface" "if0" {
  name               = "%[1]s-if0"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  ip_addresses = [
    nautobot_available_ip_address.ip4a.id,
    nautobot_available_ip_address.ip4b.id,
  ]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4a.id

  depends_on = [nautobot_vm_interface.if0]
}
`, name, vid, status, cidr, testStatus)
}

func testAccVMPrimaryIPConfigIPv4UpdateB(name string, vid int, cidr string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[3]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[4]s"
  status  = "%[3]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%[1]s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[3]s"
}

resource "nautobot_available_ip_address" "ip4a" {
  prefix_id = nautobot_prefix.p.id
  status    = "Active"
}

resource "nautobot_available_ip_address" "ip4b" {
  prefix_id = nautobot_prefix.p.id
  status    = "Active"
}

resource "nautobot_vm_interface" "if0" {
  name               = "%[1]s-if0"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  ip_addresses = [
    nautobot_available_ip_address.ip4a.id,
    nautobot_available_ip_address.ip4b.id,
  ]
}

resource "nautobot_vm_primary_ip" "test" {
  virtual_machine_id = nautobot_virtual_machine.vm.id
  primary_ip4_id     = nautobot_available_ip_address.ip4b.id

  depends_on = [nautobot_vm_interface.if0]
}
`, name, vid, status, cidr, testStatus)
}

func TestAccVMPrimaryIPResource_ipv4Single(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-primary-ip-v4single-%d", seed)
	vid := testAccVLANVid(seed, 22)
	cidr := testAccPrefixCIDR(seed, 18)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4Single(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "virtual_machine_id"),
					resource.TestCheckResourceAttrPair(
						vmPrimaryIPResourceName, "primary_ip4_id",
						"nautobot_available_ip_address.ip4", "id",
					),
					resource.TestCheckResourceAttr(vmPrimaryIPResourceName, "primary_ip6_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMPrimaryIPResource_updateIP(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-primary-ip-update-%d", seed)
	vid := testAccVLANVid(seed, 23)
	cidr := testAccPrefixCIDR(seed, 19)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4UpdateA(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "virtual_machine_id"),
					resource.TestCheckResourceAttrPair(
						vmPrimaryIPResourceName, "primary_ip4_id",
						"nautobot_available_ip_address.ip4a", "id",
					),
					resource.TestCheckResourceAttr(vmPrimaryIPResourceName, "primary_ip6_id", ""),
				),
			},
			{
				Config:             testAccVMPrimaryIPConfigIPv4UpdateB(name, vid, cidr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVMPrimaryIPConfigIPv4UpdateB(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmPrimaryIPResourceName, "virtual_machine_id"),
					resource.TestCheckResourceAttrPair(
						vmPrimaryIPResourceName, "primary_ip4_id",
						"nautobot_available_ip_address.ip4b", "id",
					),
					resource.TestCheckResourceAttr(vmPrimaryIPResourceName, "primary_ip6_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMPrimaryIPResource_import(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-primary-ip-import-%d", seed)
	vid := testAccVLANVid(seed, 24)
	cidr := testAccPrefixCIDR(seed, 20)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4Single(name, vid, cidr),
			},
			{
				ResourceName:      vmPrimaryIPResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMPrimaryIPResource_delete(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-primary-ip-del-%d", seed)
	vid := testAccVLANVid(seed, 25)
	cidr := testAccPrefixCIDR(seed, 21)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMPrimaryIPConfigIPv4Single(name, vid, cidr),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
