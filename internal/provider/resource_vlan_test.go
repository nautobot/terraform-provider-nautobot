package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	vlanResourceName = "nautobot_vlan.test"
)

func testAccVLANConfigMinimal(name string, vid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name   = "%s"
  vid    = %d
  status = "%s"
}
`, name, vid, status)
}

func testAccVLANConfigFull(name string, vid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name        = "%s"
  vid         = %d
  status      = "%s"

  description = "created by terraform acceptance test"
}
`, name, vid, status)
}

func testAccVLANConfigUpdated(name string, vid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "test" {
  name        = "%s-updated"
  vid         = %d
  status      = "%s"

  description = "updated by terraform acceptance test"
}
`, name, vid, status)
}

func testAccVLANConfigParallel(baseName string, baseVid int) string {
	status := testStatus

	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_vlan" "vlan1" {
  name   = "%s-1"
  vid    = %d
  status = "%s"
}

resource "nautobot_vlan" "vlan2" {
  name   = "%s-2"
  vid    = %d
  status = "%s"
}

resource "nautobot_vlan" "vlan3" {
  name   = "%s-3"
  vid    = %d
  status = "%s"
}
`, baseName, baseVid, status, baseName, baseVid+1, status, baseName, baseVid+2, status)
}

func TestAccVLANResource_minimal(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-minimal-%d", seed)
	vid := testAccVLANVid(seed, 26)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigMinimal(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "name", name),
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanResourceName, "status", testStatus),

					resource.TestCheckResourceAttr(vlanResourceName, "description", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "vlan_group_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "role_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(vlanResourceName, "display"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_full(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-full-%d", seed)
	vid := testAccVLANVid(seed, 27)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigFull(name, vid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "name", name),
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid)),
					resource.TestCheckResourceAttr(vlanResourceName, "status", testStatus),
					resource.TestCheckResourceAttr(vlanResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(vlanResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(vlanResourceName, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-update-%d", seed)

	vid1 := testAccVLANVid(seed, 28)
	vid2 := testAccVLANVid(seed, 29)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigFull(name, vid1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid1)),
					resource.TestCheckResourceAttr(vlanResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccVLANConfigUpdated(name, vid2),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVLANConfigUpdated(name, vid2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(vlanResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(vlanResourceName, "vid", fmt.Sprintf("%d", vid2)),
					resource.TestCheckResourceAttr(vlanResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(vlanResourceName, "tenant_id", ""),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_import(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-import-%d", seed)
	vid := testAccVLANVid(seed, 30)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigMinimal(name, vid),
			},
			{
				ResourceName:      vlanResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_delete(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	name := fmt.Sprintf("tf-acc-vlan-delete-%d", seed)
	vid := testAccVLANVid(seed, 31)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigMinimal(name, vid),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccVLANResource_parallel(t *testing.T) {
	t.Parallel()

	seed := testAccSeedForTest(t)
	baseName := fmt.Sprintf("tf-acc-vlan-parallel-%d", seed)

	baseVid := testAccVLANVid(seed, 32)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccVLANConfigParallel(baseName, baseVid),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "name", baseName+"-1"),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "vid", fmt.Sprintf("%d", baseVid)),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "description", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan1", "tags_ids.#", "0"),

					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "name", baseName+"-2"),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "vid", fmt.Sprintf("%d", baseVid+1)),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "description", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan2", "tags_ids.#", "0"),

					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "name", baseName+"-3"),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "vid", fmt.Sprintf("%d", baseVid+2)),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "description", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_vlan.vlan3", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
