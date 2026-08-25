package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vmDataSourceName = "data.nautobot_virtual_machine.test"
)

func testAccVirtualMachineDataSourceConfigMinimal(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "test" {
  name       = "%[1]s"
  cluster_id = nautobot_cluster.cl.id
  status     = "%[2]s"
}

data "nautobot_virtual_machine" "test" {
  depends_on = [nautobot_virtual_machine.test]
  name = nautobot_virtual_machine.test.name
}
`, name, status)
}

func testAccVirtualMachineDataSourceConfigFull(name string) string {
	status := "Active"

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl" {
  name            = "%[1]s-cl"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_virtual_machine" "test" {
  name                = "%[1]s"
  cluster_id          = nautobot_cluster.cl.id
  status              = "%[2]s"

  vcpus               = 4
  memory              = 8192
  disk                = 100
  comments            = "created by terraform acceptance test"
}

data "nautobot_virtual_machine" "test" {
  depends_on = [nautobot_virtual_machine.test]
  name = nautobot_virtual_machine.test.name
}
`,
		name,
		status,
	)
}

func TestAccVirtualMachineDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-minimal-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						vmDataSourceName, "cluster_id",
						"nautobot_cluster.cl", "id",
					),
					resource.TestCheckResourceAttr(vmDataSourceName, "status", "Active"),

					resource.TestCheckResourceAttr(vmDataSourceName, "vcpus", "0"),
					resource.TestCheckResourceAttr(vmDataSourceName, "memory", "0"),
					resource.TestCheckResourceAttr(vmDataSourceName, "disk", "0"),

					resource.TestCheckResourceAttr(vmDataSourceName, "comments", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "platform_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "software_version_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip4_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip6_id", ""),

					resource.TestCheckResourceAttr(vmDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVirtualMachineDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-vm-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVirtualMachineDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vmDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(
						vmDataSourceName, "id",
						vmResourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						vmDataSourceName, "cluster_id",
						"nautobot_cluster.cl", "id",
					),

					resource.TestCheckResourceAttr(vmDataSourceName, "status", "Active"),

					resource.TestCheckResourceAttr(vmDataSourceName, "vcpus", "4"),
					resource.TestCheckResourceAttr(vmDataSourceName, "memory", "8192"),
					resource.TestCheckResourceAttr(vmDataSourceName, "disk", "100"),

					resource.TestCheckResourceAttr(vmDataSourceName, "comments", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(vmDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "platform_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "software_version_id", ""),

					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip4_id", ""),
					resource.TestCheckResourceAttr(vmDataSourceName, "primary_ip6_id", ""),

					resource.TestCheckResourceAttr(vmDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vmDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
