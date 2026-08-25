package provider

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	nb "github.com/nautobot/go-nautobot/v3"
)

var (
	_ datasource.DataSource              = &GraphQLDataSource{}
	_ datasource.DataSourceWithConfigure = &GraphQLDataSource{}
)

type GraphQLDataSource struct {
	client *APIClient
}

type graphQLDataSourceModel struct {
	Query types.String `tfsdk:"query"`
	Data  types.String `tfsdk:"data"`
}

func NewGraphQLDataSource() datasource.DataSource {
	return &GraphQLDataSource{}
}

func (d *GraphQLDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_graphql"
}

func (d *GraphQLDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Provide an interface to make GraphQL calls to Nautobot as a flexible data source.",
		Attributes: map[string]dsschema.Attribute{
			"query": dsschema.StringAttribute{
				Description: "The GraphQL query that will be sent to Nautobot.",
				Required:    true,
			},
			"data": dsschema.StringAttribute{
				Description: "The data returned by the GraphQL query (JSON-encoded GraphQL `data` field).",
				Computed:    true,
			},
		},
	}
}

func (d *GraphQLDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*APIClient)
}

func (d *GraphQLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data graphQLDataSourceModel

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

	queryStr := data.Query.ValueString()
	tflog.Debug(ctx, "executing GraphQL query", map[string]any{"query": queryStr})

	body := nb.GraphQLAPIRequest{
		Query: queryStr,
	}

	gqlResp, httpResp, err := c.GraphqlAPI.
		GraphqlCreate(ctx).
		GraphQLAPIRequest(body).
		Execute()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to execute GraphQL query",
			httpErr(err, httpResp),
		)
		return
	}

	var raw string
	if gqlResp != nil {
		if gqlResp.Data != nil {
			if b, err := json.Marshal(gqlResp.Data); err == nil {
				raw = string(b)
			} else {
				resp.Diagnostics.AddError(
					"Failed to marshal GraphQL data",
					err.Error(),
				)
				return
			}
		} else {
			raw = "null"
		}
	} else {
		raw = "null"
	}

	data.Data = types.StringValue(raw)

	_ = time.Now()
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
