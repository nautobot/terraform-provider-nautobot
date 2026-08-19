package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	clusterTypeDataSourceName = "data.nautobot_cluster_type.test"
)

func testAccClusterTypeDataSourceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name = "%s"
}

data "nautobot_cluster_type" "test" {
  name = nautobot_cluster_type.test.name
}
`, name)
}

func testAccClusterTypeDataSourceConfigWithDescription(name, description string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_cluster_type" "test" {
  name        = "%s"
  description = "%s"
}

data "nautobot_cluster_type" "test" {
  name = nautobot_cluster_type.test.name
}
`, name, description)
}

func TestAccClusterTypeDataSource_minimal(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-cluster-type-min-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(
						clusterTypeDataSourceName, "id",
						"nautobot_cluster_type.test", "id",
					),

					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "description", ""),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "notes_url"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}

func TestAccClusterTypeDataSource_withDescription(t *testing.T) {
	t.Parallel()

	name := fmt.Sprintf("tfacc-ds-cluster-type-desc-%d", time.Now().Unix())
	desc := "created by terraform acceptance test (data source)"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTypeDataSourceConfigWithDescription(name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "name", name),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "id"),

					resource.TestCheckResourceAttr(clusterTypeDataSourceName, "description", desc),

					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(clusterTypeDataSourceName, "notes_url"),
				),
			},
			{
				Config: testAccProviderConfig(),
			},
		},
	})
}
