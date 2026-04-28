package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/config"
	"github.com/unbound-force/unbound-force/internal/sandbox"
)

// newSandboxCmd returns the `uf sandbox` parent command with
// all subcommands registered.
func newSandboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage containerized OpenCode sessions",
		Long: `Launch, manage, and extract changes from containerized
OpenCode development sessions. Supports Podman (local) and
Eclipse Che / Dev Spaces (CDE) backends.

Subcommands:
  create   Provision a persistent sandbox workspace
  destroy  Permanently delete a sandbox workspace
  start    Launch or resume a sandbox
  stop     Stop a sandbox (preserves persistent state)
  attach   Connect to a running sandbox's TUI
  extract  Extract changes from the sandbox as git patches
  status   Show sandbox workspace status`,
	}

	cmd.AddCommand(
		newSandboxCreateCmd(),
		newSandboxDestroyCmd(),
		newSandboxStartCmd(),
		newSandboxStopCmd(),
		newSandboxAttachCmd(),
		newSandboxExtractCmd(),
		newSandboxStatusCmd(),
	)

	return cmd
}

// applySandboxConfig applies config defaults to sandbox Options
// fields that are at their zero/default values. CLI flags take
// precedence (they're already set on the params struct).
// Also prints a deprecation warning to stderr if .uf/sandbox.yaml
// exists (per design D9 — config values flow through the cmd
// layer, not internal packages, per design D8).
func applySandboxConfig(opts *sandbox.Options, stderr io.Writer) {
	cfg, _ := config.Load(config.LoadOptions{ProjectDir: opts.ProjectDir})
	if cfg == nil {
		return
	}
	if opts.BackendName == "" && cfg.Sandbox.Backend != "" && cfg.Sandbox.Backend != "auto" {
		opts.BackendName = cfg.Sandbox.Backend
	}
	if opts.Image == "" && cfg.Sandbox.Image != "" {
		opts.Image = cfg.Sandbox.Image
	}
	if opts.Memory == "" && cfg.Sandbox.Resources.Memory != "" {
		opts.Memory = cfg.Sandbox.Resources.Memory
	}
	if opts.CPUs == "" && cfg.Sandbox.Resources.CPUs != "" {
		opts.CPUs = cfg.Sandbox.Resources.CPUs
	}
	if opts.Mode == "" && cfg.Sandbox.Mode != "" {
		opts.Mode = cfg.Sandbox.Mode
	}
	if !opts.UIDMap && cfg.Sandbox.UIDMap {
		opts.UIDMap = true
	}

	// Deprecation warning: if legacy .uf/sandbox.yaml exists,
	// warn via the injectable stderr writer (testable, per D8).
	legacyPath := filepath.Join(opts.ProjectDir, ".uf", "sandbox.yaml")
	if _, err := os.Stat(legacyPath); err == nil && stderr != nil {
		fmt.Fprintln(stderr,
			"Warning: .uf/sandbox.yaml is deprecated. Run 'uf config init' to migrate to .uf/config.yaml")
	}
}

// --- create ---

type sandboxCreateParams struct {
	projectDir string
	backend    string
	image      string
	memory     string
	cpus       string
	name       string
	detach     bool
	uidmap     bool
	demoPorts  []int
	stdout     io.Writer
	stderr     io.Writer
}

func runSandboxCreate(p sandboxCreateParams) error {
	opts := sandbox.Options{
		ProjectDir:    p.projectDir,
		BackendName:   p.backend,
		Image:         p.image,
		Memory:        p.memory,
		CPUs:          p.cpus,
		WorkspaceName: p.name,
		Detach:        p.detach,
		UIDMap:        p.uidmap,
		DemoPorts:     p.demoPorts,
		Stdout:        p.stdout,
		Stderr:        p.stderr,
	}
	applySandboxConfig(&opts, p.stderr)
	return sandbox.Create(opts)
}

func newSandboxCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Provision a persistent sandbox workspace",
		Long: `Provision a persistent sandbox workspace for the current
project. Uses Eclipse Che/Dev Spaces when configured,
Podman with named volumes otherwise.

The workspace persists across stop/start cycles. Use
'uf sandbox destroy' to permanently remove it.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backend, _ := cmd.Flags().GetString("backend")
			image, _ := cmd.Flags().GetString("image")
			memory, _ := cmd.Flags().GetString("memory")
			cpus, _ := cmd.Flags().GetString("cpus")
			name, _ := cmd.Flags().GetString("name")
			detach, _ := cmd.Flags().GetBool("detach")
			uidmap, _ := cmd.Flags().GetBool("uidmap")
			demoPorts, _ := cmd.Flags().GetIntSlice("demo-ports")

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			return runSandboxCreate(sandboxCreateParams{
				projectDir: cwd,
				backend:    backend,
				image:      image,
				memory:     memory,
				cpus:       cpus,
				name:       name,
				detach:     detach,
				uidmap:     uidmap,
				demoPorts:  demoPorts,
				stdout:     cmd.OutOrStdout(),
				stderr:     cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().String("backend", "auto",
		"Backend: auto, podman, or che")
	cmd.Flags().String("image", "",
		"Container image (Podman only; default from UF_SANDBOX_IMAGE or quay.io/unbound-force/opencode-dev:latest)")
	cmd.Flags().String("memory", "",
		"Memory limit (default \"8g\")")
	cmd.Flags().String("cpus", "",
		"CPU limit (default \"4\")")
	cmd.Flags().String("name", "",
		"Workspace name override (default \"uf-sandbox-<project-name>\")")
	cmd.Flags().Bool("detach", false,
		"Start without attaching TUI")
	cmd.Flags().IntSlice("demo-ports", nil,
		"Additional ports to expose for demos (comma-separated, e.g., 3000,8080)")
	cmd.Flags().Bool("uidmap", false,
		"Use explicit UID/GID mapping (for macOS when Podman machine virtiofs does not support --userns=keep-id)")

	return cmd
}

// --- destroy ---

type sandboxDestroyParams struct {
	projectDir string
	yes        bool
	force      bool
	stdout     io.Writer
	stderr     io.Writer
	stdin      io.Reader
}

func runSandboxDestroy(p sandboxDestroyParams) error {
	// Confirmation prompt unless --yes.
	if !p.yes {
		projName := sandbox.ProjectNameFromDir(p.projectDir)
		wsName := "uf-sandbox-" + projName
		fmt.Fprintf(p.stdout,
			"Destroy sandbox %q?\nThis will permanently delete all workspace state.\n[y/N] ",
			wsName)
		var response string
		if _, err := fmt.Fscanln(p.stdin, &response); err != nil {
			return nil
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Fprintf(p.stdout, "Cancelled.\n")
			return nil
		}
	}

	return sandbox.Destroy(sandbox.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
		Stderr:     p.stderr,
	})
}

func newSandboxDestroyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Permanently delete a sandbox workspace",
		Long: `Permanently delete the sandbox workspace and all
associated state (named volumes, CDE workspace).

Use --yes to skip the confirmation prompt.
Use --force to destroy even if the workspace is running.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			force, _ := cmd.Flags().GetBool("force")

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			return runSandboxDestroy(sandboxDestroyParams{
				projectDir: cwd,
				yes:        yes,
				force:      force,
				stdout:     cmd.OutOrStdout(),
				stderr:     cmd.ErrOrStderr(),
				stdin:      cmd.InOrStdin(),
			})
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	cmd.Flags().Bool("force", false,
		"Force destroy even if workspace is running")

	return cmd
}

// --- start ---

type sandboxStartParams struct {
	projectDir string
	mode       string
	detach     bool
	noParent   bool
	uidmap     bool
	image      string
	memory     string
	cpus       string
	backend    string
	stdout     io.Writer
	stderr     io.Writer
}

func runSandboxStart(p sandboxStartParams) error {
	opts := sandbox.Options{
		ProjectDir:  p.projectDir,
		Mode:        p.mode,
		Detach:      p.detach,
		NoParent:    p.noParent,
		UIDMap:      p.uidmap,
		Image:       p.image,
		Memory:      p.memory,
		CPUs:        p.cpus,
		BackendName: p.backend,
		Stdout:      p.stdout,
		Stderr:      p.stderr,
	}
	applySandboxConfig(&opts, p.stderr)
	return sandbox.Start(opts)
}

func newSandboxStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch or resume a sandbox",
		Long: `Start a containerized OpenCode session. If a persistent
workspace exists (from 'uf sandbox create'), resumes it.
Otherwise, starts an ephemeral container with the current
project directory mounted.

Use --detach to start without attaching (server mode).
Use --mode direct for read-write mounts (changes apply
directly to the host filesystem).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, _ := cmd.Flags().GetString("mode")
			detach, _ := cmd.Flags().GetBool("detach")
			noParent, _ := cmd.Flags().GetBool("no-parent")
			uidmap, _ := cmd.Flags().GetBool("uidmap")
			image, _ := cmd.Flags().GetString("image")
			memory, _ := cmd.Flags().GetString("memory")
			cpus, _ := cmd.Flags().GetString("cpus")
			backend, _ := cmd.Flags().GetString("backend")

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			return runSandboxStart(sandboxStartParams{
				projectDir: cwd,
				mode:       mode,
				detach:     detach,
				noParent:   noParent,
				uidmap:     uidmap,
				image:      image,
				memory:     memory,
				cpus:       cpus,
				backend:    backend,
				stdout:     cmd.OutOrStdout(),
				stderr:     cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().String("mode", "isolated",
		"Mount mode: isolated (read-only) or direct (read-write)")
	cmd.Flags().Bool("detach", false,
		"Start container without attaching TUI")
	cmd.Flags().String("image", "",
		"Container image (default from UF_SANDBOX_IMAGE or quay.io/unbound-force/opencode-dev:latest)")
	cmd.Flags().String("memory", "",
		"Container memory limit (default \"8g\")")
	cmd.Flags().String("cpus", "",
		"Container CPU limit (default \"4\")")
	cmd.Flags().String("backend", "auto",
		"Backend: auto, podman, or che")
	cmd.Flags().Bool("no-parent", false,
		"Mount only the project directory (disable parent directory mount)")
	cmd.Flags().Bool("uidmap", false,
		"Use explicit UID/GID mapping (for macOS when Podman machine virtiofs does not support --userns=keep-id)")

	return cmd
}

// --- stop ---

type sandboxStopParams struct {
	projectDir string
	stdout     io.Writer
}

func runSandboxStop(p sandboxStopParams) error {
	return sandbox.Stop(sandbox.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
	})
}

func newSandboxStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a sandbox (preserves persistent state)",
		Long: `Stop the running sandbox. For persistent workspaces
(created via 'uf sandbox create'), the workspace state
is preserved. For ephemeral containers, the container
is removed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runSandboxStop(sandboxStopParams{
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}
}

// --- attach ---

type sandboxAttachParams struct {
	stdout io.Writer
}

func runSandboxAttach(p sandboxAttachParams) error {
	return sandbox.Attach(sandbox.Options{
		Stdout: p.stdout,
	})
}

func newSandboxAttachCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Connect to a running sandbox's TUI",
		Long: `Attach the terminal to the running sandbox's OpenCode
server via 'opencode attach'. Requires the sandbox to be
running and OpenCode to be installed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSandboxAttach(sandboxAttachParams{
				stdout: cmd.OutOrStdout(),
			})
		},
	}
}

// --- extract ---

type sandboxExtractParams struct {
	yes    bool
	stdout io.Writer
	stderr io.Writer
	stdin  io.Reader
}

func runSandboxExtract(p sandboxExtractParams) error {
	return sandbox.Extract(sandbox.Options{
		Yes:    p.yes,
		Stdout: p.stdout,
		Stderr: p.stderr,
		Stdin:  p.stdin,
	})
}

func newSandboxExtractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract changes from the sandbox as git patches",
		Long: `Generate a patch from the container's git history,
present it for review, and apply it to the host repo
on confirmation. Uses git format-patch / git am for
commit-preserving round-trip extraction.

Use --yes to skip the confirmation prompt.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			yes, _ := cmd.Flags().GetBool("yes")
			return runSandboxExtract(sandboxExtractParams{
				yes:    yes,
				stdout: cmd.OutOrStdout(),
				stderr: cmd.ErrOrStderr(),
				stdin:  cmd.InOrStdin(),
			})
		},
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")

	return cmd
}

// --- status ---

type sandboxStatusParams struct {
	projectDir string
	stdout     io.Writer
}

func runSandboxStatus(p sandboxStatusParams) error {
	opts := sandbox.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
	}

	// Check for persistent workspace first.
	ws, err := sandbox.WorkspaceStatusCheck(opts)
	if err != nil {
		return err
	}
	if ws.Exists {
		sandbox.FormatWorkspaceStatus(p.stdout, ws)
		return nil
	}

	// Fall back to ephemeral status (Spec 028).
	status, err := sandbox.Status(opts)
	if err != nil {
		return err
	}
	sandbox.FormatStatus(p.stdout, status)
	return nil
}

func newSandboxStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sandbox workspace status",
		Long: `Display the current state of the sandbox workspace
including workspace name, backend, image, state, project
directory, server URL, demo endpoints, and uptime.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runSandboxStatus(sandboxStatusParams{
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}
}
