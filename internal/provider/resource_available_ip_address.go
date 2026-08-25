package provider

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &AvailableIPAddressResource{}
	_ resource.ResourceWithImportState = &AvailableIPAddressResource{}
)

type AvailableIPAddressResource struct {
	client *APIClient
}

type availableIPModel struct {
	ID        types.String `tfsdk:"id"`
	PrefixID  types.String `tfsdk:"prefix_id"`
	Address   types.String `tfsdk:"address"`
	IPVersion types.Int64  `tfsdk:"ip_version"`
	DNSName   types.String `tfsdk:"dns_name"`
	Status    types.String `tfsdk:"status"`
}

func NewAvailableIPAddressResource() resource.Resource {
	return &AvailableIPAddressResource{}
}

func (r *AvailableIPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_ip_address"
}

func (r *AvailableIPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object allocates and manages an available IP address in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Allocated IP address UUID.",
			},

			"prefix_id": rschema.StringAttribute{
				Required:    true,
				Description: "ID of the prefix to allocate the IP address from.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"address": rschema.StringAttribute{
				Computed:    true,
				Description: "Allocated IP address in CIDR notation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"ip_version": rschema.Int64Attribute{
				Computed:    true,
				Description: "IP version of the allocated IP address (4 or 6).",
			},

			"dns_name": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "DNS name associated with the IP address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status (by name) of the allocated IP address.",
			},
		},
	}
}

func (r *AvailableIPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

// randomBackoff returns an increasing backoff with jitter.
func randomBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	const (
		base     = 250 * time.Millisecond
		capDelay = 20 * time.Second
	)

	backoff := base << attempt
	if backoff > capDelay {
		backoff = capDelay
	}

	jitterMax := backoff / 2
	if jitterMax <= 0 {
		return backoff
	}
	jitter := time.Duration(rand.Int63n(int64(jitterMax)))

	return backoff + jitter
}

func (r *AvailableIPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan availableIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get status id", err.Error())
		return
	}

	statusRef := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
		String: stringPtr(statusID),
	}
	status := nb.BulkWritableCableRequestStatus{
		Id: statusRef,
	}

	body := []nb.IPAllocationRequest{
		{
			Status: status,
		},
	}
	if v := plan.DNSName.ValueString(); v != "" {
		body[0].DnsName = &v
	}

	// Remove when https://github.com/nautobot/nautobot/issues/8297 is fixed.
	const maxRetries = 10
	var alloc []nb.IPAddress
	var httpResp *http.Response

	for attempt := 0; attempt < maxRetries; attempt++ {
		alloc, httpResp, err = c.IpamAPI.
			IpamPrefixesAvailableIpsCreate(ctx, plan.PrefixID.ValueString()).
			IPAllocationRequest(body).
			Execute()
		if err == nil {
			break
		}

		formattedErr := httpErr(err, httpResp)

		isDuplicate := httpResp != nil &&
			httpResp.StatusCode == http.StatusBadRequest &&
			strings.Contains(formattedErr, "IP address with this Parent and Host already exists")

		if !isDuplicate {
			resp.Diagnostics.AddError("failed to allocate IP address", formattedErr)
			return
		}

		if attempt == maxRetries-1 {
			resp.Diagnostics.AddError("failed to allocate IP address after retries", formattedErr)
			return
		}

		backoff := randomBackoff(attempt)
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddError("failed to allocate IP address", "context cancelled while retrying after duplicate allocation")
			return
		case <-time.After(backoff):
		}
	}

	if len(alloc) == 0 || alloc[0].Id == nil || *alloc[0].Id == "" {
		resp.Diagnostics.AddError("invalid API response", "no IP address id returned from allocation")
		return
	}

	model, found, diags := r.readModel(ctx, *alloc[0].Id, plan.PrefixID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read IP address", "created IP address was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AvailableIPAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state availableIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.readModel(ctx, id, state.PrefixID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AvailableIPAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state availableIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedIPAddressRequest

	if !plan.Status.Equal(state.Status) {
		statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}

		statusRef := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(statusID),
		}
		status := nb.BulkWritableCableRequestStatus{
			Id: statusRef,
		}
		patch.Status = &status
	}

	if !plan.DNSName.Equal(state.DNSName) {
		if plan.DNSName.ValueString() == "" {
			empty := ""
			patch.DnsName = &empty
		} else {
			v := plan.DNSName.ValueString()
			patch.DnsName = &v
		}
	}

	_, httpResp, err := c.IpamAPI.
		IpamIpAddressesPartialUpdate(ctx, id).
		PatchedIPAddressRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update IP address", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id, state.PrefixID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read IP address", "updated IP address was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AvailableIPAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state availableIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete IP address", httpErr(err, httpResp))
		return
	}
}

func (r *AvailableIPAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *AvailableIPAddressResource) readModel(ctx context.Context, id string, prefixID string) (availableIPModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ip, httpResp, err := r.client.Client.IpamAPI.
		IpamIpAddressesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return availableIPModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read IP address", httpErr(err, httpResp))
		return availableIPModel{}, false, diags
	}

	// Try to derive prefixID from parent if not provided
	if prefixID == "" && ip.Parent.IsSet() {
		parent := ip.Parent.Get()
		if parent != nil && parent.Id != nil && parent.Id.String != nil && *parent.Id.String != "" {
			prefixID = *parent.Id.String
		}
	}

	var m availableIPModel
	m.ID = types.StringValue(id)
	m.PrefixID = types.StringValue(prefixID)

	m.Address = types.StringValue(ip.Address)
	m.IPVersion = types.Int64Value(int64(ip.IpVersion))

	if ip.DnsName != nil {
		m.DNSName = types.StringValue(*ip.DnsName)
	} else {
		m.DNSName = types.StringValue("")
	}

	statusName := ""
	if ip.Status.Id != nil && ip.Status.Id.String != nil && *ip.Status.Id.String != "" {
		if n, err := getStatusName(ctx, r.client.Client, *ip.Status.Id.String); err == nil {
			statusName = n
		}
	}
	m.Status = types.StringValue(statusName)

	return m, true, diags
}
