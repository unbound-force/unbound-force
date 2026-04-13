package sandbox

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// PodmanBackend implements Backend for local Podman
// containers with named volumes for persistent state.
type PodmanBackend struct{}

// Name returns the backend identifier.
func (b *PodmanBackend) Name() string { return BackendPodman }

// Create provisions a persistent Podman workspace with
// named volumes. Seeds the workspace with the project's
// source code.
//
// Steps:
//  1. Check podman is in PATH
//  2. Verify no workspace exists (named volume check)
//  3. Create named volume: uf-sandbox-<project-name>
//  4. Start container with named volume mount
//  5. Copy project source into workspace
//  6. Wait for health check
//
// Returns error if workspace already exists.
func (b *PodmanBackend) Create(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	// Verify podman is in PATH.
	if _, err := opts.LookPath("podman"); err != nil {
		return fmt.Errorf(
			"podman not found. Install: brew install podman")
	}

	volName := volumeNameForProject(opts.ProjectDir)
	ctrName := containerNameForProject(opts.ProjectDir)

	// Check if workspace already exists.
	if _, err := opts.ExecCmd("podman", "volume", "inspect", volName); err == nil {
		proj := projectName(opts.ProjectDir)
		return fmt.Errorf(
			"sandbox already exists for %s, use `uf sandbox start` or `uf sandbox destroy` first",
			proj)
	}

	// Create named volume.
	if out, err := opts.ExecCmd("podman", "volume", "create", volName); err != nil {
		return fmt.Errorf("failed to create volume: %s", strings.TrimSpace(string(out)))
	}

	// Build container run args with named volume.
	platform := DetectPlatform(opts)
	args := buildPersistentRunArgs(opts, platform, ctrName, volName)

	if out, err := opts.ExecCmd("podman", args...); err != nil {
		// Partial failure cleanup: remove volume.
		_, _ = opts.ExecCmd("podman", "rm", "-f", ctrName)
		_, _ = opts.ExecCmd("podman", "volume", "rm", volName)
		return fmt.Errorf("failed to start container: %s",
			strings.TrimSpace(string(out)))
	}

	// Copy project source into the named volume.
	src := filepath.Join(opts.ProjectDir, ".")
	dst := ctrName + ":/workspace/"
	if out, err := opts.ExecCmd("podman", "cp", src, dst); err != nil {
		// Partial failure cleanup.
		_, _ = opts.ExecCmd("podman", "rm", "-f", ctrName)
		_, _ = opts.ExecCmd("podman", "volume", "rm", volName)
		return fmt.Errorf("failed to copy source: %s",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stderr, "Waiting for OpenCode server...\n")

	// Wait for health check.
	if err := waitForHealth(opts, HealthTimeout); err != nil {
		// Partial failure cleanup.
		_, _ = opts.ExecCmd("podman", "rm", "-f", ctrName)
		_, _ = opts.ExecCmd("podman", "volume", "rm", volName)
		return err
	}

	// Set up git sync (best-effort).
	_ = setupGitSync(opts)

	fmt.Fprintf(opts.Stderr, "Sandbox created: %s\n", ctrName)
	return nil
}

// Start resumes a stopped persistent workspace.
// If no persistent workspace exists, falls back to
// ephemeral mode (Spec 028 behavior via the top-level
// Start function).
func (b *PodmanBackend) Start(opts Options) error {
	opts.defaults()
	opts = DefaultConfig(opts)

	ctrName := containerNameForProject(opts.ProjectDir)

	// Check if container exists (stopped).
	if _, err := opts.ExecCmd("podman", "inspect", ctrName); err != nil {
		// No container — fall back to ephemeral mode.
		// This shouldn't normally happen since we check
		// isPersistentWorkspace before calling this, but
		// handle it defensively.
		return fmt.Errorf("no persistent workspace found for %s",
			projectName(opts.ProjectDir))
	}

	// Start the existing container.
	if out, err := opts.ExecCmd("podman", "start", ctrName); err != nil {
		return fmt.Errorf("failed to start workspace: %s",
			strings.TrimSpace(string(out)))
	}

	fmt.Fprintf(opts.Stderr, "Waiting for OpenCode server...\n")

	if err := waitForHealth(opts, HealthTimeout); err != nil {
		return err
	}

	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)

	if opts.Detach {
		fmt.Fprintf(opts.Stdout, "Sandbox resumed (detached).\nServer: %s\n", serverURL)
		return nil
	}

	fmt.Fprintf(opts.Stderr, "Attaching to sandbox...\n")
	return b.Attach(opts)
}

// Stop stops a running persistent workspace. The container
// is stopped but the named volume is preserved.
func (b *PodmanBackend) Stop(opts Options) error {
	opts.defaults()

	ctrName := containerNameForProject(opts.ProjectDir)

	// Check if container exists.
	if _, err := opts.ExecCmd("podman", "inspect", ctrName); err != nil {
		fmt.Fprintf(opts.Stdout, "No sandbox to stop.\n")
		return nil
	}

	// Stop the container (preserve volume).
	_, _ = opts.ExecCmd("podman", "stop", ctrName)

	fmt.Fprintf(opts.Stdout, "Sandbox stopped (state preserved).\n")
	return nil
}

// Destroy permanently deletes the workspace, container,
// and named volume. Idempotent.
func (b *PodmanBackend) Destroy(opts Options) error {
	opts.defaults()

	ctrName := containerNameForProject(opts.ProjectDir)
	volName := volumeNameForProject(opts.ProjectDir)

	// Stop and remove container (ignore errors if not exists).
	_, _ = opts.ExecCmd("podman", "stop", ctrName)
	_, _ = opts.ExecCmd("podman", "rm", "-f", ctrName)

	// Remove named volume.
	_, _ = opts.ExecCmd("podman", "volume", "rm", volName)

	fmt.Fprintf(opts.Stdout, "Sandbox destroyed.\n")
	return nil
}

// Status returns the current state of the persistent
// workspace. Returns a zero-value WorkspaceStatus if no
// workspace exists.
func (b *PodmanBackend) Status(opts Options) (WorkspaceStatus, error) {
	opts.defaults()

	ctrName := containerNameForProject(opts.ProjectDir)
	volName := volumeNameForProject(opts.ProjectDir)

	// Check if volume exists.
	if _, err := opts.ExecCmd("podman", "volume", "inspect", volName); err != nil {
		return WorkspaceStatus{}, nil
	}

	ws := WorkspaceStatus{
		Exists:     true,
		Backend:    BackendPodman,
		Name:       ctrName,
		Persistent: true,
		Mode:       opts.Mode,
		ProjectDir: opts.ProjectDir,
		ServerURL:  fmt.Sprintf("http://localhost:%d", DefaultServerPort),
	}

	// Try to get container details.
	out, err := opts.ExecCmd("podman", "inspect", ctrName)
	if err == nil {
		var inspectData []podmanInspect
		if err := json.Unmarshal(out, &inspectData); err == nil && len(inspectData) > 0 {
			info := inspectData[0]
			ws.Running = info.State.Running
			ws.Image = info.ImageName
			ws.StartedAt = info.State.StartedAt
			shortID := info.ID
			if len(shortID) > 12 {
				shortID = shortID[:12]
			}
			ws.ID = shortID
			ws.ExitCode = -1
			if !info.State.Running {
				ws.ExitCode = info.State.ExitCode
			}
		}
	}

	// Populate demo endpoints from config and DemoPorts.
	cfg, _ := LoadConfig(opts)
	allPorts := mergeDemoPorts(opts.DemoPorts, cfg.DemoPorts)
	for _, port := range allPorts {
		ws.DemoEndpoints = append(ws.DemoEndpoints, DemoEndpoint{
			Name:     fmt.Sprintf("port-%d", port),
			Port:     port,
			URL:      fmt.Sprintf("http://localhost:%d", port),
			Protocol: "http",
		})
	}

	return ws, nil
}

// Attach connects the TUI to the running workspace's
// OpenCode server.
func (b *PodmanBackend) Attach(opts Options) error {
	opts.defaults()

	if _, err := opts.LookPath("opencode"); err != nil {
		return fmt.Errorf(
			"opencode not found. Install: brew install anomalyco/tap/opencode")
	}

	serverURL := fmt.Sprintf("http://localhost:%d", DefaultServerPort)
	return opts.ExecInteractive("opencode", "attach", serverURL)
}

// buildPersistentRunArgs assembles the podman run argument
// list for a persistent workspace with a named volume.
func buildPersistentRunArgs(opts Options, platform PlatformConfig, ctrName, volName string) []string {
	args := []string{
		"run", "-d",
		"--name", ctrName,
		"--hostname", ctrName,
		"-p", fmt.Sprintf("%d:%d", DefaultServerPort, DefaultServerPort),
	}

	// Named volume mount (replaces bind mount).
	volMount := fmt.Sprintf("%s:/workspace", volName)
	args = append(args, "-v", volMount)

	// Demo port mappings.
	cfg, _ := LoadConfig(opts)
	allPorts := mergeDemoPorts(opts.DemoPorts, cfg.DemoPorts)
	for _, port := range allPorts {
		args = append(args, "-p", fmt.Sprintf("%d:%d", port, port))
	}

	// Environment variables.
	args = append(args, forwardedEnvVars(opts)...)

	// Google Cloud credential mounts.
	args = append(args, googleCloudCredentialMounts(opts, platform)...)

	// Resource limits.
	args = append(args, "--memory", opts.Memory)
	args = append(args, "--cpus", opts.CPUs)

	// Image name (last argument).
	args = append(args, opts.Image)

	return args
}

// mergeDemoPorts combines CLI flag ports and config file
// ports, deduplicating.
func mergeDemoPorts(flagPorts, configPorts []int) []int {
	seen := make(map[int]bool)
	var result []int
	for _, p := range flagPorts {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	for _, p := range configPorts {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}
