package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/gateway"
)

// newGatewayCmd returns the `uf gateway` parent command
// with stop and status subcommands registered.
func newGatewayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a local LLM reverse proxy",
		Long: `Start a local reverse proxy that serves the Anthropic
Messages API on port 53147 (default). The gateway auto-detects
the cloud provider from environment variables and injects
host-side credentials into upstream requests.

Supported providers:
  - Anthropic (ANTHROPIC_API_KEY)
  - Vertex AI (CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID)
  - AWS Bedrock (CLAUDE_CODE_USE_BEDROCK=1)

Use --detach to run in the background.
Use --provider to override auto-detection.
Use --port to listen on a different port.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			provider, _ := cmd.Flags().GetString("provider")
			detach, _ := cmd.Flags().GetBool("detach")

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			return runGateway(gatewayParams{
				port:       port,
				provider:   provider,
				detach:     detach,
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
				stderr:     cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().Int("port", gateway.DefaultPort,
		"Port to listen on")
	cmd.Flags().String("provider", "",
		"Provider override: anthropic, vertex, or bedrock (auto-detected if omitted)")
	cmd.Flags().Bool("detach", false,
		"Run gateway in the background")

	cmd.AddCommand(
		newGatewayStopCmd(),
		newGatewayStatusCmd(),
	)

	return cmd
}

// gatewayParams holds testable parameters for the gateway
// command, following the established pattern from sandbox.
type gatewayParams struct {
	port       int
	provider   string
	detach     bool
	projectDir string
	stdout     io.Writer
	stderr     io.Writer
}

// runGateway executes the gateway start with testable
// parameters.
func runGateway(p gatewayParams) error {
	return gateway.Start(gateway.Options{
		Port:         p.port,
		ProviderName: p.provider,
		Detach:       p.detach,
		ProjectDir:   p.projectDir,
		Stdout:       p.stdout,
		Stderr:       p.stderr,
	})
}

// --- stop ---

// gatewayStopParams holds testable parameters for the
// gateway stop command.
type gatewayStopParams struct {
	projectDir string
	stdout     io.Writer
}

// runGatewayStop executes the gateway stop with testable
// parameters.
func runGatewayStop(p gatewayStopParams) error {
	return gateway.Stop(gateway.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
	})
}

// newGatewayStopCmd returns the `uf gateway stop`
// subcommand.
func newGatewayStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a running gateway",
		Long: `Terminate a running background gateway and remove its
PID file. Prints "No gateway running." if no gateway is found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runGatewayStop(gatewayStopParams{
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}
}

// --- status ---

// gatewayStatusParams holds testable parameters for the
// gateway status command.
type gatewayStatusParams struct {
	projectDir string
	stdout     io.Writer
}

// runGatewayStatus executes the gateway status with
// testable parameters.
func runGatewayStatus(p gatewayStatusParams) error {
	return gateway.Status(gateway.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
	})
}

// newGatewayStatusCmd returns the `uf gateway status`
// subcommand.
func newGatewayStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show gateway status",
		Long: `Display the running gateway's provider, port, PID, and
uptime. Prints "No gateway running." if no gateway is found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runGatewayStatus(gatewayStatusParams{
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}
}
