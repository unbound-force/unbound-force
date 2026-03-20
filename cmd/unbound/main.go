package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/scaffold"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	root := &cobra.Command{
		Use:     "unbound",
		Short:   "Unbound Force specification framework toolkit",
		Version: fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
	}
	root.SetVersionTemplate("unbound version {{.Version}}\n")

	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())

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
		Short: "Print the unbound version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(),
				"unbound v%s (commit %s, built %s)\n",
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
