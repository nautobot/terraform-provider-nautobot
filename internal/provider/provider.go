package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	nb "github.com/nautobot/go-nautobot/v3"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	pSchema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultStatusRequestTimeoutSeconds int64 = 10

type nautobotProvider struct{ version string }

type providerModel struct {
	URL                   types.String `tfsdk:"url"`
	Token                 types.String `tfsdk:"token"`
	SkipVersionCheck      types.Bool   `tfsdk:"skip_version_check"`
	InsecureSkipTLSVerify types.Bool   `tfsdk:"insecure_skip_tls_verify"`
	StatusRequestTimeout  types.Int64  `tfsdk:"status_request_timeout"`
}

type APIClient struct {
	Client *nb.APIClient
	Token  string
}

func New(version string) provider.Provider {
	return &nautobotProvider{version: version}
}

func (p *nautobotProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "nautobot"
	resp.Version = p.version
}

func (p *nautobotProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = pSchema.Schema{
		Attributes: map[string]pSchema.Attribute{
			"url": pSchema.StringAttribute{
				Required:    true,
				Description: "Nautobot API URL",
			},
			"token": pSchema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Admin API token",
			},
			"skip_version_check": pSchema.BoolAttribute{
				Optional:    true,
				Description: "Skip Nautobot version compatibility check. Use with caution.",
			},
			"insecure_skip_tls_verify": pSchema.BoolAttribute{
				Optional: true,
				Description: "Disable TLS certificate verification when connecting to Nautobot. " +
					"Use only for testing.",
			},
			"status_request_timeout": pSchema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds for the Nautobot status request used to verify version compatibility. Defaults to 10 seconds. Set to 0 to disable the timeout.",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
		},
	}
}

func (p *nautobotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.URL.IsNull() || cfg.URL.IsUnknown() || cfg.Token.IsNull() || cfg.Token.IsUnknown() {
		resp.Diagnostics.AddError("Missing configuration", "Both url and token must be set")
		return
	}

	skipVersionCheck := false
	if !cfg.SkipVersionCheck.IsNull() && !cfg.SkipVersionCheck.IsUnknown() {
		skipVersionCheck = cfg.SkipVersionCheck.ValueBool()
	}

	insecureSkipTLS := false
	if !cfg.InsecureSkipTLSVerify.IsNull() && !cfg.InsecureSkipTLSVerify.IsUnknown() {
		insecureSkipTLS = cfg.InsecureSkipTLSVerify.ValueBool()
	}

	statusRequestTimeoutSeconds := defaultStatusRequestTimeoutSeconds
	if !cfg.StatusRequestTimeout.IsNull() && !cfg.StatusRequestTimeout.IsUnknown() {
		statusRequestTimeoutSeconds = cfg.StatusRequestTimeout.ValueInt64()
	}

	conf := nb.NewConfiguration()
	conf.Servers[0].URL = cfg.URL.ValueString()

	baseTransport := http.DefaultTransport
	if insecureSkipTLS {
		if dt, ok := http.DefaultTransport.(*http.Transport); ok {
			tCopy := dt.Clone()
			tCopy.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
			baseTransport = tCopy
		} else {
			baseTransport = &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
		}
	}

	httpClient := &http.Client{
		Transport: &authRT{
			base:  baseTransport,
			token: cfg.Token.ValueString(),
		},
	}

	conf.HTTPClient = httpClient
	api := nb.NewAPIClient(conf)

	if !skipVersionCheck {
		if err := p.checkVersionCompatibility(ctx, api, time.Duration(statusRequestTimeoutSeconds)*time.Second); err != nil {
			resp.Diagnostics.AddError(
				"Failed to verify Nautobot version",
				err.Error(),
			)
			return
		}
	}

	client := &APIClient{Client: api, Token: cfg.Token.ValueString()}
	resp.ResourceData = client
	resp.DataSourceData = client
}

type authRT struct {
	base  http.RoundTripper
	token string
}

func (a *authRT) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Token "+a.token)
	return a.base.RoundTrip(r)
}

func (p *nautobotProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAvailableIPAddressResource,
		NewClusterResource,
		NewClusterTypeResource,
		NewManufacturerResource,
		NewNamespaceResource,
		NewPrefixResource,
		NewVirtualMachineResource,
		NewVMInterfaceResource,
		NewVMPrimaryIPResource,
		NewTenantResource,
		NewTenantGroupResource,
		NewVLANResource,
	}
}

func (p *nautobotProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAvailableIPAddressDataSource,
		NewClusterDataSource,
		NewClustersDataSource,
		NewClusterTypeDataSource,
		NewClusterTypesDataSource,
		NewGraphQLDataSource,
		NewManufacturerDataSource,
		NewManufacturersDataSource,
		NewNamespaceDataSource,
		NewNamespacesDataSource,
		NewPrefixDataSource,
		NewPrefixesDataSource,
		NewVirtualMachineDataSource,
		NewVirtualMachinesDataSource,
		NewTenantDataSource,
		NewTenantsDataSource,
		NewTenantGroupDataSource,
		NewTenantGroupsDataSource,
		NewVLANDataSource,
		NewVLANsDataSource,
	}
}

func (p *nautobotProvider) checkVersionCompatibility(ctx context.Context, api *nb.APIClient, timeout time.Duration) error {
	statusCtx := ctx
	cancel := func() {}
	if timeout > 0 {
		statusCtx, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	status, httpResp, err := api.StatusAPI.StatusRetrieve(statusCtx).Execute()
	if err != nil {
		if timeout > 0 && statusCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("Nautobot /api/status/ request timed out after %s", timeout)
		}
		if httpResp != nil {
			return fmt.Errorf("failed to retrieve Nautobot status: %s (HTTP %s)", err, httpResp.Status)
		}
		return fmt.Errorf("failed to retrieve Nautobot status: %w", err)
	}

	if status == nil || status.NautobotVersion == nil {
		return fmt.Errorf("status endpoint of Nautobot did not return a version")
	}
	nautobotVersion := strings.TrimSpace(*status.NautobotVersion)
	if nautobotVersion == "" {
		return fmt.Errorf("status endpoint of Nautobot did not return a version")
	}

	providerMajMin, err := majorMinorFromVersion(p.version)
	if err != nil {
		return nil
	}
	nautobotMajMin, err := majorMinorFromVersion(nautobotVersion)
	if err != nil {
		return fmt.Errorf("invalid Nautobot version %q: %w", nautobotVersion, err)
	}

	if providerMajMin != nautobotMajMin {
		return fmt.Errorf(
			"provider version %s is only compatible with Nautobot %s.x, but connected instance is %s",
			p.version, providerMajMin, nautobotVersion,
		)
	}

	return nil
}

func majorMinorFromVersion(v string) (string, error) {
	v = strings.TrimSpace(v)
	if strings.HasPrefix(v, "v") || strings.HasPrefix(v, "V") {
		v = v[1:]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("version %q does not have major.minor", v)
	}
	return parts[0] + "." + parts[1], nil
}
