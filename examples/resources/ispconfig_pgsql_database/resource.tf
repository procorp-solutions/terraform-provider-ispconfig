resource "ispconfig_pgsql_database" "example" {
  database_name    = "myapp_db"
  parent_domain_id = 1
  active           = true
  quota            = 500
}
