# Create a MySQL database
resource "ispconfig_web_database" "production" {
  database_name    = "myapp_production"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  active           = "y"

  # Optional: link to a database user
  database_user_id = ispconfig_web_database_user.example.id

  # Optional: quota in MB
  quota = 500
}

# Database with remote access enabled
resource "ispconfig_web_database" "remote_access" {
  database_name    = "myapp_analytics"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  active           = "y"
  remote_access    = "y"
  remote_ips       = "10.0.0.0/8" # Allow from internal network only
  quota            = 1000         # 1GB
}

# Staging database
resource "ispconfig_web_database" "staging" {
  database_name    = "myapp_staging"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  active           = "y"
  database_user_id = ispconfig_web_database_user.example.id
  quota            = 200
}

# Test database with wider remote access (for CI/CD)
resource "ispconfig_web_database" "test" {
  database_name    = "myapp_test"
  parent_domain_id = ispconfig_web_hosting.example.id
  type             = "mysql"
  active           = "y"
  remote_access    = "y"
  remote_ips       = "192.168.1.0/24,10.0.0.0/8" # Multiple CIDR blocks
  quota            = 100
}

# Output the database IDs
output "production_database_id" {
  value = ispconfig_web_database.production.id
}

output "staging_database_id" {
  value = ispconfig_web_database.staging.id
}
