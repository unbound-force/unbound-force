package sandbox

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// DevPodBackend implements Backend for DevPod workspaces
// using devcontainer.json configuration. DevPod is invoked
// as a subprocess via the ExecCmd injection pattern (D1:
// subprocess only, no Go library imports).
type DevPodBackend struct{}

// Name returns the backend identifier.
func (b *DevPodBackend) Name() string { return BackendDevPod }

// devpodWorkspaceName returns the DevPod workspace name
// for a project: "uf-sandbox-<project-name>". Matches the
// Podman persistent workspace naming convention (D5).
func devpodWorkspaceName(opts Options) string {
	return "uf-sandbox-" + projectName(opts.ProjectDir)
}

// Create provisions a DevPod workspace with the project's
// devcontainer configuration.
//
// Pre-flight checks:
//  1. podman in PATH (DevPod Podman provider requirement)
//  2. DevPod >= 0.5.0 (D5b: minimum version)
//  3. .devcontainer/devcontainer.json exists
//
// Then calls: devpod up <project-dir> --provider podman
// --id <workspace-name> --ide none [--workspace-env ...]
func (b *DevPodBackend) Create(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	// Pre-flight: podman must be installed for the DevPod
	// Podman provider (D4).
	if _, err := opts.LookPath("podman"); err != nil {
		return fmt.Errorf(
			"podman not found — DevPod requires Podman as its container provider. " +
				"Install: brew install podman")
	}

	// Pre-flight: verify DevPod >= 0.5.0 (D5b).
	if err := checkDevPodVersion(opts); err != nil {
		return err
	}

	// Pre-flight: devcontainer.json must exist.
	dcPath := filepath.Join(opts.ProjectDir,
		".devcontainer", "devcontainer.json")
	if _, err := opts.ReadFile(dcPath); err != nil {
		return fmt.Errorf(
			".devcontainer/devcontainer.json not found — "+
				"run `uf sandbox init` to create it")
	}

	wsName := devpodWorkspaceName(opts)

	// Build devpod up arguments.
	args := []string{
		"up", opts.ProjectDir,
		"--provider", "podman",
		"--id", wsName,
		"--ide", "none",
	}

	// Gateway env var injection via --workspace-env (D4).
	if opts.GatewayActive {
		args = append(args,
			"--workspace-env",
			fmt.Sprintf("ANTHROPIC_BASE_URL=http://host.containers.internal:%d",
				opts.GatewayPort),
			"--workspace-env",
			"ANTHROPIC_API_KEY=gateway",
		)
	}

	fmt.Fprintf(opts.Stderr, "Creating DevPod workspace %s...\n", wsName)

	out, err := opts.ExecCmd("devpod", args...)
	if err != nil {
		return fmt.Errorf("devpod up failed: %s",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stderr, "DevPod workspace created: %s\n", wsName)
	return nil
}

// Start resumes a stopped DevPod workspace by calling
// devpod up --id <name> (idempotent — starts or resumes).
func (b *DevPodBackend) Start(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	wsName := devpodWorkspaceName(opts)

	args := []string{"up", "--id", wsName}

	fmt.Fprintf(opts.Stderr, "Resuming DevPod workspace %s...\n", wsName)

	out, err := opts.ExecCmd("devpod", args...)
	if err != nil {
		return fmt.Errorf("devpod up failed: %s",
			strings.TrimSpace(string(out)))
	}

	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)

	if opts.Detach {
		fmt.Fprintf(opts.Stdout,
			"Sandbox resumed (detached).\nServer: %s\n", serverURL)
		return nil
	}

	fmt.Fprintf(opts.Stderr, "Attaching to sandbox...\n")
	return b.Attach(opts)
}

// Stop stops a running DevPod workspace.
func (b *DevPodBackend) Stop(opts Options) error {
	opts.defaults()

	wsName := devpodWorkspaceName(opts)

	out, err := opts.ExecCmd("devpod", "stop", wsName)
	if err != nil {
		return fmt.Errorf("devpod stop failed: %s",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stdout, "Sandbox stopped (state preserved).\n")
	return nil
}

// Destroy permanently deletes the DevPod workspace.
func (b *DevPodBackend) Destroy(opts Options) error {
	opts.defaults()

	wsName := devpodWorkspaceName(opts)

	out, err := opts.ExecCmd("devpod", "delete", wsName, "--force")
	if err != nil {
		return fmt.Errorf("devpod delete failed: %s",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stdout, "Sandbox destroyed.\n")
	return nil
}

// devpodStatusOutput is the subset of devpod status JSON
// output that we parse. DevPod returns a JSON object with
// workspace state information.
type devpodStatusOutput struct {
	ID       string `json:"id"`
	State    string `json:"state"`
	Provider string `json:"provider"`
	IDE      string `json:"ide"`
}

// Status returns the current state of the DevPod workspace.
func (b *DevPodBackend) Status(opts Options) (WorkspaceStatus, error) {
	opts.defaults()

	wsName := devpodWorkspaceName(opts)

	out, err := opts.ExecCmd("devpod", "status", wsName,
		"--output", "json")
	if err != nil {
		// Workspace does not exist.
		return WorkspaceStatus{}, nil
	}

	var status devpodStatusOutput
	if err := json.Unmarshal(out, &status); err != nil {
		return WorkspaceStatus{}, fmt.Errorf(
			"parse devpod status: %w", err)
	}

	ws := WorkspaceStatus{
		Exists:     true,
		Running:    strings.EqualFold(status.State, "Running"),
		Backend:    BackendDevPod,
		Name:       wsName,
		ID:         status.ID,
		Persistent: true,
		Mode:       opts.Mode,
		ProjectDir: opts.ProjectDir,
		ServerURL:  fmt.Sprintf("http://localhost:%d", DefaultServerPort),
	}

	return ws, nil
}

// Attach connects the TUI to the running DevPod workspace's
// OpenCode server via opencode attach.
func (b *DevPodBackend) Attach(opts Options) error {
	opts.defaults()

	if _, err := opts.LookPath("opencode"); err != nil {
		return fmt.Errorf(
			"opencode not found. Install: brew install anomalyco/tap/opencode")
	}

	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)
	return opts.ExecInteractive("opencode", "attach", serverURL)
}

// parseDevPodVersion runs `devpod version` and parses the
// semver output. Returns major, minor, patch as integers.
// Follows the parsePodmanVersion() pattern (D5b).
func parseDevPodVersion(opts Options) (int, int, int, error) {
	out, err := opts.ExecCmd("devpod", "version")
	if err != nil {
		return 0, 0, 0, fmt.Errorf(
			"failed to get devpod version: %w", err)
	}

	// DevPod version output is typically just "v0.5.18" or
	// "0.5.18". Strip leading 'v' if present.
	version := strings.TrimSpace(string(out))
	version = strings.TrimPrefix(version, "v")

	segments := strings.SplitN(version, ".", 3)
	if len(segments) < 2 {
		return 0, 0, 0, fmt.Errorf(
			"cannot parse devpod version: %s", version)
	}

	major, err := strconv.Atoi(segments[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf(
			"cannot parse major version: %w", err)
	}
	minor, err := strconv.Atoi(segments[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf(
			"cannot parse minor version: %w", err)
	}

	patch := 0
	if len(segments) >= 3 {
		// Patch may contain pre-release suffix (e.g., "18-beta").
		patchStr := segments[2]
		if dashIdx := strings.IndexByte(patchStr, '-'); dashIdx >= 0 {
			patchStr = patchStr[:dashIdx]
		}
		patch, _ = strconv.Atoi(patchStr)
	}

	return major, minor, patch, nil
}

// checkDevPodVersion verifies DevPod >= 0.5.0 is installed.
func checkDevPodVersion(opts Options) error {
	major, minor, _, err := parseDevPodVersion(opts)
	if err != nil {
		return err
	}

	// Minimum version: 0.5.0 (D5b).
	if major == 0 && minor < 5 {
		return fmt.Errorf(
			"devpod >= 0.5.0 required (current: %d.%d). "+
				"Update: https://devpod.sh/docs/getting-started/install",
			major, minor)
	}

	return nil
}

// isDevPodWorkspace checks if a DevPod workspace exists
// for the current project by calling devpod status. Returns
// true if the workspace exists (any state). Guarded by
// LookPath("devpod") — returns false if DevPod is not
// installed (D5a).
func isDevPodWorkspace(opts Options) bool {
	if _, err := opts.LookPath("devpod"); err != nil {
		return false
	}

	wsName := devpodWorkspaceName(opts)
	_, err := opts.ExecCmd("devpod", "status", wsName,
		"--output", "json")
	return err == nil
}
