package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

var (
	_ resource.Resource                = &emailInboxResource{}
	_ resource.ResourceWithConfigure   = &emailInboxResource{}
	_ resource.ResourceWithImportState = &emailInboxResource{}
)

func NewEmailInboxResource() resource.Resource {
	return &emailInboxResource{}
}

type emailInboxResource struct {
	client   *client.Client
	clientID int
	serverID int
}

type emailInboxResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	ClientID          types.Int64  `tfsdk:"client_id"`
	MailDomainID      types.Int64  `tfsdk:"maildomain_id"`
	Email             types.String `tfsdk:"email"`
	Password          types.String `tfsdk:"password"`
	Quota             types.Int64  `tfsdk:"quota"`
	ServerID          types.Int64  `tfsdk:"server_id"`
	ForwardIncomingTo types.String `tfsdk:"forward_incoming_to"`
	ForwardOutgoingTo types.String `tfsdk:"forward_outgoing_to"`
}

func (r *emailInboxResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_inbox"
}

func (r *emailInboxResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an email inbox (mailbox) in ISP Config. Inboxes must be assigned to an email domain.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the email inbox.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.Int64Attribute{
				Description: "The ISP Config client ID.",
				Optional:    true,
			},
			"maildomain_id": schema.Int64Attribute{
				Description: "The ID of the email domain this inbox belongs to.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "The full email address (e.g. user@example.com).",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "The mailbox password.",
				Required:    true,
				Sensitive:   true,
			},
			"quota": schema.Int64Attribute{
				Description: "Mailbox quota in MB. Use 0 for no mail allowed, -1 for unlimited.",
				Optional:    true,
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The mail server ID.",
				Optional:    true,
				Computed:    true,
			},
			"forward_incoming_to": schema.StringAttribute{
				Description: "Forward all incoming mail to this email address. Leave empty to disable forwarding.",
				Optional:    true,
				Computed:    true,
			},
			"forward_outgoing_to": schema.StringAttribute{
				Description: "Send a BCC copy of all outgoing mail to this email address. Leave empty to disable.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *emailInboxResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ISPConfigProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ISPConfigProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerData.Client
	r.clientID = providerData.ClientID
	r.serverID = providerData.ServerID
}

func (r *emailInboxResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailInboxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := r.clientID
	if !plan.ClientID.IsNull() {
		clientID = int(plan.ClientID.ValueInt64())
	}
	if clientID == 0 {
		resp.Diagnostics.AddError(
			"Missing Client ID",
			"Client ID must be set either in the provider configuration or in the resource configuration.",
		)
		return
	}

	emailAddr := plan.Email.ValueString()
	mailUser := &client.MailUser{
		MailDomainID: client.FlexInt(plan.MailDomainID.ValueInt64()),
		Email:        emailAddr,
		Login:        emailAddr,
		Password:     plan.Password.ValueString(),
		MoveJunk:     "n",
	}

	if !plan.Quota.IsNull() {
		mailUser.Quota = client.FlexInt(plan.Quota.ValueInt64())
	}

	if !plan.ServerID.IsNull() {
		mailUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		mailUser.ServerID = client.FlexInt(r.serverID)
	}

	if !plan.ForwardIncomingTo.IsNull() {
		mailUser.CC = plan.ForwardIncomingTo.ValueString()
	}

	if !plan.ForwardOutgoingTo.IsNull() {
		mailUser.SenderCC = plan.ForwardOutgoingTo.ValueString()
	}

	mailUserID, err := r.client.AddMailUser(mailUser, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating email inbox",
			"Could not create email inbox, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created email inbox", map[string]interface{}{"id": mailUserID})
	plan.ID = types.Int64Value(int64(mailUserID))

	created, err := r.client.GetMailUser(mailUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created email inbox",
			"Could not read created email inbox, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(created.ServerID))
	}
	if plan.Quota.IsNull() || plan.Quota.IsUnknown() {
		plan.Quota = types.Int64Value(int64(created.Quota))
	}
	if plan.ForwardIncomingTo.IsNull() || plan.ForwardIncomingTo.IsUnknown() {
		plan.ForwardIncomingTo = types.StringValue(created.CC)
	}
	if plan.ForwardOutgoingTo.IsNull() || plan.ForwardOutgoingTo.IsUnknown() {
		plan.ForwardOutgoingTo = types.StringValue(created.SenderCC)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailInboxResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailInboxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailUserID := int(state.ID.ValueInt64())

	mailUser, err := r.client.GetMailUser(mailUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading email inbox",
			fmt.Sprintf("Could not read email inbox ID %d: %s", mailUserID, err.Error()),
		)
		return
	}

	state.Email = types.StringValue(mailUser.Email)
	state.MailDomainID = types.Int64Value(int64(mailUser.MailDomainID))
	// Password is not returned by the API; keep the existing state value.
	if mailUser.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(mailUser.ServerID))
	}
	state.Quota = types.Int64Value(int64(mailUser.Quota))
	state.ForwardIncomingTo = types.StringValue(mailUser.CC)
	state.ForwardOutgoingTo = types.StringValue(mailUser.SenderCC)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *emailInboxResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emailInboxResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailUserID := int(plan.ID.ValueInt64())

	clientID := r.clientID
	if !plan.ClientID.IsNull() {
		clientID = int(plan.ClientID.ValueInt64())
	}
	if clientID == 0 {
		resp.Diagnostics.AddError(
			"Missing Client ID",
			"Client ID must be set either in the provider configuration or in the resource configuration.",
		)
		return
	}

	emailAddr := plan.Email.ValueString()
	mailUser := &client.MailUser{
		MailDomainID: client.FlexInt(plan.MailDomainID.ValueInt64()),
		Email:        emailAddr,
		Login:        emailAddr,
		Password:     plan.Password.ValueString(),
		MoveJunk:     "n",
	}

	if !plan.Quota.IsNull() {
		mailUser.Quota = client.FlexInt(plan.Quota.ValueInt64())
	}

	if !plan.ServerID.IsNull() {
		mailUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		mailUser.ServerID = client.FlexInt(r.serverID)
	}

	if !plan.ForwardIncomingTo.IsNull() {
		mailUser.CC = plan.ForwardIncomingTo.ValueString()
	}

	if !plan.ForwardOutgoingTo.IsNull() {
		mailUser.SenderCC = plan.ForwardOutgoingTo.ValueString()
	}

	err := r.client.UpdateMailUser(mailUserID, clientID, mailUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating email inbox",
			fmt.Sprintf("Could not update email inbox ID %d: %s", mailUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated email inbox", map[string]interface{}{"id": mailUserID})

	updated, err := r.client.GetMailUser(mailUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated email inbox",
			"Could not read updated email inbox, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updated.ServerID))
	}
	if plan.Quota.IsNull() || plan.Quota.IsUnknown() {
		plan.Quota = types.Int64Value(int64(updated.Quota))
	}
	if plan.ForwardIncomingTo.IsNull() || plan.ForwardIncomingTo.IsUnknown() {
		plan.ForwardIncomingTo = types.StringValue(updated.CC)
	}
	if plan.ForwardOutgoingTo.IsNull() || plan.ForwardOutgoingTo.IsUnknown() {
		plan.ForwardOutgoingTo = types.StringValue(updated.SenderCC)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailInboxResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailInboxResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailUserID := int(state.ID.ValueInt64())

	err := r.client.DeleteMailUser(mailUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting email inbox",
			fmt.Sprintf("Could not delete email inbox ID %d: %s", mailUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted email inbox", map[string]interface{}{"id": mailUserID})
}

func (r *emailInboxResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse import ID as integer: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
