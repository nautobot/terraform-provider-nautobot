package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nb "github.com/nautobot/go-nautobot/v2"
)

func dataSourceManufacturer() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a specific manufacturer in Nautobot.",

		ReadContext: dataSourceManufacturerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the manufacturer to retrieve.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "Manufacturer's UUID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"object_type": {
				Description: "Object type of the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"display": {
				Description: "Human friendly display value for the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"url": {
				Description: "URL of the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"natural_slug": {
				Description: "Natural slug for the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Manufacturer's description.",
				Type:        schema.TypeString,
				Computed:    true,
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
				Description: "Notes URL for the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceManufacturerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token

	// Get the manufacturer name from the Terraform configuration
	manufacturerName := d.Get("name").(string)

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

	// Fetch manufacturer by name
	rsp, _, err := c.DcimAPI.DcimManufacturersList(auth).Name([]string{manufacturerName}).Execute()
	if err != nil {
		return diag.Errorf("failed to get manufacturer with name %s: %s", manufacturerName, err.Error())
	}

	if len(rsp.Results) == 0 {
		return diag.Errorf("no manufacturer found with name %s", manufacturerName)
	}

	manufacturer := rsp.Results[0]

	d.SetId(manufacturer.Id)

	createdStr := ""
	if manufacturer.Created.IsSet() && manufacturer.Created.Get() != nil {
		createdStr = manufacturer.Created.Get().Format(time.RFC3339)
	}

	lastUpdatedStr := ""
	if manufacturer.LastUpdated.IsSet() && manufacturer.LastUpdated.Get() != nil {
		lastUpdatedStr = manufacturer.LastUpdated.Get().Format(time.RFC3339)
	}

	// Set the fields directly in the resource data
	d.Set("id", manufacturer.Id)
	d.Set("object_type", manufacturer.ObjectType)
	d.Set("display", manufacturer.Display)
	d.Set("url", manufacturer.Url)
	d.Set("natural_slug", manufacturer.NaturalSlug)
	d.Set("name", manufacturer.Name)
	d.Set("description", manufacturer.Description)
	d.Set("created", createdStr)
	d.Set("last_updated", lastUpdatedStr)
	d.Set("notes_url", manufacturer.NotesUrl)

	return diags
}
