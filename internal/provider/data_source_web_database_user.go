package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &webDatabaseUserDataSource{}
	_ datasource.DataSourceWithConfigure = &webDatabaseUserDataSource{}
)

// NewWebDatabaseUserDataSource is a helper function to simplify the provider implementation.
func NewWebDatabaseUserDataSource() datasource.DataSource {
	return &webDatabaseUserDataSource{}
}

// webDatabaseUserDataSource is the data source implementation.
type webDatabaseUserDataSource struct {
	client *client.Client
}

// webDatabaseUserDataSourceModel maps the data source schema data.
type webDatabaseUserDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	DatabaseUser types.String `tfsdk:"database_user"`
	ServerID     types.Int64  `tfsdk:"server_id"`
}

// Metadata returns the data source type name.
func (d *webDatabaseUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_database_user"
}

// Schema defines the schema for the data source.
func (d *webDatabaseUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a database user from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the database user.",
				Required:    true,
			},
			"database_user": schema.StringAttribute{
				Description: "The database username.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *webDatabaseUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ISPConfigProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ISPConfigProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerData.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *webDatabaseUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webDatabaseUserDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(config.ID.ValueInt64())

	dbUser, err := d.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database user",
			fmt.Sprintf("Could not read database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	// Map response to data source model
	config.DatabaseUser = types.StringValue(dbUser.DatabaseUser)
	if dbUser.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(dbUser.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

