resource "ispconfig_mysql_database_user" "example" {
  database_user     = "myapp_user"
  database_password = var.db_password
}
