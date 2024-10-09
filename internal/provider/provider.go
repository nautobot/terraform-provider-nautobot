package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	nb "github.com/nautobot/go-nautobot/v2"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"url": {
					Type:     schema.TypeString,
					Required: true,
					DefaultFunc: schema.EnvDefaultFunc(
						"NAUTOBOT_URL",
						nil,
					),
					ValidateFunc: validation.IsURLWithHTTPorHTTPS,
					Description:  "Nautobot API URL",
				},
				"token": {
					Type:      schema.TypeString,
					Required:  true,
					Sensitive: true,
					DefaultFunc: schema.EnvDefaultFunc(
						"NAUTOBOT_TOKEN",
						nil,
					),
					Description: "Admin API token",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"nautobot_available_ip_address": dataSourceAvailableIP(),
				"nautobot_cluster":              dataSourceCluster(),
				"nautobot_clusters":             dataSourceClusters(),
				"nautobot_cluster_type":         dataSourceClusterType(),
				"nautobot_cluster_types":        dataSourceClusterTypes(),
				"nautobot_manufacturer":         dataSourceManufacturer(),
				"nautobot_manufacturers":        dataSourceManufacturers(),
				"nautobot_graphql":              dataSourceGraphQL(),
				"nautobot_prefix":               dataSourcePrefix(),
				"nautobot_prefixes":             dataSourcePrefixes(),
				"nautobot_virtual_machine":      dataSourceVirtualMachine(),
				"nautobot_virtual_machines":     dataSourceVirtualMachines(),
				"nautobot_vlan":                 dataSourceVLAN(),
				"nautobot_vlans":                dataSourceVLANs(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"nautobot_available_ip_address": resourceAvailableIPAddress(),
				"nautobot_cluster":              resourceCluster(),
				"nautobot_cluster_type":         resourceClusterType(),
				"nautobot_manufacturer":         resourceManufacturer(),
				"nautobot_virtual_machine":      resourceVirtualMachine(),
				"nautobot_vm_interface":         resourceVMInterface(),
				"nautobot_vm_primary_ip":        resourcePrimaryIPAddressForVM(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

// Add whatever fields, client or connection info, etc. here
// you would need to setup to communicate with the upstream
// API.
type apiClient struct {
	Client *nb.APIClient
	Server string
	Token  *SecurityProviderNautobotToken
}

func configure(
	version string,
	p *schema.Provider,
) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		serverURL := d.Get("url").(string)
		config := nb.NewConfiguration()
		config.Servers[0].URL = serverURL
		_, hasToken := d.GetOk("token")

		var diags diag.Diagnostics = nil

		if !hasToken {
			diags = diag.FromErr(fmt.Errorf("missing token"))
			diags[0].Severity = diag.Error
			return &apiClient{Server: serverURL}, diags
		}

		token, _ := NewSecurityProviderNautobotToken(
			d.Get("token").(string),
		)

		c := nb.NewAPIClient(config)

		return &apiClient{
			Client: c,
			Server: serverURL,
			Token:  token,
		}, diags
	}
}
