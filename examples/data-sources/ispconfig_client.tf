# Query ISP Config client (customer) information
data "ispconfig_client" "current" {
  id = 1
}

# Output client details
output "client_company" {
  value       = data.ispconfig_client.current.company_name
  description = "The company name of the client"
}

output "client_contact" {
  value       = data.ispconfig_client.current.contact_name
  description = "The contact name for the client"
}

output "client_email" {
  value       = data.ispconfig_client.current.email
  description = "The email address of the client"
}

# Output resource limits
output "web_domain_limit" {
  value       = data.ispconfig_client.current.limit_web
  description = "Maximum number of web domains allowed"
}

output "database_limit" {
  value       = data.ispconfig_client.current.limit_database
  description = "Maximum number of databases allowed"
}

output "ftp_user_limit" {
  value       = data.ispconfig_client.current.limit_ftp_user
  description = "Maximum number of FTP users allowed"
}

output "shell_user_limit" {
  value       = data.ispconfig_client.current.limit_shell_user
  description = "Maximum number of shell users allowed"
}

# Example: Use client data for conditional resource creation
# This demonstrates checking limits before creating resources
locals {
  can_create_domain = data.ispconfig_client.current.limit_web > 0
}

output "can_create_domain" {
  value       = local.can_create_domain
  description = "Whether the client can create more web domains"
}
