package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/doctor"
	"github.com/unbound-force/unbound-force/internal/scaffold"
	"github.com/unbound-force/unbound-force/internal/setup"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:     "unbound-force",
		Short:   "Unbound Force specification framework toolkit (alias: uf)",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	}
	root.SetVersionTemplate("unbound-force version {{.Version}}\n")

	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newSetupCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

type initParams struct {
	targetDir   string
	force       bool
	divisorOnly bool
	lang        string
	version     string
	stdout      io.Writer
}

func runInit(p initParams) error {
	_, err := scaffold.Run(scaffold.Options{
		TargetDir:   p.targetDir,
		Force:       p.force,
		DivisorOnly: p.divisorOnly,
		Lang:        p.lang,
		Version:     p.version,
		Stdout:      p.stdout,
	})
	return err
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the unbound-force version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(),
				"unbound-force v%s (commit %s, built %s)\n",
				version, commit, date)
		},
	}
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold specification framework into current directory",
		Long: `Initialize the Unbound Force specification framework in the
current directory. This creates the Speckit templates, scripts,
OpenCode commands and agents, Divisor review personas and
convention packs, and OpenSpec schema files needed for both
strategic and tactical specification workflows.

User-owned files (templates, scripts, agents, config) are
skipped if they already exist. Tool-owned files (speckit
commands, OpenSpec schema, convention packs) are updated if
their content has changed.

Use --divisor to deploy only The Divisor review agents,
the /review-council command, and convention packs.

Use --lang to specify the project language for convention
pack selection (auto-detected from go.mod, package.json,
etc. if not provided).

Use --force to overwrite all files regardless of ownership.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			divisorOnly, _ := cmd.Flags().GetBool("divisor")
			lang, _ := cmd.Flags().GetString("lang")
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runInit(initParams{
				targetDir:   cwd,
				force:       force,
				divisorOnly: divisorOnly,
				lang:        lang,
				version:     version,
				stdout:      cmd.OutOrStdout(),
			})
		},
	}
	cmd.Flags().Bool("force", false, "Overwrite all existing files")
	cmd.Flags().Bool("divisor", false, "Deploy only Divisor review agents and convention packs")
	cmd.Flags().String("lang", "", "Project language for convention pack (auto-detected if omitted)")
	return cmd
}

// doctorParams holds testable parameters for the doctor command.
type doctorParams struct {
	targetDir string
	format    string
	stdout    io.Writer
}

// runDoctor executes the doctor command with testable parameters.
func runDoctor(p doctorParams) error {
	opts := doctor.Options{
		TargetDir: p.targetDir,
		Format:    p.format,
		Stdout:    p.stdout,
	}

	report, err := doctor.Run(opts)
	if report != nil {
		switch p.format {
		case "json":
			if fmtErr := doctor.FormatJSON(report, p.stdout); fmtErr != nil {
				return fmt.Errorf("format json: %w", fmtErr)
			}
		default:
			if fmtErr := doctor.FormatText(report, p.stdout); fmtErr != nil {
				return fmt.Errorf("format text: %w", fmtErr)
			}
		}
	}

	return err
}

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the Unbound Force development environment",
		Long: `Check for required tools, version managers, scaffolded files,
hero availability, Swarm plugin status, MCP server configuration,
and agent/skill integrity. Produces a colored terminal report by
default, or structured JSON for CI pipelines.

Exit code 0 when all checks pass or only warnings exist.
Exit code 1 when any check fails.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			format, _ := cmd.Flags().GetString("format")
			dir, _ := cmd.Flags().GetString("dir")

			// Validate format flag.
			if format != "text" && format != "json" {
				return fmt.Errorf("invalid format %q: must be 'text' or 'json'", format)
			}

			if dir == "" || dir == "." {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				dir = cwd
			}

			return runDoctor(doctorParams{
				targetDir: dir,
				format:    format,
				stdout:    cmd.OutOrStdout(),
			})
		},
	}
	cmd.Flags().String("format", "text", "Output format: text or json")
	cmd.Flags().String("dir", ".", "Target directory to check")
	return cmd
}

// setupParams holds testable parameters for the setup command.
type setupParams struct {
	targetDir string
	dryRun    bool
	yesFlag   bool
	stdout    io.Writer
	stderr    io.Writer
}

// runSetup executes the setup command with testable parameters.
func runSetup(p setupParams) error {
	opts := setup.Options{
		TargetDir: p.targetDir,
		DryRun:    p.dryRun,
		YesFlag:   p.yesFlag,
		Stdout:    p.stdout,
		Stderr:    p.stderr,
	}

	return setup.Run(opts)
}

func newSetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Install and configure the Unbound Force development tool chain",
		Long: `Detect existing version and package managers, install missing
tools through the appropriate manager, configure the Swarm plugin
in opencode.json, and scaffold project files. Idempotent -- safe
to run multiple times.

Use --dry-run to preview actions without executing.
Use --yes to skip confirmation prompts for curl|bash installs.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			yesFlag, _ := cmd.Flags().GetBool("yes")
			dir, _ := cmd.Flags().GetString("dir")

			if dir == "" || dir == "." {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("get working directory: %w", err)
				}
				dir = cwd
			}

			return runSetup(setupParams{
				targetDir: dir,
				dryRun:    dryRun,
				yesFlag:   yesFlag,
				stdout:    cmd.OutOrStdout(),
				stderr:    cmd.ErrOrStderr(),
			})
		},
	}
	cmd.Flags().String("dir", ".", "Target directory for setup")
	cmd.Flags().Bool("dry-run", false, "Print actions without executing")
	cmd.Flags().Bool("yes", false, "Skip confirmation prompts")
	return cmd
}
