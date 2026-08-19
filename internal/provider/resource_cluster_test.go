package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func testAccClusterConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "test" {
  name            = "%s"
  cluster_type_id = nautobot_cluster_type.ct.id
}
`, name, name)
}

func testAccClusterConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "test" {
  name             = "%s-updated"
  cluster_type_id  = nautobot_cluster_type.ct.id
  comments         = "updated comment"
}
`, name, name)
}

func testAccClusterConfigParallel(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%s-ct"
}

resource "nautobot_cluster" "cl1" {
  name            = "%s-1"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_cluster" "cl2" {
  name            = "%s-2"
  cluster_type_id = nautobot_cluster_type.ct.id
}

resource "nautobot_cluster" "cl3" {
  name            = "%s-3"
  cluster_type_id = nautobot_cluster_type.ct.id
}
`, name,
		name,
		name,
		name,
	)
}

func TestAccClusterResource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "name", name),
					resource.TestCheckResourceAttrSet("nautobot_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("nautobot_cluster.test", "cluster_type_id"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "cluster_group_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "location_id", ""),
					resource.TestCheckResourceAttrSet("nautobot_cluster.test", "created"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_update(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-upd-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "name", name),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", ""),
				),
			},
			{
				Config: testAccClusterConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_cluster.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "comments", "updated comment"),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tenant_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "cluster_group_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "location_id", ""),
					resource.TestCheckResourceAttr("nautobot_cluster.test", "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_import(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-import-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
			},
			{
				ResourceName:      "nautobot_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_delete(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-del-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigMinimal(name),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterResource_parallel(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-cluster-par-%d", time.Now().Unix())

	resourceName1 := "nautobot_cluster.cl1"
	resourceName2 := "nautobot_cluster.cl2"
	resourceName3 := "nautobot_cluster.cl3"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigParallel(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName1, "name", name+"-1"),
					resource.TestCheckResourceAttrSet(resourceName1, "id"),
					resource.TestCheckResourceAttrSet(resourceName1, "cluster_type_id"),
					resource.TestCheckResourceAttr(resourceName1, "comments", ""),
					resource.TestCheckResourceAttr(resourceName1, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(resourceName1, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName1, "location_id", ""),
					resource.TestCheckResourceAttr(resourceName1, "tags_ids.#", "0"),

					resource.TestCheckResourceAttr(resourceName2, "name", name+"-2"),
					resource.TestCheckResourceAttrSet(resourceName2, "id"),
					resource.TestCheckResourceAttrSet(resourceName2, "cluster_type_id"),
					resource.TestCheckResourceAttr(resourceName2, "comments", ""),
					resource.TestCheckResourceAttr(resourceName2, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(resourceName2, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName2, "location_id", ""),
					resource.TestCheckResourceAttr(resourceName2, "tags_ids.#", "0"),

					resource.TestCheckResourceAttr(resourceName3, "name", name+"-3"),
					resource.TestCheckResourceAttrSet(resourceName3, "id"),
					resource.TestCheckResourceAttrSet(resourceName3, "cluster_type_id"),
					resource.TestCheckResourceAttr(resourceName3, "comments", ""),
					resource.TestCheckResourceAttr(resourceName3, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(resourceName3, "tenant_id", ""),
					resource.TestCheckResourceAttr(resourceName3, "location_id", ""),
					resource.TestCheckResourceAttr(resourceName3, "tags_ids.#", "0"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
