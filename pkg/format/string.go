package format

import "strings"

// NormalizeString makes val comparable regardless of case or whitespace
func NormalizeString(val string) string {
	return strings.TrimSpace(strings.ToLower(val))
}
