package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

func NewSecurityProviderNautobotToken(t string) (*SecurityProviderNautobotToken, error) {
	return &SecurityProviderNautobotToken{
		token: t,
	}, nil
}

type SecurityProviderNautobotToken struct {
	token string
}

func (s *SecurityProviderNautobotToken) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", s.token))
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int) *int32 {
	val := int32(i)
	return &val
}

func isNotFoundResponse(resp *http.Response) bool {
	return resp != nil && resp.StatusCode == http.StatusNotFound
}

func getStatusName(ctx context.Context, c *nb.APIClient, statusID string) (string, error) {
	status, _, err := c.ExtrasAPI.ExtrasStatusesRetrieve(ctx, statusID).Execute()
	if err != nil {
		return "", err
	}

	if status.Name != "" {
		return status.Name, nil
	}

	return "", fmt.Errorf("status name not found for ID %s", statusID)
}

func getStatusID(ctx context.Context, c *nb.APIClient, statusName string) (string, error) {
	statuses, _, err := c.ExtrasAPI.ExtrasStatusesList(ctx).Name([]string{statusName}).Execute()
	if err != nil {
		return "", err
	}

	if len(statuses.Results) == 0 {
		return "", fmt.Errorf("status %s not found", statusName)
	}

	if statuses.Results[0].Id == nil || *statuses.Results[0].Id == "" {
		return "", fmt.Errorf("status %s returned no id", statusName)
	}

	return *statuses.Results[0].Id, nil
}

// httpErr returns exactly two lines when possible:
// 1) "<status>"
// 2) "<message> (<METHOD> <URL>)"
func httpErr(err error, resp *http.Response) string {
	var statusLine string
	if resp != nil && resp.Status != "" {
		statusLine = resp.Status
	}
	if statusLine == "" && err != nil {
		statusLine = err.Error()
	}

	var reqInfo string
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		reqInfo = fmt.Sprintf("%s %s", resp.Request.Method, resp.Request.URL.String())
	}

	// Try to extract a clean human message from the body
	var message string
	if resp != nil && resp.Body != nil {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) > 0 {
			var anyJSON any
			if json.Unmarshal(bodyBytes, &anyJSON) == nil {
				if msg := bestJSONMessage(anyJSON); msg != "" {
					message = msg
				} else {
					message = strings.TrimSpace(string(bodyBytes))
				}
			} else {
				message = strings.TrimSpace(string(bodyBytes))
			}
		}
	}

	// Fallback to the original error text if we didn't get a body message
	if message == "" && err != nil && statusLine != err.Error() {
		message = err.Error()
	}

	// Append request info at the end of the message line
	if reqInfo != "" {
		if message == "" {
			message = fmt.Sprintf("(%s)", reqInfo)
		} else {
			message = fmt.Sprintf("%s (%s)", message, reqInfo)
		}
	}

	// Produce the final two-line string
	switch {
	case statusLine != "" && message != "":
		return statusLine + "\n" + message
	case statusLine != "":
		return statusLine
	default:
		return message
	}
}

// bestJSONMessage tries to extract a single, human-friendly sentence from common DRF error shapes,
// e.g. {"detail":"..."} or {"name":["..."]} or {"non_field_errors":["..."]}. Field names are stripped.
func bestJSONMessage(v any) string {
	// 1) Exact DRF "detail"
	if m, ok := v.(map[string]any); ok {
		if d, ok := m["detail"]; ok {
			if s := stringify(d); strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}

	// Prefer non-field errors if present
	if m, ok := v.(map[string]any); ok {
		if nfe, ok := m["non_field_errors"]; ok {
			msgs := collectStrings(nfe)
			if len(msgs) > 0 {
				return strings.Join(msgs, " | ")
			}
		}
	}

	// Single field with messages -> return the first message only
	if m, ok := v.(map[string]any); ok {
		all := []string{}
		for _, val := range m {
			all = append(all, collectStrings(val)...)
		}
		if len(all) == 1 {
			return strings.TrimSpace(all[0])
		}
		if len(all) > 1 {
			return strings.Join(all, " | ")
		}
	}

	// Top-level list of messages
	if arr, ok := v.([]any); ok {
		msgs := collectStrings(arr)
		if len(msgs) > 0 {
			return strings.Join(msgs, " | ")
		}
	}

	return ""
}

// collectStrings walks typical DRF error shapes and returns only the message strings,
// ignoring field names (so {"name":["msg"]} -> ["msg"]).
func collectStrings(v any) []string {
	out := []string{}
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		if s != "" {
			out = append(out, s)
		}
	case []any:
		for _, it := range t {
			out = append(out, collectStrings(it)...)
		}
	case map[string]any:
		for _, it := range t {
			out = append(out, collectStrings(it)...)
		}
	}
	return out
}

func stringify(v any) string {
	switch s := v.(type) {
	case string:
		return s
	case fmt.Stringer:
		return s.String()
	default:
		b, _ := json.Marshal(s)
		return string(b)
	}
}

// derefStr safely dereferences a *string, returning "" if nil.
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// nullableTimeStr formats a NullableTime as RFC3339 types.String.
// Returns types.StringNull() when the value is not set or nil.
func nullableTimeStr(t nb.NullableTime) types.String {
	if t.IsSet() && t.Get() != nil {
		return types.StringValue(t.Get().Format(time.RFC3339))
	}
	return types.StringNull()
}

// nullableFKStr extracts the UUID string from a NullableApprovalWorkflowUser FK.
// Returns types.StringValue("") when not set or the nested ID is absent.
func nullableFKStr(n nb.NullableApprovalWorkflowUser) types.String {
	if n.IsSet() {
		if v := n.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			return types.StringValue(*v.Id.String)
		}
	}
	return types.StringValue("")
}

// nullableSoftwareVersionStr extracts the UUID from a nullable VM software version.
// Returns types.StringValue("") when not set or the nested ID is absent.
func nullableSoftwareVersionStr(n nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion) types.String {
	if n.IsSet() {
		if v := n.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			return types.StringValue(*v.Id.String)
		}
	}
	return types.StringValue("")
}

// makeFKUser builds a NullableApprovalWorkflowUser from a UUID string.
// An empty id produces a set-but-nil FK (clears the relation on PATCH).
func makeFKUser(id string) nb.NullableApprovalWorkflowUser {
	var fk nb.NullableApprovalWorkflowUser
	if id == "" {
		fk.Set(nil)
		return fk
	}
	fk.Set(&nb.ApprovalWorkflowUser{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(id),
		},
	})
	return fk
}

func sliceToSet(in []string) map[string]struct{} {
	s := make(map[string]struct{}, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		s[v] = struct{}{}
	}
	return s
}

// setDiff returns items in A that are not in B.
func setDiff(a, b map[string]struct{}) []string {
	out := make([]string, 0, len(a))
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}
