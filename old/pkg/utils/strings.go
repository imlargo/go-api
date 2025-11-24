package utils

import "strings"

func NormalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
