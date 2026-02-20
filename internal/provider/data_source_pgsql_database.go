package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

var (
	_ datasource.DataSource              = &pgsqlDatabaseDataSource{}
	_ datasource.DataSourceWithConfigure = &pgsqlDatabaseDataSource{}
)

func NewPgSQLDatabaseDataSource() datasource.DataSource {
	return &pgsqlDatabaseDataSource{}
}

type pgsqlDatabaseDataSource struct {
	client *client.Client
}

type pgsqlDatabaseDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	DatabaseName   types.String `tfsdk:"database_name"`
	DatabaseUserID types.Int64  `tfsdk:"database_user_id"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Quota          types.Int64  `tfsdk:"quota"`
	Active         types.Bool   `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
}

func (d *pgsqlDatabaseDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pgsql_database"
}

func (d *pgsqlDatabaseDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a PostgreSQL database from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the database.",
				Required:    true,
			},
			"database_name": schema.StringAttribute{
				Description: "The PostgreSQL database name.",
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
			"quota": schema.Int64Attribute{
				Description: "Database quota in MB.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the database is active.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Computed:    true,
			},
		},
	}
}

func (d *pgsqlDatabaseDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *pgsqlDatabaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config pgsqlDatabaseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(config.ID.ValueInt64())

	database, err := d.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PostgreSQL database",
			fmt.Sprintf("Could not read PostgreSQL database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	config.DatabaseName = types.StringValue(database.DatabaseName)
	if database.DatabaseUserID != 0 {
		config.DatabaseUserID = types.Int64Value(int64(database.DatabaseUserID))
	} else {
		config.DatabaseUserID = types.Int64Null()
	}
	config.ParentDomainID = types.Int64Value(int64(database.ParentDomainID))
	if database.DatabaseQuota != 0 {
		config.Quota = types.Int64Value(int64(database.DatabaseQuota))
	} else {
		config.Quota = types.Int64Null()
	}
	config.Active = types.BoolValue(webDBYNToBool(database.Active))
	if database.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(database.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
