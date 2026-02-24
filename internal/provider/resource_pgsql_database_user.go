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
	_ resource.Resource                = &pgsqlDatabaseUserResource{}
	_ resource.ResourceWithConfigure   = &pgsqlDatabaseUserResource{}
	_ resource.ResourceWithImportState = &pgsqlDatabaseUserResource{}
	_ resource.ResourceWithMoveState   = &pgsqlDatabaseUserResource{}
)

func NewPgSQLDatabaseUserResource() resource.Resource {
	return &pgsqlDatabaseUserResource{}
}

type pgsqlDatabaseUserResource struct {
	client   *client.Client
	clientID int
	serverID int
}

type pgsqlDatabaseUserResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ClientID         types.Int64  `tfsdk:"client_id"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	ServerID         types.Int64  `tfsdk:"server_id"`
}

func (r *pgsqlDatabaseUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pgsql_database_user"
}

func (r *pgsqlDatabaseUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PostgreSQL database user in ISP Config.",
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
				Description: "The PostgreSQL database username.",
				Required:    true,
			},
			"database_password": schema.StringAttribute{
				Description: "The PostgreSQL database password.",
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

func (r *pgsqlDatabaseUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pgsqlDatabaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan pgsqlDatabaseUserResourceModel
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

	dbUser := &client.DatabaseUser{
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
	}

	if !plan.ServerID.IsNull() {
		dbUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		dbUser.ServerID = client.FlexInt(r.serverID)
	}

	dbUserID, err := r.client.AddDatabaseUser(dbUser, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating PostgreSQL database user",
			"Could not create PostgreSQL database user, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created PostgreSQL database user", map[string]interface{}{"id": dbUserID})
	plan.ID = types.Int64Value(int64(dbUserID))

	createdUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created PostgreSQL database user",
			"Could not read created PostgreSQL database user, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdUser.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *pgsqlDatabaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state pgsqlDatabaseUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(state.ID.ValueInt64())

	dbUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading PostgreSQL database user",
			fmt.Sprintf("Could not read PostgreSQL database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	state.DatabaseUser = types.StringValue(dbUser.DatabaseUser)
	// Password is not returned by the API; keep the existing state value.
	if dbUser.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(dbUser.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *pgsqlDatabaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan pgsqlDatabaseUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(plan.ID.ValueInt64())

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

	dbUser := &client.DatabaseUser{
		DatabaseUser:     plan.DatabaseUser.ValueString(),
		DatabasePassword: plan.DatabasePassword.ValueString(),
	}

	if !plan.ServerID.IsNull() {
		dbUser.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		dbUser.ServerID = client.FlexInt(r.serverID)
	}

	err := r.client.UpdateDatabaseUser(dbUserID, clientID, dbUser)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating PostgreSQL database user",
			fmt.Sprintf("Could not update PostgreSQL database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated PostgreSQL database user", map[string]interface{}{"id": dbUserID})

	updatedUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated PostgreSQL database user",
			"Could not read updated PostgreSQL database user, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedUser.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *pgsqlDatabaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state pgsqlDatabaseUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(state.ID.ValueInt64())

	err := r.client.DeleteDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting PostgreSQL database user",
			fmt.Sprintf("Could not delete PostgreSQL database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted PostgreSQL database user", map[string]interface{}{"id": dbUserID})
}

func (r *pgsqlDatabaseUserResource) MoveState(_ context.Context) []resource.StateMover {
	return []resource.StateMover{
		{
			SourceSchema: webDatabaseUserSourceSchema(),
			StateMover: func(ctx context.Context, req resource.MoveStateRequest, resp *resource.MoveStateResponse) {
				if req.SourceTypeName != "ispconfig_web_database_user" {
					return
				}

				var src webDatabaseUserResourceModel
				resp.Diagnostics.Append(req.SourceState.Get(ctx, &src)...)
				if resp.Diagnostics.HasError() {
					return
				}

				target := pgsqlDatabaseUserResourceModel{
					ID:               src.ID,
					ClientID:         src.ClientID,
					DatabaseUser:     src.DatabaseUser,
					DatabasePassword: src.DatabasePassword,
					ServerID:         src.ServerID,
				}
				resp.Diagnostics.Append(resp.TargetState.Set(ctx, target)...)
			},
		},
	}
}

func (r *pgsqlDatabaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
