package provider

import (
	"context"

	"github.com/davispalomino/terraform-provider-vtex/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Check that VtexProvider satisfies provider interfaces
var _ provider.Provider = &VtexProvider{}

// VtexProvider is the provider implementation
type VtexProvider struct {
	version string
}

// VtexProviderModel is the provider data model
type VtexProviderModel struct {
	VtexBaseURL    types.String `tfsdk:"vtex_base_url"`
	OktaURL        types.String `tfsdk:"okta_url"`
	OktaClientID   types.String `tfsdk:"okta_client_id"`
	OktaSecret     types.String `tfsdk:"okta_secret"`
	OktaGrantType  types.String `tfsdk:"okta_grant_type"`
	OktaScope      types.String `tfsdk:"okta_scope"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VtexProvider{
			version: version,
		}
	}
}

func (p *VtexProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vtex"
	resp.Version = p.version
}

func (p *VtexProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider to manage users and roles in VTEX. IMPORTANT: You must install a VTEX Apps Service to use this provider. Without it, the provider will not work.",
		Attributes: map[string]schema.Attribute{
			"vtex_base_url": schema.StringAttribute{
				Description: "VTEX base URL (e.g. https://vendor.myvtex.com)",
				Required:    true,
			},
			"okta_url": schema.StringAttribute{
				Description: "Okta OAuth2 endpoint URL to get tokens",
				Required:    true,
			},
			"okta_client_id": schema.StringAttribute{
				Description: "Okta Client ID (ACCESS_KEY)",
				Required:    true,
				Sensitive:   true,
			},
			"okta_secret": schema.StringAttribute{
				Description: "Okta Client Secret (SECRET_KEY)",
				Required:    true,
				Sensitive:   true,
			},
			"okta_grant_type": schema.StringAttribute{
				Description: "OAuth2 grant type (e.g. authorization_code)",
				Required:    true,
			},
			"okta_scope": schema.StringAttribute{
				Description: "OAuth2 scope (e.g. scope_vendor)",
				Required:    true,
			},
		},
	}
}

func (p *VtexProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config VtexProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create VTEX client
	vtexClient, err := client.NewVtexClient(
		config.VtexBaseURL.ValueString(),
		config.OktaURL.ValueString(),
		config.OktaClientID.ValueString(),
		config.OktaSecret.ValueString(),
		config.OktaGrantType.ValueString(),
		config.OktaScope.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create VTEX client",
			"An unexpected error occurred when creating the VTEX API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	// Make client available for resources
	resp.DataSourceData = vtexClient
	resp.ResourceData = vtexClient
}

func (p *VtexProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVtexUserRoleResource,
	}
}

func (p *VtexProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Add data sources here in the future
	}
}
