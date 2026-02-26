package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

var (
	_ resource.Resource                = &cronTaskResource{}
	_ resource.ResourceWithConfigure   = &cronTaskResource{}
	_ resource.ResourceWithImportState = &cronTaskResource{}
)

func NewCronTaskResource() resource.Resource {
	return &cronTaskResource{}
}

type cronTaskResource struct {
	client   *client.Client
	clientID int
	serverID int
}

type cronTaskResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ClientID       types.Int64  `tfsdk:"client_id"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Schedule       types.String `tfsdk:"schedule"`
	Command        types.String `tfsdk:"command"`
	Type           types.String `tfsdk:"type"`
	Active         types.Bool   `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
}

func (r *cronTaskResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cron_task"
}

func (r *cronTaskResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a cron task in ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the cron task.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.Int64Attribute{
				Description: "The ISP Config client ID. Overrides the provider-level client_id.",
				Optional:    true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The ID of the parent domain this cron task belongs to.",
				Required:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "The cron schedule in standard format '* * * * *' (min hour mday month wday). Exactly 5 space-separated fields are required.",
				Required:    true,
			},
			"command": schema.StringAttribute{
				Description: "The command, script path, or URL to execute.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The cron job type. One of: url, chrooted, full. Defaults to 'url'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("url"),
			},
			"active": schema.BoolAttribute{
				Description: "Whether the cron task is active. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID. Determined automatically if not set.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *cronTaskResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// parseCronSchedule splits a cron schedule string into its 5 components.
func parseCronSchedule(schedule string) (runMin, runHour, runMday, runMonth, runWday string, err error) {
	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("schedule must have exactly 5 fields (got %d): %q", len(parts), schedule)
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4], nil
}

// buildCronSchedule reconstructs the cron schedule string from API fields.
func buildCronSchedule(runMin, runHour, runMday, runMonth, runWday string) string {
	return strings.Join([]string{runMin, runHour, runMday, runMonth, runWday}, " ")
}

func (r *cronTaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cronTaskResourceModel
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

	runMin, runHour, runMday, runMonth, runWday, err := parseCronSchedule(plan.Schedule.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Schedule", err.Error())
		return
	}

	cronJob := &client.CronJob{
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
		Command:        plan.Command.ValueString(),
		Type:           plan.Type.ValueString(),
		RunMin:         runMin,
		RunHour:        runHour,
		RunMday:        runMday,
		RunMonth:       runMonth,
		RunWday:        runWday,
		Active:         webDBBoolToYN(plan.Active.ValueBool()),
	}

	if !plan.ServerID.IsNull() {
		cronJob.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	} else if r.serverID != 0 {
		cronJob.ServerID = client.FlexInt(r.serverID)
	}

	cronJobID, err := r.client.AddCronJob(cronJob, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cron task",
			"Could not create cron task, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created cron task", map[string]interface{}{"id": cronJobID})
	plan.ID = types.Int64Value(int64(cronJobID))

	created, err := r.client.GetCronJob(cronJobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created cron task",
			"Could not read created cron task, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(created.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *cronTaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cronTaskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cronJobID := int(state.ID.ValueInt64())

	cronJob, err := r.client.GetCronJob(cronJobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cron task",
			fmt.Sprintf("Could not read cron task ID %d: %s", cronJobID, err.Error()),
		)
		return
	}

	state.ParentDomainID = types.Int64Value(int64(cronJob.ParentDomainID))
	state.Schedule = types.StringValue(buildCronSchedule(cronJob.RunMin, cronJob.RunHour, cronJob.RunMday, cronJob.RunMonth, cronJob.RunWday))
	state.Command = types.StringValue(cronJob.Command)
	state.Type = types.StringValue(cronJob.Type)
	state.Active = types.BoolValue(webDBYNToBool(cronJob.Active))
	if cronJob.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(cronJob.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cronTaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cronTaskResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cronJobID := int(plan.ID.ValueInt64())

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

	runMin, runHour, runMday, runMonth, runWday, err := parseCronSchedule(plan.Schedule.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Schedule", err.Error())
		return
	}

	cronJob := &client.CronJob{
		ParentDomainID: client.FlexInt(plan.ParentDomainID.ValueInt64()),
		Command:        plan.Command.ValueString(),
		Type:           plan.Type.ValueString(),
		RunMin:         runMin,
		RunHour:        runHour,
		RunMday:        runMday,
		RunMonth:       runMonth,
		RunWday:        runWday,
		Active:         webDBBoolToYN(plan.Active.ValueBool()),
	}

	if !plan.ServerID.IsNull() {
		cronJob.ServerID = client.FlexInt(plan.ServerID.ValueInt64())
	}

	err = r.client.UpdateCronJob(cronJobID, clientID, cronJob)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cron task",
			fmt.Sprintf("Could not update cron task ID %d: %s", cronJobID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated cron task", map[string]interface{}{"id": cronJobID})

	updated, err := r.client.GetCronJob(cronJobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated cron task",
			"Could not read updated cron task, unexpected error: "+err.Error(),
		)
		return
	}

	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updated.ServerID))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *cronTaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cronTaskResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cronJobID := int(state.ID.ValueInt64())

	err := r.client.DeleteCronJob(cronJobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting cron task",
			fmt.Sprintf("Could not delete cron task ID %d: %s", cronJobID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted cron task", map[string]interface{}{"id": cronJobID})
}

func (r *cronTaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
