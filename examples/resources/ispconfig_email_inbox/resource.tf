resource "ispconfig_email_inbox" "example" {
  maildomain_id       = ispconfig_email_domain.example.id
  email               = "user@example.com"
  password            = var.mailbox_password
  quota               = 1024
  forward_incoming_to = "backup@example.com"
  forward_outgoing_to = "archive@example.com"
}
