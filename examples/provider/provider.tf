terraform {
  required_version = ">= 1.0"

  required_providers {
    ispconfig = {
      source  = "procorp-solutions/ispconfig"
      version = "~> 1.0"
    }
  }
}

# Configure the ISPConfig provider
# Credentials can be provided via:
# 1. Provider configuration (shown below with variables)
# 2. Environment variables (ISPCONFIG_HOST, ISPCONFIG_USERNAME, ISPCONFIG_PASSWORD)
provider "ispconfig" {
  # Required: ISPConfig server host and port
  host = var.ispconfig_host

  # Required: Authentication credentials
  username = var.ispconfig_username
  password = var.ispconfig_password

  # Optional: Skip TLS certificate verification (for self-signed certificates)
  # insecure = true

  # Optional: Default client ID for all resources
  # client_id = 1

  # Optional: Default server ID for all resources (typically 1 for single-server setups)
  # server_id = 1
}

# Input variables for provider configuration
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
