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
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ resource.Resource                = &ClusterTypeResource{}
	_ resource.ResourceWithImportState = &ClusterTypeResource{}
)

type ClusterTypeResource struct {
	client *APIClient
}

type clusterTypeModel struct {
	ID          types.String `tfsdk:"id"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewClusterTypeResource() resource.Resource {
	return &ClusterTypeResource{}
}

func (r *ClusterTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_type"
}

func (r *ClusterTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a cluster type in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Cluster type's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Human-friendly display value for the cluster type.",
			},
			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "URL of the cluster type.",
			},
			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the cluster type.",
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Cluster type's name.",
			},

			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description for the cluster type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Creation date of the cluster type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the cluster type.",
			},
		},
	}
}

func (r *ClusterTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *ClusterTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	var body nb.ClusterTypeRequest
	body.Name = plan.Name.ValueString()
	if plan.Description.ValueString() != "" {
		desc := plan.Description.ValueString()
		body.Description = &desc
	}

	out, httpResp, err := c.VirtualizationAPI.
		VirtualizationClusterTypesCreate(ctx).
		ClusterTypeRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create cluster type", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created cluster type returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read cluster type", "created cluster type was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterTypeModel
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

func (r *ClusterTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clusterTypeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedClusterTypeRequest

	if !plan.Name.Equal(state.Name) {
		v := plan.Name.ValueString()
		patch.Name = &v
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

	_, httpResp, err := c.VirtualizationAPI.
		VirtualizationClusterTypesPartialUpdate(ctx, id).
		PatchedClusterTypeRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update cluster type", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read cluster type", "updated cluster type was not found")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ClusterTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterTypeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationClusterTypesDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete cluster type", httpErr(err, httpResp))
		return
	}
}

func (r *ClusterTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ClusterTypeResource) readModel(ctx context.Context, id string) (clusterTypeModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	ct, httpResp, err := r.client.Client.VirtualizationAPI.
		VirtualizationClusterTypesRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return clusterTypeModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read cluster type", httpErr(err, httpResp))
		return clusterTypeModel{}, false, diags
	}

	var m clusterTypeModel
	m.ID = types.StringValue(id)
	m.Name = types.StringValue(ct.Name)
	m.Display = types.StringValue(ct.Display)
	m.URL = types.StringValue(ct.Url)
	m.NaturalSlug = types.StringValue(ct.NaturalSlug)

	m.Description = types.StringValue(derefStr(ct.Description))

	m.Created = nullableTimeStr(ct.Created)

	m.NotesURL = types.StringValue(ct.NotesUrl)

	return m, true, diags
}
