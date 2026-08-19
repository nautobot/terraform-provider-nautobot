package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &VMPrimaryIPResource{}
	_ resource.ResourceWithImportState = &VMPrimaryIPResource{}
)

type VMPrimaryIPResource struct {
	client *APIClient
}

type VMPrimaryIPModel struct {
	ID               types.String `tfsdk:"id"`
	VirtualMachineID types.String `tfsdk:"virtual_machine_id"`
	PrimaryIP4ID     types.String `tfsdk:"primary_ip4_id"`
	PrimaryIP6ID     types.String `tfsdk:"primary_ip6_id"`
}

func NewVMPrimaryIPResource() resource.Resource {
	return &VMPrimaryIPResource{}
}

func (r *VMPrimaryIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_primary_ip"
}

func (r *VMPrimaryIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This resource sets an IP address as the primary IPv4 or IPv6 for a virtual machine in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Resource ID (same as virtual_machine_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"virtual_machine_id": rschema.StringAttribute{
				Required:    true,
				Description: "ID of the virtual machine.",
			},

			"primary_ip4_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "ID of the primary IPv4 address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"primary_ip6_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "ID of the primary IPv6 address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VMPrimaryIPResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *VMPrimaryIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VMPrimaryIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := plan.VirtualMachineID.ValueString()
	c := r.client.Client

	var patch nb.PatchedVirtualMachineRequest

	if v := strings.TrimSpace(plan.PrimaryIP4ID.ValueString()); v != "" {
		var n4 nb.NullablePrimaryIPv4
		n4.Set(&nb.PrimaryIPv4{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(v),
			},
		})
		patch.PrimaryIp4 = n4
	}

	if v := strings.TrimSpace(plan.PrimaryIP6ID.ValueString()); v != "" {
		var n6 nb.NullablePrimaryIPv6
		n6.Set(&nb.PrimaryIPv6{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(v),
			},
		})
		patch.PrimaryIp6 = n6
	}

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationVirtualMachinesPartialUpdate(ctx, vmID).
		PatchedVirtualMachineRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to set primary IP address for virtual machine", httpErr(err, httpResp))
		return
	}

	_ = resp.State.SetAttribute(ctx, path.Root("id"), vmID)

	model, found, diags := r.readModel(ctx, vmID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read virtual machine", "virtual machine was not found after setting primary IP addresses")
		return
	}
	tflog.Debug(ctx, "primary IPs set for VM", map[string]any{"vm_id": vmID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VMPrimaryIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VMPrimaryIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueString()
	if vmID == "" {
		vmID = state.VirtualMachineID.ValueString()
	}
	if vmID == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.readModel(ctx, vmID)
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

func (r *VMPrimaryIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state VMPrimaryIPModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueString()
	if vmID == "" {
		vmID = plan.VirtualMachineID.ValueString()
	}
	c := r.client.Client

	var patch nb.PatchedVirtualMachineRequest

	if !plan.PrimaryIP4ID.Equal(state.PrimaryIP4ID) {
		v := strings.TrimSpace(plan.PrimaryIP4ID.ValueString())
		if v == "" {
			var n4 nb.NullablePrimaryIPv4
			n4.Set(nil)
			patch.PrimaryIp4 = n4
		} else {
			var n4 nb.NullablePrimaryIPv4
			n4.Set(&nb.PrimaryIPv4{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(v),
				},
			})
			patch.PrimaryIp4 = n4
		}
	}

	if !plan.PrimaryIP6ID.Equal(state.PrimaryIP6ID) {
		v := strings.TrimSpace(plan.PrimaryIP6ID.ValueString())
		if v == "" {
			var n6 nb.NullablePrimaryIPv6
			n6.Set(nil)
			patch.PrimaryIp6 = n6
		} else {
			var n6 nb.NullablePrimaryIPv6
			n6.Set(&nb.PrimaryIPv6{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(v),
				},
			})
			patch.PrimaryIp6 = n6
		}
	}

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationVirtualMachinesPartialUpdate(ctx, vmID).
		PatchedVirtualMachineRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update primary IP address for virtual machine", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, vmID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read virtual machine", "virtual machine was not found after updating primary IP addresses")
		return
	}
	tflog.Debug(ctx, "primary IPs updated for VM", map[string]any{"vm_id": vmID})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VMPrimaryIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VMPrimaryIPModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmID := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedVirtualMachineRequest
	var n4 nb.NullablePrimaryIPv4
	var n6 nb.NullablePrimaryIPv6
	n4.Set(nil)
	n6.Set(nil)
	patch.PrimaryIp4 = n4
	patch.PrimaryIp6 = n6

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationVirtualMachinesPartialUpdate(ctx, vmID).
		PatchedVirtualMachineRequest(patch).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to remove primary IP address for virtual machine", httpErr(err, httpResp))
		return
	}
}

func (r *VMPrimaryIPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMPrimaryIPResource) readModel(ctx context.Context, vmID string) (VMPrimaryIPModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	vm, httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationVirtualMachinesRetrieve(ctx, vmID).
		Execute()
	if isNotFoundResponse(httpResp) {
		return VMPrimaryIPModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read virtual machine", httpErr(err, httpResp))
		return VMPrimaryIPModel{}, false, diags
	}

	var out VMPrimaryIPModel
	out.ID = types.StringValue(vmID)
	out.VirtualMachineID = types.StringValue(vmID)

	ip4 := ""
	if vm.PrimaryIp4.IsSet() {
		if v := vm.PrimaryIp4.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			ip4 = *v.Id.String
		}
	}
	out.PrimaryIP4ID = types.StringValue(ip4)

	ip6 := ""
	if vm.PrimaryIp6.IsSet() {
		if v := vm.PrimaryIp6.Get(); v != nil && v.Id != nil && v.Id.String != nil {
			ip6 = *v.Id.String
		}
	}
	out.PrimaryIP6ID = types.StringValue(ip6)

	tflog.Debug(ctx, "read primary IPs for VM", map[string]any{"vm_id": vmID, "ip4": ip4, "ip6": ip6})
	return out, true, diags
}
