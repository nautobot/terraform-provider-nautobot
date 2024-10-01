package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceCluster() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a cluster in Nautobot",

		CreateContext: resourceClusterCreate,
		ReadContext:   resourceClusterRead,
		UpdateContext: resourceClusterUpdate,
		DeleteContext: resourceClusterDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Cluster's name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"comments": {
				Description: "Comments or notes about the cluster.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cluster_type_id": {
				Description: "ID of the Cluster's type. This can be sourced from the cluster_type resource or data source.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"cluster_group_id": {
				Description: "ID of the Cluster's group.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tenant_id": {
				Description: "ID of the Tenant associated with the cluster.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"location_id": {
				Description: "ID of the Location of the cluster.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"tags_ids": {
				Description: "IDs of the Tags associated with the cluster.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "Creation date of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Last update date of the cluster.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	clusterName := d.Get("name").(string)
	existingClusters, _, err := c.VirtualizationAPI.VirtualizationClustersList(auth).Name([]string{clusterName}).Execute()
	if err != nil {
		return diag.Errorf("failed to list clusters: %s", err.Error())
	}

	// If a cluster with the same name exists, use its ID and skip creation
	if len(existingClusters.Results) > 0 {
		d.SetId(existingClusters.Results[0].Id)
		return resourceClusterRead(ctx, d, meta)
	}

	// Prepare ClusterRequest
	var cluster nb.ClusterRequest
	cluster.Name = clusterName
	cluster.ClusterType = nb.BulkWritableCableRequestStatus{
		Id: &nb.BulkWritableCableRequestStatusId{
			String: stringPtr(d.Get("cluster_type_id").(string)),
		},
	}

	// Optional fields
	if v, ok := d.GetOk("comments"); ok {
		comments := v.(string)
		cluster.Comments = &comments
	}

	if v, ok := d.GetOk("cluster_group_id"); ok {
		var clusterGroup nb.NullableBulkWritableCircuitRequestTenant
		clusterGroup.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(v.(string)),
			},
		})
		cluster.ClusterGroup = clusterGroup
	}

	if v, ok := d.GetOk("tenant_id"); ok {
		var tenant nb.NullableBulkWritableCircuitRequestTenant
		tenant.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(v.(string)),
			},
		})
		cluster.Tenant = tenant
	}

	if v, ok := d.GetOk("location_id"); ok {
		var location nb.NullableBulkWritableCircuitRequestTenant
		location.Set(&nb.BulkWritableCircuitRequestTenant{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: stringPtr(v.(string)),
			},
		})
		cluster.Location = location
	}

	if v, ok := d.GetOk("tags_ids"); ok {
		var tags []nb.BulkWritableCableRequestStatus
		for _, tag := range v.([]interface{}) {
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.BulkWritableCableRequestStatusId{
					String: stringPtr(tag.(string)),
				},
			})
		}
		cluster.Tags = tags
	}

	// Create the cluster
	rsp, _, err := c.VirtualizationAPI.VirtualizationClustersCreate(auth).ClusterRequest(cluster).Execute()
	if err != nil {
		return diag.Errorf("failed to create cluster: %s", err.Error())
	}

	// Set resource ID (Cluster ID)
	d.SetId(rsp.Id)

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Fetch cluster by ID
	clusterId := d.Id()
	cluster, _, err := c.VirtualizationAPI.VirtualizationClustersRetrieve(auth, clusterId).Execute()
	if err != nil {
		return diag.Errorf("failed to read cluster: %s", err.Error())
	}

	// Map the retrieved data back to Terraform state
	d.Set("name", cluster.Name)

	// Extract cluster_type_id safely
	if cluster.ClusterType.Id != nil && cluster.ClusterType.Id.String != nil {
		d.Set("cluster_type_id", *cluster.ClusterType.Id.String)
	}

	// Check if comments exist before setting
	if cluster.Comments != nil {
		d.Set("comments", *cluster.Comments)
	}

	// Handle nullable cluster group
	if cluster.ClusterGroup.IsSet() {
		if clusterGroup := cluster.ClusterGroup.Get(); clusterGroup != nil && clusterGroup.Id != nil {
			d.Set("cluster_group_id", *clusterGroup.Id.String)
		}
	}

	// Handle nullable tenant
	if cluster.Tenant.IsSet() {
		if tenant := cluster.Tenant.Get(); tenant != nil && tenant.Id != nil {
			d.Set("tenant_id", *tenant.Id.String)
		}
	}

	// Handle nullable location
	if cluster.Location.IsSet() {
		if location := cluster.Location.Get(); location != nil && location.Id != nil {
			d.Set("location_id", *location.Id.String)
		}
	}

	// Set tags
	if len(cluster.Tags) > 0 {
		var tags []string
		for _, tag := range cluster.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tags = append(tags, *tag.Id.String)
			}
		}
		d.Set("tags_ids", tags)
	}

	d.Set("created", cluster.Created)
	d.Set("last_updated", cluster.LastUpdated)

	return nil
}

func resourceClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	clusterId := d.Id()

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

	var cluster nb.PatchedClusterRequest

	// Update the fields that have changed
	if d.HasChange("name") {
		name := d.Get("name").(string)
		cluster.Name = &name // Set the pointer for the name
	}
	if d.HasChange("comments") {
		comments := d.Get("comments").(string)
		cluster.Comments = &comments
	}
	if d.HasChange("cluster_type_id") {
		clusterTypeID := d.Get("cluster_type_id").(string)
		cluster.ClusterType = &nb.BulkWritableCableRequestStatus{
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &clusterTypeID, // Pass pointer for the string value
			},
		}
	}
	if d.HasChange("cluster_group_id") {
		clusterGroupID := d.Get("cluster_group_id").(string)
		clusterGroup := &nb.BulkWritableCircuitRequestTenant{ // Create the cluster group as a pointer
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &clusterGroupID, // Pass pointer for the string value
			},
		}
		cluster.ClusterGroup.Set(clusterGroup) // Pass the pointer to Set()
	}
	if d.HasChange("tenant_id") {
		tenantID := d.Get("tenant_id").(string)
		tenant := &nb.BulkWritableCircuitRequestTenant{ // Create the tenant as a pointer
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &tenantID, // Pass pointer for the string value
			},
		}
		cluster.Tenant.Set(tenant) // Pass the pointer to Set()
	}
	if d.HasChange("location_id") {
		locationID := d.Get("location_id").(string)
		location := &nb.BulkWritableCircuitRequestTenant{ // Create the location as a pointer
			Id: &nb.BulkWritableCableRequestStatusId{
				String: &locationID, // Pass pointer for the string value
			},
		}
		cluster.Location.Set(location) // Pass the pointer to Set()
	}
	if d.HasChange("tags_ids") {
		var tags []nb.BulkWritableCableRequestStatus
		for _, tag := range d.Get("tags_ids").([]interface{}) {
			tagID := tag.(string)
			tags = append(tags, nb.BulkWritableCableRequestStatus{
				Id: &nb.BulkWritableCableRequestStatusId{
					String: &tagID, // Pass pointer for the string value
				},
			})
		}
		cluster.Tags = tags
	}

	// Call the API to update the cluster
	_, _, err := c.VirtualizationAPI.VirtualizationClustersPartialUpdate(auth, clusterId).PatchedClusterRequest(cluster).Execute()
	if err != nil {
		return diag.Errorf("failed to update cluster: %s", err.Error())
	}

	return resourceClusterRead(ctx, d, meta)
}

func resourceClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Delete the cluster by ID
	clusterId := d.Id()
	_, err := c.VirtualizationAPI.VirtualizationClustersDestroy(auth, clusterId).Execute()
	if err != nil {
		return diag.Errorf("failed to delete cluster: %s", err.Error())
	}

	// Clear the ID
	d.SetId("")

	return nil
}
