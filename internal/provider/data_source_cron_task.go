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
	_ datasource.DataSource              = &cronTaskDataSource{}
	_ datasource.DataSourceWithConfigure = &cronTaskDataSource{}
)

func NewCronTaskDataSource() datasource.DataSource {
	return &cronTaskDataSource{}
}

type cronTaskDataSource struct {
	client *client.Client
}

type cronTaskDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	Schedule       types.String `tfsdk:"schedule"`
	Command        types.String `tfsdk:"command"`
	Type           types.String `tfsdk:"type"`
	Active         types.Bool   `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
}

func (d *cronTaskDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cron_task"
}

func (d *cronTaskDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a cron task from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the cron task.",
				Required:    true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The ID of the parent domain this cron task belongs to.",
				Computed:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "The cron schedule in standard format '* * * * *' (min hour mday month wday).",
				Computed:    true,
			},
			"command": schema.StringAttribute{
				Description: "The URL or command to execute (max 255 characters). A URL for type 'url', or an absolute script/command path for types 'chrooted' or 'full'.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The cron job execution type: 'url' (HTTP/HTTPS URL called via wget), 'chrooted' (script inside chrooted web environment), or 'full' (script with full system access).",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the cron task is active.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID.",
				Computed:    true,
			},
		},
	}
}

func (d *cronTaskDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cronTaskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cronTaskDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cronJobID := int(config.ID.ValueInt64())

	cronJob, err := d.client.GetCronJob(cronJobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cron task",
			fmt.Sprintf("Could not read cron task ID %d: %s", cronJobID, err.Error()),
		)
		return
	}

	config.ParentDomainID = types.Int64Value(int64(cronJob.ParentDomainID))
	config.Schedule = types.StringValue(buildCronSchedule(cronJob.RunMin, cronJob.RunHour, cronJob.RunMday, cronJob.RunMonth, cronJob.RunWday))
	config.Command = types.StringValue(cronJob.Command)
	config.Type = types.StringValue(cronJob.Type)
	config.Active = types.BoolValue(webDBYNToBool(cronJob.Active))
	if cronJob.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(cronJob.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
