package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ManufacturersDataSource{}
	_ datasource.DataSourceWithConfigure = &ManufacturersDataSource{}
)

type ManufacturersDataSource struct {
	client *APIClient
}

type manufacturerItemModel struct {
	ID          types.String `tfsdk:"id"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

type manufacturersDataSourceModel struct {
	Manufacturers []manufacturerItemModel `tfsdk:"manufacturers"`
}

func NewManufacturersDataSource() datasource.DataSource {
	return &ManufacturersDataSource{}
}

func (d *ManufacturersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturers"
}

func (d *ManufacturersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Manufacturer data source in the Terraform Nautobot provider. Retrieves information about all manufacturers.",
		Attributes: map[string]dsschema.Attribute{
			"manufacturers": dsschema.ListNestedAttribute{
				Description: "List of manufacturers.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "Manufacturer's UUID.",
							Computed:    true,
						},
						"display": dsschema.StringAttribute{
							Description: "Human friendly display value for the Manufacturer.",
							Computed:    true,
						},
						"url": dsschema.StringAttribute{
							Description: "URL of the Manufacturer.",
							Computed:    true,
						},
						"natural_slug": dsschema.StringAttribute{
							Description: "Natural slug for the Manufacturer.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "Manufacturer's name.",
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
							Description: "Notes URL for the Manufacturer.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ManufacturersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *ManufacturersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state manufacturersDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	const pageLimit int32 = 200
	var offset int32 = 0

	state.Manufacturers = make([]manufacturerItemModel, 0)

	for {
		rsp, httpResp, err := c.DcimAPI.
			DcimManufacturersList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get manufacturers list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, m := range results {
			var item manufacturerItemModel

			if m.Id == nil || *m.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid manufacturer data",
					"Manufacturers list returned an item with no id (name: "+m.Name+")",
				)
				return
			}
			item.ID = types.StringValue(*m.Id)

			item.Display = types.StringValue(m.Display)
			item.URL = types.StringValue(m.Url)
			item.NaturalSlug = types.StringValue(m.NaturalSlug)
			item.Name = types.StringValue(m.Name)
			item.Description = types.StringValue(derefStr(m.Description))
			item.Created = nullableTimeStr(m.Created)
			item.LastUpdated = nullableTimeStr(m.LastUpdated)
			item.NotesURL = types.StringValue(m.NotesUrl)

			state.Manufacturers = append(state.Manufacturers, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read manufacturers", map[string]any{"count": len(state.Manufacturers)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
