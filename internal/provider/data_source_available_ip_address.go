package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &AvailableIPAddressDataSource{}
	_ datasource.DataSourceWithConfigure = &AvailableIPAddressDataSource{}
)

type AvailableIPAddressDataSource struct {
	client *APIClient
}

type AvailableIPAddressDataSourceModel struct {
	PrefixID  types.String `tfsdk:"prefix_id"`
	IPVersion types.Int64  `tfsdk:"ip_version"`
	Address   types.String `tfsdk:"address"`
}

func NewAvailableIPAddressDataSource() datasource.DataSource {
	return &AvailableIPAddressDataSource{}
}

func (d *AvailableIPAddressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_ip_address"
}

func (d *AvailableIPAddressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "This data source retrieves an available IP address from a given prefix in Nautobot.",
		Attributes: map[string]dsschema.Attribute{
			"prefix_id": dsschema.StringAttribute{
				Description: "The ID of the prefix from which to retrieve an available IP.",
				Required:    true,
			},
			"ip_version": dsschema.Int64Attribute{
				Description: "The version of the IP address (4 or 6).",
				Computed:    true,
			},
			"address": dsschema.StringAttribute{
				Description: "The available IP address.",
				Computed:    true,
			},
		},
	}
}

func (d *AvailableIPAddressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *AvailableIPAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AvailableIPAddressDataSourceModel

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

	prefixID := data.PrefixID.ValueString()
	if prefixID == "" {
		resp.Diagnostics.AddError(
			"Missing prefix_id",
			"`prefix_id` must be provided.",
		)
		return
	}

	availableIPs, httpResp, err := c.IpamAPI.
		IpamPrefixesAvailableIpsList(ctx, prefixID).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to retrieve available IPs",
			httpErr(err, httpResp),
		)
		return
	}

	if len(availableIPs) == 0 {
		resp.Diagnostics.AddError(
			"No available IPs",
			fmt.Sprintf("No available IP addresses found for prefix %s", prefixID),
		)
		return
	}

	availableIP := availableIPs[0]

	data.IPVersion = types.Int64Value(int64(availableIP.IpVersion))
	data.Address = types.StringValue(availableIP.Address)

	tflog.Debug(ctx, "read available IP", map[string]any{
		"prefix_id":  prefixID,
		"ip_version": availableIP.IpVersion,
		"address":    availableIP.Address,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
