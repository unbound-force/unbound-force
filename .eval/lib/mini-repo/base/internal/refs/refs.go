// Package refs holds helpers for working with git references.
package refs

import "strings"

// Normalize trims surrounding whitespace and a leading "refs/heads/"
// prefix so branch names compare equal however they were written.
func Normalize(ref string) string {
	return strings.TrimPrefix(strings.TrimSpace(ref), "refs/heads/")
}
