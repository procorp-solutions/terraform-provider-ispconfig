# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-01-20

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
