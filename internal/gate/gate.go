package gate

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Options configures a gate run. External dependencies are
// injected for testability per Constitution IV.
type Options struct {
	// TargetDir is the project directory to check.
	TargetDir string

	// Phase is the workflow phase to check.
	Phase string

	// Format is the output format: "text" or "json".
	Format string

	// Stdout is the writer for check results.
	Stdout io.Writer

	// Stderr is the writer for validation and internal errors.
	Stderr io.Writer
}

// defaults fills zero-value fields with production defaults.
func (o *Options) defaults() {
	if o.TargetDir == "" {
		// Error is intentionally discarded: if Getwd fails,
		// TargetDir stays empty and filepath.Abs will surface
		// the error with a "resolve directory" context.
		o.TargetDir, _ = os.Getwd()
	}
	if o.Format == "" {
		o.Format = "text"
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
}

// Run executes the gate checks for the requested phase and
// returns the report. Returns an error with the report when
// any check fails (exit code 1). Returns a nil report with
// error for internal errors (exit code 2).
func Run(opts Options) (*GateReport, error) {
	opts.defaults()

	// Validate phase.
	if !isValidPhase(opts.Phase) {
		return nil, fmt.Errorf(
			"invalid phase %q: must be one of %s",
			opts.Phase,
			strings.Join(validPhases, ", "),
		)
	}

	// Canonicalize and validate target directory (FR-010).
	absDir, err := filepath.Abs(opts.TargetDir)
	if err != nil {
		return nil, fmt.Errorf("resolve directory: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", opts.TargetDir)
		}
		return nil, fmt.Errorf("resolve symlinks: %w", err)
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", opts.TargetDir)
	}

	// Execute checks for the phase.
	checks := phaseChecks[opts.Phase]
	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		results = append(results, check(resolved))
	}

	// Compute summary.
	summary := computeSummary(results)

	// Populate provenance (Constitution III).
	branch := detectBranch(resolved)

	report := &GateReport{
		Version:   SchemaVersion,
		Producer:  ProducerName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Branch:    branch,
		Phase:     opts.Phase,
		Passed:    summary.Failed == 0,
		Checks:    results,
		Summary:   summary,
	}

	if summary.Failed > 0 {
		return report, fmt.Errorf("%d check(s) failed", summary.Failed)
	}

	return report, nil
}

// isValidPhase checks if the given phase is in the valid set.
func isValidPhase(phase string) bool {
	for _, p := range validPhases {
		if p == phase {
			return true
		}
	}
	return false
}

// computeSummary aggregates check result counts.
func computeSummary(results []CheckResult) Summary {
	var s Summary
	for _, r := range results {
		s.Total++
		if r.Passed {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}

// detectBranch returns the current git branch name, or an empty
// string if git is unavailable or the directory is not a git
// repository. Best-effort per D4.
func detectBranch(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
