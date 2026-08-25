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
	_ resource.Resource                = &TenantResource{}
	_ resource.ResourceWithImportState = &TenantResource{}
)

type TenantResource struct {
	client *APIClient
}

type tenantModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Comments      types.String `tfsdk:"comments"`
	TenantGroupID types.String `tfsdk:"tenant_group_id"`
	Created       types.String `tfsdk:"created"`
	Display       types.String `tfsdk:"display"`
	URL           types.String `tfsdk:"url"`
	NaturalSlug   types.String `tfsdk:"natural_slug"`
	NotesURL      types.String `tfsdk:"notes_url"`
}

func NewTenantResource() resource.Resource {
	return &TenantResource{}
}

func (r *TenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *TenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a tenant in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Tenant's name.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant's description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comments": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Tenant's comments.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_group_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "UUID of the tenant group this tenant belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant's creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant's display name.",
			},
			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "Tenant's URL.",
			},
			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the tenant.",
			},
			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the tenant.",
			},
		},
	}
}

func (r *TenantResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *TenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	body := nb.TenantRequest{
		Name: plan.Name.ValueString(),
	}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.Comments.ValueString(); v != "" {
		body.Comments = &v
	}
	if v := plan.TenantGroupID.ValueString(); v != "" {
		body.TenantGroup = makeFKUser(v)
	}

	out, httpResp, err := c.TenancyAPI.
		TenancyTenantsCreate(ctx).
		TenantRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create tenant", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created tenant returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read tenant", "created tenant was not found")
		return
	}

	tflog.Debug(ctx, "tenant created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel
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

func (r *TenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedTenantRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.Comments.Equal(state.Comments) {
		v := plan.Comments.ValueString()
		patch.Comments = &v
	}
	if !plan.TenantGroupID.Equal(state.TenantGroupID) {
		patch.TenantGroup = makeFKUser(plan.TenantGroupID.ValueString())
	}

	_, httpResp, err := c.TenancyAPI.
		TenancyTenantsPartialUpdate(ctx, id).
		PatchedTenantRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update tenant", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read tenant", "updated tenant was not found")
		return
	}

	tflog.Debug(ctx, "tenant updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantsDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete tenant", httpErr(err, httpResp))
		return
	}
}

func (r *TenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *TenantResource) readModel(ctx context.Context, id string) (tenantModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	m, httpResp, err := r.client.Client.TenancyAPI.
		TenancyTenantsRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return tenantModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read tenant", httpErr(err, httpResp))
		return tenantModel{}, false, diags
	}

	var out tenantModel
	out.ID = types.StringValue(id)
	out.Name = types.StringValue(m.Name)

	out.Description = types.StringValue(derefStr(m.Description))
	out.Comments = types.StringValue(derefStr(m.Comments))
	out.TenantGroupID = nullableFKStr(m.TenantGroup)
	out.Created = nullableTimeStr(m.Created)
	out.Display = types.StringValue(m.Display)
	out.URL = types.StringValue(m.Url)
	out.NaturalSlug = types.StringValue(m.NaturalSlug)
	out.NotesURL = types.StringValue(m.NotesUrl)

	tflog.Debug(ctx, "read tenant", map[string]any{"id": id})
	return out, true, diags
}
