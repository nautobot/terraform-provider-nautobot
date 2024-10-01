package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceClusterTypes() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about cluster types in Nautobot.",

		ReadContext: dataSourceClusterTypesRead,

		Schema: map[string]*schema.Schema{
			"cluster_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The UUID of the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"object_type": {
							Description: "Object type of the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"display": {
							Description: "Human-friendly display value for the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"url": {
							Description: "URL of the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"natural_slug": {
							Description: "Natural slug for the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The name of the cluster type.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"description": {
							Description: "The description of the cluster type.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"created": {
							Description: "The date the cluster type was created.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"last_updated": {
							Description: "The date the cluster type was last updated.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"notes_url": {
							Description: "Notes URL for the cluster type.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceClusterTypesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	rsp, _, err := c.VirtualizationAPI.VirtualizationClusterTypesList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get cluster types list from %s: %s", s, err.Error())
	}

	results := rsp.Results
	list := make([]map[string]interface{}, 0)

	// Iterate over the results and map each cluster type to the format expected by Terraform
	for _, clusterType := range results {
		createdStr := ""
		if clusterType.Created.IsSet() && clusterType.Created.Get() != nil {
			createdStr = clusterType.Created.Get().Format(time.RFC3339)
		}

		lastUpdatedStr := ""
		if clusterType.LastUpdated.IsSet() && clusterType.LastUpdated.Get() != nil {
			lastUpdatedStr = clusterType.LastUpdated.Get().Format(time.RFC3339)
		}

		itemMap := map[string]interface{}{
			"id":           clusterType.Id,
			"object_type":  clusterType.ObjectType,
			"display":      clusterType.Display,
			"url":          clusterType.Url,
			"natural_slug": clusterType.NaturalSlug,
			"name":         clusterType.Name,
			"description":  clusterType.Description,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
			"notes_url":    clusterType.NotesUrl,
		}
		list = append(list, itemMap)
	}

	if err := d.Set("cluster_types", list); err != nil {
		return diag.FromErr(err)
	}

	// Always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
