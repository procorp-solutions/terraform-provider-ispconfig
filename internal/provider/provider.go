package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &ISPConfigProvider{}
)

// ISPConfigProvider is the provider implementation.
type ISPConfigProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ISPConfigProviderModel describes the provider data model.
type ISPConfigProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
	ClientID types.Int64  `tfsdk:"client_id"`
	ServerID types.Int64  `tfsdk:"server_id"`
}

// Metadata returns the provider type name.
func (p *ISPConfigProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ispconfig"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *ISPConfigProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with ISP Config API to manage web hosting resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The ISP Config host and port (e.g., 'your-server.com:8080'). " +
					"Can also be set via the ISPCONFIG_HOST environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				Description: "The ISP Config username. " +
					"Can also be set via the ISPCONFIG_USERNAME environment variable.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "The ISP Config password. " +
					"Can also be set via the ISPCONFIG_PASSWORD environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Whether to skip TLS verification. Defaults to false. " +
					"Can also be set via the ISPCONFIG_INSECURE environment variable.",
				Optional: true,
			},
		"client_id": schema.Int64Attribute{
			Description: "The default ISP Config client ID to use for resources. " +
				"Can also be set via the ISPCONFIG_CLIENT_ID environment variable.",
			Optional: true,
		},
		"server_id": schema.Int64Attribute{
			Description: "The default ISP Config server ID to use for resources. " +
				"Can also be set via the ISPCONFIG_SERVER_ID environment variable.",
			Optional: true,
		},
	},
}
}

// Configure prepares a ISP Config API client for data sources and resources.
func (p *ISPConfigProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring ISP Config client")

	// Retrieve provider data from configuration
	var config ISPConfigProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown ISP Config Host",
			"The provider cannot create the ISP Config API client as there is an unknown configuration value for the ISP Config host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ISPCONFIG_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown ISP Config Username",
			"The provider cannot create the ISP Config API client as there is an unknown configuration value for the ISP Config username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ISPCONFIG_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown ISP Config Password",
			"The provider cannot create the ISP Config API client as there is an unknown configuration value for the ISP Config password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ISPCONFIG_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("ISPCONFIG_HOST")
	username := os.Getenv("ISPCONFIG_USERNAME")
	password := os.Getenv("ISPCONFIG_PASSWORD")
	insecure := os.Getenv("ISPCONFIG_INSECURE") == "true"
	clientID := int64(0)
	serverID := int64(0)

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	if !config.ClientID.IsNull() {
		clientID = config.ClientID.ValueInt64()
	}

	// Check environment variable for server_id
	if envServerID := os.Getenv("ISPCONFIG_SERVER_ID"); envServerID != "" {
		if parsed, err := strconv.ParseInt(envServerID, 10, 64); err == nil {
			serverID = parsed
		}
	}

	if !config.ServerID.IsNull() {
		serverID = config.ServerID.ValueInt64()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing ISP Config Host",
			"The provider cannot create the ISP Config API client as there is a missing or empty value for the ISP Config host. "+
				"Set the host value in the configuration or use the ISPCONFIG_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing ISP Config Username",
			"The provider cannot create the ISP Config API client as there is a missing or empty value for the ISP Config username. "+
				"Set the username value in the configuration or use the ISPCONFIG_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing ISP Config Password",
			"The provider cannot create the ISP Config API client as there is a missing or empty value for the ISP Config password. "+
				"Set the password value in the configuration or use the ISPCONFIG_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "ispconfig_host", host)
	ctx = tflog.SetField(ctx, "ispconfig_username", username)
	ctx = tflog.SetField(ctx, "ispconfig_password", "***")
	ctx = tflog.SetField(ctx, "ispconfig_insecure", insecure)
	ctx = tflog.SetField(ctx, "ispconfig_client_id", clientID)
	ctx = tflog.SetField(ctx, "ispconfig_server_id", serverID)

	tflog.Debug(ctx, "Creating ISP Config client")

	// Create a new ISP Config client using the configuration values
	apiClient := client.NewClient(host, username, password, insecure)

	// Login to establish session
	err := apiClient.Login()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Login to ISP Config API",
			"An unexpected error occurred when logging in to the ISP Config API. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"ISP Config Client Error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "ISP Config client configured successfully")

	// Store client, client_id, and server_id in provider data for use in resources and data sources
	providerData := &ISPConfigProviderData{
		Client:   apiClient,
		ClientID: int(clientID),
		ServerID: int(serverID),
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// ISPConfigProviderData contains the shared client for resources and data sources
type ISPConfigProviderData struct {
	Client   *client.Client
	ClientID int
	ServerID int
}

// Resources defines the resources implemented in the provider.
func (p *ISPConfigProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewWebHostingResource,
		NewWebUserResource,
		NewWebDatabaseResource,
		NewWebDatabaseUserResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *ISPConfigProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewWebHostingDataSource,
		NewWebUserDataSource,
		NewWebDatabaseDataSource,
		NewWebDatabaseUserDataSource,
		NewClientDataSource,
	}
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ISPConfigProvider{
			version: version,
		}
	}
}

