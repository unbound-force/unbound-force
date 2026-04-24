package sandbox

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// WorkspaceStatus represents the current state of a
// sandbox workspace, regardless of backend. Extends
// the Spec 028 ContainerStatus with backend-agnostic
// fields.
type WorkspaceStatus struct {
	// Exists is true when a workspace has been created
	// (via `uf sandbox create`).
	Exists bool

	// Running is true when the workspace is active.
	Running bool

	// Backend is the backend type ("podman" or "che").
	Backend string

	// Name is the workspace name
	// (e.g., "uf-sandbox-myproject").
	Name string

	// ID is the workspace identifier (container ID for
	// Podman, workspace ID for Che). Short form.
	ID string

	// Image is the container image or devfile used.
	Image string

	// Mode is the workspace mode. For Podman: "isolated"
	// or "direct". For CDE: always "persistent".
	Mode string

	// ProjectDir is the source project directory (host
	// path for Podman, repo URL for CDE).
	ProjectDir string

	// ServerURL is the OpenCode server URL. For Podman:
	// http://localhost:4096. For CDE: the Che endpoint URL.
	ServerURL string

	// DemoEndpoints lists exposed demo port URLs.
	DemoEndpoints []DemoEndpoint

	// StartedAt is the workspace start time.
	StartedAt string

	// ExitCode is set when the workspace has stopped.
	// -1 when running. Only applicable for Podman.
	ExitCode int

	// Persistent is true when the workspace uses named
	// volumes or CDE storage (survives stop/start).
	Persistent bool
}

// DemoEndpoint represents an exposed port in the
// workspace accessible for demo review.
type DemoEndpoint struct {
	// Name is the endpoint name from the devfile
	// (e.g., "demo-web", "demo-api").
	Name string

	// Port is the container-internal port number.
	Port int

	// URL is the externally accessible URL. For Podman:
	// http://localhost:<port>. For CDE: the Che route URL.
	URL string

	// Protocol is "http" or "https".
	Protocol string
}

// SandboxConfig is the persistent sandbox configuration
// loaded from `.uf/sandbox.yaml`. Provides defaults for
// CDE URL, Ollama endpoint, backend selection, and demo
// port mappings.
type SandboxConfig struct {
	// Che contains CDE backend configuration.
	Che CheConfig `yaml:"che"`

	// Ollama contains Ollama endpoint configuration.
	Ollama OllamaConfig `yaml:"ollama"`

	// Backend is the default backend: "auto", "podman",
	// or "che". Default: "auto".
	Backend string `yaml:"backend"`

	// DemoPorts lists port numbers to expose for demos
	// (Podman only; CDE uses devfile endpoints).
	DemoPorts []int `yaml:"demo_ports"`
}

// CheConfig contains Eclipse Che connection settings.
type CheConfig struct {
	// URL is the Che/Dev Spaces instance URL.
	// Can also be set via UF_CHE_URL env var.
	URL string `yaml:"url"`

	// Token is the authentication token for REST API.
	// Can also be set via UF_CHE_TOKEN env var.
	// Only needed when chectl is not available.
	Token string `yaml:"token"`
}

// OllamaConfig contains Ollama endpoint settings.
type OllamaConfig struct {
	// Host is the Ollama endpoint URL. Overrides the
	// default host.containers.internal:11434 for CDE
	// deployments where that hostname doesn't resolve.
	Host string `yaml:"host"`
}

// sanitizeRe matches characters that are not lowercase
// alphanumeric or hyphens.
var sanitizeRe = regexp.MustCompile(`[^a-z0-9-]`)

// ProjectNameFromDir derives a sanitized workspace name
// from the project directory path. Exported for use by
// the CLI layer (confirmation prompts).
func ProjectNameFromDir(dir string) string {
	return projectName(dir)
}

// projectName derives a sanitized workspace name from
// the project directory path. Returns the directory
// basename converted to lowercase with non-alphanumeric
// characters replaced by hyphens. Falls back to "default"
// if the result is empty.
func projectName(dir string) string {
	name := filepath.Base(dir)
	name = strings.ToLower(name)
	name = sanitizeRe.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "default"
	}
	return name
}

// containerNameForProject returns the Podman container
// name for a persistent workspace:
// "uf-sandbox-<project-name>".
func containerNameForProject(dir string) string {
	return "uf-sandbox-" + projectName(dir)
}

// volumeNameForProject returns the Podman named volume
// name for a persistent workspace:
// "uf-sandbox-<project-name>".
func volumeNameForProject(dir string) string {
	return "uf-sandbox-" + projectName(dir)
}

// LoadConfig reads the sandbox configuration from
// .uf/sandbox.yaml (or the path specified in opts.ConfigPath).
// Returns a zero-value SandboxConfig with defaults if the
// file does not exist. Environment variables override
// config file values.
//
func LoadConfig(opts Options) (SandboxConfig, error) {
	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = filepath.Join(opts.ProjectDir, DefaultConfigPath)
	}

	var cfg SandboxConfig

	data, err := opts.ReadFile(configPath)
	if err == nil {
		if parseErr := yaml.Unmarshal(data, &cfg); parseErr != nil {
			return cfg, fmt.Errorf("parse %s: %w", configPath, parseErr)
		}
	}

	// Environment variable overrides.
	if envURL := opts.Getenv("UF_CHE_URL"); envURL != "" {
		cfg.Che.URL = envURL
	}
	if envToken := opts.Getenv("UF_CHE_TOKEN"); envToken != "" {
		cfg.Che.Token = envToken
	}
	if envOllama := opts.Getenv("UF_OLLAMA_HOST"); envOllama != "" {
		cfg.Ollama.Host = envOllama
	}
	if envBackend := opts.Getenv("UF_SANDBOX_BACKEND"); envBackend != "" {
		cfg.Backend = envBackend
	}

	return cfg, nil
}

// FormatWorkspaceStatus writes a human-readable status
// report for a persistent workspace to the writer.
func FormatWorkspaceStatus(w io.Writer, ws WorkspaceStatus) {
	if !ws.Exists {
		fmt.Fprintf(w, "No sandbox workspace found.\n")
		return
	}

	state := "stopped"
	if ws.Running {
		state = "running"
	}

	modeLabel := ws.Mode
	if ws.Persistent {
		modeLabel += " (persistent)"
	}

	fmt.Fprintf(w, "Sandbox Status\n")
	fmt.Fprintf(w, "  Workspace:  %s\n", ws.Name)
	fmt.Fprintf(w, "  Backend:    %s\n", modeLabel)
	if ws.Image != "" {
		fmt.Fprintf(w, "  Image:      %s\n", ws.Image)
	}
	fmt.Fprintf(w, "  State:      %s\n", state)
	if ws.ProjectDir != "" {
		fmt.Fprintf(w, "  Project:    %s\n", ws.ProjectDir)
	}
	if ws.ServerURL != "" {
		fmt.Fprintf(w, "  Server:     %s\n", ws.ServerURL)
	}
	for i, ep := range ws.DemoEndpoints {
		if i == 0 {
			fmt.Fprintf(w, "  Demo:       %s (%s)\n", ep.URL, ep.Name)
		} else {
			fmt.Fprintf(w, "              %s (%s)\n", ep.URL, ep.Name)
		}
	}
	if ws.StartedAt != "" {
		fmt.Fprintf(w, "  Started:    %s\n", ws.StartedAt)
	}
}

// setupGitSync configures the workspace's git remote
// and branch for bidirectional sync. For Podman, this
// runs git commands inside the container. For CDE, Che
// handles git clone from the devfile projects section.
func setupGitSync(opts Options) error {
	name := containerNameForProject(opts.ProjectDir)

	// Detect current branch on host.
	branchOut, err := opts.ExecCmd("git", "-C", opts.ProjectDir,
		"rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		// Not a git repo — skip git sync setup.
		return nil
	}
	branch := strings.TrimSpace(string(branchOut))

	// Get remote URL.
	remoteOut, err := opts.ExecCmd("git", "-C", opts.ProjectDir,
		"remote", "get-url", "origin")
	if err != nil {
		// No remote — skip git sync setup.
		return nil
	}
	remote := strings.TrimSpace(string(remoteOut))

	// Configure git inside the container.
	cmds := [][]string{
		{"podman", "exec", name, "git", "-C", "/workspace", "remote", "set-url", "origin", remote},
		{"podman", "exec", name, "git", "-C", "/workspace", "checkout", branch},
	}

	for _, cmd := range cmds {
		if _, err := opts.ExecCmd(cmd[0], cmd[1:]...); err != nil {
			// Non-fatal — git sync is best-effort.
			return nil
		}
	}

	return nil
}

// checkGitSync verifies the workspace's git state is
// clean and up-to-date with the remote. Returns nil if
// clean, error with details if diverged.
func checkGitSync(opts Options) error {
	name := containerNameForProject(opts.ProjectDir)

	// Check for uncommitted changes.
	statusOut, err := opts.ExecCmd("podman", "exec", name,
		"git", "-C", "/workspace", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("check workspace git status: %w", err)
	}
	if strings.TrimSpace(string(statusOut)) != "" {
		return fmt.Errorf("uncommitted changes in workspace — commit or stash before syncing")
	}

	// Try fast-forward pull.
	if _, err := opts.ExecCmd("podman", "exec", name,
		"git", "-C", "/workspace", "pull", "--ff-only"); err != nil {
		return fmt.Errorf("workspace and remote have diverged — resolve conflicts before continuing")
	}

	return nil
}
