package utils

import "strings"

// ContainsIgnoreCase returns true if substr exists in s ignoring case.
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
