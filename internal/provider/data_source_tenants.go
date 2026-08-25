package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &TenantsDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantsDataSource{}
)

type TenantsDataSource struct {
	client *APIClient
}

type tenantItemModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Comments      types.String `tfsdk:"comments"`
	TenantGroupID types.String `tfsdk:"tenant_group_id"`
	Created       types.String `tfsdk:"created"`
	LastUpdated   types.String `tfsdk:"last_updated"`
	Display       types.String `tfsdk:"display"`
	URL           types.String `tfsdk:"url"`
	NaturalSlug   types.String `tfsdk:"natural_slug"`
	NotesURL      types.String `tfsdk:"notes_url"`
}

type tenantsDataSourceModel struct {
	Tenants []tenantItemModel `tfsdk:"tenants"`
}

func NewTenantsDataSource() datasource.DataSource {
	return &TenantsDataSource{}
}

func (d *TenantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenants"
}

func (d *TenantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all tenants in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"tenants": dsschema.ListNestedAttribute{
				Description: "List of tenants.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "Tenant's UUID.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "Tenant's name.",
							Computed:    true,
						},
						"description": dsschema.StringAttribute{
							Description: "Tenant's description.",
							Computed:    true,
						},
						"comments": dsschema.StringAttribute{
							Description: "Tenant's comments.",
							Computed:    true,
						},
						"tenant_group_id": dsschema.StringAttribute{
							Description: "UUID of the tenant group this tenant belongs to.",
							Computed:    true,
						},
						"created": dsschema.StringAttribute{
							Description: "Tenant's creation date (RFC3339).",
							Computed:    true,
						},
						"last_updated": dsschema.StringAttribute{
							Description: "Tenant's last update date (RFC3339).",
							Computed:    true,
						},
						"display": dsschema.StringAttribute{
							Description: "Human friendly display value for the tenant.",
							Computed:    true,
						},
						"url": dsschema.StringAttribute{
							Description: "URL of the tenant.",
							Computed:    true,
						},
						"natural_slug": dsschema.StringAttribute{
							Description: "Natural slug for the tenant.",
							Computed:    true,
						},
						"notes_url": dsschema.StringAttribute{
							Description: "Notes URL for the tenant.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TenantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *TenantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tenantsDataSourceModel

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

	state.Tenants = make([]tenantItemModel, 0)

	for {
		rsp, httpResp, err := c.TenancyAPI.
			TenancyTenantsList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get tenants list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, m := range results {
			var item tenantItemModel

			if m.Id == nil || *m.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid tenant data",
					"Tenants list returned an item with no id (name: "+m.Name+")",
				)
				return
			}
			item.ID = types.StringValue(*m.Id)

			item.Name = types.StringValue(m.Name)
			item.Display = types.StringValue(m.Display)
			item.URL = types.StringValue(m.Url)
			item.NaturalSlug = types.StringValue(m.NaturalSlug)
			item.NotesURL = types.StringValue(m.NotesUrl)
			item.Description = types.StringValue(derefStr(m.Description))
			item.Comments = types.StringValue(derefStr(m.Comments))
			item.TenantGroupID = nullableFKStr(m.TenantGroup)
			item.Created = nullableTimeStr(m.Created)
			item.LastUpdated = nullableTimeStr(m.LastUpdated)

			state.Tenants = append(state.Tenants, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read tenants", map[string]any{"count": len(state.Tenants)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
