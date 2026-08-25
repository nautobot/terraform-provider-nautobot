package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &VMInterfaceResource{}
	_ resource.ResourceWithImportState = &VMInterfaceResource{}
)

type VMInterfaceResource struct {
	client *APIClient
}

type vmInterfaceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	MacAddress       types.String `tfsdk:"mac_address"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	MTU              types.Int64  `tfsdk:"mtu"`
	Mode             types.String `tfsdk:"mode"`
	Description      types.String `tfsdk:"description"`
	Status           types.String `tfsdk:"status"`
	VirtualMachineID types.String `tfsdk:"virtual_machine_id"`
	UntaggedVlanID   types.String `tfsdk:"untagged_vlan_id"`
	TagsIDs          types.List   `tfsdk:"tags_ids"`
	IPAddresses      types.Set    `tfsdk:"ip_addresses"`
	Created          types.String `tfsdk:"created"`
}

func NewVMInterfaceResource() resource.Resource {
	return &VMInterfaceResource{}
}

func (r *VMInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_interface"
}

func (r *VMInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a VM Interface in Nautobot",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "VM Interface UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Name of the VM interface.",
			},
			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status of the VM interface (name).",
			},
			"virtual_machine_id": rschema.StringAttribute{
				Required:    true,
				Description: "ID of the virtual machine to which the interface belongs.",
			},
			"mac_address": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "MAC address of the interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^([0-9A-F]{2}:){5}[0-9A-F]{2}$`),
						"must be an uppercase MAC address with 6 octets, e.g. AA:BB:CC:DD:EE:FF",
					),
				},
			},
			"enabled": rschema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the interface is enabled.",
			},
			"mtu": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "MTU size of the interface.",
			},
			"mode": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Mode of the interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"untagged_vlan_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Untagged VLAN ID associated with the interface.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags_ids": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				Description: "Tags associated with the interface.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_addresses": rschema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default:     setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
				Description: "List of IP address IDs to assign to the VM interface.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the interface (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VMInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func vmInterfaceModeFromString(v string) (*nb.PatchedWritableInterfaceRequestMode, error) {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		mode := nb.PATCHEDWRITABLEINTERFACEREQUESTMODE_EMPTY
		return &mode, nil
	}

	switch s {
	case "access":
		mode := nb.PATCHEDWRITABLEINTERFACEREQUESTMODE_ACCESS
		return &mode, nil
	case "tagged":
		mode := nb.PATCHEDWRITABLEINTERFACEREQUESTMODE_TAGGED
		return &mode, nil
	case "tagged-all", "tagged_all", "tagged all":
		mode := nb.PATCHEDWRITABLEINTERFACEREQUESTMODE_TAGGED_ALL
		return &mode, nil
	default:
		return nil, fmt.Errorf("unsupported mode %q (valid values: access, tagged, tagged-all)", v)
	}
}

func (r *VMInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.UntaggedVlanID.ValueString() != "" && plan.Mode.ValueString() == "" {
		resp.Diagnostics.AddError(
			"invalid VM interface configuration",
			"mode must be set when untagged_vlan_id is specified",
		)
		return
	}

	c := r.client.Client

	statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to get status id", err.Error())
		return
	}

	var body nb.WritableVMInterfaceRequest
	body.Name = plan.Name.ValueString()

	body.Status = nb.BulkWritableCableRequestStatus{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(statusID),
		},
	}

	body.VirtualMachine = nb.BulkWritableCableRequestStatus{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(plan.VirtualMachineID.ValueString()),
		},
	}

	if v := plan.MacAddress.ValueString(); v != "" {
		mac := strings.ToUpper(v)
		body.MacAddress.Set(&mac)
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		en := plan.Enabled.ValueBool()
		body.Enabled = &en
	}
	if !plan.MTU.IsNull() && !plan.MTU.IsUnknown() {
		if v := plan.MTU.ValueInt64(); v > 0 {
			mtu := int32(v)
			body.Mtu.Set(&mtu)
		} else {
			body.Mtu.Unset()
		}
	}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}

	if v := plan.Mode.ValueString(); v != "" {
		modeEnum, err := vmInterfaceModeFromString(v)
		if err != nil {
			resp.Diagnostics.AddError("invalid mode", err.Error())
			return
		}
		body.Mode = modeEnum
	}

	if v := plan.UntaggedVlanID.ValueString(); v != "" {
		body.UntaggedVlan = makeFKUser(v)
	}

	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(tagIDs) > 0 {
			tags := make([]nb.BulkWritableCableRequestStatus, 0, len(tagIDs))
			for _, t := range tagIDs {
				if t == "" {
					continue
				}
				tags = append(tags, nb.BulkWritableCableRequestStatus{
					Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
						String: stringPtr(t),
					},
				})
			}
			body.Tags = tags
		}
	}

	created, httpResp, err := c.VirtualizationAPI.
		VirtualizationInterfacesCreate(ctx).
		WritableVMInterfaceRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create VM interface", httpErr(err, httpResp))
		return
	}
	if created.Id == nil || *created.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created VM interface returned no id")
		return
	}

	if !plan.IPAddresses.IsNull() && !plan.IPAddresses.IsUnknown() {
		var ips []string
		resp.Diagnostics.Append(plan.IPAddresses.ElementsAs(ctx, &ips, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, ip := range ips {
			if ip == "" {
				continue
			}
			if err := r.assignIPAddressToVMInterface(ctx, ip, *created.Id); err != nil {
				resp.Diagnostics.AddError("failed to assign IP address to VM interface", err.Error())
				return
			}
		}
	}

	model, found, diags := r.readModel(ctx, *created.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read VM interface", "created VM interface was not found")
		return
	}
	tflog.Debug(ctx, "vm interface created", map[string]any{"id": *created.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VMInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.readModel(ctx, id)
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

func (r *VMInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vmInterfaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.UntaggedVlanID.ValueString() != "" && plan.Mode.ValueString() == "" {
		resp.Diagnostics.AddError(
			"invalid VM interface configuration",
			"mode must be set when untagged_vlan_id is specified",
		)
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedWritableVMInterfaceRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}

	if !plan.MacAddress.Equal(state.MacAddress) {
		v := plan.MacAddress.ValueString()
		if v == "" {
			patch.MacAddress.Unset()
		} else {
			mac := strings.ToUpper(v)
			patch.MacAddress.Set(&mac)
		}
	}

	if !plan.Enabled.Equal(state.Enabled) {
		v := plan.Enabled.ValueBool()
		patch.Enabled = &v
	}

	if !plan.MTU.Equal(state.MTU) {
		v := plan.MTU.ValueInt64()
		if v > 0 {
			mtu := int32(v)
			patch.Mtu.Set(&mtu)
		} else {
			patch.Mtu.Set(nil)
		}
	}

	if !plan.Description.Equal(state.Description) {
		if plan.Description.ValueString() == "" {
			empty := ""
			patch.Description = &empty
		} else {
			v := plan.Description.ValueString()
			patch.Description = &v
		}
	}

	if !plan.Status.Equal(state.Status) {
		statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}
		statusRef := nb.BulkWritableCableRequestStatus{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(statusID),
			},
		}
		patch.Status = &statusRef
	}

	if !plan.VirtualMachineID.Equal(state.VirtualMachineID) {
		vmid := plan.VirtualMachineID.ValueString()
		vmRef := nb.BulkWritableCableRequestStatus{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(vmid),
			},
		}
		patch.VirtualMachine = &vmRef
	}

	if !plan.Mode.Equal(state.Mode) {
		v := plan.Mode.ValueString()
		modeEnum, err := vmInterfaceModeFromString(v)
		if err != nil {
			resp.Diagnostics.AddError("invalid mode", err.Error())
			return
		}
		patch.Mode = modeEnum
	}

	if !plan.UntaggedVlanID.Equal(state.UntaggedVlanID) {
		patch.UntaggedVlan = makeFKUser(plan.UntaggedVlanID.ValueString())
	}

	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tags := make([]nb.BulkWritableCableRequestStatus, 0, len(tagIDs))
		for _, t := range tagIDs {
			if t == "" {
				continue
			}
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(t),
				},
			})
		}
		patch.Tags = tags
	}

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationInterfacesPartialUpdate(ctx, id).
		PatchedWritableVMInterfaceRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update VM interface", httpErr(err, httpResp))
		return
	}

	{
		var desiredIPs []string
		resp.Diagnostics.Append(plan.IPAddresses.ElementsAs(ctx, &desiredIPs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		currentModel, found, diags := r.readModel(ctx, id)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !found {
			resp.Diagnostics.AddError("failed to read VM interface", "updated VM interface was not found")
			return
		}

		var currentIPs []string
		resp.Diagnostics.Append(currentModel.IPAddresses.ElementsAs(ctx, &currentIPs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		currentSet := sliceToSet(currentIPs)
		desiredSet := sliceToSet(desiredIPs)

		toRemove := setDiff(currentSet, desiredSet)
		toAdd := setDiff(desiredSet, currentSet)

		for _, ip := range toRemove {
			if err := r.removeIPAddressFromVMInterface(ctx, ip, id); err != nil {
				resp.Diagnostics.AddError("failed removing IP from VM interface", err.Error())
				return
			}
		}
		for _, ip := range toAdd {
			if err := r.assignIPAddressToVMInterface(ctx, ip, id); err != nil {
				resp.Diagnostics.AddError("failed assigning IP to VM interface", err.Error())
				return
			}
		}
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read VM interface", "updated VM interface was not found")
		return
	}
	tflog.Debug(ctx, "vm interface updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VMInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmInterfaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationInterfacesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete VM interface", httpErr(err, httpResp))
		return
	}
}

func (r *VMInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VMInterfaceResource) readModel(ctx context.Context, id string) (vmInterfaceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ifc, httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationInterfacesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return vmInterfaceModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read VM interface", httpErr(err, httpResp))
		return vmInterfaceModel{}, false, diags
	}

	var m vmInterfaceModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(ifc.Name)

	if ifc.MacAddress.IsSet() && ifc.MacAddress.Get() != nil {
		m.MacAddress = types.StringValue(strings.ToUpper(*ifc.MacAddress.Get()))
	} else {
		m.MacAddress = types.StringValue("")
	}

	if ifc.Enabled != nil {
		m.Enabled = types.BoolValue(*ifc.Enabled)
	} else {
		m.Enabled = types.BoolValue(false)
	}

	if ifc.Mtu.IsSet() && ifc.Mtu.Get() != nil {
		m.MTU = types.Int64Value(int64(*ifc.Mtu.Get()))
	} else {
		m.MTU = types.Int64Value(0)
	}

	m.Description = types.StringValue(derefStr(ifc.Description))

	statusName := ""
	if ifc.Status.Id != nil && ifc.Status.Id.String != nil {
		if *ifc.Status.Id.String != "" {
			if n, err := getStatusName(ctx, r.client.Client, *ifc.Status.Id.String); err == nil {
				statusName = n
			}
		}
	}
	m.Status = types.StringValue(statusName)

	vmID := ""
	if ifc.VirtualMachine.Id != nil && ifc.VirtualMachine.Id.String != nil {
		vmID = *ifc.VirtualMachine.Id.String
	}
	m.VirtualMachineID = types.StringValue(vmID)

	m.UntaggedVlanID = nullableFKStr(ifc.UntaggedVlan)

	if len(ifc.Tags) > 0 {
		vals := make([]attr.Value, 0, len(ifc.Tags))
		for _, t := range ifc.Tags {
			if t.Id != nil && t.Id.String != nil {
				vals = append(vals, types.StringValue(*t.Id.String))
			}
		}
		m.TagsIDs = types.ListValueMust(types.StringType, vals)
	} else {
		m.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	{
		ipRels, httpResp2, err2 := r.client.Client.IpamAPI.
			IpamIpAddressToInterfaceList(ctx).
			VmInterface([]string{id}).
			Execute()
		if err2 != nil {
			diags.AddError("failed to list IP assignments for VM interface", httpErr(err2, httpResp2))
			return vmInterfaceModel{}, false, diags
		}

		if len(ipRels.Results) > 0 {
			vals := make([]attr.Value, 0, len(ipRels.Results))
			for _, rel := range ipRels.Results {
				if rel.IpAddress.Id != nil && rel.IpAddress.Id.String != nil {
					vals = append(vals, types.StringValue(*rel.IpAddress.Id.String))
				}
			}
			m.IPAddresses = types.SetValueMust(types.StringType, vals)
		} else {
			m.IPAddresses = types.SetValueMust(types.StringType, []attr.Value{})
		}
	}

	mode := ""
	if ifc.Mode != nil {
		if ifc.Mode.Label != nil && *ifc.Mode.Label != "" {
			mode = string(*ifc.Mode.Label)
		} else if ifc.Mode.Value != nil && *ifc.Mode.Value != "" {
			mode = string(*ifc.Mode.Value)
		}
	}
	m.Mode = types.StringValue(mode)

	m.Created = nullableTimeStr(ifc.Created)

	tflog.Debug(ctx, "read VM interface", map[string]any{"id": id})
	return m, true, diags
}

func (r *VMInterfaceResource) assignIPAddressToVMInterface(ctx context.Context, ipAddressID, vmInterfaceID string) error {
	c := r.client.Client

	ipID := &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
		String: &ipAddressID,
	}
	ipRef := nb.BulkWritableCableRequestStatus{
		Id: ipID,
	}

	req := nb.IPAddressToInterfaceRequest{
		IpAddress:   ipRef,
		VmInterface: makeFKUser(vmInterfaceID),
	}

	_, httpResp, err := c.IpamAPI.
		IpamIpAddressToInterfaceCreate(ctx).
		IPAddressToInterfaceRequest(req).
		Execute()
	if err != nil {
		return fmt.Errorf("%s", httpErr(err, httpResp))
	}
	return nil
}

func (r *VMInterfaceResource) removeIPAddressFromVMInterface(ctx context.Context, ipAddressID, vmInterfaceID string) error {
	c := r.client.Client

	list, httpResp, err := c.IpamAPI.
		IpamIpAddressToInterfaceList(ctx).
		IpAddress([]string{ipAddressID}).
		VmInterface([]string{vmInterfaceID}).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to list IPAddressToInterface: %s", httpErr(err, httpResp))
	}

	if len(list.Results) == 0 || list.Results[0].Id == nil || *list.Results[0].Id == "" {
		return fmt.Errorf("no assignment found for IP %s on VM interface %s", ipAddressID, vmInterfaceID)
	}

	assignID := *list.Results[0].Id

	httpResp, err = c.IpamAPI.IpamIpAddressToInterfaceDestroy(ctx, assignID).Execute()
	if err != nil {
		return fmt.Errorf("%s", httpErr(err, httpResp))
	}
	return nil
}
