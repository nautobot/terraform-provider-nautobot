package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ClusterTypeDataSource{}
	_ datasource.DataSourceWithConfigure = &ClusterTypeDataSource{}
)

type ClusterTypeDataSource struct {
	client *APIClient
}

type clusterTypeDataSourceModel struct {
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

func NewClusterTypeDataSource() datasource.DataSource {
	return &ClusterTypeDataSource{}
}

func (d *ClusterTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_type"
}

func (d *ClusterTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific cluster type in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The name of the cluster type to retrieve.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "The UUID of the cluster type.",
				Computed:    true,
			},
			"display": dsschema.StringAttribute{
				Description: "Human-friendly display value for the cluster type.",
				Computed:    true,
			},
			"url": dsschema.StringAttribute{
				Description: "URL of the cluster type.",
				Computed:    true,
			},
			"natural_slug": dsschema.StringAttribute{
				Description: "Natural slug for the cluster type.",
				Computed:    true,
			},
			"description": dsschema.StringAttribute{
				Description: "The description of the cluster type.",
				Computed:    true,
			},
			"created": dsschema.StringAttribute{
				Description: "The date the cluster type was created (RFC3339).",
				Computed:    true,
			},
			"last_updated": dsschema.StringAttribute{
				Description: "The date the cluster type was last updated (RFC3339).",
				Computed:    true,
			},
			"notes_url": dsschema.StringAttribute{
				Description: "Notes URL for the cluster type.",
				Computed:    true,
			},
		},
	}
}

func (d *ClusterTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *ClusterTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clusterTypeDataSourceModel

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

	// Fetch cluster type by name
	rsp, httpResp, err := c.VirtualizationAPI.
		VirtualizationClusterTypesList(ctx).
		Name([]string{name}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster type",
			httpErr(err, httpResp),
		)
		return
	}
	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Cluster type not found",
			"No cluster type found with name "+name,
		)
		return
	}

	ct := rsp.Results[0]

	if ct.Id == nil || *ct.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid cluster type data",
			"Cluster type "+name+" returned no id",
		)
		return
	}
	id := *ct.Id
	data.ID = types.StringValue(id)

	data.Name = types.StringValue(ct.Name)
	data.Display = types.StringValue(ct.Display)
	data.URL = types.StringValue(ct.Url)
	data.NaturalSlug = types.StringValue(ct.NaturalSlug)
	data.Description = types.StringValue(derefStr(ct.Description))
	data.Created = nullableTimeStr(ct.Created)
	data.LastUpdated = nullableTimeStr(ct.LastUpdated)
	data.NotesURL = types.StringValue(ct.NotesUrl)

	tflog.Debug(ctx, "read cluster type", map[string]any{"id": id, "name": name})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
