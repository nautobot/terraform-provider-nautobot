package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &NamespacesDataSource{}
	_ datasource.DataSourceWithConfigure = &NamespacesDataSource{}
)

type NamespacesDataSource struct {
	client *APIClient
}

type namespaceItemModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	LocationID  types.String `tfsdk:"location_id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Created     types.String `tfsdk:"created"`
	LastUpdated types.String `tfsdk:"last_updated"`
	Display     types.String `tfsdk:"display"`
	URL         types.String `tfsdk:"url"`
	NaturalSlug types.String `tfsdk:"natural_slug"`
	NotesURL    types.String `tfsdk:"notes_url"`
}

type namespacesDataSourceModel struct {
	Namespaces []namespaceItemModel `tfsdk:"namespaces"`
}

func NewNamespacesDataSource() datasource.DataSource {
	return &NamespacesDataSource{}
}

func (d *NamespacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespaces"
}

func (d *NamespacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	nestedAttributes := map[string]dsschema.Attribute{
		"id":           dsschema.StringAttribute{Description: "Namespace UUID.", Computed: true},
		"name":         dsschema.StringAttribute{Description: "Namespace name.", Computed: true},
		"description":  dsschema.StringAttribute{Description: "Namespace description.", Computed: true},
		"location_id":  dsschema.StringAttribute{Description: "UUID of the location associated with the namespace.", Computed: true},
		"tenant_id":    dsschema.StringAttribute{Description: "UUID of the tenant associated with the namespace.", Computed: true},
		"created":      dsschema.StringAttribute{Description: "Namespace creation date (RFC3339).", Computed: true},
		"last_updated": dsschema.StringAttribute{Description: "Namespace last update date (RFC3339).", Computed: true},
		"display":      dsschema.StringAttribute{Description: "Human-friendly display value for the namespace.", Computed: true},
		"url":          dsschema.StringAttribute{Description: "API URL of the namespace.", Computed: true},
		"natural_slug": dsschema.StringAttribute{Description: "Natural slug for the namespace.", Computed: true},
		"notes_url":    dsschema.StringAttribute{Description: "Notes URL for the namespace.", Computed: true},
	}

	resp.Schema = dsschema.Schema{
		Description: "Retrieves information about all IPAM namespaces in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"namespaces": dsschema.ListNestedAttribute{
				Description: "List of namespaces.",
				Computed:    true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: nestedAttributes,
				},
			},
		},
	}
}

func (d *NamespacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *NamespacesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state namespacesDataSourceModel
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "API client is not configured. This is a bug in the provider configuration.")
		return
	}

	const pageLimit int32 = 200
	var offset int32
	state.Namespaces = make([]namespaceItemModel, 0)

	for {
		rsp, httpResp, err := d.client.Client.IpamAPI.
			IpamNamespacesList(ctx).
			Limit(pageLimit).
			Offset(offset).
			Sort("name").
			Execute()
		if err != nil {
			resp.Diagnostics.AddError("Failed to get namespaces list", httpErr(err, httpResp))
			return
		}

		if len(rsp.Results) == 0 {
			break
		}
		for _, n := range rsp.Results {
			if n.Id == nil || *n.Id == "" {
				resp.Diagnostics.AddError("Invalid namespace data", "Namespaces list returned an item with no id (name: "+n.Name+")")
				return
			}
			state.Namespaces = append(state.Namespaces, namespaceItemModel{
				ID:          types.StringValue(*n.Id),
				Name:        types.StringValue(n.Name),
				Description: types.StringValue(derefStr(n.Description)),
				LocationID:  nullableFKStr(n.Location),
				TenantID:    nullableFKStr(n.Tenant),
				Created:     nullableTimeStr(n.Created),
				LastUpdated: nullableTimeStr(n.LastUpdated),
				Display:     types.StringValue(n.Display),
				URL:         types.StringValue(n.Url),
				NaturalSlug: types.StringValue(n.NaturalSlug),
				NotesURL:    types.StringValue(n.NotesUrl),
			})
		}

		offset += int32(len(rsp.Results))
		if !rsp.Next.IsSet() || rsp.Next.Get() == nil || *rsp.Next.Get() == "" {
			break
		}
	}

	tflog.Debug(ctx, "read namespaces", map[string]any{"count": len(state.Namespaces)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
