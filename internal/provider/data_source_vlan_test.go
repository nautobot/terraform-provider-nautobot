package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vlanDataSourceName = "data.nautobot_vlan.test"
)

func testAccVLANDataSourceConfigMinimal(name string, vid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name   = "%[1]s"
  vid    = %[2]d
  status = "%[3]s"
}

data "nautobot_vlan" "test" {
  depends_on = [nautobot_vlan.test]
  name = nautobot_vlan.test.name
}
`, name, vid, status)
}

func testAccVLANDataSourceConfigFull(name string, vid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name        = "%[1]s"
  vid         = %[2]d
  status      = "%[3]s"
  description = "created by terraform acceptance test"
}

data "nautobot_vlan" "test" {
  depends_on = [nautobot_vlan.test]
  name = nautobot_vlan.test.name
}
`,
		name,
		vid,
		status,
	)
}

func TestAccVLANDataSource_minimal(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-ds-vlan-minimal-%d", seed)
	vid := testAccVLANVid(seed, 6)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANDataSourceConfigMinimal(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanDataSourceName, "name", name),
					resource.TestCheckResourceAttr(vlanDataSourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanDataSourceName, "status", testStatus),

					resource.TestCheckResourceAttrSet(vlanDataSourceName, "id"),
					resource.TestCheckResourceAttr(vlanDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "vlan_group_id", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vlanDataSourceName, "tags_ids.#", "0"),

					resource.TestCheckResourceAttrSet(vlanDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "notes_url"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVLANDataSource_full(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-ds-vlan-full-%d", seed)
	vid := testAccVLANVid(seed, 7)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANDataSourceConfigFull(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanDataSourceName, "name", name),
					resource.TestCheckResourceAttr(vlanDataSourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanDataSourceName, "status", testStatus),

					resource.TestCheckResourceAttr(vlanDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(vlanDataSourceName, "tenant_id", ""),

					resource.TestCheckResourceAttrSet(vlanDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(vlanDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
