package statusline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// FormatMemory converts kilobytes to a human-readable string.
// Returns "0M" for zero/negative/invalid, "NM" for megabytes, "N.NG" for gigabytes.
// Matches the behavior of format_memory() in lib/statusline.sh.
func FormatMemory(kb string) string {
	kbVal, err := strconv.ParseInt(strings.TrimSpace(kb), 10, 64)
	if err != nil || kbVal <= 0 {
		return "0M"
	}

	mb := kbVal / 1024
	if mb >= 1024 {
		// Use one decimal place, matching bash: echo "scale=1; $mb / 1024" | bc
		// bc truncates (floors) rather than rounds.
		gbTenths := mb * 10 / 1024
		whole := gbTenths / 10
		frac := gbTenths % 10
		return fmt.Sprintf("%d.%dG", whole, frac)
	}
	return fmt.Sprintf("%dM", mb)
}

// parseCWDRegex matches "current_dir":"<value>" using the same logic as the
// bash sed: sed -n 's/.*"current_dir":"\([^"]*\)".*/\1/p'
// The Go regex is greedy by default, so .* will match as far as possible,
// giving us the LAST occurrence when multiple "current_dir" keys exist.
var parseCWDRegex = regexp.MustCompile(`.*"current_dir":"([^"]*)"`)

// ParseCWDFromJSON extracts the "current_dir" value from a JSON string.
// Uses regex pattern matching (like the bash version with sed).
// Returns empty string if not found or on error.
func ParseCWDFromJSON(jsonStr string) string {
	// Replace \r\n with \n to handle Windows-style line endings,
	// then collapse to single line for regex matching (matching sed behavior on piped input).
	normalized := strings.ReplaceAll(jsonStr, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\n", "")

	matches := parseCWDRegex.FindStringSubmatch(normalized)
	if matches == nil {
		return ""
	}
	return matches[1]
}
