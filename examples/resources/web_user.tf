# Create a shell user for the domain
resource "ispconfig_web_user" "example" {
  username         = "deploy"
  password         = var.shell_user_password
  parent_domain_id = ispconfig_web_hosting.example.id
  shell            = "/bin/bash"
  quota_size       = 1000 # 1GB in MB
  active           = "y"
}

# Shell user with restricted shell (no SSH access)
resource "ispconfig_web_user" "restricted" {
  username         = "ftponly"
  password         = var.ftp_user_password
  parent_domain_id = ispconfig_web_hosting.example.id
  shell            = "/bin/false" # No shell access
  quota_size       = 500          # 500MB quota
  active           = "y"
}

# Shell user with custom directory
resource "ispconfig_web_user" "custom_dir" {
  username         = "appuser"
  password         = var.app_user_password
  parent_domain_id = ispconfig_web_hosting.example.id
  dir              = "/var/www/clients/client1/web1/app"
  shell            = "/bin/bash"
  quota_size       = 2000 # 2GB in MB
  active           = "y"
}

# Output the shell user ID
output "shell_user_id" {
  value = ispconfig_web_user.example.id
}

# Variables (define in terraform.tfvars or via environment)
variable "shell_user_password" {
  type        = string
  description = "Password for the main shell user"
  sensitive   = true
}

variable "ftp_user_password" {
  type        = string
  description = "Password for the FTP-only user"
  sensitive   = true
}

variable "app_user_password" {
  type        = string
  description = "Password for the application user"
  sensitive   = true
}
