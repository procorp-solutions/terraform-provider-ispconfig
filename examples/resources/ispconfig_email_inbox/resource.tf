resource "ispconfig_email_inbox" "example" {
  maildomain_id = ispconfig_email_domain.example.id
  email         = "user@example.com"
  password      = var.mailbox_password
  quota         = 1024
}
