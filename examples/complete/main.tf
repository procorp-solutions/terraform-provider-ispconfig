# Complete ISPConfig Terraform Provider Example
# This example demonstrates a full web hosting setup including:
# - Web domain with PHP 8.2
# - Shell user for SSH/SFTP access
# - MySQL database and database user
# - Subdomain configuration

terraform {
  required_version = ">= 1.0"

  required_providers {
    ispconfig = {
      source  = "procorp-solutions/ispconfig"
      version = "~> 1.0"
    }
  }
}

# Configure the provider
provider "ispconfig" {
  host      = var.ispconfig_host
  username  = var.ispconfig_username
  password  = var.ispconfig_password
  insecure  = var.ispconfig_insecure
  client_id = var.ispconfig_client_id
  server_id = var.ispconfig_server_id
}

# =============================================================================
# Main Web Domain
# =============================================================================

resource "ispconfig_web_hosting" "main" {
  domain      = var.domain_name
  server_id   = var.ispconfig_server_id
  ip_address  = var.domain_ip
  php         = "php-fpm"
  php_version = var.php_version
  active      = true
  ssl         = var.enable_ssl
  subdomain   = "www" # Enable www subdomain

  # Resource quotas
  hd_quota      = var.disk_quota_mb
  traffic_quota = var.traffic_quota_mb

  # PHP-FPM process manager settings
  pm              = "dynamic"
  pm_max_requests = var.php_max_requests
}

# =============================================================================
# Shell User for SSH/SFTP Access
# =============================================================================

resource "ispconfig_web_user" "deploy" {
  username         = "${replace(var.domain_name, ".", "_")}_deploy"
  password         = var.shell_user_password
  parent_domain_id = ispconfig_web_hosting.main.id
  shell            = "/bin/bash"
  quota_size       = var.shell_user_quota_mb
  active           = "y"
}

# =============================================================================
# Database User
# =============================================================================

resource "ispconfig_web_database_user" "app" {
  database_user     = "${replace(var.domain_name, ".", "_")}_user"
  database_password = var.database_password
}

# =============================================================================
# Production Database
# =============================================================================

resource "ispconfig_web_database" "production" {
  database_name    = "${replace(var.domain_name, ".", "_")}_prod"
  parent_domain_id = ispconfig_web_hosting.main.id
  type             = "mysql"
  database_user_id = ispconfig_web_database_user.app.id
  active           = "y"
  quota            = var.database_quota_mb

  # Enable remote access if needed (e.g., for external backup tools)
  remote_access = var.database_remote_access ? "y" : "n"
  remote_ips    = var.database_remote_access ? var.database_remote_ips : ""
}

# =============================================================================
# Staging Subdomain (Optional)
# =============================================================================

resource "ispconfig_web_hosting" "staging" {
  count = var.create_staging ? 1 : 0

  domain           = "staging.${var.domain_name}"
  server_id        = var.ispconfig_server_id
  type             = "subdomain"
  parent_domain_id = ispconfig_web_hosting.main.id
  php              = "php-fpm"
  php_version      = var.php_version
  active           = true
  ssl              = var.enable_ssl
  subdomain        = "none"

  # Lower quotas for staging
  hd_quota      = var.disk_quota_mb / 4
  traffic_quota = var.traffic_quota_mb / 4
}

resource "ispconfig_web_database" "staging" {
  count = var.create_staging ? 1 : 0

  database_name    = "${replace(var.domain_name, ".", "_")}_staging"
  parent_domain_id = ispconfig_web_hosting.main.id
  type             = "mysql"
  database_user_id = ispconfig_web_database_user.app.id
  active           = "y"
  quota            = var.database_quota_mb / 4
}
