package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	nb "github.com/nautobot/go-nautobot/v2"
)

func resourceManufacturer() *schema.Resource {
	return &schema.Resource{
		Description: "This object manages a manufacturer in Nautobot",

		CreateContext: resourceManufacturerCreate,
		ReadContext:   resourceManufacturerRead,
		UpdateContext: resourceManufacturerUpdate,
		DeleteContext: resourceManufacturerDelete,

		Schema: map[string]*schema.Schema{
			"created": {
				Description: "Manufacturer's creation date.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Manufacturer's description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"display": {
				Description: "Manufacturer's display name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"id": {
				Description: "Manufacturer's UUID.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_updated": {
				Description: "Manufacturer's last update date.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Manufacturer's name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"notes_url": {
				Description: "Notes URL for the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"url": {
				Description: "Manufacturer's URL.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"object_type": {
				Description: "Object type of the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"natural_slug": {
				Description: "Natural slug for the manufacturer.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceManufacturerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// Check if a manufacturer with the same name already exists
	name := d.Get("name").(string)

	existingManufacturersResp, _, err := c.DcimAPI.DcimManufacturersList(auth).Execute()
	if err != nil {
		return diag.Errorf("failed to check existing manufacturers on %s : %s", s, err.Error())
	}

	// Search through the results for a manufacturer with the given name
	for _, manufacturer := range existingManufacturersResp.Results {
		if manufacturer.Name == name {
			// Manufacturer already exists, set the ID and exit
			d.SetId(manufacturer.Id)
			return resourceManufacturerRead(ctx, d, meta)
		}
	}

	// Create a new manufacturer
	var m nb.ManufacturerRequest
	m.Name = name

	if v, ok := d.GetOk("description"); ok {
		desc := v.(string)
		m.Description = &desc
	}
	rsp, _, err := c.DcimAPI.DcimManufacturersCreate(auth).ManufacturerRequest(m).Execute()
	if err != nil {
		return diag.Errorf("failed to create manufacturer %s on %s: %s", m.Name, s, err.Error())
	}

	tflog.Trace(ctx, "manufacturer created", map[string]interface{}{
		"name": m.Name,
	})

	d.SetId(rsp.Id)

	return resourceManufacturerRead(ctx, d, meta)
}

func resourceManufacturerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	t := meta.(*apiClient).Token.token
	id := d.Get("id").(string)

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

	// Fetch manufacturer by ID
	manufacturer, _, err := c.DcimAPI.DcimManufacturersRetrieve(auth, id).Execute()
	if err != nil {
		return diag.Errorf("failed to get manufacturer %s: %s", id, err.Error())
	}

	// Set the Terraform state from the retrieved manufacturer data
	d.Set("name", manufacturer.Name)
	if manufacturer.Created.IsSet() && manufacturer.Created.Get() != nil {
		d.Set("created", manufacturer.Created.Get().Format(time.RFC3339))
	}
	if manufacturer.LastUpdated.IsSet() && manufacturer.LastUpdated.Get() != nil {
		d.Set("last_updated", manufacturer.LastUpdated.Get().Format(time.RFC3339))
	}
	d.Set("description", manufacturer.Description)
	d.Set("display", manufacturer.Display)
	d.Set("id", manufacturer.Id)
	d.Set("notes_url", manufacturer.NotesUrl)
	d.Set("url", manufacturer.Url)
	d.Set("object_type", manufacturer.ObjectType)
	d.Set("natural_slug", manufacturer.NaturalSlug)

	return nil
}

func resourceManufacturerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*apiClient).Client
	s := meta.(*apiClient).Server
	t := meta.(*apiClient).Token.token

	id := d.Get("id").(string)

	var m nb.PatchedManufacturerRequest

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

	if d.HasChange("name") {
		name := d.Get("name").(string)
		m.Name = &name
	}

	if d.HasChange("description") {
		desc := d.Get("description").(string)
		m.Description = &desc
	}

	_, _, err := c.DcimAPI.DcimManufacturersPartialUpdate(auth, id).PatchedManufacturerRequest(m).Execute()
	if err != nil {
		return diag.Errorf("failed to update manufacturer %s on %s: %s", id, s, err.Error())
	}

	tflog.Trace(ctx, "manufacturer updated", map[string]interface{}{
		"id": id,
	})

	return resourceManufacturerRead(ctx, d, meta)
}

func resourceManufacturerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	s := meta.(*apiClient).Server
	t := meta.(*apiClient).Token.token

	id := d.Get("id").(string)

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

	_, err := c.DcimAPI.DcimManufacturersDestroy(auth, id).Execute()
	if err != nil {
		return diag.Errorf("failed to delete manufacturer %s on %s: %s", id, s, err.Error())
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
