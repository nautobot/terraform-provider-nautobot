package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &NamespaceDataSource{}
	_ datasource.DataSourceWithConfigure = &NamespaceDataSource{}
)

type NamespaceDataSource struct {
	client *APIClient
}

type namespaceDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LocationID  types.String `tfsdk:"location_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewNamespaceDataSource() datasource.DataSource {
	return &NamespaceDataSource{}
}

func (d *NamespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *NamespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific IPAM namespace in Nautobot by name.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The exact name of the namespace to retrieve.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "Namespace UUID.",
				Computed:    true,
			},
			"description": dsschema.StringAttribute{
				Description: "Namespace description.",
				Computed:    true,
			},
			"location_id": dsschema.StringAttribute{
				Description: "UUID of the location associated with the namespace.",
				Computed:    true,
			},
			"tenant_id": dsschema.StringAttribute{
				Description: "UUID of the tenant associated with the namespace.",
				Computed:    true,
			},
			"created": dsschema.StringAttribute{
				Description: "Namespace creation date (RFC3339).",
				Computed:    true,
			},
			"last_updated": dsschema.StringAttribute{
				Description: "Namespace last update date (RFC3339).",
				Computed:    true,
			},
			"display": dsschema.StringAttribute{
				Description: "Human-friendly display value for the namespace.",
				Computed:    true,
			},
			"url": dsschema.StringAttribute{
				Description: "API URL of the namespace.",
				Computed:    true,
			},
			"natural_slug": dsschema.StringAttribute{
				Description: "Natural slug for the namespace.",
				Computed:    true,
			},
			"notes_url": dsschema.StringAttribute{
				Description: "Notes URL for the namespace.",
				Computed:    true,
			},
		},
	}
}

func (d *NamespaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *NamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data namespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}

	name := data.Name.ValueString()
	rsp, httpResp, err := d.client.Client.IpamAPI.
		IpamNamespacesList(ctx).
		Name([]string{name}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", httpErr(err, httpResp))
		return
	}
	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError("Namespace not found", "No namespace found with name "+name)
		return
	}
	if len(rsp.Results) > 1 {
		resp.Diagnostics.AddError("Multiple namespaces found", "More than one namespace matched name "+name)
		return
	}

	n := rsp.Results[0]
	if n.Id == nil || *n.Id == "" {
		resp.Diagnostics.AddError("Invalid namespace data", "Namespace "+name+" returned no id")
		return
	}

	data.ID = types.StringValue(*n.Id)
	data.Name = types.StringValue(n.Name)
	data.Description = types.StringValue(derefStr(n.Description))
	data.LocationID = nullableFKStr(n.Location)
	data.TenantID = nullableFKStr(n.Tenant)
	data.Created = nullableTimeStr(n.Created)
	data.LastUpdated = nullableTimeStr(n.LastUpdated)
	data.Display = types.StringValue(n.Display)
	data.URL = types.StringValue(n.Url)
	data.NaturalSlug = types.StringValue(n.NaturalSlug)
	data.NotesURL = types.StringValue(n.NotesUrl)

	tflog.Debug(ctx, "read namespace", map[string]any{"id": *n.Id, "name": name})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
