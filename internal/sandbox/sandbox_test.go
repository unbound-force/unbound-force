package sandbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

// --- Helper: mock Options builder ---

// testOpts returns an Options struct with all dependencies
// injected as no-op/success mocks. Tests override specific
// fields to exercise error paths.
func testOpts() Options {
	return Options{
		ProjectDir: "/tmp/test-project",
		Mode:       ModeIsolated,
		Stdout:     &bytes.Buffer{},
		Stderr:     &bytes.Buffer{},
		Stdin:      strings.NewReader(""),
		LookPath:   func(name string) (string, error) { return "/usr/bin/" + name, nil },
		ExecCmd: func(name string, args ...string) ([]byte, error) {
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
	args := buildRunArgs(opts, platform)

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
	args := buildRunArgs(opts, platform)

	joined := strings.Join(args, " ")

	// Verify read-write mount (no :ro) for the project directory.
	// Note: Google Cloud credential mounts may include :ro — that's
	// correct. Only the project mount should be read-write.
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
	args := buildRunArgs(opts, platform)

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
	args := buildRunArgs(opts, platform)

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
		default:
			return ""
		}
	}

	args := forwardedEnvVars(opts)
	joined := strings.Join(args, " ")

	// Verify present API keys are forwarded.
	if !strings.Contains(joined, "-e ANTHROPIC_API_KEY") {
		t.Errorf("expected ANTHROPIC_API_KEY, got: %s", joined)
	}
	if !strings.Contains(joined, "-e OPENAI_API_KEY") {
		t.Errorf("expected OPENAI_API_KEY, got: %s", joined)
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
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return []byte("true"), nil
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

	// podman inspect returns error (no container).
	// podman image exists returns error (need pull).
	// podman pull succeeds.
	// podman run succeeds.
	opts.ExecCmd = func(name string, args ...string) ([]byte, error) {
		if name == "podman" && len(args) > 0 {
			switch args[0] {
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
	args := buildRunArgs(opts, platform)
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
	args := buildRunArgs(opts, platform)
	joined := strings.Join(args, " ")

	// Check project mount is read-write (no :ro on project path).
	// Google Cloud credential mounts may include :ro — that's correct.
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
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return nil, fmt.Errorf("no such container")
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
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return nil, fmt.Errorf("no such container")
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
		if name == "podman" && len(args) > 0 && args[0] == "inspect" {
			return nil, fmt.Errorf("no such container")
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
