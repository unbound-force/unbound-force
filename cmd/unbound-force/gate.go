package main

import (
	"errors"
	"fmt"
	"io"
	"os"
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
// Returns error on check failure or internal error. Callers
// inspect the error type to determine exit code:
//
//	nil         → exit 0 (all checks pass)
//	gate error  → exit 1 (check failure)
//	other error → exit 2 (internal error)
func runGate(p gateParams) error {
	opts := gate.Options{
		TargetDir: p.targetDir,
		Phase:     p.phase,
		Format:    p.format,
		Stdout:    p.stdout,
		Stderr:    p.stderr,
	}

	report, err := gate.Run(opts)

	// Internal error (no report) — caller maps to exit 2.
	if report == nil && err != nil {
		return &gate.InternalError{Err: err}
	}

	// Format and write output to stdout (FR-009, D7).
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

// gateExitCode returns the appropriate exit code for a gate error.
// Exit 0 = pass, 1 = check failure, 2 = internal error.
func gateExitCode(err error) int {
	if err == nil {
		return 0
	}
	var internal *gate.InternalError
	if errors.As(err, &internal) {
		return 2
	}
	return 1
}

// newGateCmd creates the cobra command for the gate subcommand.
func newGateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gate",
		Short: "Enforce constitution compliance at workflow phase boundaries",
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
  1  One or more checks failed
  2  Internal error (invalid arguments, filesystem error)

Use --format json for machine-readable output with provenance
metadata (version, producer, timestamp, branch).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			phase, _ := cmd.Flags().GetString("phase")
			format, _ := cmd.Flags().GetString("format")
			dir, _ := cmd.Flags().GetString("dir")

			// --phase is required (FR-002).
			if phase == "" {
				validList := strings.Join(gate.ValidPhases(), ", ")
				return fmt.Errorf(
					"--phase is required (valid phases: %s)",
					validList)
			}

			// Validate format flag (FR-003).
			if format != "text" && format != "json" {
				return fmt.Errorf(
					"invalid format %q: must be 'text' or 'json'",
					format)
			}

			if dir == "" || dir == "." {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				dir = cwd
			}

			err := runGate(gateParams{
				targetDir: dir,
				phase:     phase,
				format:    format,
				stdout:    cmd.OutOrStdout(),
				stderr:    cmd.ErrOrStderr(),
			})

			if gateExitCode(err) == 2 {
				// Cobra prints the error; use os.Exit for
				// exit code 2 since Cobra always returns 1.
				fmt.Fprintf(cmd.ErrOrStderr(),
					"Error: %v\n", err)
				os.Exit(2)
			}

			return err
		},
	}

	cmd.Flags().String("phase", "", "Workflow phase to check (specify, plan, implement, review, pr)")
	cmd.Flags().String("format", "text", "Output format: text or json")
	cmd.Flags().String("dir", ".", "Project root directory to check")
	return cmd
}
