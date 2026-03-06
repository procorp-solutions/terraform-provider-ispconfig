package provider

// mbToAPIQuota converts a mailbox quota from MB (as provided by the user) to
// bytes, as expected by the ISPConfig API.
// Special values -1 (unlimited) and 0 (no mail) are passed through unchanged.
func mbToAPIQuota(mb int64) int64 {
	if mb < 0 {
		return mb
	}
	return mb * 1024 * 1024
}

// apiQuotaToMB converts a mailbox quota returned by the ISPConfig API (bytes)
// back to MB for storage in Terraform state.
// Special values -1 (unlimited) and 0 (no mail) are passed through unchanged.
func apiQuotaToMB(bytes int64) int64 {
	if bytes < 0 {
		return bytes
	}
	return bytes / (1024 * 1024)
}
