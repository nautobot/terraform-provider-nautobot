package provider

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceManufacturers() *schema.Resource {
	return &schema.Resource{
		Description: "Manufacturer data source in the Terraform provider Nautobot.",

		ReadContext: dataSourceManufacturersRead,

		Schema: map[string]*schema.Schema{
			"manufacturers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Manufacturer's UUID.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"object_type": {
							Description: "Object type of the Manufacturer.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"display": {
							Description: "Human friendly display value for the Manufacturer.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"url": {
							Description: "URL of the Manufacturer.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"natural_slug": {
							Description: "Natural slug for the Manufacturer.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Manufacturer's name.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"description": {
							Description: "Manufacturer's description.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"created": {
							Description: "Manufacturer's creation date.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"last_updated": {
							Description: "Manufacturer's last update.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"notes_url": {
							Description: "Notes URL for the Manufacturer.",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Use this as reference: https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers#implement-read
func dataSourceManufacturersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	s := meta.(*apiClient).Server
	t := meta.(*apiClient).Token.token
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

	rsp, _, err := c.DcimAPI.DcimManufacturersList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to get manufacturers list from %s: %s", s, err.Error())
	}

	results := rsp.Results

	list := make([]map[string]interface{}, 0)

	// Iterate over the results and map each manufacturer to the format expected by Terraform
	for _, manufacturer := range results {
		createdStr := ""
		if manufacturer.Created.IsSet() && manufacturer.Created.Get() != nil {
			createdStr = manufacturer.Created.Get().Format(time.RFC3339)
		}

		lastUpdatedStr := ""
		if manufacturer.LastUpdated.IsSet() && manufacturer.LastUpdated.Get() != nil {
			lastUpdatedStr = manufacturer.LastUpdated.Get().Format(time.RFC3339)
		}
		itemMap := map[string]interface{}{
			"id":           manufacturer.Id,
			"object_type":  manufacturer.ObjectType,
			"display":      manufacturer.Display,
			"url":          manufacturer.Url,
			"natural_slug": manufacturer.NaturalSlug,
			"name":         manufacturer.Name,
			"description":  manufacturer.Description,
			"created":      createdStr,
			"last_updated": lastUpdatedStr,
			"notes_url":    manufacturer.NotesUrl,
		}
		list = append(list, itemMap)
	}

	if err := d.Set("manufacturers", list); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
