package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ resource.Resource = &VLANResource{}
var _ resource.ResourceWithImportState = &VLANResource{}

type VLANResource struct {
	client *APIClient
}

type vlanModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Vid         types.Int64  `tfsdk:"vid"`
	Description types.String `tfsdk:"description"`
	VLANGroupID types.String `tfsdk:"vlan_group_id"`
	Status      types.String `tfsdk:"status"`
	TenantID    types.String `tfsdk:"tenant_id"`
	RoleID      types.String `tfsdk:"role_id"`
	TagsIDs     types.List   `tfsdk:"tags_ids"`

	Created     types.String `tfsdk:"created"`
	PrefixCount types.Int64  `tfsdk:"prefix_count"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewVLANResource() resource.Resource {
	return &VLANResource{}
}

func (r *VLANResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (r *VLANResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a VLAN in Nautobot",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "VLAN UUID.",
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "VLAN name.",
			},

			"vid": rschema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 4094),
				},
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status of the VLAN (name).",
			},

			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the VLAN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"vlan_group_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "VLAN group UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tenant_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant UUID associated with the VLAN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"role_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Role UUID associated with the VLAN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tags_ids": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tags associated with the VLAN.",
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

			"prefix_count": rschema.Int64Attribute{
				Computed:    true,
				Description: "Number of prefixes associated with this VLAN.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Human-friendly display value.",
			},

			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "API URL of the VLAN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the VLAN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the VLAN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VLANResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *VLANResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanModel
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

	var vlan nb.VLANRequest
	vlan.Name = plan.Name.ValueString()
	vlan.Vid = int32(plan.Vid.ValueInt64())

	vlan.Status = nb.BulkWritableCableRequestStatus{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(statusID),
		},
	}

	if !plan.Description.IsNull() {
		d := plan.Description.ValueString()
		vlan.Description = &d
	}

	if plan.VLANGroupID.ValueString() != "" {
		vlan.VlanGroup = makeFKUser(plan.VLANGroupID.ValueString())
	}

	if plan.TenantID.ValueString() != "" {
		vlan.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	if plan.RoleID.ValueString() != "" {
		vlan.Role = makeFKUser(plan.RoleID.ValueString())
	}

	if !plan.TagsIDs.IsNull() && !plan.TagsIDs.IsUnknown() {
		var tagIDs []string
		resp.Diagnostics.Append(plan.TagsIDs.ElementsAs(ctx, &tagIDs, false)...)
		if !resp.Diagnostics.HasError() && len(tagIDs) > 0 {
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
			vlan.Tags = tags
		}
	}

	out, httpResp, err := c.IpamAPI.
		IpamVlansCreate(ctx).
		VLANRequest(vlan).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create vlan", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created vlan returned no id")
		return
	}

	model, found, diags := r.buildStateModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read vlan", "created vlan was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VLANResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanModel
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

func (r *VLANResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vlanModel
	var state vlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vlanID := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedVLANRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}

	if !plan.Vid.Equal(state.Vid) {
		v := int32(plan.Vid.ValueInt64())
		patch.Vid = &v
	}

	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}

	if !plan.Status.Equal(state.Status) {
		statusID, err := getStatusID(ctx, c, plan.Status.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("failed to get status id", err.Error())
			return
		}
		patch.Status = &nb.BulkWritableCableRequestStatus{
			Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
				String: stringPtr(statusID),
			},
		}
	}

	if !plan.VLANGroupID.Equal(state.VLANGroupID) {
		patch.VlanGroup = makeFKUser(plan.VLANGroupID.ValueString())
	}

	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	if !plan.RoleID.Equal(state.RoleID) {
		patch.Role = makeFKUser(plan.RoleID.ValueString())
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

	_, httpResp, err := c.IpamAPI.
		IpamVlansPartialUpdate(ctx, vlanID).
		PatchedVLANRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update vlan", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.buildStateModel(ctx, vlanID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read vlan", "updated vlan was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *VLANResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamVlansDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete vlan", httpErr(err, httpResp))
		return
	}
}

func (r *VLANResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VLANResource) buildStateModel(ctx context.Context, id string) (vlanModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	v, httpResp, err := r.client.Client.IpamAPI.
		IpamVlansRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return vlanModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read vlan", httpErr(err, httpResp))
		return vlanModel{}, false, diags
	}

	var m vlanModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(v.Name)
	m.Vid = types.Int64Value(int64(v.Vid))

	m.Description = types.StringValue(derefStr(v.Description))

	m.VLANGroupID = nullableFKStr(v.VlanGroup)

	statusName := ""
	if v.Status.Id != nil && v.Status.Id.String != nil && *v.Status.Id.String != "" {
		if n, err := getStatusName(ctx, r.client.Client, *v.Status.Id.String); err == nil {
			statusName = n
		}
	}
	m.Status = types.StringValue(statusName)

	m.TenantID = nullableFKStr(v.Tenant)
	m.RoleID = nullableFKStr(v.Role)

	if len(v.Tags) > 0 {
		vals := make([]attr.Value, 0, len(v.Tags))
		for _, t := range v.Tags {
			if t.Id != nil && t.Id.String != nil {
				vals = append(vals, types.StringValue(*t.Id.String))
			}
		}
		m.TagsIDs = types.ListValueMust(types.StringType, vals)
	} else {
		m.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	m.Created = nullableTimeStr(v.Created)

	if v.PrefixCount != nil {
		m.PrefixCount = types.Int64Value(int64(*v.PrefixCount))
	} else {
		m.PrefixCount = types.Int64Value(0)
	}

	m.Display = types.StringValue(v.Display)
	m.URL = types.StringValue(v.Url)
	m.NaturalSlug = types.StringValue(v.NaturalSlug)
	m.NotesURL = types.StringValue(v.NotesUrl)

	tflog.Debug(ctx, "read VLAN", map[string]any{"id": id})
	return m, true, diags
}
