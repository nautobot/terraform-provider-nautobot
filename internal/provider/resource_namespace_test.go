package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const namespaceResourceName = "nautobot_namespace.test"

func testAccNamespaceConfigMinimal(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "test" {
  name = "%s"
}
`, name)
}

func testAccNamespaceConfigFull(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "namespace_test" {
  name = "%[1]s-tenant"
}

resource "nautobot_namespace" "test" {
  name        = "%[1]s"
  description = "created by terraform acceptance test"
  tenant_id   = nautobot_tenant.namespace_test.id
}
`, name)
}

func testAccNamespaceConfigUpdated(name string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_tenant" "namespace_test" {
  name = "%[1]s-tenant"
}

resource "nautobot_namespace" "test" {
  name        = "%[1]s-updated"
  description = "updated by terraform acceptance test"
}
`, name)
}

func testAccNamespaceConfigParallel(base string) string {
	return testAccProviderConfig() + fmt.Sprintf(`
resource "nautobot_namespace" "n1" {
  name = "%[1]s-1"
}

resource "nautobot_namespace" "n2" {
  name = "%[1]s-2"
}

resource "nautobot_namespace" "n3" {
  name = "%[1]s-3"
}
`, base)
}

func TestAccNamespaceResource_minimal(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-namespace-min-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceResourceName, "name", name),
					resource.TestCheckResourceAttr(namespaceResourceName, "description", ""),
					resource.TestCheckResourceAttr(namespaceResourceName, "location_id", ""),
					resource.TestCheckResourceAttr(namespaceResourceName, "tenant_id", ""),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "id"),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "created"),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "display"),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "url"),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "natural_slug"),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "notes_url"),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceResource_full(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-namespace-full-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfigFull(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceResourceName, "name", name),
					resource.TestCheckResourceAttr(namespaceResourceName, "description", "created by terraform acceptance test"),
					resource.TestCheckResourceAttrPair(namespaceResourceName, "tenant_id", "nautobot_tenant.namespace_test", "id"),
					resource.TestCheckResourceAttr(namespaceResourceName, "location_id", ""),
					resource.TestCheckResourceAttrSet(namespaceResourceName, "id"),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceResource_updateAndDrift(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-namespace-upd-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccNamespaceConfigFull(name)},
			{
				Config:             testAccNamespaceConfigUpdated(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccNamespaceConfigUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(namespaceResourceName, "name", name+"-updated"),
					resource.TestCheckResourceAttr(namespaceResourceName, "description", "updated by terraform acceptance test"),
					resource.TestCheckResourceAttr(namespaceResourceName, "tenant_id", ""),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceResource_import(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-namespace-import-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccNamespaceConfigMinimal(name)},
			{ResourceName: namespaceResourceName, ImportState: true, ImportStateVerify: true},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceResource_delete(t *testing.T) {
	t.Parallel()
	name := fmt.Sprintf("tfacc-namespace-del-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: testAccNamespaceConfigMinimal(name)},
			{Config: testAccProviderConfig()},
		},
	})
}

func TestAccNamespaceResource_parallel(t *testing.T) {
	t.Parallel()
	base := fmt.Sprintf("tfacc-namespace-par-%d", testAccSeedForTest(t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNamespaceConfigParallel(base),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nautobot_namespace.n1", "name", base+"-1"),
					resource.TestCheckResourceAttrSet("nautobot_namespace.n1", "id"),
					resource.TestCheckResourceAttr("nautobot_namespace.n2", "name", base+"-2"),
					resource.TestCheckResourceAttrSet("nautobot_namespace.n2", "id"),
					resource.TestCheckResourceAttr("nautobot_namespace.n3", "name", base+"-3"),
					resource.TestCheckResourceAttrSet("nautobot_namespace.n3", "id"),
				),
			},
			{Config: testAccProviderConfig()},
		},
	})
}
