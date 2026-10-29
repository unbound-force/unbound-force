package gate

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// FormatJSON serializes the GateReport as indented JSON per
// FR-003. Uses 2-space indentation following the doctor pattern.
func FormatJSON(report *GateReport, w io.Writer) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	_, writeErr := w.Write(append(data, '\n'))
	return writeErr
}

// FormatText renders the GateReport as colored terminal output
// per D6. Uses lipgloss for styling with automatic NO_COLOR
// and pipe detection.
func FormatText(report *GateReport, w io.Writer) error {
	renderer := lipgloss.NewRenderer(w)

	passStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	failStyle := renderer.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle := renderer.NewStyle().Bold(true)

	hasColor := renderer.ColorProfile() != termenv.Ascii

	// Phase header.
	_, _ = fmt.Fprintf(w, "%s\n\n", titleStyle.Render(fmt.Sprintf("Gate: %s", report.Phase)))

	// Individual check results.
	for _, r := range report.Checks {
		indicator := gateIndicator(r, hasColor, passStyle, failStyle)
		line := fmt.Sprintf("  %s %s", indicator, r.Name)
		if !r.Passed && r.Message != "" {
			line += " — " + r.Message
		}
		_, _ = fmt.Fprintln(w, line)
	}

	_, _ = fmt.Fprintln(w)

	// Summary line.
	if report.Passed {
		summary := fmt.Sprintf("PASS: %d of %d checks passed",
			report.Summary.Passed, report.Summary.Total)
		if hasColor {
			_, _ = fmt.Fprintln(w, passStyle.Render(summary))
		} else {
			_, _ = fmt.Fprintln(w, summary)
		}
	} else {
		summary := fmt.Sprintf("FAIL: %d of %d checks failed",
			report.Summary.Failed, report.Summary.Total)
		if hasColor {
			_, _ = fmt.Fprintln(w, failStyle.Render(summary))
		} else {
			_, _ = fmt.Fprintln(w, summary)
		}
	}

	// Provenance in dim.
	if report.Branch != "" {
		_, _ = fmt.Fprintln(w, dimStyle.Render(
			fmt.Sprintf("  branch: %s  |  %s v%s",
				report.Branch, report.Producer, report.Version)))
	}

	return nil
}

// gateIndicator returns the pass/fail symbol for a check result.
func gateIndicator(r CheckResult, hasColor bool, pass, fail lipgloss.Style) string {
	if !hasColor {
		if r.Passed {
			return "[PASS]"
		}
		return "[FAIL]"
	}
	if r.Passed {
		return pass.Render("✓")
	}
	return fail.Render("✗")
}
