package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ resource.Resource = &VirtualMachineResource{}
var _ resource.ResourceWithImportState = &VirtualMachineResource{}

type VirtualMachineResource struct {
	client *APIClient
}

type vmModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	Status            types.String `tfsdk:"status"`
	Vcpus             types.Int64  `tfsdk:"vcpus"`
	Memory            types.Int64  `tfsdk:"memory"`
	Disk              types.Int64  `tfsdk:"disk"`
	Comments          types.String `tfsdk:"comments"`
	TenantID          types.String `tfsdk:"tenant_id"`
	PlatformID        types.String `tfsdk:"platform_id"`
	RoleID            types.String `tfsdk:"role_id"`
	SoftwareVersionID types.String `tfsdk:"software_version_id"`
	TagsIDs           types.List   `tfsdk:"tags_ids"`
	Created           types.String `tfsdk:"created"`
}

func NewVirtualMachineResource() resource.Resource {
	return &VirtualMachineResource{}
}

func (r *VirtualMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *VirtualMachineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a virtual machine in Nautobot",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Virtual Machine UUID.",
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Virtual Machine's name.",
			},
			"cluster_id": rschema.StringAttribute{
				Required:    true,
				Description: "Cluster where the virtual machine belongs.",
			},
			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status of the virtual machine (name).",
			},

			"vcpus": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of virtual CPUs.",
				Default:     int64default.StaticInt64(0),
			},
			"memory": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Amount of memory in MB.",
				Default:     int64default.StaticInt64(0),
			},
			"disk": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Disk size in GB.",
				Default:     int64default.StaticInt64(0),
			},
			"comments": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Comments or notes about the virtual machine.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant associated with the virtual machine.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Platform or OS installed on the virtual machine.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Role of the virtual machine.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"software_version_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Software version installed on the virtual machine.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tags_ids": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tags associated with the virtual machine.",
				ElementType: types.StringType,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VirtualMachineResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *VirtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmModel
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

	var vm nb.VirtualMachineRequest
	vm.Name = plan.Name.ValueString()

	vm.Cluster = nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(plan.ClusterID.ValueString()),
		},
	}

	vm.Status = nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(statusID),
		},
	}

	if !plan.Vcpus.IsNull() {
		vm.Vcpus.Set(int32Ptr(int(plan.Vcpus.ValueInt64())))
	}
	if !plan.Memory.IsNull() {
		vm.Memory.Set(int32Ptr(int(plan.Memory.ValueInt64())))
	}
	if !plan.Disk.IsNull() {
		vm.Disk.Set(int32Ptr(int(plan.Disk.ValueInt64())))
	}

	if !plan.Comments.IsNull() {
		cm := plan.Comments.ValueString()
		vm.Comments = &cm
	}

	if plan.TenantID.ValueString() != "" {
		vm.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	if plan.PlatformID.ValueString() != "" {
		vm.Platform = makeFKUser(plan.PlatformID.ValueString())
	}

	if plan.RoleID.ValueString() != "" {
		vm.Role = makeFKUser(plan.RoleID.ValueString())
	}

	if plan.SoftwareVersionID.ValueString() != "" {
		var sv nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
		sv.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(plan.SoftwareVersionID.ValueString()),
			},
		})
		vm.SoftwareVersion = sv
	}

	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if !resp.Diagnostics.HasError() && len(tagIDs) > 0 {
			tags := make([]nb.ApprovalWorkflowStageResponseApprovalWorkflowStage, 0, len(tagIDs))
			for _, t := range tagIDs {
				if t == "" {
					continue
				}
				tags = append(tags, nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
					Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
						String: stringPtr(t),
					},
				})
			}
			vm.Tags = tags
		}
	}

	out, httpResp, err := c.VirtualizationAPI.
		VirtualizationVirtualMachinesCreate(ctx).
		VirtualMachineRequest(vm).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create virtual machine", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created virtual machine returned no id")
		return
	}

	model, found, diags := r.buildStateModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read virtual machine", "created virtual machine was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VirtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	if id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	model, found, diags := r.buildStateModel(ctx, id)
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

func (r *VirtualMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vmModel
	var state vmModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vmId := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedVirtualMachineRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}

	if !plan.ClusterID.Equal(state.ClusterID) {
		v := plan.ClusterID.ValueString()
		patch.Cluster = &nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: &v,
			},
		}
	}

	if !plan.Status.Equal(state.Status) {
		statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}
		patch.Status = &nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(statusID),
			},
		}
	}

	if !plan.Vcpus.Equal(state.Vcpus) {
		patch.Vcpus.Set(int32Ptr(int(plan.Vcpus.ValueInt64())))
	}
	if !plan.Memory.Equal(state.Memory) {
		patch.Memory.Set(int32Ptr(int(plan.Memory.ValueInt64())))
	}
	if !plan.Disk.Equal(state.Disk) {
		patch.Disk.Set(int32Ptr(int(plan.Disk.ValueInt64())))
	}

	if !plan.Comments.Equal(state.Comments) {
		v := plan.Comments.ValueString()
		patch.Comments = &v
	}

	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	if !plan.PlatformID.Equal(state.PlatformID) {
		patch.Platform = makeFKUser(plan.PlatformID.ValueString())
	}

	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = makeFKUser(plan.RoleID.ValueString())
	}

	if !plan.SoftwareVersionID.Equal(state.SoftwareVersionID) {
		if plan.SoftwareVersionID.ValueString() == "" {
			patch.SoftwareVersion.Set(nil)
		} else {
			var sv nb.NullableBulkWritableVirtualMachineRequestSoftwareVersion
			sv.Set(&nb.BulkWritableVirtualMachineRequestSoftwareVersion{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(plan.SoftwareVersionID.ValueString()),
				},
			})
			patch.SoftwareVersion = sv
		}
	}

	if !plan.TagsIDs.Equal(state.TagsIDs) {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tags := make([]nb.ApprovalWorkflowStageResponseApprovalWorkflowStage, 0, len(tagIDs))
		for _, t := range tagIDs {
			if t == "" {
				continue
			}
			tags = append(tags, nb.ApprovalWorkflowStageResponseApprovalWorkflowStage{
				Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
					String: stringPtr(t),
				},
			})
		}
		patch.Tags = tags
	}

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationVirtualMachinesPartialUpdate(ctx, vmId).
		PatchedVirtualMachineRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update virtual machine", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.buildStateModel(ctx, vmId)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read virtual machine", "updated virtual machine was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VirtualMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationVirtualMachinesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete virtual machine", httpErr(err, httpResp))
		return
	}
}

func (r *VirtualMachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VirtualMachineResource) buildStateModel(ctx context.Context, id string) (vmModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	vm, httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationVirtualMachinesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return vmModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read virtual machine", httpErr(err, httpResp))
		return vmModel{}, false, diags
	}

	var m vmModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(vm.Name)

	clusterID := ""
	if vm.Cluster.Id != nil && vm.Cluster.Id.String != nil {
		clusterID = *vm.Cluster.Id.String
	}
	m.ClusterID = types.StringValue(clusterID)

	statusName := ""
	if vm.Status.Id != nil && vm.Status.Id.String != nil {
		if *vm.Status.Id.String != "" {
			if n, err := getStatusName(ctx, r.client.Client, *vm.Status.Id.String); err == nil {
				statusName = n
			}
		}
	}
	m.Status = types.StringValue(statusName)

	if vm.Vcpus.IsSet() && vm.Vcpus.Get() != nil {
		m.Vcpus = types.Int64Value(int64(*vm.Vcpus.Get()))
	} else {
		m.Vcpus = types.Int64Value(0)
	}
	if vm.Memory.IsSet() && vm.Memory.Get() != nil {
		m.Memory = types.Int64Value(int64(*vm.Memory.Get()))
	} else {
		m.Memory = types.Int64Value(0)
	}
	if vm.Disk.IsSet() && vm.Disk.Get() != nil {
		m.Disk = types.Int64Value(int64(*vm.Disk.Get()))
	} else {
		m.Disk = types.Int64Value(0)
	}

	m.Comments = types.StringValue(derefStr(vm.Comments))

	m.TenantID = nullableFKStr(vm.Tenant)
	m.PlatformID = nullableFKStr(vm.Platform)
	m.RoleID = nullableFKStr(vm.Role)

	m.SoftwareVersionID = nullableSoftwareVersionStr(vm.SoftwareVersion)

	if len(vm.Tags) > 0 {
		vals := make([]attr.Value, 0, len(vm.Tags))
		for _, t := range vm.Tags {
			if t.Id != nil && t.Id.String != nil {
				vals = append(vals, types.StringValue(*t.Id.String))
			}
		}
		m.TagsIDs = types.ListValueMust(types.StringType, vals)
	} else {
		m.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	m.Created = nullableTimeStr(vm.Created)

	tflog.Debug(ctx, "read VM", map[string]any{"id": id})
	return m, true, diags
}
