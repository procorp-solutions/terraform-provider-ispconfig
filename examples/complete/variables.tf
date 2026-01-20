# =============================================================================
# Provider Configuration Variables
# =============================================================================

variable "ispconfig_host" {
  type        = string
  description = "ISPConfig server host and port (e.g., 'ispconfig.example.com:8080')"
}

variable "ispconfig_username" {
  type        = string
  description = "ISPConfig API username"
  sensitive   = true
}

variable "ispconfig_password" {
  type        = string
  description = "ISPConfig API password"
  sensitive   = true
}

variable "ispconfig_insecure" {
  type        = bool
  description = "Skip TLS certificate verification (for self-signed certificates)"
  default     = false
}

variable "ispconfig_client_id" {
  type        = number
  description = "ISPConfig client ID"
  default     = 1
}

variable "ispconfig_server_id" {
  type        = number
  description = "ISPConfig server ID (typically 1 for single-server setups)"
  default     = 1
}

# =============================================================================
# Domain Configuration Variables
# =============================================================================

variable "domain_name" {
  type        = string
  description = "The primary domain name (e.g., 'example.com')"

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]*\\.[a-z]{2,}$", var.domain_name))
    error_message = "Domain name must be a valid domain (e.g., 'example.com')."
  }
}

variable "domain_ip" {
  type        = string
  description = "IP address for the domain ('*' for automatic assignment)"
  default     = "*"
}

variable "enable_ssl" {
  type        = bool
  description = "Enable SSL/TLS for the domain"
  default     = true
}

variable "php_version" {
  type        = string
  description = "PHP version to use"
  default     = "8.2"

  validation {
    condition     = contains(["7.4", "8.0", "8.1", "8.2", "8.3", "8.4"], var.php_version)
    error_message = "PHP version must be one of: 7.4, 8.0, 8.1, 8.2, 8.3, 8.4."
  }
}

variable "php_max_requests" {
  type        = number
  description = "PHP-FPM max requests per process"
  default     = 500
}

# =============================================================================
# Resource Quota Variables
# =============================================================================

variable "disk_quota_mb" {
  type        = number
  description = "Disk quota in MB"
  default     = 10000 # 10GB
}

variable "traffic_quota_mb" {
  type        = number
  description = "Traffic quota in MB"
  default     = 100000 # 100GB
}

variable "database_quota_mb" {
  type        = number
  description = "Database quota in MB"
  default     = 500
}

variable "shell_user_quota_mb" {
  type        = number
  description = "Shell user disk quota in MB"
  default     = 5000 # 5GB
}

# =============================================================================
# Credential Variables
# =============================================================================

variable "shell_user_password" {
  type        = string
  description = "Password for the shell/SFTP user"
  sensitive   = true
}

variable "database_password" {
  type        = string
  description = "Password for the database user"
  sensitive   = true
}

# =============================================================================
# Optional Features
# =============================================================================

variable "create_staging" {
  type        = bool
  description = "Create a staging subdomain and database"
  default     = false
}

variable "database_remote_access" {
  type        = bool
  description = "Enable remote database access"
  default     = false
}

variable "database_remote_ips" {
  type        = string
  description = "Allowed IP addresses/CIDR for remote database access"
  default     = ""
}
