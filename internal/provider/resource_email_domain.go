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
	_ resource.Resource                = &emailDomainResource{}
	_ resource.ResourceWithConfigure   = &emailDomainResource{}
	_ resource.ResourceWithImportState = &emailDomainResource{}
)

func NewEmailDomainResource() resource.Resource {
	return &emailDomainResource{}
}

type emailDomainResource struct {
	client   *client.Client
	clientID int
	serverID int
}

type emailDomainResourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	ClientID types.Int64  `tfsdk:"client_id"`
	Domain   types.String `tfsdk:"domain"`
	ServerID types.Int64  `tfsdk:"server_id"`
	Active   types.String `tfsdk:"active"`
}

func (r *emailDomainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_domain"
}

func (r *emailDomainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an email domain in ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the email domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.Int64Attribute{
				Description: "The ISP Config client ID.",
				Optional:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The email domain name (e.g. example.com).",
				Required:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The mail server ID.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.StringAttribute{
				Description: "Whether the domain is active. Accepted values: 'y' or 'n'.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *emailDomainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *emailDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailDomainResourceModel
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

	mailDomain := &client.MailDomain{
		Domain: plan.Domain.ValueString(),
	}

	if !plan.ServerID.IsNull() {
		mailDomain.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		mailDomain.ServerID = client.FlexInt(r.serverID)
	}

	if !plan.Active.IsNull() {
		mailDomain.Active = plan.Active.ValueString()
	} else {
		mailDomain.Active = "y"
	}

	mailDomainID, err := r.client.AddMailDomain(mailDomain, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating email domain",
			"Could not create email domain, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created email domain", map[string]interface{}{"id": mailDomainID})
	plan.ID = types.Int64Value(int64(mailDomainID))

	created, err := r.client.GetMailDomain(mailDomainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created email domain",
			"Could not read created email domain, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(created.ServerID))
	}
	if plan.Active.IsNull() || plan.Active.IsUnknown() {
		plan.Active = types.StringValue(created.Active)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state emailDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailDomainID := int(state.ID.ValueInt64())

	mailDomain, err := r.client.GetMailDomain(mailDomainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading email domain",
			fmt.Sprintf("Could not read email domain ID %d: %s", mailDomainID, err.Error()),
		)
		return
	}

	state.Domain = types.StringValue(mailDomain.Domain)
	if mailDomain.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(mailDomain.ServerID))
	}
	if mailDomain.Active != "" {
		state.Active = types.StringValue(mailDomain.Active)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *emailDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emailDomainResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailDomainID := int(plan.ID.ValueInt64())

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

	mailDomain := &client.MailDomain{
		Domain: plan.Domain.ValueString(),
	}

	if !plan.ServerID.IsNull() {
		mailDomain.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	}

	if !plan.Active.IsNull() {
		mailDomain.Active = plan.Active.ValueString()
	}

	err := r.client.UpdateMailDomain(mailDomainID, clientID, mailDomain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating email domain",
			fmt.Sprintf("Could not update email domain ID %d: %s", mailDomainID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated email domain", map[string]interface{}{"id": mailDomainID})

	updated, err := r.client.GetMailDomain(mailDomainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated email domain",
			"Could not read updated email domain, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updated.ServerID))
	}
	if plan.Active.IsNull() || plan.Active.IsUnknown() {
		plan.Active = types.StringValue(updated.Active)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *emailDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state emailDomainResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mailDomainID := int(state.ID.ValueInt64())

	err := r.client.DeleteMailDomain(mailDomainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting email domain",
			fmt.Sprintf("Could not delete email domain ID %d: %s", mailDomainID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted email domain", map[string]interface{}{"id": mailDomainID})
}

func (r *emailDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
