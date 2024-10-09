package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourcePrefixes() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about all prefixes in Nautobot.",

		ReadContext: dataSourcePrefixesRead,

		Schema: map[string]*schema.Schema{
			"prefixes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				},
			},
		},
	}
}

func dataSourcePrefixesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	rsp, _, err := c.IpamAPI.IpamPrefixesList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get prefixes: %s", err.Error())
	}

	results := rsp.Results
	list := make([]map[string]interface{}, 0)

	for _, prefix := range results {
		createdStr := ""
		if prefix.Created.IsSet() && prefix.Created.Get() != nil {
			createdStr = prefix.Created.Get().Format(time.RFC3339)
		}

		lastUpdatedStr := ""
		if prefix.LastUpdated.IsSet() && prefix.LastUpdated.Get() != nil {
			lastUpdatedStr = prefix.LastUpdated.Get().Format(time.RFC3339)
		}

		itemMap := map[string]interface{}{
			"id":           prefix.Id,
			"prefix":       prefix.Prefix,
			"description":  prefix.Description,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
		}

		if prefix.Status.Id != nil && prefix.Status.Id.String != nil {
			statusID := *prefix.Status.Id.String
			statusName, err := getStatusName(ctx, c, t, statusID)
			if err != nil {
				return diag.Errorf("failed to get status name for ID %s: %s", statusID, err.Error())
			}
			itemMap["status"] = statusName
		}

		if prefix.Tenant.IsSet() {
			if tenant := prefix.Tenant.Get(); tenant != nil && tenant.Id != nil && tenant.Id.String != nil {
				itemMap["tenant_id"] = *tenant.Id.String
			}
		}

		if prefix.Role.IsSet() {
			if role := prefix.Role.Get(); role != nil && role.Id != nil && role.Id.String != nil {
				itemMap["role_id"] = *role.Id.String
			}
		}

		if prefix.Rir.IsSet() {
			if rir := prefix.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
				itemMap["rir_id"] = *rir.Id.String
			}
		}

		if prefix.Namespace != nil && prefix.Namespace.Id != nil && prefix.Namespace.Id.String != nil {
			itemMap["namespace_id"] = *prefix.Namespace.Id.String
		}

		list = append(list, itemMap)
	}

	if err := d.Set("prefixes", list); err != nil {
		return diag.FromErr(err)
	}

	// Set ID for the data source
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
