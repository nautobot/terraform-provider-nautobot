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
	_ resource.Resource                = &ManufacturerResource{}
	_ resource.ResourceWithImportState = &ManufacturerResource{}
)

type ManufacturerResource struct {
	client *APIClient
}

type manufacturerModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewManufacturerResource() resource.Resource {
	return &ManufacturerResource{}
}

func (r *ManufacturerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturer"
}

func (r *ManufacturerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "This object manages a manufacturer in Nautobot.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Manufacturer's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Manufacturer's name.",
			},

			"description": rschema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Manufacturer's description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"created": rschema.StringAttribute{
				Computed:    true,
				Description: "Manufacturer's creation date (RFC3339).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"display": rschema.StringAttribute{
				Computed:    true,
				Description: "Manufacturer's display name.",
			},
			"url": rschema.StringAttribute{
				Computed:    true,
				Description: "Manufacturer's URL.",
			},
			"natural_slug": rschema.StringAttribute{
				Computed:    true,
				Description: "Natural slug for the manufacturer.",
			},
			"notes_url": rschema.StringAttribute{
				Computed:    true,
				Description: "Notes URL for the manufacturer.",
			},
		},
	}
}

func (r *ManufacturerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *ManufacturerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan manufacturerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.client.Client

	var body nb.ManufacturerRequest
	body.Name = plan.Name.ValueString()
	if v := plan.Description.ValueString(); v != "" {
		body.Description = &v
	}

	out, httpResp, err := c.DcimAPI.
		DcimManufacturersCreate(ctx).
		ManufacturerRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to create manufacturer", httpErr(err, httpResp))
		return
	}
	if out.Id == nil || *out.Id == "" {
		resp.Diagnostics.AddError("invalid API response", "created manufacturer returned no id")
		return
	}

	model, found, diags := r.readModel(ctx, *out.Id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read manufacturer", "created manufacturer was not found")
		return
	}

	tflog.Debug(ctx, "manufacturer created", map[string]any{"id": *out.Id, "name": plan.Name.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ManufacturerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state manufacturerModel
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

func (r *ManufacturerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state manufacturerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	c := r.client.Client

	var patch nb.PatchedManufacturerRequest

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

	_, httpResp, err := c.DcimAPI.
		DcimManufacturersPartialUpdate(ctx, id).
		PatchedManufacturerRequest(patch).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("failed to update manufacturer", httpErr(err, httpResp))
		return
	}

	model, found, diags := r.readModel(ctx, id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("failed to read manufacturer", "updated manufacturer was not found")
		return
	}

	tflog.Debug(ctx, "manufacturer updated", map[string]any{"id": id})
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ManufacturerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state manufacturerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.Client.DcimAPI.
		DcimManufacturersDestroy(ctx, state.ID.ValueString()).
		Execute()
	if err != nil && !isNotFoundResponse(httpResp) {
		resp.Diagnostics.AddError("failed to delete manufacturer", httpErr(err, httpResp))
		return
	}
}

func (r *ManufacturerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ManufacturerResource) readModel(ctx context.Context, id string) (manufacturerModel, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	m, httpResp, err := r.client.Client.DcimAPI.
		DcimManufacturersRetrieve(ctx, id).
		Execute()
	if isNotFoundResponse(httpResp) {
		return manufacturerModel{}, false, diags
	}
	if err != nil {
		diags.AddError("failed to read manufacturer", httpErr(err, httpResp))
		return manufacturerModel{}, false, diags
	}

	var out manufacturerModel
	out.ID = types.StringValue(id)
	out.Name = types.StringValue(m.Name)

	out.Description = types.StringValue(derefStr(m.Description))

	out.Created = nullableTimeStr(m.Created)

	out.Display = types.StringValue(m.Display)
	out.URL = types.StringValue(m.Url)
	out.NaturalSlug = types.StringValue(m.NaturalSlug)
	out.NotesURL = types.StringValue(m.NotesUrl)

	tflog.Debug(ctx, "read manufacturer", map[string]any{"id": id})
	return out, true, diags
}
