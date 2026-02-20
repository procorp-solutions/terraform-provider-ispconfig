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
	_ datasource.DataSource              = &mysqlDatabaseUserDataSource{}
	_ datasource.DataSourceWithConfigure = &mysqlDatabaseUserDataSource{}
)

func NewMySQLDatabaseUserDataSource() datasource.DataSource {
	return &mysqlDatabaseUserDataSource{}
}

type mysqlDatabaseUserDataSource struct {
	client *client.Client
}

type mysqlDatabaseUserDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	DatabaseUser types.String `tfsdk:"database_user"`
	ServerID     types.Int64  `tfsdk:"server_id"`
}

func (d *mysqlDatabaseUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_database_user"
}

func (d *mysqlDatabaseUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a MySQL database user from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the database user.",
				Required:    true,
			},
			"database_user": schema.StringAttribute{
				Description: "The MySQL database username.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Computed:    true,
			},
		},
	}
}

func (d *mysqlDatabaseUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mysqlDatabaseUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config mysqlDatabaseUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(config.ID.ValueInt64())

	dbUser, err := d.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MySQL database user",
			fmt.Sprintf("Could not read MySQL database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	config.DatabaseUser = types.StringValue(dbUser.DatabaseUser)
	if dbUser.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(dbUser.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
