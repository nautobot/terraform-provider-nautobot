package provider

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	nb "github.com/nautobot/go-nautobot/v3"
)

func testAccAPIClient() *nb.APIClient {
	cfg := nb.NewConfiguration()
	cfg.Servers = nb.ServerConfigurations{{URL: testURL + "/api"}}
	cfg.HTTPClient = &http.Client{
		Transport: &authRT{base: http.DefaultTransport, token: testToken},
	}
	return nb.NewAPIClient(cfg)
}

func testCaptureResourceID(resourceAddr string, target *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resourceAddr)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("%s has an empty ID", resourceAddr)
		}
		*target = rs.Primary.ID
		return nil
	}
}

func testCheckTenantAbsent(id *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if *id == "" {
			return fmt.Errorf("tenant ID was not captured")
		}
		_, resp, err := testAccAPIClient().TenancyAPI.TenancyTenantsRetrieve(context.Background(), *id).Execute()
		if isNotFoundResponse(resp) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("retrieve tenant %s: %w", *id, err)
		}
		return fmt.Errorf("tenant %s still exists", *id)
	}
}

func testCheckTenantGroupAbsent(id *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if *id == "" {
			return fmt.Errorf("tenant group ID was not captured")
		}
		_, resp, err := testAccAPIClient().TenancyAPI.TenancyTenantGroupsRetrieve(context.Background(), *id).Execute()
		if isNotFoundResponse(resp) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("retrieve tenant group %s: %w", *id, err)
		}
		return fmt.Errorf("tenant group %s still exists", *id)
	}
}

func testMutateTenantByID(t *testing.T, id, name, description string) {
	t.Helper()
	if id == "" {
		t.Fatal("cannot mutate tenant: ID was not captured")
	}

	patch := nb.PatchedTenantRequest{Name: &name, Description: &description}
	_, resp, err := testAccAPIClient().TenancyAPI.
		TenancyTenantsPartialUpdate(context.Background(), id).
		PatchedTenantRequest(patch).
		Execute()
	if err != nil {
		t.Fatalf("mutate tenant %s: %s", id, httpErr(err, resp))
	}
}

func testMutateTenantGroupByID(t *testing.T, id, name, description string) {
	t.Helper()
	if id == "" {
		t.Fatal("cannot mutate tenant group: ID was not captured")
	}

	patch := nb.PatchedTenantGroupRequest{Name: &name, Description: &description}
	_, resp, err := testAccAPIClient().TenancyAPI.
		TenancyTenantGroupsPartialUpdate(context.Background(), id).
		PatchedTenantGroupRequest(patch).
		Execute()
	if err != nil {
		t.Fatalf("mutate tenant group %s: %s", id, httpErr(err, resp))
	}
}

func testDeleteTenantOutOfBand(resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("cannot delete missing tenant resource: %s", resourceAddr)
		}
		resp, err := testAccAPIClient().TenancyAPI.
			TenancyTenantsDestroy(context.Background(), rs.Primary.ID).
			Execute()
		if err != nil && !isNotFoundResponse(resp) {
			return fmt.Errorf("delete tenant %s: %s", rs.Primary.ID, httpErr(err, resp))
		}
		return nil
	}
}

func testDeleteTenantGroupOutOfBand(resourceAddr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceAddr]
		if !ok || rs.Primary.ID == "" {
			return fmt.Errorf("cannot delete missing tenant group resource: %s", resourceAddr)
		}
		resp, err := testAccAPIClient().TenancyAPI.
			TenancyTenantGroupsDestroy(context.Background(), rs.Primary.ID).
			Execute()
		if err != nil && !isNotFoundResponse(resp) {
			return fmt.Errorf("delete tenant group %s: %s", rs.Primary.ID, httpErr(err, resp))
		}
		return nil
	}
}

var (
	testURL   = strings.TrimRight(testEnvironmentValue("NAUTOBOT_TEST_URL", "http://nautobot:8080"), "/")
	testToken = testEnvironmentValue("NAUTOBOT_TEST_TOKEN", "0123456789abcdef0123456789abcdef01234567")
)

const testStatus = "Active"

func testEnvironmentValue(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func testAccSeedForTest(t *testing.T) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(t.Name()))
	nameHash := h.Sum64()

	seed := uint64(time.Now().UnixNano()) ^ nameHash
	return int64(seed)
}

func testAccPrefixCIDR(seed int64, offset int) string {
	h := fnv.New64a()

	var b [16]byte
	binary.LittleEndian.PutUint64(b[0:8], uint64(seed))
	binary.LittleEndian.PutUint64(b[8:16], uint64(offset))
	_, _ = h.Write(b[:])

	x := h.Sum64()

	o2 := int((x >> 0) & 0xFF)
	o3 := int((x >> 8) & 0xFF)
	o4 := int((x >> 16) & 0xF0)

	if o2 == 0 {
		o2 = 1
	}
	if o3 == 0 {
		o3 = 1
	}

	return fmt.Sprintf("21.%d.%d.%d/28", o2, o3, o4)
}

func testAccVLANVid(seed int64, offset int) int {
	base := int(uint64(seed) % 2000)
	return 1000 + base + offset
}

func testCheckTypeSetContainsResourceAttr(setOwnerAddr, setAttr, otherAddr, otherAttr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		so, ok := s.RootModule().Resources[setOwnerAddr]
		if !ok {
			return fmt.Errorf("not found: %s", setOwnerAddr)
		}
		ov, ok := s.RootModule().Resources[otherAddr]
		if !ok {
			return fmt.Errorf("not found: %s", otherAddr)
		}

		want := ov.Primary.Attributes[otherAttr]
		if want == "" {
			return fmt.Errorf("%s.%s is empty", otherAddr, otherAttr)
		}

		prefix := setAttr + "."
		for k, v := range so.Primary.Attributes {
			if k == setAttr+".#" {
				continue
			}
			if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
				if v == want {
					return nil
				}
			}
		}

		return fmt.Errorf("%s: expected %s to contain %s.%s=%q", setOwnerAddr, setAttr, otherAddr, otherAttr, want)
	}
}

func testCountAtLeast(dsAddr, listAttr string, min int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		raw := rs.Primary.Attributes[listAttr+".#"]
		if raw == "" {
			return fmt.Errorf("%s: %s.# is empty", dsAddr, listAttr)
		}

		n, err := strconv.Atoi(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, raw, err)
		}
		if n < min {
			return fmt.Errorf("%s: expected at least %d %s, got %d", dsAddr, min, listAttr, n)
		}
		return nil
	}
}

func testFindListIndexByAttr(dsAddr, listAttr, matchField, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes[listAttr+".#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		for i := 0; i < n; i++ {
			k := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if rs.Primary.Attributes[k] == want {
				return nil
			}
		}
		return fmt.Errorf("%s: expected to find %s.%s=%q", dsAddr, listAttr, matchField, want)
	}
}

func testCheckListItemHasAttrs(dsAddr, listAttr, matchField, matchValue string, want map[string]string, requiredSet []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes[listAttr+".#"]
		if rawN == "" {
			return fmt.Errorf("%s: %s.# is empty", dsAddr, listAttr)
		}
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if rs.Primary.Attributes[k] == matchValue {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find %s with %s=%q", dsAddr, listAttr, matchField, matchValue)
		}

		for field, expected := range want {
			k := fmt.Sprintf("%s.%d.%s", listAttr, idx, field)
			got := rs.Primary.Attributes[k]
			if got != expected {
				return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, expected, got)
			}
		}

		for _, field := range requiredSet {
			k := fmt.Sprintf("%s.%d.%s", listAttr, idx, field)
			if rs.Primary.Attributes[k] == "" {
				return fmt.Errorf("%s: %s expected to be set, got empty", dsAddr, k)
			}
		}

		return nil
	}
}

func testCheckListItemAttrNull(dsAddr, listAttr, matchField, matchValue, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}

		rawN := rs.Primary.Attributes[listAttr+".#"]
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		for i := 0; i < n; i++ {
			matchKey := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if rs.Primary.Attributes[matchKey] != matchValue {
				continue
			}
			key := fmt.Sprintf("%s.%d.%s", listAttr, i, field)
			if _, exists := rs.Primary.Attributes[key]; exists {
				return fmt.Errorf("%s: %s expected to be null, got %q", dsAddr, key, rs.Primary.Attributes[key])
			}
			return nil
		}

		return fmt.Errorf("%s: expected to find %s with %s=%q", dsAddr, listAttr, matchField, matchValue)
	}
}

func testCheckListItemAttrEqualsResourceAttr(dsAddr, listAttr, matchField, matchValue, field, resAddr, resField string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		r, ok := s.RootModule().Resources[resAddr]
		if !ok {
			return fmt.Errorf("not found: %s", resAddr)
		}
		want := r.Primary.Attributes[resField]
		if want == "" {
			return fmt.Errorf("%s.%s is empty", resAddr, resField)
		}

		rawN := ds.Primary.Attributes[listAttr+".#"]
		if rawN == "" {
			return fmt.Errorf("%s: %s.# is empty", dsAddr, listAttr)
		}
		n, err := strconv.Atoi(rawN)
		if err != nil {
			return fmt.Errorf("%s: cannot parse %s.#=%q: %w", dsAddr, listAttr, rawN, err)
		}

		idx := -1
		for i := 0; i < n; i++ {
			k := fmt.Sprintf("%s.%d.%s", listAttr, i, matchField)
			if ds.Primary.Attributes[k] == matchValue {
				idx = i
				break
			}
		}
		if idx == -1 {
			return fmt.Errorf("%s: expected to find %s with %s=%q", dsAddr, listAttr, matchField, matchValue)
		}

		k := fmt.Sprintf("%s.%d.%s", listAttr, idx, field)
		got := ds.Primary.Attributes[k]
		if got != want {
			return fmt.Errorf("%s: %s expected %q, got %q", dsAddr, k, want, got)
		}
		return nil
	}
}

func testCheckDataSourceAddressNotEqualAllocated(dsAddr string, allocatedAddrs ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		dsVal := rs.Primary.Attributes["address"]
		if dsVal == "" {
			return fmt.Errorf("%s: address is empty", dsAddr)
		}
		for _, a := range allocatedAddrs {
			if a == "" {
				continue
			}
			if dsVal == a {
				return fmt.Errorf("%s: expected data source address to differ from allocated address %q", dsAddr, a)
			}
		}
		return nil
	}
}

func testCheckDataSourceAddressNotInAllocatedResources(dsAddr string, resAddrs ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dsAddr]
		if !ok {
			return fmt.Errorf("not found: %s", dsAddr)
		}
		dsVal := ds.Primary.Attributes["address"]
		if dsVal == "" {
			return fmt.Errorf("%s: address is empty", dsAddr)
		}

		for _, r := range resAddrs {
			rs, ok := s.RootModule().Resources[r]
			if !ok {
				return fmt.Errorf("not found: %s", r)
			}
			alloc := rs.Primary.Attributes["address"]
			if alloc == "" {
				continue
			}
			if dsVal == alloc {
				return fmt.Errorf("%s: data source address %q equals allocated %s.address", dsAddr, dsVal, r)
			}
		}
		return nil
	}
}

// TODO: Remove
func testCheckVLANsCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "vlans", min)
}

func testFindVLANIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "vlans", "name", wantName)
}

func testCheckVLANInListHasAttrs(dsAddr, vlanName string, want map[string]string) resource.TestCheckFunc {
	return testCheckListItemHasAttrs(dsAddr, "vlans", "name", vlanName, want, nil)
}

func testCheckVirtualMachinesCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "virtual_machines", min)
}

func testFindVMIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "virtual_machines", "name", wantName)
}

func testCheckVMInListHasAttrs(dsAddr, vmName string, want map[string]string) resource.TestCheckFunc {
	return testCheckListItemHasAttrs(dsAddr, "virtual_machines", "name", vmName, want, nil)
}

func testCheckPrefixesCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "prefixes", min)
}

func testFindPrefixIndexByCIDR(dsAddr, wantCIDR string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "prefixes", "prefix", wantCIDR)
}

func testCheckPrefixInListHasAttrs(dsAddr, cidr string, want map[string]string) resource.TestCheckFunc {
	return testCheckListItemHasAttrs(dsAddr, "prefixes", "prefix", cidr, want, nil)
}

func testCheckPrefixInListAttrEqualsResourceAttr(dsAddr, cidr, field, resAddr, resField string) resource.TestCheckFunc {
	return testCheckListItemAttrEqualsResourceAttr(dsAddr, "prefixes", "prefix", cidr, field, resAddr, resField)
}

func testCheckManufacturersCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "manufacturers", min)
}

func testFindManufacturerIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "manufacturers", "name", wantName)
}

func testCheckManufacturerInListHasAttrs(dsAddr, mName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return testCheckListItemHasAttrs(dsAddr, "manufacturers", "name", mName, want, requiredComputed)
}

func testCheckNamespacesCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "namespaces", min)
}

func testFindNamespaceIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "namespaces", "name", wantName)
}

func testCheckNamespaceInListHasAttrs(dsAddr, namespaceName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return testCheckListItemHasAttrs(dsAddr, "namespaces", "name", namespaceName, want, requiredComputed)
}

func testCheckClustersCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "clusters", min)
}

func testFindClusterIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "clusters", "name", wantName)
}

func testCheckClusterInListHasAttrs(dsAddr, clName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "cluster_type_id", "created", "last_updated"}
	return testCheckListItemHasAttrs(dsAddr, "clusters", "name", clName, want, requiredComputed)
}

func testCheckClusterTypesCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "cluster_types", min)
}

func testFindClusterTypeIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "cluster_types", "name", wantName)
}

func testCheckClusterTypeInListHasAttrs(dsAddr, ctName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return testCheckListItemHasAttrs(dsAddr, "cluster_types", "name", ctName, want, requiredComputed)
}

func testCheckTenantsCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "tenants", min)
}

func testFindTenantIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "tenants", "name", wantName)
}

func testCheckTenantInListHasAttrs(dsAddr, tName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return testCheckListItemHasAttrs(dsAddr, "tenants", "name", tName, want, requiredComputed)
}

func testCheckTenantGroupsCountAtLeast(dsAddr string, min int) resource.TestCheckFunc {
	return testCountAtLeast(dsAddr, "tenant_groups", min)
}

func testFindTenantGroupIndexByName(dsAddr, wantName string) resource.TestCheckFunc {
	return testFindListIndexByAttr(dsAddr, "tenant_groups", "name", wantName)
}

func testCheckTenantGroupInListHasAttrs(dsAddr, tgName string, want map[string]string) resource.TestCheckFunc {
	requiredComputed := []string{"id", "display", "url", "natural_slug", "created", "last_updated", "notes_url"}
	return testCheckListItemHasAttrs(dsAddr, "tenant_groups", "name", tgName, want, requiredComputed)
}
