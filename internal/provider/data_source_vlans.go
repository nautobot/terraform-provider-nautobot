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
	_ datasource.DataSource              = &VLANsDataSource{}
	_ datasource.DataSourceWithConfigure = &VLANsDataSource{}
)

type VLANsDataSource struct {
	client *APIClient
}

type vlanItemModel struct {
	ID          types.String `tfsdk:"id"`
	Vid         types.Int64  `tfsdk:"vid"`
	Name        types.String `tfsdk:"name"`
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

type vlansDataSourceModel struct {
	VLANs []vlanItemModel `tfsdk:"vlans"`
}

func NewVLANsDataSource() datasource.DataSource {
	return &VLANsDataSource{}
}

func (d *VLANsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlans"
}

func (d *VLANsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all VLANs in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"vlans": dsschema.ListNestedAttribute{
				Description: "List of VLANs.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "The UUID of the VLAN.",
							Computed:    true,
						},
						"vid": dsschema.Int64Attribute{
							Description: "The ID (VID) of the VLAN.",
							Computed:    true,
						},
						"name": dsschema.StringAttribute{
							Description: "The name of the VLAN.",
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
				},
			},
		},
	}
}

func (d *VLANsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *VLANsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vlansDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"API client is not configured. This is a bug in the provider configuration.",
		)
		return
	}

	c := d.client.Client

	const pageLimit int32 = 200
	var offset int32 = 0

	state.VLANs = make([]vlanItemModel, 0)

	for {
		rsp, httpResp, err := c.IpamAPI.
			IpamVlansList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get VLANs list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, vlan := range results {
			var item vlanItemModel

			if vlan.Id == nil || *vlan.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid VLAN data",
					"VLANs list returned an item with no id (name: "+vlan.Name+")",
				)
				return
			}
			item.ID = types.StringValue(*vlan.Id)

			item.Vid = types.Int64Value(int64(vlan.Vid))
			item.Name = types.StringValue(vlan.Name)
			item.Description = types.StringValue(derefStr(vlan.Description))
			item.Created = nullableTimeStr(vlan.Created)
			item.LastUpdated = nullableTimeStr(vlan.LastUpdated)

			item.VLANGroupID = nullableFKStr(vlan.VlanGroup)

			statusName := ""
			if vlan.Status.Id != nil && vlan.Status.Id.String != nil {
				statusID := *vlan.Status.Id.String
				if statusID != "" {
					if name, err := getStatusName(ctx, c, statusID); err == nil {
						statusName = name
					}
				}
			}
			item.Status = types.StringValue(statusName)

			item.TenantID = nullableFKStr(vlan.Tenant)
			item.RoleID = nullableFKStr(vlan.Role)

			if len(vlan.Tags) > 0 {
				tagVals := make([]attr.Value, 0, len(vlan.Tags))
				for _, tag := range vlan.Tags {
					if tag.Id != nil && tag.Id.String != nil {
						tagVals = append(tagVals, types.StringValue(*tag.Id.String))
					}
				}
				item.TagsIDs = types.ListValueMust(types.StringType, tagVals)
			} else {
				item.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}

			if vlan.PrefixCount != nil {
				item.PrefixCount = types.Int64Value(int64(*vlan.PrefixCount))
			} else {
				item.PrefixCount = types.Int64Value(0)
			}

			item.Display = types.StringValue(vlan.Display)
			item.URL = types.StringValue(vlan.Url)
			item.NaturalSlug = types.StringValue(vlan.NaturalSlug)
			item.NotesURL = types.StringValue(vlan.NotesUrl)

			state.VLANs = append(state.VLANs, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read VLANs", map[string]any{"count": len(state.VLANs)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
