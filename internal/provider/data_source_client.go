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
	_ datasource.DataSource              = &clientDataSource{}
	_ datasource.DataSourceWithConfigure = &clientDataSource{}
)

// NewClientDataSource is a helper function to simplify the provider implementation.
func NewClientDataSource() datasource.DataSource {
	return &clientDataSource{}
}

// clientDataSource is the data source implementation.
type clientDataSource struct {
	client *client.Client
}

// clientDataSourceModel maps the data source schema data.
type clientDataSourceModel struct {
	ID                int64        `tfsdk:"id"`
	CompanyName       types.String `tfsdk:"company_name"`
	ContactName       types.String `tfsdk:"contact_name"`
	CustomerNo        types.String `tfsdk:"customer_no"`
	VATNumber         types.String `tfsdk:"vat_number"`
	Street            types.String `tfsdk:"street"`
	Zip               types.String `tfsdk:"zip"`
	City              types.String `tfsdk:"city"`
	State             types.String `tfsdk:"state"`
	Country           types.String `tfsdk:"country"`
	Phone             types.String `tfsdk:"phone"`
	Mobile            types.String `tfsdk:"mobile"`
	Fax               types.String `tfsdk:"fax"`
	Email             types.String `tfsdk:"email"`
	Internet          types.String `tfsdk:"internet"`
	Username          types.String `tfsdk:"username"`
	Locked            types.String `tfsdk:"locked"`
	Canceled          types.String `tfsdk:"canceled"`
	DefaultWebserver  types.Int64  `tfsdk:"default_webserver"`
	DefaultMailserver types.Int64  `tfsdk:"default_mailserver"`
	DefaultDBserver   types.Int64  `tfsdk:"default_dbserver"`
	LimitWeb          types.Int64  `tfsdk:"limit_web"`
	LimitDatabase     types.Int64  `tfsdk:"limit_database"`
	LimitFTPUser      types.Int64  `tfsdk:"limit_ftp_user"`
}

// Metadata returns the data source type name.
func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

// Schema defines the schema for the data source.
func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an ISP Config client (customer) from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the client.",
				Required:    true,
			},
			"company_name": schema.StringAttribute{
				Description: "The company name.",
				Computed:    true,
			},
			"contact_name": schema.StringAttribute{
				Description: "The contact name.",
				Computed:    true,
			},
			"customer_no": schema.StringAttribute{
				Description: "The customer number.",
				Computed:    true,
			},
			"vat_number": schema.StringAttribute{
				Description: "The VAT number.",
				Computed:    true,
			},
			"street": schema.StringAttribute{
				Description: "The street address.",
				Computed:    true,
			},
			"zip": schema.StringAttribute{
				Description: "The ZIP code.",
				Computed:    true,
			},
			"city": schema.StringAttribute{
				Description: "The city.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The state.",
				Computed:    true,
			},
			"country": schema.StringAttribute{
				Description: "The country.",
				Computed:    true,
			},
			"phone": schema.StringAttribute{
				Description: "The phone number.",
				Computed:    true,
			},
			"mobile": schema.StringAttribute{
				Description: "The mobile number.",
				Computed:    true,
			},
			"fax": schema.StringAttribute{
				Description: "The fax number.",
				Computed:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address.",
				Computed:    true,
			},
			"internet": schema.StringAttribute{
				Description: "The internet URL.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username.",
				Computed:    true,
			},
			"locked": schema.StringAttribute{
				Description: "Whether the client is locked.",
				Computed:    true,
			},
			"canceled": schema.StringAttribute{
				Description: "Whether the client is canceled.",
				Computed:    true,
			},
			"default_webserver": schema.Int64Attribute{
				Description: "The default web server ID.",
				Computed:    true,
			},
			"default_mailserver": schema.Int64Attribute{
				Description: "The default mail server ID.",
				Computed:    true,
			},
			"default_dbserver": schema.Int64Attribute{
				Description: "The default database server ID.",
				Computed:    true,
			},
			"limit_web": schema.Int64Attribute{
				Description: "The web domain limit.",
				Computed:    true,
			},
			"limit_database": schema.Int64Attribute{
				Description: "The database limit.",
				Computed:    true,
			},
			"limit_ftp_user": schema.Int64Attribute{
				Description: "The FTP user limit.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *clientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config clientDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := int(config.ID)

	ispClient, err := d.client.GetClient(clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading client",
			fmt.Sprintf("Could not read client ID %d: %s", clientID, err.Error()),
		)
		return
	}

	// Map response to data source model
	config.CompanyName = types.StringValue(ispClient.CompanyName)
	config.ContactName = types.StringValue(ispClient.ContactName)
	config.CustomerNo = types.StringValue(ispClient.CustomerNo)
	config.VATNumber = types.StringValue(ispClient.VATNumber)
	config.Street = types.StringValue(ispClient.Street)
	config.Zip = types.StringValue(ispClient.Zip)
	config.City = types.StringValue(ispClient.City)
	config.State = types.StringValue(ispClient.State)
	config.Country = types.StringValue(ispClient.Country)
	config.Phone = types.StringValue(ispClient.Phone)
	config.Mobile = types.StringValue(ispClient.Mobile)
	config.Fax = types.StringValue(ispClient.Fax)
	config.Email = types.StringValue(ispClient.Email)
	config.Internet = types.StringValue(ispClient.Internet)
	config.Username = types.StringValue(ispClient.Username)
	config.Locked = types.StringValue(ispClient.Locked)
	config.Canceled = types.StringValue(ispClient.Canceled)

	if ispClient.DefaultWebserver != 0 {
		config.DefaultWebserver = types.Int64Value(int64(ispClient.DefaultWebserver))
	} else {
		config.DefaultWebserver = types.Int64Null()
	}
	if ispClient.DefaultMailserver != 0 {
		config.DefaultMailserver = types.Int64Value(int64(ispClient.DefaultMailserver))
	} else {
		config.DefaultMailserver = types.Int64Null()
	}
	if ispClient.DefaultDBserver != 0 {
		config.DefaultDBserver = types.Int64Value(int64(ispClient.DefaultDBserver))
	} else {
		config.DefaultDBserver = types.Int64Null()
	}
	if ispClient.LimitWeb != 0 {
		config.LimitWeb = types.Int64Value(int64(ispClient.LimitWeb))
	} else {
		config.LimitWeb = types.Int64Null()
	}
	if ispClient.LimitDatabase != 0 {
		config.LimitDatabase = types.Int64Value(int64(ispClient.LimitDatabase))
	} else {
		config.LimitDatabase = types.Int64Null()
	}
	if ispClient.LimitFTPUser != 0 {
		config.LimitFTPUser = types.Int64Value(int64(ispClient.LimitFTPUser))
	} else {
		config.LimitFTPUser = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

