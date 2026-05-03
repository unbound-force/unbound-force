package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/unbound-force/unbound-force/internal/config"
	"github.com/unbound-force/unbound-force/internal/ollamaproxy"
)

// newOllamaProxyCmd returns the `uf ollama-proxy` parent
// command with stop and status subcommands registered.
func newOllamaProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ollama-proxy",
		Short: "Start a local Ollama-compatible proxy to Vertex AI",
		Long: `Start a local proxy that serves the Ollama API on port
11434 (default). The proxy translates embedding requests to
Vertex AI and generation requests to the uf gateway.

This enables GPU-less development with Dewey and other
Ollama-compatible tools by routing requests to cloud APIs.

Use --detach to run in the background.
Use --embed-model to override the Vertex AI embedding model.
Use --gateway-url to override the uf gateway URL.
Use --port to listen on a different port.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			port, _ := cmd.Flags().GetInt("port")
			embedModel, _ := cmd.Flags().GetString("embed-model")
			gatewayURL, _ := cmd.Flags().GetString("gateway-url")
			detach, _ := cmd.Flags().GetBool("detach")

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			return runOllamaProxy(ollamaProxyParams{
				port:       port,
				embedModel: embedModel,
				gatewayURL: gatewayURL,
				detach:     detach,
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
				stderr:     cmd.ErrOrStderr(),
			})
		},
	}

	cmd.Flags().Int("port", ollamaproxy.DefaultPort,
		"Port to listen on")
	cmd.Flags().String("embed-model", "",
		"Vertex AI embedding model (default: text-embedding-005)")
	cmd.Flags().String("gateway-url", "",
		"Gateway URL for generation (default: http://localhost:53147)")
	cmd.Flags().Bool("detach", false,
		"Run proxy in the background")

	cmd.AddCommand(
		newOllamaProxyStopCmd(),
		newOllamaProxyStatusCmd(),
	)

	return cmd
}

// ollamaProxyParams holds testable parameters for the
// ollama-proxy command, following the established pattern
// from gateway.go.
type ollamaProxyParams struct {
	port       int
	embedModel string
	gatewayURL string
	detach     bool
	projectDir string
	stdout     io.Writer
	stderr     io.Writer
}

// runOllamaProxy executes the ollama-proxy start with
// testable parameters. CLI flags take precedence over
// config values, which take precedence over compiled
// defaults in the ollamaproxy package.
func runOllamaProxy(p ollamaProxyParams) error {
	opts := ollamaproxy.Options{
		Port:       p.port,
		EmbedModel: p.embedModel,
		GatewayURL: p.gatewayURL,
		Detach:     p.detach,
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
		Stderr:     p.stderr,
	}

	// Apply config defaults when CLI flags are at zero
	// value. Config fills the gap between CLI flags and
	// the proxy package's compiled defaults.
	cfg, _ := config.Load(config.LoadOptions{
		ProjectDir: p.projectDir,
	})
	if cfg != nil {
		if opts.Port == ollamaproxy.DefaultPort && cfg.OllamaProxy.Port != 0 {
			opts.Port = cfg.OllamaProxy.Port
		}
		if opts.EmbedModel == "" && cfg.OllamaProxy.EmbedModel != "" {
			opts.EmbedModel = cfg.OllamaProxy.EmbedModel
		}
		if opts.GatewayURL == "" && cfg.OllamaProxy.GatewayURL != "" {
			opts.GatewayURL = cfg.OllamaProxy.GatewayURL
		}
	}

	return ollamaproxy.Start(opts)
}

// --- stop ---

// ollamaProxyStopParams holds testable parameters for the
// ollama-proxy stop command.
type ollamaProxyStopParams struct {
	projectDir string
	stdout     io.Writer
}

// runOllamaProxyStop executes the ollama-proxy stop with
// testable parameters.
func runOllamaProxyStop(p ollamaProxyStopParams) error {
	return ollamaproxy.Stop(ollamaproxy.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
	})
}

// newOllamaProxyStopCmd returns the `uf ollama-proxy stop`
// subcommand.
func newOllamaProxyStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop a running ollama-proxy",
		Long: `Terminate a running background ollama-proxy and remove
its PID file. Prints "No ollama-proxy running." if no proxy
is found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runOllamaProxyStop(ollamaProxyStopParams{
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}
}

// --- status ---

// ollamaProxyStatusParams holds testable parameters for
// the ollama-proxy status command.
type ollamaProxyStatusParams struct {
	projectDir string
	stdout     io.Writer
}

// runOllamaProxyStatus executes the ollama-proxy status
// with testable parameters.
func runOllamaProxyStatus(p ollamaProxyStatusParams) error {
	return ollamaproxy.Status(ollamaproxy.Options{
		ProjectDir: p.projectDir,
		Stdout:     p.stdout,
	})
}

// newOllamaProxyStatusCmd returns the `uf ollama-proxy
// status` subcommand.
func newOllamaProxyStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show ollama-proxy status",
		Long: `Display the running ollama-proxy's port, model, PID,
and uptime. Prints "No ollama-proxy running." if no proxy
is found.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}
			return runOllamaProxyStatus(ollamaProxyStatusParams{
				projectDir: cwd,
				stdout:     cmd.OutOrStdout(),
			})
		},
	}
}
