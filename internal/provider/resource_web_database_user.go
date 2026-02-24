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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webDatabaseUserResource{}
	_ resource.ResourceWithConfigure   = &webDatabaseUserResource{}
	_ resource.ResourceWithImportState = &webDatabaseUserResource{}
)

// NewWebDatabaseUserResource is a helper function to simplify the provider implementation.
func NewWebDatabaseUserResource() resource.Resource {
	return &webDatabaseUserResource{}
}

// webDatabaseUserResource is the resource implementation.
type webDatabaseUserResource struct {
	client   *client.Client
	clientID int
	serverID int
}

// webDatabaseUserResourceModel maps the resource schema data.
type webDatabaseUserResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ClientID         types.Int64  `tfsdk:"client_id"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	ServerID         types.Int64  `tfsdk:"server_id"`
}

// Metadata returns the resource type name.
func (r *webDatabaseUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_database_user"
}

// Schema defines the schema for the resource.
func (r *webDatabaseUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a database user in ISP Config. Deprecated: use `ispconfig_mysql_database_user` or `ispconfig_pgsql_database_user` instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the database user.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.Int64Attribute{
				Description: "The ISP Config client ID.",
				Optional:    true,
			},
			"database_user": schema.StringAttribute{
				Description: "The database username.",
				Required:    true,
			},
			"database_password": schema.StringAttribute{
				Description: "The database password.",
				Required:    true,
				Sensitive:   true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *webDatabaseUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *webDatabaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webDatabaseUserResourceModel
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

	// Build DatabaseUser struct
	dbUser := &client.DatabaseUser{
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
	}

	if !plan.ServerID.IsNull() {
		dbUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		// Inherit server_id from provider config
		dbUser.ServerID = client.FlexInt(r.serverID)
	}

	// Create database user
	dbUserID, err := r.client.AddDatabaseUser(dbUser, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating database user",
			"Could not create database user, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created database user", map[string]interface{}{"id": dbUserID})

	plan.ID = types.Int64Value(int64(dbUserID))

	// Read back the created resource to get computed values
	createdUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created database user",
			"Could not read created database user, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with computed values - always set when Unknown or Null
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdUser.ServerID))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webDatabaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webDatabaseUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(state.ID.ValueInt64())

	dbUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database user",
			fmt.Sprintf("Could not read database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	// Update state
	state.DatabaseUser = types.StringValue(dbUser.DatabaseUser)
	// Note: Password is not returned by the API, so we keep the existing value
	if dbUser.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(dbUser.ServerID))
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webDatabaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webDatabaseUserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(plan.ID.ValueInt64())

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

	// Build DatabaseUser struct
	dbUser := &client.DatabaseUser{
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
	}

	if !plan.ServerID.IsNull() {
		dbUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		// Inherit server_id from provider config
		dbUser.ServerID = client.FlexInt(r.serverID)
	}

	// Update database user
	err := r.client.UpdateDatabaseUser(dbUserID, clientID, dbUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating database user",
			fmt.Sprintf("Could not update database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated database user", map[string]interface{}{"id": dbUserID})

	// Read back the updated resource
	updatedUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated database user",
			"Could not read updated database user, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with computed values - always set when Unknown or Null
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedUser.ServerID))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webDatabaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webDatabaseUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(state.ID.ValueInt64())

	err := r.client.DeleteDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting database user",
			fmt.Sprintf("Could not delete database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted database user", map[string]interface{}{"id": dbUserID})
}

// webDatabaseUserSourceSchema returns the web_database_user schema for use in MoveState movers.
func webDatabaseUserSourceSchema() *schema.Schema {
	return &schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":                schema.Int64Attribute{Computed: true},
			"client_id":        schema.Int64Attribute{Optional: true},
			"database_user":    schema.StringAttribute{Optional: true, Computed: true},
			"database_password": schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
			"server_id":        schema.Int64Attribute{Optional: true, Computed: true},
		},
	}
}

// ImportState imports the resource state.
func (r *webDatabaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

