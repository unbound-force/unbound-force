package textutil

import (
	"fmt"
	"strings"
	"testing"
)

func TestTruncateOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		maxLines int
		want     string
	}{
		{
			name:     "empty input",
			input:    []byte(""),
			maxLines: 20,
			want:     "",
		},
		{
			name:     "whitespace only",
			input:    []byte("  \n  \n  "),
			maxLines: 20,
			want:     "",
		},
		{
			name:     "short output 3 lines",
			input:    []byte("line1\nline2\nline3\n"),
			maxLines: 20,
			want:     "line1\nline2\nline3",
		},
		{
			name:     "exactly 20 lines",
			input:    []byte(generateLines(20)),
			maxLines: 20,
			want:     strings.TrimSpace(generateLines(20)),
		},
		{
			name:     "21 lines truncated to last 10",
			input:    []byte(generateLines(21)),
			maxLines: 20,
			want:     "... (11 lines omitted)\n" + generateLastNLines(21, 10),
		},
		{
			name:     "50 lines truncated to last 10",
			input:    []byte(generateLines(50)),
			maxLines: 20,
			want:     "... (40 lines omitted)\n" + generateLastNLines(50, 10),
		},
		{
			name:     "output with no trailing newline",
			input:    []byte("line1\nline2\nline3"),
			maxLines: 20,
			want:     "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateOutput(tt.input, tt.maxLines)
			if got != tt.want {
				t.Errorf("TruncateOutput() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

// generateLines creates a string with n lines like "line 1\nline 2\n...".
func generateLines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

// generateLastNLines returns the last n lines of a total-line string,
// joined by newlines (no trailing newline).
func generateLastNLines(total, n int) string {
	var lines []string
	for i := total - n + 1; i <= total; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	return strings.Join(lines, "\n")
}
