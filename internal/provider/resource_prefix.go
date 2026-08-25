package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var _ resource.Resource = &PrefixResource{}
var _ resource.ResourceWithImportState = &PrefixResource{}

type PrefixResource struct {
	client *APIClient
}

type prefixModel struct {
	ID            types.String `tfsdk:"id"`
	VLANID        types.String `tfsdk:"vlan_id"`
	Prefix        types.String `tfsdk:"prefix"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	ParentID      types.String `tfsdk:"parent_id"`
	RoleID        types.String `tfsdk:"role_id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	RirID         types.String `tfsdk:"rir_id"`
	NamespaceID   types.String `tfsdk:"namespace_id"`
	Created       types.String `tfsdk:"created"`
	Network       types.String `tfsdk:"network"`
	Broadcast     types.String `tfsdk:"broadcast"`
	PrefixLength  types.Int64  `tfsdk:"prefix_length"`
	IPVersion     types.Int64  `tfsdk:"ip_version"`
	DateAllocated types.String `tfsdk:"date_allocated"`
	TagsIDs       types.List   `tfsdk:"tags_ids"`
	Display       types.String `tfsdk:"display"`
	URL           types.String `tfsdk:"url"`
	NaturalSlug   types.String `tfsdk:"natural_slug"`
	NotesURL      types.String `tfsdk:"notes_url"`
}

func NewPrefixResource() resource.Resource {
	return &PrefixResource{}
}

func (r *PrefixResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefix"
}

func (r *PrefixResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a prefix in Nautobot",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Prefix UUID.",
			},

			"prefix": rschema.StringAttribute{
				Required:    true,
				Description: "The prefix in CIDR notation (e.g. 10.0.0.0/24).",
			},

			"status": rschema.StringAttribute{
				Required:    true,
				Description: "Status of the prefix (name).",
			},

			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"vlan_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "VLAN UUID associated with this prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"tenant_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant UUID associated with the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"role_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Role UUID associated with the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"parent_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Parent prefix UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"rir_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "RIR UUID associated with the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"date_allocated": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Date this prefix was allocated/reserved (RFC3339).",
			},

			"namespace_id": rschema.StringAttribute{
				Computed:    true,
				Description: "The ID of the namespace associated with the prefix.",
			},

			"tags_ids": rschema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Tags associated with the prefix.",
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

			"network": rschema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 network address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"broadcast": rschema.StringAttribute{
				Computed:    true,
				Description: "IPv4 or IPv6 broadcast address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"prefix_length": rschema.Int64Attribute{
				Computed:    true,
				Description: "Length of the prefix, in bits.",
			},

			"ip_version": rschema.Int64Attribute{
				Computed:    true,
				Description: "IP version (4 or 6).",
			},

			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Human-friendly display value.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "API URL of the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the prefix.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PrefixResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *PrefixResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan prefixModel
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

	var pr nb.WritablePrefixRequest
	pr.Prefix = plan.Prefix.ValueString()
	pr.Status = nb.BulkWritableCableRequestStatus{
		Id: &nb.ApprovalWorkflowApprovalWorkflowDefinitionId{
			String: stringPtr(statusID),
		},
	}

	if !plan.Description.IsNull() {
		v := plan.Description.ValueString()
		pr.Description = &v
	}

	if plan.VLANID.ValueString() != "" {
		pr.Vlan = makeFKUser(plan.VLANID.ValueString())
	}

	if plan.TenantID.ValueString() != "" {
		pr.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	if plan.RoleID.ValueString() != "" {
		pr.Role = makeFKUser(plan.RoleID.ValueString())
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
			pr.Tags = tags
		}
	}

	out, httpResp, err := c.IpamAPI.
		IpamPrefixesCreate(ctx).
		WritablePrefixRequest(pr).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create prefix", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created prefix returned no id")
		return
	}

	model, found, diags := r.buildStateModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read prefix", "created prefix was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PrefixResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state prefixModel
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

func (r *PrefixResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan prefixModel
	var state prefixModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedWritablePrefixRequest

	if !plan.Prefix.Equal(state.Prefix) {
		v := plan.Prefix.ValueString()
		patch.Prefix = &v
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

	if !plan.VLANID.Equal(state.VLANID) {
		patch.Vlan = makeFKUser(plan.VLANID.ValueString())
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
		IpamPrefixesPartialUpdate(ctx, id).
		PatchedWritablePrefixRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update prefix", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.buildStateModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read prefix", "updated prefix was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *PrefixResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state prefixModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamPrefixesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete prefix", httpErr(err, httpResp))
		return
	}
}

func (r *PrefixResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PrefixResource) buildStateModel(ctx context.Context, id string) (prefixModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, httpResp, err := r.client.Client.IpamAPI.
		IpamPrefixesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return prefixModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read prefix", httpErr(err, httpResp))
		return prefixModel{}, false, diags
	}

	var m prefixModel
	m.ID = types.StringValue(id)
	m.Prefix = types.StringValue(p.Prefix)

	m.Description = types.StringValue(derefStr(p.Description))

	m.Created = nullableTimeStr(p.Created)

	statusName := ""
	if p.Status.Id != nil && p.Status.Id.String != nil && *p.Status.Id.String != "" {
		if n, err := getStatusName(ctx, r.client.Client, *p.Status.Id.String); err == nil {
			statusName = n
		}
	}
	m.Status = types.StringValue(statusName)

	parentID := ""
	if p.Parent.IsSet() {
		if parent := p.Parent.Get(); parent != nil && parent.Id != nil && parent.Id.String != nil {
			parentID = *parent.Id.String
		}
	}
	m.ParentID = types.StringValue(parentID)
	m.TenantID = nullableFKStr(p.Tenant)
	m.RoleID = nullableFKStr(p.Role)
	rirID := ""
	if p.Rir.IsSet() {
		if rir := p.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
			rirID = *rir.Id.String
		}
	}
	m.RirID = types.StringValue(rirID)
	namespaceID := ""
	if p.Namespace != nil && p.Namespace.Id != nil && p.Namespace.Id.String != nil {
		namespaceID = *p.Namespace.Id.String
	}
	m.NamespaceID = types.StringValue(namespaceID)
	m.VLANID = nullableFKStr(p.Vlan)

	m.Network = types.StringValue(p.Network)
	m.Broadcast = types.StringValue(p.Broadcast)
	m.PrefixLength = types.Int64Value(int64(p.PrefixLength))
	m.IPVersion = types.Int64Value(int64(p.IpVersion))

	m.DateAllocated = nullableTimeStr(p.DateAllocated)

	if len(p.Tags) > 0 {
		vals := make([]attr.Value, 0, len(p.Tags))
		for _, t := range p.Tags {
			if t.Id != nil && t.Id.String != nil {
				vals = append(vals, types.StringValue(*t.Id.String))
			}
		}
		m.TagsIDs = types.ListValueMust(types.StringType, vals)
	} else {
		m.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	m.Display = types.StringValue(p.Display)
	m.URL = types.StringValue(p.Url)
	m.NaturalSlug = types.StringValue(p.NaturalSlug)
	m.NotesURL = types.StringValue(p.NotesUrl)

	tflog.Debug(ctx, "read Prefix", map[string]any{"id": id})
	return m, true, diags
}
