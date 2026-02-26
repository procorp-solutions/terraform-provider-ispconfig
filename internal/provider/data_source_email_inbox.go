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
	_ datasource.DataSource              = &emailInboxDataSource{}
	_ datasource.DataSourceWithConfigure = &emailInboxDataSource{}
)

func NewEmailInboxDataSource() datasource.DataSource {
	return &emailInboxDataSource{}
}

type emailInboxDataSource struct {
	client *client.Client
}

type emailInboxDataSourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Email             types.String `tfsdk:"email"`
	MailDomainID      types.Int64  `tfsdk:"maildomain_id"`
	Quota             types.Int64  `tfsdk:"quota"`
	ServerID          types.Int64  `tfsdk:"server_id"`
	ForwardIncomingTo types.String `tfsdk:"forward_incoming_to"`
	ForwardOutgoingTo types.String `tfsdk:"forward_outgoing_to"`
}

func (d *emailInboxDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_inbox"
}

func (d *emailInboxDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an email inbox (mailbox) from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the email inbox.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "The full email address.",
				Computed:    true,
			},
			"maildomain_id": schema.Int64Attribute{
				Description: "The ID of the email domain this inbox belongs to.",
				Computed:    true,
			},
			"quota": schema.Int64Attribute{
				Description: "Mailbox quota in MB.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The mail server ID.",
				Computed:    true,
			},
			"forward_incoming_to": schema.StringAttribute{
				Description: "Address that incoming mail is forwarded to.",
				Computed:    true,
			},
			"forward_outgoing_to": schema.StringAttribute{
				Description: "Address that receives a BCC copy of all outgoing mail.",
				Computed:    true,
			},
		},
	}
}

func (d *emailInboxDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *emailInboxDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config emailInboxDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailUserID := int(config.ID.ValueInt64())

	mailUser, err := d.client.GetMailUser(mailUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading email inbox",
			fmt.Sprintf("Could not read email inbox ID %d: %s", mailUserID, err.Error()),
		)
		return
	}

	config.Email = types.StringValue(mailUser.Email)
	config.MailDomainID = types.Int64Value(int64(mailUser.MailDomainID))
	config.Quota = types.Int64Value(int64(mailUser.Quota))
	if mailUser.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(mailUser.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}
	config.ForwardIncomingTo = types.StringValue(mailUser.CC)
	config.ForwardOutgoingTo = types.StringValue(mailUser.SenderCC)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
