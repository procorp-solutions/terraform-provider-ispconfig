# Create a database user
resource "ispconfig_web_database_user" "example" {
  database_user     = "app_user"
  database_password = var.database_user_password
}

# Create a read-only database user (for reporting)
resource "ispconfig_web_database_user" "readonly" {
  database_user     = "reporting_user"
  database_password = var.reporting_user_password
}

# Create a database and link it to the user
resource "ispconfig_web_database" "linked_db" {
  database_name    = "app_production"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  database_user_id = ispconfig_web_database_user.example.id
  active           = "y"
}

# Create a secondary database for the same user
resource "ispconfig_web_database" "linked_db_staging" {
  database_name    = "app_staging"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  database_user_id = ispconfig_web_database_user.example.id
  active           = "y"
}

# Output the database user ID
output "database_user_id" {
  value = ispconfig_web_database_user.example.id
}

# Variables (define in terraform.tfvars or via environment)
variable "database_user_password" {
  type        = string
  description = "Password for the main database user"
  sensitive   = true
}

variable "reporting_user_password" {
  type        = string
  description = "Password for the reporting database user"
  sensitive   = true
}
