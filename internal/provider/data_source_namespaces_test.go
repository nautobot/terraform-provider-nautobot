package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const namespacesDataSourceName = "data.nautobot_namespaces.test"

func testAccNamespacesDataSourceConfig(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "n1" {
  name = "%[1]s-1"
}

resource "nautobot_namespace" "n2" {
  name        = "%[1]s-2"
  description = "n2 created by terraform acceptance test"
}

resource "nautobot_namespace" "n3" {
  name        = "%[1]s-3"
  description = "n3 created by terraform acceptance test"
}

data "nautobot_namespaces" "test" {
  depends_on = [
    nautobot_namespace.n1,
    nautobot_namespace.n2,
    nautobot_namespace.n3,
  ]
}
`, base)
}

func TestAccNamespacesDataSource_list(t *testing.T) {
	t.Parallel()
	base := fmt.Sprintf("tfacc-ds-namespaces-%d", testAccSeedForTest(t))
	n1, n2, n3 := base+"-1", base+"-2", base+"-3"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNamespacesDataSourceConfig(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckNamespacesCountAtLeast(namespacesDataSourceName, 3),
					testFindNamespaceIndexByName(namespacesDataSourceName, n1),
					testFindNamespaceIndexByName(namespacesDataSourceName, n2),
					testFindNamespaceIndexByName(namespacesDataSourceName, n3),
					testCheckNamespaceInListHasAttrs(namespacesDataSourceName, n1, map[string]string{
						"name": n1, "description": "", "location_id": "", "tenant_id": "",
					}),
					testCheckNamespaceInListHasAttrs(namespacesDataSourceName, n2, map[string]string{
						"name": n2, "description": "n2 created by terraform acceptance test", "location_id": "", "tenant_id": "",
					}),
					testCheckNamespaceInListHasAttrs(namespacesDataSourceName, n3, map[string]string{
						"name": n3, "description": "n3 created by terraform acceptance test", "location_id": "", "tenant_id": "",
					}),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}
