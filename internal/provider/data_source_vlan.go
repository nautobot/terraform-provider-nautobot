package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &VLANDataSource{}
	_ datasource.DataSourceWithConfigure = &VLANDataSource{}
)

type VLANDataSource struct {
	client *APIClient
}

type vlanDataSourceModel struct {
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Vid         types.Int64  `tfsdk:"vid"`
	Description types.String `tfsdk:"description"`
	VLANGroupID types.String `tfsdk:"vlan_group_id"`
	Status      types.String `tfsdk:"status"`
	TenantID    types.String `tfsdk:"tenant_id"`
	RoleID      types.String `tfsdk:"role_id"`
	TagsIDs     types.List   `tfsdk:"tags_ids"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	PrefixCount types.Int64  `tfsdk:"prefix_count"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

func NewVLANDataSource() datasource.DataSource {
	return &VLANDataSource{}
}

func (d *VLANDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (d *VLANDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a specific VLAN in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"name": dsschema.StringAttribute{
				Description: "The name of the VLAN to retrieve.",
				Required:    true,
			},
			"id": dsschema.StringAttribute{
				Description: "The UUID of the VLAN.",
				Computed:    true,
			},
			"vid": dsschema.Int64Attribute{
				Description: "The ID (VID) of the VLAN.",
				Computed:    true,
			},
			"description": dsschema.StringAttribute{
				Description: "Description of the VLAN.",
				Computed:    true,
			},
			"vlan_group_id": dsschema.StringAttribute{
				Description: "The ID of the VLAN group.",
				Computed:    true,
			},
			"status": dsschema.StringAttribute{
				Description: "The status of the VLAN (name).",
				Computed:    true,
			},
			"tenant_id": dsschema.StringAttribute{
				Description: "The ID of the tenant associated with the VLAN.",
				Computed:    true,
			},
			"role_id": dsschema.StringAttribute{
				Description: "The ID of the role associated with the VLAN.",
				Computed:    true,
			},
			"tags_ids": dsschema.ListAttribute{
				Description: "The IDs of the tags associated with the VLAN.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"created": dsschema.StringAttribute{
				Description: "The creation date of the VLAN (RFC3339).",
				Computed:    true,
			},
			"last_updated": dsschema.StringAttribute{
				Description: "The last update date of the VLAN (RFC3339).",
				Computed:    true,
			},

			"prefix_count": dsschema.Int64Attribute{
				Description: "Number of prefixes associated with this VLAN.",
				Computed:    true,
			},
			"display": dsschema.StringAttribute{
				Description: "Human-friendly display value.",
				Computed:    true,
			},
			"url": dsschema.StringAttribute{
				Description: "API URL of the VLAN.",
				Computed:    true,
			},
			"natural_slug": dsschema.StringAttribute{
				Description: "Natural slug for the VLAN.",
				Computed:    true,
			},
			"notes_url": dsschema.StringAttribute{
				Description: "Notes URL for the VLAN.",
				Computed:    true,
			},
		},
	}
}

func (d *VLANDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *VLANDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data vlanDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	vlanName := data.Name.ValueString()

	// Fetch VLAN by name
	rsp, httpResp, err := c.IpamAPI.
		IpamVlansList(ctx).
		Name([]string{vlanName}).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get VLAN",
			httpErr(err, httpResp),
		)
		return
	}
	if len(rsp.Results) == 0 {
		resp.Diagnostics.AddError(
			"VLAN not found",
			"No VLAN found with name "+vlanName,
		)
		return
	}

	vlan := rsp.Results[0]

	if vlan.Id == nil || *vlan.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid VLAN data",
			"VLAN "+vlanName+" returned no id",
		)
		return
	}
	vlanID := *vlan.Id
	data.ID = types.StringValue(vlanID)

	data.Vid = types.Int64Value(int64(vlan.Vid))
	data.Name = types.StringValue(vlan.Name)
	data.Description = types.StringValue(derefStr(vlan.Description))
	data.Created = nullableTimeStr(vlan.Created)
	data.LastUpdated = nullableTimeStr(vlan.LastUpdated)

	data.VLANGroupID = nullableFKStr(vlan.VlanGroup)

	statusName := ""
	if vlan.Status.Id != nil && vlan.Status.Id.String != nil {
		statusID := *vlan.Status.Id.String
		if statusID != "" {
			if name, err := getStatusName(ctx, c, statusID); err == nil {
				statusName = name
			}
		}
	}
	data.Status = types.StringValue(statusName)

	data.TenantID = nullableFKStr(vlan.Tenant)
	data.RoleID = nullableFKStr(vlan.Role)

	if len(vlan.Tags) > 0 {
		tagVals := make([]attr.Value, 0, len(vlan.Tags))
		for _, tag := range vlan.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tagVals = append(tagVals, types.StringValue(*tag.Id.String))
			}
		}
		data.TagsIDs = types.ListValueMust(types.StringType, tagVals)
	} else {
		data.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if vlan.PrefixCount != nil {
		data.PrefixCount = types.Int64Value(int64(*vlan.PrefixCount))
	} else {
		data.PrefixCount = types.Int64Value(0)
	}

	data.Display = types.StringValue(vlan.Display)
	data.URL = types.StringValue(vlan.Url)
	data.NaturalSlug = types.StringValue(vlan.NaturalSlug)
	data.NotesURL = types.StringValue(vlan.NotesUrl)

	tflog.Debug(ctx, "read VLAN", map[string]any{"id": vlanID, "name": vlanName})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
