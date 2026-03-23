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

	// Bright, distinctive colors for visual scanning
	// (gcal-organizer style).
	passStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	warnStyle := renderer.NewStyle().Foreground(lipgloss.Color("11"))
	failStyle := renderer.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle := renderer.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	boxStyle := renderer.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1)

	// Detect color support — if no color, use plain indicators.
	hasColor := renderer.ColorProfile() != termenv.Ascii

	// Header with stethoscope emoji.
	fmt.Fprintln(w, titleStyle.Render("🩺 Unbound Force Doctor"))
	fmt.Fprintln(w)

	boldStyle := renderer.NewStyle().Bold(true)

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

			// Fix hint on indented line below in subtle gray.
			if r.InstallHint != "" {
				fmt.Fprintln(w, dimStyle.Render("     Fix: "+r.InstallHint))
			}
			if r.InstallURL != "" {
				fmt.Fprintln(w, dimStyle.Render("     Docs: "+r.InstallURL))
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

	// Boxed summary with emoji counters.
	summaryContent := fmt.Sprintf("  ✅ %d passed  ⚠️  %d warnings  ❌ %d failed",
		report.Summary.Passed, report.Summary.Warned, report.Summary.Failed)
	fmt.Fprintln(w, boxStyle.Render(summaryContent))

	// Contextual completion message.
	if report.Summary.Failed == 0 && report.Summary.Warned == 0 {
		fmt.Fprintln(w, passStyle.Render("🎉 Everything looks good!"))
	} else if report.Summary.Failed > 0 {
		fmt.Fprintln(w, dimStyle.Render("  Run 'uf doctor' after fixes."))
	} else {
		fmt.Fprintln(w, dimStyle.Render("  All critical checks passed."))
	}

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

	// Emoji indicators for visual scanning (gcal-organizer style).
	switch r.Severity {
	case Pass:
		// Optional-absent items show gray ⊘ per data-model.md.
		if r.InstallHint != "" {
			return dim.Render("⊘")
		}
		return pass.Render("✅")
	case Warn:
		return warn.Render("⚠️ ")
	case Fail:
		return fail.Render("❌")
	default:
		return "?"
	}
}
