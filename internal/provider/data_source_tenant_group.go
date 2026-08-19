package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &TenantGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &TenantGroupDataSource{}
)

type TenantGroupDataSource struct {
	client *APIClient
}

type tenantGroupDataSourceModel struct {
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

func NewTenantGroupDataSource() datasource.DataSource {
	return &TenantGroupDataSource{}
}

func (d *TenantGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_group"
}

func (d *TenantGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific tenant group in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The name of the tenant group to retrieve.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "Tenant group's UUID.",
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
	}
}

func (d *TenantGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *TenantGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tenantGroupDataSourceModel

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

	rsp, httpResp, err := c.TenancyAPI.
		TenancyTenantGroupsList(ctx).
		Name([]string{name}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get tenant group",
			httpErr(err, httpResp),
		)
		return
	}

	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Tenant group not found",
			"No tenant group found with name "+name,
		)
		return
	}

	m := rsp.Results[0]

	if m.Id == nil || *m.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid tenant group data",
			"Tenant group "+name+" returned no id",
		)
		return
	}

	data.ID = types.StringValue(*m.Id)
	data.Name = types.StringValue(m.Name)
	data.Display = types.StringValue(m.Display)
	data.URL = types.StringValue(m.Url)
	data.NaturalSlug = types.StringValue(m.NaturalSlug)
	data.NotesURL = types.StringValue(m.NotesUrl)
	data.Description = types.StringValue(derefStr(m.Description))
	data.ParentID = nullableFKStr(m.Parent)
	data.Created = nullableTimeStr(m.Created)
	data.LastUpdated = nullableTimeStr(m.LastUpdated)

	tflog.Debug(ctx, "read tenant group", map[string]any{"id": *m.Id, "name": name})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
