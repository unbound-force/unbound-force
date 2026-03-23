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

	// Verify install order: opencode (brew), gaze (brew), node version check,
	// swarm (npm), swarm setup, opencode.json, swarm init.
	expectedCmds := []string{
		"brew install anomalyco/tap/opencode",
		"brew install unbound-force/tap/gaze",
		"node --version",
		"npm install -g opencode-swarm-plugin@latest",
		"swarm setup",
		"swarm init",
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
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"node":     "/usr/local/bin/node",
			"npm":      "/usr/local/bin/npm",
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

	// Read the modified opencode.json.
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

	// Verify plugin array has opencode-swarm-plugin.
	var plugins []string
	if pluginRaw, ok := parsed["plugin"]; ok {
		if pErr := json.Unmarshal(pluginRaw, &plugins); pErr != nil {
			t.Fatalf("unmarshal plugin: %v", pErr)
		}
	} else {
		t.Fatal("plugin key not found")
	}

	found := false
	for _, p := range plugins {
		if p == "opencode-swarm-plugin" {
			found = true
		}
	}
	if !found {
		t.Error("opencode-swarm-plugin not in plugin array")
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

	// Verify opencode.json was created.
	data, readErr := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if readErr != nil {
		t.Fatalf("opencode.json not created: %v", readErr)
	}

	// Verify it has $schema and plugin array.
	var parsed map[string]json.RawMessage
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("invalid JSON: %v", jsonErr)
	}

	if _, ok := parsed["$schema"]; !ok {
		t.Error("$schema not found in created opencode.json")
	}

	var plugins []string
	if pluginRaw, ok := parsed["plugin"]; ok {
		if pErr := json.Unmarshal(pluginRaw, &plugins); pErr != nil {
			t.Fatalf("unmarshal plugin: %v", pErr)
		}
	}
	found := false
	for _, p := range plugins {
		if p == "opencode-swarm-plugin" {
			found = true
		}
	}
	if !found {
		t.Error("opencode-swarm-plugin not in plugin array")
	}
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
	// Malformed JSON causes a "skipped" step (not "failed"), so
	// Run may or may not return an error depending on other steps.
	// The key assertion is that the output warns about malformed JSON.
	_ = err

	output := buf.String()
	if !strings.Contains(output, "malformed") && !strings.Contains(output, "skipped") {
		t.Error("expected malformed JSON warning in output")
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

func TestSetupRun_OllamaTip(t *testing.T) {
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
			// ollama NOT in PATH
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
	if !strings.Contains(output, "Tip") {
		t.Error("expected Ollama tip when ollama not installed")
	}
	if !strings.Contains(output, "ollama") {
		t.Error("expected 'ollama' in tip message")
	}
	if !strings.Contains(output, "granite-embedding:30m") {
		t.Error("expected 'granite-embedding:30m' in tip message")
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
