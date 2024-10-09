package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceVLAN() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a specific VLAN in Nautobot.",

		ReadContext: dataSourceVLANRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the VLAN to retrieve.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The UUID of the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vid": {
				Description: "The ID (VID) of the VLAN.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"description": {
				Description: "Description of the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vlan_group_id": {
				Description: "The ID of the VLAN group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "The status of the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tenant_id": {
				Description: "The ID of the tenant associated with the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				Description: "The ID of the role associated with the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"locations": {
				Description: "The IDs of the locations associated with the VLAN.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags_ids": {
				Description: "The IDs of the tags associated with the VLAN.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Description: "The creation date of the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "The last update date of the VLAN.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceVLANRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Get the VLAN name from the Terraform configuration
	vlanName := d.Get("name").(string)

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

	// Fetch VLAN by name
	rsp, _, err := c.IpamAPI.IpamVlansList(auth).Name([]string{vlanName}).Execute()
	if err != nil {
		return diag.Errorf("failed to get VLAN with name %s: %s", vlanName, err.Error())
	}

	if len(rsp.Results) == 0 {
		return diag.Errorf("no VLAN found with name %s", vlanName)
	}

	vlan := rsp.Results[0]

	d.SetId(vlan.Id)

	createdStr := ""
	if vlan.Created.IsSet() && vlan.Created.Get() != nil {
		createdStr = vlan.Created.Get().Format(time.RFC3339)
	}

	lastUpdatedStr := ""
	if vlan.LastUpdated.IsSet() && vlan.LastUpdated.Get() != nil {
		lastUpdatedStr = vlan.LastUpdated.Get().Format(time.RFC3339)
	}

	// Set the fields directly in the resource data
	d.Set("id", vlan.Id)
	d.Set("vid", vlan.Vid)
	d.Set("name", vlan.Name)
	d.Set("description", vlan.Description)
	d.Set("created", createdStr)
	d.Set("last_updated", lastUpdatedStr)

	// Handle nullable VlanGroup
	if vlan.VlanGroup.IsSet() {
		if vlanGroup := vlan.VlanGroup.Get(); vlanGroup != nil && vlanGroup.Id != nil {
			d.Set("vlan_group_id", *vlanGroup.Id)
		}
	}

	if vlan.Status.Id != nil && vlan.Status.Id.String != nil {
		statusID := *vlan.Status.Id.String
		statusName, err := getStatusName(ctx, c, t, statusID)
		if err != nil {
			return diag.Errorf("failed to get status name for ID %s: %s", statusID, err.Error())
		}
		d.Set("status", statusName)
	}

	// Handle nullable Tenant
	if vlan.Tenant.IsSet() {
		if tenant := vlan.Tenant.Get(); tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
			d.Set("tenant_id", *tenant.Id.String)
		}
	}

	// Handle nullable Role
	if vlan.Role.IsSet() {
		if role := vlan.Role.Get(); role != nil && role.Id != nil && role.Id.String != nil {
			d.Set("role_id", *role.Id.String)
		}
	}

	// Handle locations
	var locations []string
	for _, location := range vlan.Locations {
		if location.Id != nil && location.Id.String != nil {
			locations = append(locations, *location.Id.String)
		}
	}
	d.Set("locations", locations)

	// Handle Tags
	var tags []string
	for _, tag := range vlan.Tags {
		if tag.Id != nil && tag.Id.String != nil {
			tags = append(tags, *tag.Id.String)
		}
	}
	d.Set("tags_ids", tags)

	return diags
}
