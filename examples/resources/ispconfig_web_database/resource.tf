resource "ispconfig_web_database" "example" {
  database_name    = "myapp_db"
  parent_domain_id = 1
  type             = "mysql"
  active           = "y"
  quota            = 500
}
