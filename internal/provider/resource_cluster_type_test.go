package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccClusterTypeConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name = "%s"
}
`, name)
}

func testAccClusterTypeConfigWithDescription(name, description string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name        = "%s"
  description = "%s"
}
`, name, description)
}

func testAccClusterTypeConfigParallel(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct1" {
  name = "%s-1"
}

resource "nautobot_cluster_type" "ct2" {
  name = "%s-2"
}

resource "nautobot_cluster_type" "ct3" {
  name = "%s-3"
}
`, base, base, base)
}

func TestAccClusterTypeResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "id"),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "display"),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "url"),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "natural_slug"),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.test", "created"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_update(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-upd-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", ""),
				),
			},
			{
				Config: testAccClusterTypeConfigWithDescription(name, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster_type.test", "description", "updated description"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
			},
			{
				ResourceName:      "nautobot_cluster_type.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-type-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeResource_parallel(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-cluster-type-par-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct1", "name", base+"-1"),
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct1", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.ct1", "id"),

					resource.TestCheckResourceAttr("nautobot_cluster_type.ct2", "name", base+"-2"),
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct2", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.ct2", "id"),

					resource.TestCheckResourceAttr("nautobot_cluster_type.ct3", "name", base+"-3"),
					resource.TestCheckResourceAttr("nautobot_cluster_type.ct3", "description", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster_type.ct3", "id"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
