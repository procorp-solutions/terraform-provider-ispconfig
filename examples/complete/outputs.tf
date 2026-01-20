# =============================================================================
# Domain Outputs
# =============================================================================

output "main_domain_id" {
  value       = ispconfig_web_hosting.main.id
  description = "The ID of the main web domain"
}

output "main_domain_name" {
  value       = ispconfig_web_hosting.main.domain
  description = "The main domain name"
}

output "main_domain_document_root" {
  value       = ispconfig_web_hosting.main.document_root
  description = "The document root path for the main domain"
}

output "staging_domain_id" {
  value       = var.create_staging ? ispconfig_web_hosting.staging[0].id : null
  description = "The ID of the staging subdomain (if created)"
}

# =============================================================================
# Shell User Outputs
# =============================================================================

output "shell_user_id" {
  value       = ispconfig_web_user.deploy.id
  description = "The ID of the shell user"
}

output "shell_username" {
  value       = ispconfig_web_user.deploy.username
  description = "The username for SSH/SFTP access"
}

output "shell_user_directory" {
  value       = ispconfig_web_user.deploy.dir
  description = "The home directory of the shell user"
}

# =============================================================================
# Database Outputs
# =============================================================================

output "database_user_id" {
  value       = ispconfig_web_database_user.app.id
  description = "The ID of the database user"
}

output "database_username" {
  value       = ispconfig_web_database_user.app.database_user
  description = "The database username"
}

output "production_database_id" {
  value       = ispconfig_web_database.production.id
  description = "The ID of the production database"
}

output "production_database_name" {
  value       = ispconfig_web_database.production.database_name
  description = "The name of the production database"
}

output "staging_database_id" {
  value       = var.create_staging ? ispconfig_web_database.staging[0].id : null
  description = "The ID of the staging database (if created)"
}

output "staging_database_name" {
  value       = var.create_staging ? ispconfig_web_database.staging[0].database_name : null
  description = "The name of the staging database (if created)"
}

# =============================================================================
# Connection Information
# =============================================================================

output "sftp_connection" {
  value       = "sftp://${ispconfig_web_user.deploy.username}@${var.domain_name}"
  description = "SFTP connection string"
}

output "database_connection" {
  value = {
    host     = var.domain_name
    database = ispconfig_web_database.production.database_name
    username = ispconfig_web_database_user.app.database_user
  }
  description = "Database connection information (password not included)"
}
