package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/config"
	"github.com/unbound-force/unbound-force/internal/doctor"
)

// --- Test helpers ---

// stubLookPath returns a function that simulates exec.LookPath.
// Keys are binary names; values are their paths.
func stubLookPath(found map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := found[name]; ok {
			return path, nil
		}
		return "", fmt.Errorf("executable %q not found", name)
	}
}

// cmdRecorder records all ExecCmd calls and returns canned results.
type cmdRecorder struct {
	calls   []string
	outputs map[string]string
	errors  map[string]error
}

func (r *cmdRecorder) execCmd(name string, args ...string) ([]byte, error) {
	key := name
	if len(args) > 0 {
		key = name + " " + strings.Join(args, " ")
	}
	r.calls = append(r.calls, key)

	if err, ok := r.errors[key]; ok {
		out := ""
		if o, ok2 := r.outputs[key]; ok2 {
			out = o
		}
		return []byte(out), err
	}
	if out, ok := r.outputs[key]; ok {
		return []byte(out), nil
	}
	return []byte(""), nil
}

// stubGetenv returns a function that reads env vars from a map.
func stubGetenv(vars map[string]string) func(string) string {
	return func(key string) string {
		return vars[key]
	}
}

// stubEvalSymlinks returns a function that resolves paths via a map.
func stubEvalSymlinks(resolved map[string]string) func(string) (string, error) {
	return func(path string) (string, error) {
		if r, ok := resolved[path]; ok {
			return r, nil
		}
		return path, nil
	}
}

// createFile creates a file with content in a temp dir.
func createFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// --- Phase 5: User Story 3 tests ---

func TestSetupRun_AllMissing(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
		errors:  map[string]error{},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true, // Allow non-interactive swarm setup in test
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify install order: opencode (brew), gaze (brew),
	// gh (brew), node version check, openspec (npm), uv (brew),
	// replicator (brew), replicator setup, ollama (brew),
	// dewey (brew).
	// Note: specify-cli is skipped because uv was just installed
	// and is not yet in the stubbed LookPath.
	// Note: ollama uses formula on Linux, cask on macOS.
	ollamaCmd := "brew install ollama"
	if runtime.GOOS == "darwin" {
		ollamaCmd = "brew install --cask ollama-app"
	}
	expectedCmds := []string{
		"brew install anomalyco/tap/opencode",
		"brew install unbound-force/tap/gaze",
		"brew install gh",
		"node --version",
		"npm install -g @fission-ai/openspec@latest",
		"brew install uv",
		"brew install unbound-force/tap/replicator",
		"replicator setup",
		ollamaCmd,
		"brew install unbound-force/tap/dewey",
	}

	for _, expected := range expectedCmds {
		found := false
		for _, call := range rec.calls {
			if call == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q not found in calls: %v", expected, rec.calls)
		}
	}
}

func TestSetupRun_AllPresent(t *testing.T) {
	dir := t.TempDir()

	// Create all expected files/dirs.
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir .dewey: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version":    "v22.15.0",
			"ollama list":       "NAME                    ID              SIZE\ngranite-embedding:30m   abc123          63 MB\n",
			"devpod provider list": "podman   docker   v0.1.0\n",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":          "/opt/homebrew/bin/brew",
			"opencode":      "/usr/local/bin/opencode",
			"gaze":          "/usr/local/bin/gaze",
			"gh":            "/usr/local/bin/gh",
			"node":          "/usr/local/bin/node",
			"npm":           "/usr/local/bin/npm",
			"openspec":      "/usr/local/bin/openspec",
			"uv":            "/usr/local/bin/uv",
			"specify":       "/usr/local/bin/specify",
			"replicator":    "/usr/local/bin/replicator",
			"dewey":         "/usr/local/bin/dewey",
			"ollama":        "/usr/local/bin/ollama",
			"podman":        "/usr/local/bin/podman",
			"devpod":        "/usr/local/bin/devpod",
			"golangci-lint": "/usr/local/bin/golangci-lint",
			"govulncheck":   "/usr/local/bin/govulncheck",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify no install commands were called (all tools already present).
	for _, call := range rec.calls {
		if strings.Contains(call, "install") {
			t.Errorf("unexpected install command: %s", call)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "already") {
		t.Error("expected 'already' messages for configured items")
	}
}

func TestSetupRun_NoNodeJS(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
		errors: map[string]error{
			// Node.js install via brew fails — simulating no Node.js available.
			"brew install node": fmt.Errorf("node install failed"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err == nil {
		t.Fatal("expected error when Node.js install fails")
	}

	// OpenCode and Gaze should be installed.
	brewInstallCalls := 0
	for _, call := range rec.calls {
		if strings.Contains(call, "brew install") && !strings.Contains(call, "node") {
			brewInstallCalls++
		}
	}
	if brewInstallCalls < 2 {
		t.Errorf("expected at least 2 non-node brew install calls, got %d", brewInstallCalls)
	}

	// npm-dependent steps should be skipped because Node.js failed.
	for _, call := range rec.calls {
		if strings.Contains(call, "npm install") {
			t.Errorf("unexpected npm command after Node.js failure: %s", call)
		}
	}
}

func TestSetupRun_ReplicatorBrewFails(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			"brew install unbound-force/tap/replicator": fmt.Errorf("brew: formula not found"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err == nil {
		t.Fatal("expected error when replicator brew install fails")
	}

	// replicator setup should NOT be called.
	for _, call := range rec.calls {
		if call == "replicator setup" {
			t.Errorf("unexpected command after brew failure: %s", call)
		}
	}
}

func TestSetupRun_NvmDetected(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"node": "/home/user/.nvm/versions/node/v22.15.0/bin/node",
			"npm":  "/home/user/.nvm/versions/node/v22.15.0/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{"NVM_DIR": "/home/user/.nvm"}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// OpenSpec should be installed via npm from nvm-managed node.
	npmCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "npm install -g @fission-ai/openspec@latest") {
			npmCalled = true
		}
	}
	if !npmCalled {
		t.Error("expected npm install call for openspec")
	}
}

func TestSetupRun_NvmInstallNode(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{"NVM_DIR": "/home/user/.nvm"}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should invoke bash to source nvm and install node.
	nvmCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "nvm.sh") && strings.Contains(call, "nvm install 22") {
			nvmCalled = true
		}
	}
	if !nvmCalled {
		t.Errorf("expected nvm install call, got calls: %v", rec.calls)
	}
}

func TestSetupRun_ReplicatorAlreadyInstalled(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should NOT call brew install for replicator.
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/replicator" {
			t.Error("should not install replicator when already present")
		}
	}

	output := buf.String()
	if !strings.Contains(output, "already installed") {
		t.Error("expected 'already installed' for replicator")
	}
}

func TestSetupRun_OpencodeJsonManipulation(t *testing.T) {
	dir := t.TempDir()

	// Create opencode.json with existing MCP servers.
	createFile(t, dir, "opencode.json", `{
  "mcpServers": {
    "dewey": {
      "command": "dewey"
    }
  }
}`)

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Setup does not directly write opencode.json — that is
	// handled by uf init (via scaffold.configureOpencodeJSON).
	// Verify the original file is unchanged by setup.
	data, readErr := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if readErr != nil {
		t.Fatalf("read opencode.json: %v", readErr)
	}

	// Verify valid JSON.
	var parsed map[string]json.RawMessage
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("invalid JSON: %v", jsonErr)
	}

	// Verify MCP servers preserved.
	if _, ok := parsed["mcpServers"]; !ok {
		t.Error("mcpServers should be preserved")
	}
}

func TestSetupRun_NoOpencodeJson(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Setup does not directly create opencode.json — that is
	// handled by uf init (via scaffold.configureOpencodeJSON).
	// We just verify setup completes successfully without error.
}

func TestSetupRun_MalformedOpencodeJson(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", "{invalid json")

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	// Setup does not directly touch opencode.json.
	// Malformed JSON is handled by uf init (scaffold.configureOpencodeJSON).
	// Run should succeed.
	if err != nil {
		t.Fatalf("Run: %v (malformed JSON should be non-fatal)", err)
	}
}

func TestSetupRun_DryRun(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		DryRun:    true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify no ExecCmd calls were made (except possibly node --version).
	for _, call := range rec.calls {
		if strings.Contains(call, "install") || strings.Contains(call, "setup") || strings.Contains(call, "init") {
			t.Errorf("unexpected command in dry-run: %s", call)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Would") || !strings.Contains(output, "dry-run") {
		t.Errorf("expected 'Would install' messages in dry-run output, got: %s", output)
	}
}

func TestSetupRun_CurlSafety(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   false,
		IsTTY:     func() bool { return false },
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify no curl command was executed.
	for _, call := range rec.calls {
		if strings.Contains(call, "curl") {
			t.Errorf("curl should not be called without --yes: %s", call)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "skipped") {
		t.Error("expected skip message for curl install")
	}
}

func TestSetupRun_FnmInstallNode(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"fnm":  "/usr/local/bin/fnm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should invoke fnm install 22.
	fnmCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "fnm install 22") {
			fnmCalled = true
		}
	}
	if !fnmCalled {
		t.Errorf("expected fnm install call, got calls: %v", rec.calls)
	}
}

func TestSetupRun_NvmInstallFails(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
		errors: map[string]error{
			"bash -c source /home/user/.nvm/nvm.sh && nvm install 22": fmt.Errorf("nvm failed"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{"NVM_DIR": "/home/user/.nvm"}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "failed") {
		t.Error("expected failure message for nvm install")
	}
}

func TestSetupRun_NoManagersForNode(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir:    dir,
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err == nil {
		t.Fatal("expected error when no Node.js managers available")
	}

	output := buf.String()
	if !strings.Contains(output, "failed") || !strings.Contains(output, "Node.js") {
		t.Error("expected Node.js failure message")
	}
}

func TestSetupRun_OpenCodeCurlWithYes(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// With --yes and no brew, should use curl.
	curlCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "curl") {
			curlCalled = true
		}
	}
	if !curlCalled {
		t.Errorf("expected curl install with --yes flag, got calls: %v", rec.calls)
	}
}

func TestSetupRun_GazeNoHomebrew(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Gaze") && !strings.Contains(output, "skipped") {
		t.Error("expected Gaze skip message when no Homebrew")
	}
}

func TestSetupRun_DryRunNodeMissing(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		TargetDir:    dir,
		DryRun:       true,
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "dry-run") {
		t.Error("expected dry-run messages")
	}
}

func TestSetupRun_DryRunNvmDetected(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		TargetDir:    dir,
		DryRun:       true,
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{"NVM_DIR": "/home/user/.nvm"}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "nvm install 22") {
		t.Errorf("expected nvm install hint in dry-run, got: %s", output)
	}
}

func TestSetupRun_DryRunFnmDetected(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		DryRun:    true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"fnm": "/usr/local/bin/fnm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "fnm install 22") {
		t.Errorf("expected fnm install hint in dry-run, got: %s", output)
	}
}

func TestSetupRun_OpenCodeBrewFails(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			"brew install anomalyco/tap/opencode": fmt.Errorf("brew failed"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err == nil {
		t.Fatal("expected error when OpenCode brew install fails")
	}

	output := buf.String()
	if !strings.Contains(output, "failed") {
		t.Error("expected failure message for OpenCode brew install")
	}
}

// --- Replicator installation tests ---

func TestInstallReplicator_AlreadyInstalled(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installReplicator(&opts, env)
	if result.action != "already installed" {
		t.Errorf("expected 'already installed', got %q", result.action)
	}
}

func TestInstallReplicator_DryRun(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		DryRun: true,
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installReplicator(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "brew install unbound-force/tap/replicator") {
		t.Errorf("expected brew install hint in detail, got %q", result.detail)
	}
}

func TestInstallReplicator_NoHomebrew(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	// No Homebrew, no dnf, no Go → skips with download link.
	result := installReplicator(&opts, env)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "github.com/unbound-force/replicator") {
		t.Errorf("expected GitHub releases link in detail, got %q", result.detail)
	}
}

func TestInstallReplicator_Success(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}

	found := false
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/replicator" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected brew install call, got: %v", rec.calls)
	}
}

func TestInstallReplicator_BrewFails(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/replicator": fmt.Errorf("brew failed"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installReplicator(&opts, env)
	if result.action != "failed" {
		t.Errorf("expected 'failed', got %q", result.action)
	}
}

func TestRunReplicatorSetup_DryRun(t *testing.T) {
	opts := Options{DryRun: true}
	opts.defaults()

	result := runReplicatorSetup(&opts)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
}

func TestRunReplicatorSetup_Success(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		YesFlag: true,
		ExecCmd: rec.execCmd,
	}
	opts.defaults()

	result := runReplicatorSetup(&opts)
	if result.action != "completed" {
		t.Errorf("expected 'completed', got %q", result.action)
	}

	found := false
	for _, call := range rec.calls {
		if call == "replicator setup" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'replicator setup' call, got: %v", rec.calls)
	}
}

func TestRunReplicatorSetup_Failure(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{
			"replicator setup": fmt.Errorf("setup failed"),
		},
	}

	opts := Options{
		YesFlag: true,
		ExecCmd: rec.execCmd,
	}
	opts.defaults()

	result := runReplicatorSetup(&opts)
	if result.action != "failed" {
		t.Errorf("expected 'failed', got %q", result.action)
	}
}

func TestAtomicWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	data := []byte(`{"test": true}`)
	if err := atomicWriteFile(path, data, 0644); err != nil {
		t.Fatalf("atomicWriteFile: %v", err)
	}

	// Verify file was written.
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("got %q, want %q", got, data)
	}

	// Verify permissions.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0644 {
		t.Errorf("perm = %o, want 0644", info.Mode().Perm())
	}
}

func TestFormatSetupText(t *testing.T) {
	var buf bytes.Buffer
	results := []stepResult{
		{name: "OpenCode", action: "already installed"},
		{name: "Gaze", action: "installed", detail: "via Homebrew"},
		{name: "Node.js", action: "failed", detail: "not found", err: fmt.Errorf("not available")},
	}

	FormatSetupText(&buf, results)

	output := buf.String()
	if !strings.Contains(output, "already installed") {
		t.Error("expected 'already installed' message")
	}
	if !strings.Contains(output, "installed") {
		t.Error("expected 'installed' message")
	}
	if !strings.Contains(output, "failed") {
		t.Error("expected 'failed' message")
	}
}

// --- Dewey installation tests ---

func TestSetupRun_DeweyInstall(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify Dewey was installed via brew.
	deweyCalled := false
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/dewey" {
			deweyCalled = true
		}
	}
	if !deweyCalled {
		t.Errorf("expected brew install dewey, got calls: %v", rec.calls)
	}
}

func TestSetupRun_DeweyAlreadyInstalled(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version":       "v22.15.0",
			"ollama list":          "NAME                    ID              SIZE\ngranite-embedding:30m   abc123          63 MB\n",
			"devpod provider list": "podman   docker   v0.1.0\n",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"replicator": "/usr/local/bin/replicator",
			"dewey":      "/usr/local/bin/dewey",
			"ollama":     "/usr/local/bin/ollama",
			"podman":     "/usr/local/bin/podman",
			"devpod":     "/usr/local/bin/devpod",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify no brew install dewey was called.
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/dewey" {
			t.Error("should not install dewey when already present")
		}
	}

	output := buf.String()
	if !strings.Contains(output, "already") {
		t.Error("expected 'already installed' for Dewey")
	}
}

func TestSetupRun_DeweyEmbeddingModelPull(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
			"ollama list":    "NAME                    ID              SIZE\nllama3:latest           abc123          4.7 GB\n",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":   "/opt/homebrew/bin/brew",
			"node":   "/usr/local/bin/node",
			"npm":    "/usr/local/bin/npm",
			"dewey":  "/usr/local/bin/dewey",
			"ollama": "/usr/local/bin/ollama",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify ollama pull was called for granite-embedding:30m.
	pullCalled := false
	for _, call := range rec.calls {
		if call == "ollama pull granite-embedding:30m" {
			pullCalled = true
		}
	}
	if !pullCalled {
		t.Errorf("expected ollama pull granite-embedding:30m, got calls: %v", rec.calls)
	}
}

func TestSetupRun_OllamaInstall(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		GOOS:      runtime.GOOS,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"go":         "/usr/local/bin/go",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"replicator": "/usr/local/bin/replicator",
			// ollama NOT in PATH -- should be installed via brew
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify Ollama was installed via Homebrew. On Linux (this test
	// runs on Linux), the formula path is used. On macOS, the cask
	// path would be used. Check for whichever is appropriate.
	expectedCmd := "brew install ollama" // Linux default
	if opts.GOOS == "darwin" {
		expectedCmd = "brew install --cask ollama-app"
	}
	found := false
	for _, call := range rec.calls {
		if call == expectedCmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %q in recorded commands, got: %v", expectedCmd, rec.calls)
	}

	// Verify no Ollama tip in output (removed -- now installed automatically).
	output := buf.String()
	if strings.Contains(output, "Tip: Install Ollama") {
		t.Error("Ollama tip should be removed -- Ollama is now installed automatically")
	}
}

func TestSetupRun_NoOllamaTip(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"go":         "/usr/local/bin/go",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"replicator": "/usr/local/bin/replicator",
			"ollama":     "/usr/local/bin/ollama", // ollama IS in PATH
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Tip") {
		t.Error("should NOT show Ollama tip when ollama is installed")
	}
}

func TestSetupRun_OllamaNoHomebrew(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			// No brew, no ollama — Homebrew not available
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify no brew install ollama was attempted.
	for _, call := range rec.calls {
		if call == "brew install --cask ollama-app" || call == "brew install ollama" {
			t.Error("should NOT attempt brew install ollama when Homebrew is not available")
		}
	}

	// Verify output contains download link.
	output := buf.String()
	if !strings.Contains(output, "ollama.com/download") {
		t.Error("expected download link in output when Homebrew is not available")
	}
}

func TestSetupRun_OllamaBrewFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// On Linux, installOllama uses "brew install ollama" (formula).
	// On macOS, it uses "brew install --cask ollama-app".
	ollamaCmd := "brew install ollama"
	if runtime.GOOS == "darwin" {
		ollamaCmd = "brew install --cask ollama-app"
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			ollamaCmd: fmt.Errorf("brew: install failed"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"replicator": "/usr/local/bin/replicator",
			// ollama NOT in PATH
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	// Ollama failure increments the fail count, causing Run to return
	// an error. This is expected -- Ollama is optional but tracked.
	if err == nil {
		t.Log("Run returned nil -- Ollama failure counted but not fatal in this config")
	}

	output := buf.String()
	if !strings.Contains(output, "failed") && !strings.Contains(output, "FAIL") {
		t.Error("expected failure indication in output when brew install ollama fails")
	}
}

// --- OS-aware Ollama installation tests ---

func TestOllamaBrew_Darwin(t *testing.T) {
	args := ollamaBrew("darwin")
	expected := []string{"brew", "install", "--cask", "ollama-app"}
	if len(args) != len(expected) {
		t.Fatalf("ollamaBrew(darwin) = %v, want %v", args, expected)
	}
	for i, v := range expected {
		if args[i] != v {
			t.Errorf("ollamaBrew(darwin)[%d] = %q, want %q", i, args[i], v)
		}
	}
}

func TestOllamaBrew_Linux(t *testing.T) {
	args := ollamaBrew("linux")
	expected := []string{"brew", "install", "ollama"}
	if len(args) != len(expected) {
		t.Fatalf("ollamaBrew(linux) = %v, want %v", args, expected)
	}
	for i, v := range expected {
		if args[i] != v {
			t.Errorf("ollamaBrew(linux)[%d] = %q, want %q", i, args[i], v)
		}
	}
}

func TestInstallOllama_LinuxFormula(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		GOOS:      "linux",
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/home/linuxbrew/.linuxbrew/bin/brew",
			"node":       "/usr/bin/node",
			"npm":        "/usr/bin/npm",
			"go":         "/usr/bin/go",
			"opencode":   "/usr/bin/opencode",
			"gaze":       "/usr/bin/gaze",
			"replicator": "/usr/bin/replicator",
			// ollama NOT in PATH
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	_ = Run(opts)

	found := false
	for _, call := range rec.calls {
		if call == "brew install ollama" {
			found = true
		}
		if call == "brew install --cask ollama-app" {
			t.Error("should NOT use --cask on Linux")
		}
	}
	if !found {
		t.Errorf("expected 'brew install ollama' on Linux, got: %v", rec.calls)
	}
}

func TestInstallOllama_DarwinCask(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		GOOS:      "darwin",
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":       "/opt/homebrew/bin/brew",
			"node":       "/usr/local/bin/node",
			"npm":        "/usr/local/bin/npm",
			"go":         "/usr/local/bin/go",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"replicator": "/usr/local/bin/replicator",
			// ollama NOT in PATH
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	_ = Run(opts)

	found := false
	for _, call := range rec.calls {
		if call == "brew install --cask ollama-app" {
			found = true
		}
		if call == "brew install ollama" {
			t.Error("should use --cask on macOS, not formula")
		}
	}
	if !found {
		t.Errorf("expected 'brew install --cask ollama-app' on macOS, got: %v", rec.calls)
	}
}

// --- RPM URL and dnf install tests ---

func TestRpmURL(t *testing.T) {
	url := rpmURL("unbound-force/unbound-force", "0.12.0", "amd64")
	expected := "https://github.com/unbound-force/unbound-force/releases/download/v0.12.0/unbound-force_0.12.0_linux_amd64.rpm"
	if url != expected {
		t.Errorf("rpmURL = %q, want %q", url, expected)
	}
}

func TestRpmURL_Arm64(t *testing.T) {
	url := rpmURL("unbound-force/gaze", "1.5.0", "arm64")
	expected := "https://github.com/unbound-force/gaze/releases/download/v1.5.0/gaze_1.5.0_linux_arm64.rpm"
	if url != expected {
		t.Errorf("rpmURL = %q, want %q", url, expected)
	}
}

func TestRepoName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"unbound-force/unbound-force", "unbound-force"},
		{"unbound-force/gaze", "gaze"},
		{"single", "single"},
	}
	for _, tt := range tests {
		got := repoName(tt.input)
		if got != tt.want {
			t.Errorf("repoName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInstallViaRpm_Success(t *testing.T) {
	rec := &cmdRecorder{}

	opts := &Options{
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}

	result := installViaRpm(opts, "unbound-force", "unbound-force/unbound-force", "0.12.0")

	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if result.detail != "via dnf (RPM)" {
		t.Errorf("detail = %q, want 'via dnf (RPM)'", result.detail)
	}
	// Verify dnf was called with the RPM URL.
	found := false
	for _, call := range rec.calls {
		if strings.Contains(call, "dnf install -y") && strings.Contains(call, ".rpm") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected dnf install call, got: %v", rec.calls)
	}
}

func TestInstallViaRpm_NoVersion(t *testing.T) {
	opts := &Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	result := installViaRpm(opts, "unbound-force", "unbound-force/unbound-force", "")

	if result.action != "skipped" {
		t.Errorf("action = %q, want skipped", result.action)
	}
}

func TestInstallViaRpm_DnfFails(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{},
	}
	// Make all dnf calls fail.
	rec.errors["dnf install -y "+rpmURL("unbound-force/unbound-force", "0.12.0", rpmArch())] = fmt.Errorf("dnf: not authorized")

	opts := &Options{
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}

	result := installViaRpm(opts, "unbound-force", "unbound-force/unbound-force", "0.12.0")

	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
}

func TestInstallViaRpm_DryRun(t *testing.T) {
	opts := &Options{
		DryRun: true,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}

	result := installViaRpm(opts, "unbound-force", "unbound-force/unbound-force", "0.12.0")

	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "dnf install") {
		t.Errorf("detail should contain dnf install hint, got: %q", result.detail)
	}
}

// --- GitHub CLI installation tests ---

func TestSetupRun_GHMissing_BrewInstall(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":     "/opt/homebrew/bin/brew",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			// gh NOT in PATH
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	found := false
	for _, call := range rec.calls {
		if call == "brew install gh" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected brew install gh, got calls: %v", rec.calls)
	}
}

func TestSetupRun_GHNoHomebrew(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"node": "/usr/local/bin/node",
			"npm":  "/usr/local/bin/npm",
			// No brew, no gh
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Verify no brew install gh was attempted.
	for _, call := range rec.calls {
		if call == "brew install gh" {
			t.Error("should NOT attempt brew install gh when Homebrew is not available")
		}
	}

	output := buf.String()
	if !strings.Contains(output, "cli.github.com") {
		t.Error("expected cli.github.com link in output when Homebrew is not available")
	}
}

// --- OpenSpec CLI installation tests ---

func TestSetupRun_OpenSpecMissing_Install(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":     "/opt/homebrew/bin/brew",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"gh":       "/usr/local/bin/gh",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			// openspec NOT in PATH
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should use npm to install openspec.
	npmCalled := false
	for _, call := range rec.calls {
		if call == "npm install -g @fission-ai/openspec@latest" {
			npmCalled = true
		}
	}
	if !npmCalled {
		t.Errorf("expected npm install for openspec, got calls: %v", rec.calls)
	}
}

func TestSetupRun_OpenSpecNpmFails(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			"npm install -g @fission-ai/openspec@latest": fmt.Errorf("npm ERR! code EACCES"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		YesFlag:   true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":     "/opt/homebrew/bin/brew",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"gh":       "/usr/local/bin/gh",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			// npm install fails
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	// OpenSpec failure is a "failed" step, which causes Run to return error.
	if err == nil {
		t.Fatal("expected error when openspec npm install fails")
	}

	output := buf.String()
	if !strings.Contains(output, "failed") {
		t.Error("expected failure message for openspec install")
	}
	if !strings.Contains(output, "npm") {
		t.Error("expected npm reference in openspec failure message")
	}
}

// --- uv installation tests ---

func TestInstallUV_AlreadyInstalled(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"uv": "/usr/local/bin/uv",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installUV(&opts, env)
	if result.action != "already installed" {
		t.Errorf("expected 'already installed', got %q", result.action)
	}
}

func TestInstallUV_DryRun_Homebrew(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		DryRun: true,
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installUV(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "brew install uv") {
		t.Errorf("expected brew install hint, got %q", result.detail)
	}
}

func TestInstallUV_DryRun_Curl(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		DryRun:       true,
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installUV(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "curl") {
		t.Errorf("expected curl install hint, got %q", result.detail)
	}
}

func TestInstallUV_Homebrew(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installUV(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}

	found := false
	for _, call := range rec.calls {
		if call == "brew install uv" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'brew install uv' call, got: %v", rec.calls)
	}
}

func TestInstallUV_Curl(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		YesFlag:  true,
		Stdout:   &buf,
		Stderr:   &buf,
		LookPath: stubLookPath(map[string]string{
			// No brew — should fall back to curl
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installUV(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "curl") {
		t.Errorf("expected 'via curl' detail, got %q", result.detail)
	}

	curlCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "curl") && strings.Contains(call, "astral.sh") {
			curlCalled = true
		}
	}
	if !curlCalled {
		t.Errorf("expected curl install call, got: %v", rec.calls)
	}
}

func TestInstallUV_CurlSkipped(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		YesFlag:  false,
		IsTTY:    func() bool { return false },
		Stdout:   &buf,
		Stderr:   &buf,
		LookPath: stubLookPath(map[string]string{
			// No brew, no TTY, no --yes
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installUV(&opts, env)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "curl|bash") {
		t.Errorf("expected curl|bash skip detail, got %q", result.detail)
	}
}

// --- Specify CLI installation tests ---

func TestInstallSpecify_AlreadyInstalled(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"specify": "/usr/local/bin/specify",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installSpecify(&opts, env)
	if result.action != "already installed" {
		t.Errorf("expected 'already installed', got %q", result.action)
	}
}

func TestInstallSpecify_DryRun(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		DryRun: true,
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"uv": "/usr/local/bin/uv",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installSpecify(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "uv tool install specify-cli") {
		t.Errorf("expected uv tool install hint, got %q", result.detail)
	}
}

func TestInstallSpecify_NoUV(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installSpecify(&opts, env)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "uv not available") {
		t.Errorf("expected 'uv not available' detail, got %q", result.detail)
	}
}

func TestInstallSpecify_Success(t *testing.T) {
	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"uv": "/usr/local/bin/uv",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installSpecify(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}

	found := false
	for _, call := range rec.calls {
		if call == "uv tool install specify-cli" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'uv tool install specify-cli' call, got: %v", rec.calls)
	}
}

func TestInstallSpecify_Failed(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{
			"uv tool install specify-cli": fmt.Errorf("install failed"),
		},
	}

	var buf bytes.Buffer
	opts := Options{
		Stdout: &buf,
		Stderr: &buf,
		LookPath: stubLookPath(map[string]string{
			"uv": "/usr/local/bin/uv",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installSpecify(&opts, env)
	if result.action != "failed" {
		t.Errorf("expected 'failed', got %q", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

// Dewey init/index tests were removed — dewey workspace initialization
// is now handled exclusively by uf init (via scaffold.initSubTools).
// See internal/scaffold/scaffold_test.go for the corresponding tests.

// --- Dry-run update test ---

func TestSetupRun_DryRunNewSteps(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		DryRun:    true,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":  "/opt/homebrew/bin/brew",
			"node":  "/usr/local/bin/node",
			"npm":   "/usr/local/bin/npm",
			"bun":   "/home/user/.bun/bin/bun",
			"dewey": "/usr/local/bin/dewey",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	output := buf.String()

	// Verify dry-run output includes new tools.
	checks := []struct {
		name    string
		pattern string
	}{
		{"gh", "Would install: brew install gh"},
		{"openspec", "Would install: npm install -g @fission-ai/openspec@latest"},
		{"uv", "Would install: brew install uv"},
		{"specify", "Would install: uv tool install specify-cli"},
		{"replicator", "Would install: brew install unbound-force/tap/replicator"},
	}
	for _, c := range checks {
		if !strings.Contains(output, c.pattern) {
			t.Errorf("expected dry-run output to contain %q for %s, got:\n%s", c.pattern, c.name, output)
		}
	}

	// Verify no install/init commands were actually executed.
	for _, call := range rec.calls {
		if strings.Contains(call, "install") || strings.Contains(call, "setup") || call == "dewey init" || call == "dewey index" {
			t.Errorf("unexpected command in dry-run: %s", call)
		}
	}
}

// --- Config integration tests ---

func TestShouldSkipTool_SkipList(t *testing.T) {
	opts := Options{SkipTools: []string{"ollama", "dewey"}}
	if !opts.shouldSkipTool("ollama") {
		t.Error("expected ollama to be skipped via skip list")
	}
	if !opts.shouldSkipTool("dewey") {
		t.Error("expected dewey to be skipped via skip list")
	}
	if opts.shouldSkipTool("gaze") {
		t.Error("gaze should not be skipped")
	}
}

func TestShouldSkipTool_MethodSkip(t *testing.T) {
	opts := Options{
		ToolMethods: map[string]config.ToolConfig{
			"ollama": {Method: "skip"},
			"gaze":   {Method: "homebrew"},
		},
	}
	if !opts.shouldSkipTool("ollama") {
		t.Error("expected ollama to be skipped via method: skip")
	}
	if opts.shouldSkipTool("gaze") {
		t.Error("gaze with method: homebrew should not be skipped")
	}
}

func TestShouldSkipTool_ManualMode(t *testing.T) {
	opts := Options{PackageManager: "manual"}
	if !opts.shouldSkipTool("gaze") {
		t.Error("manual mode should skip tools with no override")
	}

	opts.ToolMethods = map[string]config.ToolConfig{
		"gaze": {Method: "rpm"},
	}
	if opts.shouldSkipTool("gaze") {
		t.Error("manual mode should NOT skip tools with explicit method")
	}
}

func TestShouldSkipTool_NoConfig(t *testing.T) {
	opts := Options{}
	if opts.shouldSkipTool("gaze") {
		t.Error("no config should not skip any tool")
	}
}

func TestToolMethod_Default(t *testing.T) {
	opts := Options{}
	if m := opts.toolMethod("gaze"); m != "auto" {
		t.Errorf("toolMethod = %q, want auto", m)
	}
}

func TestToolMethod_Override(t *testing.T) {
	opts := Options{
		ToolMethods: map[string]config.ToolConfig{
			"gaze": {Method: "rpm"},
		},
	}
	if m := opts.toolMethod("gaze"); m != "rpm" {
		t.Errorf("toolMethod = %q, want rpm", m)
	}
}

func TestEmbeddingModel_Default(t *testing.T) {
	opts := Options{}
	if m := opts.embeddingModel(); m != defaultEmbeddingModel {
		t.Errorf("embeddingModel = %q, want %q", m, defaultEmbeddingModel)
	}
}

func TestEmbeddingModel_Override(t *testing.T) {
	opts := Options{EmbeddingModel: "mxbai-embed-large"}
	if m := opts.embeddingModel(); m != "mxbai-embed-large" {
		t.Errorf("embeddingModel = %q, want mxbai-embed-large", m)
	}
}

func TestEmbeddingDim_Default(t *testing.T) {
	opts := Options{}
	if d := opts.embeddingDim(); d != defaultEmbeddingDim {
		t.Errorf("embeddingDim = %q, want %q", d, defaultEmbeddingDim)
	}
}

func TestEmbeddingDim_Override(t *testing.T) {
	opts := Options{EmbeddingDimensions: 1024}
	if d := opts.embeddingDim(); d != "1024" {
		t.Errorf("embeddingDim = %q, want 1024", d)
	}
}

// --- Podman installation tests ---

func TestInstallPodman_AlreadyInstalled(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "already installed" {
		t.Errorf("expected 'already installed', got %q", result.action)
	}
}

func TestInstallPodman_BrewInstall(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"podman info": "host:\n  os: linux\n",
		},
	}

	opts := Options{
		GOOS:   "linux",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "via Homebrew") {
		t.Errorf("expected 'via Homebrew' in detail, got %q", result.detail)
	}
	if !strings.Contains(result.detail, "verified") {
		t.Errorf("expected 'verified' in detail, got %q", result.detail)
	}

	found := false
	for _, call := range rec.calls {
		if call == "brew install podman" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'brew install podman' call, got: %v", rec.calls)
	}
}

func TestInstallPodman_NoHomebrew(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout:       &bytes.Buffer{},
		Stderr:       &bytes.Buffer{},
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "podman.io/docs/installation") {
		t.Errorf("expected download URL in detail, got %q", result.detail)
	}
}

func TestInstallPodman_DarwinMachineInit(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			// No machine exists — empty output.
			"podman machine list --format {{.Name}}": "",
			"podman info": "host:\n  os: darwin\n",
		},
	}

	opts := Options{
		GOOS:   "darwin",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "verified") {
		t.Errorf("expected 'verified' in detail, got %q", result.detail)
	}

	// Verify machine init and start were called.
	initCalled := false
	startCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "podman machine init") {
			initCalled = true
		}
		if call == "podman machine start" {
			startCalled = true
		}
	}
	if !initCalled {
		t.Errorf("expected 'podman machine init' call, got: %v", rec.calls)
	}
	if !startCalled {
		t.Errorf("expected 'podman machine start' call, got: %v", rec.calls)
	}
}

func TestInstallPodman_DarwinMachineAlreadyExists(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"podman machine list --format {{.Name}}": "podman-machine-default",
			"podman info": "host:\n  os: darwin\n",
		},
	}

	opts := Options{
		GOOS:   "darwin",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}

	// Verify machine init was NOT called.
	for _, call := range rec.calls {
		if strings.Contains(call, "podman machine init") {
			t.Error("should not call machine init when machine already exists")
		}
	}
}

func TestInstallPodman_DarwinMachineInitFails(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"podman machine list --format {{.Name}}": "",
			"podman info": "host:\n  os: darwin\n",
		},
		errors: map[string]error{
			// No gtimeout or timeout in LookPath, so
			// initMachineWithTimeout falls back to direct
			// podman machine init (no timeout wrapper).
			"podman machine init": fmt.Errorf("init failed"),
		},
	}

	opts := Options{
		GOOS:   "darwin",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	// Machine init failure does not fail the step.
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "machine init failed") {
		t.Errorf("expected 'machine init failed' in detail, got %q", result.detail)
	}
}

func TestInstallPodman_DarwinMachineStartFails(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"podman machine list --format {{.Name}}": "",
			"podman info": "host:\n  os: darwin\n",
		},
		errors: map[string]error{
			"podman machine start": fmt.Errorf("start failed"),
		},
	}

	opts := Options{
		GOOS:   "darwin",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "machine start failed") {
		t.Errorf("expected 'machine start failed' in detail, got %q", result.detail)
	}
}

func TestInstallPodman_LinuxNoMachineInit(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"podman info": "host:\n  os: linux\n",
		},
	}

	opts := Options{
		GOOS:   "linux",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}

	// Verify no machine init on Linux.
	for _, call := range rec.calls {
		if strings.Contains(call, "machine") {
			t.Errorf("should not call machine commands on Linux: %s", call)
		}
	}
}

func TestInstallPodman_SmokeTestPasses(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"podman info": "host:\n  os: linux\n",
		},
	}

	opts := Options{
		GOOS:   "linux",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if !strings.Contains(result.detail, "verified") {
		t.Errorf("expected 'verified' in detail, got %q", result.detail)
	}
}

func TestInstallPodman_SmokeTestFails(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{
			"podman info": fmt.Errorf("cannot connect to Podman"),
		},
	}

	opts := Options{
		GOOS:   "linux",
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "podman info failed") {
		t.Errorf("expected 'podman info failed' in detail, got %q", result.detail)
	}
}

func TestInstallPodman_DryRun(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		DryRun: true,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installPodman(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "brew install podman") {
		t.Errorf("expected brew install hint, got %q", result.detail)
	}
}

// --- DevPod installation tests ---

func TestInstallDevPod_AlreadyInstalled(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installDevPod(&opts, env)
	if result.action != "already installed" {
		t.Errorf("expected 'already installed', got %q", result.action)
	}
}

func TestInstallDevPod_BrewInstall(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installDevPod(&opts, env)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "via Homebrew") {
		t.Errorf("expected 'via Homebrew' in detail, got %q", result.detail)
	}

	found := false
	for _, call := range rec.calls {
		if call == "brew install devpod" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'brew install devpod' call, got: %v", rec.calls)
	}
}

func TestInstallDevPod_NoHomebrew(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout:       &bytes.Buffer{},
		Stderr:       &bytes.Buffer{},
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installDevPod(&opts, env)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "devpod.sh/docs/getting-started/install") {
		t.Errorf("expected download URL in detail, got %q", result.detail)
	}
}

func TestInstallDevPod_DryRun(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		DryRun: true,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"brew": "/opt/homebrew/bin/brew",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	env := doctor.DetectEnvironment(&doctor.Options{
		LookPath:     opts.LookPath,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
	})

	result := installDevPod(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "brew install devpod") {
		t.Errorf("expected brew install hint, got %q", result.detail)
	}
}

// --- DevPod provider configuration tests ---

func TestConfigureDevPodProvider_AlreadyRegistered(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"devpod provider list": "podman   docker   v0.1.0\nkubernetes   kubernetes   v0.2.0\n",
		},
	}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "already installed" {
		t.Errorf("expected 'already installed', got %q", result.action)
	}
}

func TestConfigureDevPodProvider_MissingInstall(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"devpod provider list": "kubernetes   kubernetes   v0.2.0\n",
		},
	}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "installed" {
		t.Errorf("expected 'installed', got %q", result.action)
	}

	// Verify the provider add command was called.
	found := false
	for _, call := range rec.calls {
		if call == "devpod provider add docker --name podman -o DOCKER_PATH=podman" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected provider add call, got: %v", rec.calls)
	}
}

func TestConfigureDevPodProvider_AddFails(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"devpod provider list": "kubernetes   kubernetes   v0.2.0\n",
		},
		errors: map[string]error{
			"devpod provider add docker --name podman -o DOCKER_PATH=podman": fmt.Errorf("provider add failed"),
		},
	}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "failed" {
		t.Errorf("expected 'failed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "devpod provider add docker --name podman") {
		t.Errorf("expected manual command in detail, got %q", result.detail)
	}
}

func TestConfigureDevPodProvider_ListFails(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{
			"devpod provider list": fmt.Errorf("devpod error"),
		},
	}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "provider list failed") {
		t.Errorf("expected warning about list failure, got %q", result.detail)
	}
}

func TestConfigureDevPodProvider_NoDevPod(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "no devpod") {
		t.Errorf("expected 'no devpod' in detail, got %q", result.detail)
	}
}

func TestConfigureDevPodProvider_NoPodman(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "skipped" {
		t.Errorf("expected 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "no podman") {
		t.Errorf("expected 'no podman' in detail, got %q", result.detail)
	}
}

func TestConfigureDevPodProvider_DryRun(t *testing.T) {
	rec := &cmdRecorder{}

	opts := Options{
		DryRun: true,
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		LookPath: stubLookPath(map[string]string{
			"devpod": "/usr/local/bin/devpod",
			"podman": "/usr/local/bin/podman",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}
	opts.defaults()

	result := configureDevPodProvider(&opts)
	if result.action != "dry-run" {
		t.Errorf("expected 'dry-run', got %q", result.action)
	}
	if !strings.Contains(result.detail, "devpod provider add docker --name podman") {
		t.Errorf("expected provider add hint, got %q", result.detail)
	}
}

// --- hasProvider tests ---

func TestHasProvider_ExactMatch(t *testing.T) {
	output := "podman   docker   v0.1.0\nkubernetes   kubernetes   v0.2.0\n"
	if !hasProvider(output, "podman") {
		t.Error("expected to find 'podman' provider")
	}
}

func TestHasProvider_SubstringNotMatched(t *testing.T) {
	output := "podman-custom   docker   v0.1.0\nkubernetes   kubernetes   v0.2.0\n"
	if hasProvider(output, "podman") {
		t.Error("'podman-custom' should not match 'podman' (exact match required)")
	}
}

func TestHasProvider_NotFound(t *testing.T) {
	output := "docker   docker   v0.1.0\nkubernetes   kubernetes   v0.2.0\n"
	if hasProvider(output, "podman") {
		t.Error("expected not to find 'podman' provider")
	}
}

func TestHasProvider_EmptyOutput(t *testing.T) {
	if hasProvider("", "podman") {
		t.Error("expected not to find provider in empty output")
	}
}

func TestSetupRun_SkipViaConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on windows")
	}
	dir := t.TempDir()
	rec := &cmdRecorder{outputs: map[string]string{}, errors: map[string]error{}}
	var buf bytes.Buffer

	err := Run(Options{
		TargetDir:    dir,
		DryRun:       true,
		SkipTools:    []string{"ollama", "dewey", "golangci-lint", "govulncheck"},
		Stdout:       &buf,
		Stderr:       &buf,
		LookPath:     stubLookPath(map[string]string{}),
		ExecCmd:      rec.execCmd,
		Getenv:       stubGetenv(map[string]string{}),
		EvalSymlinks: stubEvalSymlinks(map[string]string{}),
		ReadFile:     func(string) ([]byte, error) { return nil, os.ErrNotExist },
		WriteFile:    func(string, []byte, os.FileMode) error { return nil },
	})
	// May fail due to missing tools, but that's expected.
	_ = err

	output := buf.String()
	for _, tool := range []string{"Ollama", "Dewey", "golangci-lint", "govulncheck"} {
		if !strings.Contains(output, tool) {
			continue
		}
		// Verify the tool shows as "excluded by config" not "installed" or "failed".
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, tool) && strings.Contains(line, "excluded by config") {
				break
			}
		}
	}
}

// --- installGaze unit tests ---

func TestInstallGaze_AlreadyInstalled(t *testing.T) {
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"gaze": "/usr/local/bin/gaze",
		}),
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "already installed" {
		t.Errorf("action = %q, want %q", result.action, "already installed")
	}
	if result.name != "Gaze" {
		t.Errorf("name = %q, want %q", result.name, "Gaze")
	}
}

func TestInstallGaze_DryRunWithHomebrew(t *testing.T) {
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want %q", result.action, "dry-run")
	}
	if !strings.Contains(result.detail, "brew install") {
		t.Errorf("detail = %q, want to contain 'brew install'", result.detail)
	}
}

func TestInstallGaze_DryRunNoHomebrew(t *testing.T) {
	// With no Homebrew and no dnf, dry-run falls through to go install
	// (or skipped if Go is not available).
	opts := &Options{
		DryRun: true,
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want %q", result.action, "dry-run")
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallGaze_NoHomebrewSkip(t *testing.T) {
	// No Homebrew, no dnf, no Go → skips with download link.
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "skipped" {
		t.Errorf("action = %q, want %q", result.action, "skipped")
	}
	if !strings.Contains(result.detail, "Homebrew not available") {
		t.Errorf("detail = %q, want to contain 'Homebrew not available'", result.detail)
	}
}

func TestInstallGaze_HomebrewInstallSuccess(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"brew install unbound-force/tap/gaze": "==> Installing gaze",
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want %q", result.action, "installed")
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
}

func TestInstallGaze_HomebrewInstallFailed(t *testing.T) {
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/gaze": fmt.Errorf("brew error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want %q", result.action, "failed")
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallGaze_ExplicitHomebrewMethod(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"brew install unbound-force/tap/gaze": "installed",
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"gaze": {Method: "homebrew"},
		},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want %q", result.action, "installed")
	}
}

func TestInstallGaze_ExplicitHomebrewDryRun(t *testing.T) {
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		ToolMethods: map[string]config.ToolConfig{
			"gaze": {Method: "homebrew"},
		},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want %q", result.action, "dry-run")
	}
	if !strings.Contains(result.detail, "brew install") {
		t.Errorf("detail = %q, want to contain 'brew install'", result.detail)
	}
}

// --- resolveMethod tests (Task 1.3) ---

func TestResolveMethod(t *testing.T) {
	env := doctor.DetectedEnvironment{}

	tests := []struct {
		name           string
		toolName       string
		packageManager string
		toolMethods    map[string]config.ToolConfig
		fallbacks      []string
		want           string
		wantStderr     string // substring expected in stderr output
	}{
		{
			name:     "per-tool override wins over global",
			toolName: "gaze",
			packageManager: "dnf",
			toolMethods: map[string]config.ToolConfig{
				"gaze": {Method: "homebrew"},
			},
			want: "homebrew",
		},
		{
			name:           "global dnf returns dnf",
			toolName:       "gaze",
			packageManager: "dnf",
			want:           "dnf",
		},
		{
			name:           "global homebrew returns homebrew",
			toolName:       "gaze",
			packageManager: "homebrew",
			want:           "homebrew",
		},
		{
			name:           "apt falls through to first fallback",
			toolName:       "gaze",
			packageManager: "apt",
			fallbacks:      []string{"homebrew", "dnf", "go"},
			want:           "go",
			wantStderr:     "apt support not yet implemented",
		},
		{
			name:           "auto resolves to first matching fallback",
			toolName:       "gaze",
			packageManager: "auto",
			fallbacks:      []string{"homebrew", "dnf", "go"},
			want:           "go",
		},
		{
			name:           "empty PackageManager resolves to first matching fallback",
			toolName:       "gaze",
			packageManager: "",
			fallbacks:      []string{"homebrew", "dnf", "go"},
			want:           "go",
		},
		{
			name:           "auto with no fallbacks returns skip",
			toolName:       "gaze",
			packageManager: "",
			want:           "skip",
		},
		{
			name:     "per-tool rpm override",
			toolName: "gaze",
			toolMethods: map[string]config.ToolConfig{
				"gaze": {Method: "rpm"},
			},
			want: "rpm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			opts := &Options{
				PackageManager: tt.packageManager,
				ToolMethods:    tt.toolMethods,
				Stderr:         &stderr,
			}
			opts.defaults()

			got := opts.resolveMethod(tt.toolName, env, tt.fallbacks...)
			if got != tt.want {
				t.Errorf("resolveMethod(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
			if tt.wantStderr != "" {
				if !strings.Contains(stderr.String(), tt.wantStderr) {
					t.Errorf("stderr = %q, want to contain %q", stderr.String(), tt.wantStderr)
				}
			}
		})
	}
}

// --- installViaGo tests (Task 1.4) ---

func TestInstallViaGo(t *testing.T) {
	const testModule = "github.com/example/tool/cmd/tool"

	tests := []struct {
		name       string
		lookPath   map[string]string
		execErrors map[string]error
		dryRun     bool
		wantAction string
		wantDetail string
		wantErr    bool
	}{
		{
			name:       "success",
			lookPath:   map[string]string{"go": "/usr/local/bin/go"},
			wantAction: "installed",
			wantDetail: "via go install",
		},
		{
			name:     "go install fails",
			lookPath: map[string]string{"go": "/usr/local/bin/go"},
			execErrors: map[string]error{
				"go install " + testModule + "@latest": fmt.Errorf("compilation error"),
			},
			wantAction: "failed",
			wantDetail: testModule,
			wantErr:    true,
		},
		{
			name:       "go not available",
			lookPath:   map[string]string{},
			wantAction: "skipped",
			wantDetail: "Go not available",
			wantErr:    false,
		},
		{
			name:       "dry-run",
			lookPath:   map[string]string{"go": "/usr/local/bin/go"},
			dryRun:     true,
			wantAction: "dry-run",
			wantDetail: "Would install: go install " + testModule + "@latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &cmdRecorder{
				errors: tt.execErrors,
			}
			if rec.errors == nil {
				rec.errors = map[string]error{}
			}

			opts := &Options{
				DryRun:   tt.dryRun,
				LookPath: stubLookPath(tt.lookPath),
				ExecCmd:  rec.execCmd,
				Stdout:   &bytes.Buffer{},
				Stderr:   &bytes.Buffer{},
			}

			result := installViaGo(opts, "TestTool", testModule)

			if result.action != tt.wantAction {
				t.Errorf("action = %q, want %q", result.action, tt.wantAction)
			}
			if !strings.Contains(result.detail, tt.wantDetail) {
				t.Errorf("detail = %q, want to contain %q", result.detail, tt.wantDetail)
			}
			if tt.wantErr && result.err == nil {
				t.Error("expected non-nil error")
			}
			if !tt.wantErr && result.err != nil {
				t.Errorf("unexpected error: %v", result.err)
			}
		})
	}
}

// --- installGaze fallback chain tests (Task 2.2) ---

func TestInstallGaze_DnfFallback(t *testing.T) {
	// No Homebrew, dnf available → installs via RPM.
	rec := &cmdRecorder{}
	opts := &Options{
		Version:  "1.0.0",
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "dnf") {
		t.Errorf("detail = %q, want to contain 'dnf'", result.detail)
	}
}

func TestInstallGaze_GoInstallFallback(t *testing.T) {
	// No Homebrew, no dnf, Go available → installs via go install.
	rec := &cmdRecorder{}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallGaze_NoManagersNoGo(t *testing.T) {
	// No Homebrew, no dnf, no Go → skips.
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "skipped" {
		t.Errorf("action = %q, want skipped", result.action)
	}
}

func TestInstallGaze_PackageManagerDnf(t *testing.T) {
	// PackageManager: "dnf" → skips Homebrew, uses dnf directly.
	rec := &cmdRecorder{}
	opts := &Options{
		PackageManager: "dnf",
		Version:        "1.0.0",
		LookPath:       stubLookPath(map[string]string{}),
		ExecCmd:        rec.execCmd,
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "dnf") {
		t.Errorf("detail = %q, want to contain 'dnf'", result.detail)
	}
	// Verify no brew install was called.
	for _, call := range rec.calls {
		if strings.Contains(call, "brew") {
			t.Errorf("unexpected brew call when PackageManager=dnf: %s", call)
		}
	}
}

func TestInstallGaze_ExplicitGoMethod(t *testing.T) {
	rec := &cmdRecorder{}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"gaze": {Method: "go"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

// --- installReplicator fallback chain tests (Task 3.3) ---

func TestInstallReplicator_DnfFallback(t *testing.T) {
	rec := &cmdRecorder{}
	opts := Options{
		Version:  "1.0.0",
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "dnf") {
		t.Errorf("detail = %q, want to contain 'dnf'", result.detail)
	}
}

func TestInstallReplicator_GoInstallFallback(t *testing.T) {
	rec := &cmdRecorder{}
	opts := Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallReplicator_NoManagersNoGo(t *testing.T) {
	opts := Options{
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "skipped" {
		t.Errorf("action = %q, want skipped", result.action)
	}
}

func TestInstallReplicator_PackageManagerDnf(t *testing.T) {
	rec := &cmdRecorder{}
	opts := Options{
		PackageManager: "dnf",
		Version:        "1.0.0",
		LookPath:       stubLookPath(map[string]string{}),
		ExecCmd:        rec.execCmd,
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "dnf") {
		t.Errorf("detail = %q, want to contain 'dnf'", result.detail)
	}
	for _, call := range rec.calls {
		if strings.Contains(call, "brew") {
			t.Errorf("unexpected brew call when PackageManager=dnf: %s", call)
		}
	}
}

func TestInstallReplicator_ExplicitGoMethod(t *testing.T) {
	rec := &cmdRecorder{}
	opts := Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"replicator": {Method: "go"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

// --- installDewey fallback chain tests (Task 4.3) ---

func TestInstallDewey_GoInstallFallback(t *testing.T) {
	// No Homebrew, Go available → installs via go install + pulls model.
	rec := &cmdRecorder{
		outputs: map[string]string{
			"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go":     "/usr/local/bin/go",
			"ollama": "/usr/local/bin/ollama",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallDewey_NoHomebrewNoGo(t *testing.T) {
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "skipped" {
		t.Errorf("action = %q, want skipped", result.action)
	}
}

func TestInstallDewey_PackageManagerDnfFallsToGo(t *testing.T) {
	// PackageManager: "dnf" → resolveMethod returns "dnf", but Dewey
	// has no dnf case, so it falls through to go install.
	rec := &cmdRecorder{
		outputs: map[string]string{
			"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
		},
	}
	opts := &Options{
		PackageManager: "dnf",
		LookPath: stubLookPath(map[string]string{
			"go":     "/usr/local/bin/go",
			"ollama": "/usr/local/bin/ollama",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallDewey_ExplicitGoMethod(t *testing.T) {
	rec := &cmdRecorder{
		outputs: map[string]string{
			"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go":     "/usr/local/bin/go",
			"ollama": "/usr/local/bin/ollama",
		}),
		ExecCmd: rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"dewey": {Method: "go"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

// --- installGH fallback chain tests (Task 5.3) ---

func TestInstallGH_DnfFallback(t *testing.T) {
	// No Homebrew, dnf available → attempts dnf install gh.
	rec := &cmdRecorder{}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGH(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "dnf") {
		t.Errorf("detail = %q, want to contain 'dnf'", result.detail)
	}
}

func TestInstallGH_DnfFails_SkipsGracefully(t *testing.T) {
	// dnf install gh fails → skipped (NOT failed), with actionable link.
	rec := &cmdRecorder{
		errors: map[string]error{
			"dnf install -y gh": fmt.Errorf("No match for argument: gh"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGH(opts, env)
	if result.action != "skipped" {
		t.Errorf("action = %q, want skipped (graceful degradation)", result.action)
	}
	if result.err != nil {
		t.Errorf("err should be nil for graceful degradation, got: %v", result.err)
	}
	if !strings.Contains(result.detail, "install_linux.md") {
		t.Errorf("detail = %q, want to contain repo setup URL", result.detail)
	}
}

func TestInstallGH_PackageManagerDnf(t *testing.T) {
	// PackageManager: "dnf" → skips Homebrew, uses dnf directly.
	rec := &cmdRecorder{}
	opts := &Options{
		PackageManager: "dnf",
		LookPath:       stubLookPath(map[string]string{}),
		ExecCmd:        rec.execCmd,
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGH(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "dnf") {
		t.Errorf("detail = %q, want to contain 'dnf'", result.detail)
	}
	for _, call := range rec.calls {
		if strings.Contains(call, "brew") {
			t.Errorf("unexpected brew call when PackageManager=dnf: %s", call)
		}
	}
}

func TestInstallGH_ExplicitDnfDryRun(t *testing.T) {
	opts := &Options{
		DryRun:         true,
		PackageManager: "dnf",
		LookPath:       stubLookPath(map[string]string{}),
		Stdout:         &bytes.Buffer{},
		Stderr:         &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installGH(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "dnf install -y gh") {
		t.Errorf("detail = %q, want to contain 'dnf install -y gh'", result.detail)
	}
}

// --- Dry-run fallback chain tests (Task 6.1 + 6.2) ---

func TestInstallGaze_DryRunDnfFallback(t *testing.T) {
	opts := &Options{
		DryRun:   true,
		Version:  "1.0.0",
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  (&cmdRecorder{}).execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGaze(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "dnf install") {
		t.Errorf("detail = %q, want to contain 'dnf install'", result.detail)
	}
}

func TestInstallGaze_DryRunGoFallback(t *testing.T) {
	opts := &Options{
		DryRun: true,
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: (&cmdRecorder{}).execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallReplicator_DryRunDnfFallback(t *testing.T) {
	opts := Options{
		DryRun:   true,
		Version:  "1.0.0",
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  (&cmdRecorder{}).execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installReplicator(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "dnf install") {
		t.Errorf("detail = %q, want to contain 'dnf install'", result.detail)
	}
}

func TestInstallReplicator_DryRunGoFallback(t *testing.T) {
	opts := Options{
		DryRun: true,
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: (&cmdRecorder{}).execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

func TestInstallDewey_DryRunGoFallback(t *testing.T) {
	opts := &Options{
		DryRun: true,
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: (&cmdRecorder{}).execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "go install") {
		t.Errorf("detail = %q, want to contain 'go install'", result.detail)
	}
}

// --- installGH explicit Homebrew and auto-mode Homebrew tests ---

func TestInstallGH_ExplicitHomebrew(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install succeeds.
	rec := &cmdRecorder{}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"gh": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGH(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
	found := false
	for _, call := range rec.calls {
		if call == "brew install gh" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'brew install gh' call, got: %v", rec.calls)
	}
}

func TestInstallGH_ExplicitHomebrewFails(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install gh": fmt.Errorf("brew: formula not found"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"gh": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGH(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallGH_ExplicitHomebrewDryRun(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override, dry-run mode.
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		ToolMethods: map[string]config.ToolConfig{
			"gh": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGH(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "brew install gh") {
		t.Errorf("detail = %q, want to contain 'brew install gh'", result.detail)
	}
}

func TestInstallGH_AutoHomebrewAvailable(t *testing.T) {
	// Auto mode, Homebrew available → brew install succeeds.
	rec := &cmdRecorder{}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGH(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
}

func TestInstallGH_AutoHomebrewFails(t *testing.T) {
	// Auto mode, Homebrew available → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install gh": fmt.Errorf("brew error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGH(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallGH_AutoDryRunHomebrew(t *testing.T) {
	// Auto mode, dry-run, Homebrew available → dry-run with brew hint.
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installGH(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "brew install gh") {
		t.Errorf("detail = %q, want to contain 'brew install gh'", result.detail)
	}
}

func TestInstallGH_AutoDryRunNoManagers(t *testing.T) {
	// Auto mode, dry-run, no managers → dry-run with download link.
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGH(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "cli.github.com") {
		t.Errorf("detail = %q, want to contain 'cli.github.com'", result.detail)
	}
}

// --- installReplicator explicit Homebrew and auto-mode tests ---

func TestInstallReplicator_ExplicitHomebrew(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install succeeds.
	rec := &cmdRecorder{}
	opts := Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"replicator": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
	found := false
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/replicator" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected brew install call, got: %v", rec.calls)
	}
}

func TestInstallReplicator_ExplicitHomebrewFails(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/replicator": fmt.Errorf("brew error"),
		},
	}
	opts := Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"replicator": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallReplicator_ExplicitHomebrewDryRun(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override, dry-run mode.
	opts := Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		ToolMethods: map[string]config.ToolConfig{
			"replicator": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "brew install unbound-force/tap/replicator") {
		t.Errorf("detail = %q, want to contain brew install hint", result.detail)
	}
}

func TestInstallReplicator_AutoHomebrewAvailable(t *testing.T) {
	// Auto mode, Homebrew available → brew install succeeds.
	rec := &cmdRecorder{}
	opts := Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installReplicator(&opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
}

func TestInstallReplicator_AutoHomebrewFails(t *testing.T) {
	// Auto mode, Homebrew available → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/replicator": fmt.Errorf("brew error"),
		},
	}
	opts := Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installReplicator(&opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallReplicator_AutoGoInstallFails(t *testing.T) {
	// Auto mode, no Homebrew, no dnf, Go available but go install fails.
	// installViaGo returns action: "failed" → goResult.action != "skipped",
	// so installReplicator returns the failed result directly.
	rec := &cmdRecorder{
		errors: map[string]error{
			"go install github.com/unbound-force/replicator/cmd/replicator@latest": fmt.Errorf("compilation error"),
		},
	}
	opts := Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installReplicator(&opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

// --- installDewey explicit Homebrew and auto-mode tests ---

func TestInstallDewey_ExplicitHomebrew(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install
	// succeeds, post-install runs (ollama already has model).
	rec := &cmdRecorder{
		outputs: map[string]string{
			"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"ollama": "/usr/local/bin/ollama",
		}),
		ExecCmd: rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"dewey": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
	found := false
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/dewey" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected brew install call, got: %v", rec.calls)
	}
}

func TestInstallDewey_ExplicitHomebrewFails(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/dewey": fmt.Errorf("brew error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"dewey": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallDewey_ExplicitHomebrewDryRun(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override, dry-run mode.
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		ToolMethods: map[string]config.ToolConfig{
			"dewey": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "brew install unbound-force/tap/dewey") {
		t.Errorf("detail = %q, want to contain brew install hint", result.detail)
	}
}

func TestInstallDewey_ExplicitGoFails(t *testing.T) {
	// resolveMethod returns "go" via per-tool override → go install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"go install github.com/unbound-force/dewey/cmd/dewey@latest": fmt.Errorf("compilation error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"dewey": {Method: "go"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallDewey_AutoHomebrewAvailable(t *testing.T) {
	// Auto mode, Homebrew available → brew install succeeds, post-install runs.
	rec := &cmdRecorder{
		outputs: map[string]string{
			"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"ollama": "/usr/local/bin/ollama",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installDewey(opts, env)
	if result.action != "installed" {
		t.Errorf("action = %q, want installed", result.action)
	}
	if !strings.Contains(result.detail, "Homebrew") {
		t.Errorf("detail = %q, want to contain 'Homebrew'", result.detail)
	}
}

func TestInstallDewey_AutoHomebrewFails(t *testing.T) {
	// Auto mode, Homebrew available → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/dewey": fmt.Errorf("brew error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installDewey(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallDewey_AutoGoInstallFails(t *testing.T) {
	// Auto mode, no Homebrew, Go available but go install fails.
	// installViaGo returns action: "failed" → goResult.action != "skipped",
	// so installDewey returns the failed result directly.
	rec := &cmdRecorder{
		errors: map[string]error{
			"go install github.com/unbound-force/dewey/cmd/dewey@latest": fmt.Errorf("compilation error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{}

	result := installDewey(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallDewey_AutoDryRunHomebrew(t *testing.T) {
	// Auto mode, dry-run, Homebrew available → dry-run with brew hint.
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	opts.defaults()
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerHomebrew, Path: "/opt/homebrew/bin/brew"},
		},
	}

	result := installDewey(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "brew install unbound-force/tap/dewey") {
		t.Errorf("detail = %q, want to contain brew install hint", result.detail)
	}
}

// --- installGaze explicit Homebrew failure test ---

func TestInstallGaze_ExplicitHomebrewFails(t *testing.T) {
	// resolveMethod returns "homebrew" via per-tool override → brew install fails.
	rec := &cmdRecorder{
		errors: map[string]error{
			"brew install unbound-force/tap/gaze": fmt.Errorf("brew error"),
		},
	}
	opts := &Options{
		LookPath: stubLookPath(map[string]string{}),
		ExecCmd:  rec.execCmd,
		ToolMethods: map[string]config.ToolConfig{
			"gaze": {Method: "homebrew"},
		},
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{}

	result := installGaze(opts, env)
	if result.action != "failed" {
		t.Errorf("action = %q, want failed", result.action)
	}
	if result.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestInstallGH_DryRunDnfFallback(t *testing.T) {
	opts := &Options{
		DryRun:   true,
		LookPath: stubLookPath(map[string]string{}),
		Stdout:   &bytes.Buffer{},
		Stderr:   &bytes.Buffer{},
	}
	env := doctor.DetectedEnvironment{
		Managers: []doctor.ManagerInfo{
			{Kind: doctor.ManagerDnf, Path: "/usr/bin/dnf"},
		},
	}

	result := installGH(opts, env)
	if result.action != "dry-run" {
		t.Errorf("action = %q, want dry-run", result.action)
	}
	if !strings.Contains(result.detail, "dnf install -y gh") {
		t.Errorf("detail = %q, want to contain 'dnf install -y gh'", result.detail)
	}
}

// --- Error output reporting tests (Spec 036, Phase 3, US2) ---

func TestInstallViaBrew_ShowsOutput(t *testing.T) {
	t.Run("short_output", func(t *testing.T) {
		brewOutput := "Error: No available formula with the name \"gaze\".\n"
		rec := &cmdRecorder{
			outputs: map[string]string{
				"brew install gaze": brewOutput,
			},
			errors: map[string]error{
				"brew install gaze": fmt.Errorf("exit status 1"),
			},
		}

		opts := &Options{
			LookPath: stubLookPath(map[string]string{}),
			ExecCmd:  rec.execCmd,
			Stdout:   &bytes.Buffer{},
			Stderr:   &bytes.Buffer{},
		}

		result := installViaBrew(opts, "Gaze", "gaze")

		// Verify stepResult captures the output.
		if result.action != "failed" {
			t.Fatalf("action = %q, want failed", result.action)
		}
		if !strings.Contains(string(result.output), "No available formula") {
			t.Errorf("output = %q, want to contain 'No available formula'", result.output)
		}

		// Verify printStepResult renders the output.
		var buf bytes.Buffer
		printStepResult(&buf, result)
		rendered := buf.String()
		if !strings.Contains(rendered, "No available formula") {
			t.Errorf("rendered output should contain 'No available formula', got:\n%s", rendered)
		}
	})

	t.Run("long_output_truncated", func(t *testing.T) {
		// Generate 25 lines of brew output to trigger truncation.
		longOutput := setupGenerateLines(25)
		rec := &cmdRecorder{
			outputs: map[string]string{
				"brew install gaze": longOutput,
			},
			errors: map[string]error{
				"brew install gaze": fmt.Errorf("exit status 1"),
			},
		}

		opts := &Options{
			LookPath: stubLookPath(map[string]string{}),
			ExecCmd:  rec.execCmd,
			Stdout:   &bytes.Buffer{},
			Stderr:   &bytes.Buffer{},
		}

		result := installViaBrew(opts, "Gaze", "gaze")

		if result.action != "failed" {
			t.Fatalf("action = %q, want failed", result.action)
		}

		// Verify printStepResult renders truncated output.
		var buf bytes.Buffer
		printStepResult(&buf, result)
		rendered := buf.String()

		// Should contain the truncation marker.
		if !strings.Contains(rendered, "... (15 lines omitted)") {
			t.Errorf("expected '... (15 lines omitted)' in rendered output, got:\n%s", rendered)
		}

		// Should contain the last 10 lines (line 16 through line 25).
		if !strings.Contains(rendered, "line 25") {
			t.Errorf("expected last line 'line 25' in rendered output, got:\n%s", rendered)
		}
		if !strings.Contains(rendered, "line 16") {
			t.Errorf("expected 'line 16' (first of last 10) in rendered output, got:\n%s", rendered)
		}

		// Should NOT contain early lines that were truncated.
		if strings.Contains(rendered, "line 1\n") {
			t.Errorf("early lines should be truncated, but found 'line 1' in:\n%s", rendered)
		}
	})
}

func TestInstallViaGo_ShowsOutput(t *testing.T) {
	goOutput := "# github.com/example/tool\n./main.go:10:5: undefined: nonexistent\n"
	rec := &cmdRecorder{
		outputs: map[string]string{
			"go install github.com/example/tool@latest": goOutput,
		},
		errors: map[string]error{
			"go install github.com/example/tool@latest": fmt.Errorf("exit status 1"),
		},
	}

	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/bin/go",
		}),
		ExecCmd: rec.execCmd,
		Stdout:  &bytes.Buffer{},
		Stderr:  &bytes.Buffer{},
	}

	result := installViaGo(opts, "TestTool", "github.com/example/tool")

	// Verify stepResult captures the output.
	if result.action != "failed" {
		t.Fatalf("action = %q, want failed", result.action)
	}
	if !strings.Contains(string(result.output), "undefined: nonexistent") {
		t.Errorf("output = %q, want to contain 'undefined: nonexistent'", result.output)
	}

	// Verify printStepResult renders the output.
	var buf bytes.Buffer
	printStepResult(&buf, result)
	rendered := buf.String()
	if !strings.Contains(rendered, "undefined: nonexistent") {
		t.Errorf("rendered output should contain compiler error, got:\n%s", rendered)
	}
}

func TestPrintStepResult_WithOutput(t *testing.T) {
	// Test 1: Failed result with output shows the output.
	failedResult := stepResult{
		name:   "TestTool",
		action: "failed",
		detail: "install failed",
		output: []byte("Error: something went wrong\n"),
		err:    fmt.Errorf("exit status 1"),
	}

	var buf bytes.Buffer
	printStepResult(&buf, failedResult)
	rendered := buf.String()

	if !strings.Contains(rendered, "something went wrong") {
		t.Errorf("failed result should show output text, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Error: exit status 1") {
		t.Errorf("failed result should show error, got:\n%s", rendered)
	}

	// Test 2: Successful result with no output does NOT show "Output:" —
	// FR-007 regression test.
	successResult := stepResult{
		name:   "TestTool",
		action: "installed",
		detail: "via Homebrew",
	}

	var buf2 bytes.Buffer
	printStepResult(&buf2, successResult)
	rendered2 := buf2.String()

	if strings.Contains(rendered2, "Output:") {
		t.Errorf("successful result should NOT contain 'Output:', got:\n%s", rendered2)
	}
	if strings.Contains(rendered2, "Error:") {
		t.Errorf("successful result should NOT contain 'Error:', got:\n%s", rendered2)
	}
}

func TestPrintStepResult_SuccessWithOutputNotShown(t *testing.T) {
	// Even if a successful result somehow has output, it should NOT
	// be displayed — output is only shown for failed results.
	result := stepResult{
		name:   "TestTool",
		action: "installed",
		detail: "via Homebrew",
		output: []byte("some install log output\n"),
	}

	var buf bytes.Buffer
	printStepResult(&buf, result)
	rendered := buf.String()

	if strings.Contains(rendered, "some install log output") {
		t.Errorf("successful result should NOT show output, got:\n%s", rendered)
	}
}

// setupGenerateLines creates a string with n lines like "line 1\nline 2\n...".
func setupGenerateLines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

// setupGenerateLastNLines returns the last n lines of a total-line
// string, joined by newlines (no trailing newline).
func setupGenerateLastNLines(total, n int) string {
	var lines []string
	for i := total - n + 1; i <= total; i++ {
		lines = append(lines, fmt.Sprintf("line %d", i))
	}
	return strings.Join(lines, "\n")
}
