package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceAvailableIP() *schema.Resource {
	return &schema.Resource{
		Description: "This data source retrieves an available IP address from a given prefix in Nautobot",

		ReadContext: dataSourceAvailableIPRead,

		Schema: map[string]*schema.Schema{
			"prefix_id": {
				Description: "The ID of the prefix from which to retrieve an available IP.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ip_version": {
				Description: "The version of the IP address (4 or 6).",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"address": {
				Description: "The available IP address.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceAvailableIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	prefixID := d.Get("prefix_id").(string)

	// Fetch the available IPs from the given prefix
	availableIPs, _, err := c.IpamAPI.IpamPrefixesAvailableIpsList(auth, prefixID).Execute()
	if err != nil {
		return diag.Errorf("failed to retrieve available IPs from prefix %s: %s", prefixID, err.Error())
	}

	// Check if there are available IPs
	if len(availableIPs) == 0 {
		return diag.Errorf("no available IP addresses found for prefix %s", prefixID)
	}

	// Use the first available IP from the list
	availableIP := availableIPs[0]

	// Set values in Terraform state
	d.Set("ip_version", availableIP.IpVersion)
	d.Set("address", availableIP.Address)

	// Set resource ID to the available IP address
	d.SetId(fmt.Sprintf("%s-%d", availableIP.Address, availableIP.IpVersion))

	return nil
}
