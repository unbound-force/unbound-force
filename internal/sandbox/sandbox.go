package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Options configures sandbox operations. All external
// dependencies are injected as function fields for
// testability per Constitution Principle IV.
type Options struct {
	// ProjectDir is the project directory to mount into
	// the container. Defaults to current working directory.
	ProjectDir string

	// Mode is the mount mode: "isolated" (read-only with
	// overlay) or "direct" (read-write). Default: "isolated".
	Mode string

	// Detach skips auto-attach after container start.
	// When true, prints the server URL and exits.
	Detach bool

	// Yes skips confirmation prompts (for extract).
	Yes bool

	// NoParent disables parent directory mounting. When
	// false (default), the project's parent directory is
	// mounted at /workspace so sibling repos are
	// accessible via relative paths (e.g., ../dewey).
	// When true, only the project directory is mounted
	// (pre-Spec 034 behavior).
	NoParent bool

	// Image is the container image to use.
	// Default: "quay.io/unbound-force/opencode-dev:latest".
	// Overridden by UF_SANDBOX_IMAGE env var or --image flag.
	Image string

	// Memory is the container memory limit (e.g., "8g").
	Memory string

	// CPUs is the container CPU limit (e.g., "4").
	CPUs string

	// Stdout is the writer for user-facing output.
	Stdout io.Writer

	// Stderr is the writer for progress/status messages.
	Stderr io.Writer

	// Stdin is the reader for user input (confirmation prompts).
	Stdin io.Reader

	// LookPath finds a binary in PATH.
	LookPath func(string) (string, error)

	// ExecCmd runs a command and returns combined output.
	ExecCmd func(name string, args ...string) ([]byte, error)

	// ExecInteractive runs a command with stdin/stdout/stderr
	// connected to the terminal. Used for `opencode attach`
	// which requires interactive I/O.
	ExecInteractive func(name string, args ...string) error

	// Getenv reads an environment variable.
	Getenv func(string) string

	// ReadFile reads a file's contents.
	ReadFile func(string) ([]byte, error)

	// HTTPGet performs an HTTP GET request and returns the
	// status code. Used for health check polling.
	HTTPGet func(url string) (int, error)

	// --- New fields for Spec 029 ---

	// BackendName selects the backend: "auto", "podman",
	// or "che". Default: "auto" (auto-detect).
	BackendName string

	// WorkspaceName overrides the auto-generated workspace
	// name. Default: "uf-sandbox-<project-name>".
	WorkspaceName string

	// DemoPorts lists additional ports to expose for demos
	// (Podman only). Merged with config file ports.
	DemoPorts []int

	// ConfigPath is the path to .uf/sandbox.yaml.
	// Default: "<ProjectDir>/.uf/sandbox.yaml".
	ConfigPath string

	// CheURL is the Eclipse Che server URL. Overrides
	// config file. Can also be set via UF_CHE_URL env var.
	CheURL string

	// HTTPDo performs an HTTP request and returns the
	// response. Used for CDE REST API calls. Defaults to
	// http.DefaultClient.Do.
	HTTPDo func(req *http.Request) (*http.Response, error)

	// UIDMap enables explicit UID/GID mapping via
	// --uidmap/--gidmap flags instead of --userns=keep-id.
	// Use on macOS when the Podman machine's virtiofs does
	// not support keep-id UID mapping.
	UIDMap bool

	// Platform overrides the auto-detected PlatformConfig.
	// When non-nil, Start() uses this instead of calling
	// DetectPlatform(). Allows tests to inject macOS
	// platform properties on Linux CI.
	Platform *PlatformConfig
}

// ContainerStatus represents the current state of the
// sandbox container, parsed from `podman inspect` output.
type ContainerStatus struct {
	// Running is true when the container is active.
	Running bool

	// Name is the container name (always "uf-sandbox").
	Name string

	// ID is the container ID (short form).
	ID string

	// Image is the container image used.
	Image string

	// Mode is the mount mode ("isolated" or "direct").
	// Determined by inspecting the volume mount flags.
	Mode string

	// ProjectDir is the mounted project directory.
	ProjectDir string

	// ServerURL is the OpenCode server URL.
	ServerURL string

	// StartedAt is the container start time.
	StartedAt string

	// ExitCode is set when the container has stopped.
	// -1 when the container is running.
	ExitCode int
}

// PatchSummary describes changes available for extraction
// from the sandbox container.
type PatchSummary struct {
	// CommitCount is the number of commits since the
	// mount point (origin/HEAD).
	CommitCount int

	// FilesChanged is the number of files modified.
	FilesChanged int

	// Insertions is the total lines added.
	Insertions int

	// Deletions is the total lines removed.
	Deletions int

	// Patch is the raw patch content (format-patch output).
	Patch string

	// StatOutput is the human-readable diffstat.
	StatOutput string
}

// defaults fills zero-value fields with production implementations.
func (o *Options) defaults() {
	if o.ProjectDir == "" {
		o.ProjectDir, _ = os.Getwd()
	}
	if o.Mode == "" {
		o.Mode = ModeIsolated
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
	if o.Stdin == nil {
		o.Stdin = os.Stdin
	}
	if o.LookPath == nil {
		o.LookPath = exec.LookPath
	}
	if o.ExecCmd == nil {
		o.ExecCmd = defaultExecCmd
	}
	if o.ExecInteractive == nil {
		o.ExecInteractive = defaultExecInteractive
	}
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.ReadFile == nil {
		o.ReadFile = os.ReadFile
	}
	if o.HTTPGet == nil {
		o.HTTPGet = defaultHTTPGet
	}
	if o.HTTPDo == nil {
		o.HTTPDo = http.DefaultClient.Do
	}
}

// defaultExecCmd is the production implementation of ExecCmd.
func defaultExecCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// defaultExecInteractive runs a command with stdin/stdout/stderr
// connected to the terminal for interactive TUI commands.
func defaultExecInteractive(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// defaultHTTPGet performs an HTTP GET and returns the status code.
func defaultHTTPGet(url string) (int, error) {
	resp, err := http.Get(url) //nolint:gosec // URL is constructed internally, not user-supplied
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, nil
}

// GatewayDefaultPort is the default port for the LLM gateway.
// Matches gateway.DefaultPort — duplicated here to avoid a
// circular import between sandbox and gateway packages.
const GatewayDefaultPort = 53147

// gatewayHealthCheck checks if a gateway is already running
// on the given port by probing its health endpoint. Returns
// true if the health endpoint responds with HTTP 200.
func gatewayHealthCheck(httpGet func(string) (int, error), port int) bool {
	url := fmt.Sprintf("http://localhost:%d/health", port)
	code, err := httpGet(url)
	return err == nil && code == http.StatusOK
}

// autoStartGateway detects a cloud provider from environment
// variables and starts the gateway if one is found. Returns
// the port the gateway is listening on, whether the gateway
// is active, and any error.
//
// Detection priority (same as gateway.DetectProvider):
//  1. CLAUDE_CODE_USE_VERTEX=1 + ANTHROPIC_VERTEX_PROJECT_ID → Vertex
//  2. CLAUDE_CODE_USE_BEDROCK=1 → Bedrock
//  3. ANTHROPIC_API_KEY → Anthropic
//  4. None → no gateway (direct API key usage)
//
// If a gateway is already running (health check passes),
// it is reused without starting a new one (FR-010).
func autoStartGateway(opts Options) (int, bool, error) {
	port := GatewayDefaultPort

	// Detect provider from env vars.
	hasVertex := opts.Getenv("CLAUDE_CODE_USE_VERTEX") == "1" &&
		opts.Getenv("ANTHROPIC_VERTEX_PROJECT_ID") != ""
	hasBedrock := opts.Getenv("CLAUDE_CODE_USE_BEDROCK") == "1"
	hasAnthropic := opts.Getenv("ANTHROPIC_API_KEY") != ""

	if !hasVertex && !hasBedrock && !hasAnthropic {
		// No provider detected — direct API key usage.
		return 0, false, nil
	}

	// Check if a gateway is already running.
	if gatewayHealthCheck(opts.HTTPGet, port) {
		return port, true, nil
	}

	// Start the gateway in detached mode.
	_, err := opts.ExecCmd("uf", "gateway", "--detach",
		"--port", fmt.Sprintf("%d", port))
	if err != nil {
		return 0, false, fmt.Errorf("failed to start gateway: %w", err)
	}

	// Wait for the gateway health endpoint with exponential
	// backoff (same pattern as waitForHealth).
	deadline := time.Now().Add(10 * time.Second)
	interval := 200 * time.Millisecond
	maxInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		if gatewayHealthCheck(opts.HTTPGet, port) {
			return port, true, nil
		}
		time.Sleep(interval)
		if interval < maxInterval {
			interval *= 2
			if interval > maxInterval {
				interval = maxInterval
			}
		}
	}

	return 0, false, fmt.Errorf(
		"gateway started but health check timed out on port %d", port)
}

// isContainerRunning checks if a container named uf-sandbox
// exists and is in the running state.
func isContainerRunning(opts Options) (bool, error) {
	out, err := opts.ExecCmd("podman", "inspect",
		"--format", "{{.State.Running}}", ContainerName)
	if err != nil {
		// Container does not exist — not an error, just not running.
		return false, nil
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// isContainerExists checks if a container named uf-sandbox
// exists (running or stopped).
func isContainerExists(opts Options) bool {
	_, err := opts.ExecCmd("podman", "inspect", ContainerName)
	return err == nil
}

// waitForHealth polls the OpenCode server health endpoint
// with exponential backoff until it responds or the timeout
// expires. Initial interval: 500ms, doubling to 5s max.
// Total timeout per FR-005: 60 seconds.
func waitForHealth(opts Options, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	interval := 500 * time.Millisecond
	maxInterval := 5 * time.Second
	url := fmt.Sprintf("http://localhost:%d", DefaultServerPort)

	for time.Now().Before(deadline) {
		code, err := opts.HTTPGet(url)
		if err == nil && code == 200 {
			return nil
		}
		time.Sleep(interval)
		if interval < maxInterval {
			interval *= 2
			if interval > maxInterval {
				interval = maxInterval
			}
		}
	}
	return fmt.Errorf(
		"health check timed out after %s. Check container logs: podman logs %s",
		timeout, ContainerName)
}

// Create provisions a persistent sandbox workspace using the
// selected backend. Resolves the backend from flags, env, config,
// or auto-detection, then delegates to backend.Create().
// After successful creation, auto-attaches the TUI unless
// --detach is set.
func Create(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	backend, err := ResolveBackend(opts)
	if err != nil {
		return err
	}

	if err := backend.Create(opts); err != nil {
		return err
	}

	// Display demo endpoint URLs after create.
	ws, err := backend.Status(opts)
	if err == nil && len(ws.DemoEndpoints) > 0 {
		fmt.Fprintf(opts.Stderr, "Demo endpoints:\n")
		for _, ep := range ws.DemoEndpoints {
			fmt.Fprintf(opts.Stderr, "  %s: %s\n", ep.Name, ep.URL)
		}
	}

	// Auto-attach unless --detach.
	if opts.Detach {
		serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)
		if ws.ServerURL != "" {
			serverURL = ws.ServerURL
		}
		fmt.Fprintf(opts.Stdout, "Sandbox created (detached).\nServer: %s\n", serverURL)
		return nil
	}

	return backend.Attach(opts)
}

// Destroy permanently deletes the sandbox workspace and all
// associated state. Resolves the backend and delegates to
// backend.Destroy(). Idempotent.
func Destroy(opts Options) error {
	opts.defaults()

	backend, err := ResolveBackend(opts)
	if err != nil {
		return err
	}

	return backend.Destroy(opts)
}

// isPersistentWorkspace checks if a persistent workspace
// exists for the current project by looking for a named
// volume.
func isPersistentWorkspace(opts Options) bool {
	volName := volumeNameForProject(opts.ProjectDir)
	_, err := opts.ExecCmd("podman", "volume", "inspect", volName)
	return err == nil
}

// Start launches a sandbox container with the project directory
// mounted. Checks prerequisites (Podman, OpenCode), detects
// platform, pulls the image if needed, starts the container,
// waits for the health check, and attaches the TUI (unless
// Detach is true).
//
// For persistent workspaces (created via `uf sandbox create`),
// Start resumes the existing workspace. For ephemeral mode
// (no prior create), Start uses the Spec 028 behavior.
func Start(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	// Persistent workspace detection: if a named volume
	// exists for this project, delegate to the backend's
	// Start method to resume the persistent workspace.
	if isPersistentWorkspace(opts) {
		backend, err := ResolveBackend(opts)
		if err != nil {
			return err
		}
		return backend.Start(opts)
	}

	// --- Ephemeral mode (Spec 028 backward compatibility) ---

	// FR-001: Verify podman is in PATH.
	if _, err := opts.LookPath("podman"); err != nil {
		return fmt.Errorf(
			"podman not found. Install: brew install podman or https://podman.io")
	}

	// Podman version check: --userns=keep-id:uid=N,gid=N
	// requires Podman >= 4.3.
	major, minor, verErr := parsePodmanVersion(opts)
	if verErr != nil {
		return fmt.Errorf("podman version check: %w", verErr)
	}
	if major < 4 || (major == 4 && minor < 3) {
		return fmt.Errorf(
			"podman >= 4.3 required for --userns=keep-id:uid=N,gid=N (current: %d.%d)",
			major, minor)
	}

	// Rootless check: --uidmap is only safe under rootless
	// Podman. Rootful Podman uses different UID semantics.
	if opts.UIDMap && !isRootlessPodman(opts) {
		return fmt.Errorf(
			"--uidmap is only safe under rootless Podman")
	}

	// Verify opencode is in PATH when attach is needed.
	if !opts.Detach {
		if _, err := opts.LookPath("opencode"); err != nil {
			return fmt.Errorf(
				"opencode not found. Install: brew install anomalyco/tap/opencode")
		}
	}

	// FR-016: Check for already-running container.
	running, err := isContainerRunning(opts)
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if running {
		return fmt.Errorf(
			"sandbox already running, use `uf sandbox attach` or `uf sandbox stop` first")
	}

	// Clean up dead (stopped) container before starting a new one.
	if isContainerExists(opts) {
		fmt.Fprintf(opts.Stderr, "Removing stopped sandbox container...\n")
		_, _ = opts.ExecCmd("podman", "rm", ContainerName)
	}

	// FR-003: Pull image if not cached.
	if _, err := opts.ExecCmd("podman", "image", "exists", opts.Image); err != nil {
		fmt.Fprintf(opts.Stderr, "Pulling image %s...\n", opts.Image)
		if out, pullErr := opts.ExecCmd("podman", "pull", opts.Image); pullErr != nil {
			return fmt.Errorf("failed to pull image %s: %s", opts.Image, string(out))
		}
	}

	// Ollama warning: when OLLAMA_HOST is empty and Ollama is
	// not detected, print a warning but continue.
	if opts.Getenv("OLLAMA_HOST") == "" {
		if _, err := opts.LookPath("ollama"); err != nil {
			fmt.Fprintf(opts.Stderr,
				"Warning: Ollama not detected. AI features requiring local models may not work.\n")
		}
	}

	// Auto-start gateway for cloud provider credential
	// isolation (FR-010, FR-012). The gateway runs on the
	// host and proxies LLM requests so the container does
	// not need credential files or API keys.
	gatewayPort, gatewayActive, gwErr := autoStartGateway(opts)
	if gwErr != nil {
		// Gateway start failed — cloud provider LLM access
		// will be unavailable inside the container.
		fmt.Fprintf(opts.Stderr,
			"Gateway start failed (%v) — cloud provider LLM access unavailable\n", gwErr)
	}
	if gatewayActive {
		fmt.Fprintf(opts.Stderr,
			"Gateway active on port %d — credentials proxied\n", gatewayPort)
	}

	// Detect platform for volume flags. Use injected
	// Platform when provided (for test overrides), otherwise
	// detect from the host environment.
	var platform PlatformConfig
	if opts.Platform != nil {
		platform = *opts.Platform
	} else {
		platform = DetectPlatform(opts)
	}

	// macOS UID mapping guard: if the Podman machine does
	// not support keep-id and the user has not opted into
	// explicit --uidmap, return an actionable error.
	if platform.OS == "darwin" && !platform.UIDMapSupported && !opts.UIDMap {
		return fmt.Errorf(
			"podman machine UID mapping not supported.\n\n" +
				"Your Podman machine's virtiofs does not support --userns=keep-id UID mapping.\n" +
				"Fix the Podman machine (recommended):\n" +
				"  podman machine stop\n" +
				"  podman machine rm\n" +
				"  podman machine init --rootful=false\n" +
				"  podman machine start\n\n" +
				"Or use the --uidmap flag as a workaround:\n" +
				"  uf sandbox start --uidmap")
	}

	// Build and execute podman run.
	args := buildRunArgs(opts, platform, gatewayActive, gatewayPort)
	if out, err := opts.ExecCmd("podman", args...); err != nil {
		return fmt.Errorf("failed to start container: %s", strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stderr, "Waiting for OpenCode server...\n")

	// FR-005: Wait for health check.
	if err := waitForHealth(opts, HealthTimeout); err != nil {
		return err
	}

	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)

	// FR-007: Detach mode — print URL and return.
	if opts.Detach {
		fmt.Fprintf(opts.Stdout, "Sandbox started (detached).\nServer: %s\n", serverURL)
		return nil
	}

	// FR-006: Attach TUI.
	fmt.Fprintf(opts.Stderr, "Attaching to sandbox...\n")
	if err := opts.ExecInteractive("opencode", "attach", serverURL); err != nil {
		return fmt.Errorf(
			"failed to attach: %v. Connect manually: opencode attach %s", err, serverURL)
	}

	return nil
}

// Stop stops and removes the sandbox container.
// Returns nil if no container is running (idempotent).
//
// For persistent workspaces, stops the container but
// preserves the named volume. For ephemeral mode,
// stops and removes the container (Spec 028 behavior).
func Stop(opts Options) error {
	opts.defaults()

	// Persistent workspace: delegate to backend.Stop()
	// which preserves the named volume.
	if isPersistentWorkspace(opts) {
		backend, err := ResolveBackend(opts)
		if err != nil {
			return err
		}
		return backend.Stop(opts)
	}

	// --- Ephemeral mode (Spec 028 backward compatibility) ---

	// Check if container exists at all.
	if !isContainerExists(opts) {
		fmt.Fprintf(opts.Stdout, "No sandbox to stop.\n")
		return nil
	}

	// Stop the container (ignore error if already stopped).
	_, _ = opts.ExecCmd("podman", "stop", ContainerName)

	// Remove the container.
	if out, err := opts.ExecCmd("podman", "rm", ContainerName); err != nil {
		return fmt.Errorf("failed to remove container: %s", strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stdout, "Sandbox stopped.\n")
	return nil
}

// Attach connects the TUI to the running sandbox's OpenCode
// server via `opencode attach`.
func Attach(opts Options) error {
	opts.defaults()

	// Verify opencode is in PATH.
	if _, err := opts.LookPath("opencode"); err != nil {
		return fmt.Errorf(
			"opencode not found. Install: brew install anomalyco/tap/opencode")
	}

	// Verify container is running.
	running, err := isContainerRunning(opts)
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !running {
		return fmt.Errorf("no sandbox running, run `uf sandbox start`")
	}

	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)
	return opts.ExecInteractive("opencode", "attach", serverURL)
}

// Extract generates a patch from the container's git history,
// presents it for review, and applies it to the host repo on
// confirmation.
//
// For persistent workspaces with git push access, suggests
// using `git pull` instead of extract.
func Extract(opts Options) error {
	opts.defaults()

	// Direct mode: changes are already on the host filesystem.
	if opts.Mode == ModeDirect {
		fmt.Fprintf(opts.Stdout,
			"Sandbox is in direct mode — changes are already on the host filesystem.\n")
		return nil
	}

	// For persistent CDE workspaces, suggest git pull.
	if isPersistentWorkspace(opts) {
		// Check if this is a CDE workspace (has Che URL configured).
		cheURL := resolveCheURL(opts)
		if cheURL != "" {
			fmt.Fprintf(opts.Stdout,
				"This workspace has git push access. Use `git pull` on the host instead of extract.\n")
			return nil
		}
	}

	// Verify container is running.
	running, err := isContainerRunning(opts)
	if err != nil {
		return fmt.Errorf("check container: %w", err)
	}
	if !running {
		return fmt.Errorf("no sandbox running")
	}

	// Count commits since mount point.
	logOut, err := opts.ExecCmd("podman", "exec", ContainerName,
		"git", "-C", "/workspace", "log", "--oneline", "origin/HEAD..HEAD")
	if err != nil || strings.TrimSpace(string(logOut)) == "" {
		fmt.Fprintf(opts.Stdout, "No changes to extract.\n")
		return nil
	}

	commitLines := strings.Split(strings.TrimSpace(string(logOut)), "\n")
	commitCount := len(commitLines)

	// Generate patch.
	patchOut, err := opts.ExecCmd("podman", "exec", ContainerName,
		"git", "-C", "/workspace", "format-patch", "origin/HEAD..HEAD", "--stdout")
	if err != nil {
		return fmt.Errorf("failed to generate patch: %v", err)
	}
	patch := string(patchOut)

	// Display patch summary.
	fmt.Fprintf(opts.Stdout, "\nPatch Summary:\n")
	fmt.Fprintf(opts.Stdout, "  Commits: %d\n", commitCount)
	for _, line := range commitLines {
		fmt.Fprintf(opts.Stdout, "    %s\n", line)
	}
	fmt.Fprintf(opts.Stdout, "\n")

	// Confirmation prompt.
	if !opts.Yes {
		fmt.Fprintf(opts.Stdout, "Apply this patch to the host repository? [y/N] ")
		var response string
		if _, err := fmt.Fscanln(opts.Stdin, &response); err != nil || !isYes(response) {
			fmt.Fprintf(opts.Stdout, "Patch not applied.\n")
			return nil
		}
	}

	// Write patch to temp file and apply via git am.
	tmpFile, err := os.CreateTemp("", "uf-sandbox-*.patch")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(patch); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write patch: %w", err)
	}
	_ = tmpFile.Close()

	if out, err := opts.ExecCmd("git", "am", tmpFile.Name()); err != nil {
		return fmt.Errorf(
			"patch conflict: %s\nrun `git am --abort` to undo",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stdout, "Patch applied successfully (%d commits).\n", commitCount)
	return nil
}

// WorkspaceStatusCheck returns the workspace status for
// persistent workspaces. Returns a zero-value WorkspaceStatus
// if no persistent workspace exists.
func WorkspaceStatusCheck(opts Options) (WorkspaceStatus, error) {
	opts.defaults()

	if !isPersistentWorkspace(opts) {
		return WorkspaceStatus{}, nil
	}

	backend, err := ResolveBackend(opts)
	if err != nil {
		return WorkspaceStatus{}, err
	}

	return backend.Status(opts)
}

// Status returns the current state of the sandbox container.
// Returns a zero-value ContainerStatus with Running=false if
// no container exists.
func Status(opts Options) (ContainerStatus, error) {
	opts.defaults()

	out, err := opts.ExecCmd("podman", "inspect", ContainerName)
	if err != nil {
		// Container does not exist.
		return ContainerStatus{}, nil
	}

	// podman inspect returns a JSON array.
	var inspectData []podmanInspect
	if err := json.Unmarshal(out, &inspectData); err != nil {
		return ContainerStatus{}, fmt.Errorf("parse inspect output: %w", err)
	}
	if len(inspectData) == 0 {
		return ContainerStatus{}, nil
	}

	info := inspectData[0]

	// Determine mode and project directory from volume mounts.
	mode := ModeIsolated
	projectDir := ""
	for _, m := range info.Mounts {
		if m.Destination == "/workspace" {
			projectDir = m.Source
			if m.RW {
				mode = ModeDirect
			}
			break
		}
	}

	exitCode := -1
	if !info.State.Running {
		exitCode = info.State.ExitCode
	}

	// Truncate container ID to short form (12 chars).
	shortID := info.ID
	if len(shortID) > 12 {
		shortID = shortID[:12]
	}

	return ContainerStatus{
		Running:    info.State.Running,
		Name:       info.Name,
		ID:         shortID,
		Image:      info.ImageName,
		Mode:       mode,
		ProjectDir: projectDir,
		ServerURL:  fmt.Sprintf("http://localhost:%d", DefaultServerPort),
		StartedAt:  info.State.StartedAt,
		ExitCode:   exitCode,
	}, nil
}

// FormatStatus writes a human-readable status report to the writer.
func FormatStatus(w io.Writer, s ContainerStatus) {
	if !s.Running {
		fmt.Fprintf(w, "No sandbox running.\n")
		return
	}
	fmt.Fprintf(w, "Sandbox Status\n")
	fmt.Fprintf(w, "  Container:  %s (%s)\n", s.Name, s.ID)
	fmt.Fprintf(w, "  Image:      %s\n", s.Image)
	fmt.Fprintf(w, "  Mode:       %s\n", s.Mode)
	fmt.Fprintf(w, "  Project:    %s\n", s.ProjectDir)
	fmt.Fprintf(w, "  Server:     %s\n", s.ServerURL)
	if s.StartedAt != "" {
		fmt.Fprintf(w, "  Started:    %s\n", s.StartedAt)
	}
}

// isYes returns true if the response is a yes confirmation.
func isYes(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "y" || s == "yes"
}

// podmanInspect is the subset of podman inspect JSON output
// that we parse. Only the fields we need are included.
type podmanInspect struct {
	ID        string `json:"Id"`
	Name      string `json:"Name"`
	ImageName string `json:"ImageName"`
	State     struct {
		Running   bool   `json:"Running"`
		StartedAt string `json:"StartedAt"`
		ExitCode  int    `json:"ExitCode"`
	} `json:"State"`
	Mounts []struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		RW          bool   `json:"RW"`
	} `json:"Mounts"`
}
