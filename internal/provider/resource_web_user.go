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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webUserResource{}
	_ resource.ResourceWithConfigure   = &webUserResource{}
	_ resource.ResourceWithImportState = &webUserResource{}
)

// NewWebUserResource is a helper function to simplify the provider implementation.
func NewWebUserResource() resource.Resource {
	return &webUserResource{}
}

// webUserResource is the resource implementation.
type webUserResource struct {
	client   *client.Client
	clientID int
}

// webUserResourceModel maps the resource schema data.
type webUserResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ClientID       types.Int64  `tfsdk:"client_id"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Dir            types.String `tfsdk:"dir"`
	Shell          types.String `tfsdk:"shell"`
	QuotaSize      types.Int64  `tfsdk:"quota_size"`
	Active         types.String `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
	UID            types.String `tfsdk:"uid"`
	GID            types.String `tfsdk:"gid"`
}

// Metadata returns the resource type name.
func (r *webUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_user"
}

// Schema defines the schema for the resource.
func (r *webUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a shell user in ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the shell user.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.Int64Attribute{
				Description: "The ISP Config client ID.",
				Optional:    true,
			},
			"username": schema.StringAttribute{
				Description: "The shell username.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "The shell user password.",
				Required:    true,
				Sensitive:   true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The parent domain ID.",
				Required:    true,
			},
		"dir": schema.StringAttribute{
			Description: "The shell user directory path.",
			Optional:    true,
			Computed:    true,
		},
		"shell": schema.StringAttribute{
			Description: "The shell for the user (e.g., '/bin/bash', '/bin/sh', '/bin/false', '/sbin/nologin').",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("/bin/bash"),
		},
		"quota_size": schema.Int64Attribute{
			Description: "Quota size in MB.",
			Optional:    true,
			Computed:    true,
		},
			"active": schema.StringAttribute{
				Description: "Whether the shell user is active ('y' or 'n').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("y"),
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Optional:    true,
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

// Configure adds the provider configured client to the resource.
func (r *webUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
}

// Create creates the resource and sets the initial Terraform state.
func (r *webUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine client ID
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

	// Fetch parent domain to get system user/group
	parentDomain, err := r.client.GetWebDomain(int(plan.ParentDomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching parent domain",
			"Could not fetch parent domain to get system user/group: "+err.Error(),
		)
		return
	}

	// Build ShellUser struct
	shellUser := &client.ShellUser{
		Username:       plan.Username.ValueString(),
		Password:       plan.Password.ValueString(),
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
		PUser:          parentDomain.System,      // system_user from parent
		PGroup:         parentDomain.SystemGroup, // system_group from parent
	}

	if !plan.Dir.IsNull() {
		shellUser.Dir = plan.Dir.ValueString()
	}
	if !plan.Shell.IsNull() {
		shellUser.Shell = plan.Shell.ValueString()
	}
	if !plan.QuotaSize.IsNull() {
		shellUser.QuotaSize = client.FlexInt(plan.QuotaSize.ValueInt64())
	}
	if !plan.Active.IsNull() {
		shellUser.Active = plan.Active.ValueString()
	}
	if !plan.ServerID.IsNull() {
		shellUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	}

	// Create shell user
	userID, err := r.client.AddShellUser(shellUser, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating shell user",
			"Could not create shell user, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created shell user", map[string]interface{}{"id": userID})

	plan.ID = types.Int64Value(int64(userID))

	// Read back the created resource to get computed values
	createdUser, err := r.client.GetShellUser(userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created shell user",
			"Could not read created shell user, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with computed values
	if (plan.Dir.IsNull() || plan.Dir.IsUnknown()) && createdUser.Dir != "" {
		plan.Dir = types.StringValue(createdUser.Dir)
	}
	if (plan.Shell.IsNull() || plan.Shell.IsUnknown()) && createdUser.Shell != "" {
		plan.Shell = types.StringValue(createdUser.Shell)
	}
	// Always set computed values from API response when Unknown or Null
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdUser.ServerID))
	}
	if plan.QuotaSize.IsNull() || plan.QuotaSize.IsUnknown() {
		plan.QuotaSize = types.Int64Value(int64(createdUser.QuotaSize))
	}
	if (plan.Active.IsNull() || plan.Active.IsUnknown()) && createdUser.Active != "" {
		plan.Active = types.StringValue(createdUser.Active)
	}
	// UID and GID are computed-only, always set them from API response
	if createdUser.UID != "" {
		plan.UID = types.StringValue(createdUser.UID)
	} else if plan.UID.IsUnknown() || plan.UID.IsNull() {
		plan.UID = types.StringValue("")
	}
	if createdUser.GID != "" {
		plan.GID = types.StringValue(createdUser.GID)
	} else if plan.GID.IsUnknown() || plan.GID.IsNull() {
		plan.GID = types.StringValue("")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := int(state.ID.ValueInt64())

	shellUser, err := r.client.GetShellUser(userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading shell user",
			fmt.Sprintf("Could not read shell user ID %d: %s", userID, err.Error()),
		)
		return
	}

	// Update state
	state.Username = types.StringValue(shellUser.Username)
	// Note: Password is not returned by the API, so we keep the existing value
	state.ParentDomainID = types.Int64Value(int64(shellUser.ParentDomainID))
	state.Dir = types.StringValue(shellUser.Dir)
	state.Shell = types.StringValue(shellUser.Shell)
	if shellUser.QuotaSize != 0 {
		state.QuotaSize = types.Int64Value(int64(shellUser.QuotaSize))
	}
	state.Active = types.StringValue(shellUser.Active)
	if shellUser.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(shellUser.ServerID))
	}
	state.UID = types.StringValue(shellUser.UID)
	state.GID = types.StringValue(shellUser.GID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := int(plan.ID.ValueInt64())

	// Determine client ID
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

	// Fetch parent domain to get system user/group
	parentDomain, err := r.client.GetWebDomain(int(plan.ParentDomainID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error fetching parent domain",
			"Could not fetch parent domain to get system user/group: "+err.Error(),
		)
		return
	}

	// Build ShellUser struct
	shellUser := &client.ShellUser{
		Username:       plan.Username.ValueString(),
		Password:       plan.Password.ValueString(),
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
		PUser:          parentDomain.System,      // system_user from parent
		PGroup:         parentDomain.SystemGroup, // system_group from parent
	}

	if !plan.Dir.IsNull() {
		shellUser.Dir = plan.Dir.ValueString()
	}
	if !plan.Shell.IsNull() {
		shellUser.Shell = plan.Shell.ValueString()
	}
	if !plan.QuotaSize.IsNull() {
		shellUser.QuotaSize = client.FlexInt(plan.QuotaSize.ValueInt64())
	}
	if !plan.Active.IsNull() {
		shellUser.Active = plan.Active.ValueString()
	}
	if !plan.ServerID.IsNull() {
		shellUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	}

	// Update shell user
	err = r.client.UpdateShellUser(userID, clientID, shellUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating shell user",
			fmt.Sprintf("Could not update shell user ID %d: %s", userID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated shell user", map[string]interface{}{"id": userID})

	// Read back the updated resource
	updatedUser, err := r.client.GetShellUser(userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated shell user",
			"Could not read updated shell user, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with computed values
	if (plan.Dir.IsNull() || plan.Dir.IsUnknown()) && updatedUser.Dir != "" {
		plan.Dir = types.StringValue(updatedUser.Dir)
	}
	if (plan.Shell.IsNull() || plan.Shell.IsUnknown()) && updatedUser.Shell != "" {
		plan.Shell = types.StringValue(updatedUser.Shell)
	}
	// Always set computed values from API response when Unknown or Null
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedUser.ServerID))
	}
	if plan.QuotaSize.IsNull() || plan.QuotaSize.IsUnknown() {
		plan.QuotaSize = types.Int64Value(int64(updatedUser.QuotaSize))
	}
	if (plan.Active.IsNull() || plan.Active.IsUnknown()) && updatedUser.Active != "" {
		plan.Active = types.StringValue(updatedUser.Active)
	}
	// UID and GID are computed-only, always set them from API response
	if updatedUser.UID != "" {
		plan.UID = types.StringValue(updatedUser.UID)
	} else if plan.UID.IsUnknown() || plan.UID.IsNull() {
		plan.UID = types.StringValue("")
	}
	if updatedUser.GID != "" {
		plan.GID = types.StringValue(updatedUser.GID)
	} else if plan.GID.IsUnknown() || plan.GID.IsNull() {
		plan.GID = types.StringValue("")
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := int(state.ID.ValueInt64())

	err := r.client.DeleteShellUser(userID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting shell user",
			fmt.Sprintf("Could not delete shell user ID %d: %s", userID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted shell user", map[string]interface{}{"id": userID})
}

// ImportState imports the resource state.
func (r *webUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Convert the import ID (string) to int64
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
