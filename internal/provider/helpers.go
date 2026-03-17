package provider

import (
	"fmt"
	"strings"
)

// boolToYN converts a Go bool to the "y"/"n" string expected by the ISPConfig API.
func boolToYN(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

// ynToBool converts an ISPConfig "y"/"n" string to a Go bool.
func ynToBool(s string) bool {
	return s == "y" || s == "Y"
}

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

// parseCronSchedule splits a cron schedule string into its 5 components.
func parseCronSchedule(schedule string) (runMin, runHour, runMday, runMonth, runWday string, err error) {
	parts := strings.Fields(schedule)
	if len(parts) != 5 {
		return "", "", "", "", "", fmt.Errorf("schedule must have exactly 5 fields (got %d): %q", len(parts), schedule)
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4], nil
}

// buildCronSchedule reconstructs the cron schedule string from API fields.
func buildCronSchedule(runMin, runHour, runMday, runMonth, runWday string) string {
	return strings.Join([]string{runMin, runHour, runMday, runMonth, runWday}, " ")
}
