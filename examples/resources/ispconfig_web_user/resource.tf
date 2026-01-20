resource "ispconfig_web_user" "example" {
  username         = "deploy"
  password         = var.shell_password
  parent_domain_id = 1
  shell            = "/bin/bash"
  quota_size       = 1000
  active           = "y"
}

variable "shell_password" {
  type      = string
  sensitive = true
}
