package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

var (
	_ resource.Resource                = &mysqlDatabaseResource{}
	_ resource.ResourceWithConfigure   = &mysqlDatabaseResource{}
	_ resource.ResourceWithImportState = &mysqlDatabaseResource{}
)

func NewMySQLDatabaseResource() resource.Resource {
	return &mysqlDatabaseResource{}
}

type mysqlDatabaseResource struct {
	client   *client.Client
	clientID int
	serverID int
}

type mysqlDatabaseResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ClientID       types.Int64  `tfsdk:"client_id"`
	DatabaseName   types.String `tfsdk:"database_name"`
	DatabaseUserID types.Int64  `tfsdk:"database_user_id"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Quota          types.Int64  `tfsdk:"quota"`
	Active         types.Bool   `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
	RemoteAccess   types.Bool   `tfsdk:"remote_access"`
	RemoteIPs      types.String `tfsdk:"remote_ips"`
}

func (r *mysqlDatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_database"
}

func (r *mysqlDatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a MySQL database in ISP Config.",
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
				Description: "The MySQL database name.",
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
			"quota": schema.Int64Attribute{
				Description: "Database quota in MB.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the database is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Optional:    true,
				Computed:    true,
			},
			"remote_access": schema.BoolAttribute{
				Description: "Enable remote access.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"remote_ips": schema.StringAttribute{
				Description: "Comma-separated list of IPs allowed for remote access.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *mysqlDatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mysqlDatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mysqlDatabaseResourceModel
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

	database := &client.Database{
		DatabaseName:   plan.DatabaseName.ValueString(),
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
		Type:           "mysql",
	}

	if !plan.DatabaseUserID.IsNull() {
		database.DatabaseUserID = client.FlexInt(plan.DatabaseUserID.ValueInt64())
	}
	if !plan.Quota.IsNull() {
		database.DatabaseQuota = client.FlexInt(plan.Quota.ValueInt64())
	}
	if !plan.Active.IsNull() {
		database.Active = webDBBoolToYN(plan.Active.ValueBool())
	}
	if !plan.ServerID.IsNull() {
		database.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else {
		parentDomain, err := r.client.GetWebDomain(int(plan.ParentDomainID.ValueInt64()))
		if err == nil && parentDomain.ServerID != 0 {
			database.ServerID = parentDomain.ServerID
		} else if r.serverID != 0 {
			database.ServerID = client.FlexInt(r.serverID)
		}
	}
	if !plan.RemoteAccess.IsNull() {
		database.RemoteAccess = webDBBoolToYN(plan.RemoteAccess.ValueBool())
	}
	if !plan.RemoteIPs.IsNull() {
		database.RemoteIPs = plan.RemoteIPs.ValueString()
	}

	databaseID, err := r.client.AddDatabase(database, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating MySQL database",
			"Could not create MySQL database, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created MySQL database", map[string]interface{}{"id": databaseID})
	plan.ID = types.Int64Value(int64(databaseID))

	createdDB, err := r.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created MySQL database",
			"Could not read created MySQL database, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdDB.ServerID))
	}
	if plan.DatabaseUserID.IsNull() || plan.DatabaseUserID.IsUnknown() {
		plan.DatabaseUserID = types.Int64Value(int64(createdDB.DatabaseUserID))
	}
	if plan.Quota.IsNull() || plan.Quota.IsUnknown() {
		plan.Quota = types.Int64Value(int64(createdDB.DatabaseQuota))
	}
	if plan.Active.IsNull() || plan.Active.IsUnknown() {
		plan.Active = types.BoolValue(webDBYNToBool(createdDB.Active))
	}
	if plan.RemoteAccess.IsNull() || plan.RemoteAccess.IsUnknown() {
		plan.RemoteAccess = types.BoolValue(webDBYNToBool(createdDB.RemoteAccess))
	}
	if plan.RemoteIPs.IsNull() || plan.RemoteIPs.IsUnknown() {
		plan.RemoteIPs = types.StringValue(createdDB.RemoteIPs)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *mysqlDatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mysqlDatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(state.ID.ValueInt64())

	database, err := r.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading MySQL database",
			fmt.Sprintf("Could not read MySQL database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	state.DatabaseName = types.StringValue(database.DatabaseName)
	state.ParentDomainID = types.Int64Value(int64(database.ParentDomainID))
	if database.DatabaseUserID != 0 {
		state.DatabaseUserID = types.Int64Value(int64(database.DatabaseUserID))
	}
	if database.DatabaseQuota != 0 {
		state.Quota = types.Int64Value(int64(database.DatabaseQuota))
	}
	state.Active = types.BoolValue(webDBYNToBool(database.Active))
	if database.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(database.ServerID))
	}
	state.RemoteAccess = types.BoolValue(webDBYNToBool(database.RemoteAccess))
	state.RemoteIPs = types.StringValue(database.RemoteIPs)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mysqlDatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mysqlDatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(plan.ID.ValueInt64())

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

	database := &client.Database{
		DatabaseName:   plan.DatabaseName.ValueString(),
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
		Type:           "mysql",
	}

	if !plan.DatabaseUserID.IsNull() {
		database.DatabaseUserID = client.FlexInt(plan.DatabaseUserID.ValueInt64())
	}
	if !plan.Quota.IsNull() {
		database.DatabaseQuota = client.FlexInt(plan.Quota.ValueInt64())
	}
	if !plan.Active.IsNull() {
		database.Active = webDBBoolToYN(plan.Active.ValueBool())
	}
	if !plan.ServerID.IsNull() {
		database.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else {
		parentDomain, err := r.client.GetWebDomain(int(plan.ParentDomainID.ValueInt64()))
		if err == nil && parentDomain.ServerID != 0 {
			database.ServerID = parentDomain.ServerID
		} else if r.serverID != 0 {
			database.ServerID = client.FlexInt(r.serverID)
		}
	}
	if !plan.RemoteAccess.IsNull() {
		database.RemoteAccess = webDBBoolToYN(plan.RemoteAccess.ValueBool())
	}
	if !plan.RemoteIPs.IsNull() {
		database.RemoteIPs = plan.RemoteIPs.ValueString()
	}

	err := r.client.UpdateDatabase(databaseID, clientID, database)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating MySQL database",
			fmt.Sprintf("Could not update MySQL database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated MySQL database", map[string]interface{}{"id": databaseID})

	updatedDB, err := r.client.GetDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated MySQL database",
			"Could not read updated MySQL database, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedDB.ServerID))
	}
	if plan.DatabaseUserID.IsNull() || plan.DatabaseUserID.IsUnknown() {
		plan.DatabaseUserID = types.Int64Value(int64(updatedDB.DatabaseUserID))
	}
	if plan.Quota.IsNull() || plan.Quota.IsUnknown() {
		plan.Quota = types.Int64Value(int64(updatedDB.DatabaseQuota))
	}
	if plan.Active.IsNull() || plan.Active.IsUnknown() {
		plan.Active = types.BoolValue(webDBYNToBool(updatedDB.Active))
	}
	if plan.RemoteAccess.IsNull() || plan.RemoteAccess.IsUnknown() {
		plan.RemoteAccess = types.BoolValue(webDBYNToBool(updatedDB.RemoteAccess))
	}
	if plan.RemoteIPs.IsNull() || plan.RemoteIPs.IsUnknown() {
		plan.RemoteIPs = types.StringValue(updatedDB.RemoteIPs)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *mysqlDatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mysqlDatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	databaseID := int(state.ID.ValueInt64())

	err := r.client.DeleteDatabase(databaseID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting MySQL database",
			fmt.Sprintf("Could not delete MySQL database ID %d: %s", databaseID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted MySQL database", map[string]interface{}{"id": databaseID})
}

func (r *mysqlDatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
