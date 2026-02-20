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
	_ resource.Resource                = &mysqlDatabaseUserResource{}
	_ resource.ResourceWithConfigure   = &mysqlDatabaseUserResource{}
	_ resource.ResourceWithImportState = &mysqlDatabaseUserResource{}
)

func NewMySQLDatabaseUserResource() resource.Resource {
	return &mysqlDatabaseUserResource{}
}

type mysqlDatabaseUserResource struct {
	client   *client.Client
	clientID int
	serverID int
}

type mysqlDatabaseUserResourceModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ClientID         types.Int64  `tfsdk:"client_id"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	ServerID         types.Int64  `tfsdk:"server_id"`
}

func (r *mysqlDatabaseUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_database_user"
}

func (r *mysqlDatabaseUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MySQL database user in ISP Config.",
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
				Description: "The MySQL database username.",
				Required:    true,
			},
			"database_password": schema.StringAttribute{
				Description: "The MySQL database password.",
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

func (r *mysqlDatabaseUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mysqlDatabaseUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mysqlDatabaseUserResourceModel
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
			"Error creating MySQL database user",
			"Could not create MySQL database user, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created MySQL database user", map[string]interface{}{"id": dbUserID})
	plan.ID = types.Int64Value(int64(dbUserID))

	createdUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created MySQL database user",
			"Could not read created MySQL database user, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdUser.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *mysqlDatabaseUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mysqlDatabaseUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(state.ID.ValueInt64())

	dbUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MySQL database user",
			fmt.Sprintf("Could not read MySQL database user ID %d: %s", dbUserID, err.Error()),
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

func (r *mysqlDatabaseUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mysqlDatabaseUserResourceModel
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
			"Error updating MySQL database user",
			fmt.Sprintf("Could not update MySQL database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated MySQL database user", map[string]interface{}{"id": dbUserID})

	updatedUser, err := r.client.GetDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated MySQL database user",
			"Could not read updated MySQL database user, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedUser.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *mysqlDatabaseUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mysqlDatabaseUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbUserID := int(state.ID.ValueInt64())

	err := r.client.DeleteDatabaseUser(dbUserID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting MySQL database user",
			fmt.Sprintf("Could not delete MySQL database user ID %d: %s", dbUserID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted MySQL database user", map[string]interface{}{"id": dbUserID})
}

func (r *mysqlDatabaseUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
