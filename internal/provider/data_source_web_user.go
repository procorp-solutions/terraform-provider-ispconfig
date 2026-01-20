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
	_ datasource.DataSource              = &webUserDataSource{}
	_ datasource.DataSourceWithConfigure = &webUserDataSource{}
)

// NewWebUserDataSource is a helper function to simplify the provider implementation.
func NewWebUserDataSource() datasource.DataSource {
	return &webUserDataSource{}
}

// webUserDataSource is the data source implementation.
type webUserDataSource struct {
	client *client.Client
}

// webUserDataSourceModel maps the data source schema data.
type webUserDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Username       types.String `tfsdk:"username"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Dir            types.String `tfsdk:"dir"`
	QuotaSize      types.Int64  `tfsdk:"quota_size"`
	Active         types.String `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
	UID            types.String `tfsdk:"uid"`
	GID            types.String `tfsdk:"gid"`
}

// Metadata returns the data source type name.
func (d *webUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_user"
}

// Schema defines the schema for the data source.
func (d *webUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a shell user from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the shell user.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "The shell username.",
				Computed:    true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The parent domain ID.",
				Computed:    true,
			},
			"dir": schema.StringAttribute{
				Description: "The shell user directory path.",
				Computed:    true,
			},
			"quota_size": schema.Int64Attribute{
				Description: "Quota size in MB.",
				Computed:    true,
			},
			"active": schema.StringAttribute{
				Description: "Whether the shell user is active.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Computed:    true,
			},
			"uid": schema.StringAttribute{
				Description: "The user ID.",
				Computed:    true,
			},
			"gid": schema.StringAttribute{
				Description: "The group ID.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *webUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *webUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webUserDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := int(config.ID.ValueInt64())

	shellUser, err := d.client.GetShellUser(userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading shell user",
			fmt.Sprintf("Could not read shell user ID %d: %s", userID, err.Error()),
		)
		return
	}

	// Map response to data source model
	config.Username = types.StringValue(shellUser.Username)
	config.ParentDomainID = types.Int64Value(int64(shellUser.ParentDomainID))
	config.Dir = types.StringValue(shellUser.Dir)
	if shellUser.QuotaSize != 0 {
		config.QuotaSize = types.Int64Value(int64(shellUser.QuotaSize))
	} else {
		config.QuotaSize = types.Int64Null()
	}
	config.Active = types.StringValue(shellUser.Active)
	if shellUser.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(shellUser.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}
	config.UID = types.StringValue(shellUser.UID)
	config.GID = types.StringValue(shellUser.GID)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

