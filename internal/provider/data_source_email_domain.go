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
	_ datasource.DataSource              = &emailDomainDataSource{}
	_ datasource.DataSourceWithConfigure = &emailDomainDataSource{}
)

func NewEmailDomainDataSource() datasource.DataSource {
	return &emailDomainDataSource{}
}

type emailDomainDataSource struct {
	client *client.Client
}

type emailDomainDataSourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Domain   types.String `tfsdk:"domain"`
	ServerID types.Int64  `tfsdk:"server_id"`
	Active   types.String `tfsdk:"active"`
}

func (d *emailDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_domain"
}

func (d *emailDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an email domain from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the email domain.",
				Required:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The email domain name.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The mail server ID.",
				Computed:    true,
			},
			"active": schema.StringAttribute{
				Description: "Whether the domain is active ('y' or 'n').",
				Computed:    true,
			},
		},
	}
}

func (d *emailDomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *emailDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config emailDomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailDomainID := int(config.ID.ValueInt64())

	mailDomain, err := d.client.GetMailDomain(mailDomainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading email domain",
			fmt.Sprintf("Could not read email domain ID %d: %s", mailDomainID, err.Error()),
		)
		return
	}

	config.Domain = types.StringValue(mailDomain.Domain)
	if mailDomain.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(mailDomain.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}
	if mailDomain.Active != "" {
		config.Active = types.StringValue(mailDomain.Active)
	} else {
		config.Active = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
