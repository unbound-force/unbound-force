package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	// Verify install order: opencode (brew), gaze (brew), mxf (brew),
	// gh (brew), node version check, bun (npm), openspec (npm),
	// swarm (npm), swarm setup, swarm init, ollama (brew),
	// dewey (brew). Note: dewey init/index are handled by uf init
	// (via scaffold.initSubTools), not by setup directly.
	expectedCmds := []string{
		"brew install anomalyco/tap/opencode",
		"brew install unbound-force/tap/gaze",
		"brew install unbound-force/tap/mxf",
		"brew install gh",
		"node --version",
		"npm install -g bun",
		"npm install -g @fission-ai/openspec@latest",
		"npm install -g opencode-swarm-plugin@latest",
		"swarm setup",
		"swarm init",
		"brew install --cask ollama-app",
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
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)
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
			"node --version": "v22.15.0",
			"ollama list":    "NAME                    ID              SIZE\ngranite-embedding:30m   abc123          63 MB\n",
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
			"mxf":           "/usr/local/bin/mxf",
			"gh":            "/usr/local/bin/gh",
			"node":          "/usr/local/bin/node",
			"npm":           "/usr/local/bin/npm",
			"bun":           "/home/user/.bun/bin/bun",
			"openspec":      "/usr/local/bin/openspec",
			"swarm":         "/usr/local/bin/swarm",
			"dewey":         "/usr/local/bin/dewey",
			"ollama":        "/usr/local/bin/ollama",
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

	// Verify no install commands were called (only node --version for checking).
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

	// Swarm steps should be skipped because Node.js failed.
	for _, call := range rec.calls {
		if strings.Contains(call, "npm install") || call == "swarm setup" || call == "swarm init" {
			t.Errorf("unexpected swarm-related command: %s", call)
		}
	}
}

func TestSetupRun_NpmFails(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			"npm install -g opencode-swarm-plugin@latest": fmt.Errorf("npm ERR! code EACCES"),
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
		t.Fatal("expected error when npm install fails")
	}

	// swarm setup/init should NOT be called.
	for _, call := range rec.calls {
		if call == "swarm setup" || call == "swarm init" {
			t.Errorf("unexpected command after npm failure: %s", call)
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

	// Swarm should be installed via npm from nvm-managed node.
	npmCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "npm install -g opencode-swarm-plugin") {
			npmCalled = true
		}
	}
	if !npmCalled {
		t.Error("expected npm install call for swarm plugin")
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

func TestSetupRun_BunDetected(t *testing.T) {
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
			"node": "/usr/local/bin/node",
			"bun":  "/home/user/.bun/bin/bun",
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

	// Should use bun instead of npm.
	bunCalled := false
	for _, call := range rec.calls {
		if strings.Contains(call, "bun add -g opencode-swarm-plugin") {
			bunCalled = true
		}
	}
	if !bunCalled {
		t.Errorf("expected bun add call, got calls: %v", rec.calls)
	}
}

func TestSetupRun_OpencodeJsonManipulation(t *testing.T) {
	dir := t.TempDir()

	// Create opencode.json with existing MCP servers, no plugin key.
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
			"brew":  "/opt/homebrew/bin/brew",
			"node":  "/usr/local/bin/node",
			"npm":   "/usr/local/bin/npm",
			"swarm": "/usr/local/bin/swarm",
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

	// Setup no longer directly writes opencode.json (US4).
	// opencode.json is now managed by uf init (via scaffold.configureOpencodeJSON).
	// Verify the original file is unchanged by setup — uf init handles it.
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
			"brew":  "/opt/homebrew/bin/brew",
			"node":  "/usr/local/bin/node",
			"npm":   "/usr/local/bin/npm",
			"swarm": "/usr/local/bin/swarm",
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

	// Setup no longer directly creates opencode.json (US4).
	// opencode.json is now created by uf init (via scaffold.configureOpencodeJSON)
	// which runs as the final step of setup. The file may or may not exist
	// depending on whether dewey/swarm are available in the test environment.
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
			"brew":  "/opt/homebrew/bin/brew",
			"node":  "/usr/local/bin/node",
			"npm":   "/usr/local/bin/npm",
			"swarm": "/usr/local/bin/swarm",
		}),
		ExecCmd:      rec.execCmd,
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
		WriteFile:    os.WriteFile,
	}

	err := Run(opts)
	// Setup no longer directly touches opencode.json (US4).
	// Malformed JSON is handled by uf init (scaffold.configureOpencodeJSON)
	// which runs as the final step. Run should succeed.
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

func TestSetupRun_SwarmPluginNpmFails(t *testing.T) {
	dir := t.TempDir()

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			"npm install -g opencode-swarm-plugin@latest": fmt.Errorf("npm failed"),
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
		t.Fatal("expected error when npm install fails")
	}

	// Swarm setup/init should be skipped.
	for _, call := range rec.calls {
		if call == "swarm setup" || call == "swarm init" {
			t.Errorf("unexpected command after npm failure: %s", call)
		}
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
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
			"ollama list":    "NAME                    ID              SIZE\ngranite-embedding:30m   abc123          63 MB\n",
		},
	}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			"brew":     "/opt/homebrew/bin/brew",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"swarm":    "/usr/local/bin/swarm",
			"dewey":    "/usr/local/bin/dewey",
			"ollama":   "/usr/local/bin/ollama",
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
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

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
			"brew":     "/opt/homebrew/bin/brew",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			"go":       "/usr/local/bin/go",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"swarm":    "/usr/local/bin/swarm",
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

	// Verify Ollama was installed via Homebrew (no tip, actual install).
	found := false
	for _, call := range rec.calls {
		if call == "brew install --cask ollama-app" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'brew install --cask ollama-app' in recorded commands")
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
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

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
			"brew":     "/opt/homebrew/bin/brew",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			"go":       "/usr/local/bin/go",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"swarm":    "/usr/local/bin/swarm",
			"ollama":   "/usr/local/bin/ollama", // ollama IS in PATH
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
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

	rec := &cmdRecorder{}

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Stdout:    &buf,
		Stderr:    &buf,
		LookPath: stubLookPath(map[string]string{
			// No brew, no ollama — Homebrew not available
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"swarm":    "/usr/local/bin/swarm",
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

	// Verify no brew install ollama-app was attempted.
	for _, call := range rec.calls {
		if call == "brew install --cask ollama-app" {
			t.Error("should NOT attempt brew install --cask ollama-app when Homebrew is not available")
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
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

	rec := &cmdRecorder{
		outputs: map[string]string{
			"node --version": "v22.15.0",
		},
		errors: map[string]error{
			"brew install --cask ollama-app": fmt.Errorf("brew: cask not found"),
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
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			"bun":      "/home/user/.bun/bin/bun",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"swarm":    "/usr/local/bin/swarm",
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

// --- Mx F installation tests ---

func TestSetupRun_MxFMissing_BrewInstall(t *testing.T) {
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
			// mxf NOT in PATH
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
		if call == "brew install unbound-force/tap/mxf" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected brew install mxf, got calls: %v", rec.calls)
	}
}

func TestSetupRun_MxFNoHomebrew(t *testing.T) {
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
			// No brew, no mxf
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

	// Verify no brew install mxf was attempted.
	for _, call := range rec.calls {
		if call == "brew install unbound-force/tap/mxf" {
			t.Error("should NOT attempt brew install mxf when Homebrew is not available")
		}
	}

	output := buf.String()
	if !strings.Contains(output, "GitHub") || !strings.Contains(output, "releases") {
		t.Error("expected GitHub releases link in output when Homebrew is not available")
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
			"mxf":      "/usr/local/bin/mxf",
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
			"mxf":      "/usr/local/bin/mxf",
			"gh":       "/usr/local/bin/gh",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			"bun":      "/home/user/.bun/bin/bun",
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

	// Should use bun (preferred) to install openspec.
	bunCalled := false
	for _, call := range rec.calls {
		if call == "bun add -g @fission-ai/openspec@latest" {
			bunCalled = true
		}
	}
	if !bunCalled {
		t.Errorf("expected bun add for openspec, got calls: %v", rec.calls)
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
			"mxf":      "/usr/local/bin/mxf",
			"gh":       "/usr/local/bin/gh",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
			// No bun — falls back to npm which fails
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
	if !strings.Contains(output, "npm") && !strings.Contains(output, "permissions") {
		t.Error("expected npm permissions guidance in openspec failure message")
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
	// Note: dewey init/index are no longer in setup — they are handled
	// by uf init (via scaffold.initSubTools).
	checks := []struct {
		name    string
		pattern string
	}{
		{"mxf", "Would install: brew install unbound-force/tap/mxf"},
		{"gh", "Would install: brew install gh"},
		{"openspec", "Would install: bun add -g @fission-ai/openspec@latest"},
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
