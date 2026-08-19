package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &TenantGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantGroupsDataSource{}
)

type TenantGroupsDataSource struct {
	client *APIClient
}

type tenantGroupItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	ParentID    types.String `tfsdk:"parent_id"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

type tenantGroupsDataSourceModel struct {
	TenantGroups []tenantGroupItemModel `tfsdk:"tenant_groups"`
}

func NewTenantGroupsDataSource() datasource.DataSource {
	return &TenantGroupsDataSource{}
}

func (d *TenantGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_groups"
}

func (d *TenantGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all tenant groups in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"tenant_groups": dsschema.ListNestedAttribute{
				Description: "List of tenant groups.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "Tenant group's UUID.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "Tenant group's name.",
							Computed:    true,
						},
						"description": dsschema.StringAttribute{
							Description: "Tenant group's description.",
							Computed:    true,
						},
						"parent_id": dsschema.StringAttribute{
							Description: "UUID of the parent tenant group.",
							Computed:    true,
						},
						"created": dsschema.StringAttribute{
							Description: "Tenant group's creation date (RFC3339).",
							Computed:    true,
						},
						"last_updated": dsschema.StringAttribute{
							Description: "Tenant group's last update date (RFC3339).",
							Computed:    true,
						},
						"display": dsschema.StringAttribute{
							Description: "Human friendly display value for the tenant group.",
							Computed:    true,
						},
						"url": dsschema.StringAttribute{
							Description: "URL of the tenant group.",
							Computed:    true,
						},
						"natural_slug": dsschema.StringAttribute{
							Description: "Natural slug for the tenant group.",
							Computed:    true,
						},
						"notes_url": dsschema.StringAttribute{
							Description: "Notes URL for the tenant group.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TenantGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *TenantGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tenantGroupsDataSourceModel

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

	state.TenantGroups = make([]tenantGroupItemModel, 0)

	for {
		rsp, httpResp, err := c.TenancyAPI.
			TenancyTenantGroupsList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get tenant groups list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, m := range results {
			var item tenantGroupItemModel

			if m.Id == nil || *m.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid tenant group data",
					"Tenant groups list returned an item with no id (name: "+m.Name+")",
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
			item.ParentID = nullableFKStr(m.Parent)
			item.Created = nullableTimeStr(m.Created)
			item.LastUpdated = nullableTimeStr(m.LastUpdated)

			state.TenantGroups = append(state.TenantGroups, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read tenant groups", map[string]any{"count": len(state.TenantGroups)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
