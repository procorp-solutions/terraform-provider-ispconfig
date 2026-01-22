package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &webHostingDataSource{}
	_ datasource.DataSourceWithConfigure = &webHostingDataSource{}
)

// NewWebHostingDataSource is a helper function to simplify the provider implementation.
func NewWebHostingDataSource() datasource.DataSource {
	return &webHostingDataSource{}
}

// webHostingDataSource is the data source implementation.
type webHostingDataSource struct {
	client *client.Client
}

// webHostingDataSourceModel maps the data source schema data.
type webHostingDataSourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Domain         types.String `tfsdk:"domain"`
	IPAddress      types.String `tfsdk:"ip_address"`
	IPv6Address    types.String `tfsdk:"ipv6_address"`
	Type           types.String `tfsdk:"type"`
	ParentDomainID types.Int64  `tfsdk:"parent_domain_id"`
	DocumentRoot   types.String `tfsdk:"document_root"`
	PHP            types.String `tfsdk:"php"`
	Active         types.String `tfsdk:"active"`
	ServerID       types.Int64  `tfsdk:"server_id"`
	HdQuota        types.Int64  `tfsdk:"hd_quota"`
	TrafficQuota   types.Int64  `tfsdk:"traffic_quota"`
	CGI            types.String `tfsdk:"cgi"`
	SSI            types.String `tfsdk:"ssi"`
	Perl           types.String `tfsdk:"perl"`
	Ruby           types.String `tfsdk:"ruby"`
	Python         types.String `tfsdk:"python"`
	SuExec         types.String `tfsdk:"suexec"`
	SSL             types.String `tfsdk:"ssl"`
	RedirectType    types.String `tfsdk:"redirect_type"`
	RedirectPath    types.String `tfsdk:"redirect_path"`
	PHPOpenBasedir          types.String `tfsdk:"php_open_basedir"`
	ApacheDirectives        types.String `tfsdk:"apache_directives"`
	DisableSymlinkNotOwner  types.String `tfsdk:"disable_symlink_restriction"`
}

// Metadata returns the data source type name.
func (d *webHostingDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_hosting"
}

// Schema defines the schema for the data source.
func (d *webHostingDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a web hosting domain from ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the web hosting domain.",
				Required:    true,
			},
			"domain": schema.StringAttribute{
				Description: "The domain name.",
				Computed:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "The IP address for the domain.",
				Computed:    true,
			},
			"ipv6_address": schema.StringAttribute{
				Description: "The IPv6 address for the domain.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of domain.",
				Computed:    true,
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The parent domain ID for subdomains.",
				Computed:    true,
			},
			"document_root": schema.StringAttribute{
				Description: "The document root for the domain.",
				Computed:    true,
			},
			"php": schema.StringAttribute{
				Description: "PHP mode.",
				Computed:    true,
			},
			"active": schema.StringAttribute{
				Description: "Whether the domain is active.",
				Computed:    true,
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID where the domain is hosted.",
				Computed:    true,
			},
			"hd_quota": schema.Int64Attribute{
				Description: "Hard disk quota in MB.",
				Computed:    true,
			},
			"traffic_quota": schema.Int64Attribute{
				Description: "Traffic quota in MB.",
				Computed:    true,
			},
			"cgi": schema.StringAttribute{
				Description: "CGI enabled.",
				Computed:    true,
			},
			"ssi": schema.StringAttribute{
				Description: "SSI enabled.",
				Computed:    true,
			},
			"perl": schema.StringAttribute{
				Description: "Perl enabled.",
				Computed:    true,
			},
			"ruby": schema.StringAttribute{
				Description: "Ruby enabled.",
				Computed:    true,
			},
			"python": schema.StringAttribute{
				Description: "Python enabled.",
				Computed:    true,
			},
			"suexec": schema.StringAttribute{
				Description: "SuExec enabled.",
				Computed:    true,
			},
			"ssl": schema.StringAttribute{
				Description: "SSL enabled.",
				Computed:    true,
			},
			"redirect_type": schema.StringAttribute{
				Description: "The redirect type.",
				Computed:    true,
			},
			"redirect_path": schema.StringAttribute{
				Description: "The redirect path.",
				Computed:    true,
			},
			"php_open_basedir": schema.StringAttribute{
				Description: "PHP open_basedir restriction. Limits which directories PHP can access.",
				Computed:    true,
			},
			"apache_directives": schema.StringAttribute{
				Description: "Custom Apache directives included in the vhost configuration.",
				Computed:    true,
			},
			"disable_symlink_restriction": schema.StringAttribute{
				Description: "Deactivate symlinks restriction of the web space ('y' or 'n').",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *webHostingDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (d *webHostingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config webHostingDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := int(config.ID.ValueInt64())

	domain, err := d.client.GetWebDomain(domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading web hosting",
			fmt.Sprintf("Could not read web hosting ID %d: %s", domainID, err.Error()),
		)
		return
	}

	// Map response to data source model
	config.Domain = types.StringValue(domain.Domain)
	config.IPAddress = types.StringValue(domain.IPAddress)
	config.IPv6Address = types.StringValue(domain.IPv6Address)
	config.Type = types.StringValue(domain.Type)
	if domain.ParentDomainID != 0 {
		config.ParentDomainID = types.Int64Value(int64(domain.ParentDomainID))
	} else {
		config.ParentDomainID = types.Int64Null()
	}
	config.DocumentRoot = types.StringValue(domain.DocumentRoot)
	config.PHP = types.StringValue(domain.PHPVersion)
	config.Active = types.StringValue(domain.Active)
	if domain.ServerID != 0 {
		config.ServerID = types.Int64Value(int64(domain.ServerID))
	} else {
		config.ServerID = types.Int64Null()
	}
	if domain.HdQuota != 0 {
		config.HdQuota = types.Int64Value(int64(domain.HdQuota))
	} else {
		config.HdQuota = types.Int64Null()
	}
	if domain.TrafficQuota != 0 {
		config.TrafficQuota = types.Int64Value(int64(domain.TrafficQuota))
	} else {
		config.TrafficQuota = types.Int64Null()
	}
	config.CGI = types.StringValue(domain.CGI)
	config.SSI = types.StringValue(domain.SSI)
	config.Perl = types.StringValue(domain.Perl)
	config.Ruby = types.StringValue(domain.Ruby)
	config.Python = types.StringValue(domain.Python)
	config.SuExec = types.StringValue(domain.SuExec)
	config.SSL = types.StringValue(domain.SSL)
	config.RedirectType = types.StringValue(domain.RedirectType)
	config.RedirectPath = types.StringValue(domain.RedirectPath)
	config.PHPOpenBasedir = types.StringValue(domain.PHPOpenBasedir)
	config.ApacheDirectives = types.StringValue(domain.ApacheDirectives)
	config.DisableSymlinkNotOwner = types.StringValue(domain.DisableSymlinkNotOwner)

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}

