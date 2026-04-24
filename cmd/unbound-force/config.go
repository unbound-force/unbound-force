package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	goyaml "github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/config"
)

// --- Init subcommand ---

type configInitParams struct {
	targetDir string
	stdout    io.Writer
}

func runConfigInit(p configInitParams) error {
	result, err := config.InitFile(config.InitOptions{
		ProjectDir: p.targetDir,
	})
	if err != nil {
		return err
	}

	if result.Created {
		fmt.Fprintf(p.stdout, "Created %s\n", result.Path)
		fmt.Fprintln(p.stdout, "All values are commented out — uncomment what you want to change.")
		return nil
	}

	if result.Updated {
		fmt.Fprintf(p.stdout, "Updated %s\n", result.Path)
		if len(result.SectionsAdded) > 0 {
			fmt.Fprintf(p.stdout, "  Added sections: %v\n", result.SectionsAdded)
		}
		if len(result.SectionsRemoved) > 0 {
			fmt.Fprintf(p.stdout, "  Removed deprecated sections: %v\n", result.SectionsRemoved)
		}
		fmt.Fprintln(p.stdout, "A backup was saved to .uf/config.yaml.bak")
		return nil
	}

	fmt.Fprintf(p.stdout, "%s is already up to date.\n", result.Path)
	return nil
}

// --- Show subcommand ---

type configShowParams struct {
	targetDir string
	format    string
	stdout    io.Writer
}

func runConfigShow(p configShowParams) error {
	cfg, err := config.Load(config.LoadOptions{
		ProjectDir: p.targetDir,
	})
	if err != nil {
		return err
	}

	switch p.format {
	case "json":
		data, jsonErr := json.MarshalIndent(cfg, "", "  ")
		if jsonErr != nil {
			return fmt.Errorf("marshal JSON: %w", jsonErr)
		}
		fmt.Fprintln(p.stdout, string(data))
	default:
		data, yamlErr := goyaml.Marshal(cfg)
		if yamlErr != nil {
			return fmt.Errorf("marshal YAML: %w", yamlErr)
		}
		fmt.Fprint(p.stdout, string(data))
	}
	return nil
}

// --- Validate subcommand ---

type configValidateParams struct {
	targetDir string
	format    string
	stdout    io.Writer
}

func runConfigValidate(p configValidateParams) error {
	configPath := config.RepoConfigPath(p.targetDir)
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Missing file is valid — defaults are used.
		fmt.Fprintln(p.stdout, "No config file found at", configPath)
		fmt.Fprintln(p.stdout, "This is valid — compiled defaults are used.")
		return nil
	}

	// Try to parse the YAML.
	var cfg config.Config
	if parseErr := goyaml.Unmarshal(data, &cfg); parseErr != nil {
		fmt.Fprintf(p.stdout, "FAIL: %s is not valid YAML\n", configPath)
		fmt.Fprintf(p.stdout, "  Error: %v\n", parseErr)
		return fmt.Errorf("config validation failed")
	}

	// Validate known field values.
	errors := config.Validate(cfg)

	if p.format == "json" {
		type result struct {
			Path   string   `json:"path"`
			Valid  bool     `json:"valid"`
			Errors []string `json:"errors,omitempty"`
		}
		r := result{Path: configPath, Valid: len(errors) == 0, Errors: errors}
		data, _ := json.MarshalIndent(r, "", "  ")
		fmt.Fprintln(p.stdout, string(data))
		if len(errors) > 0 {
			return fmt.Errorf("config validation failed")
		}
		return nil
	}

	if len(errors) > 0 {
		fmt.Fprintf(p.stdout, "Config validation: %s\n\n", configPath)
		for _, e := range errors {
			fmt.Fprintf(p.stdout, "  FAIL: %s\n", e)
		}
		fmt.Fprintf(p.stdout, "\n%d error(s) found\n", len(errors))
		return fmt.Errorf("config validation failed")
	}

	fmt.Fprintf(p.stdout, "Config validation: %s\n", configPath)
	fmt.Fprintln(p.stdout, "  All checks passed.")
	return nil
}

// --- Command factory ---

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Unbound Force configuration",
		Long: `Manage the unified .uf/config.yaml configuration file.

Subcommands:
  init      Create or update the config file
  show      Display effective config after all layers merge
  validate  Validate config against known field values`,
	}

	// --- init subcommand ---
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create or update .uf/config.yaml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			return runConfigInit(configInitParams{
				targetDir: dir,
				stdout:    cmd.OutOrStdout(),
			})
		},
	}
	initCmd.Flags().String("dir", ".", "Target directory")

	// --- show subcommand ---
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Display effective configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			format, _ := cmd.Flags().GetString("format")
			return runConfigShow(configShowParams{
				targetDir: dir,
				format:    format,
				stdout:    cmd.OutOrStdout(),
			})
		},
	}
	showCmd.Flags().String("dir", ".", "Target directory")
	showCmd.Flags().String("format", "text", "Output format (text|json)")

	// --- validate subcommand ---
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			format, _ := cmd.Flags().GetString("format")
			return runConfigValidate(configValidateParams{
				targetDir: dir,
				format:    format,
				stdout:    cmd.OutOrStdout(),
			})
		},
	}
	validateCmd.Flags().String("dir", ".", "Target directory")
	validateCmd.Flags().String("format", "text", "Output format (text|json)")

	cmd.AddCommand(initCmd)
	cmd.AddCommand(showCmd)
	cmd.AddCommand(validateCmd)

	return cmd
}
