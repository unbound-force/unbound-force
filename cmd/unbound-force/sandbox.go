package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/sandbox"
)

// newSandboxCmd returns the `uf sandbox` parent command with
// all subcommands registered.
func newSandboxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Manage containerized OpenCode sessions",
		Long: `Launch, manage, and extract changes from containerized
OpenCode development sessions using Podman. The sandbox
provides an isolated environment for AI-assisted coding
with automatic API key forwarding and platform detection.

Subcommands:
  start    Launch a new sandbox container
  stop     Stop and remove the sandbox container
  attach   Connect to a running sandbox's TUI
  extract  Extract changes from the sandbox as git patches
  status   Show sandbox container status`,
	}

	cmd.AddCommand(
		newSandboxStartCmd(),
		newSandboxStopCmd(),
		newSandboxAttachCmd(),
		newSandboxExtractCmd(),
		newSandboxStatusCmd(),
	)

	return cmd
}

// --- start ---

type sandboxStartParams struct {
	projectDir string
	mode       string
	detach     bool
	image      string
	memory     string
	cpus       string
	stdout     io.Writer
	stderr     io.Writer
}

func runSandboxStart(p sandboxStartParams) error {
	return sandbox.Start(sandbox.Options{
		ProjectDir: p.projectDir,
		Mode:       p.mode,
		Detach:     p.detach,
		Image:      p.image,
		Memory:     p.memory,
		CPUs:       p.cpus,
		Stdout:     p.stdout,
		Stderr:     p.stderr,
	})
}

func newSandboxStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch a new sandbox container",
		Long: `Start a containerized OpenCode session with the current
project directory mounted. Detects platform capabilities,
pulls the container image if needed, starts the container,
waits for the OpenCode server health check, and attaches
the TUI.

Use --detach to start without attaching (server mode).
Use --mode direct for read-write mounts (changes apply
directly to the host filesystem).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			mode, _ := cmd.Flags().GetString("mode")
			detach, _ := cmd.Flags().GetBool("detach")
			image, _ := cmd.Flags().GetString("image")
			memory, _ := cmd.Flags().GetString("memory")
			cpus, _ := cmd.Flags().GetString("cpus")

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			return runSandboxStart(sandboxStartParams{
				projectDir: cwd,
				mode:       mode,
				detach:     detach,
				image:      image,
				memory:     memory,
				cpus:       cpus,
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

	return cmd
}

// --- stop ---

type sandboxStopParams struct {
	stdout io.Writer
}

func runSandboxStop(p sandboxStopParams) error {
	return sandbox.Stop(sandbox.Options{
		Stdout: p.stdout,
	})
}

func newSandboxStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop and remove the sandbox container",
		Long: `Stop the running sandbox container and remove it.
Idempotent — safe to run when no sandbox is running.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSandboxStop(sandboxStopParams{
				stdout: cmd.OutOrStdout(),
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
	stdout io.Writer
}

func runSandboxStatus(p sandboxStatusParams) error {
	status, err := sandbox.Status(sandbox.Options{
		Stdout: p.stdout,
	})
	if err != nil {
		return err
	}
	sandbox.FormatStatus(p.stdout, status)
	return nil
}

func newSandboxStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sandbox container status",
		Long: `Display the current state of the sandbox container
including container ID, image, mount mode, project
directory, server URL, and uptime.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSandboxStatus(sandboxStatusParams{
				stdout: cmd.OutOrStdout(),
			})
		},
	}
}
