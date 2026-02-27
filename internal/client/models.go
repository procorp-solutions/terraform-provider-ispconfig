package client

import (
	"encoding/json"
	"strconv"
)

// FlexInt is a helper type that can unmarshal from either string or int
type FlexInt int

// UnmarshalJSON implements the json.Unmarshaler interface
func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*fi = FlexInt(i)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*fi = 0
			return nil
		}
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*fi = FlexInt(i)
		return nil
	}

	return nil
}

// LoginRequest represents the login request to ISP Config API
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response from ISP Config API
type LoginResponse struct {
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	Response interface{} `json:"response"` // session_id (string on success) or false (bool on error)
}

// APIResponse represents a generic API response
type APIResponse struct {
	Code     string      `json:"code"`
	Message  string      `json:"message"`
	Response interface{} `json:"response"`
}

// WebDomain represents a web hosting domain
type WebDomain struct {
	ID              FlexInt `json:"domain_id,omitempty"`
	SysUserID       FlexInt `json:"sys_userid,omitempty"`
	SysGroupID      FlexInt `json:"sys_groupid,omitempty"`
	SysPerm         string  `json:"sys_perm_user,omitempty"`
	ClientID        FlexInt `json:"client_id,omitempty"`
	ServerID        FlexInt `json:"server_id,omitempty"`
	IPAddress       string  `json:"ip_address,omitempty"`
	IPv6Address     string  `json:"ipv6_address,omitempty"`
	Domain          string  `json:"domain"`
	Type            string  `json:"type,omitempty"`
	ParentDomainID  FlexInt `json:"parent_domain_id,omitempty"`
	Vhost           string  `json:"vhost_type,omitempty"`
	DocumentRoot    string  `json:"document_root,omitempty"`
	System          string  `json:"system_user,omitempty"`
	SystemGroup     string  `json:"system_group,omitempty"`
	HdQuota         FlexInt `json:"hd_quota,omitempty"`
	TrafficQuota    FlexInt `json:"traffic_quota,omitempty"`
	CGI             string  `json:"cgi,omitempty"`
	SSI             string  `json:"ssi,omitempty"`
	Perl            string  `json:"perl,omitempty"`
	Ruby            string  `json:"ruby,omitempty"`
	Python          string  `json:"python,omitempty"`
	SuExec          string  `json:"suexec,omitempty"`
	ErrDoc          string  `json:"errordocs,omitempty"`
	Subdomain       string  `json:"subdomain,omitempty"`
	SSL             string  `json:"ssl,omitempty"`
	SSLState        string  `json:"ssl_state,omitempty"`
	SSLLocality     string  `json:"ssl_locality,omitempty"`
	SSLOrganization string  `json:"ssl_organisation,omitempty"`
	SSLOrgUnit      string  `json:"ssl_organisation_unit,omitempty"`
	SSLCountry      string  `json:"ssl_country,omitempty"`
	SSLDomain       string  `json:"ssl_domain,omitempty"`
	SSLRequest      string  `json:"ssl_request,omitempty"`
	SSLCert         string  `json:"ssl_cert,omitempty"`
	SSLBundle       string  `json:"ssl_bundle,omitempty"`
	SSLKey          string  `json:"ssl_key,omitempty"`
	SSLAction       string  `json:"ssl_action,omitempty"`
	PHPVersion         string  `json:"php,omitempty"`
	ServerPHPID        FlexInt `json:"server_php_id,omitempty"`
	FastcgiPHPVersion  string  `json:"fastcgi_php_version,omitempty"`
	Active          string  `json:"active,omitempty"`
	RedirectType    string  `json:"redirect_type,omitempty"`
	RedirectPath    string  `json:"redirect_path,omitempty"`
	SEOURL          string  `json:"seo_redirect,omitempty"`
	RewriteRules    string  `json:"rewrite_rules,omitempty"`
	Added           string  `json:"added_date,omitempty"`
	AddedBy         string  `json:"added_by,omitempty"`
	// Additional fields required by ISPConfig
	AllowOverride     string  `json:"allow_override,omitempty"`
	PHPFPMChroot      string  `json:"php_fpm_chroot,omitempty"`
	PHPFPMIni         string  `json:"php_fpm_ini_dir,omitempty"`
	PM                string  `json:"pm,omitempty"`
	PMMaxChildren     FlexInt `json:"pm_max_children,omitempty"`
	PMStartServers    FlexInt `json:"pm_start_servers,omitempty"`
	PMMinSpareServers FlexInt `json:"pm_min_spare_servers,omitempty"`
	PMMaxSpareServers FlexInt `json:"pm_max_spare_servers,omitempty"`
	PMProcess         string  `json:"pm_process_idle_timeout,omitempty"`
	PMMaxRequests     FlexInt `json:"pm_max_requests,omitempty"`
	HTTPPort          FlexInt `json:"http_port,omitempty"`
	HTTPSPort         FlexInt `json:"https_port,omitempty"`
	PHPOpenBasedir    string  `json:"php_open_basedir,omitempty"`
	ApacheDirectives       string  `json:"apache_directives,omitempty"`
	DisableSymlinkNotOwner string  `json:"disable_symlinknotowner,omitempty"`
	CustomPHPIni           string  `json:"custom_php_ini,omitempty"`
	BackupInterval    string  `json:"backup_interval,omitempty"`
	BackupCopies      FlexInt `json:"backup_copies,omitempty"`
	Stats             string  `json:"stats_type,omitempty"`
	StatsPassword     string  `json:"stats_password,omitempty"`
}

// FTPUser represents an FTP user
type FTPUser struct {
	ID             FlexInt `json:"ftp_user_id,omitempty"`
	SysUserID      FlexInt `json:"sys_userid,omitempty"`
	SysGroupID     FlexInt `json:"sys_groupid,omitempty"`
	ServerID       FlexInt `json:"server_id,omitempty"`
	ParentDomainID FlexInt `json:"parent_domain_id"`
	Username       string  `json:"username"`
	Password       string  `json:"password,omitempty"`
	QuotaSize      FlexInt `json:"quota_size,omitempty"`
	Active         string  `json:"active,omitempty"`
	UID            string  `json:"uid,omitempty"`
	GID            string  `json:"gid,omitempty"`
	Dir            string  `json:"dir,omitempty"`
	QuotaFiles     FlexInt `json:"quota_files,omitempty"`
	ULRatio        FlexInt `json:"ul_ratio,omitempty"`
	DLRatio        FlexInt `json:"dl_ratio,omitempty"`
	ULBandwidth    FlexInt `json:"ul_bandwidth,omitempty"`
	DLBandwidth    FlexInt `json:"dl_bandwidth,omitempty"`
}

// ShellUser represents a shell user
type ShellUser struct {
	ID             FlexInt `json:"shell_user_id,omitempty"`
	SysUserID      FlexInt `json:"sys_userid,omitempty"`
	SysGroupID     FlexInt `json:"sys_groupid,omitempty"`
	ServerID       FlexInt `json:"server_id,omitempty"`
	ParentDomainID FlexInt `json:"parent_domain_id"`
	Username       string  `json:"username"`
	Password       string  `json:"password,omitempty"`
	Shell          string  `json:"shell,omitempty"`
	Active         string  `json:"active,omitempty"`
	UID            string  `json:"uid,omitempty"`
	GID            string  `json:"gid,omitempty"`
	Dir            string  `json:"dir,omitempty"`
	QuotaSize      FlexInt `json:"quota_size,omitempty"`
	QuotaFiles     FlexInt `json:"quota_files,omitempty"`
	PUser          string  `json:"puser,omitempty"`   // System user from parent domain
	PGroup         string  `json:"pgroup,omitempty"`  // System group from parent domain
}

// Database represents a database
type Database struct {
	ID               FlexInt `json:"database_id,omitempty"`
	SysUserID        FlexInt `json:"sys_userid,omitempty"`
	SysGroupID       FlexInt `json:"sys_groupid,omitempty"`
	ServerID         FlexInt `json:"server_id,omitempty"`
	ParentDomainID   FlexInt `json:"parent_domain_id"`
	Type             string  `json:"type,omitempty"`
	DatabaseName     string  `json:"database_name"`
	DatabaseNameOrig string  `json:"database_name_orig,omitempty"`
	DatabaseUser     string  `json:"database_user,omitempty"`
	DatabaseUserID   FlexInt `json:"database_user_id,omitempty"`
	DatabasePassword string  `json:"database_password,omitempty"`
	DatabaseCharset  string  `json:"database_charset,omitempty"`
	RemoteAccess     string  `json:"remote_access,omitempty"`
	RemoteIPs        string  `json:"remote_ips,omitempty"`
	BackupInterval   string  `json:"backup_interval,omitempty"`
	BackupCopies     FlexInt `json:"backup_copies,omitempty"`
	Active           string  `json:"active,omitempty"`
	DatabaseQuota    FlexInt `json:"database_quota,omitempty"`
}

// DatabaseUser represents a database user
type DatabaseUser struct {
	ID               FlexInt `json:"database_user_id,omitempty"`
	SysUserID        FlexInt `json:"sys_userid,omitempty"`
	SysGroupID       FlexInt `json:"sys_groupid,omitempty"`
	ServerID         FlexInt `json:"server_id,omitempty"`
	DatabaseUser     string  `json:"database_user"`
	DatabaseUserOrig string  `json:"database_user_orig,omitempty"`
	DatabasePassword string  `json:"database_password"`
}

// MailDomain represents an ISPConfig mail domain
type MailDomain struct {
	ID       FlexInt `json:"maildomain_id,omitempty"`
	ServerID FlexInt `json:"server_id,omitempty"`
	Domain   string  `json:"domain"`
	// Active and LocalDelivery must always be sent so ISPConfig does not default to wrong values.
	Active        string `json:"active"`
	LocalDelivery string `json:"local_delivery"`
}

// MailUser represents an ISPConfig mailbox (email inbox)
type MailUser struct {
	ID           FlexInt `json:"mailuser_id,omitempty"`
	ServerID     FlexInt `json:"server_id,omitempty"`
	MailDomainID FlexInt `json:"maildomain_id,omitempty"`
	Email        string  `json:"email"`
	Login        string  `json:"login,omitempty"`
	Password     string  `json:"password,omitempty"`
	Maildir      string  `json:"maildir,omitempty"`
	Quota        FlexInt `json:"quota,omitempty"`
	Active       string  `json:"active,omitempty"`
	CC           string  `json:"cc,omitempty"`
	SenderCC     string  `json:"sender_cc,omitempty"`
	// The following fields must always be sent explicitly; the mail_user table
	// uses strict column types and rejects empty strings for these columns.
	MoveJunk      string `json:"move_junk"`       // CHAR(1): 'y' or 'n'
	PurgeTrashDays string `json:"purge_trash_days"` // INT: days before purging trash (0 = never)
	PurgeJunkDays  string `json:"purge_junk_days"`  // INT: days before purging junk (0 = never)
}

// CronJob represents an ISPConfig cron task
type CronJob struct {
	ID             FlexInt `json:"cron_id,omitempty"`
	ServerID       FlexInt `json:"server_id,omitempty"`
	ParentDomainID FlexInt `json:"parent_domain_id"`
	Type           string  `json:"type,omitempty"`
	Command        string  `json:"command"`
	RunMin         string  `json:"run_min"`
	RunHour        string  `json:"run_hour"`
	RunMday        string  `json:"run_mday"`
	RunMonth       string  `json:"run_month"`
	RunWday        string  `json:"run_wday"`
	Active         string  `json:"active,omitempty"`
	Log            string  `json:"log,omitempty"`
}

// ISPConfigClient represents an ISP Config client
type ISPConfigClient struct {
	ID                    FlexInt `json:"client_id,omitempty"`
	SysUserID             FlexInt `json:"sys_userid,omitempty"`
	SysGroupID            FlexInt `json:"sys_groupid,omitempty"`
	ParentClientID        FlexInt `json:"parent_client_id,omitempty"`
	CompanyName           string  `json:"company_name,omitempty"`
	ContactName           string  `json:"contact_name,omitempty"`
	CustomerNo            string  `json:"customer_no,omitempty"`
	VATNumber             string  `json:"vat_id,omitempty"`
	Street                string  `json:"street,omitempty"`
	Zip                   string  `json:"zip,omitempty"`
	City                  string  `json:"city,omitempty"`
	State                 string  `json:"state,omitempty"`
	Country               string  `json:"country,omitempty"`
	Phone                 string  `json:"telephone,omitempty"`
	Mobile                string  `json:"mobile,omitempty"`
	Fax                   string  `json:"fax,omitempty"`
	Email                 string  `json:"email,omitempty"`
	Internet              string  `json:"internet,omitempty"`
	ICQ                   string  `json:"icq,omitempty"`
	Notes                 string  `json:"notes,omitempty"`
	DefaultMailserver     FlexInt `json:"default_mailserver,omitempty"`
	LimitMailDomain       FlexInt `json:"limit_maildomain,omitempty"`
	LimitMailbox          FlexInt `json:"limit_mailbox,omitempty"`
	LimitMailAlias        FlexInt `json:"limit_mailalias,omitempty"`
	LimitMailAliasPattern FlexInt `json:"limit_mailaliasdomain,omitempty"`
	LimitMailForward      FlexInt `json:"limit_mailforward,omitempty"`
	LimitMailCatchall     FlexInt `json:"limit_mailcatchall,omitempty"`
	LimitMailRouting      FlexInt `json:"limit_mailrouting,omitempty"`
	LimitMailFilter       FlexInt `json:"limit_mailfilter,omitempty"`
	LimitFetchmail        FlexInt `json:"limit_fetchmail,omitempty"`
	LimitMailQuota        FlexInt `json:"limit_mailquota,omitempty"`
	LimitSpamfilter       string  `json:"limit_spamfilter_wblist,omitempty"`
	LimitSpamfilterUser   string  `json:"limit_spamfilter_user,omitempty"`
	LimitSpamfilterPolicy string  `json:"limit_spamfilter_policy,omitempty"`
	DefaultWebserver      FlexInt `json:"default_webserver,omitempty"`
	LimitWeb              FlexInt `json:"limit_web_domain,omitempty"`
	LimitWebQuota         FlexInt `json:"limit_web_quota,omitempty"`
	WebPHP                string  `json:"web_php_options,omitempty"`
	LimitCGI              string  `json:"limit_cgi,omitempty"`
	LimitSSI              string  `json:"limit_ssi,omitempty"`
	LimitPerl             string  `json:"limit_perl,omitempty"`
	LimitRuby             string  `json:"limit_ruby,omitempty"`
	LimitPython           string  `json:"limit_python,omitempty"`
	ForceSubdomain        string  `json:"force_suexec,omitempty"`
	LimitHTTPdirs         string  `json:"limit_hterror,omitempty"`
	LimitWildcard         string  `json:"limit_wildcard,omitempty"`
	LimitSSL              string  `json:"limit_ssl,omitempty"`
	LimitSSLLetsencrypt   string  `json:"limit_ssl_letsencrypt,omitempty"`
	LimitTrafficQuota     FlexInt `json:"limit_traffic_quota,omitempty"`
	LimitWebAlias         FlexInt `json:"limit_web_aliasdomain,omitempty"`
	LimitWebSubdomain     FlexInt `json:"limit_web_subdomain,omitempty"`
	LimitFTPUser          FlexInt `json:"limit_ftp_user,omitempty"`
	LimitShellUser        FlexInt `json:"limit_shell_user,omitempty"`
	SSHChroot             string  `json:"ssh_chroot,omitempty"`
	LimitWebdavUser       FlexInt `json:"limit_webdav_user,omitempty"`
	DefaultDNSserver      FlexInt `json:"default_dnsserver,omitempty"`
	LimitDNSZone          FlexInt `json:"limit_dns_zone,omitempty"`
	LimitDNSSlaveZone     FlexInt `json:"limit_dns_slave_zone,omitempty"`
	LimitDNSRecord        FlexInt `json:"limit_dns_record,omitempty"`
	DefaultDBserver       FlexInt `json:"default_dbserver,omitempty"`
	LimitDatabase         FlexInt `json:"limit_database,omitempty"`
	LimitDatabaseQuota    FlexInt `json:"limit_database_quota,omitempty"`
	LimitCronType         string  `json:"limit_cron_type,omitempty"`
	LimitCron             FlexInt `json:"limit_cron,omitempty"`
	LimitCronFrequency    FlexInt `json:"limit_cron_frequency,omitempty"`
	Locked                string  `json:"locked,omitempty"`
	Canceled              string  `json:"canceled,omitempty"`
	Created               string  `json:"created_at,omitempty"`
	Username              string  `json:"username,omitempty"`
	Password              string  `json:"password,omitempty"`
	Language              string  `json:"language,omitempty"`
	UseTheme              string  `json:"usertheme,omitempty"`
	TemplateMenu          string  `json:"template_master,omitempty"`
	TemplateAdditional    string  `json:"template_additional,omitempty"`
	Created_at            string  `json:"created,omitempty"`
}
