# Query an existing database user by ID
data "ispconfig_web_database_user" "existing" {
  id = 321
}

# Output database user details
output "database_username" {
  value       = data.ispconfig_web_database_user.existing.database_user
  description = "The username of the database user"
}

# Use an existing database user with a new database
resource "ispconfig_web_database" "new_db" {
  database_name    = "new_application_db"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  database_user_id = data.ispconfig_web_database_user.existing.id
  active           = "y"
  quota            = 500
}
