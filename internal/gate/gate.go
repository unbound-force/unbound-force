package gate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// gitCommandTimeout bounds each git subprocess invocation so a
// wedged git binary or hung filesystem cannot stall a headless
// gate run indefinitely.
const gitCommandTimeout = 10 * time.Second

// Options configures a gate run. External dependencies are
// injected for testability per Constitution IV.
type Options struct {
	// TargetDir is the project directory to check.
	TargetDir string

	// Phase is the workflow phase to check.
	Phase string

	// BuildVersion identifies the uf binary that produced the report.
	BuildVersion string

	// Format is the output format: "text" or "json". Reserved:
	// Run does not consume it; the CLI boundary owns formatting
	// (design D7).
	Format string

	// Stdout is reserved for future output; Run does not write
	// to it. The CLI boundary formats the report and writes it
	// through its own writer (design D7).
	Stdout io.Writer

	// Stderr is reserved for future diagnostics; Run does not
	// write to it. The CLI boundary routes error output to
	// stderr via Cobra (design D7).
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
	if o.BuildVersion == "" {
		o.BuildVersion = "dev"
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
// error for validation errors (exit code 1) or internal
// errors wrapped as *InternalError (exit code 2).
func Run(opts Options) (*GateReport, error) {
	opts.defaults()

	// Validate phase — user error, not internal (exit 1).
	if !isValidPhase(opts.Phase) {
		return nil, fmt.Errorf(
			"invalid phase %q: must be one of %s",
			opts.Phase,
			strings.Join(validPhases, ", "),
		)
	}

	// Canonicalize and validate target directory (GATE-FR-010).
	absDir, err := filepath.Abs(opts.TargetDir)
	if err != nil {
		return nil, &InternalError{
			Err: fmt.Errorf("resolve directory: %w", err),
		}
	}

	resolved, err := filepath.EvalSymlinks(absDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// User-supplied path does not exist — exit 1.
			return nil, fmt.Errorf("directory does not exist: %s", opts.TargetDir)
		}
		return nil, &InternalError{
			Err: fmt.Errorf("resolve symlinks: %w", err),
		}
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return nil, &InternalError{
			Err: fmt.Errorf("stat directory: %w", err),
		}
	}
	if !info.IsDir() {
		// User-supplied path is a file — exit 1.
		return nil, fmt.Errorf("path is not a directory: %s", opts.TargetDir)
	}

	// Resolve the active change once so every check sees a single
	// branch snapshot and provenance matches the resolution the
	// checks used.
	res := resolveChange(resolved)
	if res.ResolveErr != nil {
		return nil, &InternalError{
			Err: fmt.Errorf("resolve change: %w", res.ResolveErr),
		}
	}
	ctx := runContext{dir: resolved, res: res}

	// Execute checks for the phase.
	checks := phaseChecks[opts.Phase]
	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		result := check(ctx)
		if result.internalErr != nil {
			return nil, &InternalError{Err: result.internalErr}
		}
		results = append(results, result)
	}

	// Compute summary.
	summary := computeSummary(results)

	report := &GateReport{
		Version:     SchemaVersion,
		ToolVersion: opts.BuildVersion,
		Producer:    ProducerName,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Branch:      res.Branch,
		Commit:      detectCommit(resolved),
		Phase:       opts.Phase,
		Passed:      summary.Failed == 0,
		Checks:      results,
		Summary:     summary,
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

// errGitUnavailable reports that the git binary is not on PATH.
var errGitUnavailable = errors.New("git is not available — run 'uf doctor' to check tool availability")

// execGit runs git with the given arguments in dir under a bounded
// timeout with explicit argument separation (no shell invocation,
// Constitution V), returning trimmed stdout. Failures distinguish
// git being unavailable, timeouts, and command errors with stderr
// detail, so best-effort and strict callers share one contract.
func execGit(dir string, args ...string) (string, error) {
	gitCtx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(gitCtx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", gitError(gitCtx, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// gitError classifies a failed git invocation: git missing from
// PATH, a timeout, or the command's own failure with stderr detail.
func gitError(gitCtx context.Context, err error) error {
	if _, lookErr := exec.LookPath("git"); lookErr != nil {
		return errGitUnavailable
	}
	if gitCtx.Err() != nil {
		return fmt.Errorf("git did not respond within %s", gitCommandTimeout)
	}
	detail := err.Error()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
		detail = strings.TrimSpace(string(exitErr.Stderr))
	}
	return fmt.Errorf("git failed: %s", detail)
}

// detectBranch returns the current git branch name ("HEAD" for a
// detached HEAD) and a detailed error when detection fails: git
// unavailable, timeout, or not a git repository. The error is
// carried on the change resolution so checks that consume branch
// identity report the same failure the run observed (design D2).
func detectBranch(dir string) (string, error) {
	branch, err := execGit(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("cannot determine git branch: %w", err)
	}
	return branch, nil
}

// detectCommit returns the checked HEAD commit, or an empty string when git
// is unavailable, times out, or the directory is not a repository.
func detectCommit(dir string) string {
	commit, err := execGit(dir, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return commit
}
