package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseRuntimeVersion_AcceptsAnchoredASCIIVersions(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		allowVPrefix bool
		want         runtimeVersion
	}{
		{name: "node v prefix", output: "v20.0.0", allowVPrefix: true, want: runtimeVersion{20, 0, 0}},
		{name: "node without v prefix", output: "24.1.2\n", allowVPrefix: true, want: runtimeVersion{24, 1, 2}},
		{name: "npm CRLF and ASCII edge trim", output: "\t 11.6.2 \t\r\n", want: runtimeVersion{11, 6, 2}},
		{name: "uint32 boundary", output: "24.4294967295.0", allowVPrefix: true, want: runtimeVersion{24, 4294967295, 0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseRuntimeVersion([]byte(test.output), test.allowVPrefix)
			if err != nil {
				t.Fatalf("parseRuntimeVersion() error: %v", err)
			}
			if got != test.want {
				t.Errorf("parseRuntimeVersion() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseRuntimeVersion_RejectsMalformedAndOverflowValues(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		allowVPrefix bool
	}{
		{name: "explicit npm v prefix", output: "v10.11.0"},
		{name: "major leading zero", output: "020.1.1", allowVPrefix: true},
		{name: "minor leading zero", output: "20.01.1", allowVPrefix: true},
		{name: "patch leading zero", output: "20.1.01", allowVPrefix: true},
		{name: "major overflow", output: "4294967296.1.1", allowVPrefix: true},
		{name: "minor overflow", output: "20.4294967296.1", allowVPrefix: true},
		{name: "patch overflow", output: "20.1.4294967296", allowVPrefix: true},
		{name: "malformed major", output: "x20.1.1", allowVPrefix: true},
		{name: "malformed minor", output: "20.x1.1", allowVPrefix: true},
		{name: "malformed patch", output: "20.1.x1", allowVPrefix: true},
		{name: "trailing garbage", output: "20.1.1garbage", allowVPrefix: true},
		{name: "negative component", output: "-20.1.1", allowVPrefix: true},
		{name: "float component", output: "20.1.1.0", allowVPrefix: true},
		{name: "empty component", output: "20..1", allowVPrefix: true},
		{name: "whitespace component", output: "20. 1.1", allowVPrefix: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseRuntimeVersion([]byte(test.output), test.allowVPrefix)
			if err == nil {
				t.Error("parseRuntimeVersion() want error, got nil")
			}
		})
	}
}

func TestRun_InstallsProbesAndAtomicallyActivatesReviewPlugins(t *testing.T) {
	targetDir := t.TempDir()
	opencodeDir := filepath.Join(targetDir, ".opencode")
	configPath := filepath.Join(targetDir, "opencode.json")
	config := []byte("{\n  \"plugin\": [\"user-owned-plugin\"]\n}\n")
	if err := os.WriteFile(configPath, config, 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	var events []string
	var stageRoot string
	renameCount := 0
	pluginDestinations := []string{
		filepath.Join(opencodeDir, "plugins", "invoke-agent"),
		filepath.Join(opencodeDir, "plugins", "review-dispatch"),
	}
	assertPluginsAbsent := func(phase string) {
		t.Helper()
		for _, destination := range pluginDestinations {
			if _, err := os.Stat(filepath.Join(destination, "index.ts")); !os.IsNotExist(err) {
				t.Fatalf("%s: plugin source became auto-discoverable at %s: %v", phase, destination, err)
			}
		}
	}

	opts := Options{
		TargetDir: targetDir,
		Version:   "1.0.0-test",
		Stdout:    &bytes.Buffer{},
		LookPath: func(name string) (string, error) {
			events = append(events, "lookpath "+name)
			switch name {
			case "node", "npm", "opencode":
				return filepath.Join("/tools", name), nil
			default:
				return "", fmt.Errorf("executable %q not found", name)
			}
		},
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected cwd-dependent command %s %v", name, args)
			return nil, nil
		},
		ExecCmdInDir: func(dir, name string, args ...string) ([]byte, error) {
			if dir != opencodeDir {
				t.Fatalf("command directory = %q, want %q", dir, opencodeDir)
			}
			assertPluginsAbsent("before " + filepath.Base(name))
			command := filepath.Base(name) + " " + strings.Join(args, " ")
			switch command {
			case "node --version":
				events = append(events, command)
				return []byte("v22.15.0\r\n"), nil
			case "npm --version":
				events = append(events, command)
				return []byte("10.9.2\n"), nil
			case "npm ci --ignore-scripts --omit=dev":
				events = append(events, command)
				return nil, nil
			case "opencode debug config":
				if stageRoot == "" {
					t.Fatal("probe ran before external staging")
				}
				if strings.HasPrefix(stageRoot, filepath.Join(opencodeDir, "plugins")+string(filepath.Separator)) {
					t.Fatalf("stage %q is inside auto-discovery", stageRoot)
				}
				for _, stagedPath := range []string{
					filepath.Join(stageRoot, "plugins", "invoke-agent", "index.ts"),
					filepath.Join(stageRoot, "plugins", "review-dispatch", "index.ts"),
					filepath.Join(stageRoot, "lib", "review-dispatch-lesson-proposal.ts"),
					filepath.Join(stageRoot, "lib", "review-dispatch-sibling-evidence.ts"),
					filepath.Join(stageRoot, "lib", "reviewer-manifest.ts"),
				} {
					if _, err := os.Stat(stagedPath); err != nil {
						t.Fatalf("probe ran before staged input %s existed: %v", stagedPath, err)
					}
				}
				probe, err := os.ReadFile(filepath.Join(opencodeDir, "plugins", pluginProbeDirectoryName, "index.ts"))
				if err != nil {
					t.Fatalf("read generated probe: %v", err)
				}
				if strings.Contains(string(probe), "plan_review_dispatch") {
					events = append(events, "tool-definition probe")
				} else {
					events = append(events, "import probe")
				}
				return []byte("{}\n"), nil
			default:
				t.Fatalf("unexpected command %q", command)
				return nil, nil
			}
		},
		MkdirTemp: func(dir, pattern string) (string, error) {
			if dir != opencodeDir || pattern != ".uf-review-plugins-" {
				t.Fatalf("MkdirTemp(%q, %q), want target .opencode external stage", dir, pattern)
			}
			var err error
			stageRoot, err = os.MkdirTemp(dir, pattern)
			return stageRoot, err
		},
		Rename: func(source, destination string) error {
			if renameCount == 0 {
				assertPluginsAbsent("before atomic activation")
			} else if _, err := os.Stat(filepath.Join(destination, "index.ts")); !os.IsNotExist(err) {
				t.Fatalf("activation destination %q already contains source: %v", destination, err)
			}
			if stageRoot == "" || !strings.HasPrefix(source, stageRoot+string(filepath.Separator)) {
				t.Fatalf("activation source %q is outside stage %q", source, stageRoot)
			}
			if strings.HasPrefix(stageRoot, filepath.Join(opencodeDir, "plugins")+string(filepath.Separator)) {
				t.Fatalf("stage %q is inside auto-discovery", stageRoot)
			}
			if _, err := os.Stat(filepath.Join(source, "index.ts")); err != nil {
				t.Fatalf("staged source is incomplete: %v", err)
			}
			events = append(events, "activate "+filepath.Base(destination))
			renameCount++
			return os.Rename(source, destination)
		},
	}

	result, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	wantActivationEvents := []string{
		"lookpath node",
		"node --version",
		"lookpath npm",
		"npm --version",
		"npm ci --ignore-scripts --omit=dev",
		"lookpath opencode",
		"import probe",
		"tool-definition probe",
		"activate invoke-agent",
		"activate review-dispatch",
	}
	if len(events) < len(wantActivationEvents) || !reflect.DeepEqual(events[:len(wantActivationEvents)], wantActivationEvents) {
		t.Errorf("activation operation order =\n%v\nwant prefix\n%v", events, wantActivationEvents)
	}

	for _, assetPath := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset} {
		targetPath := filepath.Join(targetDir, mapAssetPath(assetPath))
		got, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatalf("read activated plugin %s: %v", assetPath, err)
		}
		want, err := assetContent(assetPath)
		if err != nil {
			t.Fatalf("read embedded plugin %s: %v", assetPath, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("activated plugin %s differs from embedded source", assetPath)
		}
		if !containsPath(result.Created, mapAssetPath(assetPath)) {
			t.Errorf("activated plugin %s missing from result.Created", assetPath)
		}
	}
	gotConfig, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	var gotOC map[string]json.RawMessage
	if err := json.Unmarshal(gotConfig, &gotOC); err != nil {
		t.Fatalf("parse opencode.json: %v", err)
	}
	gotPlugins := getPlugins(t, gotOC)
	wantPlugins := []string{
		"user-owned-plugin",
		"./" + mapAssetPath(invokeAgentPluginAsset),
		"./" + mapAssetPath(reviewDispatchPluginAsset),
	}
	if len(gotPlugins) != len(wantPlugins) {
		t.Fatalf("plugin array = %v, want %v", gotPlugins, wantPlugins)
	}
	for i := range wantPlugins {
		if gotPlugins[i] != wantPlugins[i] {
			t.Errorf("plugin[%d] = %q, want %q", i, gotPlugins[i], wantPlugins[i])
		}
	}
}

// reviewPluginRunOpts builds an Options whose injected command fakes drive
// the review plugin activation flow. When failInstall is true, `npm ci`
// returns an error (soft failure); otherwise the full install → probe →
// atomic activate path succeeds.
func reviewPluginRunOpts(t *testing.T, targetDir string, stdout io.Writer, failInstall bool) Options {
	t.Helper()
	opencodeDir := filepath.Join(targetDir, ".opencode")
	return Options{
		TargetDir: targetDir,
		Version:   "1.0.0-test",
		Stdout:    stdout,
		LookPath: func(name string) (string, error) {
			switch name {
			case "node", "npm", "opencode":
				return filepath.Join("/tools", name), nil
			default:
				return "", fmt.Errorf("executable %q not found", name)
			}
		},
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected cwd-dependent command %s %v", name, args)
			return nil, nil
		},
		ExecCmdInDir: func(dir, name string, args ...string) ([]byte, error) {
			if dir != opencodeDir {
				t.Fatalf("command directory = %q, want %q", dir, opencodeDir)
			}
			command := filepath.Base(name) + " " + strings.Join(args, " ")
			switch command {
			case "node --version":
				return []byte("v22.15.0\r\n"), nil
			case "npm --version":
				return []byte("10.9.2\n"), nil
			case "npm ci --ignore-scripts --omit=dev":
				if failInstall {
					return []byte("npm ERR! code ENOTCACHED\n"), fmt.Errorf("npm ci failed: offline")
				}
				return nil, nil
			case "opencode debug config":
				return []byte("{}\n"), nil
			default:
				t.Fatalf("unexpected command %q", command)
				return nil, nil
			}
		},
		MkdirTemp: func(dir, pattern string) (string, error) {
			return os.MkdirTemp(dir, pattern)
		},
		Rename: func(source, destination string) error {
			return os.Rename(source, destination)
		},
	}
}

func TestRun_PartialResultOnPluginActivationFailure(t *testing.T) {
	targetDir := t.TempDir()
	opencodeDir := filepath.Join(targetDir, ".opencode")
	var stdout bytes.Buffer

	result, err := Run(reviewPluginRunOpts(t, targetDir, &stdout, true))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil (soft failure exits zero)", err)
	}
	if result.Status != resultStatusPartial {
		t.Errorf("Run() Status = %q, want %q", result.Status, resultStatusPartial)
	}
	if result.FailedSubTools != 1 {
		t.Errorf("Run() FailedSubTools = %d, want 1", result.FailedSubTools)
	}
	if len(result.Created) == 0 {
		t.Error("Run() Created is empty, want repairable scaffold assets retained")
	}

	// Plugin sources must not be left in auto-discovery on failure.
	for _, destination := range []string{
		filepath.Join(opencodeDir, "plugins", "invoke-agent", "index.ts"),
		filepath.Join(opencodeDir, "plugins", "review-dispatch", "index.ts"),
	} {
		if _, statErr := os.Stat(destination); !os.IsNotExist(statErr) {
			t.Errorf("plugin source became auto-discoverable at %s: %v", destination, statErr)
		}
	}

	// The failed and inactive sub-tool plus remediation must be surfaced.
	out := stdout.String()
	for _, want := range []string{"review-plugins", "failed", "activation: inactive", "npm ci --ignore-scripts --omit=dev"} {
		if !strings.Contains(out, want) {
			t.Errorf("Run() output does not surface %q:\n%s", want, out)
		}
	}
}

func TestRun_IndependentSubtoolsContinueOnPluginFailure(t *testing.T) {
	targetDir := t.TempDir()
	var stdout bytes.Buffer

	result, err := Run(reviewPluginRunOpts(t, targetDir, &stdout, true))
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	// The .gitignore sub-tool runs independently of plugin activation.
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	data, readErr := os.ReadFile(gitignorePath)
	if readErr != nil {
		t.Fatalf("read .gitignore: %v", readErr)
	}
	if !strings.Contains(string(data), gitignoreMarker) {
		t.Error(".gitignore missing UF ignore block; independent sub-tool did not run")
	}

	// Other scaffold assets (non-plugin) were still deployed.
	if !containsPath(result.Created, ".opencode/agents/cobalt-crush-dev.md") {
		t.Error("independent scaffold assets missing from result.Created despite plugin failure")
	}
}

func TestRun_IdempotentRetryAfterPluginFailure(t *testing.T) {
	targetDir := t.TempDir()

	// First run: npm ci fails → soft partial result, no plugin source.
	firstOut := &bytes.Buffer{}
	first, err := Run(reviewPluginRunOpts(t, targetDir, firstOut, true))
	if err != nil {
		t.Fatalf("first Run() error = %v, want nil", err)
	}
	if first.Status != resultStatusPartial || first.FailedSubTools != 1 {
		t.Fatalf("first Run() = Status %q / FailedSubTools %d, want partial/1", first.Status, first.FailedSubTools)
	}

	// Second run: npm ci succeeds → plugins activate, result fully populated.
	secondOut := &bytes.Buffer{}
	second, err := Run(reviewPluginRunOpts(t, targetDir, secondOut, false))
	if err != nil {
		t.Fatalf("second Run() error = %v, want nil (idempotent retry)", err)
	}
	if second.Status != "" {
		t.Errorf("second Run() Status = %q, want empty (success)", second.Status)
	}
	if second.FailedSubTools != 0 {
		t.Errorf("second Run() FailedSubTools = %d, want 0", second.FailedSubTools)
	}
	for _, assetPath := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset} {
		targetPath := filepath.Join(targetDir, mapAssetPath(assetPath))
		if _, statErr := os.Stat(targetPath); statErr != nil {
			t.Errorf("retry did not activate plugin source %s: %v", assetPath, statErr)
		}
		if !containsPath(second.Created, mapAssetPath(assetPath)) {
			t.Errorf("retry did not record activated plugin %s in result.Created", assetPath)
		}
	}
}

func TestAtomicallyActivateReviewPlugins_Branches(t *testing.T) {
	// Test the three branches of the idempotency guard by actually calling atomicallyActivateReviewPlugins
	// rather than testing the logic directly
	
	tempDir := t.TempDir()
	pluginsDir := filepath.Join(tempDir, ".opencode", "plugins")
	
	// Test case 1: Force refresh
	t.Run("Force refresh", func(t *testing.T) {
		// Create staged plugin directories and files
		stagedBase := filepath.Join(tempDir, "stage1")
		stagedInvokeAgentDir := filepath.Join(stagedBase, "plugins", "invoke-agent")
		stagedReviewDispatchDir := filepath.Join(stagedBase, "plugins", "review-dispatch")
		
		if err := os.MkdirAll(stagedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir staged invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(stagedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir staged review-dispatch plugin dir: %v", err)
		}
		
		if err := os.WriteFile(filepath.Join(stagedInvokeAgentDir, "index.ts"), []byte("// invoke-agent content"), 0o644); err != nil {
			t.Fatalf("write staged invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedReviewDispatchDir, "index.ts"), []byte("// review-dispatch content"), 0o644); err != nil {
			t.Fatalf("write staged review-dispatch index.ts: %v", err)
		}
		
		// Debug: Check if the source directories exist
		if _, err := os.Stat(stagedInvokeAgentDir); err != nil {
			t.Fatalf("staged invoke-agent directory does not exist: %v", err)
		}
		if _, err := os.Stat(stagedReviewDispatchDir); err != nil {
			t.Fatalf("staged review-dispatch directory does not exist: %v", err)
		}
		if _, err := os.Stat(filepath.Join(stagedInvokeAgentDir, "index.ts")); err != nil {
			t.Fatalf("staged invoke-agent index.ts does not exist: %v", err)
		}
		if _, err := os.Stat(filepath.Join(stagedReviewDispatchDir, "index.ts")); err != nil {
			t.Fatalf("staged review-dispatch index.ts does not exist: %v", err)
		}
		
		// Print the actual paths for debugging
		t.Logf("stageRoot: %s", stagedBase)
		t.Logf("stagedInvokeAgentDir: %s", stagedInvokeAgentDir)
		t.Logf("stagedReviewDispatchDir: %s", stagedReviewDispatchDir)
		
		// Create installed plugin directories and files with different content
		installedInvokeAgentDir := filepath.Join(pluginsDir, "invoke-agent")
		installedReviewDispatchDir := filepath.Join(pluginsDir, "review-dispatch")
		
		if err := os.MkdirAll(installedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir installed invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir installed review-dispatch plugin dir: %v", err)
		}
		
		if err := os.WriteFile(filepath.Join(installedInvokeAgentDir, "index.ts"), []byte("// old invoke-agent content"), 0o644); err != nil {
			t.Fatalf("write installed invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedReviewDispatchDir, "index.ts"), []byte("// old review-dispatch content"), 0o644); err != nil {
			t.Fatalf("write installed review-dispatch index.ts: %v", err)
		}
		
		// Create options with Force=true
		opts := &Options{
			TargetDir: tempDir,
			Force:     true,
			Stat:      os.Stat,
			Lstat:     os.Lstat,
			ReadFile:  os.ReadFile,
			RemoveAll: os.RemoveAll,
			Rename:    os.Rename,
			MkdirAll:  os.MkdirAll,
		}
		
		// Ensure the plugins directory exists before calling the function
		if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
			t.Fatalf("mkdir plugins directory: %v", err)
		}
		
		// Call the actual function
		activated, err := atomicallyActivateReviewPlugins(opts, stagedBase)
		if err != nil {
			t.Fatalf("atomicallyActivateReviewPlugins failed: %v", err)
		}
		
		// Verify that plugins were activated
		if len(activated) != 2 {
			t.Errorf("expected 2 activated plugins, got %d", len(activated))
		}
		
		// Verify that staged directories were removed
		if _, err := os.Stat(stagedInvokeAgentDir); !os.IsNotExist(err) {
			t.Error("staged invoke-agent directory should have been removed")
		}
		if _, err := os.Stat(stagedReviewDispatchDir); !os.IsNotExist(err) {
			t.Error("staged review-dispatch directory should have been removed")
		}
		
		// Verify that installed files now have the new content
		invokeContent, err := os.ReadFile(filepath.Join(installedInvokeAgentDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed invoke-agent content: %v", err)
		}
		if string(invokeContent) != "// invoke-agent content" {
			t.Error("installed invoke-agent content was not updated")
		}
		
		reviewContent, err := os.ReadFile(filepath.Join(installedReviewDispatchDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed review-dispatch content: %v", err)
		}
		if string(reviewContent) != "// review-dispatch content" {
			t.Error("installed review-dispatch content was not updated")
		}
	})
	
	// Test case 2: Content diff refresh
	t.Run("Content diff refresh", func(t *testing.T) {
		// Recreate staged plugin directories and files
		stagedBase := filepath.Join(tempDir, "stage2")
		stagedInvokeAgentDir := filepath.Join(stagedBase, "plugins", "invoke-agent")
		stagedReviewDispatchDir := filepath.Join(stagedBase, "plugins", "review-dispatch")
		
		if err := os.MkdirAll(stagedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir staged invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(stagedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir staged review-dispatch plugin dir: %v", err)
		}
		
		if err := os.WriteFile(filepath.Join(stagedInvokeAgentDir, "index.ts"), []byte("// new invoke-agent content"), 0o644); err != nil {
			t.Fatalf("write staged invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedReviewDispatchDir, "index.ts"), []byte("// new review-dispatch content"), 0o644); err != nil {
			t.Fatalf("write staged review-dispatch index.ts: %v", err)
		}
		
		// Create installed plugin directories and files with different content
		installedInvokeAgentDir := filepath.Join(pluginsDir, "invoke-agent")
		installedReviewDispatchDir := filepath.Join(pluginsDir, "review-dispatch")
		
		if err := os.MkdirAll(installedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir installed invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir installed review-dispatch plugin dir: %v", err)
		}
		
		if err := os.WriteFile(filepath.Join(installedInvokeAgentDir, "index.ts"), []byte("// old invoke-agent content"), 0o644); err != nil {
			t.Fatalf("write installed invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedReviewDispatchDir, "index.ts"), []byte("// old review-dispatch content"), 0o644); err != nil {
			t.Fatalf("write installed review-dispatch index.ts: %v", err)
		}
		
		// Create options with Force=false
		opts := &Options{
			TargetDir: tempDir,
			Force:     false,
			Stat:      os.Stat,
			Lstat:     os.Lstat,
			ReadFile:  os.ReadFile,
			RemoveAll: os.RemoveAll,
			Rename:    os.Rename,
			MkdirAll:  os.MkdirAll,
		}
		
		// Call the actual function
		activated, err := atomicallyActivateReviewPlugins(opts, stagedBase)
		if err != nil {
			t.Fatalf("atomicallyActivateReviewPlugins failed: %v", err)
		}
		
		// Verify that plugins were activated
		if len(activated) != 2 {
			t.Errorf("expected 2 activated plugins, got %d", len(activated))
		}
		
		// Verify that staged directories were removed
		if _, err := os.Stat(stagedInvokeAgentDir); !os.IsNotExist(err) {
			t.Error("staged invoke-agent directory should have been removed")
		}
		if _, err := os.Stat(stagedReviewDispatchDir); !os.IsNotExist(err) {
			t.Error("staged review-dispatch directory should have been removed")
		}
		
		// Verify that installed files now have the new content
		invokeContent, err := os.ReadFile(filepath.Join(installedInvokeAgentDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed invoke-agent content: %v", err)
		}
		if string(invokeContent) != "// new invoke-agent content" {
			t.Error("installed invoke-agent content was not updated")
		}
		
		reviewContent, err := os.ReadFile(filepath.Join(installedReviewDispatchDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed review-dispatch content: %v", err)
		}
		if string(reviewContent) != "// new review-dispatch content" {
			t.Error("installed review-dispatch content was not updated")
		}
	})
	
	// Test case 3: Identical skip
	t.Run("Identical skip", func(t *testing.T) {
		// Recreate staged plugin directories and files
		stagedBase := filepath.Join(tempDir, "stage3")
		stagedInvokeAgentDir := filepath.Join(stagedBase, "plugins", "invoke-agent")
		stagedReviewDispatchDir := filepath.Join(stagedBase, "plugins", "review-dispatch")
		
		if err := os.MkdirAll(stagedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir staged invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(stagedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir staged review-dispatch plugin dir: %v", err)
		}
		
		if err := os.WriteFile(filepath.Join(stagedInvokeAgentDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write staged invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedReviewDispatchDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write staged review-dispatch index.ts: %v", err)
		}
		
		// Create installed plugin directories and files with identical content
		installedInvokeAgentDir := filepath.Join(pluginsDir, "invoke-agent")
		installedReviewDispatchDir := filepath.Join(pluginsDir, "review-dispatch")
		
		if err := os.MkdirAll(installedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir installed invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir installed review-dispatch plugin dir: %v", err)
		}
		
		if err := os.WriteFile(filepath.Join(installedInvokeAgentDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write installed invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedReviewDispatchDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write installed review-dispatch index.ts: %v", err)
		}
		
		// Create options with Force=false
		opts := &Options{
			TargetDir: tempDir,
			Force:     false,
			Stat:      os.Stat,
			Lstat:     os.Lstat,
			ReadFile:  os.ReadFile,
			RemoveAll: os.RemoveAll,
			Rename:    os.Rename,
			MkdirAll:  os.MkdirAll,
		}
		
		// Call the actual function
		activated, err := atomicallyActivateReviewPlugins(opts, stagedBase)
		if err != nil {
			t.Fatalf("atomicallyActivateReviewPlugins failed: %v", err)
		}
		
		// Verify that plugins were activated
		if len(activated) != 2 {
			t.Errorf("expected 2 activated plugins, got %d", len(activated))
		}
		
		// Verify that staged directories were removed (they should be removed in the skip case too)
		if _, err := os.Stat(stagedInvokeAgentDir); !os.IsNotExist(err) {
			t.Error("staged invoke-agent directory should have been removed")
		}
		if _, err := os.Stat(stagedReviewDispatchDir); !os.IsNotExist(err) {
			t.Error("staged review-dispatch directory should have been removed")
		}
		
		// Verify that installed files still have the original content
		invokeContent, err := os.ReadFile(filepath.Join(installedInvokeAgentDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed invoke-agent content: %v", err)
		}
		if string(invokeContent) != "// identical content" {
			t.Error("installed invoke-agent content should not have changed")
		}
		
		reviewContent, err := os.ReadFile(filepath.Join(installedReviewDispatchDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed review-dispatch content: %v", err)
		}
		if string(reviewContent) != "// identical content" {
			t.Error("installed review-dispatch content should not have changed")
		}
	})
}