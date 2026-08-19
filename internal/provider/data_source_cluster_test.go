package provider

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	clusterDataSourceName = "data.nautobot_cluster.test"
)

func testAccClusterDataSourceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "test" {
  name            = "%[1]s"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
}

data "nautobot_cluster" "test" {
  name = nautobot_cluster.test.name
}
`, name, "")
}

func testAccClusterDataSourceConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "test" {
  name             = "%[1]s"
  cluster_type_id  = nautobot_cluster_type.ct.id
  tenant_id        = "%[2]s"
  comments         = "created by terraform acceptance test"
  cluster_group_id = "%[3]s"
  location_id      = "%[4]s"
}

data "nautobot_cluster" "test" {
  name = nautobot_cluster.test.name
}
`, name, "", "", "")
}

func TestAccClusterDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-minimal-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "id",
						"nautobot_cluster.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "cluster_type_id",
						"nautobot_cluster.test", "cluster_type_id",
					),

					resource.TestCheckResourceAttr(clusterDataSourceName, "comments", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "location_id", ""),

					resource.TestCheckResourceAttr(clusterDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterDataSource_full(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-full-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "id"),

					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "id",
						"nautobot_cluster.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						clusterDataSourceName, "cluster_type_id",
						"nautobot_cluster.test", "cluster_type_id",
					),

					resource.TestCheckResourceAttr(clusterDataSourceName, "comments", "created by terraform acceptance test"),
					resource.TestCheckResourceAttr(clusterDataSourceName, "cluster_group_id", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttr(clusterDataSourceName, "location_id", ""),

					resource.TestCheckResourceAttr(clusterDataSourceName, "tags_ids.#", "0"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterDataSourceName, "last_updated"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterDataSource_notFound(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tf-acc-ds-cluster-notfound-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "nautobot_cluster" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexpMustCompile(`Cluster not found`),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func regexpMustCompile(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}
