package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/gate"
)

// gateParams holds testable parameters for the gate command.
type gateParams struct {
	targetDir string
	phase     string
	format    string
	stdout    io.Writer
	stderr    io.Writer
}

// runGate executes the gate command with testable parameters.
// Returns error on check failure or internal error. main maps the
// returned error to an exit code through the exitCoder interface:
//
//	nil           → exit 0 (all checks pass)
//	plain error   → exit 1 (validation or check failure)
//	*gate.InternalError → exit 2 (implements exitCoder)
func runGate(p gateParams) error {
	opts := gate.Options{
		TargetDir:    p.targetDir,
		Phase:        p.phase,
		BuildVersion: version,
		Format:       p.format,
		Stdout:       p.stdout,
		Stderr:       p.stderr,
	}

	report, err := gate.Run(opts)

	// No report and error — gate.Run already wraps true
	// internal errors as *gate.InternalError (exit 2).
	// Validation errors are plain errors (exit 1).
	if report == nil && err != nil {
		return err
	}

	// Format and write output to stdout (GATE-FR-009, D7).
	if report != nil {
		switch p.format {
		case "json":
			if fmtErr := gate.FormatJSON(report, p.stdout); fmtErr != nil {
				return &gate.InternalError{
					Err: fmt.Errorf("format json: %w", fmtErr),
				}
			}
		default:
			if fmtErr := gate.FormatText(report, p.stdout); fmtErr != nil {
				return &gate.InternalError{
					Err: fmt.Errorf("format text: %w", fmtErr),
				}
			}
		}
	}

	// Check failure — caller maps to exit 1.
	return err
}

// newGateCmd creates the cobra command for the gate subcommand.
func newGateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Enforce constitution compliance at workflow phase boundaries",
		// Headless tool: never emit usage on the error path. Errors
		// go to stderr; the report (if any) is the only stdout
		// content, keeping --format json output parseable.
		SilenceUsage: true,
		Long: `Run phase-specific constitution compliance checks. Designed for
headless use in CI pipelines, validation loops, and pre-PR hooks.

Phases:
  specify    — spec artifact exists
  plan       — spec + plan/design artifacts exist
  implement  — spec + plan + tasks + coverage strategy exist
  review     — all tasks complete + review marker present
  pr         — code review passed + not on main branch

Exit codes:
  0  All checks pass
  1  One or more checks failed, or invalid arguments
  2  Internal error (unexpected filesystem failure)

Coverage strategy detection (implement phase):
  plan/tasks artifacts are scanned (case-insensitive) for a
  heading containing 'coverage' — negated headings such as 'No
  coverage strategy' do not count — body text containing
  'coverage target', 'unit test', 'integration test', or
  'e2e test', or a line matching coverage.*%.

Task completion (review phase):
  tasks.md lines matching '- [ ]' fail the check, including
  indented subtasks; prose mid-line and examples inside fenced
  code blocks do not count.

Artifact resolution:
  When the current branch names a change (opsx/<name> or
  NNN-<name>), checks resolve artifacts from that change only
  (openspec/changes/<name>/ or specs/<name>/). quick/<name>
  branches carry no spec artifacts, so their checks always fail.
  On a detached HEAD (the default CI checkout state), the branch
  is recovered from GITHUB_HEAD_REF or GITHUB_REF_NAME before
  falling back. On branches without a change mapping, checks
  resolve the first alphabetical change directory across
  openspec/changes/* or specs/* — one change for the whole run;
  artifacts of different changes are never combined.

Use --format json for machine-readable output with provenance
metadata (schema version, producer, tool version, timestamp, branch, commit).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			phase, _ := cmd.Flags().GetString("phase")
			format, _ := cmd.Flags().GetString("format")
			dir, _ := cmd.Flags().GetString("dir")

			// --phase is required (GATE-FR-002).
			if phase == "" {
				validList := strings.Join(gate.ValidPhases(), ", ")
				return fmt.Errorf(
					"--phase is required (valid phases: %s)",
					validList)
			}

			// Validate format flag (GATE-FR-003).
			if format != "text" && format != "json" {
				return fmt.Errorf(
					"invalid format %q: must be 'text' or 'json'",
					format)
			}

			// --dir passes through as-is: gate.Run owns the default
			// ("." resolves via filepath.Abs) so a Getwd failure is
			// classified as an internal error (exit 2) on every path
			// (GATE-FR-009), not exit 1 from a CLI-side pre-resolution.

			err := runGate(gateParams{
				targetDir: dir,
				phase:     phase,
				format:    format,
				stdout:    cmd.OutOrStdout(),
				stderr:    cmd.ErrOrStderr(),
			})

			return err
		},
	}

	cmd.Flags().String("phase", "", "Workflow phase to check (specify, plan, implement, review, pr)")
	cmd.Flags().String("format", "text", "Output format: text or json")
	cmd.Flags().String("dir", ".", "Project root directory to check")
	return cmd
}
