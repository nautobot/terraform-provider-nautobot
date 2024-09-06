package provider

import (
	"context"
	"time"

	nb "github.com/TobiPeterG/go-nautobot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceClusterType() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a specific cluster type in Nautobot.",

		ReadContext: dataSourceClusterTypeRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the cluster type to retrieve.",
				Type:        schema.TypeString,
				Required:    true,
			},
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
			"cluster_count": {
				Description: "Count of clusters associated with the cluster type.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"description": {
				Description: "The description of the cluster type.",
				Type:        schema.TypeString,
				Computed:    true,
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
			"custom_fields": {
				Description: "Custom fields for the cluster type.",
				Type:        schema.TypeMap,
				Computed:    true,
			},
		},
	}
}

func dataSourceClusterTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Get the cluster type name from the Terraform configuration
	clusterTypeName := d.Get("name").(string)

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

	// Fetch cluster types by name
	rsp, _, err := c.VirtualizationAPI.VirtualizationClusterTypesList(auth).Name([]string{clusterTypeName}).Execute()
	if err != nil {
		return diag.Errorf("failed to get cluster types with name %s: %s", clusterTypeName, err.Error())
	}

	if len(rsp.Results) == 0 {
		return diag.Errorf("no cluster type found with name %s", clusterTypeName)
	}

	clusterType := rsp.Results[0]

	d.SetId(clusterType.Id)

	createdStr := ""
	if clusterType.Created.IsSet() && clusterType.Created.Get() != nil {
		createdStr = clusterType.Created.Get().Format(time.RFC3339)
	}

	lastUpdatedStr := ""
	if clusterType.LastUpdated.IsSet() && clusterType.LastUpdated.Get() != nil {
		lastUpdatedStr = clusterType.LastUpdated.Get().Format(time.RFC3339)
	}

	// Set the fields directly in the resource data
	d.Set("id", clusterType.Id)
	d.Set("object_type", clusterType.ObjectType)
	d.Set("display", clusterType.Display)
	d.Set("url", clusterType.Url)
	d.Set("natural_slug", clusterType.NaturalSlug)
	d.Set("cluster_count", clusterType.ClusterCount)
	d.Set("name", clusterType.Name)
	d.Set("description", clusterType.Description)
	d.Set("created", createdStr)
	d.Set("last_updated", lastUpdatedStr)
	d.Set("notes_url", clusterType.NotesUrl)
	d.Set("custom_fields", clusterType.CustomFields)

	return diags
}