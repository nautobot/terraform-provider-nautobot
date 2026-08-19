package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const namespaceDataSourceName = "data.nautobot_namespace.test"

func testAccNamespaceDataSourceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "test" {
  name = "%s"
}

data "nautobot_namespace" "test" {
  name = nautobot_namespace.test.name
}
`, name)
}

func testAccNamespaceDataSourceConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "namespace_ds" {
  name = "%[1]s-tenant"
}

resource "nautobot_namespace" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
  tenant_id   = nautobot_tenant.namespace_ds.id
}

data "nautobot_namespace" "test" {
  name = nautobot_namespace.test.name
}
`, name)
}

func TestAccNamespaceDataSource_minimal(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-namespace-min-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfigMinimal(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceDataSourceName, "name", name),
					resource.TestCheckResourceAttrPair(namespaceDataSourceName, "id", namespaceResourceName, "id"),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "description", ""),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "location_id", ""),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "tenant_id", ""),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "created"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "last_updated"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "display"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "url"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(namespaceDataSourceName, "notes_url"),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceDataSource_full(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-namespace-full-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceDataSourceConfigFull(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceDataSourceName, "name", name),
					resource.TestCheckResourceAttrPair(namespaceDataSourceName, "id", namespaceResourceName, "id"),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(namespaceDataSourceName, "tenant_id", "nautobot_tenant.namespace_ds", "id"),
					resource.TestCheckResourceAttr(namespaceDataSourceName, "location_id", ""),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceDataSource_notFound(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-ds-namespace-notfound-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig() + fmt.Sprintf(`
data "nautobot_namespace" "test" {
  name = "%s"
}
`, name),
				ExpectError: regexp.MustCompile(`Namespace not found`),
			},
			{Config: testAccProviderConfig()},
		},
	})
}
