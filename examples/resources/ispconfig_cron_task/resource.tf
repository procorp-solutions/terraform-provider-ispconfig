resource "ispconfig_cron_task" "example" {
  parent_domain_id = 1
  schedule         = "*/5 * * * *"
  command          = "https://example.com/cron-endpoint"
  type             = "url"
  active           = true
}
