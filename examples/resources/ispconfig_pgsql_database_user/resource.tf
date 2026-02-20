resource "ispconfig_pgsql_database_user" "example" {
  database_user     = "myapp_user"
  database_password = var.db_password
}
