# Query an existing database by ID
data "ispconfig_web_database" "existing" {
  id = 789
}

# Output database details
output "database_name" {
  value       = data.ispconfig_web_database.existing.database_name
  description = "The name of the database"
}

output "database_type" {
  value       = data.ispconfig_web_database.existing.type
  description = "The type of database (mysql, etc.)"
}

output "database_quota" {
  value       = data.ispconfig_web_database.existing.quota
  description = "The quota in MB for this database"
}

output "database_active" {
  value       = data.ispconfig_web_database.existing.active
  description = "Whether the database is active"
}

output "remote_access_enabled" {
  value       = data.ispconfig_web_database.existing.remote_access
  description = "Whether remote access is enabled"
}
