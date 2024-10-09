package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"
)

func dataSourceGraphQL() *schema.Resource {
	return &schema.Resource{
		Description: "Provide an interface to make GraphQL calls to Nautobot as a flexible data source.",

		ReadContext: dataSourceGraphQLRead,

		Schema: map[string]*schema.Schema{
			"query": {
				Description: "The GraphQL query that will be sent to Nautobot.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"data": {
				Description: "The data returned by the GraphQL query.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

type reqBody struct {
	Query string `json:"query"`
}

// Use this as reference: https://learn.hashicorp.com/tutorials/terraform/provider-setup?in=terraform/providers#implement-read
func dataSourceGraphQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(*apiClient).Client
	s := fmt.Sprintf("%s/graphql/", meta.(*apiClient).Server)
	t := meta.(*apiClient).Token
	query := d.Get("query").(string)

	queryBody, _ := json.Marshal(reqBody{Query: query})
	req, err := http.NewRequestWithContext(ctx, "POST", s, bytes.NewBuffer(queryBody))
	if err != nil {
		return diag.Errorf("failed to create request with context %s: %s", s, err.Error())
	}
	req.Header.Add("Content-Type", "application/json")
	// Add the authorization header to our request.
	t.Intercept(ctx, req)

	rsp, err := c.GetConfig().HTTPClient.Do(req)
	if err != nil {
		return diag.Errorf("failed to successfully call %s: %s", s, err.Error())
	}
	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return diag.Errorf("failed to decode GraphQL response from %s: %s", s, err.Error())
	}

	data := gjson.Get(string(body), "data")
	if err := d.Set("data", data.Raw); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
