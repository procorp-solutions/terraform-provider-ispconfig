package provider

// mbToAPIQuota converts a mailbox quota from MB (as provided by the user) to the
// unit expected by the ISPConfig API (kB, where 1 kB = 1000 bytes).
// Special values -1 (unlimited) and 0 (no mail) are passed through unchanged.
func mbToAPIQuota(mb int64) int64 {
	if mb < 0 {
		return mb
	}
	return mb * 1024 * 1024 / 1000
}

// apiQuotaToMB converts a mailbox quota returned by the ISPConfig API (kB, where
// 1 kB = 1000 bytes) back to MB for storage in Terraform state.
// Special values -1 (unlimited) and 0 (no mail) are passed through unchanged.
func apiQuotaToMB(kb int64) int64 {
	if kb < 0 {
		return kb
	}
	return kb * 1000 / (1024 * 1024)
}
