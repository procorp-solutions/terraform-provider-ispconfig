# Basic web hosting domain with PHP 8.2
resource "ispconfig_web_hosting" "example" {
  domain      = "example.com"
  server_id   = 1 # Required: typically 1 for single-server setups
  ip_address  = "*"
  php         = "php-fpm"
  php_version = "8.2"
  active      = true
  ssl         = true

  # Optional: resource quotas
  hd_quota      = 10000  # 10GB in MB
  traffic_quota = 100000 # 100GB in MB

  # Optional: PHP-FPM settings
  pm              = "dynamic"
  pm_max_requests = 500
}

# Web hosting with custom document root subdirectory
# Useful when deploying frameworks that use a public/ directory
resource "ispconfig_web_hosting" "laravel_app" {
  domain      = "app.example.com"
  server_id   = 1
  php         = "php-fpm"
  php_version = "8.2"
  active      = true
  ssl         = true

  # ISPConfig will generate base path like /var/www/clients/client2/web10
  # and this will append web/public to create /var/www/clients/client2/web10/web/public
  root_subdir = "web/public"
}

# Subdomain linked to parent domain
resource "ispconfig_web_hosting" "blog" {
  domain           = "blog.example.com"
  server_id        = 1
  type             = "subdomain"
  parent_domain_id = ispconfig_web_hosting.example.id
  php              = "php-fpm"
  php_version      = "8.1"
  active           = true
  ssl              = true
}

# API domain without www subdomain
resource "ispconfig_web_hosting" "api" {
  domain    = "api.example.com"
  server_id = 1
  subdomain = "none" # Disable www subdomain (only api.example.com will work)
  php       = "php-fpm"
  active    = true
  ssl       = true
}

# Static site without PHP
resource "ispconfig_web_hosting" "static" {
  domain    = "static.example.com"
  server_id = 1
  php       = "no" # Disable PHP
  active    = true
  ssl       = true
  subdomain = "*" # Wildcard subdomain support
}

# Domain with CGI and Perl enabled (legacy applications)
resource "ispconfig_web_hosting" "legacy" {
  domain    = "legacy.example.com"
  server_id = 1
  active    = true
  ssl       = false
  cgi       = true
  perl      = true
  ssi       = true
}

# Development domain with relaxed settings
resource "ispconfig_web_hosting" "dev" {
  domain      = "dev.example.com"
  server_id   = 1
  php         = "php-fpm"
  php_version = "8.3"
  active      = true
  ssl         = true
  suexec      = false # Disable suexec for development

  # Lower quotas for development
  hd_quota      = 2000  # 2GB
  traffic_quota = 10000 # 10GB
}

# Output the domain IDs
output "main_domain_id" {
  value = ispconfig_web_hosting.example.id
}

output "app_domain_id" {
  value = ispconfig_web_hosting.laravel_app.id
}

output "blog_domain_id" {
  value = ispconfig_web_hosting.blog.id
}
