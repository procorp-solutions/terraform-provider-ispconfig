package provider

import (
	"context"
	"fmt"
	filepath "path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/procorp-solutions/ispconfig-terraform-provider/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webHostingResource{}
	_ resource.ResourceWithConfigure   = &webHostingResource{}
	_ resource.ResourceWithImportState = &webHostingResource{}
)

// NewWebHostingResource is a helper function to simplify the provider implementation.
func NewWebHostingResource() resource.Resource {
	return &webHostingResource{}
}

// webHostingResource is the resource implementation.
type webHostingResource struct {
	client   *client.Client
	clientID int
	serverID int
}

// webHostingResourceModel maps the resource schema data.
type webHostingResourceModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	ClientID             types.Int64  `tfsdk:"client_id"`
	Domain               types.String `tfsdk:"domain"`
	IPAddress            types.String `tfsdk:"ip_address"`
	IPv6Address          types.String `tfsdk:"ipv6_address"`
	Type                 types.String `tfsdk:"type"`
	ParentDomainID       types.Int64  `tfsdk:"parent_domain_id"`
	DocumentRoot         types.String `tfsdk:"document_root"`
	RootSubdir           types.String `tfsdk:"root_subdir"`
	PHP                  types.String `tfsdk:"php"`
	PHPVersion           types.String `tfsdk:"php_version"`
	Active               types.Bool   `tfsdk:"active"`
	ServerID             types.Int64  `tfsdk:"server_id"`
	HdQuota              types.Int64  `tfsdk:"hd_quota"`
	TrafficQuota         types.Int64  `tfsdk:"traffic_quota"`
	CGI                  types.Bool   `tfsdk:"cgi"`
	SSI                  types.Bool   `tfsdk:"ssi"`
	Perl                 types.Bool   `tfsdk:"perl"`
	Ruby                 types.Bool   `tfsdk:"ruby"`
	Python               types.Bool   `tfsdk:"python"`
	SuExec               types.Bool   `tfsdk:"suexec"`
	SSL                  types.Bool   `tfsdk:"ssl"`
	Subdomain            types.String `tfsdk:"subdomain"`
	RedirectType         types.String `tfsdk:"redirect_type"`
	RedirectPath         types.String `tfsdk:"redirect_path"`
	AllowOverride        types.String `tfsdk:"allow_override"`
	PM                   types.String `tfsdk:"pm"`
	PMProcessIdleTimeout types.String `tfsdk:"pm_process_idle_timeout"`
	PMMaxRequests        types.Int64  `tfsdk:"pm_max_requests"`
	HTTPPort             types.Int64  `tfsdk:"http_port"`
	HTTPSPort            types.Int64  `tfsdk:"https_port"`
	PHPOpenBasedir          types.String `tfsdk:"php_open_basedir"`
	ApacheDirectives        types.String `tfsdk:"apache_directives"`
	DisableSymlinkNotOwner  types.Bool   `tfsdk:"disable_symlink_restriction"`
}

// Helper functions for bool to Y/N conversion
func boolToYN(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

func ynToBool(s string) bool {
	return s == "y" || s == "Y"
}

// PHP version to server_php_id mapping
var phpVersionToIDMap = map[string]int{
	"7.0": 2,
	"7.1": 3,
	"7.2": 4,
	"7.3": 5,
	"7.4": 6,
	"8.0": 7,
	"8.1": 8,
	"8.2": 9,
	"8.3": 10,
	"8.4": 11,
}

// Reverse mapping: server_php_id to PHP version
var phpIDToVersionMap = map[int]string{
	2:  "7.0",
	3:  "7.1",
	4:  "7.2",
	5:  "7.3",
	6:  "7.4",
	7:  "8.0",
	8:  "8.1",
	9:  "8.2",
	10: "8.3",
	11: "8.4",
}

// phpVersionToID converts PHP version string to server_php_id
func phpVersionToID(version string) (int, error) {
	id, ok := phpVersionToIDMap[version]
	if !ok {
		return 0, fmt.Errorf("invalid PHP version: %s. Valid versions are: 7.0, 7.1, 7.2, 7.3, 7.4, 8.0, 8.1, 8.2, 8.3, 8.4", version)
	}
	return id, nil
}

// phpIDToVersion converts server_php_id to PHP version string
func phpIDToVersion(id int) string {
	version, ok := phpIDToVersionMap[id]
	if !ok {
		return "" // Return empty string if ID not found
	}
	return version
}

// combineDocumentRoot combines a base document root path with a subdirectory
func combineDocumentRoot(basePath, subdir string) string {
	if subdir == "" {
		return basePath
	}
	// Clean the subdir to remove leading/trailing slashes
	subdir = strings.Trim(subdir, "/")
	// Use filepath.Join for proper path combination
	return filepath.Join(basePath, subdir)
}

// Metadata returns the resource type name.
func (r *webHostingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_web_hosting"
}

// Schema defines the schema for the resource.
func (r *webHostingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a web hosting domain in ISP Config.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the web hosting domain.",
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
				Description: "The domain name.",
				Required:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "The IP address for the domain.",
				Optional:    true,
				Computed:    true,
			},
			"ipv6_address": schema.StringAttribute{
				Description: "The IPv6 address for the domain.",
				Optional:    true,
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The type of domain (e.g., 'vhost', 'subdomain').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("vhost"),
			},
			"parent_domain_id": schema.Int64Attribute{
				Description: "The parent domain ID for subdomains.",
				Optional:    true,
				Computed:    true,
			},
		"document_root": schema.StringAttribute{
			Description: "The document root for the domain.",
			Optional:    true,
			Computed:    true,
		},
		"root_subdir": schema.StringAttribute{
			Description: "Subdirectory path to append to the ISPConfig-generated base document root (e.g., 'web/www'). Cannot be used with document_root.",
			Optional:    true,
		},
		"php": schema.StringAttribute{
				Description: "PHP mode (e.g., 'php-fpm', 'fast-cgi', 'mod', 'no').",
				Optional:    true,
				Computed:    true,
			},
			"php_version": schema.StringAttribute{
				Description: "PHP version: 7.0, 7.1, 7.2, 7.3, 7.4, 8.0, 8.1, 8.2, 8.3, or 8.4.",
				Optional:    true,
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the domain is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"server_id": schema.Int64Attribute{
				Description: "The server ID where the domain is hosted. Can be set in provider configuration or here. Required by ISPConfig (typically 1 for single-server setups).",
				Optional:    true,
				Computed:    true,
			},
			"hd_quota": schema.Int64Attribute{
				Description: "Hard disk quota in MB.",
				Optional:    true,
				Computed:    true,
			},
			"traffic_quota": schema.Int64Attribute{
				Description: "Traffic quota in MB.",
				Optional:    true,
				Computed:    true,
			},
			"cgi": schema.BoolAttribute{
				Description: "Enable CGI.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ssi": schema.BoolAttribute{
				Description: "Enable SSI.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"perl": schema.BoolAttribute{
				Description: "Enable Perl.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"ruby": schema.BoolAttribute{
				Description: "Enable Ruby.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"python": schema.BoolAttribute{
				Description: "Enable Python.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"suexec": schema.BoolAttribute{
				Description: "Enable SuExec.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"ssl": schema.BoolAttribute{
				Description: "Enable SSL.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"subdomain": schema.StringAttribute{
				Description: "Subdomain auto-redirect setting (e.g., 'www', 'none', '*'). Default 'www' creates www subdomain alias.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("www"),
			},
			"redirect_type": schema.StringAttribute{
				Description: "The redirect type (e.g., '', 'R', 'L', 'R=301', 'R=302').",
				Optional:    true,
				Computed:    true,
			},
			"redirect_path": schema.StringAttribute{
				Description: "The redirect path.",
				Optional:    true,
				Computed:    true,
			},
			"allow_override": schema.StringAttribute{
				Description: "Apache AllowOverride directive (e.g., 'All', 'None').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("All"),
			},
			"pm": schema.StringAttribute{
				Description: "PHP-FPM process manager type: 'dynamic', 'static', 'ondemand'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ondemand"),
			},
			"pm_process_idle_timeout": schema.StringAttribute{
				Description: "PHP-FPM process idle timeout in seconds.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("10"),
			},
			"pm_max_requests": schema.Int64Attribute{
				Description: "PHP-FPM max requests per process. Leave unset to use ISPConfig default.",
				Optional:    true,
				Computed:    true,
			},
			"http_port": schema.Int64Attribute{
				Description: "HTTP port number.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(80),
			},
			"https_port": schema.Int64Attribute{
				Description: "HTTPS port number.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(443),
			},
			"php_open_basedir": schema.StringAttribute{
				Description: "PHP open_basedir restriction. Limits which directories PHP can access.",
				Optional:    true,
				Computed:    true,
			},
			"apache_directives": schema.StringAttribute{
				Description: "Custom Apache directives to include in the vhost configuration.",
				Optional:    true,
				Computed:    true,
			},
			"disable_symlink_restriction": schema.BoolAttribute{
				Description: "Deactivate symlinks restriction of the web space. When true, allows following symlinks regardless of owner.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *webHostingResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *webHostingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webHostingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that document_root and root_subdir are not both set in config
	if !plan.DocumentRoot.IsNull() && !plan.RootSubdir.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot specify both document_root and root_subdir in configuration. Use root_subdir to append to the ISPConfig-generated base path, or document_root to specify the full path.",
		)
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

	// Build WebDomain struct
	domain := &client.WebDomain{
		Domain:   plan.Domain.ValueString(),
		ClientID: client.FlexInt(clientID),
	}

	if !plan.IPAddress.IsNull() {
		domain.IPAddress = plan.IPAddress.ValueString()
	}
	if !plan.IPv6Address.IsNull() {
		domain.IPv6Address = plan.IPv6Address.ValueString()
	}
	if !plan.Type.IsNull() {
		domain.Type = plan.Type.ValueString()
	}
	if !plan.ParentDomainID.IsNull() {
		domain.ParentDomainID = client.FlexInt(plan.ParentDomainID.ValueInt64())
	}
	// Only set document_root if root_subdir is not specified
	// If root_subdir is set, we'll let ISPConfig generate the base path and append the subdir after creation
	if !plan.DocumentRoot.IsNull() && plan.RootSubdir.IsNull() {
		domain.DocumentRoot = plan.DocumentRoot.ValueString()
	}
	if !plan.PHP.IsNull() {
		domain.PHPVersion = plan.PHP.ValueString()
	}
	if !plan.PHPVersion.IsNull() {
		phpID, err := phpVersionToID(plan.PHPVersion.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid PHP Version",
				err.Error(),
			)
			return
		}
		domain.ServerPHPID = client.FlexInt(phpID)
	}
	if !plan.Active.IsNull() {
		domain.Active = boolToYN(plan.Active.ValueBool())
	}
	// ServerID: use resource value if set, otherwise use provider default
	serverID := r.serverID
	if !plan.ServerID.IsNull() {
		serverID = int(plan.ServerID.ValueInt64())
	}
	if serverID == 0 {
		resp.Diagnostics.AddError(
			"Missing Server ID",
			"Server ID must be set either in the provider configuration or in the resource configuration.",
		)
		return
	}
	domain.ServerID = client.FlexInt(serverID)
	if !plan.HdQuota.IsNull() {
		domain.HdQuota = client.FlexInt(plan.HdQuota.ValueInt64())
	}
	if !plan.TrafficQuota.IsNull() {
		domain.TrafficQuota = client.FlexInt(plan.TrafficQuota.ValueInt64())
	}
	if !plan.CGI.IsNull() {
		domain.CGI = boolToYN(plan.CGI.ValueBool())
	}
	if !plan.SSI.IsNull() {
		domain.SSI = boolToYN(plan.SSI.ValueBool())
	}
	if !plan.Perl.IsNull() {
		domain.Perl = boolToYN(plan.Perl.ValueBool())
	}
	if !plan.Ruby.IsNull() {
		domain.Ruby = boolToYN(plan.Ruby.ValueBool())
	}
	if !plan.Python.IsNull() {
		domain.Python = boolToYN(plan.Python.ValueBool())
	}
	if !plan.SuExec.IsNull() {
		domain.SuExec = boolToYN(plan.SuExec.ValueBool())
	}
	if !plan.SSL.IsNull() {
		domain.SSL = boolToYN(plan.SSL.ValueBool())
	}
	if !plan.Subdomain.IsNull() {
		domain.Subdomain = plan.Subdomain.ValueString()
	}
	if !plan.RedirectType.IsNull() {
		domain.RedirectType = plan.RedirectType.ValueString()
	}
	if !plan.RedirectPath.IsNull() {
		domain.RedirectPath = plan.RedirectPath.ValueString()
	}
	if !plan.AllowOverride.IsNull() {
		domain.AllowOverride = plan.AllowOverride.ValueString()
	}
	if !plan.PM.IsNull() {
		domain.PM = plan.PM.ValueString()
	}
	if !plan.PMProcessIdleTimeout.IsNull() {
		domain.PMProcess = plan.PMProcessIdleTimeout.ValueString()
	}
	if !plan.PMMaxRequests.IsNull() {
		domain.PMMaxRequests = client.FlexInt(plan.PMMaxRequests.ValueInt64())
	}
	if !plan.HTTPPort.IsNull() {
		domain.HTTPPort = client.FlexInt(plan.HTTPPort.ValueInt64())
	}
	if !plan.HTTPSPort.IsNull() {
		domain.HTTPSPort = client.FlexInt(plan.HTTPSPort.ValueInt64())
	}
	if !plan.PHPOpenBasedir.IsNull() {
		domain.PHPOpenBasedir = plan.PHPOpenBasedir.ValueString()
	}
	if !plan.ApacheDirectives.IsNull() {
		domain.ApacheDirectives = plan.ApacheDirectives.ValueString()
	}
	// Always send disable_symlink_restriction (defaults to false/"n")
	domain.DisableSymlinkNotOwner = boolToYN(plan.DisableSymlinkNotOwner.ValueBool())

	// Create web domain
	domainID, err := r.client.AddWebDomain(domain, clientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating web hosting",
			"Could not create web hosting, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Created web hosting", map[string]interface{}{"id": domainID})

	plan.ID = types.Int64Value(int64(domainID))

	// Read back the created resource to get computed values
	createdDomain, err := r.client.GetWebDomain(domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created web hosting",
			"Could not read created web hosting, unexpected error: "+err.Error(),
		)
		return
	}

	// If root_subdir is specified, update the document_root with the combined path
	if !plan.RootSubdir.IsNull() && plan.RootSubdir.ValueString() != "" {
		baseDocRoot := createdDomain.DocumentRoot
		newDocRoot := combineDocumentRoot(baseDocRoot, plan.RootSubdir.ValueString())
		
		tflog.Debug(ctx, "Updating document root with subdir", map[string]interface{}{
			"base_path": baseDocRoot,
			"subdir":    plan.RootSubdir.ValueString(),
			"new_path":  newDocRoot,
		})

		// Update the domain with the new document root
		updateDomain := &client.WebDomain{
			Domain:       createdDomain.Domain,
			ClientID:     createdDomain.ClientID,
			DocumentRoot: newDocRoot,
		}
		
		err = r.client.UpdateWebDomain(domainID, clientID, updateDomain)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating document root with subdir",
				fmt.Sprintf("Could not update document root: %s", err.Error()),
			)
			return
		}

		// Read back again to get the updated document root
		createdDomain, err = r.client.GetWebDomain(domainID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading updated web hosting",
				"Could not read updated web hosting, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Update plan with computed values - always set when Unknown or Null
	if plan.IPAddress.IsNull() || plan.IPAddress.IsUnknown() {
		plan.IPAddress = types.StringValue(createdDomain.IPAddress)
	}
	if plan.IPv6Address.IsNull() || plan.IPv6Address.IsUnknown() {
		plan.IPv6Address = types.StringValue(createdDomain.IPv6Address)
	}
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(createdDomain.ServerID))
	}
	if plan.DocumentRoot.IsNull() || plan.DocumentRoot.IsUnknown() {
		plan.DocumentRoot = types.StringValue(createdDomain.DocumentRoot)
	}
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		plan.Type = types.StringValue(createdDomain.Type)
	}
	if plan.PHP.IsNull() || plan.PHP.IsUnknown() {
		plan.PHP = types.StringValue(createdDomain.PHPVersion)
	}
	if plan.PHPVersion.IsNull() || plan.PHPVersion.IsUnknown() {
		phpVersion := phpIDToVersion(int(createdDomain.ServerPHPID))
		plan.PHPVersion = types.StringValue(phpVersion)
	}
	if plan.ParentDomainID.IsNull() || plan.ParentDomainID.IsUnknown() {
		plan.ParentDomainID = types.Int64Value(int64(createdDomain.ParentDomainID))
	}
	if plan.HdQuota.IsNull() || plan.HdQuota.IsUnknown() {
		plan.HdQuota = types.Int64Value(int64(createdDomain.HdQuota))
	}
	if plan.TrafficQuota.IsNull() || plan.TrafficQuota.IsUnknown() {
		plan.TrafficQuota = types.Int64Value(int64(createdDomain.TrafficQuota))
	}
	if plan.RedirectType.IsNull() || plan.RedirectType.IsUnknown() {
		plan.RedirectType = types.StringValue(createdDomain.RedirectType)
	}
	if plan.RedirectPath.IsNull() || plan.RedirectPath.IsUnknown() {
		plan.RedirectPath = types.StringValue(createdDomain.RedirectPath)
	}
	if plan.PMMaxRequests.IsNull() || plan.PMMaxRequests.IsUnknown() {
		plan.PMMaxRequests = types.Int64Value(int64(createdDomain.PMMaxRequests))
	}
	if plan.PHPOpenBasedir.IsNull() || plan.PHPOpenBasedir.IsUnknown() {
		plan.PHPOpenBasedir = types.StringValue(createdDomain.PHPOpenBasedir)
	}
	if plan.ApacheDirectives.IsNull() || plan.ApacheDirectives.IsUnknown() {
		plan.ApacheDirectives = types.StringValue(createdDomain.ApacheDirectives)
	}
	if plan.DisableSymlinkNotOwner.IsNull() || plan.DisableSymlinkNotOwner.IsUnknown() {
		plan.DisableSymlinkNotOwner = types.BoolValue(ynToBool(createdDomain.DisableSymlinkNotOwner))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webHostingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webHostingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := int(state.ID.ValueInt64())

	domain, err := r.client.GetWebDomain(domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading web hosting",
			fmt.Sprintf("Could not read web hosting ID %d: %s", domainID, err.Error()),
		)
		return
	}

	// Update state
	state.Domain = types.StringValue(domain.Domain)
	state.IPAddress = types.StringValue(domain.IPAddress)
	state.IPv6Address = types.StringValue(domain.IPv6Address)
	state.Type = types.StringValue(domain.Type)
	if domain.ParentDomainID != 0 {
		state.ParentDomainID = types.Int64Value(int64(domain.ParentDomainID))
	}
	state.DocumentRoot = types.StringValue(domain.DocumentRoot)
	// Note: root_subdir is preserved from state as-is since it's configuration-only
	// and not returned by the API
	state.PHP = types.StringValue(domain.PHPVersion)
	if domain.ServerPHPID != 0 {
		phpVersion := phpIDToVersion(int(domain.ServerPHPID))
		if phpVersion != "" {
			state.PHPVersion = types.StringValue(phpVersion)
		}
	}
	state.Active = types.BoolValue(ynToBool(domain.Active))
	if domain.ServerID != 0 {
		state.ServerID = types.Int64Value(int64(domain.ServerID))
	}
	if domain.HdQuota != 0 {
		state.HdQuota = types.Int64Value(int64(domain.HdQuota))
	}
	if domain.TrafficQuota != 0 {
		state.TrafficQuota = types.Int64Value(int64(domain.TrafficQuota))
	}
	state.CGI = types.BoolValue(ynToBool(domain.CGI))
	state.SSI = types.BoolValue(ynToBool(domain.SSI))
	state.Perl = types.BoolValue(ynToBool(domain.Perl))
	state.Ruby = types.BoolValue(ynToBool(domain.Ruby))
	state.Python = types.BoolValue(ynToBool(domain.Python))
	state.SuExec = types.BoolValue(ynToBool(domain.SuExec))
	state.SSL = types.BoolValue(ynToBool(domain.SSL))
	state.Subdomain = types.StringValue(domain.Subdomain)
	state.RedirectType = types.StringValue(domain.RedirectType)
	state.RedirectPath = types.StringValue(domain.RedirectPath)
	state.AllowOverride = types.StringValue(domain.AllowOverride)
	state.PM = types.StringValue(domain.PM)
	state.PMProcessIdleTimeout = types.StringValue(domain.PMProcess)
	if domain.PMMaxRequests != 0 {
		state.PMMaxRequests = types.Int64Value(int64(domain.PMMaxRequests))
	}
	if domain.HTTPPort != 0 {
		state.HTTPPort = types.Int64Value(int64(domain.HTTPPort))
	}
	if domain.HTTPSPort != 0 {
		state.HTTPSPort = types.Int64Value(int64(domain.HTTPSPort))
	}
	state.PHPOpenBasedir = types.StringValue(domain.PHPOpenBasedir)
	state.ApacheDirectives = types.StringValue(domain.ApacheDirectives)
	state.DisableSymlinkNotOwner = types.BoolValue(ynToBool(domain.DisableSymlinkNotOwner))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webHostingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webHostingResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that document_root and root_subdir are not both explicitly set in config
	// We check the config directly to avoid false positives from computed state values
	var config webHostingResourceModel
	diags = req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	if !config.DocumentRoot.IsNull() && !config.RootSubdir.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Configuration",
			"Cannot specify both document_root and root_subdir in configuration. Use root_subdir to append to the ISPConfig-generated base path, or document_root to specify the full path.",
		)
		return
	}

	domainID := int(plan.ID.ValueInt64())

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

	// Get current state to check if root_subdir is being added/changed/removed
	var currentState webHostingResourceModel
	diags = req.State.Get(ctx, &currentState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	
	// Get current domain to find the base path
	currentDomain, err2 := r.client.GetWebDomain(domainID)
	if err2 != nil {
		resp.Diagnostics.AddError(
			"Error reading current web hosting",
			fmt.Sprintf("Could not read web hosting ID %d: %s", domainID, err2.Error()),
		)
		return
	}
	currentDocRoot := currentDomain.DocumentRoot

	// Build WebDomain struct
	domain := &client.WebDomain{
		Domain:   plan.Domain.ValueString(),
		ClientID: client.FlexInt(clientID),
	}

	if !plan.IPAddress.IsNull() {
		domain.IPAddress = plan.IPAddress.ValueString()
	}
	if !plan.IPv6Address.IsNull() {
		domain.IPv6Address = plan.IPv6Address.ValueString()
	}
	if !plan.Type.IsNull() {
		domain.Type = plan.Type.ValueString()
	}
	if !plan.ParentDomainID.IsNull() {
		domain.ParentDomainID = client.FlexInt(plan.ParentDomainID.ValueInt64())
	}
	
	// Handle document_root and root_subdir transitions
	configRootSubdir := config.RootSubdir.ValueString()
	stateRootSubdir := currentState.RootSubdir.ValueString()
	
	if configRootSubdir != "" {
		// root_subdir is set in config - combine with base path
		basePath := currentDocRoot
		if stateRootSubdir != "" {
			// Remove old subdir from the path to get the base
			oldSubdir := strings.Trim(stateRootSubdir, "/")
			if strings.HasSuffix(basePath, oldSubdir) {
				basePath = strings.TrimSuffix(basePath, "/"+oldSubdir)
			}
		}
		domain.DocumentRoot = combineDocumentRoot(basePath, configRootSubdir)
		tflog.Debug(ctx, "Updating document root with subdir", map[string]interface{}{
			"base_path": basePath,
			"subdir":    configRootSubdir,
			"new_path":  domain.DocumentRoot,
		})
	} else if stateRootSubdir != "" && configRootSubdir == "" {
		// root_subdir was removed - revert to base path
		basePath := currentDocRoot
		oldSubdir := strings.Trim(stateRootSubdir, "/")
		if strings.HasSuffix(basePath, oldSubdir) {
			basePath = strings.TrimSuffix(basePath, "/"+oldSubdir)
		}
		domain.DocumentRoot = basePath
		tflog.Debug(ctx, "Reverting document root (root_subdir removed)", map[string]interface{}{
			"old_path": currentDocRoot,
			"new_path": basePath,
		})
	} else if !config.DocumentRoot.IsNull() {
		// document_root explicitly set in config
		domain.DocumentRoot = plan.DocumentRoot.ValueString()
	}
	if !plan.PHP.IsNull() {
		domain.PHPVersion = plan.PHP.ValueString()
	}
	if !plan.PHPVersion.IsNull() {
		phpID, err := phpVersionToID(plan.PHPVersion.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid PHP Version",
				err.Error(),
			)
			return
		}
		domain.ServerPHPID = client.FlexInt(phpID)
	}
	if !plan.Active.IsNull() {
		domain.Active = boolToYN(plan.Active.ValueBool())
	}
	// ServerID: use resource value if set, otherwise use provider default
	serverID := r.serverID
	if !plan.ServerID.IsNull() {
		serverID = int(plan.ServerID.ValueInt64())
	}
	if serverID == 0 {
		resp.Diagnostics.AddError(
			"Missing Server ID",
			"Server ID must be set either in the provider configuration or in the resource configuration.",
		)
		return
	}
	domain.ServerID = client.FlexInt(serverID)
	if !plan.HdQuota.IsNull() {
		domain.HdQuota = client.FlexInt(plan.HdQuota.ValueInt64())
	}
	if !plan.TrafficQuota.IsNull() {
		domain.TrafficQuota = client.FlexInt(plan.TrafficQuota.ValueInt64())
	}
	if !plan.CGI.IsNull() {
		domain.CGI = boolToYN(plan.CGI.ValueBool())
	}
	if !plan.SSI.IsNull() {
		domain.SSI = boolToYN(plan.SSI.ValueBool())
	}
	if !plan.Perl.IsNull() {
		domain.Perl = boolToYN(plan.Perl.ValueBool())
	}
	if !plan.Ruby.IsNull() {
		domain.Ruby = boolToYN(plan.Ruby.ValueBool())
	}
	if !plan.Python.IsNull() {
		domain.Python = boolToYN(plan.Python.ValueBool())
	}
	if !plan.SuExec.IsNull() {
		domain.SuExec = boolToYN(plan.SuExec.ValueBool())
	}
	if !plan.SSL.IsNull() {
		domain.SSL = boolToYN(plan.SSL.ValueBool())
	}
	if !plan.Subdomain.IsNull() {
		domain.Subdomain = plan.Subdomain.ValueString()
	}
	if !plan.RedirectType.IsNull() {
		domain.RedirectType = plan.RedirectType.ValueString()
	}
	if !plan.RedirectPath.IsNull() {
		domain.RedirectPath = plan.RedirectPath.ValueString()
	}
	if !plan.AllowOverride.IsNull() {
		domain.AllowOverride = plan.AllowOverride.ValueString()
	}
	if !plan.PM.IsNull() {
		domain.PM = plan.PM.ValueString()
	}
	if !plan.PMProcessIdleTimeout.IsNull() {
		domain.PMProcess = plan.PMProcessIdleTimeout.ValueString()
	}
	if !plan.PMMaxRequests.IsNull() {
		domain.PMMaxRequests = client.FlexInt(plan.PMMaxRequests.ValueInt64())
	}
	if !plan.HTTPPort.IsNull() {
		domain.HTTPPort = client.FlexInt(plan.HTTPPort.ValueInt64())
	}
	if !plan.HTTPSPort.IsNull() {
		domain.HTTPSPort = client.FlexInt(plan.HTTPSPort.ValueInt64())
	}
	if !plan.PHPOpenBasedir.IsNull() {
		domain.PHPOpenBasedir = plan.PHPOpenBasedir.ValueString()
	}
	if !plan.ApacheDirectives.IsNull() {
		domain.ApacheDirectives = plan.ApacheDirectives.ValueString()
	}
	// Always send disable_symlink_restriction (defaults to false/"n")
	domain.DisableSymlinkNotOwner = boolToYN(plan.DisableSymlinkNotOwner.ValueBool())

	// Update web domain
	err := r.client.UpdateWebDomain(domainID, clientID, domain)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating web hosting",
			fmt.Sprintf("Could not update web hosting ID %d: %s", domainID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Updated web hosting", map[string]interface{}{"id": domainID})

	// Read back the updated resource
	updatedDomain, err := r.client.GetWebDomain(domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated web hosting",
			"Could not read updated web hosting, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with any computed values - always set when Unknown or Null
	if plan.IPAddress.IsNull() || plan.IPAddress.IsUnknown() {
		plan.IPAddress = types.StringValue(updatedDomain.IPAddress)
	}
	if plan.IPv6Address.IsNull() || plan.IPv6Address.IsUnknown() {
		plan.IPv6Address = types.StringValue(updatedDomain.IPv6Address)
	}
	if plan.ServerID.IsNull() || plan.ServerID.IsUnknown() {
		plan.ServerID = types.Int64Value(int64(updatedDomain.ServerID))
	}
	if plan.DocumentRoot.IsNull() || plan.DocumentRoot.IsUnknown() {
		plan.DocumentRoot = types.StringValue(updatedDomain.DocumentRoot)
	}
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		plan.Type = types.StringValue(updatedDomain.Type)
	}
	if plan.PHP.IsNull() || plan.PHP.IsUnknown() {
		plan.PHP = types.StringValue(updatedDomain.PHPVersion)
	}
	if plan.PHPVersion.IsNull() || plan.PHPVersion.IsUnknown() {
		phpVersion := phpIDToVersion(int(updatedDomain.ServerPHPID))
		plan.PHPVersion = types.StringValue(phpVersion)
	}
	if plan.ParentDomainID.IsNull() || plan.ParentDomainID.IsUnknown() {
		plan.ParentDomainID = types.Int64Value(int64(updatedDomain.ParentDomainID))
	}
	if plan.HdQuota.IsNull() || plan.HdQuota.IsUnknown() {
		plan.HdQuota = types.Int64Value(int64(updatedDomain.HdQuota))
	}
	if plan.TrafficQuota.IsNull() || plan.TrafficQuota.IsUnknown() {
		plan.TrafficQuota = types.Int64Value(int64(updatedDomain.TrafficQuota))
	}
	if plan.RedirectType.IsNull() || plan.RedirectType.IsUnknown() {
		plan.RedirectType = types.StringValue(updatedDomain.RedirectType)
	}
	if plan.RedirectPath.IsNull() || plan.RedirectPath.IsUnknown() {
		plan.RedirectPath = types.StringValue(updatedDomain.RedirectPath)
	}
	if plan.PMMaxRequests.IsNull() || plan.PMMaxRequests.IsUnknown() {
		plan.PMMaxRequests = types.Int64Value(int64(updatedDomain.PMMaxRequests))
	}
	if plan.PHPOpenBasedir.IsNull() || plan.PHPOpenBasedir.IsUnknown() {
		plan.PHPOpenBasedir = types.StringValue(updatedDomain.PHPOpenBasedir)
	}
	if plan.ApacheDirectives.IsNull() || plan.ApacheDirectives.IsUnknown() {
		plan.ApacheDirectives = types.StringValue(updatedDomain.ApacheDirectives)
	}
	if plan.DisableSymlinkNotOwner.IsNull() || plan.DisableSymlinkNotOwner.IsUnknown() {
		plan.DisableSymlinkNotOwner = types.BoolValue(ynToBool(updatedDomain.DisableSymlinkNotOwner))
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webHostingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webHostingResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID := int(state.ID.ValueInt64())

	err := r.client.DeleteWebDomain(domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting web hosting",
			fmt.Sprintf("Could not delete web hosting ID %d: %s", domainID, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "Deleted web hosting", map[string]interface{}{"id": domainID})
}

// ImportState imports the resource state.
func (r *webHostingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
