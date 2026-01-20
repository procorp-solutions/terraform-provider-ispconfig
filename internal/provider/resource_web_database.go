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
	_ resource.Resource                = &webDatabaseResource{}
	_ resource.ResourceWithConfigure   = &webDatabaseResource{}
	_ resource.ResourceWithImportState = &webDatabaseResource{}
)

// NewWebDatabaseResource is a helper function to simplify the provider implementation.
func NewWebDatabaseResource() resource.Resource {
	return &webDatabaseResource{}
}

// webDatabaseResource is the resource implementation.
type webDatabaseResource struct {
	client   *client.Client
	clientID int
}

// webDatabaseResourceModel maps the resource schema data.
type webDatabaseResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ClientID       types.Int64  `tfsdk:"client_id"`
	DatabaseName   types.String `tfsdk:"database_name"`
	DatabaseUserID types.Int64  `tfsdk:"database_user_id"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Type           types.String `tfsdk:"type"`
	Quota          types.Int64  `tfsdk:"quota"`
	Active         types.String `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
	RemoteAccess   types.String `tfsdk:"remote_access"`
	RemoteIPs      types.String `tfsdk:"remote_ips"`
}

// Metadata returns the resource type name.
func (r *webDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_database"
}

// Schema defines the schema for the resource.
func (r *webDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a database in ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the database.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.Int64Attribute{
				Description: "The ISP Config client ID.",
				Optional:    true,
			},
			"database_name": schema.StringAttribute{
				Description: "The database name.",
				Required:    true,
			},
			"database_user_id": schema.Int64Attribute{
				Description: "The database user ID.",
				Optional:    true,
				Computed:    true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The parent domain ID.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The database type (e.g., 'mysql', 'postgresql').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("mysql"),
			},
			"quota": schema.Int64Attribute{
				Description: "Database quota in MB.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.StringAttribute{
				Description: "Whether the database is active ('y' or 'n').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("y"),
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Optional:    true,
				Computed:    true,
			},
			"remote_access": schema.StringAttribute{
				Description: "Enable remote access ('y' or 'n').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("n"),
			},
			"remote_ips": schema.StringAttribute{
				Description: "Comma-separated list of IPs allowed for remote access.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *webDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *webDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webDatabaseResourceModel
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

	// Build Database struct
	database := &client.Database{
		DatabaseName:   plan.DatabaseName.ValueString(),
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
	}

	if !plan.DatabaseUserID.IsNull() {
		database.DatabaseUserID = client.FlexInt(plan.DatabaseUserID.ValueInt64())
	}
	if !plan.Type.IsNull() {
		database.Type = plan.Type.ValueString()
	}
	if !plan.Quota.IsNull() {
		database.DatabaseQuota = client.FlexInt(plan.Quota.ValueInt64())
	}
	if !plan.Active.IsNull() {
		database.Active = plan.Active.ValueString()
	}
	if !plan.ServerID.IsNull() {
		database.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	}
	if !plan.RemoteAccess.IsNull() {
		database.RemoteAccess = plan.RemoteAccess.ValueString()
	}
	if !plan.RemoteIPs.IsNull() {
		database.RemoteIPs = plan.RemoteIPs.ValueString()
	}

	// Create database
	databaseID, err := r.client.AddDatabase(database, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating database",
			"Could not create database, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created database", map[string]interface{}{"id": databaseID})

	plan.ID = types.Int64Value(int64(databaseID))

	// Read back the created resource to get computed values
	createdDB, err := r.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created database",
			"Could not read created database, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with computed values - always set when Unknown or Null
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdDB.ServerID))
	}
	if plan.DatabaseUserID.IsNull() || plan.DatabaseUserID.IsUnknown() {
		plan.DatabaseUserID = types.Int64Value(int64(createdDB.DatabaseUserID))
	}
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		plan.Type = types.StringValue(createdDB.Type)
	}
	if plan.Quota.IsNull() || plan.Quota.IsUnknown() {
		plan.Quota = types.Int64Value(int64(createdDB.DatabaseQuota))
	}
	if plan.Active.IsNull() || plan.Active.IsUnknown() {
		plan.Active = types.StringValue(createdDB.Active)
	}
	if plan.RemoteAccess.IsNull() || plan.RemoteAccess.IsUnknown() {
		plan.RemoteAccess = types.StringValue(createdDB.RemoteAccess)
	}
	if plan.RemoteIPs.IsNull() || plan.RemoteIPs.IsUnknown() {
		plan.RemoteIPs = types.StringValue(createdDB.RemoteIPs)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webDatabaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(state.ID.ValueInt64())

	database, err := r.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading database",
			fmt.Sprintf("Could not read database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	// Update state
	state.DatabaseName = types.StringValue(database.DatabaseName)
	state.ParentDomainID = types.Int64Value(int64(database.ParentDomainID))
	if database.DatabaseUserID != 0 {
		state.DatabaseUserID = types.Int64Value(int64(database.DatabaseUserID))
	}
	state.Type = types.StringValue(database.Type)
	if database.DatabaseQuota != 0 {
		state.Quota = types.Int64Value(int64(database.DatabaseQuota))
	}
	state.Active = types.StringValue(database.Active)
	if database.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(database.ServerID))
	}
	state.RemoteAccess = types.StringValue(database.RemoteAccess)
	state.RemoteIPs = types.StringValue(database.RemoteIPs)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webDatabaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(plan.ID.ValueInt64())

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

	// Build Database struct
	database := &client.Database{
		DatabaseName:   plan.DatabaseName.ValueString(),
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
	}

	if !plan.DatabaseUserID.IsNull() {
		database.DatabaseUserID = client.FlexInt(plan.DatabaseUserID.ValueInt64())
	}
	if !plan.Type.IsNull() {
		database.Type = plan.Type.ValueString()
	}
	if !plan.Quota.IsNull() {
		database.DatabaseQuota = client.FlexInt(plan.Quota.ValueInt64())
	}
	if !plan.Active.IsNull() {
		database.Active = plan.Active.ValueString()
	}
	if !plan.ServerID.IsNull() {
		database.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	}
	if !plan.RemoteAccess.IsNull() {
		database.RemoteAccess = plan.RemoteAccess.ValueString()
	}
	if !plan.RemoteIPs.IsNull() {
		database.RemoteIPs = plan.RemoteIPs.ValueString()
	}

	// Update database
	err := r.client.UpdateDatabase(databaseID, clientID, database)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating database",
			fmt.Sprintf("Could not update database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated database", map[string]interface{}{"id": databaseID})

	// Read back the updated resource
	updatedDB, err := r.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated database",
			"Could not read updated database, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with computed values - always set when Unknown or Null
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedDB.ServerID))
	}
	if plan.DatabaseUserID.IsNull() || plan.DatabaseUserID.IsUnknown() {
		plan.DatabaseUserID = types.Int64Value(int64(updatedDB.DatabaseUserID))
	}
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		plan.Type = types.StringValue(updatedDB.Type)
	}
	if plan.Quota.IsNull() || plan.Quota.IsUnknown() {
		plan.Quota = types.Int64Value(int64(updatedDB.DatabaseQuota))
	}
	if plan.Active.IsNull() || plan.Active.IsUnknown() {
		plan.Active = types.StringValue(updatedDB.Active)
	}
	if plan.RemoteAccess.IsNull() || plan.RemoteAccess.IsUnknown() {
		plan.RemoteAccess = types.StringValue(updatedDB.RemoteAccess)
	}
	if plan.RemoteIPs.IsNull() || plan.RemoteIPs.IsUnknown() {
		plan.RemoteIPs = types.StringValue(updatedDB.RemoteIPs)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webDatabaseResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(state.ID.ValueInt64())

	err := r.client.DeleteDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting database",
			fmt.Sprintf("Could not delete database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted database", map[string]interface{}{"id": databaseID})
}

// ImportState imports the resource state.
func (r *webDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

