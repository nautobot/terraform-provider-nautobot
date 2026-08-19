package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	nb "github.com/nautobot/go-nautobot/v3"
)

func TestNewSecurityProviderNautobotToken(t *testing.T) {
	sp, err := NewSecurityProviderNautobotToken("abc123")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if sp == nil {
		t.Fatalf("expected non-nil provider")
	}
}

func TestSecurityProviderNautobotToken_Intercept(t *testing.T) {
	sp, _ := NewSecurityProviderNautobotToken("abc123")
	req, _ := http.NewRequest(http.MethodGet, testURL, nil)

	if err := sp.Intercept(context.Background(), req); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}

	got := req.Header.Get("Authorization")
	if got != "Token abc123" {
		t.Fatalf("expected Authorization header %q, got %q", "Token abc123", got)
	}
}

func TestStringPtr(t *testing.T) {
	p := stringPtr("x")
	if p == nil || *p != "x" {
		t.Fatalf("expected ptr to %q, got %+v", "x", p)
	}
}

func TestInt32Ptr(t *testing.T) {
	p := int32Ptr(42)
	if p == nil || *p != int32(42) {
		t.Fatalf("expected ptr to %d, got %+v", 42, p)
	}
}

func TestDerefStr(t *testing.T) {
	if got := derefStr(nil); got != "" {
		t.Fatalf("expected nil pointer to become an empty string, got %q", got)
	}
	empty := ""
	if got := derefStr(&empty); got != "" {
		t.Fatalf("expected empty string to remain empty, got %q", got)
	}
	want := "value"
	if got := derefStr(&want); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNullableTimeStr(t *testing.T) {
	var unset nb.NullableTime
	if got := nullableTimeStr(unset); !got.IsNull() {
		t.Fatalf("expected unset time to become null, got %v", got)
	}
	var explicitNull nb.NullableTime
	explicitNull.Set(nil)
	if got := nullableTimeStr(explicitNull); !got.IsNull() {
		t.Fatalf("expected explicit null time to become null, got %v", got)
	}

	want := time.Date(2026, time.July, 22, 12, 34, 56, 0, time.FixedZone("test", 2*60*60))
	var value nb.NullableTime
	value.Set(&want)
	if got := nullableTimeStr(value).ValueString(); got != want.Format(time.RFC3339) {
		t.Fatalf("expected %q, got %q", want.Format(time.RFC3339), got)
	}
}

func TestNullableFKStr(t *testing.T) {
	var unset nb.NullableApprovalWorkflowUser
	if got := nullableFKStr(unset).ValueString(); got != "" {
		t.Fatalf("expected unset FK to become empty, got %q", got)
	}

	var explicitNull nb.NullableApprovalWorkflowUser
	explicitNull.Set(nil)
	if got := nullableFKStr(explicitNull).ValueString(); got != "" {
		t.Fatalf("expected explicit null FK to become empty, got %q", got)
	}

	var missingID nb.NullableApprovalWorkflowUser
	missingID.Set(&nb.ApprovalWorkflowUser{})
	if got := nullableFKStr(missingID).ValueString(); got != "" {
		t.Fatalf("expected FK without an ID to become empty, got %q", got)
	}

	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	if got := nullableFKStr(makeFKUser(want)).ValueString(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNullableSoftwareVersionStr(t *testing.T) {
	var unset nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	if got := nullableSoftwareVersionStr(unset).ValueString(); got != "" {
		t.Fatalf("expected unset software version to become empty, got %q", got)
	}

	var missingID nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	missingID.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{})
	if got := nullableSoftwareVersionStr(missingID).ValueString(); got != "" {
		t.Fatalf("expected software version without an ID to become empty, got %q", got)
	}

	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	var populated nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
	populated.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{String: &want},
	})
	if got := nullableSoftwareVersionStr(populated).ValueString(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestMakeFKUserJSON(t *testing.T) {
	empty, err := json.Marshal(makeFKUser(""))
	if err != nil {
		t.Fatalf("marshal empty FK: %v", err)
	}
	if string(empty) != "null" {
		t.Fatalf("expected empty FK to marshal as null, got %s", empty)
	}

	want := "748ca2dd-a3ac-5bb6-8b4a-276b7e3c33c7"
	populated, err := json.Marshal(makeFKUser(want))
	if err != nil {
		t.Fatalf("marshal populated FK: %v", err)
	}
	if !strings.Contains(string(populated), want) {
		t.Fatalf("expected populated FK JSON to contain %q, got %s", want, populated)
	}
}

func TestSliceToSetAndSetDiff(t *testing.T) {
	a := sliceToSet([]string{"a", "b", "b", "", "c"})
	b := sliceToSet([]string{"b", "c"})

	diff := setDiff(a, b)

	if len(diff) != 1 || diff[0] != "a" {
		t.Fatalf("expected diff [a], got %v", diff)
	}
}

func TestCollectStrings(t *testing.T) {
	in := map[string]any{
		"name": []any{"bad"},
		"nested": map[string]any{
			"other": []any{"also bad"},
		},
		"empty": "",
	}
	out := collectStrings(in)

	if len(out) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(out), out)
	}
	if out[0] == "" || out[1] == "" {
		t.Fatalf("expected non-empty messages, got %v", out)
	}
}

func TestBestJSONMessage(t *testing.T) {
	m1 := map[string]any{"detail": "nope"}
	if got := bestJSONMessage(m1); got != "nope" {
		t.Fatalf("expected %q, got %q", "nope", got)
	}

	m2 := map[string]any{"non_field_errors": []any{"x", "y"}, "name": []any{"ignored"}}
	if got := bestJSONMessage(m2); got != "x | y" {
		t.Fatalf("expected %q, got %q", "x | y", got)
	}

	m3 := map[string]any{"name": []any{"bad name"}, "vid": []any{"bad vid"}}
	got3 := bestJSONMessage(m3)
	if got3 == "" {
		t.Fatalf("expected non-empty message, got empty")
	}
}

func TestHttpErr_TwoLineFormatAndBodyPreserved(t *testing.T) {
	body := `{"detail":"boom"}`
	req, _ := http.NewRequest(http.MethodPost, testURL+"/api/x", nil)
	resp := &http.Response{
		Status:  "400 Bad Request",
		Body:    io.NopCloser(bytes.NewBufferString(body)),
		Request: req,
	}

	out := httpErr(errors.New("ignored"), resp)

	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), out)
	}
	if lines[0] != "400 Bad Request" {
		t.Fatalf("expected status line %q, got %q", "400 Bad Request", lines[0])
	}
	if !strings.Contains(lines[1], "boom") {
		t.Fatalf("expected message line to contain %q, got %q", "boom", lines[1])
	}
	if !strings.Contains(lines[1], "(POST "+testURL+"/api/x)") {
		t.Fatalf("expected message line to contain request info, got %q", lines[1])
	}

	b2, _ := io.ReadAll(resp.Body)
	if string(b2) != body {
		t.Fatalf("expected body preserved %q, got %q", body, string(b2))
	}
}

func TestAccGetStatusIDAndGetStatusName(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC is not set; skipping acceptance test")
	}

	token := testToken
	baseURL := testURL + "/api"

	cfg := nb.NewConfiguration()
	cfg.Servers = nb.ServerConfigurations{
		{URL: baseURL},
	}

	httpClient := &http.Client{
		Transport: &authRT{
			base:  http.DefaultTransport,
			token: token,
		},
	}
	cfg.HTTPClient = httpClient

	client := nb.NewAPIClient(cfg)
	ctx := context.Background()

	id, err := getStatusID(ctx, client, testStatus)
	if err != nil {
		t.Fatalf("getStatusID: expected nil err, got %v", err)
	}
	if id == "" {
		t.Fatalf("getStatusID: expected non-empty id")
	}

	name, err := getStatusName(ctx, client, id)
	if err != nil {
		t.Fatalf("getStatusName: expected nil err, got %v", err)
	}
	if name != testStatus {
		t.Fatalf("getStatusName: expected %q, got %q", testStatus, name)
	}
}
