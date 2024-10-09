package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceClusterType() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a cluster type in Nautobot.",

		CreateContext: resourceClusterTypeCreate,
		ReadContext:   resourceClusterTypeRead,
		UpdateContext: resourceClusterTypeUpdate,
		DeleteContext: resourceClusterTypeDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Cluster type's UUID.",
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
				Description: "Cluster type's name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description for the cluster type.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"created": {
				Description: "Creation date of the cluster type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Last update date of the cluster type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"notes_url": {
				Description: "Notes URL for the cluster type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceClusterTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	clusterTypeName := d.Get("name").(string)
	existingClusterTypes, _, err := c.VirtualizationAPI.VirtualizationClusterTypesList(auth).Name([]string{clusterTypeName}).Execute()
	if err != nil {
		return diag.Errorf("failed to list cluster types: %s", err.Error())
	}

	// If a cluster type with the same name exists, use its ID and skip creation
	if len(existingClusterTypes.Results) > 0 {
		d.SetId(existingClusterTypes.Results[0].Id)
		return resourceClusterTypeRead(ctx, d, meta)
	}

	// Prepare ClusterTypeRequest
	var clusterType nb.ClusterTypeRequest
	clusterType.Name = clusterTypeName

	if v, ok := d.GetOk("description"); ok {
		description := v.(string)
		clusterType.Description = &description
	}

	// Create the cluster type using VirtualizationAPI
	rsp, _, err := c.VirtualizationAPI.VirtualizationClusterTypesCreate(auth).ClusterTypeRequest(clusterType).Execute()
	if err != nil {
		return diag.Errorf("failed to create cluster type: %s", err.Error())
	}

	// Set resource ID (Cluster Type ID)
	d.SetId(rsp.Id)

	return resourceClusterTypeRead(ctx, d, meta)
}

func resourceClusterTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Fetch cluster type by ID using VirtualizationAPI
	clusterTypeId := d.Id()
	clusterType, _, err := c.VirtualizationAPI.VirtualizationClusterTypesRetrieve(auth, clusterTypeId).Execute()
	if err != nil {
		return diag.Errorf("failed to read cluster type: %s", err.Error())
	}

	// Map the retrieved data back to Terraform state
	d.Set("name", clusterType.Name)
	d.Set("object_type", clusterType.ObjectType)
	d.Set("display", clusterType.Display)
	d.Set("url", clusterType.Url)
	d.Set("natural_slug", clusterType.NaturalSlug)
	if clusterType.Description != nil {
		d.Set("description", *clusterType.Description)
	}
	d.Set("created", clusterType.Created)
	d.Set("last_updated", clusterType.LastUpdated)
	d.Set("notes_url", clusterType.NotesUrl)

	return nil
}

func resourceClusterTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	clusterTypeId := d.Id()

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

	var clusterType nb.PatchedClusterTypeRequest

	// Update the fields that have changed
	if d.HasChange("name") {
		name := d.Get("name").(string)
		clusterType.Name = &name
	}
	if d.HasChange("description") {
		description := d.Get("description").(string)
		clusterType.Description = &description
	}

	// Call the API to update the cluster type
	_, _, err := c.VirtualizationAPI.VirtualizationClusterTypesPartialUpdate(auth, clusterTypeId).PatchedClusterTypeRequest(clusterType).Execute()
	if err != nil {
		return diag.Errorf("failed to update cluster type: %s", err.Error())
	}

	return resourceClusterTypeRead(ctx, d, meta)
}

func resourceClusterTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Delete the cluster type by ID using VirtualizationAPI
	clusterTypeId := d.Id()
	_, err := c.VirtualizationAPI.VirtualizationClusterTypesDestroy(auth, clusterTypeId).Execute()
	if err != nil {
		return diag.Errorf("failed to delete cluster type: %s", err.Error())
	}

	// Clear the ID
	d.SetId("")

	return nil
}
