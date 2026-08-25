package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &ClusterDataSource{}
	_ datasource.DataSourceWithConfigure = &ClusterDataSource{}
)

type ClusterDataSource struct {
	client *APIClient
}

type clusterDataSourceModel struct {
	Name           types.String `tfsdk:"name"`
	ID             types.String `tfsdk:"id"`
	ClusterTypeID  types.String `tfsdk:"cluster_type_id"`
	ClusterGroupID types.String `tfsdk:"cluster_group_id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	LocationID     types.String `tfsdk:"location_id"`
	TagsIDs        types.List   `tfsdk:"tags_ids"`
	Comments       types.String `tfsdk:"comments"`
	Created        types.String `tfsdk:"created"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

func (d *ClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *ClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific cluster in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The name of the cluster.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "The UUID of the cluster.",
				Computed:    true,
			},
			"cluster_type_id": dsschema.StringAttribute{
				Description: "The ID of the cluster type.",
				Computed:    true,
			},
			"cluster_group_id": dsschema.StringAttribute{
				Description: "The ID of the cluster group.",
				Computed:    true,
			},
			"tenant_id": dsschema.StringAttribute{
				Description: "The ID of the tenant associated with the cluster.",
				Computed:    true,
			},
			"location_id": dsschema.StringAttribute{
				Description: "The ID of the location associated with the cluster.",
				Computed:    true,
			},
			"tags_ids": dsschema.ListAttribute{
				Description: "The IDs of the tags associated with the cluster.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"comments": dsschema.StringAttribute{
				Description: "Comments or notes about the cluster.",
				Computed:    true,
			},
			"created": dsschema.StringAttribute{
				Description: "The creation date of the cluster (RFC3339).",
				Computed:    true,
			},
			"last_updated": dsschema.StringAttribute{
				Description: "The last update date of the cluster (RFC3339).",
				Computed:    true,
			},
		},
	}
}

func (d *ClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data clusterDataSourceModel

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

	clusterName := data.Name.ValueString()

	rsp, httpResp, err := c.VirtualizationAPI.
		VirtualizationClustersList(ctx).
		Name([]string{clusterName}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster",
			httpErr(err, httpResp),
		)
		return
	}

	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError(
			"Cluster not found",
			"No cluster found with name "+clusterName,
		)
		return
	}

	cluster := rsp.Results[0]

	if cluster.Id == nil || *cluster.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid cluster data",
			"Cluster "+clusterName+" returned no id",
		)
		return
	}
	id := *cluster.Id
	data.ID = types.StringValue(id)

	data.Name = types.StringValue(cluster.Name)

	data.Comments = types.StringValue(derefStr(cluster.Comments))
	data.Created = nullableTimeStr(cluster.Created)
	data.LastUpdated = nullableTimeStr(cluster.LastUpdated)

	if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
		data.ClusterTypeID = types.StringValue(*cluster.ClusterType.Id.String)
	} else {
		data.ClusterTypeID = types.StringValue("")
	}

	data.ClusterGroupID = nullableFKStr(cluster.ClusterGroup)
	data.TenantID = nullableFKStr(cluster.Tenant)
	data.LocationID = nullableFKStr(cluster.Location)

	if len(cluster.Tags) > 0 {
		tagVals := make([]attr.Value, 0, len(cluster.Tags))
		for _, tag := range cluster.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tagVals = append(tagVals, types.StringValue(*tag.Id.String))
			}
		}
		data.TagsIDs = types.ListValueMust(types.StringType, tagVals)
	} else {
		data.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	tflog.Debug(ctx, "read cluster", map[string]any{"id": id, "name": clusterName})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
