package main

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/homebrew"
)

// homebrewCaskParams holds testable parameters for the Homebrew Cask command.
type homebrewCaskParams struct {
	inputPath  string
	outputPath string
	stdout     io.Writer
}

// runHomebrewCask transforms one generated Homebrew Cask hook.
func runHomebrewCask(params homebrewCaskParams) error {
	result, err := homebrew.Run(homebrew.Options{
		InputPath:  params.inputPath,
		OutputPath: params.outputPath,
	})
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(params.stdout, "Transformed Homebrew Cask: %s\n", result.OutputPath); err != nil {
		return fmt.Errorf("write command output: %w", err)
	}
	return nil
}

// newHomebrewCaskCmd returns the release-only Cask transformation command.
func newHomebrewCaskCmd() *cobra.Command {
	command := &cobra.Command{
		Use:    "transform-homebrew-cask",
		Short:  "Transform the generated Homebrew Cask postflight hook",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			inputPath, err := cmd.Flags().GetString("input")
			if err != nil {
				return fmt.Errorf("read input flag: %w", err)
			}
			outputPath, err := cmd.Flags().GetString("output")
			if err != nil {
				return fmt.Errorf("read output flag: %w", err)
			}

			return runHomebrewCask(homebrewCaskParams{
				inputPath:  inputPath,
				outputPath: outputPath,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}

	command.Flags().String("input", "", "Generated Cask input path")
	command.Flags().String("output", "", "Transformed Cask output path")
	return command
}
