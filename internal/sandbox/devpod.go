package sandbox

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// validIDEs lists the IDE values supported by DevPod.
// Used by validateIDE to reject unknown values before
// invoking devpod up.
var validIDEs = []string{
	"none", "vscode", "openvscode",
	"fleet", "jupyternotebook", "cursor",
}

// safeWSNameRe validates that a workspace name contains only
// characters safe for shell command interpolation. This is
// the enforcement boundary for Principle V (Security by Default):
// even though projectName() sanitizes upstream, this guard
// ensures the invariant holds regardless of how wsName is
// derived in the future.
var safeWSNameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

// validateIDE checks that the IDE value is in the set of
// DevPod-supported IDEs. Returns an error listing all valid
// values when the input is not recognized.
func validateIDE(ide string) error {
	for _, v := range validIDEs {
		if ide == v {
			return nil
		}
	}
	return fmt.Errorf(
		"unknown IDE: %s, use one of: %s",
		ide, strings.Join(validIDEs, ", "))
}

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

	// Validate IDE before building args — fail early with
	// a clear error listing valid values (D3).
	if err := validateIDE(opts.IDE); err != nil {
		return err
	}

	wsName := devpodWorkspaceName(opts)

	// Build devpod up arguments.
	args := []string{
		"up", opts.ProjectDir,
		"--provider", "podman",
		"--id", wsName,
		"--ide", opts.IDE,
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
		// D12: Check workspace status before reporting error.
		// DevPod v0.6.x has a Bun tunnel bug that causes
		// non-zero exit even after successful workspace
		// creation. Use workspace status as source of truth.
		if b.isWorkspaceRunning(opts, wsName) {
			fmt.Fprintf(opts.Stderr,
				"DevPod workspace created: %s "+
					"(IDE tunnel had a non-fatal error)\n", wsName)
		} else {
			return fmt.Errorf("devpod up failed: %w: %s", err,
				strings.TrimSpace(string(out)))
		}
	} else {
		fmt.Fprintf(opts.Stderr, "DevPod workspace created: %s\n", wsName)
	}

	// D11: Wait for OpenCode server health check after
	// creation, matching the pattern in Start().
	fmt.Fprintf(opts.Stderr, "Waiting for OpenCode server...\n")
	if err := waitForHealth(opts, opts.HealthCheckTimeout); err != nil {
		// D13: SSH fallback — start server manually for
		// workspaces where postStartCommand didn't run.
		if sshErr := b.startServerViaSSH(opts, wsName); sshErr == nil {
			// Server started via SSH, wait again with
			// shorter timeout.
			if err2 := waitForHealth(opts, 15*time.Second); err2 == nil {
				return b.attachOrDetach(opts)
			}
		}
		// Health check failed and SSH fallback didn't help.
		// Report the workspace as created but without the
		// server — don't attempt attach.
		fmt.Fprintf(opts.Stderr,
			"Warning: OpenCode server not responding.\n"+
				"Workspace created. Try: uf sandbox attach\n")
		return nil
	}

	return b.attachOrDetach(opts)
}

// Start resumes a stopped DevPod workspace by calling
// devpod up --id <name> (idempotent — starts or resumes).
func (b *DevPodBackend) Start(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	// Validate IDE before building args — fail early with
	// a clear error listing valid values (D3).
	if err := validateIDE(opts.IDE); err != nil {
		return err
	}

	wsName := devpodWorkspaceName(opts)

	args := []string{"up", "--id", wsName, "--ide", opts.IDE}

	fmt.Fprintf(opts.Stderr, "Resuming DevPod workspace %s...\n", wsName)

	out, err := opts.ExecCmd("devpod", args...)
	if err != nil {
		// D12: Check workspace status before reporting error.
		if b.isWorkspaceRunning(opts, wsName) {
			fmt.Fprintf(opts.Stderr,
				"DevPod workspace resumed: %s "+
					"(IDE tunnel had a non-fatal error)\n", wsName)
		} else {
			return fmt.Errorf("devpod up failed: %w: %s", err,
				strings.TrimSpace(string(out)))
		}
	}

	// Wait for the OpenCode server health check before
	// attaching. DevPod containers may take a moment to
	// start the server via postStartCommand.
	//
	// D3/D8: Health timeout is non-fatal for DevPod
	// because VS Code connects independently via its
	// own tunnel. The user can still attach manually
	// via "uf sandbox attach".
	fmt.Fprintf(opts.Stderr, "Waiting for OpenCode server...\n")
	if err := waitForHealth(opts, opts.HealthCheckTimeout); err != nil {
		// D13: SSH fallback — start server manually for
		// workspaces where postStartCommand didn't run.
		if sshErr := b.startServerViaSSH(opts, wsName); sshErr == nil {
			if err2 := waitForHealth(opts, 15*time.Second); err2 == nil {
				return b.attachOrDetach(opts)
			}
		}
		fmt.Fprintf(opts.Stderr,
			"Warning: OpenCode server not responding.\n"+
				"VS Code may still be connected. Try: uf sandbox attach\n")
		return nil
	}

	return b.attachOrDetach(opts)
}

// Stop stops a running DevPod workspace.
func (b *DevPodBackend) Stop(opts Options) error {
	opts.defaults()

	wsName := devpodWorkspaceName(opts)

	out, err := opts.ExecCmd("devpod", "stop", wsName)
	if err != nil {
		return fmt.Errorf("devpod stop failed: %w: %s", err,
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
		return fmt.Errorf("devpod delete failed: %w: %s", err,
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stdout, "Sandbox destroyed.\n")
	return nil
}

// isWorkspaceRunning checks if a DevPod workspace is in the
// Running state by calling devpod status. Used by Create()
// and Start() to determine if a non-zero exit from devpod
// up was a tunnel error (workspace is fine) or a real
// failure (D12).
func (b *DevPodBackend) isWorkspaceRunning(opts Options, wsName string) bool {
	out, err := opts.ExecCmd("devpod", "status", wsName,
		"--output", "json")
	if err != nil {
		return false
	}
	var status devpodStatusOutput
	if err := json.Unmarshal(out, &status); err != nil {
		return false
	}
	return strings.EqualFold(status.State, "Running")
}

// startServerViaSSH starts the OpenCode server inside a
// DevPod workspace via SSH. This is a fallback for
// workspaces created before the postStartCommand was added
// to the devcontainer template, or when postStartCommand
// failed to execute (D13).
//
// Uses --start-services=false to disable port forwarding
// (which would hang if DevPod already has a tunnel open on
// port 4096) and --command for clean command execution.
//
// Injection safety: wsName is derived from opts.ProjectDir
// via projectName(), which sanitizes to [a-z0-9-]. The guard
// below enforces this invariant at the shell-execution boundary
// so that future changes to wsName derivation cannot silently
// reintroduce an injection vector (Principle V).
func (b *DevPodBackend) startServerViaSSH(opts Options, wsName string) error {
	if !safeWSNameRe.MatchString(wsName) {
		return fmt.Errorf("unsafe workspace name %q: "+
			"must match [a-z0-9][a-z0-9-]*[a-z0-9]", wsName)
	}
	_, err := opts.ExecCmd("devpod", "ssh", wsName,
		"--start-services=false",
		"--workdir", "/",
		"--command",
		"cd /workspaces/"+wsName+
			" && nohup opencode serve --port 4096"+
			" > /tmp/opencode-server.log 2>&1 &")
	return err
}

// attachOrDetach handles the detach/attach decision after
// a successful workspace create or start. If --detach is
// set, prints the server URL and returns. Otherwise,
// attaches the TUI.
func (b *DevPodBackend) attachOrDetach(opts Options) error {
	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)
	if opts.Detach {
		fmt.Fprintf(opts.Stdout,
			"Sandbox ready (detached).\nServer: %s\n", serverURL)
		return nil
	}
	fmt.Fprintf(opts.Stderr, "Attaching to sandbox...\n")
	return b.Attach(opts)
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
