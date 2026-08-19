package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	clustersDataSourceName = "data.nautobot_clusters.test"
)

func testAccClustersDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct" {
  name = "%[1]s-ct"
}

resource "nautobot_cluster" "cl1" {
  name            = "%[1]s-1"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
}

resource "nautobot_cluster" "cl2" {
  name            = "%[1]s-2"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
  comments        = "cl2 created by terraform acceptance test"
}

resource "nautobot_cluster" "cl3" {
  name            = "%[1]s-3"
  cluster_type_id = nautobot_cluster_type.ct.id
  tenant_id       = "%[2]s"
  comments        = "cl3 created by terraform acceptance test"
}

data "nautobot_clusters" "test" {
  depends_on = [
    nautobot_cluster.cl1,
    nautobot_cluster.cl2,
    nautobot_cluster.cl3,
  ]
}
`, base, "")
}

func TestAccClustersDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-clusters-%d", time.Now().Unix())

	cl1 := base + "-1"
	cl2 := base + "-2"
	cl3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckClustersCountAtLeast(clustersDataSourceName, 3),

					testFindClusterIndexByName(clustersDataSourceName, cl1),
					testFindClusterIndexByName(clustersDataSourceName, cl2),
					testFindClusterIndexByName(clustersDataSourceName, cl3),

					testCheckClusterInListHasAttrs(clustersDataSourceName, cl1, map[string]string{
						"name":             cl1,
						"tenant_id":        "",
						"comments":         "",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
					testCheckClusterInListHasAttrs(clustersDataSourceName, cl2, map[string]string{
						"name":             cl2,
						"tenant_id":        "",
						"comments":         "cl2 created by terraform acceptance test",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
					testCheckClusterInListHasAttrs(clustersDataSourceName, cl3, map[string]string{
						"name":             cl3,
						"tenant_id":        "",
						"comments":         "cl3 created by terraform acceptance test",
						"cluster_group_id": "",
						"location_id":      "",
						"tags_ids.#":       "0",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
