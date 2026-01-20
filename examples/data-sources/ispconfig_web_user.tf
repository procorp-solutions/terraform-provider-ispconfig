# Query an existing shell user by ID
data "ispconfig_web_user" "existing" {
  id = 456
}

# Output shell user details
output "shell_username" {
  value       = data.ispconfig_web_user.existing.username
  description = "The username of the shell user"
}

output "shell_directory" {
  value       = data.ispconfig_web_user.existing.dir
  description = "The home directory of the shell user"
}

output "shell_type" {
  value       = data.ispconfig_web_user.existing.shell
  description = "The shell assigned to the user"
}

output "shell_active" {
  value       = data.ispconfig_web_user.existing.active
  description = "Whether the shell user is active"
}

output "parent_domain" {
  value       = data.ispconfig_web_user.existing.parent_domain_id
  description = "The parent domain ID for this shell user"
}
