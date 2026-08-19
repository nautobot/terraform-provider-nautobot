package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vmInterfaceResourceName = "nautobot_vm_interface.test"
)

func testAccVMInterfaceConfigMinimal(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_vm_interface" "test" {
  name               = "%s-if0"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}
`, name, name, name, status, name, testStatus)
}

func testAccVMInterfaceConfigFull(name string, vid int, cidr string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[5]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[5]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[5]s"
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
  status     = "%[4]s"
}

resource "nautobot_vm_interface" "test" {
  name               = "%[1]s-if0"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  mac_address      = "AA:BB:CC:DD:EE:FF"
  enabled          = false
  mtu              = 1500
  description      = "created by terraform acceptance test"
  mode             = "Access"
  untagged_vlan_id = nautobot_vlan.v.id
  ip_addresses     = [nautobot_available_ip_address.ip1.id]
}
`, name, vid, cidr, status, testStatus)
}

func testAccVMInterfaceConfigUpdated(name string, vid int, cidr string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "v" {
  name   = "%[1]s-vlan"
  vid    = %[2]d
  status = "%[5]s"
}

resource "nautobot_prefix" "p" {
  prefix  = "%[3]s"
  status  = "%[5]s"
  vlan_id = nautobot_vlan.v.id
}

resource "nautobot_available_ip_address" "ip1" {
  prefix_id = nautobot_prefix.p.id
  status    = "%[5]s"
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
  status     = "%[4]s"
}

resource "nautobot_vm_interface" "test" {
  name               = "%[1]s-if0-updated"
  status             = "%[5]s"
  virtual_machine_id = nautobot_virtual_machine.vm.id

  mac_address      = "AA:BB:CC:DD:EE:11"
  enabled          = true
  mtu              = 9000
  description      = "updated by terraform acceptance test"
  mode             = "Tagged"
  untagged_vlan_id = nautobot_vlan.v.id
  ip_addresses     = [nautobot_available_ip_address.ip1.id]
}
`, name, vid, cidr, status, testStatus)
}

func testAccVMInterfaceConfigParallel(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm" {
  name       = "%s-vm"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_vm_interface" "if1" {
  name               = "%s-if1"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}

resource "nautobot_vm_interface" "if2" {
  name               = "%s-if2"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}

resource "nautobot_vm_interface" "if3" {
  name               = "%s-if3"
  status             = "%s"
  virtual_machine_id = nautobot_virtual_machine.vm.id
}
`,
		name, name, name, status,
		name, testStatus,
		name, testStatus,
		name, testStatus,
	)
}

func TestAccVMInterfaceResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testStatus),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "virtual_machine_id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "untagged_vlan_id", ""),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_full(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-if-full-%d", seed)
	vid := testAccVLANVid(seed, 20)
	cidr := testAccPrefixCIDR(seed, 16)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigFull(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testStatus),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "virtual_machine_id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Access"),
					resource.TestCheckResourceAttrPair(vmInterfaceResourceName, "untagged_vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					testCheckTypeSetContainsResourceAttr(vmInterfaceResourceName, "ip_addresses", "nautobot_available_ip_address.ip1", "id"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(vmInterfaceResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tfacc-vm-if-upd-%d", seed)
	vid := testAccVLANVid(seed, 21)
	cidr := testAccPrefixCIDR(seed, 17)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigFull(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testStatus),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:FF"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "1500"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "created by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Access"),
					resource.TestCheckResourceAttrPair(vmInterfaceResourceName, "untagged_vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					testCheckTypeSetContainsResourceAttr(vmInterfaceResourceName, "ip_addresses", "nautobot_available_ip_address.ip1", "id"),
				),
			},
			{
				Config:             testAccVMInterfaceConfigUpdated(name, vid, cidr),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVMInterfaceConfigUpdated(name, vid, cidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "name", name+"-if0-updated"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "status", testStatus),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mac_address", "AA:BB:CC:DD:EE:11"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mtu", "9000"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "description", "updated by terraform acceptance test"),

					resource.TestCheckResourceAttr(vmInterfaceResourceName, "mode", "Tagged"),
					resource.TestCheckResourceAttrPair(vmInterfaceResourceName, "untagged_vlan_id", "nautobot_vlan.v", "id"),
					resource.TestCheckResourceAttr(vmInterfaceResourceName, "ip_addresses.#", "1"),
					testCheckTypeSetContainsResourceAttr(vmInterfaceResourceName, "ip_addresses", "nautobot_available_ip_address.ip1", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
			},
			{
				ResourceName:      vmInterfaceResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVMInterfaceResource_parallel(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-vm-if-par-%d", time.Now().Unix())

	resourceName1 := "nautobot_vm_interface.if1"
	resourceName2 := "nautobot_vm_interface.if2"
	resourceName3 := "nautobot_vm_interface.if3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVMInterfaceConfigParallel(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", name+"-if1"),
					resource.TestCheckResourceAttr(resourceName1, "status", testStatus),
					resource.TestCheckResourceAttrSet(resourceName1, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName1, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName1, "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr(resourceName2, "name", name+"-if2"),
					resource.TestCheckResourceAttr(resourceName2, "status", testStatus),
					resource.TestCheckResourceAttrSet(resourceName2, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName2, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName2, "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr(resourceName3, "name", name+"-if3"),
					resource.TestCheckResourceAttr(resourceName3, "status", testStatus),
					resource.TestCheckResourceAttrSet(resourceName3, "virtual_machine_id"),
					resource.TestCheckResourceAttr(resourceName3, "tags_ids.#", "0"),
					resource.TestCheckResourceAttr(resourceName3, "ip_addresses.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
