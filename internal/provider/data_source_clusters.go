package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceClusters() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about clusters in Nautobot.",

		ReadContext: dataSourceClustersRead,

		Schema: map[string]*schema.Schema{
			"clusters": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The UUID of the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The name of the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"cluster_type_id": {
							Description: "The ID of the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"cluster_group_id": {
							Description: "The ID of the cluster group.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tenant_id": {
							Description: "The ID of the tenant associated with the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"location_id": {
							Description: "The ID of the location associated with the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"tags_ids": {
							Description: "The IDs of the tags associated with the cluster.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"comments": {
							Description: "Comments or notes about the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"created": {
							Description: "The creation date of the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"last_updated": {
							Description: "The last update date of the cluster.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceClustersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	s := meta.(*apiClient).Server
	t := meta.(*apiClient).Token.token

	// Auth context
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    t,
				Prefix: "Token",
			},
		},
	)

	// Fetch clusters list
	rsp, _, err := c.VirtualizationAPI.VirtualizationClustersList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get clusters list from %s: %s", s, err.Error())
	}

	results := rsp.Results
	list := make([]map[string]interface{}, 0)

	// Iterate over the results and map each cluster to the format expected by Terraform
	for _, cluster := range results {
		createdStr := ""
		if cluster.Created.IsSet() && cluster.Created.Get() != nil {
			createdStr = cluster.Created.Get().Format(time.RFC3339)
		}

		lastUpdatedStr := ""
		if cluster.LastUpdated.IsSet() && cluster.LastUpdated.Get() != nil {
			lastUpdatedStr = cluster.LastUpdated.Get().Format(time.RFC3339)
		}

		// Prepare itemMap with mandatory fields
		itemMap := map[string]interface{}{
			"id":           cluster.Id,
			"name":         cluster.Name,
			"comments":     cluster.Comments,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
		}

		// Extract cluster_type_id safely
		if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
			itemMap["cluster_type_id"] = *cluster.ClusterType.Id.String
		}

		// Handle nullable ClusterGroup
		if cluster.ClusterGroup.IsSet() {
			if clusterGroup := cluster.ClusterGroup.Get(); clusterGroup != nil && clusterGroup.Id != nil {
				itemMap["cluster_group_id"] = *clusterGroup.Id.String
			}
		}

		// Handle nullable Tenant
		if cluster.Tenant.IsSet() {
			if tenant := cluster.Tenant.Get(); tenant != nil && tenant.Id != nil {
				itemMap["tenant_id"] = *tenant.Id.String
			}
		}

		// Handle nullable Location
		if cluster.Location.IsSet() {
			if location := cluster.Location.Get(); location != nil && location.Id != nil {
				itemMap["location_id"] = *location.Id.String
			}
		}

		// Handle Tags
		var tags []string
		for _, tag := range cluster.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tags = append(tags, *tag.Id.String)
			}
		}
		itemMap["tags_ids"] = tags

		// Add the cluster to the list
		list = append(list, itemMap)
	}

	// Set the clusters list in the resource data
	if err := d.Set("clusters", list); err != nil {
		return diag.FromErr(err)
	}

	// Always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
