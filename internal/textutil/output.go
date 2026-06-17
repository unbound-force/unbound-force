// Package textutil provides shared text formatting helpers for
// CLI output across the unbound-force toolchain.
package textutil

import (
	"fmt"
	"strings"
)

// TruncateOutput returns a string representation of command output,
// truncated to the last maxLines/2 lines if the output exceeds
// maxLines lines. Empty output returns an empty string.
func TruncateOutput(output []byte, maxLines int) string {
	s := strings.TrimSpace(string(output))
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	tail := maxLines / 2
	omitted := len(lines) - tail
	return fmt.Sprintf("... (%d lines omitted)\n%s", omitted, strings.Join(lines[len(lines)-tail:], "\n"))
}
