package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

// Ensure ScaleGridProvider satisfies the provider.Provider interface.
var _ provider.Provider = (*ScaleGridProvider)(nil)

// ScaleGridProvider is the provider implementation.
type ScaleGridProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and run locally, and "test" during acceptance tests.
	version string
}

// ScaleGridProviderModel maps provider schema data to a Go type.
type ScaleGridProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	Email    types.String `tfsdk:"email"`
	APIKey   types.String `tfsdk:"api_key"`
	AuthMode types.String `tfsdk:"auth_mode"`
}

// New returns a function that builds the provider, capturing the version.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ScaleGridProvider{version: version}
	}
}

func (p *ScaleGridProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scalegrid"
	resp.Version = p.version
}

func (p *ScaleGridProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The ScaleGrid provider manages database deployments (MongoDB, Redis, MySQL, " +
			"PostgreSQL) and related resources through the ScaleGrid REST API.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional: true,
				Description: "Base URL of the ScaleGrid API. Defaults to `" + client.DefaultBaseURL +
					"`. May also be set with the `SCALEGRID_BASE_URL` environment variable.",
			},
			"email": schema.StringAttribute{
				Optional: true,
				Description: "ScaleGrid account email, used as the username for basic authentication. " +
					"May also be set with the `SCALEGRID_EMAIL` environment variable.",
			},
			"api_key": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "ScaleGrid API key generated from the console. May also be set with the " +
					"`SCALEGRID_API_KEY` environment variable.",
			},
			"auth_mode": schema.StringAttribute{
				Optional: true,
				Description: "Authentication scheme: `basic` (email + api_key, the default) or " +
					"`bearer` (api_key as a bearer token). May also be set with the " +
					"`SCALEGRID_AUTH_MODE` environment variable.",
			},
		},
	}
}

func (p *ScaleGridProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ScaleGridProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve configuration, letting explicit config win over environment.
	baseURL := firstNonEmpty(stringValue(config.BaseURL), os.Getenv("SCALEGRID_BASE_URL"))
	email := firstNonEmpty(stringValue(config.Email), os.Getenv("SCALEGRID_EMAIL"))
	apiKey := firstNonEmpty(stringValue(config.APIKey), os.Getenv("SCALEGRID_API_KEY"))
	authMode := firstNonEmpty(stringValue(config.AuthMode), os.Getenv("SCALEGRID_AUTH_MODE"), string(client.AuthBasic))

	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing ScaleGrid API key",
			"The provider requires an API key. Set the `api_key` attribute or the "+
				"`SCALEGRID_API_KEY` environment variable.",
		)
	}
	if client.AuthMode(authMode) == client.AuthBasic && email == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("email"),
			"Missing ScaleGrid account email",
			"Basic authentication requires an account email. Set the `email` attribute, the "+
				"`SCALEGRID_EMAIL` environment variable, or switch `auth_mode` to `bearer`.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.NewClient(client.Config{
		BaseURL:   baseURL,
		Email:     email,
		APIKey:    apiKey,
		AuthMode:  client.AuthMode(authMode),
		UserAgent: "terraform-provider-scalegrid/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create ScaleGrid API client", err.Error())
		return
	}

	// Make the client available to resources and data sources.
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *ScaleGridProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
		NewFirewallRuleResource,
		NewCloudProfileResource,
		NewBackupResource,
	}
}

func (p *ScaleGridProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewClusterDataSource,
		NewClustersDataSource,
		NewCloudProfileDataSource,
	}
}
