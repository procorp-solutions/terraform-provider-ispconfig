# Query an existing web hosting domain by ID
data "ispconfig_web_hosting" "existing" {
  id = 123
}

# Use the data to reference domain configuration
output "domain_name" {
  value       = data.ispconfig_web_hosting.existing.domain
  description = "The domain name of the existing site"
}

output "domain_ip" {
  value       = data.ispconfig_web_hosting.existing.ip_address
  description = "The IP address assigned to the domain"
}

output "php_version" {
  value       = data.ispconfig_web_hosting.existing.php_version
  description = "The PHP version configured for this domain"
}

output "ssl_enabled" {
  value       = data.ispconfig_web_hosting.existing.ssl
  description = "Whether SSL is enabled for this domain"
}

output "document_root" {
  value       = data.ispconfig_web_hosting.existing.document_root
  description = "The document root path for this domain"
}

# Example: Create a shell user for an existing domain
resource "ispconfig_web_user" "for_existing_domain" {
  username         = "admin"
  password         = var.admin_password
  parent_domain_id = data.ispconfig_web_hosting.existing.id
  shell            = "/bin/bash"
  active           = "y"
}

variable "admin_password" {
  type        = string
  description = "Password for the admin shell user"
  sensitive   = true
}
