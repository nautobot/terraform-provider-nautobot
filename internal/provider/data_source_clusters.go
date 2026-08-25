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
	_ datasource.DataSource              = &ClustersDataSource{}
	_ datasource.DataSourceWithConfigure = &ClustersDataSource{}
)

type ClustersDataSource struct {
	client *APIClient
}

type clusterItemModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	ClusterTypeID  types.String `tfsdk:"cluster_type_id"`
	ClusterGroupID types.String `tfsdk:"cluster_group_id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	LocationID     types.String `tfsdk:"location_id"`
	TagsIDs        types.List   `tfsdk:"tags_ids"`
	Comments       types.String `tfsdk:"comments"`
	Created        types.String `tfsdk:"created"`
	LastUpdated    types.String `tfsdk:"last_updated"`
}

type clustersDataSourceModel struct {
	Clusters []clusterItemModel `tfsdk:"clusters"`
}

func NewClustersDataSource() datasource.DataSource {
	return &ClustersDataSource{}
}

func (d *ClustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *ClustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about clusters in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"clusters": dsschema.ListNestedAttribute{
				Description: "List of clusters.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "The UUID of the cluster.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "The name of the cluster.",
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
				},
			},
		},
	}
}

func (d *ClustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *ClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state clustersDataSourceModel

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

	state.Clusters = make([]clusterItemModel, 0)

	for {
		rsp, httpResp, err := c.VirtualizationAPI.
			VirtualizationClustersList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get clusters list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, cluster := range results {
			var item clusterItemModel

			if cluster.Id == nil || *cluster.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid cluster data",
					"Clusters list returned an item with no id (name: "+cluster.Name+")",
				)
				return
			}
			item.ID = types.StringValue(*cluster.Id)

			item.Name = types.StringValue(cluster.Name)

			item.Comments = types.StringValue(derefStr(cluster.Comments))
			item.Created = nullableTimeStr(cluster.Created)
			item.LastUpdated = nullableTimeStr(cluster.LastUpdated)

			if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
				item.ClusterTypeID = types.StringValue(*cluster.ClusterType.Id.String)
			} else {
				item.ClusterTypeID = types.StringValue("")
			}

			item.ClusterGroupID = nullableFKStr(cluster.ClusterGroup)
			item.TenantID = nullableFKStr(cluster.Tenant)
			item.LocationID = nullableFKStr(cluster.Location)

			if len(cluster.Tags) > 0 {
				tagVals := make([]attr.Value, 0, len(cluster.Tags))
				for _, tag := range cluster.Tags {
					if tag.Id != nil && tag.Id.String != nil {
						tagVals = append(tagVals, types.StringValue(*tag.Id.String))
					}
				}
				item.TagsIDs = types.ListValueMust(types.StringType, tagVals)
			} else {
				item.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}

			state.Clusters = append(state.Clusters, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read clusters", map[string]any{"count": len(state.Clusters)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
