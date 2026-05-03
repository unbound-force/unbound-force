package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// --- Helper: mock Options builder ---

// testOpts returns an Options struct with all dependencies
// injected as no-op/success mocks. Tests override specific
// fields to exercise error paths.
//
// Key defaults for backward compatibility:
//   - LookPath finds podman and opencode but NOT devpod
//     (prevents auto-detection of DevPod backend)
//   - ExecCmd returns error for "podman volume inspect"
//     (prevents persistent workspace detection)
func testOpts() Options {
	return Options{
		ProjectDir: "/tmp/test-project",
		Mode:       ModeIsolated,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		Stdin:      strings.NewReader(""),
		LookPath: func(name string) (string, error) {
			// Don't find devpod by default — prevents auto-detect.
			if name == "devpod" {
				return "", fmt.Errorf("not found")
			}
			return "/usr/bin/" + name, nil
		},
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			// Volume inspect fails by default — prevents persistent
			// workspace detection in ephemeral-mode tests.
			if name == "podman" && len(args) > 0 && args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			// Podman version check: return valid version so
			// Start() passes the >= 4.3 gate.
			if name == "podman" && len(args) > 0 && args[0] == "--version" {
				return []byte("podman version 5.0.0\n"), nil
			}
			// Rootless check: return "true" by default so
			// --uidmap tests pass the rootless gate.
			if name == "podman" && len(args) > 0 && args[0] == "info" {
				return []byte("true\n"), nil
			}
			return []byte(""), nil
		},
		ExecInteractive: func(name string, args ...string) error { return nil },
		Getenv:          func(key string) string { return "" },
		ReadFile:        func(path string) ([]byte, error) { return nil, fmt.Errorf("not found") },
		HTTPGet:         func(url string) (int, error) { return 200, nil },
	}
}

// stdout returns the captured stdout content from test Options.
func stdout(opts Options) string {
	return opts.Stdout.(*bytes.Buffer).String()
}

// --- DetectPlatform tests ---

func TestDetectPlatform_MacOSArm64(t *testing.T) {
	opts := testOpts()
	// On macOS, probeUIDMapping calls ExecCmd with podman run.
	// On Linux, getenforce may be called. Allow both.
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		// Allow the UID mapping probe (podman run --rm ...).
		if name == "podman" && len(args) > 0 && args[0] == "run" {
			return []byte("1000\n"), nil
		}
		// Allow getenforce on Linux.
		if name == "getenforce" {
			return []byte("Disabled\n"), nil
		}
		t.Fatalf("unexpected ExecCmd call: %s %v", name, args)
		return nil, nil
	}

	p := DetectPlatform(opts)

	// On macOS (where tests run), SELinux is always false.
	if p.SELinux {
		t.Error("expected SELinux=false on macOS")
	}
	if p.OS == "" {
		t.Error("expected OS to be set")
	}
	if p.Arch == "" {
		t.Error("expected Arch to be set")
	}
}

func TestDetectPlatform_FedoraSELinux(t *testing.T) {
	// This test can only verify the logic path on Linux.
	// On macOS, DetectPlatform returns early before checking
	// SELinux. We test the SELinux detection logic directly.
	opts := testOpts()
	opts.ReadFile = func(path string) ([]byte, error) {
		if path == "/etc/selinux/config" {
			return []byte("SELINUX=enforcing\nSELINUXTYPE=targeted\n"), nil
		}
		return nil, fmt.Errorf("not found")
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "getenforce" {
			return []byte("Enforcing\n"), nil
		}
		return nil, fmt.Errorf("unknown command")
	}

	p := DetectPlatform(opts)

	// On macOS test host, the function returns early before
	// checking SELinux. We verify the function doesn't crash.
	// The SELinux path is tested via the config builder tests.
	if p.OS == "linux" && !p.SELinux {
		t.Error("expected SELinux=true on Linux with enforcing config")
	}
}

func TestDetectPlatform_FedoraNoSELinux(t *testing.T) {
	opts := testOpts()
	opts.ReadFile = func(path string) ([]byte, error) {
		if path == "/etc/selinux/config" {
			return []byte("SELINUX=disabled\nSELINUXTYPE=targeted\n"), nil
		}
		return nil, fmt.Errorf("not found")
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "getenforce" {
			return []byte("Disabled\n"), nil
		}
		return nil, fmt.Errorf("unknown command")
	}

	p := DetectPlatform(opts)

	if p.SELinux {
		t.Error("expected SELinux=false when disabled")
	}
}

// --- config.go tests ---

func TestBuildRunArgs_Isolated(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)

	joined := strings.Join(args, " ")

	// Verify key flags are present.
	if !strings.Contains(joined, "--name uf-sandbox") {
		t.Errorf("expected --name uf-sandbox, got: %s", joined)
	}
	if !strings.Contains(joined, "-p 4096:4096") {
		t.Errorf("expected -p 4096:4096, got: %s", joined)
	}
	if !strings.Contains(joined, "--memory 8g") {
		t.Errorf("expected --memory 8g, got: %s", joined)
	}
	if !strings.Contains(joined, "--cpus 4") {
		t.Errorf("expected --cpus 4, got: %s", joined)
	}
	// Verify parent directory mounted with :ro for isolated mode.
	if !strings.Contains(joined, "/tmp:/workspace:ro") {
		t.Errorf("expected parent mount /tmp:/workspace:ro, got: %s", joined)
	}
	// Verify workdir set to project subdirectory.
	if !strings.Contains(joined, "--workdir /workspace/test-project") {
		t.Errorf("expected --workdir /workspace/test-project, got: %s", joined)
	}
	// Verify WORKSPACE env var set for entrypoint (FR-044).
	if !strings.Contains(joined, "WORKSPACE=/workspace/test-project") {
		t.Errorf("expected WORKSPACE=/workspace/test-project, got: %s", joined)
	}
	// Verify image is last argument.
	if args[len(args)-1] != DefaultImage {
		t.Errorf("expected image as last arg, got: %s", args[len(args)-1])
	}
}

func TestBuildRunArgs_Direct(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeDirect
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)

	joined := strings.Join(args, " ")

	// Verify parent directory mounted read-write (no :ro) for direct mode.
	if !strings.Contains(joined, "/tmp:/workspace") {
		t.Errorf("expected parent mount /tmp:/workspace, got: %s", joined)
	}
	if strings.Contains(joined, "/tmp:/workspace:ro") {
		t.Errorf("expected no :ro on parent mount for direct mode, got: %s", joined)
	}
	// Verify workdir set to project subdirectory.
	if !strings.Contains(joined, "--workdir /workspace/test-project") {
		t.Errorf("expected --workdir /workspace/test-project, got: %s", joined)
	}
	// Verify WORKSPACE env var set for entrypoint (FR-044).
	if !strings.Contains(joined, "WORKSPACE=/workspace/test-project") {
		t.Errorf("expected WORKSPACE=/workspace/test-project, got: %s", joined)
	}
}

func TestBuildVolumeMounts_NoParentFlag(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeDirect
	opts.NoParent = true

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	mounts := buildVolumeMounts(opts, platform)
	joined := strings.Join(mounts, " ")

	// With --no-parent, project dir mounted directly.
	if !strings.Contains(joined, "/tmp/test-project:/workspace") {
		t.Errorf("expected project-only mount, got: %s", joined)
	}
	// No parent mount.
	if strings.Contains(joined, "/tmp:/workspace") {
		t.Errorf("expected no parent mount with --no-parent, got: %s", joined)
	}
}

func TestBuildRunArgs_NoParentNoWorkdir(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeDirect
	opts.NoParent = true
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	// No --workdir when --no-parent is set.
	if strings.Contains(joined, "--workdir") {
		t.Errorf("expected no --workdir with --no-parent, got: %s", joined)
	}
	// No WORKSPACE env var when --no-parent is set (FR-044).
	if strings.Contains(joined, "WORKSPACE=") {
		t.Errorf("expected no WORKSPACE with --no-parent, got: %s", joined)
	}
}

func TestBuildVolumeMounts_RootFallback(t *testing.T) {
	opts := testOpts()
	opts.ProjectDir = "/myproject"
	opts.Mode = ModeDirect

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	mounts := buildVolumeMounts(opts, platform)
	joined := strings.Join(mounts, " ")

	// Parent of /myproject is /, should fall back to
	// project-only mount.
	if !strings.Contains(joined, "/myproject:/workspace") {
		t.Errorf("expected project-only mount for root parent, got: %s", joined)
	}
}

func TestBuildRunArgs_RootFallbackNoWorkdir(t *testing.T) {
	opts := testOpts()
	opts.ProjectDir = "/myproject"
	opts.Mode = ModeDirect
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	// No --workdir when parent is root (fallback).
	if strings.Contains(joined, "--workdir") {
		t.Errorf("expected no --workdir for root parent fallback, got: %s", joined)
	}
	// No WORKSPACE when parent is root (FR-044).
	if strings.Contains(joined, "WORKSPACE=") {
		t.Errorf("expected no WORKSPACE for root parent fallback, got: %s", joined)
	}
}

func TestBuildVolumeMounts_ParentMountSELinux(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated // SELinux + isolated = :ro,Z

	platform := PlatformConfig{OS: "linux", Arch: "amd64", SELinux: true}
	mounts := buildVolumeMounts(opts, platform)
	joined := strings.Join(mounts, " ")

	// Parent mount with isolated mode + SELinux: :ro,Z.
	if !strings.Contains(joined, "/tmp:/workspace:ro,Z") {
		t.Errorf("expected :ro,Z on parent mount with SELinux, got: %s", joined)
	}
}

func TestBuildRunArgs_SELinux(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "linux", Arch: "amd64", SELinux: true}
	args := buildRunArgs(opts, platform, false, 0)

	joined := strings.Join(args, " ")

	// Verify :Z suffix on volume mount when SELinux is enforcing.
	if !strings.Contains(joined, ",Z") {
		t.Errorf("expected ,Z suffix for SELinux, got: %s", joined)
	}
}

func TestBuildRunArgs_CustomImage(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated
	opts.Image = "my-registry.io/custom-image:v2"
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)

	// Verify custom image is used.
	if args[len(args)-1] != "my-registry.io/custom-image:v2" {
		t.Errorf("expected custom image, got: %s", args[len(args)-1])
	}
}

func TestDefaultConfig_ImagePrecedence(t *testing.T) {
	// Test 1: Flag value takes precedence.
	opts := testOpts()
	opts.Image = "flag-image:latest"
	opts.Getenv = func(key string) string {
		if key == "UF_SANDBOX_IMAGE" {
			return "env-image:latest"
		}
		return ""
	}

	result := DefaultConfig(opts)
	if result.Image != "flag-image:latest" {
		t.Errorf("expected flag image, got: %s", result.Image)
	}

	// Test 2: Env var when no flag.
	opts.Image = ""
	result = DefaultConfig(opts)
	if result.Image != "env-image:latest" {
		t.Errorf("expected env image, got: %s", result.Image)
	}

	// Test 3: Default constant when neither flag nor env.
	opts.Getenv = func(key string) string { return "" }
	opts.Image = ""
	result = DefaultConfig(opts)
	if result.Image != DefaultImage {
		t.Errorf("expected default image, got: %s", result.Image)
	}
}

func TestDefaultConfig_MemoryAndCPUsPrecedence(t *testing.T) {
	// Flag values override defaults.
	opts := testOpts()
	opts.Memory = "16g"
	opts.CPUs = "8"

	result := DefaultConfig(opts)
	if result.Memory != "16g" {
		t.Errorf("expected 16g, got: %s", result.Memory)
	}
	if result.CPUs != "8" {
		t.Errorf("expected 8, got: %s", result.CPUs)
	}

	// Defaults when no flag.
	opts.Memory = ""
	opts.CPUs = ""
	result = DefaultConfig(opts)
	if result.Memory != DefaultMemory {
		t.Errorf("expected default memory, got: %s", result.Memory)
	}
	if result.CPUs != DefaultCPUs {
		t.Errorf("expected default cpus, got: %s", result.CPUs)
	}
}

func TestForwardedEnvVars(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		switch key {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		case "OPENAI_API_KEY":
			return "sk-xxx"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-gcp-project"
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		default:
			return ""
		}
	}

	args := forwardedEnvVars(opts, false)
	joined := strings.Join(args, " ")

	// Verify present API keys are forwarded.
	if !strings.Contains(joined, "-e ANTHROPIC_API_KEY") {
		t.Errorf("expected ANTHROPIC_API_KEY, got: %s", joined)
	}
	if !strings.Contains(joined, "-e OPENAI_API_KEY") {
		t.Errorf("expected OPENAI_API_KEY, got: %s", joined)
	}
	// Verify Vertex-specific vars are forwarded.
	if !strings.Contains(joined, "-e ANTHROPIC_VERTEX_PROJECT_ID") {
		t.Errorf("expected ANTHROPIC_VERTEX_PROJECT_ID, got: %s", joined)
	}
	if !strings.Contains(joined, "-e CLAUDE_CODE_USE_VERTEX") {
		t.Errorf("expected CLAUDE_CODE_USE_VERTEX, got: %s", joined)
	}
	// Verify absent keys are NOT forwarded.
	if strings.Contains(joined, "GEMINI_API_KEY") {
		t.Errorf("expected no GEMINI_API_KEY (not set), got: %s", joined)
	}
	// Verify OLLAMA_HOST is always set.
	if !strings.Contains(joined, "OLLAMA_HOST=host.containers.internal:11434") {
		t.Errorf("expected OLLAMA_HOST override, got: %s", joined)
	}
}

// --- Start() tests ---

func TestStart_PodmanMissing(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		if name == "podman" || name == "devpod" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when podman is missing")
	}
	if !strings.Contains(err.Error(), "podman not found") {
		t.Errorf("expected podman install hint, got: %s", err.Error())
	}
}

func TestStart_AlreadyRunning(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "--version" {
				return []byte("podman version 5.0.0\n"), nil
			}
			if args[0] == "inspect" {
				return []byte("true"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when sandbox is already running")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("expected already running message, got: %s", err.Error())
	}
}

func TestStart_DetachMode(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	interactiveCalled := false

	// podman volume inspect returns error (no persistent workspace).
	// podman inspect returns error (no container).
	// podman image exists returns error (need pull).
	// podman pull succeeds.
	// podman run succeeds.
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return nil, fmt.Errorf("image not found")
			case "pull":
				return []byte("pulled"), nil
			case "run":
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}
	opts.ExecInteractive = func(name string, args ...string) error {
		interactiveCalled = true
		return nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if interactiveCalled {
		t.Error("ExecInteractive should NOT be called in detach mode")
	}
	out := stdout(opts)
	if !strings.Contains(out, "Sandbox started (detached)") {
		t.Errorf("expected detach message, got: %s", out)
	}
	if !strings.Contains(out, "http://localhost:4096") {
		t.Errorf("expected server URL, got: %s", out)
	}
}

func TestStart_IsolatedMount(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, ":ro") {
		t.Errorf("expected :ro for isolated mode, got: %s", joined)
	}
}

func TestStart_DirectMount(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeDirect
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	// Check project mount is read-write (no :ro on project path).
	if strings.Contains(joined, "/tmp/test-project:/workspace:ro") {
		t.Errorf("expected no :ro on project mount for direct mode, got: %s", joined)
	}
}

func TestStart_HealthTimeout(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return []byte(""), nil // image exists
			case "run":
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}
	// HTTPGet always fails — simulates timeout.
	opts.HTTPGet = func(url string) (int, error) {
		return 0, fmt.Errorf("connection refused")
	}

	// Use a very short timeout to avoid slow tests.
	// We test waitForHealth directly with a short timeout.
	err := waitForHealth(opts, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout message, got: %s", err.Error())
	}
}

func TestStart_DeadContainerCleanup(t *testing.T) {
	rmCalled := false
	runCalled := false

	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "inspect":
				if len(args) > 1 && args[1] == "--format" {
					// isContainerRunning: container exists but not running.
					return []byte("false"), nil
				}
				// isContainerExists: container exists.
				return []byte("{}"), nil
			case "rm":
				rmCalled = true
				return []byte(""), nil
			case "image":
				return []byte(""), nil // image exists
			case "run":
				runCalled = true
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rmCalled {
		t.Error("expected podman rm to be called for dead container cleanup")
	}
	if !runCalled {
		t.Error("expected podman run to be called after cleanup")
	}
}

func TestStart_HappyPathWithAttach(t *testing.T) {
	attachCalled := false
	attachArgs := []string{}

	opts := testOpts()
	opts.Detach = false
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return []byte(""), nil // image exists
			case "run":
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}
	opts.ExecInteractive = func(name string, args ...string) error {
		attachCalled = true
		attachArgs = append(attachArgs, name)
		attachArgs = append(attachArgs, args...)
		return nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !attachCalled {
		t.Error("expected ExecInteractive to be called for TUI attach")
	}
	if len(attachArgs) < 3 {
		t.Fatalf("expected at least 3 attach args, got: %v", attachArgs)
	}
	if attachArgs[0] != "opencode" {
		t.Errorf("expected opencode command, got: %s", attachArgs[0])
	}
	if attachArgs[1] != "attach" {
		t.Errorf("expected attach subcommand, got: %s", attachArgs[1])
	}
	if attachArgs[2] != "http://localhost:4096" {
		t.Errorf("expected server URL, got: %s", attachArgs[2])
	}
}

// --- Stop() tests ---

func TestStop_RunningContainer(t *testing.T) {
	stopCalled := false
	rmCalled := false

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "inspect":
				return []byte("{}"), nil // container exists
			case "stop":
				stopCalled = true
				return []byte(""), nil
			case "rm":
				rmCalled = true
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	err := Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Error("expected podman stop to be called")
	}
	if !rmCalled {
		t.Error("expected podman rm to be called")
	}
	if !strings.Contains(stdout(opts), "Sandbox stopped") {
		t.Errorf("expected stopped message, got: %s", stdout(opts))
	}
}

func TestStop_NoContainer(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "inspect" {
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	err := Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout(opts), "No sandbox to stop") {
		t.Errorf("expected no sandbox message, got: %s", stdout(opts))
	}
}

// --- Attach() tests ---

func TestAttach_NoContainer(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "inspect" {
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	err := Attach(opts)
	if err == nil {
		t.Fatal("expected error when no container running")
	}
	if !strings.Contains(err.Error(), "no sandbox running") {
		t.Errorf("expected no sandbox message, got: %s", err.Error())
	}
}

func TestAttach_OpenCodeMissing(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		if name == "opencode" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}

	err := Attach(opts)
	if err == nil {
		t.Fatal("expected error when opencode is missing")
	}
	if !strings.Contains(err.Error(), "opencode not found") {
		t.Errorf("expected opencode install hint, got: %s", err.Error())
	}
}

// --- Extract() tests ---

func TestExtract_NoChanges(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "inspect":
				if len(args) > 1 && args[1] == "--format" {
					return []byte("true"), nil // container running
				}
				return []byte("{}"), nil
			case "exec":
				// git log returns empty (no commits).
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout(opts), "No changes to extract") {
		t.Errorf("expected no changes message, got: %s", stdout(opts))
	}
}

func TestExtract_UserDeclines(t *testing.T) {
	opts := testOpts()
	opts.Stdin = strings.NewReader("n\n")
	gitAmCalled := false

	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "inspect":
				if len(args) > 1 && args[1] == "--format" {
					return []byte("true"), nil
				}
				return []byte("{}"), nil
			case "exec":
				// Check if this is the log command or format-patch.
				for _, a := range args {
					if a == "log" {
						return []byte("abc1234 First commit\ndef5678 Second commit\n"), nil
					}
					if a == "format-patch" {
						return []byte("From abc1234...\n---\npatch content\n"), nil
					}
				}
				return []byte(""), nil
			}
		}
		if name == "git" && len(args) > 0 && args[0] == "am" {
			gitAmCalled = true
			return []byte(""), nil
		}
		return []byte(""), nil
	}

	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gitAmCalled {
		t.Error("git am should NOT be called when user declines")
	}
	if !strings.Contains(stdout(opts), "Patch not applied") {
		t.Errorf("expected decline message, got: %s", stdout(opts))
	}
}

func TestExtract_DirectModeWarning(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeDirect

	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout(opts), "direct mode") {
		t.Errorf("expected direct mode message, got: %s", stdout(opts))
	}
}

func TestExtract_NoContainer(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "inspect" {
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	err := Extract(opts)
	if err == nil {
		t.Fatal("expected error when no container running")
	}
	if !strings.Contains(err.Error(), "no sandbox running") {
		t.Errorf("expected no sandbox message, got: %s", err.Error())
	}
}

func TestExtract_HappyPathWithYes(t *testing.T) {
	gitAmCalled := false
	gitAmFile := ""

	opts := testOpts()
	opts.Yes = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "inspect":
				if len(args) > 1 && args[1] == "--format" {
					return []byte("true"), nil
				}
				return []byte("{}"), nil
			case "exec":
				for _, a := range args {
					if a == "log" {
						return []byte("abc1234 First commit\n"), nil
					}
					if a == "format-patch" {
						return []byte("From abc1234...\n---\npatch content\n"), nil
					}
				}
				return []byte(""), nil
			}
		}
		if name == "git" && len(args) > 0 && args[0] == "am" {
			gitAmCalled = true
			if len(args) > 1 {
				gitAmFile = args[1]
			}
			return []byte("Applying: First commit\n"), nil
		}
		return []byte(""), nil
	}

	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gitAmCalled {
		t.Error("expected git am to be called")
	}
	if gitAmFile == "" {
		t.Error("expected git am to receive a patch file path")
	}
	out := stdout(opts)
	if !strings.Contains(out, "Patch applied successfully") {
		t.Errorf("expected success message, got: %s", out)
	}
	if !strings.Contains(out, "1 commits") {
		t.Errorf("expected commit count, got: %s", out)
	}
}

// --- Status() tests ---

func TestStatus_Running(t *testing.T) {
	inspectJSON := []podmanInspect{{
		ID:        "abc123def456789012345678",
		Name:      "uf-sandbox",
		ImageName: "quay.io/unbound-force/opencode-dev:latest",
		State: struct {
			Running   bool   `json:"Running"`
			StartedAt string `json:"StartedAt"`
			ExitCode  int    `json:"ExitCode"`
		}{
			Running:   true,
			StartedAt: "2026-04-12T10:00:00Z",
			ExitCode:  0,
		},
		Mounts: []struct {
			Source      string `json:"Source"`
			Destination string `json:"Destination"`
			RW          bool   `json:"RW"`
		}{
			{Source: "/home/dev/project", Destination: "/workspace", RW: false},
		},
	}}
	data, _ := json.Marshal(inspectJSON)

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return data, nil
		}
		return []byte(""), nil
	}

	status, err := Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Running {
		t.Error("expected Running=true")
	}
	if status.Name != "uf-sandbox" {
		t.Errorf("expected name uf-sandbox, got: %s", status.Name)
	}
	if status.ID != "abc123def456" {
		t.Errorf("expected short ID abc123def456, got: %s", status.ID)
	}
	if status.Image != "quay.io/unbound-force/opencode-dev:latest" {
		t.Errorf("expected default image, got: %s", status.Image)
	}
	if status.Mode != ModeIsolated {
		t.Errorf("expected isolated mode (RW=false), got: %s", status.Mode)
	}
	if status.ProjectDir != "/home/dev/project" {
		t.Errorf("expected project dir, got: %s", status.ProjectDir)
	}
	if status.ExitCode != -1 {
		t.Errorf("expected ExitCode=-1 for running, got: %d", status.ExitCode)
	}
}

func TestStatus_Stopped(t *testing.T) {
	inspectJSON := []podmanInspect{{
		ID:        "abc123def456789012345678",
		Name:      "uf-sandbox",
		ImageName: "quay.io/unbound-force/opencode-dev:latest",
		State: struct {
			Running   bool   `json:"Running"`
			StartedAt string `json:"StartedAt"`
			ExitCode  int    `json:"ExitCode"`
		}{
			Running:  false,
			ExitCode: 137,
		},
		Mounts: []struct {
			Source      string `json:"Source"`
			Destination string `json:"Destination"`
			RW          bool   `json:"RW"`
		}{
			{Source: "/home/dev/project", Destination: "/workspace", RW: true},
		},
	}}
	data, _ := json.Marshal(inspectJSON)

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return data, nil
		}
		return []byte(""), nil
	}

	status, err := Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Running {
		t.Error("expected Running=false")
	}
	if status.ExitCode != 137 {
		t.Errorf("expected ExitCode=137, got: %d", status.ExitCode)
	}
	if status.Mode != ModeDirect {
		t.Errorf("expected direct mode (RW=true), got: %s", status.Mode)
	}
}

func TestStatus_NoContainer(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return nil, fmt.Errorf("no such container")
		}
		return []byte(""), nil
	}

	status, err := Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Running {
		t.Error("expected Running=false when no container")
	}
}

// --- Health check tests ---

func TestWaitForHealth_ImmediateSuccess(t *testing.T) {
	opts := testOpts()
	opts.HTTPGet = func(url string) (int, error) {
		return 200, nil
	}

	err := waitForHealth(opts, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForHealth_DelayedSuccess(t *testing.T) {
	callCount := 0
	opts := testOpts()
	opts.HTTPGet = func(url string) (int, error) {
		callCount++
		if callCount < 3 {
			return 0, fmt.Errorf("connection refused")
		}
		return 200, nil
	}

	err := waitForHealth(opts, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount < 3 {
		t.Errorf("expected at least 3 calls, got: %d", callCount)
	}
}

func TestWaitForHealth_Timeout(t *testing.T) {
	opts := testOpts()
	opts.HTTPGet = func(url string) (int, error) {
		return 0, fmt.Errorf("connection refused")
	}

	err := waitForHealth(opts, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout message, got: %s", err.Error())
	}
}

// --- isContainerRunning tests ---

func TestIsContainerRunning_Running(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("true\n"), nil
	}

	running, err := isContainerRunning(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !running {
		t.Error("expected running=true")
	}
}

func TestIsContainerRunning_NotRunning(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("false\n"), nil
	}

	running, err := isContainerRunning(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if running {
		t.Error("expected running=false")
	}
}

func TestIsContainerRunning_NoContainer(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("no such container")
	}

	running, err := isContainerRunning(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if running {
		t.Error("expected running=false when no container")
	}
}

// --- FormatStatus tests ---

func TestFormatStatus_Running(t *testing.T) {
	var buf bytes.Buffer
	FormatStatus(&buf, ContainerStatus{
		Running:    true,
		Name:       "uf-sandbox",
		ID:         "abc123def456",
		Image:      DefaultImage,
		Mode:       ModeIsolated,
		ProjectDir: "/home/dev/project",
		ServerURL:  "http://localhost:4096",
		StartedAt:  "2026-04-12T10:00:00Z",
	})

	out := buf.String()
	if !strings.Contains(out, "Sandbox Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, "uf-sandbox") {
		t.Errorf("expected container name, got: %s", out)
	}
	if !strings.Contains(out, "isolated") {
		t.Errorf("expected mode, got: %s", out)
	}
}

func TestFormatStatus_NotRunning(t *testing.T) {
	var buf bytes.Buffer
	FormatStatus(&buf, ContainerStatus{Running: false})

	out := buf.String()
	if !strings.Contains(out, "No sandbox running") {
		t.Errorf("expected no sandbox message, got: %s", out)
	}
}

// --- isYes tests ---

func TestIsYes(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"y", true},
		{"Y", true},
		{"yes", true},
		{"YES", true},
		{"Yes", true},
		{"n", false},
		{"no", false},
		{"", false},
		{"maybe", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isYes(tt.input); got != tt.want {
				t.Errorf("isYes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ============================================================
// Spec 029: Backend Interface + Persistent Workspace Tests
// ============================================================

// --- projectName tests ---

func TestProjectName_Simple(t *testing.T) {
	got := projectName("/home/dev/my-project")
	if got != "my-project" {
		t.Errorf("expected my-project, got: %s", got)
	}
}

func TestProjectName_SpecialChars(t *testing.T) {
	got := projectName("/home/dev/My Project (v2)")
	want := "my-project--v2-"
	// Sanitize: lowercase, special chars → hyphens, trim trailing hyphens.
	if got != "my-project--v2" {
		t.Errorf("expected sanitized name without trailing hyphens, got: %s (want prefix of %s)", got, want)
	}
}

func TestProjectName_Empty(t *testing.T) {
	got := projectName("/")
	if got != "default" {
		t.Errorf("expected default, got: %s", got)
	}
}

// --- LoadConfig tests ---

func TestLoadConfig_HappyPath(t *testing.T) {
	yamlContent := `
backend: podman
ollama:
  host: http://ollama.internal:11434
demo_ports:
  - 3000
  - 8080
`
	opts := testOpts()
	opts.ReadFile = func(path string) ([]byte, error) {
		return []byte(yamlContent), nil
	}

	cfg, err := LoadConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Backend != "podman" {
		t.Errorf("expected backend podman, got: %s", cfg.Backend)
	}
	if cfg.Ollama.Host != "http://ollama.internal:11434" {
		t.Errorf("expected ollama host, got: %s", cfg.Ollama.Host)
	}
	if len(cfg.DemoPorts) != 2 || cfg.DemoPorts[0] != 3000 || cfg.DemoPorts[1] != 8080 {
		t.Errorf("expected demo ports [3000, 8080], got: %v", cfg.DemoPorts)
	}
}

func TestLoadConfig_Missing(t *testing.T) {
	opts := testOpts()
	// ReadFile returns error by default (file not found).

	cfg, err := LoadConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return zero-value defaults.
	if cfg.Backend != "" {
		t.Errorf("expected empty backend, got: %s", cfg.Backend)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	yamlContent := `
backend: podman
`
	opts := testOpts()
	opts.ReadFile = func(path string) ([]byte, error) {
		return []byte(yamlContent), nil
	}
	opts.Getenv = func(key string) string {
		switch key {
		case "UF_SANDBOX_BACKEND":
			return "podman"
		case "UF_OLLAMA_HOST":
			return "http://custom-ollama:11434"
		}
		return ""
	}

	cfg, err := LoadConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Env var should override config file.
	if cfg.Backend != "podman" {
		t.Errorf("expected env backend override, got: %s", cfg.Backend)
	}
	if cfg.Ollama.Host != "http://custom-ollama:11434" {
		t.Errorf("expected env ollama host override, got: %s", cfg.Ollama.Host)
	}
}

// --- ResolveBackend tests ---

func TestResolveBackend_AutoPodman(t *testing.T) {
	opts := testOpts()
	// Default testOpts: auto-detect returns Podman.

	backend, err := ResolveBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendPodman {
		t.Errorf("expected podman backend, got: %s", backend.Name())
	}
}

func TestResolveBackend_ExplicitPodman(t *testing.T) {
	opts := testOpts()
	opts.BackendName = BackendPodman

	backend, err := ResolveBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendPodman {
		t.Errorf("expected podman backend, got: %s", backend.Name())
	}
}

func TestResolveBackend_CheMigrationError(t *testing.T) {
	opts := testOpts()
	opts.BackendName = "che"

	_, err := ResolveBackend(opts)
	if err == nil {
		t.Fatal("expected error for che backend")
	}
	if !strings.Contains(err.Error(), "che backend removed") {
		t.Errorf("expected migration error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "devpod") {
		t.Errorf("expected devpod suggestion, got: %s", err.Error())
	}
}

func TestResolveBackend_UnknownBackend(t *testing.T) {
	opts := testOpts()
	opts.BackendName = "docker"

	_, err := ResolveBackend(opts)
	if err == nil {
		t.Fatal("expected error for unknown backend")
	}
	if !strings.Contains(err.Error(), "unknown backend") {
		t.Errorf("expected unknown backend message, got: %s", err.Error())
	}
}

// --- PodmanBackend tests ---

func TestPodmanCreate_HappyPath(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.Detach = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				return []byte("volume-created"), nil
			case "run":
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the sequence: volume create, run, cp.
	hasVolumeCreate := false
	hasRun := false
	hasCp := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "volume create") {
			hasVolumeCreate = true
		}
		if strings.Contains(cmd, "podman run") {
			hasRun = true
		}
		if strings.Contains(cmd, "podman cp") {
			hasCp = true
		}
	}
	if !hasVolumeCreate {
		t.Error("expected podman volume create")
	}
	if !hasRun {
		t.Error("expected podman run")
	}
	if !hasCp {
		t.Error("expected podman cp")
	}
}

func TestPodmanCreate_AlreadyExists(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "volume" && args[1] == "inspect" {
			return []byte("{}"), nil // volume exists
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when workspace already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected already exists message, got: %s", err.Error())
	}
}

func TestPodmanCreate_VolumeCreateFails(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			if args[1] == "inspect" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[1] == "create" {
				return []byte("permission denied"), fmt.Errorf("exit 1")
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when volume create fails")
	}
	if !strings.Contains(err.Error(), "failed to create volume") {
		t.Errorf("expected volume create error, got: %s", err.Error())
	}
}

func TestPodmanCreate_WithDemoPorts(t *testing.T) {
	var runArgs string
	opts := testOpts()
	opts.DemoPorts = []int{3000, 8080}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				return []byte(""), nil
			case "run":
				runArgs = strings.Join(args, " ")
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(runArgs, "-p 3000:3000") {
		t.Errorf("expected -p 3000:3000, got: %s", runArgs)
	}
	if !strings.Contains(runArgs, "-p 8080:8080") {
		t.Errorf("expected -p 8080:8080, got: %s", runArgs)
	}
}

func TestPodmanStart_PersistentResume(t *testing.T) {
	startCalled := false
	opts := testOpts()
	opts.Detach = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return []byte("{}"), nil // volume exists
			case "inspect":
				return []byte("{}"), nil // container exists
			case "start":
				startCalled = true
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !startCalled {
		t.Error("expected podman start to be called")
	}
}

func TestPodmanStart_EphemeralFallback(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "inspect" {
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Start(opts)
	if err == nil {
		t.Fatal("expected error when no persistent workspace")
	}
	if !strings.Contains(err.Error(), "no persistent workspace") {
		t.Errorf("expected no persistent workspace message, got: %s", err.Error())
	}
}

func TestPodmanStop_PersistentPreservesVolume(t *testing.T) {
	stopCalled := false
	rmCalled := false
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return []byte("{}"), nil // volume exists
			case "inspect":
				return []byte("{}"), nil // container exists
			case "stop":
				stopCalled = true
				return []byte(""), nil
			case "rm":
				rmCalled = true
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Error("expected podman stop to be called")
	}
	if rmCalled {
		t.Error("expected podman rm NOT to be called (persistent mode)")
	}
	if !strings.Contains(stdout(opts), "state preserved") {
		t.Errorf("expected state preserved message, got: %s", stdout(opts))
	}
}

func TestPodmanStop_EphemeralRemoves(t *testing.T) {
	// This tests the top-level Stop() in ephemeral mode.
	stopCalled := false
	rmCalled := false
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "inspect":
				return []byte("{}"), nil // container exists
			case "stop":
				stopCalled = true
				return []byte(""), nil
			case "rm":
				rmCalled = true
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	err := Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Error("expected podman stop to be called")
	}
	if !rmCalled {
		t.Error("expected podman rm to be called (ephemeral mode)")
	}
}

func TestPodmanDestroy_HappyPath(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Destroy(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasRm := false
	hasVolumeRm := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "podman rm") {
			hasRm = true
		}
		if strings.Contains(cmd, "volume rm") {
			hasVolumeRm = true
		}
	}
	if !hasRm {
		t.Error("expected podman rm")
	}
	if !hasVolumeRm {
		t.Error("expected podman volume rm")
	}
	if !strings.Contains(stdout(opts), "Sandbox destroyed") {
		t.Errorf("expected destroyed message, got: %s", stdout(opts))
	}
}

func TestPodmanDestroy_NoWorkspace(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		// All commands return error (nothing exists).
		return nil, fmt.Errorf("not found")
	}

	b := &PodmanBackend{}
	err := b.Destroy(opts)
	// Destroy is idempotent — no error even when nothing exists.
	if err != nil {
		t.Fatalf("expected no error (idempotent), got: %v", err)
	}
}

func TestPodmanDestroy_RunningWorkspace(t *testing.T) {
	stopCalled := false
	rmCalled := false
	volumeRmCalled := false
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "stop":
				stopCalled = true
				return []byte(""), nil
			case "rm":
				rmCalled = true
				return []byte(""), nil
			case "volume":
				if args[1] == "rm" {
					volumeRmCalled = true
				}
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Destroy(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Error("expected podman stop before destroy")
	}
	if !rmCalled {
		t.Error("expected podman rm")
	}
	if !volumeRmCalled {
		t.Error("expected podman volume rm")
	}
}

func TestPodmanStatus_PersistentRunning(t *testing.T) {
	inspectJSON := []podmanInspect{{
		ID:        "abc123def456789012345678",
		Name:      "uf-sandbox-test-project",
		ImageName: DefaultImage,
		State: struct {
			Running   bool   `json:"Running"`
			StartedAt string `json:"StartedAt"`
			ExitCode  int    `json:"ExitCode"`
		}{
			Running:   true,
			StartedAt: "2026-04-13T10:00:00Z",
		},
		Mounts: nil,
	}}
	data, _ := json.Marshal(inspectJSON)

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" && args[1] == "inspect" {
				return []byte("{}"), nil // volume exists
			}
			if args[0] == "inspect" {
				return data, nil
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	ws, err := b.Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Exists {
		t.Error("expected Exists=true")
	}
	if !ws.Running {
		t.Error("expected Running=true")
	}
	if !ws.Persistent {
		t.Error("expected Persistent=true")
	}
	if ws.Backend != BackendPodman {
		t.Errorf("expected podman backend, got: %s", ws.Backend)
	}
	if ws.ID != "abc123def456" {
		t.Errorf("expected short ID, got: %s", ws.ID)
	}
}

func TestPodmanStatus_PersistentStopped(t *testing.T) {
	inspectJSON := []podmanInspect{{
		ID:        "abc123def456789012345678",
		Name:      "uf-sandbox-test-project",
		ImageName: DefaultImage,
		State: struct {
			Running   bool   `json:"Running"`
			StartedAt string `json:"StartedAt"`
			ExitCode  int    `json:"ExitCode"`
		}{
			Running:  false,
			ExitCode: 0,
		},
		Mounts: nil,
	}}
	data, _ := json.Marshal(inspectJSON)

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" && args[1] == "inspect" {
				return []byte("{}"), nil
			}
			if args[0] == "inspect" {
				return data, nil
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	ws, err := b.Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Exists {
		t.Error("expected Exists=true")
	}
	if ws.Running {
		t.Error("expected Running=false")
	}
	if ws.ExitCode != 0 {
		t.Errorf("expected ExitCode=0, got: %d", ws.ExitCode)
	}
}

// --- FormatWorkspaceStatus tests ---

func TestFormatWorkspaceStatus_Podman(t *testing.T) {
	var buf bytes.Buffer
	FormatWorkspaceStatus(&buf, WorkspaceStatus{
		Exists:     true,
		Running:    true,
		Backend:    BackendPodman,
		Name:       "uf-sandbox-myproject",
		Image:      DefaultImage,
		Mode:       ModeIsolated,
		ProjectDir: "/home/dev/myproject",
		ServerURL:  "http://localhost:4096",
		StartedAt:  "2026-04-13T10:00:00Z",
		Persistent: true,
	})

	out := buf.String()
	if !strings.Contains(out, "Sandbox Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, "uf-sandbox-myproject") {
		t.Errorf("expected workspace name, got: %s", out)
	}
	if !strings.Contains(out, "persistent") {
		t.Errorf("expected persistent label, got: %s", out)
	}
	if !strings.Contains(out, "running") {
		t.Errorf("expected running state, got: %s", out)
	}
}

func TestFormatWorkspaceStatus_WithDemoEndpoints(t *testing.T) {
	var buf bytes.Buffer
	FormatWorkspaceStatus(&buf, WorkspaceStatus{
		Exists:  true,
		Running: true,
		Name:    "uf-sandbox-myproject",
		Mode:    ModeIsolated,
		DemoEndpoints: []DemoEndpoint{
			{Name: "demo-web", Port: 3000, URL: "http://localhost:3000", Protocol: "http"},
			{Name: "demo-api", Port: 8080, URL: "http://localhost:8080", Protocol: "http"},
		},
		Persistent: true,
	})

	out := buf.String()
	if !strings.Contains(out, "demo-web") {
		t.Errorf("expected demo-web endpoint, got: %s", out)
	}
	if !strings.Contains(out, "demo-api") {
		t.Errorf("expected demo-api endpoint, got: %s", out)
	}
	if !strings.Contains(out, "http://localhost:3000") {
		t.Errorf("expected port 3000 URL, got: %s", out)
	}
}

func TestFormatWorkspaceStatus_NoWorkspace(t *testing.T) {
	var buf bytes.Buffer
	FormatWorkspaceStatus(&buf, WorkspaceStatus{Exists: false})

	out := buf.String()
	if !strings.Contains(out, "No sandbox workspace found") {
		t.Errorf("expected no workspace message, got: %s", out)
	}
}

// --- Backward compatibility tests ---

func TestStart_EphemeralMode(t *testing.T) {
	// Verify that `uf sandbox start` without prior `create`
	// uses ephemeral mode (Spec 028 behavior).
	runCalled := false
	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return []byte(""), nil
			case "run":
				runCalled = true
				// Verify ephemeral container name.
				for _, a := range args {
					if a == ContainerName {
						return []byte("container-id"), nil
					}
				}
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runCalled {
		t.Error("expected podman run for ephemeral mode")
	}
}

func TestStop_EphemeralMode(t *testing.T) {
	// Verify that `uf sandbox stop` in ephemeral mode
	// removes the container (Spec 028 behavior).
	rmCalled := false
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "inspect":
				return []byte("{}"), nil
			case "stop":
				return []byte(""), nil
			case "rm":
				rmCalled = true
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	err := Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rmCalled {
		t.Error("expected podman rm in ephemeral mode")
	}
}

func TestAttach_Unchanged(t *testing.T) {
	// Verify attach works with both persistent and ephemeral.
	attachCalled := false
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return []byte("true"), nil // container running
		}
		return []byte(""), nil
	}
	opts.ExecInteractive = func(name string, args ...string) error {
		attachCalled = true
		return nil
	}

	err := Attach(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !attachCalled {
		t.Error("expected attach to be called")
	}
}

func TestExtract_Unchanged(t *testing.T) {
	// Verify extract works in ephemeral mode.
	opts := testOpts()
	opts.Mode = ModeDirect
	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout(opts), "direct mode") {
		t.Errorf("expected direct mode message, got: %s", stdout(opts))
	}
}

func TestStatus_EphemeralFallback(t *testing.T) {
	// Verify status shows Spec 028 format for ephemeral.
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "inspect" {
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	status, err := Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Running {
		t.Error("expected Running=false for no container")
	}
}

// --- Git sync tests ---

func TestSetupGitSync_PodmanBackend(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		if name == "git" {
			for _, a := range args {
				if a == "rev-parse" {
					return []byte("main\n"), nil
				}
				if a == "get-url" {
					return []byte("https://github.com/org/repo.git\n"), nil
				}
			}
		}
		return []byte(""), nil
	}

	err := setupGitSync(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasSetURL := false
	hasCheckout := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "set-url") {
			hasSetURL = true
		}
		if strings.Contains(cmd, "checkout") {
			hasCheckout = true
		}
	}
	if !hasSetURL {
		t.Error("expected git remote set-url")
	}
	if !hasCheckout {
		t.Error("expected git checkout")
	}
}

func TestCheckGitSync_Clean(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "exec" {
			for _, a := range args {
				if a == "status" {
					return []byte(""), nil // clean
				}
				if a == "pull" {
					return []byte("Already up to date.\n"), nil
				}
			}
		}
		return []byte(""), nil
	}

	err := checkGitSync(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckGitSync_Diverged(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "exec" {
			for _, a := range args {
				if a == "status" {
					return []byte(""), nil // clean
				}
				if a == "pull" {
					return nil, fmt.Errorf("fatal: Not possible to fast-forward")
				}
			}
		}
		return []byte(""), nil
	}

	err := checkGitSync(opts)
	if err == nil {
		t.Fatal("expected error when diverged")
	}
	if !strings.Contains(err.Error(), "diverged") {
		t.Errorf("expected diverged message, got: %s", err.Error())
	}
}

func TestExtract_PersistentWorkspaceEarlyReturn(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			return []byte("{}"), nil // volume exists → persistent
		}
		return []byte(""), nil
	}

	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout(opts)
	if !strings.Contains(out, "persistent workspace") {
		t.Errorf("expected persistent workspace message, got: %s", out)
	}
	if !strings.Contains(out, "git push") {
		t.Errorf("expected git push suggestion, got: %s", out)
	}
}

// --- Create/Destroy dispatch tests ---

func TestCreate_DispatchPodman(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.BackendName = BackendPodman
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if len(args) > 1 && args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				return []byte(""), nil
			case "run":
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	err := Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout(opts)
	if !strings.Contains(out, "Sandbox created (detached)") {
		t.Errorf("expected detached message, got: %s", out)
	}
}

func TestDestroy_DispatchPodman(t *testing.T) {
	opts := testOpts()
	opts.BackendName = BackendPodman
	err := Destroy(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout(opts), "Sandbox destroyed") {
		t.Errorf("expected destroyed message, got: %s", stdout(opts))
	}
}

func TestWorkspaceStatusCheck_NoPersistent(t *testing.T) {
	opts := testOpts()
	// Default testOpts has no persistent workspace.
	ws, err := WorkspaceStatusCheck(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ws.Exists {
		t.Error("expected Exists=false when no persistent workspace")
	}
}

func TestWorkspaceStatusCheck_Persistent(t *testing.T) {
	inspectJSON := []podmanInspect{{
		ID:        "abc123def456789012345678",
		Name:      "uf-sandbox-test-project",
		ImageName: DefaultImage,
		State: struct {
			Running   bool   `json:"Running"`
			StartedAt string `json:"StartedAt"`
			ExitCode  int    `json:"ExitCode"`
		}{Running: true},
		Mounts: nil,
	}}
	data, _ := json.Marshal(inspectJSON)

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return []byte("{}"), nil // volume exists
			}
			if args[0] == "inspect" {
				return data, nil
			}
		}
		return []byte(""), nil
	}

	ws, err := WorkspaceStatusCheck(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Exists {
		t.Error("expected Exists=true")
	}
	if !ws.Running {
		t.Error("expected Running=true")
	}
}

// --- mergeDemoPorts tests ---

func TestMergeDemoPorts_Dedup(t *testing.T) {
	result := mergeDemoPorts([]int{3000, 8080}, []int{8080, 9090})
	if len(result) != 3 {
		t.Errorf("expected 3 ports, got: %d (%v)", len(result), result)
	}
}

// ============================================================
// Spec 033: Gateway Integration Tests (T070-T081)
// ============================================================

// --- gatewayHealthCheck tests (T070) ---

func TestGatewayHealthCheck_Success(t *testing.T) {
	httpGet := func(url string) (int, error) {
		return 200, nil
	}

	if !gatewayHealthCheck(httpGet, 53147) {
		t.Error("expected true when health returns 200")
	}
}

func TestGatewayHealthCheck_Failure(t *testing.T) {
	httpGet := func(url string) (int, error) {
		return 0, fmt.Errorf("connection refused")
	}

	if gatewayHealthCheck(httpGet, 53147) {
		t.Error("expected false when health check fails")
	}
}

func TestGatewayHealthCheck_NonOKStatus(t *testing.T) {
	httpGet := func(url string) (int, error) {
		return 500, nil
	}

	if gatewayHealthCheck(httpGet, 53147) {
		t.Error("expected false when health returns non-200")
	}
}

func TestGatewayHealthCheck_URL(t *testing.T) {
	var capturedURL string
	httpGet := func(url string) (int, error) {
		capturedURL = url
		return 200, nil
	}

	gatewayHealthCheck(httpGet, 9000)
	if capturedURL != "http://localhost:9000/health" {
		t.Errorf("expected http://localhost:9000/health, got: %s", capturedURL)
	}
}

// --- autoStartGateway tests (T071-T073) ---

func TestAutoStartGateway_ProviderDetected(t *testing.T) {
	execCmdCalled := false
	opts := testOpts()
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}
	// Health check fails first (no gateway running), then
	// succeeds after ExecCmd starts the gateway.
	healthCallCount := 0
	opts.HTTPGet = func(url string) (int, error) {
		healthCallCount++
		if healthCallCount == 1 {
			// First call: gateway not running yet.
			return 0, fmt.Errorf("connection refused")
		}
		// Subsequent calls: gateway is running.
		return 200, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "uf" && len(args) > 0 && args[0] == "gateway" {
			execCmdCalled = true
			return []byte(""), nil
		}
		// Volume inspect fails (no persistent workspace).
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			return nil, fmt.Errorf("no such volume")
		}
		return []byte(""), nil
	}

	port, active, err := autoStartGateway(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected gateway to be active")
	}
	if port != GatewayDefaultPort {
		t.Errorf("expected port %d, got: %d", GatewayDefaultPort, port)
	}
	if !execCmdCalled {
		t.Error("expected uf gateway --detach to be called")
	}
}

func TestAutoStartGateway_NoProvider(t *testing.T) {
	opts := testOpts()
	// No provider env vars set (default testOpts).

	port, active, err := autoStartGateway(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if active {
		t.Error("expected gateway to be inactive when no provider")
	}
	if port != 0 {
		t.Errorf("expected port 0, got: %d", port)
	}
}

func TestAutoStartGateway_ExistingGateway(t *testing.T) {
	execCmdCalled := false
	opts := testOpts()
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}
	// Health check succeeds immediately (gateway already running).
	opts.HTTPGet = func(url string) (int, error) {
		return 200, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "uf" {
			execCmdCalled = true
		}
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			return nil, fmt.Errorf("no such volume")
		}
		return []byte(""), nil
	}

	port, active, err := autoStartGateway(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected gateway to be active (reused)")
	}
	if port != GatewayDefaultPort {
		t.Errorf("expected port %d, got: %d", GatewayDefaultPort, port)
	}
	if execCmdCalled {
		t.Error("expected ExecCmd NOT called (reuse existing gateway)")
	}
}

func TestAutoStartGateway_VertexDetected(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		switch key {
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		}
		return ""
	}
	opts.HTTPGet = func(url string) (int, error) {
		return 200, nil // Gateway already running.
	}

	_, active, err := autoStartGateway(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected gateway to be active for Vertex")
	}
}

func TestAutoStartGateway_BedrockDetected(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		if key == "CLAUDE_CODE_USE_BEDROCK" {
			return "1"
		}
		return ""
	}
	opts.HTTPGet = func(url string) (int, error) {
		return 200, nil
	}

	_, active, err := autoStartGateway(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !active {
		t.Error("expected gateway to be active for Bedrock")
	}
}

// --- gatewayEnvVars tests (T074) ---

func TestGatewayEnvVars(t *testing.T) {
	args := gatewayEnvVars(53147)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "ANTHROPIC_BASE_URL=http://host.containers.internal:53147") {
		t.Errorf("expected ANTHROPIC_BASE_URL, got: %s", joined)
	}
	if !strings.Contains(joined, "ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected ANTHROPIC_API_KEY=gateway, got: %s", joined)
	}
}

func TestGatewayEnvVars_CustomPort(t *testing.T) {
	args := gatewayEnvVars(9000)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "ANTHROPIC_BASE_URL=http://host.containers.internal:9000") {
		t.Errorf("expected port 9000 in URL, got: %s", joined)
	}
}

// --- forwardedEnvVars with gateway tests (T075-T076) ---

func TestForwardedEnvVars_GatewayActive(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		switch key {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		case "OPENAI_API_KEY":
			return "sk-xxx"
		case "GEMINI_API_KEY":
			return "gemini-xxx"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		case "GOOGLE_CLOUD_PROJECT":
			return "gcp-project"
		case "VERTEX_LOCATION":
			return "us-central1"
		}
		return ""
	}

	args := forwardedEnvVars(opts, true)
	joined := strings.Join(args, " ")

	// Skipped keys when gateway is active.
	if strings.Contains(joined, "-e ANTHROPIC_API_KEY") {
		t.Errorf("ANTHROPIC_API_KEY should be skipped with gateway, got: %s", joined)
	}
	if strings.Contains(joined, "-e ANTHROPIC_VERTEX_PROJECT_ID") {
		t.Errorf("ANTHROPIC_VERTEX_PROJECT_ID should be skipped with gateway, got: %s", joined)
	}
	if strings.Contains(joined, "-e CLAUDE_CODE_USE_VERTEX") {
		t.Errorf("CLAUDE_CODE_USE_VERTEX should be skipped with gateway, got: %s", joined)
	}
	if strings.Contains(joined, "-e GOOGLE_CLOUD_PROJECT") {
		t.Errorf("GOOGLE_CLOUD_PROJECT should be skipped with gateway, got: %s", joined)
	}
	if strings.Contains(joined, "-e VERTEX_LOCATION") {
		t.Errorf("VERTEX_LOCATION should be skipped with gateway, got: %s", joined)
	}

	// Non-proxied keys should still be forwarded.
	if !strings.Contains(joined, "-e OPENAI_API_KEY") {
		t.Errorf("OPENAI_API_KEY should be forwarded, got: %s", joined)
	}
	if !strings.Contains(joined, "-e GEMINI_API_KEY") {
		t.Errorf("GEMINI_API_KEY should be forwarded, got: %s", joined)
	}
	// OLLAMA_HOST always present.
	if !strings.Contains(joined, "OLLAMA_HOST=host.containers.internal:11434") {
		t.Errorf("OLLAMA_HOST should always be present, got: %s", joined)
	}
}

func TestForwardedEnvVars_GatewayInactive(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		switch key {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		case "OPENAI_API_KEY":
			return "sk-xxx"
		case "ANTHROPIC_VERTEX_PROJECT_ID":
			return "my-project"
		case "CLAUDE_CODE_USE_VERTEX":
			return "1"
		}
		return ""
	}

	args := forwardedEnvVars(opts, false)
	joined := strings.Join(args, " ")

	// All keys should be forwarded when gateway is inactive.
	if !strings.Contains(joined, "-e ANTHROPIC_API_KEY") {
		t.Errorf("ANTHROPIC_API_KEY should be forwarded, got: %s", joined)
	}
	if !strings.Contains(joined, "-e OPENAI_API_KEY") {
		t.Errorf("OPENAI_API_KEY should be forwarded, got: %s", joined)
	}
	if !strings.Contains(joined, "-e ANTHROPIC_VERTEX_PROJECT_ID") {
		t.Errorf("ANTHROPIC_VERTEX_PROJECT_ID should be forwarded, got: %s", joined)
	}
	if !strings.Contains(joined, "-e CLAUDE_CODE_USE_VERTEX") {
		t.Errorf("CLAUDE_CODE_USE_VERTEX should be forwarded, got: %s", joined)
	}
}

// --- buildRunArgs with gateway tests (T079-T080) ---

func TestBuildRunArgs_GatewayActive(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs
	opts.Getenv = func(key string) string {
		switch key {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		case "OPENAI_API_KEY":
			return "sk-xxx"
		}
		return ""
	}

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, true, 53147)
	joined := strings.Join(args, " ")

	// Verify gateway env vars present.
	if !strings.Contains(joined, "ANTHROPIC_BASE_URL=http://host.containers.internal:53147") {
		t.Errorf("expected ANTHROPIC_BASE_URL, got: %s", joined)
	}
	if !strings.Contains(joined, "ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected ANTHROPIC_API_KEY, got: %s", joined)
	}

	// Verify host's real ANTHROPIC_API_KEY is not forwarded.
	// The gateway placeholder (ANTHROPIC_API_KEY=gateway) IS
	// present, but the bare "-e ANTHROPIC_API_KEY" (which
	// reads from host env) should NOT be. Count occurrences:
	// exactly 1 (the gateway placeholder).
	count := strings.Count(joined, "ANTHROPIC_API_KEY")
	if count != 1 {
		t.Errorf("expected exactly 1 ANTHROPIC_API_KEY (gateway), got %d in: %s", count, joined)
	}

	// Verify OPENAI_API_KEY IS forwarded (not proxied by gateway).
	if !strings.Contains(joined, "-e OPENAI_API_KEY") {
		t.Errorf("OPENAI_API_KEY should be forwarded, got: %s", joined)
	}
}

func TestBuildRunArgs_GatewayInactive(t *testing.T) {
	opts := testOpts()
	opts.Mode = ModeIsolated
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-xxx"
		}
		return ""
	}

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	// Verify no gateway env vars.
	if strings.Contains(joined, "ANTHROPIC_BASE_URL") {
		t.Errorf("expected no ANTHROPIC_BASE_URL without gateway, got: %s", joined)
	}
	if strings.Contains(joined, "ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected no ANTHROPIC_API_KEY=gateway without gateway, got: %s", joined)
	}

	// Verify ANTHROPIC_API_KEY IS forwarded from host.
	if !strings.Contains(joined, "-e ANTHROPIC_API_KEY") {
		t.Errorf("ANTHROPIC_API_KEY should be forwarded without gateway, got: %s", joined)
	}
}

// --- Start() auto-starts gateway test (T081) ---

func TestStart_AutoStartsGateway(t *testing.T) {
	gatewayStarted := false
	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	healthCallCount := 0
	opts.HTTPGet = func(url string) (int, error) {
		healthCallCount++
		// Gateway health check: first call fails, then succeeds.
		if strings.Contains(url, "53147") {
			if healthCallCount == 1 {
				return 0, fmt.Errorf("connection refused")
			}
			return 200, nil
		}
		// OpenCode server health check: always succeeds.
		return 200, nil
	}

	var runArgs string
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "uf" && len(args) > 0 && args[0] == "gateway" {
			gatewayStarted = true
			return []byte(""), nil
		}
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return []byte(""), nil
			case "run":
				runArgs = strings.Join(args, " ")
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !gatewayStarted {
		t.Error("expected gateway to be auto-started")
	}

	// Verify container received gateway env vars.
	if !strings.Contains(runArgs, "ANTHROPIC_BASE_URL") {
		t.Errorf("expected ANTHROPIC_BASE_URL in container args, got: %s", runArgs)
	}
	if !strings.Contains(runArgs, "ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected ANTHROPIC_API_KEY=gateway in container args, got: %s", runArgs)
	}

	// Verify host's real ANTHROPIC_API_KEY is not forwarded.
	// Only the gateway placeholder (ANTHROPIC_API_KEY=gateway)
	// should be present, not the bare forwarded form.
	count := strings.Count(runArgs, "ANTHROPIC_API_KEY")
	if count != 1 {
		t.Errorf("expected exactly 1 ANTHROPIC_API_KEY (gateway), got %d in: %s", count, runArgs)
	}

	// Verify stderr contains gateway active message.
	stderrOut := opts.Stderr.(*bytes.Buffer).String()
	if !strings.Contains(stderrOut, "Gateway active") {
		t.Errorf("expected gateway active message in stderr, got: %s", stderrOut)
	}
}

func TestStart_NoGatewayFallback(t *testing.T) {
	// When no provider env vars are set, Start() should
	// fall back to credential mount behavior (backward
	// compatible, identical to pre-gateway behavior).
	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	// No provider env vars set (default testOpts).

	var runArgs string
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return []byte(""), nil
			case "run":
				runArgs = strings.Join(args, " ")
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no gateway env vars in container args.
	if strings.Contains(runArgs, "ANTHROPIC_BASE_URL") {
		t.Errorf("expected no ANTHROPIC_BASE_URL without gateway, got: %s", runArgs)
	}
	if strings.Contains(runArgs, "ANTHROPIC_API_KEY") {
		t.Errorf("expected no ANTHROPIC_API_KEY without gateway, got: %s", runArgs)
	}
}

// ============================================================
// UID Mapping Tests (sandbox-uid-mapping change, Task Group 8)
// ============================================================

// --- uidMappingArgs tests (8.1, 8.2) ---

func TestUIDMappingArgs_Default(t *testing.T) {
	result := uidMappingArgs(Options{})
	if len(result) != 1 {
		t.Fatalf("expected 1 element, got %d: %v", len(result), result)
	}
	if result[0] != "--userns=keep-id:uid=1000,gid=1000" {
		t.Errorf("expected --userns=keep-id:uid=1000,gid=1000, got: %s", result[0])
	}
}

func TestUIDMappingArgs_UIDMap(t *testing.T) {
	result := uidMappingArgs(Options{UIDMap: true})
	if len(result) != 12 {
		t.Fatalf("expected 12 elements, got %d: %v", len(result), result)
	}
	joined := strings.Join(result, " ")
	if !strings.Contains(joined, "--uidmap") {
		t.Errorf("expected --uidmap in result, got: %s", joined)
	}
	if !strings.Contains(joined, "--gidmap") {
		t.Errorf("expected --gidmap in result, got: %s", joined)
	}
	if strings.Contains(joined, "--userns") {
		t.Errorf("expected no --userns when UIDMap=true, got: %s", joined)
	}
}

// --- buildRunArgs integration tests (8.3, 8.4) ---

func TestBuildRunArgs_IncludesUserNS(t *testing.T) {
	opts := testOpts()
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "linux", Arch: "amd64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "--userns=keep-id:uid=1000,gid=1000") {
		t.Errorf("expected --userns=keep-id:uid=1000,gid=1000, got: %s", joined)
	}
	// Verify it appears before the image argument.
	usrIdx := -1
	imgIdx := -1
	for i, a := range args {
		if a == "--userns=keep-id:uid=1000,gid=1000" {
			usrIdx = i
		}
		if a == DefaultImage {
			imgIdx = i
		}
	}
	if usrIdx < 0 {
		t.Fatal("--userns not found in args")
	}
	if imgIdx < 0 {
		t.Fatal("image not found in args")
	}
	if usrIdx >= imgIdx {
		t.Errorf("--userns (idx %d) should appear before image (idx %d)", usrIdx, imgIdx)
	}
}

func TestBuildRunArgs_UIDMapOverride(t *testing.T) {
	opts := testOpts()
	opts.UIDMap = true
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "linux", Arch: "amd64"}
	args := buildRunArgs(opts, platform, false, 0)
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "--uidmap") {
		t.Errorf("expected --uidmap in args, got: %s", joined)
	}
	if !strings.Contains(joined, "--gidmap") {
		t.Errorf("expected --gidmap in args, got: %s", joined)
	}
	if strings.Contains(joined, "--userns") {
		t.Errorf("expected no --userns when UIDMap=true, got: %s", joined)
	}
}

// --- buildPersistentRunArgs integration tests (8.5, 8.6) ---

func TestBuildPersistentRunArgs_IncludesUserNS(t *testing.T) {
	opts := testOpts()
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "linux", Arch: "amd64"}
	args := buildPersistentRunArgs(opts, platform, "uf-sandbox-test", "uf-vol-test")
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "--userns=keep-id:uid=1000,gid=1000") {
		t.Errorf("expected --userns=keep-id:uid=1000,gid=1000, got: %s", joined)
	}
	// Verify it appears before the image argument.
	usrIdx := -1
	imgIdx := -1
	for i, a := range args {
		if a == "--userns=keep-id:uid=1000,gid=1000" {
			usrIdx = i
		}
		if a == DefaultImage {
			imgIdx = i
		}
	}
	if usrIdx < 0 {
		t.Fatal("--userns not found in args")
	}
	if imgIdx < 0 {
		t.Fatal("image not found in args")
	}
	if usrIdx >= imgIdx {
		t.Errorf("--userns (idx %d) should appear before image (idx %d)", usrIdx, imgIdx)
	}
}

// --- buildPersistentRunArgs gateway tests (Task Group 3) ---

func TestBuildPersistentRunArgs_GatewayActive(t *testing.T) {
	opts := testOpts()
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs
	opts.GatewayActive = true
	opts.GatewayPort = GatewayDefaultPort
	opts.Getenv = func(key string) string {
		switch key {
		case "ANTHROPIC_API_KEY":
			return "sk-ant-xxx"
		case "OPENAI_API_KEY":
			return "sk-xxx"
		}
		return ""
	}

	platform := PlatformConfig{OS: "linux", Arch: "amd64"}
	args := buildPersistentRunArgs(opts, platform, "uf-sandbox-test", "uf-vol-test")
	joined := strings.Join(args, " ")

	// Verify gateway env vars present.
	if !strings.Contains(joined, "ANTHROPIC_BASE_URL=http://host.containers.internal:53147") {
		t.Errorf("expected ANTHROPIC_BASE_URL, got: %s", joined)
	}
	if !strings.Contains(joined, "ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected ANTHROPIC_API_KEY=gateway, got: %s", joined)
	}

	// Verify host's real ANTHROPIC_API_KEY is NOT forwarded
	// (skipped by gatewaySkippedKeys).
	count := strings.Count(joined, "ANTHROPIC_API_KEY")
	if count != 1 {
		t.Errorf("expected exactly 1 ANTHROPIC_API_KEY (gateway), got %d in: %s", count, joined)
	}

	// Verify non-proxied keys are still forwarded.
	if !strings.Contains(joined, "-e OPENAI_API_KEY") {
		t.Errorf("OPENAI_API_KEY should be forwarded, got: %s", joined)
	}
}

func TestBuildPersistentRunArgs_GatewayInactive(t *testing.T) {
	opts := testOpts()
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs
	opts.GatewayActive = false
	opts.GatewayPort = 0
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-xxx"
		}
		return ""
	}

	platform := PlatformConfig{OS: "linux", Arch: "amd64"}
	args := buildPersistentRunArgs(opts, platform, "uf-sandbox-test", "uf-vol-test")
	joined := strings.Join(args, " ")

	// Verify no gateway env vars.
	if strings.Contains(joined, "ANTHROPIC_BASE_URL") {
		t.Errorf("expected no ANTHROPIC_BASE_URL without gateway, got: %s", joined)
	}
	if strings.Contains(joined, "ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected no ANTHROPIC_API_KEY=gateway without gateway, got: %s", joined)
	}

	// Verify ANTHROPIC_API_KEY IS forwarded from host.
	if !strings.Contains(joined, "-e ANTHROPIC_API_KEY") {
		t.Errorf("ANTHROPIC_API_KEY should be forwarded without gateway, got: %s", joined)
	}
}

func TestBuildPersistentRunArgs_UIDMapOverride(t *testing.T) {
	opts := testOpts()
	opts.UIDMap = true
	opts.Image = DefaultImage
	opts.Memory = DefaultMemory
	opts.CPUs = DefaultCPUs

	platform := PlatformConfig{OS: "linux", Arch: "amd64"}
	args := buildPersistentRunArgs(opts, platform, "uf-sandbox-test", "uf-vol-test")
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "--uidmap") {
		t.Errorf("expected --uidmap in args, got: %s", joined)
	}
	if !strings.Contains(joined, "--gidmap") {
		t.Errorf("expected --gidmap in args, got: %s", joined)
	}
	if strings.Contains(joined, "--userns") {
		t.Errorf("expected no --userns when UIDMap=true, got: %s", joined)
	}
}

// --- probeUIDMapping tests (8.7-8.11) ---

func TestProbeUIDMapping_Success(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("1000\n"), nil
	}

	if !probeUIDMapping(opts) {
		t.Error("expected probeUIDMapping to return true when output is 1000")
	}
}

func TestProbeUIDMapping_Failure(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("0\n"), nil
	}

	if probeUIDMapping(opts) {
		t.Error("expected probeUIDMapping to return false when output is 0")
	}
}

func TestProbeUIDMapping_Error(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("podman not available")
	}

	if probeUIDMapping(opts) {
		t.Error("expected probeUIDMapping to return false on error (fail-safe)")
	}
}

func TestProbeUIDMapping_UnexpectedOutput(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("nobody\n"), nil
	}

	if probeUIDMapping(opts) {
		t.Error("expected probeUIDMapping to return false for non-numeric output")
	}
}

func TestProbeUIDMapping_EmptyOutput(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte(""), nil
	}

	if probeUIDMapping(opts) {
		t.Error("expected probeUIDMapping to return false for empty output")
	}
}

// --- DetectPlatform test (8.12) ---

func TestDetectPlatform_LinuxAlwaysSupported(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("test only runs on Linux")
	}
	opts := testOpts()
	p := DetectPlatform(opts)
	if !p.UIDMapSupported {
		t.Error("expected UIDMapSupported=true on Linux")
	}
}

// --- Start() integration with Platform injection (8.13, 8.14) ---

func TestStart_DarwinUIDMapNotSupported(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{
		OS:              "darwin",
		Arch:            "arm64",
		UIDMapSupported: false,
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "--version" {
				return []byte("podman version 5.0.0\n"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when macOS UID mapping not supported")
	}
	if !strings.Contains(err.Error(), "podman machine UID mapping") {
		t.Errorf("expected 'podman machine UID mapping' in error, got: %s", err.Error())
	}
}

func TestStart_DarwinUIDMapOverride(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.UIDMap = true
	opts.Platform = &PlatformConfig{
		OS:              "darwin",
		Arch:            "arm64",
		UIDMapSupported: false,
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "info":
				return []byte("true\n"), nil // rootless
			case "inspect":
				return nil, fmt.Errorf("no such container")
			case "image":
				return []byte(""), nil
			case "run":
				return []byte("container-id"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	// The UID mapping error should NOT be returned because
	// --uidmap overrides the detection. It may fail later
	// for other reasons — just verify the UID mapping error
	// is NOT the one returned.
	if err != nil && strings.Contains(err.Error(), "podman machine UID mapping") {
		t.Errorf("expected --uidmap to bypass probe error, got: %s", err.Error())
	}
}

// --- parsePodmanVersion tests (8.15, 8.16) ---

func TestParsePodmanVersion_Valid(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("podman version 4.9.3\n"), nil
	}

	major, minor, err := parsePodmanVersion(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if major != 4 {
		t.Errorf("expected major=4, got: %d", major)
	}
	if minor != 9 {
		t.Errorf("expected minor=9, got: %d", minor)
	}
}

func TestParsePodmanVersion_TooOld(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			if args[0] == "volume" {
				return nil, fmt.Errorf("no such volume")
			}
			if args[0] == "--version" {
				return []byte("podman version 4.2.1\n"), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error for old Podman version")
	}
	if !strings.Contains(err.Error(), "podman >= 4.3 required") {
		t.Errorf("expected 'podman >= 4.3 required' in error, got: %s", err.Error())
	}
}

// --- isRootlessPodman tests (8.17, 8.18, 8.19) ---

func TestIsRootlessPodman_True(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("true\n"), nil
	}

	if !isRootlessPodman(opts) {
		t.Error("expected isRootlessPodman to return true")
	}
}

func TestIsRootlessPodman_False(t *testing.T) {
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		return []byte("false\n"), nil
	}

	if isRootlessPodman(opts) {
		t.Error("expected isRootlessPodman to return false")
	}
}

func TestStart_UIDMapRejectedUnderRootful(t *testing.T) {
	opts := testOpts()
	opts.Detach = true
	opts.UIDMap = true
	opts.Platform = &PlatformConfig{OS: "linux", Arch: "amd64", UIDMapSupported: true}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
			case "--version":
				return []byte("podman version 5.0.0\n"), nil
			case "info":
				return []byte("false\n"), nil // rootful
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err == nil {
		t.Fatal("expected error when --uidmap under rootful Podman")
	}
	if !strings.Contains(err.Error(), "only safe under rootless") {
		t.Errorf("expected 'only safe under rootless' in error, got: %s", err.Error())
	}
}

// --- PodmanBackend.Create chown tests (8.23, 8.24) ---

func TestPodmanCreate_ChownAfterCopy(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.Detach = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if len(args) > 1 && args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				return []byte("volume-created"), nil
			case "run":
				// UID mapping probe (busybox) or container start.
				for _, a := range args {
					if a == probeImage {
						return []byte("1000\n"), nil
					}
				}
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "exec":
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify chown is called after cp.
	cpIdx := -1
	chownIdx := -1
	for i, cmd := range commands {
		if strings.Contains(cmd, "podman cp") {
			cpIdx = i
		}
		if strings.Contains(cmd, "chown -R dev:dev /workspace") {
			chownIdx = i
		}
	}
	if cpIdx < 0 {
		t.Fatal("expected podman cp in commands")
	}
	if chownIdx < 0 {
		t.Fatal("expected chown command in commands")
	}
	if chownIdx <= cpIdx {
		t.Errorf("chown (idx %d) should come after cp (idx %d)", chownIdx, cpIdx)
	}

	// Verify stderr contains the progress message.
	stderrOut := opts.Stderr.(*bytes.Buffer).String()
	if !strings.Contains(stderrOut, "Fixing workspace permissions...") {
		t.Errorf("expected 'Fixing workspace permissions...' in stderr, got: %s", stderrOut)
	}
}

func TestPodmanCreate_ChownFailure(t *testing.T) {
	rmCalled := false
	volumeRmCalled := false
	opts := testOpts()
	opts.Detach = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if len(args) > 1 && args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				if len(args) > 1 && args[1] == "rm" {
					volumeRmCalled = true
					return []byte(""), nil
				}
				return []byte("volume-created"), nil
			case "run":
				// UID mapping probe (busybox) or container start.
				for _, a := range args {
					if a == probeImage {
						return []byte("1000\n"), nil
					}
				}
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "exec":
				// chown fails.
				return []byte("permission denied"), fmt.Errorf("exit 1")
			case "rm":
				rmCalled = true
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	b := &PodmanBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when chown fails")
	}
	if !strings.Contains(err.Error(), "failed to fix permissions") {
		t.Errorf("expected 'failed to fix permissions' in error, got: %s", err.Error())
	}
	if !rmCalled {
		t.Error("expected podman rm -f for partial cleanup")
	}
	if !volumeRmCalled {
		t.Error("expected podman volume rm for partial cleanup")
	}
}

// ============================================================
// Gateway Wiring in Dispatch Functions (Task Group 2)
// ============================================================

func TestCreate_AutoStartsGateway(t *testing.T) {
	gatewayStarted := false
	opts := testOpts()
	opts.Detach = true
	opts.BackendName = BackendPodman
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	healthCallCount := 0
	opts.HTTPGet = func(url string) (int, error) {
		healthCallCount++
		// Gateway health: first call fails, then succeeds.
		if strings.Contains(url, "53147") {
			if healthCallCount == 1 {
				return 0, fmt.Errorf("connection refused")
			}
			return 200, nil
		}
		// OpenCode server health: always succeeds.
		return 200, nil
	}

	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "uf" && len(args) > 0 && args[0] == "gateway" {
			gatewayStarted = true
			return []byte(""), nil
		}
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if len(args) > 1 && args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				return []byte(""), nil
			case "run":
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	err := Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gatewayStarted {
		t.Error("expected gateway to be auto-started in Create()")
	}

	stderrOut := opts.Stderr.(*bytes.Buffer).String()
	if !strings.Contains(stderrOut, "Gateway active") {
		t.Errorf("expected 'Gateway active' in stderr, got: %s", stderrOut)
	}
}

func TestStart_PersistentAutoStartsGateway(t *testing.T) {
	gatewayStarted := false
	opts := testOpts()
	opts.Detach = true
	opts.Getenv = func(key string) string {
		if key == "ANTHROPIC_API_KEY" {
			return "sk-ant-test"
		}
		return ""
	}

	healthCallCount := 0
	opts.HTTPGet = func(url string) (int, error) {
		healthCallCount++
		if strings.Contains(url, "53147") {
			if healthCallCount == 1 {
				return 0, fmt.Errorf("connection refused")
			}
			return 200, nil
		}
		return 200, nil
	}

	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "uf" && len(args) > 0 && args[0] == "gateway" {
			gatewayStarted = true
			return []byte(""), nil
		}
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if len(args) > 1 && args[1] == "inspect" {
					// Volume exists → persistent workspace.
					return []byte("{}"), nil
				}
				return []byte(""), nil
			case "inspect":
				return []byte("{}"), nil // container exists
			case "start":
				return []byte(""), nil
			}
		}
		return []byte(""), nil
	}

	err := Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gatewayStarted {
		t.Error("expected gateway to be auto-started for persistent Start()")
	}

	stderrOut := opts.Stderr.(*bytes.Buffer).String()
	if !strings.Contains(stderrOut, "Gateway active") {
		t.Errorf("expected 'Gateway active' in stderr, got: %s", stderrOut)
	}
}

// ============================================================
// DevPod Backend Tests (Task Group 4)
// ============================================================

func TestDevPodCreate_Success(t *testing.T) {
	var capturedArgs []string
	opts := testOpts()
	opts.Detach = true
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" {
			if len(args) > 0 && args[0] == "version" {
				return []byte("v0.5.18\n"), nil
			}
			if len(args) > 0 && args[0] == "up" {
				capturedArgs = args
				return []byte("workspace created"), nil
			}
		}
		return []byte(""), nil
	}
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "devcontainer.json") {
			return []byte(`{"image":"test"}`), nil
		}
		return nil, fmt.Errorf("not found")
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := strings.Join(capturedArgs, " ")
	if !strings.Contains(joined, "--provider podman") {
		t.Errorf("expected --provider podman, got: %s", joined)
	}
	if !strings.Contains(joined, "--id uf-sandbox-test-project") {
		t.Errorf("expected --id uf-sandbox-test-project, got: %s", joined)
	}
	if !strings.Contains(joined, "--ide none") {
		t.Errorf("expected --ide none, got: %s", joined)
	}
}

func TestDevPodCreate_NotInstalled(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		if name == "podman" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "version" {
			return []byte("v0.5.18\n"), nil
		}
		return []byte(""), nil
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when podman not installed")
	}
	if !strings.Contains(err.Error(), "podman not found") {
		t.Errorf("expected podman install hint, got: %s", err.Error())
	}
}

func TestDevPodStop_Success(t *testing.T) {
	stopCalled := false
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "stop" {
			stopCalled = true
			if args[1] != "uf-sandbox-test-project" {
				t.Errorf("expected workspace name uf-sandbox-test-project, got: %s", args[1])
			}
			return []byte(""), nil
		}
		return []byte(""), nil
	}

	b := &DevPodBackend{}
	err := b.Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stopCalled {
		t.Error("expected devpod stop to be called")
	}
	if !strings.Contains(stdout(opts), "state preserved") {
		t.Errorf("expected state preserved message, got: %s", stdout(opts))
	}
}

func TestDevPodDestroy_Success(t *testing.T) {
	var capturedArgs []string
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "delete" {
			capturedArgs = args
			return []byte(""), nil
		}
		return []byte(""), nil
	}

	b := &DevPodBackend{}
	err := b.Destroy(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout(opts), "Sandbox destroyed") {
		t.Errorf("expected destroyed message, got: %s", stdout(opts))
	}

	joined := strings.Join(capturedArgs, " ")
	if !strings.Contains(joined, "uf-sandbox-test-project") {
		t.Errorf("expected workspace name, got: %s", joined)
	}
	if !strings.Contains(joined, "--force") {
		t.Errorf("expected --force flag, got: %s", joined)
	}
}

func TestDevPodStatus_Running(t *testing.T) {
	statusJSON := `{"id":"abc123","state":"Running","provider":"podman","ide":"none"}`
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "status" {
			return []byte(statusJSON), nil
		}
		return []byte(""), nil
	}

	b := &DevPodBackend{}
	ws, err := b.Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Exists {
		t.Error("expected Exists=true")
	}
	if !ws.Running {
		t.Error("expected Running=true")
	}
	if ws.Backend != BackendDevPod {
		t.Errorf("expected devpod backend, got: %s", ws.Backend)
	}
	if ws.Name != "uf-sandbox-test-project" {
		t.Errorf("expected uf-sandbox-test-project, got: %s", ws.Name)
	}
	if ws.ID != "abc123" {
		t.Errorf("expected abc123, got: %s", ws.ID)
	}
	if !ws.Persistent {
		t.Error("expected Persistent=true")
	}
}

func TestDevPodStatus_Stopped(t *testing.T) {
	statusJSON := `{"id":"abc123","state":"Stopped","provider":"podman","ide":"none"}`
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "status" {
			return []byte(statusJSON), nil
		}
		return []byte(""), nil
	}

	b := &DevPodBackend{}
	ws, err := b.Status(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ws.Exists {
		t.Error("expected Exists=true")
	}
	if ws.Running {
		t.Error("expected Running=false for stopped workspace")
	}
}

func TestDevPodCreate_GatewayEnvInjection(t *testing.T) {
	var capturedArgs []string
	opts := testOpts()
	opts.Detach = true
	opts.GatewayActive = true
	opts.GatewayPort = GatewayDefaultPort
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" {
			if len(args) > 0 && args[0] == "version" {
				return []byte("v0.5.18\n"), nil
			}
			if len(args) > 0 && args[0] == "up" {
				capturedArgs = args
				return []byte("workspace created"), nil
			}
		}
		return []byte(""), nil
	}
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "devcontainer.json") {
			return []byte(`{"image":"test"}`), nil
		}
		return nil, fmt.Errorf("not found")
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	joined := strings.Join(capturedArgs, " ")
	if !strings.Contains(joined, "--workspace-env ANTHROPIC_BASE_URL=http://host.containers.internal:53147") {
		t.Errorf("expected ANTHROPIC_BASE_URL via --workspace-env, got: %s", joined)
	}
	if !strings.Contains(joined, "--workspace-env ANTHROPIC_API_KEY=gateway") {
		t.Errorf("expected ANTHROPIC_API_KEY=gateway via --workspace-env, got: %s", joined)
	}
}

func TestResolveBackend_DevPod(t *testing.T) {
	opts := testOpts()
	opts.BackendName = BackendDevPod
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	backend, err := ResolveBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendDevPod {
		t.Errorf("expected devpod backend, got: %s", backend.Name())
	}
}

func TestAutoDetect_PrefersDevPod(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "devcontainer.json") {
			return []byte(`{"image":"test"}`), nil
		}
		return nil, fmt.Errorf("not found")
	}

	backend, err := autoDetectBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendDevPod {
		t.Errorf("expected devpod backend, got: %s", backend.Name())
	}
}

func TestAutoDetect_FallsBackToPodman(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	// ReadFile returns error for devcontainer.json (not found).
	opts.ReadFile = func(path string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}

	backend, err := autoDetectBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendPodman {
		t.Errorf("expected podman backend (fallback), got: %s", backend.Name())
	}
}

func TestDevPodCreate_MissingDevcontainer(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "version" {
			return []byte("v0.5.18\n"), nil
		}
		return []byte(""), nil
	}
	// ReadFile returns error (no devcontainer.json).
	opts.ReadFile = func(path string) ([]byte, error) {
		return nil, fmt.Errorf("not found")
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when devcontainer.json missing")
	}
	if !strings.Contains(err.Error(), "devcontainer.json not found") {
		t.Errorf("expected devcontainer.json error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "uf sandbox init") {
		t.Errorf("expected uf sandbox init hint, got: %s", err.Error())
	}
}

func TestDevPodCreate_DevPodUpFails(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" {
			if len(args) > 0 && args[0] == "version" {
				return []byte("v0.5.18\n"), nil
			}
			if len(args) > 0 && args[0] == "up" {
				return []byte("provider error: podman crashed"), fmt.Errorf("exit 1")
			}
		}
		return []byte(""), nil
	}
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.Contains(path, "devcontainer.json") {
			return []byte(`{"image":"test"}`), nil
		}
		return nil, fmt.Errorf("not found")
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when devpod up fails")
	}
	if !strings.Contains(err.Error(), "devpod up failed") {
		t.Errorf("expected devpod up failed error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "provider error") {
		t.Errorf("expected devpod output in error, got: %s", err.Error())
	}
}

func TestDevPodCreate_PodmanNotInstalled(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		if name == "podman" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when podman not installed")
	}
	if !strings.Contains(err.Error(), "podman not found") {
		t.Errorf("expected podman not found error, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "DevPod requires Podman") {
		t.Errorf("expected DevPod-specific hint, got: %s", err.Error())
	}
}

func TestDevPodCreate_VersionTooOld(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "devpod" && len(args) > 0 && args[0] == "version" {
			return []byte("v0.4.9\n"), nil
		}
		return []byte(""), nil
	}

	b := &DevPodBackend{}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when DevPod version too old")
	}
	if !strings.Contains(err.Error(), "devpod >= 0.5.0 required") {
		t.Errorf("expected version requirement error, got: %s", err.Error())
	}
}

func TestDevPodAttach_Success(t *testing.T) {
	var attachArgs []string
	opts := testOpts()
	opts.ExecInteractive = func(name string, args ...string) error {
		attachArgs = append([]string{name}, args...)
		return nil
	}

	b := &DevPodBackend{}
	err := b.Attach(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attachArgs) < 3 {
		t.Fatalf("expected at least 3 args, got: %v", attachArgs)
	}
	if attachArgs[0] != "opencode" {
		t.Errorf("expected opencode, got: %s", attachArgs[0])
	}
	if attachArgs[1] != "attach" {
		t.Errorf("expected attach, got: %s", attachArgs[1])
	}
	if attachArgs[2] != "http://localhost:4096" {
		t.Errorf("expected http://localhost:4096, got: %s", attachArgs[2])
	}
}

func TestIsPersistentWorkspace_DevPod(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		// Podman volume inspect fails (no Podman workspace).
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			return nil, fmt.Errorf("no such volume")
		}
		// DevPod status succeeds (workspace exists).
		if name == "devpod" && len(args) > 0 && args[0] == "status" {
			return []byte(`{"id":"abc","state":"Running"}`), nil
		}
		return []byte(""), nil
	}

	if !isPersistentWorkspace(opts) {
		t.Error("expected isPersistentWorkspace=true for DevPod workspace")
	}
}

func TestExtract_PersistentWorkspace(t *testing.T) {
	// Test with Podman persistent workspace.
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			return []byte("{}"), nil // volume exists → persistent
		}
		return []byte(""), nil
	}

	err := Extract(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout(opts)
	if !strings.Contains(out, "persistent workspace") {
		t.Errorf("expected persistent workspace message, got: %s", out)
	}

	// Test with DevPod persistent workspace.
	opts2 := testOpts()
	opts2.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}
	opts2.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 && args[0] == "volume" {
			return nil, fmt.Errorf("no such volume")
		}
		if name == "devpod" && len(args) > 0 && args[0] == "status" {
			return []byte(`{"id":"abc","state":"Running"}`), nil
		}
		return []byte(""), nil
	}

	err = Extract(opts2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out2 := stdout(opts2)
	if !strings.Contains(out2, "persistent workspace") {
		t.Errorf("expected persistent workspace message for DevPod, got: %s", out2)
	}
}

func TestCreate_NoCloudProvider_SkipsGateway(t *testing.T) {
	gatewayStarted := false
	opts := testOpts()
	opts.Detach = true
	opts.BackendName = BackendPodman
	// No provider env vars set (default testOpts).

	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "uf" && len(args) > 0 && args[0] == "gateway" {
			gatewayStarted = true
			return []byte(""), nil
		}
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				if len(args) > 1 && args[1] == "inspect" {
					return nil, fmt.Errorf("no such volume")
				}
				return []byte(""), nil
			case "run":
				return []byte("container-id"), nil
			case "cp":
				return []byte(""), nil
			case "inspect":
				return nil, fmt.Errorf("no such container")
			}
		}
		return []byte(""), nil
	}

	err := Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gatewayStarted {
		t.Error("expected gateway NOT to be started when no cloud provider")
	}

	stderrOut := opts.Stderr.(*bytes.Buffer).String()
	if strings.Contains(stderrOut, "Gateway active") {
		t.Errorf("expected no 'Gateway active' in stderr, got: %s", stderrOut)
	}
}

// --- InitDevcontainer tests ---

// testDevcontainerTemplate returns a minimal devcontainer.json
// template for testing. Mirrors the embedded asset structure.
func testDevcontainerTemplate() []byte {
	return []byte(`{
  "_comment": "ANTHROPIC_API_KEY=gateway is a sentinel value.",
  "image": "quay.io/unbound-force/opencode-dev:latest",
  "forwardPorts": [4096],
  "containerEnv": {
    "ANTHROPIC_BASE_URL": "http://host.containers.internal:53147",
    "ANTHROPIC_API_KEY": "gateway"
  },
  "remoteUser": "dev"
}
`)
}

func TestRunSandboxInit_Creates(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := InitDevcontainer(InitDevcontainerOptions{
		ProjectDir:      dir,
		Stdout:          &buf,
		TemplateContent: testDevcontainerTemplate(),
	})
	if err != nil {
		t.Fatalf("InitDevcontainer error: %v", err)
	}

	// Verify file was created.
	outPath := filepath.Join(dir, ".devcontainer", "devcontainer.json")
	data, readErr := os.ReadFile(outPath)
	if readErr != nil {
		t.Fatalf("read devcontainer.json: %v", readErr)
	}

	// Parse and verify fields.
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("parse devcontainer.json: %v", jsonErr)
	}

	if img, ok := parsed["image"].(string); !ok || img != "quay.io/unbound-force/opencode-dev:latest" {
		t.Errorf("image = %v, want quay.io/unbound-force/opencode-dev:latest", parsed["image"])
	}

	ports, ok := parsed["forwardPorts"].([]interface{})
	if !ok || len(ports) != 1 {
		t.Errorf("forwardPorts = %v, want [4096]", parsed["forwardPorts"])
	} else if ports[0].(float64) != 4096 {
		t.Errorf("forwardPorts[0] = %v, want 4096", ports[0])
	}

	env, ok := parsed["containerEnv"].(map[string]interface{})
	if !ok {
		t.Fatal("containerEnv not a map")
	}
	if env["ANTHROPIC_API_KEY"] != "gateway" {
		t.Errorf("ANTHROPIC_API_KEY = %v, want gateway", env["ANTHROPIC_API_KEY"])
	}
	if env["ANTHROPIC_BASE_URL"] != "http://host.containers.internal:53147" {
		t.Errorf("ANTHROPIC_BASE_URL = %v, want gateway URL", env["ANTHROPIC_BASE_URL"])
	}

	if parsed["remoteUser"] != "dev" {
		t.Errorf("remoteUser = %v, want dev", parsed["remoteUser"])
	}

	// Verify _comment key is present (sentinel explanation).
	if _, hasComment := parsed["_comment"]; !hasComment {
		t.Error("expected _comment key explaining sentinel value")
	}

	// Verify output message.
	if !strings.Contains(buf.String(), "Created") {
		t.Errorf("expected 'Created' in output, got: %s", buf.String())
	}
}

func TestRunSandboxInit_ExistingSkips(t *testing.T) {
	dir := t.TempDir()

	// Create existing devcontainer.json.
	dcDir := filepath.Join(dir, ".devcontainer")
	if err := os.MkdirAll(dcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	existing := []byte(`{"image":"custom"}`)
	if err := os.WriteFile(filepath.Join(dcDir, "devcontainer.json"), existing, 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := InitDevcontainer(InitDevcontainerOptions{
		ProjectDir:      dir,
		Stdout:          &buf,
		TemplateContent: testDevcontainerTemplate(),
	})
	if err != nil {
		t.Fatalf("InitDevcontainer error: %v", err)
	}

	// Verify file was NOT overwritten.
	data, _ := os.ReadFile(filepath.Join(dcDir, "devcontainer.json"))
	if string(data) != string(existing) {
		t.Error("existing devcontainer.json should not be overwritten without --force")
	}

	// Verify skip message.
	if !strings.Contains(buf.String(), "already exists") {
		t.Errorf("expected 'already exists' in output, got: %s", buf.String())
	}
}

func TestRunSandboxInit_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()

	// Create existing devcontainer.json.
	dcDir := filepath.Join(dir, ".devcontainer")
	if err := os.MkdirAll(dcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dcDir, "devcontainer.json"), []byte(`{"image":"old"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := InitDevcontainer(InitDevcontainerOptions{
		ProjectDir:      dir,
		Force:           true,
		Stdout:          &buf,
		TemplateContent: testDevcontainerTemplate(),
	})
	if err != nil {
		t.Fatalf("InitDevcontainer error: %v", err)
	}

	// Verify file was overwritten with template content.
	data, _ := os.ReadFile(filepath.Join(dcDir, "devcontainer.json"))
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("parse overwritten devcontainer.json: %v", jsonErr)
	}
	if parsed["image"] != "quay.io/unbound-force/opencode-dev:latest" {
		t.Errorf("image = %v, want template default", parsed["image"])
	}

	// Verify overwrite message.
	if !strings.Contains(buf.String(), "Overwritten") {
		t.Errorf("expected 'Overwritten' in output, got: %s", buf.String())
	}
}

func TestRunSandboxInit_CustomDemoPorts(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := InitDevcontainer(InitDevcontainerOptions{
		ProjectDir:      dir,
		DemoPorts:       []int{3000, 8080},
		Stdout:          &buf,
		TemplateContent: testDevcontainerTemplate(),
	})
	if err != nil {
		t.Fatalf("InitDevcontainer error: %v", err)
	}

	// Verify forwardPorts includes default 4096 + custom ports.
	data, _ := os.ReadFile(filepath.Join(dir, ".devcontainer", "devcontainer.json"))
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("parse devcontainer.json: %v", jsonErr)
	}

	ports, ok := parsed["forwardPorts"].([]interface{})
	if !ok {
		t.Fatal("forwardPorts not an array")
	}
	if len(ports) != 3 {
		t.Fatalf("expected 3 ports (4096 + 3000 + 8080), got %d: %v", len(ports), ports)
	}

	// Verify all ports are present.
	portSet := make(map[float64]bool)
	for _, p := range ports {
		portSet[p.(float64)] = true
	}
	for _, expected := range []float64{4096, 3000, 8080} {
		if !portSet[expected] {
			t.Errorf("expected port %v in forwardPorts, got: %v", expected, ports)
		}
	}
}

// --- Helper ---

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
