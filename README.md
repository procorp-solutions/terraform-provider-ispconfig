# Terraform Provider for ISPConfig

[![CI](https://github.com/procorp-solutions/ispconfig-terraform-provider/actions/workflows/ci.yml/badge.svg)](https://github.com/procorp-solutions/ispconfig-terraform-provider/actions/workflows/ci.yml)
[![Release](https://github.com/procorp-solutions/ispconfig-terraform-provider/actions/workflows/release.yml/badge.svg)](https://github.com/procorp-solutions/ispconfig-terraform-provider/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This Terraform provider enables you to manage [ISPConfig](https://www.ispconfig.org/) resources through Infrastructure as Code. ISPConfig is a popular open-source hosting control panel for Linux servers.

## Features

- **Web Hosting Management** - Create and manage web domains with PHP, SSL, and custom configurations
- **Shell Users** - Manage SSH/SFTP users with quotas and shell assignments
- **Databases** - Create MySQL databases with quota and remote access controls
- **Database Users** - Manage database users and credentials
- **Email Domains** - Create and manage mail domains
- **Email Inboxes** - Create and manage mailboxes (email inboxes) assigned to a mail domain
- **Cron Tasks** - Schedule cron jobs using standard cron format (`* * * * *`)
- **Data Sources** - Query existing ISPConfig resources for reference in your configurations
- **Import Support** - Import existing resources into Terraform state

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- ISPConfig 3.x with remote API enabled

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    ispconfig = {
      source  = "procorp-solutions/ispconfig"
      version = "~> 1.0"
    }
  }
}
```

Run `terraform init` to download and install the provider.

### Building from Source

1. Clone the repository:

```bash
git clone https://github.com/procorp-solutions/ispconfig-terraform-provider.git
cd ispconfig-terraform-provider
```

2. Build the provider:

```bash
go build -o terraform-provider-ispconfig
```

3. Install the provider locally:

**Linux/macOS:**
```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/procorp-solutions/ispconfig/1.0.0/linux_amd64
cp terraform-provider-ispconfig ~/.terraform.d/plugins/registry.terraform.io/procorp-solutions/ispconfig/1.0.0/linux_amd64/
```

**Windows:**
```powershell
mkdir -p $env:APPDATA\terraform.d\plugins\registry.terraform.io\procorp-solutions\ispconfig\1.0.0\windows_amd64
cp terraform-provider-ispconfig.exe $env:APPDATA\terraform.d\plugins\registry.terraform.io\procorp-solutions\ispconfig\1.0.0\windows_amd64\
```

## Quick Start

### Provider Configuration

```hcl
provider "ispconfig" {
  host     = "your-ispconfig-server.com:8080"
  username = var.ispconfig_username
  password = var.ispconfig_password
  
  # Optional settings
  insecure  = false  # Set to true for self-signed certificates
  client_id = 1      # Default client ID for resources
  server_id = 1      # Default server ID for resources
}
```

### Environment Variables

You can also configure the provider using environment variables:

| Variable | Description |
|----------|-------------|
| `ISPCONFIG_HOST` | ISPConfig host and port |
| `ISPCONFIG_USERNAME` | ISPConfig username |
| `ISPCONFIG_PASSWORD` | ISPConfig password |
| `ISPCONFIG_INSECURE` | Set to "true" to skip TLS verification |
| `ISPCONFIG_CLIENT_ID` | Default client ID |
| `ISPCONFIG_SERVER_ID` | Default server ID |

### Basic Example

```hcl
# Create a web hosting domain with PHP 8.2
resource "ispconfig_web_hosting" "example" {
  domain      = "example.com"
  server_id   = 1
  php         = "php-fpm"
  php_version = "8.2"
  active      = true
  ssl         = true
}

# Create a shell user
resource "ispconfig_web_user" "deploy" {
  username         = "deploy"
  password         = var.shell_password
  parent_domain_id = ispconfig_web_hosting.example.id
  shell            = "/bin/bash"
}

# Create a database user
resource "ispconfig_web_database_user" "app" {
  database_user     = "app_user"
  database_password = var.db_password
}

# Create a database
resource "ispconfig_web_database" "production" {
  database_name    = "app_production"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  database_user_id = ispconfig_web_database_user.app.id
  active           = "y"
}
```

## Resources

### ispconfig_web_hosting

Manages a web hosting domain.

**Required Arguments:**
- `domain` - The domain name
- `server_id` - The server ID where the domain is hosted

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `ip_address` - IP address for the domain (default: auto-assigned)
- `type` - Domain type: `vhost`, `subdomain` (default: `vhost`)
- `parent_domain_id` - Parent domain ID for subdomains
- `document_root` - Full path to the document root directory
- `root_subdir` - Subdirectory path to append to the ISPConfig-generated base document root
- `php` - PHP mode: `php-fpm`, `fast-cgi`, `mod`, `no`
- `php_version` - PHP version: `7.4`, `8.0`, `8.1`, `8.2`, `8.3`, `8.4`
- `active` - Activate domain (default: `true`)
- `ssl` - Enable SSL (default: `false`)
- `subdomain` - Subdomain auto-redirect: `www`, `none`, `*` (default: `www`)
- `hd_quota` - Hard disk quota in MB
- `traffic_quota` - Traffic quota in MB
- `pm` - PHP-FPM process manager: `dynamic`, `static`, `ondemand`
- `pm_max_requests` - PHP-FPM max requests per process
- `cgi`, `ssi`, `perl`, `ruby`, `python` - Enable respective features
- `suexec` - Enable SuExec (default: `true`)
- `http_port`, `https_port` - Custom port numbers

### ispconfig_web_user

Manages a shell/SFTP user.

**Required Arguments:**
- `username` - The shell username
- `password` - The shell user password
- `parent_domain_id` - The parent domain ID

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `dir` - The shell user directory path
- `shell` - The shell: `/bin/bash`, `/bin/sh`, `/bin/false`, `/sbin/nologin`
- `quota_size` - Quota size in MB
- `active` - Whether active: `y` or `n`

### ispconfig_web_database

Manages a MySQL database.

**Required Arguments:**
- `database_name` - The database name
- `parent_domain_id` - The parent domain ID

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `type` - Database type (default: `mysql`)
- `database_user_id` - Link to a database user
- `active` - Whether active: `y` or `n`
- `remote_access` - Enable remote access: `y` or `n`
- `remote_ips` - Allowed IPs for remote access (CIDR notation)
- `quota` - Quota in MB

### ispconfig_web_database_user

Manages a database user.

**Required Arguments:**
- `database_user` - The database username
- `database_password` - The database user password

**Optional Arguments:**
- `client_id` - Override the provider's default client ID

> **Deprecated:** Use `ispconfig_mysql_database_user` or `ispconfig_pgsql_database_user` instead.

### ispconfig_mysql_database

Manages a MySQL database.

**Required Arguments:**
- `database_name` - The database name
- `parent_domain_id` - The parent domain ID

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `database_user_id` - Link to a database user
- `quota` - Quota in MB
- `active` - Whether active (default: `true`)
- `server_id` - The server ID
- `remote_access` - Enable remote access (default: `false`)
- `remote_ips` - Comma-separated list of IPs allowed for remote access

### ispconfig_mysql_database_user

Manages a MySQL database user.

**Required Arguments:**
- `database_user` - The database username
- `database_password` - The database user password

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `server_id` - The server ID

### ispconfig_pgsql_database

Manages a PostgreSQL database.

**Required Arguments:**
- `database_name` - The database name
- `parent_domain_id` - The parent domain ID

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `database_user_id` - Link to a database user
- `quota` - Quota in MB
- `active` - Whether active (default: `true`)
- `server_id` - The server ID
- `remote_access` - Enable remote access (default: `false`)
- `remote_ips` - Comma-separated list of IPs allowed for remote access

### ispconfig_pgsql_database_user

Manages a PostgreSQL database user.

**Required Arguments:**
- `database_user` - The database username
- `database_password` - The database user password

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `server_id` - The server ID

### ispconfig_email_domain

Manages a mail domain. Email inboxes are assigned to a domain.

**Required Arguments:**
- `domain` - The email domain name (e.g. `example.com`)

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `server_id` - The mail server ID
- `active` - Whether the domain is active: `y` or `n` (default: `y`)

### ispconfig_email_inbox

Manages an email inbox (mailbox) assigned to a mail domain.

**Required Arguments:**
- `maildomain_id` - The ID of the email domain this inbox belongs to
- `email` - The full email address (e.g. `user@example.com`)
- `password` - The mailbox password

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `quota` - Mailbox quota in MB; `0` = no mail allowed, `-1` = unlimited
- `server_id` - The mail server ID
- `forward_incoming_to` - Forward all incoming mail to this email address
- `forward_outgoing_to` - BCC all outgoing mail to this email address

### ispconfig_cron_task

Manages a cron task (scheduled job) in ISP Config.

**Required Arguments:**
- `parent_domain_id` - The ID of the parent domain this cron task belongs to
- `schedule` - The cron schedule in standard format `"* * * * *"` (min hour mday month wday)
- `command` - The command, script path, or URL to execute

**Optional Arguments:**
- `client_id` - Override the provider's default client ID
- `type` - The cron job type: `url` (default), `chrooted`, or `full`
- `active` - Whether the cron task is active (default: `true`)
- `server_id` - The server ID

## Data Sources

All resources have corresponding data sources for querying existing resources:

- `ispconfig_web_hosting` - Query web hosting domains
- `ispconfig_web_user` - Query shell users
- `ispconfig_web_database` - Query databases (deprecated)
- `ispconfig_web_database_user` - Query database users (deprecated)
- `ispconfig_mysql_database` - Query MySQL databases
- `ispconfig_mysql_database_user` - Query MySQL database users
- `ispconfig_pgsql_database` - Query PostgreSQL databases
- `ispconfig_pgsql_database_user` - Query PostgreSQL database users
- `ispconfig_email_domain` - Query email domains
- `ispconfig_email_inbox` - Query email inboxes
- `ispconfig_cron_task` - Query cron tasks
- `ispconfig_client` - Query ISPConfig client information

```hcl
# Query an existing domain
data "ispconfig_web_hosting" "existing" {
  id = 123
}

# Query an email domain
data "ispconfig_email_domain" "mail" {
  id = 42
}

# Query client limits
data "ispconfig_client" "current" {
  id = 1
}

output "web_domain_limit" {
  value = data.ispconfig_client.current.limit_web
}
```

## Importing Existing Resources

You can import existing ISPConfig resources using their ID:

```bash
# Import a web hosting domain
terraform import ispconfig_web_hosting.example 123

# Import a shell user
terraform import ispconfig_web_user.deploy 456

# Import a database
terraform import ispconfig_web_database.production 789

# Import a database user
terraform import ispconfig_web_database_user.app 321

# Import an email domain
terraform import ispconfig_email_domain.example 10

# Import an email inbox
terraform import ispconfig_email_inbox.user 20

# Import a cron task
terraform import ispconfig_cron_task.backup 30
```

## Examples

See the [examples](./examples/) directory for complete usage examples:

- [Provider Configuration](./examples/provider/) - Basic provider setup
- [Resources](./examples/resources/) - Individual resource examples
- [Data Sources](./examples/data-sources/) - Data source usage examples
- [Complete Setup](./examples/complete/) - Full web hosting environment

## Development

### Prerequisites

- Go 1.21+
- Terraform 1.0+
- Access to an ISPConfig instance with remote API enabled

### Building

```bash
go build -o terraform-provider-ispconfig
```

### Running Tests

```bash
go test -v ./...
```

### Acceptance Tests

Acceptance tests require a running ISPConfig instance:

```bash
export ISPCONFIG_HOST="your-host:8080"
export ISPCONFIG_USERNAME="your-username"
export ISPCONFIG_PASSWORD="your-password"
export TF_ACC=1

go test -v ./internal/provider -timeout 120m
```

### Debugging

Run the provider in debug mode:

```bash
./terraform-provider-ispconfig -debug
```

Use the `TF_REATTACH_PROVIDERS` environment variable as instructed.

## Troubleshooting

### TLS Certificate Errors

For self-signed certificates, enable insecure mode:

```hcl
provider "ispconfig" {
  host     = "your-server:8080"
  username = var.username
  password = var.password
  insecure = true
}
```

### Authentication Errors

Ensure that:
1. The remote API is enabled in ISPConfig
2. Your user has the necessary API permissions
3. The credentials are correct

### Session Timeout

The provider maintains a session with the ISPConfig API. If you experience session timeout issues, the provider will automatically re-authenticate.

## ISPConfig API Reference

This provider uses the ISPConfig remote API. The following methods are utilized:

| Resource | API Methods |
|----------|-------------|
| Web Domain | `sites_web_domain_add`, `sites_web_domain_get`, `sites_web_domain_update`, `sites_web_domain_delete` |
| Shell User | `sites_shell_user_add`, `sites_shell_user_get`, `sites_shell_user_update`, `sites_shell_user_delete` |
| Database | `sites_database_add`, `sites_database_get`, `sites_database_update`, `sites_database_delete` |
| Database User | `sites_database_user_add`, `sites_database_user_get`, `sites_database_user_update`, `sites_database_user_delete` |
| Email Domain | `mail_domain_add`, `mail_domain_get`, `mail_domain_update`, `mail_domain_delete` |
| Email Inbox | `mail_user_add`, `mail_user_get`, `mail_user_update`, `mail_user_delete` |
| Cron Task | `sites_cron_add`, `sites_cron_get`, `sites_cron_update`, `sites_cron_delete` |
| Client | `client_get`, `client_get_all` |
| Authentication | `login`, `logout` |

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
- [ISPConfig](https://www.ispconfig.org/)
- [GoReleaser](https://goreleaser.com/)
