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
	_ datasource.DataSource              = &PrefixesDataSource{}
	_ datasource.DataSourceWithConfigure = &PrefixesDataSource{}
)

type PrefixesDataSource struct {
	client *APIClient
}

type prefixItemModel struct {
	ID            types.String `tfsdk:"id"`
	Prefix        types.String `tfsdk:"prefix"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	ParentID      types.String `tfsdk:"parent_id"`
	RoleID        types.String `tfsdk:"role_id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	RirID         types.String `tfsdk:"rir_id"`
	NamespaceID   types.String `tfsdk:"namespace_id"`
	VLANID        types.String `tfsdk:"vlan_id"`
	Created       types.String `tfsdk:"created"`
	LastUpdated   types.String `tfsdk:"last_updated"`
	Network       types.String `tfsdk:"network"`
	Broadcast     types.String `tfsdk:"broadcast"`
	PrefixLength  types.Int64  `tfsdk:"prefix_length"`
	IPVersion     types.Int64  `tfsdk:"ip_version"`
	DateAllocated types.String `tfsdk:"date_allocated"`
	TagsIDs       types.List   `tfsdk:"tags_ids"`
	Display       types.String `tfsdk:"display"`
	URL           types.String `tfsdk:"url"`
	NaturalSlug   types.String `tfsdk:"natural_slug"`
	NotesURL      types.String `tfsdk:"notes_url"`
}

type prefixesDataSourceModel struct {
	Prefixes []prefixItemModel `tfsdk:"prefixes"`
}

func NewPrefixesDataSource() datasource.DataSource {
	return &PrefixesDataSource{}
}

func (d *PrefixesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefixes"
}

func (d *PrefixesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all prefixes in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"prefixes": dsschema.ListNestedAttribute{
				Description: "List of prefixes.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Description: "The UUID of the prefix.",
							Computed:    true,
						},
						"prefix": dsschema.StringAttribute{
							Description: "The prefix in CIDR notation.",
							Computed:    true,
						},
						"description": dsschema.StringAttribute{
							Description: "Description of the prefix.",
							Computed:    true,
						},
						"status": dsschema.StringAttribute{
							Description: "The status of the prefix (name).",
							Computed:    true,
						},
						"parent_id": dsschema.StringAttribute{
							Description: "The ID of the parent of this prefix.",
							Computed:    true,
						},
						"role_id": dsschema.StringAttribute{
							Description: "The ID of the role associated with the prefix.",
							Computed:    true,
						},
						"tenant_id": dsschema.StringAttribute{
							Description: "The ID of the tenant associated with the prefix.",
							Computed:    true,
						},
						"rir_id": dsschema.StringAttribute{
							Description: "The ID of the RIR associated with the prefix.",
							Computed:    true,
						},
						"namespace_id": dsschema.StringAttribute{
							Description: "The ID of the namespace associated with the prefix.",
							Computed:    true,
						},
						"vlan_id": dsschema.StringAttribute{
							Description: "The UUID of the VLAN the prefix belongs to.",
							Computed:    true,
						},
						"created": dsschema.StringAttribute{
							Description: "The creation date of the prefix (RFC3339).",
							Computed:    true,
						},
						"last_updated": dsschema.StringAttribute{
							Description: "The last update date of the prefix (RFC3339).",
							Computed:    true,
						},
						"network": dsschema.StringAttribute{
							Description: "IPv4 or IPv6 network address.",
							Computed:    true,
						},
						"broadcast": dsschema.StringAttribute{
							Description: "IPv4 or IPv6 broadcast address.",
							Computed:    true,
						},
						"prefix_length": dsschema.Int64Attribute{
							Description: "Length of the network prefix, in bits.",
							Computed:    true,
						},
						"ip_version": dsschema.Int64Attribute{
							Description: "IP version of the prefix (4 or 6).",
							Computed:    true,
						},
						"date_allocated": dsschema.StringAttribute{
							Description: "Date this prefix was allocated/reserved (RFC3339).",
							Computed:    true,
						},
						"tags_ids": dsschema.ListAttribute{
							Description: "The IDs of the tags associated with the prefix.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"display": dsschema.StringAttribute{
							Description: "Human-friendly display value.",
							Computed:    true,
						},
						"url": dsschema.StringAttribute{
							Description: "API URL of the prefix.",
							Computed:    true,
						},
						"natural_slug": dsschema.StringAttribute{
							Description: "Natural slug for the prefix.",
							Computed:    true,
						},
						"notes_url": dsschema.StringAttribute{
							Description: "Notes URL for the prefix.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *PrefixesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *PrefixesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state prefixesDataSourceModel

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

	state.Prefixes = make([]prefixItemModel, 0)

	for {
		rsp, httpResp, err := c.IpamAPI.
			IpamPrefixesList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("prefix").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get prefixes list",
				httpErr(err, httpResp),
			)
			return
		}

		results := rsp.Results
		if len(results) == 0 {
			break
		}

		for _, prefix := range results {
			var item prefixItemModel

			if prefix.Id == nil || *prefix.Id == "" {
				resp.Diagnostics.AddError(
					"Invalid prefix data",
					"Prefixes list returned an item with no id (prefix: "+prefix.Prefix+")",
				)
				return
			}
			item.ID = types.StringValue(*prefix.Id)

			item.Prefix = types.StringValue(prefix.Prefix)
			item.Description = types.StringValue(derefStr(prefix.Description))
			item.Created = nullableTimeStr(prefix.Created)
			item.LastUpdated = nullableTimeStr(prefix.LastUpdated)

			statusName := ""
			if prefix.Status.Id != nil && prefix.Status.Id.String != nil {
				if statusID := *prefix.Status.Id.String; statusID != "" {
					if name, err := getStatusName(ctx, c, statusID); err == nil {
						statusName = name
					}
				}
			}
			item.Status = types.StringValue(statusName)

			parentID := ""
			if prefix.Parent.IsSet() {
				if parent := prefix.Parent.Get(); parent != nil && parent.Id != nil && parent.Id.String != nil {
					parentID = *parent.Id.String
				}
			}
			item.ParentID = types.StringValue(parentID)
			item.TenantID = nullableFKStr(prefix.Tenant)
			item.RoleID = nullableFKStr(prefix.Role)
			rirID := ""
			if prefix.Rir.IsSet() {
				if rir := prefix.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
					rirID = *rir.Id.String
				}
			}
			item.RirID = types.StringValue(rirID)

			namespaceID := ""
			if prefix.Namespace != nil && prefix.Namespace.Id != nil && prefix.Namespace.Id.String != nil {
				namespaceID = *prefix.Namespace.Id.String
			}
			item.NamespaceID = types.StringValue(namespaceID)

			item.VLANID = nullableFKStr(prefix.Vlan)

			item.Network = types.StringValue(prefix.Network)
			item.Broadcast = types.StringValue(prefix.Broadcast)
			item.PrefixLength = types.Int64Value(int64(prefix.PrefixLength))
			item.IPVersion = types.Int64Value(int64(prefix.IpVersion))

			item.DateAllocated = nullableTimeStr(prefix.DateAllocated)

			if len(prefix.Tags) > 0 {
				tagVals := make([]attr.Value, 0, len(prefix.Tags))
				for _, tag := range prefix.Tags {
					if tag.Id != nil && tag.Id.String != nil {
						tagVals = append(tagVals, types.StringValue(*tag.Id.String))
					}
				}
				item.TagsIDs = types.ListValueMust(types.StringType, tagVals)
			} else {
				item.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
			}

			item.Display = types.StringValue(prefix.Display)
			item.URL = types.StringValue(prefix.Url)
			item.NaturalSlug = types.StringValue(prefix.NaturalSlug)
			item.NotesURL = types.StringValue(prefix.NotesUrl)

			state.Prefixes = append(state.Prefixes, item)
		}

		offset += int32(len(results))

		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read prefixes", map[string]any{"count": len(state.Prefixes)})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
