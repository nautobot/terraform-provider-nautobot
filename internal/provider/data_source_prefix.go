package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &PrefixDataSource{}
	_ datasource.DataSourceWithConfigure = &PrefixDataSource{}
)

type PrefixDataSource struct {
	client *APIClient
}

type prefixDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	VLANID        types.String `tfsdk:"vlan_id"`
	Prefix        types.String `tfsdk:"prefix"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	ParentID      types.String `tfsdk:"parent_id"`
	RoleID        types.String `tfsdk:"role_id"`
	TenantID      types.String `tfsdk:"tenant_id"`
	RirID         types.String `tfsdk:"rir_id"`
	NamespaceID   types.String `tfsdk:"namespace_id"`
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

func NewPrefixDataSource() datasource.DataSource {
	return &PrefixDataSource{}
}

func (d *PrefixDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefix"
}

func (d *PrefixDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about a Prefix in Nautobot by either its ID or the combination of an exact prefix and namespace UUID.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Description: "The UUID of the prefix. Provide either `id`, or both `prefix` and `namespace_id`.",
				Optional:    true,
				Computed:    true,
			},
			"vlan_id": dsschema.StringAttribute{
				Description: "The UUID of the VLAN associated with the prefix.",
				Computed:    true,
			},

			"prefix": dsschema.StringAttribute{
				Description: "The exact prefix in CIDR notation. Must be provided together with `namespace_id` when `id` is not used.",
				Optional:    true,
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
				Description: "The namespace UUID associated with the prefix. Must be provided together with `prefix` when `id` is not used.",
				Optional:    true,
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
	}
}

func (d *PrefixDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *PrefixDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data prefixDataSourceModel

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

	idStr := data.ID.ValueString()
	prefixStr := data.Prefix.ValueString()
	namespaceIDStr := data.NamespaceID.ValueString()

	idProvided := idStr != ""
	prefixProvided := prefixStr != ""
	namespaceProvided := namespaceIDStr != ""

	if !idProvided && !prefixProvided && !namespaceProvided {
		resp.Diagnostics.AddError(
			"Missing selector",
			"Provide either `id`, or both `prefix` and `namespace_id`.",
		)
		return
	}
	if idProvided && (prefixProvided || namespaceProvided) {
		resp.Diagnostics.AddError(
			"Conflicting selectors",
			"`id` cannot be combined with `prefix` or `namespace_id`. Provide either `id`, or both `prefix` and `namespace_id`.",
		)
		return
	}
	if !idProvided && (!prefixProvided || !namespaceProvided) {
		resp.Diagnostics.AddError(
			"Incomplete prefix selector",
			"`prefix` and `namespace_id` must be provided together.",
		)
		return
	}

	var prefix *nb.Prefix

	if idProvided {
		// Fetch prefix by ID
		rsp, httpResp, err := c.IpamAPI.
			IpamPrefixesRetrieve(ctx, idStr).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get prefix by ID",
				httpErr(err, httpResp),
			)
			return
		}
		prefix = rsp
	} else {
		rsp, httpResp, err := c.IpamAPI.
			IpamPrefixesList(ctx).
			Prefix([]string{prefixStr}).
			Namespace([]string{namespaceIDStr}).
			Execute()
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to get prefix by prefix",
				httpErr(err, httpResp),
			)
			return
		}

		if len(rsp.Results) == 0 {
			resp.Diagnostics.AddError(
				"Prefix not found",
				"No prefix found matching "+prefixStr+" in namespace "+namespaceIDStr,
			)
			return
		}
		if len(rsp.Results) > 1 {
			resp.Diagnostics.AddError(
				"Multiple prefixes found",
				"More than one prefix matched "+prefixStr+" in namespace "+namespaceIDStr+". This violates Nautobot's per-namespace prefix uniqueness constraint.",
			)
			return
		}
		prefix = &rsp.Results[0]
	}

	if prefix == nil {
		resp.Diagnostics.AddError(
			"Prefix not found",
			"Prefix lookup returned no data.",
		)
		return
	}

	if prefix.Id == nil || *prefix.Id == "" {
		resp.Diagnostics.AddError(
			"Invalid prefix data",
			"Prefix returned no id",
		)
		return
	}
	resID := *prefix.Id
	data.ID = types.StringValue(resID)

	data.Prefix = types.StringValue(prefix.Prefix)

	data.Description = types.StringValue(derefStr(prefix.Description))
	data.Created = nullableTimeStr(prefix.Created)
	data.LastUpdated = nullableTimeStr(prefix.LastUpdated)

	statusName := ""
	if prefix.Status.Id != nil && prefix.Status.Id.String != nil {
		if statusID := *prefix.Status.Id.String; statusID != "" {
			if name, err := getStatusName(ctx, c, statusID); err == nil {
				statusName = name
			}
		}
	}
	data.Status = types.StringValue(statusName)

	parentID := ""
	if prefix.Parent.IsSet() {
		if parent := prefix.Parent.Get(); parent != nil && parent.Id != nil && parent.Id.String != nil {
			parentID = *parent.Id.String
		}
	}
	data.ParentID = types.StringValue(parentID)
	data.TenantID = nullableFKStr(prefix.Tenant)
	data.RoleID = nullableFKStr(prefix.Role)
	rirID := ""
	if prefix.Rir.IsSet() {
		if rir := prefix.Rir.Get(); rir != nil && rir.Id != nil && rir.Id.String != nil {
			rirID = *rir.Id.String
		}
	}
	data.RirID = types.StringValue(rirID)

	namespaceID := ""
	if prefix.Namespace != nil && prefix.Namespace.Id != nil && prefix.Namespace.Id.String != nil {
		namespaceID = *prefix.Namespace.Id.String
	}
	data.NamespaceID = types.StringValue(namespaceID)

	data.VLANID = nullableFKStr(prefix.Vlan)

	data.Network = types.StringValue(prefix.Network)
	data.Broadcast = types.StringValue(prefix.Broadcast)
	data.PrefixLength = types.Int64Value(int64(prefix.PrefixLength))
	data.IPVersion = types.Int64Value(int64(prefix.IpVersion))

	data.DateAllocated = nullableTimeStr(prefix.DateAllocated)

	if len(prefix.Tags) > 0 {
		tagVals := make([]attr.Value, 0, len(prefix.Tags))
		for _, tag := range prefix.Tags {
			if tag.Id != nil && tag.Id.String != nil {
				tagVals = append(tagVals, types.StringValue(*tag.Id.String))
			}
		}
		data.TagsIDs = types.ListValueMust(types.StringType, tagVals)
	} else {
		data.TagsIDs = types.ListValueMust(types.StringType, []attr.Value{})
	}

	data.Display = types.StringValue(prefix.Display)
	data.URL = types.StringValue(prefix.Url)
	data.NaturalSlug = types.StringValue(prefix.NaturalSlug)
	data.NotesURL = types.StringValue(prefix.NotesUrl)

	tflog.Debug(ctx, "read Prefix", map[string]any{"id": resID, "prefix": data.Prefix.ValueString(), "namespace_id": data.NamespaceID.ValueString()})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
