package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	clusterTypesDataSourceName = "data.nautobot_cluster_types.test"
)

func testAccClusterTypesDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "ct1" {
  name = "%[1]s-1"
}

resource "nautobot_cluster_type" "ct2" {
  name        = "%[1]s-2"
  description = "ct2 created by terraform acceptance test"
}

resource "nautobot_cluster_type" "ct3" {
  name        = "%[1]s-3"
  description = "ct3 created by terraform acceptance test"
}

data "nautobot_cluster_types" "test" {
  depends_on = [
    nautobot_cluster_type.ct1,
    nautobot_cluster_type.ct2,
    nautobot_cluster_type.ct3,
  ]
}
`, base)
}

func TestAccClusterTypesDataSource_list(t *testing.T) {
	t.Parallel()

	base := fmt.Sprintf("tfacc-ds-cluster-types-%d", time.Now().Unix())

	ct1 := base + "-1"
	ct2 := base + "-2"
	ct3 := base + "-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypesDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckClusterTypesCountAtLeast(clusterTypesDataSourceName, 3),

					testFindClusterTypeIndexByName(clusterTypesDataSourceName, ct1),
					testFindClusterTypeIndexByName(clusterTypesDataSourceName, ct2),
					testFindClusterTypeIndexByName(clusterTypesDataSourceName, ct3),

					testCheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct1, map[string]string{
						"name":        ct1,
						"description": "",
					}),

					testCheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct2, map[string]string{
						"name":        ct2,
						"description": "ct2 created by terraform acceptance test",
					}),
					testCheckClusterTypeInListHasAttrs(clusterTypesDataSourceName, ct3, map[string]string{
						"name":        ct3,
						"description": "ct3 created by terraform acceptance test",
					}),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
