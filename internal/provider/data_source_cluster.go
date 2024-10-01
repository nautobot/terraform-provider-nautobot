package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a specific cluster in Nautobot.",

		ReadContext: dataSourceClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the cluster.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The UUID of the cluster.",
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
	}
}

func dataSourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
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

	// Fetch the cluster name from Terraform configuration
	clusterName := d.Get("name").(string)

	// Fetch clusters by name
	rsp, _, err := c.VirtualizationAPI.VirtualizationClustersList(auth).Name([]string{clusterName}).Execute()
	if err != nil {
		return diag.Errorf("failed to get cluster with name %s: %s", clusterName, err.Error())
	}

	// Ensure at least one result is returned
	if len(rsp.Results) == 0 {
		return diag.Errorf("no cluster found with name %s", clusterName)
	}

	cluster := rsp.Results[0]

	d.SetId(cluster.Id)

	// Set basic fields
	d.Set("id", cluster.Id)
	d.Set("name", cluster.Name)
	d.Set("comments", cluster.Comments)

	// Convert created and last updated fields to strings
	createdStr := ""
	if cluster.Created.IsSet() && cluster.Created.Get() != nil {
		createdStr = cluster.Created.Get().Format(time.RFC3339)
	}
	d.Set("created", createdStr)

	lastUpdatedStr := ""
	if cluster.LastUpdated.IsSet() && cluster.LastUpdated.Get() != nil {
		lastUpdatedStr = cluster.LastUpdated.Get().Format(time.RFC3339)
	}
	d.Set("last_updated", lastUpdatedStr)

	// Handle cluster_type_id
	if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
		d.Set("cluster_type_id", *cluster.ClusterType.Id.String)
	}

	// Handle cluster_group_id
	if cluster.ClusterGroup.IsSet() {
		if clusterGroup := cluster.ClusterGroup.Get(); clusterGroup != nil && clusterGroup.Id != nil {
			d.Set("cluster_group_id", *clusterGroup.Id.String)
		}
	}

	// Handle tenant_id
	if cluster.Tenant.IsSet() {
		if tenant := cluster.Tenant.Get(); tenant != nil && tenant.Id != nil {
			d.Set("tenant_id", *tenant.Id.String)
		}
	}

	// Handle location_id
	if cluster.Location.IsSet() {
		if location := cluster.Location.Get(); location != nil && location.Id != nil {
			d.Set("location_id", *location.Id.String)
		}
	}

	// Handle tags
	var tags []string
	for _, tag := range cluster.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tags = append(tags, *tag.Id.String)
		}
	}
	d.Set("tags_ids", tags)

	return diags
}
