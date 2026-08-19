package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const manufacturerResourceName = "nautobot_manufacturer.test"

func testAccManufacturerConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name = "%s"
}
`, name)
}

func testAccManufacturerConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name        = "%s"
  description = "created by terraform acceptance test"
}
`, name)
}

func testAccManufacturerConfigUpdated(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "test" {
  name        = "%s-updated"
  description = "updated by terraform acceptance test"
}
`, name)
}

func testAccManufacturerConfigParallel(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_manufacturer" "m1" {
  name = "%s-1"
}

resource "nautobot_manufacturer" "m2" {
  name = "%s-2"
}

resource "nautobot_manufacturer" "m3" {
  name = "%s-3"
}
`, base, base, base)
}

func TestAccManufacturerResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "id"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "display"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "url"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "id"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "display"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "url"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(manufacturerResourceName, "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_updateAndDrift(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-upd-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", "created by terraform acceptance test"),
				),
			},
			{
				Config:             testAccManufacturerConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccManufacturerConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(manufacturerResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(manufacturerResourceName, "description", "updated by terraform acceptance test"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigMinimal(name),
			},
			{
				ResourceName:      manufacturerResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-manufacturer-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccManufacturerResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-manufacturer-par-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccManufacturerConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_manufacturer.m1", "name", base+"-1"),
					resource.TestCheckResourceAttr("nautobot_manufacturer.m1", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_manufacturer.m1", "id"),

					resource.TestCheckResourceAttr("nautobot_manufacturer.m2", "name", base+"-2"),
					resource.TestCheckResourceAttr("nautobot_manufacturer.m2", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_manufacturer.m2", "id"),

					resource.TestCheckResourceAttr("nautobot_manufacturer.m3", "name", base+"-3"),
					resource.TestCheckResourceAttr("nautobot_manufacturer.m3", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_manufacturer.m3", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
