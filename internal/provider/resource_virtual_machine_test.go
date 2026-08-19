package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vmResourceName = "nautobot_virtual_machine.test"
)

func testAccVirtualMachineConfigMinimal(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "test" {
  name       = "%s"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}
`, name, name, name, status)
}

func testAccVirtualMachineConfigFull(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "test" {
  name                = "%s"
  cluster_id          = nautobot_cluster.cl.id
  status              = "%s"

  vcpus               = 4
  memory              = 8192
  disk                = 100
  comments            = "created by terraform acceptance test"
}
`,
		name,
		name,
		name,
		status,
	)
}

func testAccVirtualMachineConfigUpdated(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "test" {
  name                = "%s-updated"
  cluster_id          = nautobot_cluster.cl.id
  status              = "%s"

  vcpus               = 8
  memory              = 16384
  disk                = 200
  comments            = "updated by terraform acceptance test"

}
`,
		name,
		name,
		name,
		status,
	)
}

func testAccVirtualMachineConfigParallel(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "vm1" {
  name       = "%s-1"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_virtual_machine" "vm2" {
  name       = "%s-2"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}

resource "nautobot_virtual_machine" "vm3" {
  name       = "%s-3"
  cluster_id = nautobot_cluster.cl.id
  status     = "%s"
}
`,
		name,
		name,
		name, status,
		name, status,
		name, status,
	)
}

func TestAccVirtualMachineResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-vm-minimal-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmResourceName, "name", name),
					resource.TestCheckResourceAttr(vmResourceName, "vcpus", "0"),
					resource.TestCheckResourceAttr(vmResourceName, "memory", "0"),
					resource.TestCheckResourceAttr(vmResourceName, "disk", "0"),
					resource.TestCheckResourceAttr(vmResourceName, "comments", ""),
					resource.TestCheckResourceAttr(vmResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vmResourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-vm-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmResourceName, "name", name),
					resource.TestCheckResourceAttr(vmResourceName, "vcpus", "4"),
					resource.TestCheckResourceAttr(vmResourceName, "memory", "8192"),
					resource.TestCheckResourceAttr(vmResourceName, "disk", "100"),
					resource.TestCheckResourceAttr(vmResourceName, "comments", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(vmResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vmResourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-vm-update-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmResourceName, "vcpus", "4"),
					resource.TestCheckResourceAttr(vmResourceName, "memory", "8192"),
					resource.TestCheckResourceAttr(vmResourceName, "disk", "100"),
				),
			},
			{
				Config:             testAccVirtualMachineConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVirtualMachineConfigUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(vmResourceName, "vcpus", "8"),
					resource.TestCheckResourceAttr(vmResourceName, "memory", "16384"),
					resource.TestCheckResourceAttr(vmResourceName, "disk", "200"),
					resource.TestCheckResourceAttr(vmResourceName, "comments", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(vmResourceName, "tenant_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-vm-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfigMinimal(name),
			},
			{
				ResourceName:      vmResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-vm-delete-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineResource_parallel(t *testing.T) {
	t.Parallel()

	baseName := fmt.Sprintf("tf-acc-vm-parallel-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineConfigParallel(baseName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "name", baseName+"-1"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "vcpus", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "memory", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "disk", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "comments", ""),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm1", "tags_ids.#", "0"),

					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "name", baseName+"-2"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "vcpus", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "memory", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "disk", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "comments", ""),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm2", "tags_ids.#", "0"),

					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "name", baseName+"-3"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "vcpus", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "memory", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "disk", "0"),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "comments", ""),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_virtual_machine.vm3", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
