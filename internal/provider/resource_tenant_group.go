package provider

import (
	"context"

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
	_ resource.Resource                = &TenantGroupResource{}
	_ resource.ResourceWithImportState = &TenantGroupResource{}
)

type TenantGroupResource struct {
	client *APIClient
}

type tenantGroupModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ParentID    types.String `tfsdk:"parent_id"`
	Created     types.String `tfsdk:"created"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewTenantGroupResource() resource.Resource {
	return &TenantGroupResource{}
}

func (r *TenantGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_group"
}

func (r *TenantGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a tenant group in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant group's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Tenant group's name.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant group's description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parent_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "UUID of the parent tenant group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant group's creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant group's display name.",
			},
			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant group's URL.",
			},
			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the tenant group.",
			},
			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the tenant group.",
			},
		},
	}
}

func (r *TenantGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *TenantGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	body := nb.TenantGroupRequest{
		Name: plan.Name.ValueString(),
	}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.ParentID.ValueString(); v != "" {
		body.Parent = makeFKUser(v)
	}

	out, httpResp, err := c.TenancyAPI.
		TenancyTenantGroupsCreate(ctx).
		TenantGroupRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create tenant group", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created tenant group returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read tenant group", "created tenant group was not found")
		return
	}

	tflog.Debug(ctx, "tenant group created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantGroupModel
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

func (r *TenantGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tenantGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedTenantGroupRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.ParentID.Equal(state.ParentID) {
		patch.Parent = makeFKUser(plan.ParentID.ValueString())
	}

	_, httpResp, err := c.TenancyAPI.
		TenancyTenantGroupsPartialUpdate(ctx, id).
		PatchedTenantGroupRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update tenant group", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read tenant group", "updated tenant group was not found")
		return
	}

	tflog.Debug(ctx, "tenant group updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantGroupsDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete tenant group", httpErr(err, httpResp))
		return
	}
}

func (r *TenantGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TenantGroupResource) readModel(ctx context.Context, id string) (tenantGroupModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	m, httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantGroupsRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return tenantGroupModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read tenant group", httpErr(err, httpResp))
		return tenantGroupModel{}, false, diags
	}

	var out tenantGroupModel
	out.ID = types.StringValue(id)
	out.Name = types.StringValue(m.Name)

	out.Description = types.StringValue(derefStr(m.Description))
	out.ParentID = nullableFKStr(m.Parent)
	out.Created = nullableTimeStr(m.Created)
	out.Display = types.StringValue(m.Display)
	out.URL = types.StringValue(m.Url)
	out.NaturalSlug = types.StringValue(m.NaturalSlug)
	out.NotesURL = types.StringValue(m.NotesUrl)

	tflog.Debug(ctx, "read tenant group", map[string]any{"id": id})
	return out, true, diags
}
