package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ManufacturerDataSource{}
	_ datasource.DataSourceWithConfigure = &ManufacturerDataSource{}
)

type ManufacturerDataSource struct {
	client *APIClient
}

type manufacturerDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewManufacturerDataSource() datasource.DataSource {
	return &ManufacturerDataSource{}
}

func (d *ManufacturerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturer"
}

func (d *ManufacturerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific manufacturer in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The name of the manufacturer to retrieve.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "Manufacturer's UUID.",
				Computed:    true,
			},
			"display": dsschema.StringAttribute{
				Description: "Human friendly display value for the manufacturer.",
				Computed:    true,
			},
			"url": dsschema.StringAttribute{
				Description: "URL of the manufacturer.",
				Computed:    true,
			},
			"natural_slug": dsschema.StringAttribute{
				Description: "Natural slug for the manufacturer.",
				Computed:    true,
			},
			"description": dsschema.StringAttribute{
				Description: "Manufacturer's description.",
				Computed:    true,
			},
			"created": dsschema.StringAttribute{
				Description: "Manufacturer's creation date (RFC3339).",
				Computed:    true,
			},
			"last_updated": dsschema.StringAttribute{
				Description: "Manufacturer's last update date (RFC3339).",
				Computed:    true,
			},
			"notes_url": dsschema.StringAttribute{
				Description: "Notes URL for the manufacturer.",
				Computed:    true,
			},
		},
	}
}

func (d *ManufacturerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *ManufacturerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data manufacturerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client
	name := data.Name.ValueString()

	rsp, httpResp, err := c.DcimAPI.
		DcimManufacturersList(ctx).
		Name([]string{name}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get manufacturer",
			httpErr(err, httpResp),
		)
		return
	}

	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Manufacturer not found",
			"No manufacturer found with name "+name,
		)
		return
	}

	m := rsp.Results[0]

	if m.Id == nil || *m.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid manufacturer data",
			"Manufacturer "+name+" returned no id",
		)
		return
	}
	id := *m.Id
	data.ID = types.StringValue(id)

	data.Name = types.StringValue(m.Name)
	data.Display = types.StringValue(m.Display)
	data.URL = types.StringValue(m.Url)
	data.NaturalSlug = types.StringValue(m.NaturalSlug)
	data.Description = types.StringValue(derefStr(m.Description))
	data.Created = nullableTimeStr(m.Created)
	data.LastUpdated = nullableTimeStr(m.LastUpdated)
	data.NotesURL = types.StringValue(m.NotesUrl)

	tflog.Debug(ctx, "read manufacturer", map[string]any{"id": id, "name": name})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
