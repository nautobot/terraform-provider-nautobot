package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourcePrefix() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a Prefix in Nautobot by its associated VLAN ID.",

		ReadContext: dataSourcePrefixRead,

		Schema: map[string]*schema.Schema{
			"vlan_id": {
				Description: "The UUID of the VLAN to retrieve the prefix for.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The UUID of the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"prefix": {
				Description: "The prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description of the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "The status of the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				Description: "The ID of the role associated with the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tenant_id": {
				Description: "The ID of the tenant associated with the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"rir_id": {
				Description: "The ID of the RIR associated with the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"namespace_id": {
				Description: "The ID of the namespace associated with the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created": {
				Description: "The creation date of the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "The last update date of the prefix.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourcePrefixRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Get the VLAN ID from the Terraform configuration
	vlanID := d.Get("vlan_id").(string)

	// Prepare the VLAN ID as a []*string slice
	vlanIDList := []*string{&vlanID}

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

	// Fetch prefixes by VLAN ID
	rsp, _, err := c.IpamAPI.IpamPrefixesList(auth).VlanId(vlanIDList).Execute()
	if err != nil {
		return diag.Errorf("failed to get prefix for VLAN ID %s: %s", vlanID, err.Error())
	}

	if len(rsp.Results) == 0 {
		return diag.Errorf("no prefix found for VLAN ID %s", vlanID)
	}

	prefix := rsp.Results[0]

	d.SetId(prefix.Id)

	createdStr := ""
	if prefix.Created.IsSet() && prefix.Created.Get() != nil {
		createdStr = prefix.Created.Get().Format(time.RFC3339)
	}

	lastUpdatedStr := ""
	if prefix.LastUpdated.IsSet() && prefix.LastUpdated.Get() != nil {
		lastUpdatedStr = prefix.LastUpdated.Get().Format(time.RFC3339)
	}

	// Set the fields directly in the resource data
	d.Set("id", prefix.Id)
	d.Set("prefix", prefix.Prefix)
	d.Set("description", prefix.Description)
	d.Set("created", createdStr)
	d.Set("last_updated", lastUpdatedStr)

	// Handle nullable status
	if prefix.Status.Id != nil && prefix.Status.Id.String != nil {
		statusID := *prefix.Status.Id.String
		statusName, err := getStatusName(ctx, c, t, statusID)
		if err != nil {
			return diag.Errorf("failed to get status name for ID %s: %s", statusID, err.Error())
		}
		d.Set("status", statusName)
	}

	// Handle nullable Tenant
	if prefix.Tenant.IsSet() {
		if tenant := prefix.Tenant.Get(); tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
			d.Set("tenant_id", *tenant.Id.String)
		}
	}

	// Handle nullable Role
	if prefix.Role.IsSet() {
		if role := prefix.Role.Get(); role != nil && role.Id != nil && role.Id.String != nil {
			d.Set("role_id", *role.Id.String)
		}
	}

	// Handle nullable RIR
	if prefix.Rir.IsSet() {
		if rir := prefix.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
			d.Set("rir_id", *rir.Id.String)
		}
	}

	// Handle nullable Namespace (without using IsSet)
	if prefix.Namespace != nil && prefix.Namespace.Id != nil && prefix.Namespace.Id.String != nil {
		d.Set("namespace_id", *prefix.Namespace.Id.String)
	}

	return diags
}
