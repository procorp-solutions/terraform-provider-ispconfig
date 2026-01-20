resource "ispconfig_web_hosting" "example" {
  domain      = "example.com"
  server_id   = 1
  ip_address  = "*"
  php         = "php-fpm"
  php_version = "8.2"
  active      = true
  ssl         = true
  hd_quota    = 10000
}
