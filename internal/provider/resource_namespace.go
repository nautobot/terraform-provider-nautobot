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
	_ resource.Resource                = &NamespaceResource{}
	_ resource.ResourceWithImportState = &NamespaceResource{}
)

type NamespaceResource struct {
	client *APIClient
}

type namespaceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LocationID  types.String `tfsdk:"location_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Created     types.String `tfsdk:"created"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

func (r *NamespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages an IPAM namespace in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Namespace UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Namespace name.",
			},
			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Namespace description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"location_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "UUID of the location associated with the namespace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tenant_id": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "UUID of the tenant associated with the namespace.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Namespace creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Human-friendly display value for the namespace.",
			},
			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "API URL of the namespace.",
			},
			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the namespace.",
			},
			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the namespace.",
			},
		},
	}
}

func (r *NamespaceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := nb.NamespaceRequest{Name: plan.Name.ValueString()}
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}
	if v := plan.LocationID.ValueString(); v != "" {
		body.Location = makeFKUser(v)
	}
	if v := plan.TenantID.ValueString(); v != "" {
		body.Tenant = makeFKUser(v)
	}

	out, httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesCreate(ctx).
		NamespaceRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create namespace", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created namespace returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read namespace", "created namespace was not found")
		return
	}

	tflog.Debug(ctx, "namespace created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namespaceModel
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

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	var patch nb.PatchedNamespaceRequest
	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
	}
	if !plan.Description.Equal(state.Description) {
		v := plan.Description.ValueString()
		patch.Description = &v
	}
	if !plan.LocationID.Equal(state.LocationID) {
		patch.Location = makeFKUser(plan.LocationID.ValueString())
	}
	if !plan.TenantID.Equal(state.TenantID) {
		patch.Tenant = makeFKUser(plan.TenantID.ValueString())
	}

	_, httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesPartialUpdate(ctx, id).
		PatchedNamespaceRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update namespace", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read namespace", "updated namespace was not found")
		return
	}

	tflog.Debug(ctx, "namespace updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete namespace", httpErr(err, httpResp))
	}
}

func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NamespaceResource) readModel(ctx context.Context, id string) (namespaceModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	n, httpResp, err := r.client.Client.IpamAPI.
		IpamNamespacesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return namespaceModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read namespace", httpErr(err, httpResp))
		return namespaceModel{}, false, diags
	}

	out := namespaceModel{
		ID:          types.StringValue(id),
		Name:        types.StringValue(n.Name),
		Description: types.StringValue(derefStr(n.Description)),
		LocationID:  nullableFKStr(n.Location),
		TenantID:    nullableFKStr(n.Tenant),
		Created:     nullableTimeStr(n.Created),
		Display:     types.StringValue(n.Display),
		URL:         types.StringValue(n.Url),
		NaturalSlug: types.StringValue(n.NaturalSlug),
		NotesURL:    types.StringValue(n.NotesUrl),
	}

	tflog.Debug(ctx, "read namespace", map[string]any{"id": id})
	return out, true, diags
}
