# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2025-03-17

First stable release. This version consolidates all features and fixes from the 0.x series and is suitable for production use.

### Provider

- **Terraform Plugin Framework** — Built on the Terraform Plugin Framework (not SDKv2). Requires Terraform >= 1.0.
- **Authentication** — Session-based auth against the ISPConfig remote JSON API (`/remote/json.php`). Credentials via config or environment variables (`ISPCONFIG_HOST`, `ISPCONFIG_USERNAME`, `ISPCONFIG_PASSWORD`, `ISPCONFIG_INSECURE`, `ISPCONFIG_CLIENT_ID`, `ISPCONFIG_SERVER_ID`).
- **TLS** — Optional insecure mode for self-signed certificates. Sensitive attributes (e.g. passwords) are marked sensitive in the schema.

### Resources (11)

| Resource | Description |
|----------|-------------|
| `ispconfig_web_hosting` | Web domains with PHP version, SSL, quotas, `php_open_basedir`, `apache_directives`, `disable_symlink_restriction` |
| `ispconfig_web_user` | Shell/SFTP users with quotas and shell assignment |
| `ispconfig_mysql_database` | MySQL databases with quota and remote access (replaces deprecated `ispconfig_web_database` for MySQL) |
| `ispconfig_mysql_database_user` | MySQL database users |
| `ispconfig_pgsql_database` | PostgreSQL databases with quota and remote access (replaces deprecated `ispconfig_web_database` for PostgreSQL) |
| `ispconfig_pgsql_database_user` | PostgreSQL database users |
| `ispconfig_web_database` | *(Deprecated)* Use `ispconfig_mysql_database` or `ispconfig_pgsql_database` |
| `ispconfig_web_database_user` | *(Deprecated)* Use `ispconfig_mysql_database_user` or `ispconfig_pgsql_database_user` |
| `ispconfig_email_domain` | Mail domains with `local_delivery` and `active` |
| `ispconfig_email_inbox` | Mailboxes (inbox, quota, purge settings, forwarding) |
| `ispconfig_cron_task` | Cron jobs (schedule, command/URL, type: `url`/`chrooted`/`full`, `parent_domain_id`) |

### Data Sources (13)

| Data Source | Description |
|-------------|-------------|
| `ispconfig_web_hosting` | Query web hosting domain details |
| `ispconfig_web_user` | Query shell user details |
| `ispconfig_mysql_database` | Query MySQL database details |
| `ispconfig_mysql_database_user` | Query MySQL database user details |
| `ispconfig_pgsql_database` | Query PostgreSQL database details |
| `ispconfig_pgsql_database_user` | Query PostgreSQL database user details |
| `ispconfig_web_database` | Query web database details |
| `ispconfig_web_database_user` | Query web database user details |
| `ispconfig_client` | Query ISPConfig client/customer information |
| `ispconfig_email_domain` | Query email domain details |
| `ispconfig_email_inbox` | Query email inbox details |
| `ispconfig_cron_task` | Query cron task details |

### Features

- **Import** — All resources support `terraform import`.
- **State moves** — `ispconfig_mysql_database` / `ispconfig_pgsql_database` and their user resources support `moved` blocks from deprecated `ispconfig_web_database` / `ispconfig_web_database_user`.
- **Dynamic PHP versions** — `ispconfig_web_hosting` discovers PHP versions from the ISPConfig server API; `php_version` (e.g. `8.4`) is mapped to `server_php_id` automatically.
- **Quotas** — Quota attributes accept MB; conversion to API units (including special values `0` and `-1`) is handled internally.
- **API quirks** — Boolean fields use `y`/`n` in the API; the provider exposes booleans in the schema. Inconsistent API types (e.g. FlexInt) are handled in the client layer.

### Requirements

- Terraform >= 1.0
- Go >= 1.21 (for building from source)
- ISPConfig 3.x with remote API enabled

### Documentation & Examples

- Generated provider and resource docs (e.g. `terraform providers schema`, `go generate`).
- Example configurations in `examples/` for all resources and data sources, plus a complete example.

---

## [0.4.7] - 2026-03-05

### Fixed

- Fixed `ispconfig_email_inbox` `quota` attribute not converting MB to the unit expected by the ISPConfig API. The `quota` attribute accepts megabytes (e.g. `250`); values are now multiplied by `1024×1024÷1000` before being sent to the API and divided back when reading state. Special values `0` (no mail) and `-1` (unlimited) are passed through unchanged.

## [0.4.6] - 2026-02-27

### Added

- Added `local_delivery` boolean attribute to `ispconfig_email_domain` resource and data source. Controls whether mail for the domain is delivered locally on the server (`true`) or relayed externally (`false`). Defaults to `true`. Maps to ISPConfig's `local_delivery` field.

### Fixed

- Fixed `ispconfig_email_inbox` creation failing with `Incorrect integer value` for `purge_trash_days` and `purge_junk_days`. These INT columns reject empty strings; they are now always sent as `'0'` (never purge).
- Fixed `ispconfig_email_domain` being created as inactive. ISPConfig's `mail_domain_add` API ignores the `active` parameter; an immediate `mail_domain_update` call is now issued after creation to apply the correct `active` and `server_id` values.

## [0.4.5] - 2026-02-27

### Fixed

- Fixed `ispconfig_email_inbox` creation failing with `Data truncated for column 'move_junk'`. The `move_junk` column in the ISPConfig `mail_user` table is `CHAR(1)` and rejects an empty string; the field is now always sent as `'n'` (disabled) for both Create and Update.
- Fixed `ispconfig_email_inbox` Update not falling back to the provider-level `server_id` when the resource does not specify one.
- Fixed `ispconfig_email_domain` Update not falling back to the provider-level `server_id` and not defaulting `active` to `'y'` when not explicitly set.
- Removed `omitempty` from the `active` field in `MailDomain` so that ISPConfig always receives an explicit value instead of omitting it and defaulting to inactive.

## [0.4.4] - 2026-02-27

### Fixed

- Fixed `ispconfig_cron_task` Get, Update, and Delete operations using the wrong parameter name (`primary_id`) for the ISPConfig REST API. The `sites_cron_*` functions expect `cron_id` (matched by name, not position), causing reads and deletes to silently operate on no record.

## [0.4.3] - 2026-02-26

### Fixed

- Debugging errors after `ispconfig_cron_task` creation response.

## [0.4.2] - 2026-02-26

### Fixed

- Fixed `ispconfig_cron_task` Read failing with `cannot unmarshal array into Go value` when ISPConfig returns the cron record wrapped in a JSON array (`[{...}]`) rather than a plain object.

## [0.4.1] - 2026-02-26

### Fixed

- Fixed `ispconfig_cron_task` Create and Update failing with `server_id_error_empty` when `server_id` was not explicitly set in the resource. The resource now falls back to the provider-level `server_id` and returns a clear error if neither is set.
- Removed post-create `GetCronJob` call that caused spurious errors; computed `server_id` is now populated directly from the known value.

### Changed

- Improved `command` attribute description to clarify the 255-character limit and expected format per cron type.
- Improved `type` attribute description to explain the behaviour of each allowed value (`url`, `chrooted`, `full`).

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
- **Note:** supersedes v0.1.0 which was an internal pre-release tag.
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
