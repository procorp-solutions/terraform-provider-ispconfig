# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [0.4.0] - 2026-02-26

### Added

- Added `ispconfig_cron_task` resource and data source for managing scheduled cron jobs in ISP Config. Supports standard cron schedule format (`* * * * *`), command or URL execution, `parent_domain_id`, job `type` (`url`, `chrooted`, `full`), and `active` boolean.
- Added `forward_incoming_to` and `forward_outgoing_to` attributes to `ispconfig_email_inbox` resource and data source for configuring incoming mail forwarding and outgoing BCC respectively.

## [0.3.5] - 2026-02-23

### Fixed

- Fixed linters errors in CI/CD Pipeline.

## [0.3.4] - 2026-02-23

### Added

- Added `ResourceWithMoveState` support to `ispconfig_mysql_database_user` and `ispconfig_pgsql_database_user`, allowing state to be moved from the deprecated `ispconfig_web_database_user` using a `moved` block.

## [0.3.3] - 2026-02-23

### Added

- Added `ResourceWithMoveState` support to `ispconfig_mysql_database` and `ispconfig_pgsql_database`, allowing state to be moved from the deprecated `ispconfig_web_database` using a `moved` block without recreating the underlying resource.

## [0.3.2] - 2026-02-23

### Added

- Added `ispconfig_email_domain` resource for managing email domains in ISPConfig. **Experimental** — requires confirming API method names (`mail_domain_*`) against your ISPConfig version.
- Added `ispconfig_email_inbox` resource for managing email inboxes (mailboxes) assigned to an email domain. **Experimental** — requires confirming API method names (`mail_user_*`) against your ISPConfig version.
- Added `ispconfig_email_domain` data source for reading email domain details.
- Added `ispconfig_email_inbox` data source for reading email inbox details.

### Fixed

- Fixed `ispconfig_pgsql_database` resource missing `remote_access` and `remote_ips` attributes that were already present in `ispconfig_mysql_database`.
- Fixed `ispconfig_pgsql_database` data source missing `remote_access` and `remote_ips` attributes.

## [0.3.1] - 2026-02-20

### Changed

- Regenerated provider documentation with `go generate` to update the Terraform Registry docs.

## [0.3.0] - 2026-02-20

### Added

- Added `ispconfig_mysql_database` resource for managing MySQL databases (engine type hardcoded, no `type` attribute required).
- Added `ispconfig_mysql_database_user` resource for managing MySQL database users.
- Added `ispconfig_pgsql_database` resource for managing PostgreSQL databases (engine type hardcoded, no `type` attribute required).
- Added `ispconfig_pgsql_database_user` resource for managing PostgreSQL database users.
- Added `ispconfig_mysql_database` data source for reading MySQL database details.
- Added `ispconfig_mysql_database_user` data source for reading MySQL database user details.
- Added `ispconfig_pgsql_database` data source for reading PostgreSQL database details.
- Added `ispconfig_pgsql_database_user` data source for reading PostgreSQL database user details.

### Deprecated

- `ispconfig_web_database` resource is now deprecated. Migrate to `ispconfig_mysql_database` or `ispconfig_pgsql_database`. The resource remains functional for backward compatibility and will be removed in a future major release.
- `ispconfig_web_database_user` resource is now deprecated. Migrate to `ispconfig_mysql_database_user` or `ispconfig_pgsql_database_user`.

## [0.2.1] - 2026-02-13

- Re-tag of v0.2.0; no code changes.

## [0.2.0] - 2026-02-13

### Added

- Added dynamic PHP version discovery for `ispconfig_web_hosting` via ISPConfig server API, including automatic mapping between `php_version` (e.g. `8.4`) and `server_php_id`

### Changed

- Improved `ispconfig_web_hosting` `php_version` schema/docs to reflect dynamically fetched versions instead of a static hardcoded list
- Enhanced `server_id` fallback behavior across resources to honor provider-level defaults when resource-specific values are not set

### Fixed

- Fixed `ispconfig_web_database` `server_id` inheritance by deriving it from `parent_domain_id` and falling back to provider configuration when needed
- Fixed `ispconfig_web_database_user` creation/update without explicit `server_id` by inheriting provider-level `server_id`
- Fixed `ispconfig_web_user` server assignment edge cases by prioritizing resource `server_id`, then parent domain, then provider-level fallback
- Fixed web hosting documentation examples to align with provider behavior and current `php_version` handling

## [0.1.5] - 2026-01-22

### Added

- Added `disable_symlink_restriction` parameter to `ispconfig_web_hosting` resource and data source to deactivate symlinks restriction of the web space (maps to ISPConfig's `disable_symlinknotowner` field)

## [0.1.4] - 2026-01-22

### Fixed

- Fixed state migration issue: Added automatic state upgrade from version 0 (string `"y"`/`"n"` values) to version 1 (boolean values) for `ispconfig_web_user` and `ispconfig_web_database` resources
- This resolves the "a bool is required" error when reading existing state files that contain string values for `active` and `remote_access` attributes

## [0.1.3] - 2026-01-22

### Fixed

- Fixed `ispconfig_web_user` resource not inheriting `server_id` from parent domain, causing users to be created without server assignment in ISPConfig UI

### Changed

- **BREAKING:** Changed `active` attribute in `ispconfig_web_user` resource and data source from string (`"y"`/`"n"`) to boolean (`true`/`false`)
- **BREAKING:** Changed `active` and `remote_access` attributes in `ispconfig_web_database` resource and data source from string (`"y"`/`"n"`) to boolean (`true`/`false`)
- This change unifies the boolean handling across all resources to match `ispconfig_web_hosting`

## [0.1.2] - 2026-01-21

### Added

- Added `php_open_basedir` parameter to `ispconfig_web_hosting` resource for PHP open_basedir restrictions
- Added `apache_directives` parameter to `ispconfig_web_hosting` resource for custom Apache configuration directives

## [0.1.1] - 2026-01-20

### Added

- Initial release of ISPConfig Terraform Provider
- **Resources:**
  - `ispconfig_web_hosting` - Manage web hosting domains with PHP, SSL, quotas, and more
  - `ispconfig_web_user` - Manage shell/SFTP users with quotas and shell assignments
  - `ispconfig_web_database` - Manage MySQL databases with quotas and remote access
  - `ispconfig_web_database_user` - Manage database users
- **Data Sources:**
  - `ispconfig_web_hosting` - Query existing web hosting domains
  - `ispconfig_web_user` - Query existing shell users
  - `ispconfig_web_database` - Query existing databases
  - `ispconfig_web_database_user` - Query existing database users
  - `ispconfig_client` - Query ISPConfig client/customer information
- Session-based authentication with ISPConfig remote API
- TLS/SSL support with optional insecure mode for self-signed certificates
- Support for environment variables configuration
- Resource import support for all resources
- Provider-level default client_id and server_id settings
- Comprehensive examples and documentation
- GitHub Actions CI/CD workflows for testing and releases
- GPG-signed releases for Terraform Registry

### Known Limitations

- Requires Go 1.21+ for building from source
- Session management does not implement automatic session refresh on expiration
- Only MySQL database type is currently tested

### Security

- All sensitive attributes (passwords) are marked as sensitive in schema
- Examples use variables instead of hardcoded credentials
- Provider configuration supports environment variables for secrets
