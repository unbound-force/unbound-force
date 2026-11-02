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
func FormatText(report *GateReport, w io.Writer) (err error) {
	renderer := lipgloss.NewRenderer(w)

	passStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	failStyle := renderer.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle := renderer.NewStyle().Foreground(lipgloss.Color("241"))
	titleStyle := renderer.NewStyle().Bold(true)

	hasColor := renderer.ColorProfile() != termenv.Ascii

	// write records the first write error so a failing sink is not
	// silently ignored, matching FormatJSON's error contract.
	failed := false
	write := func(f string, args ...any) {
		if failed {
			return
		}
		if _, e := fmt.Fprintf(w, f, args...); e != nil && err == nil {
			err = e
			failed = true
		}
	}

	// Phase header.
	write("%s\n\n", titleStyle.Render(fmt.Sprintf("Gate: %s", report.Phase)))

	// Individual check results.
	for _, r := range report.Checks {
		indicator := gateIndicator(r, hasColor, passStyle, failStyle)
		line := fmt.Sprintf("  %s %s", indicator, r.Name)
		if !r.Passed && r.Message != "" {
			line += " — " + r.Message
		}
		write("%s\n", line)
	}

	write("\n")

	// Summary line.
	if report.Passed {
		summary := fmt.Sprintf("PASS: %d of %d checks passed",
			report.Summary.Passed, report.Summary.Total)
		if hasColor {
			write("%s\n", passStyle.Render(summary))
		} else {
			write("%s\n", summary)
		}
	} else {
		summary := fmt.Sprintf("FAIL: %d of %d checks failed",
			report.Summary.Failed, report.Summary.Total)
		if hasColor {
			write("%s\n", failStyle.Render(summary))
		} else {
			write("%s\n", summary)
		}
	}

	// Provenance in dim.
	if report.Branch != "" {
		write("%s\n", dimStyle.Render(
			fmt.Sprintf("  branch: %s  |  %s v%s",
				report.Branch, report.Producer, report.Version)))
	}

	return err
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
