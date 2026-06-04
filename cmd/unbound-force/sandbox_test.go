package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/sandbox"
)

// TestApplySandboxConfig_UIDMap verifies that applySandboxConfig
// propagates UIDMap from config to opts when the CLI flag is not
// set, and that the CLI flag takes precedence over config.
func TestApplySandboxConfig_UIDMap(t *testing.T) {
	// Create a temp project dir with .uf/config.yaml
	// containing uid_map: true.
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte("sandbox:\n  uid_map: true\n")
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Config sets UIDMap when CLI flag is not set.
	opts := sandbox.Options{
		ProjectDir: dir,
		UIDMap:     false,
	}
	var stderr bytes.Buffer
	applySandboxConfig(&opts, &stderr)
	if !opts.UIDMap {
		t.Error("expected UIDMap=true from config when CLI flag not set")
	}

	// Test 2: CLI flag (UIDMap=true) takes precedence — already true,
	// config should not change it.
	opts2 := sandbox.Options{
		ProjectDir: dir,
		UIDMap:     true,
	}
	applySandboxConfig(&opts2, &stderr)
	if !opts2.UIDMap {
		t.Error("expected UIDMap=true preserved when CLI flag already set")
	}

	// Test 3: Config with uid_map: false (default) does not
	// override CLI flag.
	dir2 := t.TempDir()
	ufDir2 := filepath.Join(dir2, ".uf")
	if err := os.MkdirAll(ufDir2, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData2 := []byte("sandbox:\n  uid_map: false\n")
	if err := os.WriteFile(filepath.Join(ufDir2, "config.yaml"), cfgData2, 0o644); err != nil {
		t.Fatal(err)
	}
	opts3 := sandbox.Options{
		ProjectDir: dir2,
		UIDMap:     true,
	}
	applySandboxConfig(&opts3, &stderr)
	if !opts3.UIDMap {
		t.Error("expected UIDMap=true preserved when config has uid_map: false")
	}
}

// --- Destroy confirmation tests (Task 6.4/D14) ---

func TestRunSandboxDestroy_EmptyInputCancels(t *testing.T) {
	var stdout bytes.Buffer
	p := sandboxDestroyParams{
		projectDir: "/tmp/test",
		yes:        false,
		stdout:     &stdout,
		stderr:     &bytes.Buffer{},
		stdin:      strings.NewReader("\n"),
	}
	err := runSandboxDestroy(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Cancelled.") {
		t.Errorf("expected 'Cancelled.' in output, got: %s", stdout.String())
	}
}

func TestRunSandboxDestroy_EOFCancels(t *testing.T) {
	var stdout bytes.Buffer
	p := sandboxDestroyParams{
		projectDir: "/tmp/test",
		yes:        false,
		stdout:     &stdout,
		stderr:     &bytes.Buffer{},
		stdin:      strings.NewReader(""),
	}
	err := runSandboxDestroy(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Cancelled.") {
		t.Errorf("expected 'Cancelled.' on EOF, got: %s", stdout.String())
	}
}

func TestRunSandboxDestroy_NoCancels(t *testing.T) {
	var stdout bytes.Buffer
	p := sandboxDestroyParams{
		projectDir: "/tmp/test",
		yes:        false,
		stdout:     &stdout,
		stderr:     &bytes.Buffer{},
		stdin:      strings.NewReader("n\n"),
	}
	err := runSandboxDestroy(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "Cancelled.") {
		t.Errorf("expected 'Cancelled.' on 'n', got: %s", stdout.String())
	}
}

// --- runSandboxStatus tests ---

func TestNewSandboxStatusCmd_Registered(t *testing.T) {
	cmd := newSandboxCmd()
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'status' subcommand to be registered on sandbox")
	}
}

func TestNewSandboxCmd_AllSubcommands(t *testing.T) {
	cmd := newSandboxCmd()
	expected := []string{"init", "create", "destroy", "start", "stop", "attach", "extract", "status"}

	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Use] = true
	}

	for _, want := range expected {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestApplySandboxConfig_AllFields(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte(`
sandbox:
  backend: devpod
  image: custom-image:latest
  ide: vscode
  mode: direct
  resources:
    memory: "16g"
    cpus: "8"
`)
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	opts := sandbox.Options{ProjectDir: dir}
	var stderr bytes.Buffer
	applySandboxConfig(&opts, &stderr)

	if opts.BackendName != "devpod" {
		t.Errorf("expected backend devpod, got: %s", opts.BackendName)
	}
	if opts.Image != "custom-image:latest" {
		t.Errorf("expected image custom-image:latest, got: %s", opts.Image)
	}
	if opts.IDE != "vscode" {
		t.Errorf("expected IDE vscode, got: %s", opts.IDE)
	}
	if opts.Mode != "direct" {
		t.Errorf("expected mode direct, got: %s", opts.Mode)
	}
	if opts.Memory != "16g" {
		t.Errorf("expected memory 16g, got: %s", opts.Memory)
	}
	if opts.CPUs != "8" {
		t.Errorf("expected cpus 8, got: %s", opts.CPUs)
	}
}

func TestApplySandboxConfig_CLIFlagPrecedence(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgData := []byte(`
sandbox:
  backend: devpod
  image: config-image:latest
`)
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), cfgData, 0o644); err != nil {
		t.Fatal(err)
	}

	// CLI flags already set — config should NOT override.
	opts := sandbox.Options{
		ProjectDir:  dir,
		BackendName: "podman",
		Image:       "cli-image:latest",
	}
	var stderr bytes.Buffer
	applySandboxConfig(&opts, &stderr)

	if opts.BackendName != "podman" {
		t.Errorf("expected CLI backend preserved, got: %s", opts.BackendName)
	}
	if opts.Image != "cli-image:latest" {
		t.Errorf("expected CLI image preserved, got: %s", opts.Image)
	}
}

func TestApplySandboxConfig_LegacyWarning(t *testing.T) {
	dir := t.TempDir()
	// Create legacy .uf/sandbox.yaml file.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ufDir, "sandbox.yaml"), []byte("backend: podman\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	opts := sandbox.Options{ProjectDir: dir}
	var stderr bytes.Buffer
	applySandboxConfig(&opts, &stderr)

	if !strings.Contains(stderr.String(), "deprecated") {
		t.Errorf("expected deprecation warning, got: %s", stderr.String())
	}
}

func TestApplySandboxConfig_NoConfig(t *testing.T) {
	dir := t.TempDir()
	opts := sandbox.Options{ProjectDir: dir}
	var stderr bytes.Buffer

	// Should not panic when no config exists.
	applySandboxConfig(&opts, &stderr)

	// BackendName should remain empty — config.Load returns
	// a Config with zero-value fields when no file exists.
	if opts.BackendName != "" {
		t.Errorf("expected empty backend, got: %s", opts.BackendName)
	}
	// No deprecation warning should be emitted.
	if strings.Contains(stderr.String(), "deprecated") {
		t.Errorf("expected no deprecation warning, got: %s", stderr.String())
	}
}
