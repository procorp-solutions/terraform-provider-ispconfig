resource "ispconfig_web_database_user" "example" {
  database_user     = "app_user"
  database_password = var.database_password
}

variable "database_password" {
  type      = string
  sensitive = true
}
