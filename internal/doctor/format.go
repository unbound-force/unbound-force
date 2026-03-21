package doctor

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// FormatJSON serializes the Report as indented JSON per FR-017.
// Uses 2-space indentation. Severity values serialize as lowercase
// strings via MarshalJSON.
func FormatJSON(report *Report, w io.Writer) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	_, writeErr := w.Write(append(data, '\n'))
	return writeErr
}

// FormatText renders the Report as colored terminal output per
// FR-019 and contracts/cli-schema.md. Uses lipgloss for styling
// with automatic NO_COLOR and pipe detection.
func FormatText(report *Report, w io.Writer) error {
	renderer := lipgloss.NewRenderer(w)

	// Define styles per research.md R2.
	passStyle := renderer.NewStyle().Foreground(lipgloss.Color("2"))
	warnStyle := renderer.NewStyle().Foreground(lipgloss.Color("3"))
	failStyle := renderer.NewStyle().Foreground(lipgloss.Color("1"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle := renderer.NewStyle().Bold(true)

	// Detect color support — if no color, use plain indicators.
	hasColor := renderer.ColorProfile() != termenv.Ascii

	// Header.
	fmt.Fprintln(w, boldStyle.Render("Unbound Force Doctor"))
	fmt.Fprintln(w, "====================")
	fmt.Fprintln(w)

	for _, group := range report.Groups {
		fmt.Fprintln(w, boldStyle.Render(group.Name))

		for _, r := range group.Results {
			indicator := formatIndicator(r, hasColor, passStyle, warnStyle, failStyle, dimStyle)
			name := fmt.Sprintf("%-18s", r.Name)
			msg := r.Message
			if r.Detail != "" {
				msg += " (" + r.Detail + ")"
			}
			fmt.Fprintf(w, "  %s %s %s\n", indicator, name, msg)

			// Install hint on indented line below per cli-schema.md.
			if r.InstallHint != "" {
				fmt.Fprintf(w, "                     Install: %s\n", r.InstallHint)
			}
			if r.InstallURL != "" {
				fmt.Fprintf(w, "                     Docs: %s\n", r.InstallURL)
			}
		}

		// Render embedded output (e.g., swarm doctor) between separators.
		if group.Embed != "" {
			separator := strings.Repeat("─", 40)
			fmt.Fprintf(w, "  ── %s %s\n", group.Name+" ", separator[:max(0, 38-len(group.Name))])
			for _, line := range strings.Split(strings.TrimRight(group.Embed, "\n"), "\n") {
				fmt.Fprintf(w, "  %s\n", line)
			}
			fmt.Fprintf(w, "  %s\n", separator)
		}

		fmt.Fprintln(w)
	}

	// Summary line.
	summaryLine := fmt.Sprintf("Summary: %d passed, %d warnings, %d failed",
		report.Summary.Passed, report.Summary.Warned, report.Summary.Failed)
	fmt.Fprintln(w, summaryLine)

	return nil
}

// formatIndicator returns the appropriate symbol for a check result.
func formatIndicator(r CheckResult, hasColor bool, pass, warn, fail, dim lipgloss.Style) string {
	if !hasColor {
		// Plain text fallback per FR-019.
		switch r.Severity {
		case Pass:
			// Optional-absent items use INFO indicator per data-model.md.
			if r.Severity == Pass && r.InstallHint != "" {
				return "[INFO]"
			}
			return "[PASS]"
		case Warn:
			return "[WARN]"
		case Fail:
			return "[FAIL]"
		default:
			return "[????]"
		}
	}

	// Colored symbols per cli-schema.md.
	switch r.Severity {
	case Pass:
		// Optional-absent items show gray circle per data-model.md.
		if r.InstallHint != "" {
			return dim.Render("○")
		}
		return pass.Render("✓")
	case Warn:
		return warn.Render("!")
	case Fail:
		return fail.Render("✗")
	default:
		return "?"
	}
}
