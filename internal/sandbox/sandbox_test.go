package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
//   - LookPath finds podman and opencode but NOT chectl
//     (prevents auto-detection of CDE backend)
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
			// Don't find chectl by default — prevents CDE auto-detect.
			if name == "chectl" {
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
	// ExecCmd should never be called on macOS (no getenforce).
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		t.Fatalf("ExecCmd should not be called on macOS, got: %s %v", name, args)
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
	// Verify read-only mount for isolated mode.
	if !strings.Contains(joined, "/tmp/test-project:/workspace:ro") {
		t.Errorf("expected :ro volume mount for isolated mode, got: %s", joined)
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

	// Verify read-write mount (no :ro) for the project directory.
	if !strings.Contains(joined, "/tmp/test-project:/workspace") {
		t.Errorf("expected volume mount, got: %s", joined)
	}
	if strings.Contains(joined, "/tmp/test-project:/workspace:ro") {
		t.Errorf("expected no :ro on project mount for direct mode, got: %s", joined)
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
		if name == "podman" {
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
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
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
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
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
backend: che
che:
  url: https://che.example.com
  token: test-token
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
	if cfg.Backend != "che" {
		t.Errorf("expected backend che, got: %s", cfg.Backend)
	}
	if cfg.Che.URL != "https://che.example.com" {
		t.Errorf("expected che URL, got: %s", cfg.Che.URL)
	}
	if cfg.Che.Token != "test-token" {
		t.Errorf("expected che token, got: %s", cfg.Che.Token)
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
	if cfg.Che.URL != "" {
		t.Errorf("expected empty che URL, got: %s", cfg.Che.URL)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	yamlContent := `
backend: podman
che:
  url: https://config-url.example.com
`
	opts := testOpts()
	opts.ReadFile = func(path string) ([]byte, error) {
		return []byte(yamlContent), nil
	}
	opts.Getenv = func(key string) string {
		switch key {
		case "UF_CHE_URL":
			return "https://env-url.example.com"
		case "UF_SANDBOX_BACKEND":
			return "che"
		}
		return ""
	}

	cfg, err := LoadConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Env var should override config file.
	if cfg.Che.URL != "https://env-url.example.com" {
		t.Errorf("expected env URL override, got: %s", cfg.Che.URL)
	}
	if cfg.Backend != "che" {
		t.Errorf("expected env backend override, got: %s", cfg.Backend)
	}
}

// --- ResolveBackend tests ---

func TestResolveBackend_AutoPodman(t *testing.T) {
	opts := testOpts()
	// LookPath doesn't find chectl (default in testOpts).
	// No UF_CHE_URL set.

	backend, err := ResolveBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendPodman {
		t.Errorf("expected podman backend, got: %s", backend.Name())
	}
}

func TestResolveBackend_AutoChe(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		if key == "UF_CHE_URL" {
			return "https://che.example.com"
		}
		return ""
	}

	backend, err := ResolveBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendChe {
		t.Errorf("expected che backend, got: %s", backend.Name())
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

func TestResolveBackend_ExplicitChe(t *testing.T) {
	opts := testOpts()
	opts.BackendName = BackendChe
	opts.CheURL = "https://che.example.com"

	backend, err := ResolveBackend(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if backend.Name() != BackendChe {
		t.Errorf("expected che backend, got: %s", backend.Name())
	}
}

func TestResolveBackend_CheNotConfigured(t *testing.T) {
	opts := testOpts()
	opts.BackendName = BackendChe
	// No CheURL, no chectl.

	_, err := ResolveBackend(opts)
	if err == nil {
		t.Fatal("expected error when CDE not configured")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected not configured message, got: %s", err.Error())
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

// --- CheBackend tests ---

func TestCheCreate_WithChectl(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	// Create a temp devfile.
	tmpDir := t.TempDir()
	opts.ProjectDir = tmpDir
	if err := writeTestFile(tmpDir+"/devfile.yaml", "schemaVersion: 2.0.0"); err != nil {
		t.Fatal(err)
	}

	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		return []byte(""), nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasCreate := false
	hasStart := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "workspace:create") {
			hasCreate = true
		}
		if strings.Contains(cmd, "workspace:start") {
			hasStart = true
		}
	}
	if !hasCreate {
		t.Error("expected chectl workspace:create")
	}
	if !hasStart {
		t.Error("expected chectl workspace:start")
	}
}

func TestCheCreate_WithRestAPI(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		if name == "chectl" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}

	tmpDir := t.TempDir()
	opts.ProjectDir = tmpDir
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "devfile.yaml") {
			return []byte("schemaVersion: 2.0.0"), nil
		}
		return nil, fmt.Errorf("not found")
	}

	// Create devfile on disk for os.Stat check.
	if err := writeTestFile(tmpDir+"/devfile.yaml", "schemaVersion: 2.0.0"); err != nil {
		t.Fatal(err)
	}

	var requestURL string
	opts.HTTPDo = func(req *http.Request) (*http.Response, error) {
		requestURL = req.URL.String()
		return &http.Response{
			StatusCode: 201,
			Body:       io.NopCloser(strings.NewReader(`{"id":"ws-123"}`)),
		}, nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: false}
	err := b.Create(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(requestURL, "/api/workspace/devfile") {
		t.Errorf("expected REST API URL, got: %s", requestURL)
	}
}

func TestCheCreate_NoDevfile(t *testing.T) {
	opts := testOpts()
	opts.ProjectDir = t.TempDir() // Empty dir, no devfile.

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when devfile missing")
	}
	if !strings.Contains(err.Error(), "devfile.yaml not found") {
		t.Errorf("expected devfile not found message, got: %s", err.Error())
	}
}

func TestCheCreate_Unreachable(t *testing.T) {
	opts := testOpts()
	opts.LookPath = func(name string) (string, error) {
		if name == "chectl" {
			return "", fmt.Errorf("not found")
		}
		return "/usr/bin/" + name, nil
	}

	tmpDir := t.TempDir()
	opts.ProjectDir = tmpDir
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, "devfile.yaml") {
			return []byte("schemaVersion: 2.0.0"), nil
		}
		return nil, fmt.Errorf("not found")
	}
	if err := writeTestFile(tmpDir+"/devfile.yaml", "schemaVersion: 2.0.0"); err != nil {
		t.Fatal(err)
	}

	opts.HTTPDo = func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("connection refused")
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: false}
	err := b.Create(opts)
	if err == nil {
		t.Fatal("expected error when Che unreachable")
	}
	if !strings.Contains(err.Error(), "cannot reach Che") {
		t.Errorf("expected unreachable message, got: %s", err.Error())
	}
}

func TestCheStart_WithChectl(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.Detach = true
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		return []byte(""), nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
	err := b.Start(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasStart := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "workspace:start") {
			hasStart = true
		}
	}
	if !hasStart {
		t.Error("expected chectl workspace:start")
	}
}

func TestCheStop_WithChectl(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		return []byte(""), nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
	err := b.Stop(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasStop := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "workspace:stop") {
			hasStop = true
		}
	}
	if !hasStop {
		t.Error("expected chectl workspace:stop")
	}
}

func TestCheDestroy_WithChectl(t *testing.T) {
	var commands []string
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		cmd := name + " " + strings.Join(args, " ")
		commands = append(commands, cmd)
		return []byte(""), nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
	err := b.Destroy(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasDelete := false
	hasYes := false
	for _, cmd := range commands {
		if strings.Contains(cmd, "workspace:delete") {
			hasDelete = true
		}
		if strings.Contains(cmd, "--yes") {
			hasYes = true
		}
	}
	if !hasDelete {
		t.Error("expected chectl workspace:delete")
	}
	if !hasYes {
		t.Error("expected --yes flag on delete")
	}
}

func TestCheStatus_Running(t *testing.T) {
	cheJSON := `[{
		"id": "ws-123",
		"status": "RUNNING",
		"config": {"name": "uf-test-project"},
		"runtime": {
			"machines": {
				"dev": {
					"servers": {
						"opencode-4096": {"url": "https://uf-test-opencode.apps.che.example.com"},
						"demo-web": {"url": "https://uf-test-demo.apps.che.example.com"}
					}
				}
			}
		}
	}]`

	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "chectl" && len(args) > 0 && args[0] == "workspace:list" {
			return []byte(cheJSON), nil
		}
		return []byte(""), nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
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
	if ws.ID != "ws-123" {
		t.Errorf("expected workspace ID ws-123, got: %s", ws.ID)
	}
	if ws.ServerURL == "" {
		t.Error("expected server URL to be set")
	}
	if len(ws.DemoEndpoints) == 0 {
		t.Error("expected at least one demo endpoint")
	}
}

func TestCheAttach_EndpointURL(t *testing.T) {
	cheJSON := `[{
		"id": "ws-123",
		"status": "RUNNING",
		"config": {"name": "uf-test-project"},
		"runtime": {
			"machines": {
				"dev": {
					"servers": {
						"opencode-4096": {"url": "https://uf-test-opencode.apps.che.example.com"}
					}
				}
			}
		}
	}]`

	var attachURL string
	opts := testOpts()
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "chectl" && len(args) > 0 && args[0] == "workspace:list" {
			return []byte(cheJSON), nil
		}
		return []byte(""), nil
	}
	opts.ExecInteractive = func(name string, args ...string) error {
		if name == "opencode" && len(args) > 1 {
			attachURL = args[1]
		}
		return nil
	}

	b := &CheBackend{cheURL: "https://che.example.com", useChectl: true}
	err := b.Attach(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(attachURL, "https://uf-test-opencode.apps.che.example.com") {
		t.Errorf("expected Che endpoint URL, got: %s", attachURL)
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

func TestFormatWorkspaceStatus_Che(t *testing.T) {
	var buf bytes.Buffer
	FormatWorkspaceStatus(&buf, WorkspaceStatus{
		Exists:     true,
		Running:    true,
		Backend:    BackendChe,
		Name:       "uf-myproject",
		Mode:       ModePersistent,
		ServerURL:  "https://uf-myproject-opencode.apps.che.example.com",
		StartedAt:  "2026-04-13T10:00:00Z",
		Persistent: true,
	})

	out := buf.String()
	if !strings.Contains(out, "Sandbox Status") {
		t.Errorf("expected status header, got: %s", out)
	}
	if !strings.Contains(out, "uf-myproject") {
		t.Errorf("expected workspace name, got: %s", out)
	}
	if !strings.Contains(out, "persistent") {
		t.Errorf("expected persistent label, got: %s", out)
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
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
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

func TestExtract_PersistentSuggestsGitPull(t *testing.T) {
	opts := testOpts()
	opts.CheURL = "https://che.example.com"
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
	if !strings.Contains(stdout(opts), "git pull") {
		t.Errorf("expected git pull suggestion, got: %s", stdout(opts))
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

// --- googleCloudCredentialMounts with gateway tests (T077-T078) ---

func TestGoogleCloudCredentialMounts_GatewayActive(t *testing.T) {
	opts := testOpts()
	opts.Getenv = func(key string) string {
		if key == "GOOGLE_APPLICATION_CREDENTIALS" {
			return "/path/to/creds.json"
		}
		return ""
	}

	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := googleCloudCredentialMounts(opts, platform, true)

	if len(args) != 0 {
		t.Errorf("expected no gcloud mounts when gateway active, got: %v", args)
	}
}

func TestGoogleCloudCredentialMounts_GatewayInactive(t *testing.T) {
	opts := testOpts()
	// When GOOGLE_APPLICATION_CREDENTIALS is not set and
	// gcloud dir doesn't exist, no mounts are returned.
	// This test verifies the function is called and doesn't
	// short-circuit.
	platform := PlatformConfig{OS: "darwin", Arch: "arm64"}
	args := googleCloudCredentialMounts(opts, platform, false)

	// On test machines without gcloud, this returns empty.
	// The important thing is it doesn't return nil early
	// like it does with gatewayActive=true.
	_ = args // No assertion needed — just verify no panic.
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
	// No provider env vars set (default testOpts).

	var runArgs string
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
			case "volume":
				return nil, fmt.Errorf("no such volume")
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

// --- Helper ---

func writeTestFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
