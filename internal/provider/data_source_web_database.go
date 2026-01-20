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
	_ datasource.DataSource              = &webDatabaseDataSource{}
	_ datasource.DataSourceWithConfigure = &webDatabaseDataSource{}
)

// NewWebDatabaseDataSource is a helper function to simplify the provider implementation.
func NewWebDatabaseDataSource() datasource.DataSource {
	return &webDatabaseDataSource{}
}

// webDatabaseDataSource is the data source implementation.
type webDatabaseDataSource struct {
	client *client.Client
}

// webDatabaseDataSourceModel maps the data source schema data.
type webDatabaseDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	DatabaseName   types.String `tfsdk:"database_name"`
	DatabaseUserID types.Int64  `tfsdk:"database_user_id"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Type           types.String `tfsdk:"type"`
	Quota          types.Int64  `tfsdk:"quota"`
	Active         types.String `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
	RemoteAccess   types.String `tfsdk:"remote_access"`
	RemoteIPs      types.String `tfsdk:"remote_ips"`
}

// Metadata returns the data source type name.
func (d *webDatabaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_database"
}

// Schema defines the schema for the data source.
func (d *webDatabaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a database from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the database.",
				Required:    true,
			},
			"database_name": schema.StringAttribute{
				Description: "The database name.",
				Computed:    true,
			},
			"database_user_id": schema.Int64Attribute{
				Description: "The database user ID.",
				Computed:    true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The parent domain ID.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The database type.",
				Computed:    true,
			},
			"quota": schema.Int64Attribute{
				Description: "Database quota in MB.",
				Computed:    true,
			},
			"active": schema.StringAttribute{
				Description: "Whether the database is active.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Computed:    true,
			},
			"remote_access": schema.StringAttribute{
				Description: "Remote access enabled.",
				Computed:    true,
			},
			"remote_ips": schema.StringAttribute{
				Description: "Comma-separated list of IPs allowed for remote access.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *webDatabaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *webDatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webDatabaseDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(config.ID.ValueInt64())

	database, err := d.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database",
			fmt.Sprintf("Could not read database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	// Map response to data source model
	config.DatabaseName = types.StringValue(database.DatabaseName)
	if database.DatabaseUserID != 0 {
		config.DatabaseUserID = types.Int64Value(int64(database.DatabaseUserID))
	} else {
		config.DatabaseUserID = types.Int64Null()
	}
	config.ParentDomainID = types.Int64Value(int64(database.ParentDomainID))
	config.Type = types.StringValue(database.Type)
	if database.DatabaseQuota != 0 {
		config.Quota = types.Int64Value(int64(database.DatabaseQuota))
	} else {
		config.Quota = types.Int64Null()
	}
	config.Active = types.StringValue(database.Active)
	if database.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(database.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}
	config.RemoteAccess = types.StringValue(database.RemoteAccess)
	config.RemoteIPs = types.StringValue(database.RemoteIPs)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

