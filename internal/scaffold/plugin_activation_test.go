package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
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
			case "npm install --ignore-scripts --omit=dev --no-package-lock --no-audit --no-fund":
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
		WriteFile: func(path string, data []byte, perm os.FileMode) error {
			installedPluginsDir := filepath.Join(opencodeDir, "plugins")
			invokeDest := filepath.Join(installedPluginsDir, "invoke-agent", "index.ts")
			reviewDest := filepath.Join(installedPluginsDir, "review-dispatch", "index.ts")
			switch path {
			case invokeDest:
				if renameCount == 0 {
					assertPluginsAbsent("before atomic activation")
				}
				events = append(events, "activate invoke-agent")
				renameCount++
			case reviewDest:
				if renameCount == 0 {
					assertPluginsAbsent("before atomic activation")
				}
				events = append(events, "activate review-dispatch")
				renameCount++
			}
			return os.WriteFile(path, data, perm)
		},
		Rename: func(source, destination string) error {
			t.Fatalf("unexpected Rename call: %s -> %s", source, destination)
			return nil
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
		"npm install --ignore-scripts --omit=dev --no-package-lock --no-audit --no-fund",
		"lookpath opencode",
		"import probe",
		"tool-definition probe",
		"activate invoke-agent",
		"activate review-dispatch",
	}
	if len(events) < len(wantActivationEvents) || !reflect.DeepEqual(events[:len(wantActivationEvents)], wantActivationEvents) {
		t.Errorf("activation operation order =\n%v\nwant prefix\n%v", events, wantActivationEvents)
	}

	for _, assetPath := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset, ufWorkflowPluginAsset, ufFinalePluginAsset} {
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
		"./" + mapAssetPath(ufWorkflowPluginAsset),
		"./" + mapAssetPath(ufFinalePluginAsset),
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

// TestSCFR003_ForeignLockSurvivesRun composes the observable behavior
// behind the "Lock written by another tool" scenario: a foreign lock
// written into the target before Run() survives byte-identical, is
// never reported as written, and the activation install runs with
// --no-package-lock so the lock is neither read nor written.
func TestSCFR003_ForeignLockSurvivesRun(t *testing.T) {
	targetDir := t.TempDir()
	opencodeDir := filepath.Join(targetDir, ".opencode")
	if err := os.MkdirAll(opencodeDir, 0o755); err != nil {
		t.Fatalf("create .opencode dir: %v", err)
	}

	// opencode.json with a user-owned plugin so the activation probe
	// has a config to read, matching the other activation tests.
	opencodeJSON := "{\n  \"plugin\": [\"user-owned-plugin\"]\n}\n"
	if err := os.WriteFile(filepath.Join(targetDir, "opencode.json"), []byte(opencodeJSON), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	// A realistic foreign lock, diverging from anything the scaffold
	// ships, pre-written before Run().
	foreignLock := []byte(`{
  "name": "foreign-project",
  "version": "1.0.0",
  "lockfileVersion": 3,
  "requires": true,
  "packages": {
    "": {
      "name": "foreign-project",
      "version": "1.0.0",
      "dependencies": {
        "left-pad": "^1.3.0"
      }
    },
    "node_modules/left-pad": {
      "version": "1.3.0",
      "resolved": "https://registry.npmjs.org/left-pad/-/left-pad-1.3.0.tgz",
      "integrity": "sha512-XI5MPZVNTjYzN4Gk+lnMdFuZ2swrEHP7t6v/psU7hWSKVstLgvXWraHbjS4bPT/MGY4wdJhWN7EjJSbO9MBzA==",
      "license": "MIT"
    }
  }
}
`)
	lockPath := filepath.Join(opencodeDir, "package-lock.json")
	if err := os.WriteFile(lockPath, foreignLock, 0o644); err != nil {
		t.Fatalf("pre-write foreign lock: %v", err)
	}

	var stdout bytes.Buffer
	opts := reviewPluginRunOpts(t, targetDir, &stdout, false)
	baseExec := opts.ExecCmdInDir
	var installCommands []string
	opts.ExecCmdInDir = func(dir, name string, args ...string) ([]byte, error) {
		command := filepath.Base(name) + " " + strings.Join(args, " ")
		if strings.HasPrefix(command, "npm install") {
			installCommands = append(installCommands, command)
		}
		return baseExec(dir, name, args...)
	}

	result, err := Run(opts)
	if err != nil {
		t.Fatalf("Run() error = %v, want nil", err)
	}

	// The foreign lock must survive byte-identical.
	got, readErr := os.ReadFile(lockPath)
	if readErr != nil {
		t.Fatalf("foreign lock removed by Run(): %v", readErr)
	}
	if !bytes.Equal(got, foreignLock) {
		t.Error("foreign lock content was modified by Run()")
	}

	// Run() must not report the foreign lock in any write group.
	lockRel := filepath.ToSlash(filepath.Join(".opencode", "package-lock.json"))
	for name, group := range map[string][]string{
		"Created":     result.Created,
		"Overwritten": result.Overwritten,
		"Updated":     result.Updated,
	} {
		if containsPath(group, lockRel) {
			t.Errorf("foreign lock reported as %s by Run()", name)
		}
	}

	// The activation install must have run without touching the lock.
	if len(installCommands) == 0 {
		t.Fatal("activation install never ran; test cannot verify the lock is not read")
	}
	for _, command := range installCommands {
		if !strings.Contains(command, "--no-package-lock") {
			t.Errorf("install command %q must pin --no-package-lock so the foreign lock is neither read nor written", command)
		}
	}
}

// reviewPluginRunOpts builds an Options whose injected command fakes drive
// the review plugin activation flow. When failInstall is true, `npm install`
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
			case "npm install --ignore-scripts --omit=dev --no-package-lock --no-audit --no-fund":
				if failInstall {
					return []byte("npm ERR! code ENOTCACHED\n"), fmt.Errorf("npm install failed: offline")
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
	for _, want := range []string{"review-plugins", "failed", "activation: inactive", "npm install --ignore-scripts --omit=dev --no-package-lock"} {
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

	// First run: npm install fails → soft partial result, no plugin source.
	firstOut := &bytes.Buffer{}
	first, err := Run(reviewPluginRunOpts(t, targetDir, firstOut, true))
	if err != nil {
		t.Fatalf("first Run() error = %v, want nil", err)
	}
	if first.Status != resultStatusPartial || first.FailedSubTools != 1 {
		t.Fatalf("first Run() = Status %q / FailedSubTools %d, want partial/1", first.Status, first.FailedSubTools)
	}

	// Second run: npm install succeeds → plugins activate, result fully populated.
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
	for _, assetPath := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset, ufWorkflowPluginAsset, ufFinalePluginAsset} {
		targetPath := filepath.Join(targetDir, mapAssetPath(assetPath))
		if _, statErr := os.Stat(targetPath); statErr != nil {
			t.Errorf("retry did not activate plugin source %s: %v", assetPath, statErr)
		}
		if !containsPath(second.Created, mapAssetPath(assetPath)) {
			t.Errorf("retry did not record activated plugin %s in result.Created", assetPath)
		}
	}
}

func TestStageReviewPluginSources_CoversImportedLibModules(t *testing.T) {
	stageRoot := t.TempDir()
	opts := &Options{
		MkdirAll:  os.MkdirAll,
		WriteFile: os.WriteFile,
	}
	if err := stageReviewPluginSources(opts, stageRoot); err != nil {
		t.Fatalf("stageReviewPluginSources() error: %v", err)
	}

	// The lib layout is flat: every module lives at opencode/lib/<name>.ts and
	// imports siblings as "./<name>.js". Specifiers that escape the flat
	// grammars below — parent-relative "../x.js" imports or nested "./x/y.js"
	// paths — would silently skip the closure walk, so they fail the test
	// loudly through the non-flat scans instead.
	pluginLibSpecifier := regexp.MustCompile(`["']\.\./\.\./lib/([A-Za-z0-9._-]+)\.js["']`)
	libLocalSpecifier := regexp.MustCompile(`["']\./([A-Za-z0-9._-]+)\.js["']`)
	nonFlatPluginLibSpecifier := regexp.MustCompile(`["']\.\./\.\./lib/[A-Za-z0-9._-]*/[^"']*\.js["']`)
	nonFlatLibSpecifier := regexp.MustCompile(`["']\.\./[^"']*\.js["']|["']\./[A-Za-z0-9._-]*/[^"']*\.js["']`)

	stagedPluginAssets := []string{
		invokeAgentPluginAsset,
		reviewDispatchPluginAsset,
		ufWorkflowPluginAsset,
		ufFinalePluginAsset,
	}

	// requiredModules is the transitive closure of lib modules referenced by
	// the staged plugin sources and their staged lib imports.
	requiredModules := make(map[string]bool)
	var requireModule func(name string)
	requireModule = func(name string) {
		if requiredModules[name] {
			return
		}
		requiredModules[name] = true
		content, err := assetContent("opencode/lib/" + name + ".ts")
		if err != nil {
			// assetContent failures are reported by the assertion loop
			// below; the walk cannot traverse a module with no
			// embedded asset.
			return
		}
		for _, match := range libLocalSpecifier.FindAllStringSubmatch(string(content), -1) {
			requireModule(match[1])
		}
		for _, match := range nonFlatLibSpecifier.FindAllString(string(content), -1) {
			t.Errorf("lib module %s contains a non-flat local import %s; the closure walk covers only the flat opencode/lib layout", name, match)
		}
	}

	for _, assetPath := range stagedPluginAssets {
		content, err := assetContent(assetPath)
		if err != nil {
			t.Fatalf("read embedded plugin %s: %v", assetPath, err)
		}
		for _, match := range pluginLibSpecifier.FindAllStringSubmatch(string(content), -1) {
			requireModule(match[1])
		}
		for _, match := range nonFlatPluginLibSpecifier.FindAllString(string(content), -1) {
			t.Errorf("plugin %s contains a non-flat lib import %s; the closure walk covers only the flat opencode/lib layout", assetPath, match)
		}
	}

	if len(requiredModules) == 0 {
		t.Fatal("no lib modules derived from staged plugin sources; closure extraction found no specifiers")
	}

	for module := range requiredModules {
		assetPath := "opencode/lib/" + module + ".ts"
		embedded, err := assetContent(assetPath)
		if err != nil {
			t.Errorf("lib module %s referenced by staged sources has no embedded asset %s: %v", module, assetPath, err)
			continue
		}
		stagedPath := filepath.Join(stageRoot, "lib", module+".ts")
		staged, err := os.ReadFile(stagedPath)
		if err != nil {
			t.Errorf("lib module %s referenced by staged sources is missing from the staging list (%s): %v", module, stagedPath, err)
			continue
		}
		// Content lockstep: staged sources must stay byte-identical to the
		// embedded assets the closure walk derives from. A divergence would
		// deploy review plugins whose staged copy silently drifts from the
		// embedded registry.
		if !bytes.Equal(staged, embedded) {
			t.Errorf("staged lib module %s (%s) differs from embedded asset %s", module, stagedPath, assetPath)
		}
		// Ownership lockstep: every module the staged plugins import must be
		// tool-owned so uf init overwrites it in generated projects, and a
		// Divisor asset so DivisorOnly deployments stage it. A module missing
		// from either registry would stage once and then drift.
		if !isToolOwned(assetPath) {
			t.Errorf("lib module %s is imported by staged plugin sources but isToolOwned(%s) is false; uf init would never update it", module, assetPath)
		}
		if !isDivisorAsset(assetPath) {
			t.Errorf("lib module %s is imported by staged plugin sources but isDivisorAsset(%s) is false; DivisorOnly deployments would not stage it", module, assetPath)
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
		stagedUfWorkflowDir := filepath.Join(stagedBase, "plugins", "uf-workflow")
		if err := os.MkdirAll(stagedUfWorkflowDir, 0o755); err != nil {
			t.Fatalf("mkdir staged uf-workflow plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedUfWorkflowDir, "index.ts"), []byte("// uf-workflow content"), 0o644); err != nil {
			t.Fatalf("write staged uf-workflow index.ts: %v", err)
		}
		stagedUfFinaleDir := filepath.Join(stagedBase, "plugins", "uf-finale")
		if err := os.MkdirAll(stagedUfFinaleDir, 0o755); err != nil {
			t.Fatalf("mkdir staged uf-finale plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedUfFinaleDir, "index.ts"), []byte("// uf-finale content"), 0o644); err != nil {
			t.Fatalf("write staged uf-finale index.ts: %v", err)
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
		installedUfWorkflowDir := filepath.Join(pluginsDir, "uf-workflow")
		if err := os.MkdirAll(installedUfWorkflowDir, 0o755); err != nil {
			t.Fatalf("mkdir installed uf-workflow plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedUfWorkflowDir, "index.ts"), []byte("// old uf-workflow content"), 0o644); err != nil {
			t.Fatalf("write installed uf-workflow index.ts: %v", err)
		}
		installedUfFinaleDir := filepath.Join(pluginsDir, "uf-finale")
		if err := os.MkdirAll(installedUfFinaleDir, 0o755); err != nil {
			t.Fatalf("mkdir installed uf-finale plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedUfFinaleDir, "index.ts"), []byte("// old uf-finale content"), 0o644); err != nil {
			t.Fatalf("write installed uf-finale index.ts: %v", err)
		}

		// Create options with Force=true
		opts := &Options{
			TargetDir: tempDir,
			Force:     true,
			Stat:      os.Stat,
			Lstat:     os.Lstat,
			ReadFile:  os.ReadFile,
			WriteFile: os.WriteFile,
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
		if len(activated) != 4 {
			t.Errorf("expected 4 activated plugins, got %d", len(activated))
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
		ufWorkflowContent, err := os.ReadFile(filepath.Join(installedUfWorkflowDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed uf-workflow content: %v", err)
		}
		if string(ufWorkflowContent) != "// uf-workflow content" {
			t.Error("installed uf-workflow content was not updated")
		}
		ufFinaleContent, err := os.ReadFile(filepath.Join(installedUfFinaleDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed uf-finale content: %v", err)
		}
		if string(ufFinaleContent) != "// uf-finale content" {
			t.Error("installed uf-finale content was not updated")
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
		stagedUfWorkflowDir := filepath.Join(stagedBase, "plugins", "uf-workflow")
		if err := os.MkdirAll(stagedUfWorkflowDir, 0o755); err != nil {
			t.Fatalf("mkdir staged uf-workflow plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedUfWorkflowDir, "index.ts"), []byte("// new uf-workflow content"), 0o644); err != nil {
			t.Fatalf("write staged uf-workflow index.ts: %v", err)
		}
		stagedUfFinaleDir := filepath.Join(stagedBase, "plugins", "uf-finale")
		if err := os.MkdirAll(stagedUfFinaleDir, 0o755); err != nil {
			t.Fatalf("mkdir staged uf-finale plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedUfFinaleDir, "index.ts"), []byte("// new uf-finale content"), 0o644); err != nil {
			t.Fatalf("write staged uf-finale index.ts: %v", err)
		}

		// Create installed plugin directories and files with different content
		installedInvokeAgentDir := filepath.Join(pluginsDir, "invoke-agent")
		installedReviewDispatchDir := filepath.Join(pluginsDir, "review-dispatch")
		installedUfWorkflowDir := filepath.Join(pluginsDir, "uf-workflow")
		installedUfFinaleDir := filepath.Join(pluginsDir, "uf-finale")

		if err := os.MkdirAll(installedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir installed invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir installed review-dispatch plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedUfWorkflowDir, 0o755); err != nil {
			t.Fatalf("mkdir installed uf-workflow plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedUfFinaleDir, 0o755); err != nil {
			t.Fatalf("mkdir installed uf-finale plugin dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(installedInvokeAgentDir, "index.ts"), []byte("// old invoke-agent content"), 0o644); err != nil {
			t.Fatalf("write installed invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedReviewDispatchDir, "index.ts"), []byte("// old review-dispatch content"), 0o644); err != nil {
			t.Fatalf("write installed review-dispatch index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedUfWorkflowDir, "index.ts"), []byte("// old uf-workflow content"), 0o644); err != nil {
			t.Fatalf("write installed uf-workflow index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedUfFinaleDir, "index.ts"), []byte("// old uf-finale content"), 0o644); err != nil {
			t.Fatalf("write installed uf-finale index.ts: %v", err)
		}

		// Create options with Force=false
		opts := &Options{
			TargetDir: tempDir,
			Force:     false,
			Stat:      os.Stat,
			Lstat:     os.Lstat,
			ReadFile:  os.ReadFile,
			WriteFile: os.WriteFile,
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
		if len(activated) != 4 {
			t.Errorf("expected 4 activated plugins, got %d", len(activated))
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
		ufWorkflowContent, err := os.ReadFile(filepath.Join(installedUfWorkflowDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed uf-workflow content: %v", err)
		}
		if string(ufWorkflowContent) != "// new uf-workflow content" {
			t.Error("installed uf-workflow content was not updated")
		}
		ufFinaleContent, err := os.ReadFile(filepath.Join(installedUfFinaleDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed uf-finale content: %v", err)
		}
		if string(ufFinaleContent) != "// new uf-finale content" {
			t.Error("installed uf-finale content was not updated")
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
		stagedUfWorkflowDir := filepath.Join(stagedBase, "plugins", "uf-workflow")
		if err := os.MkdirAll(stagedUfWorkflowDir, 0o755); err != nil {
			t.Fatalf("mkdir staged uf-workflow plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedUfWorkflowDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write staged uf-workflow index.ts: %v", err)
		}
		stagedUfFinaleDir := filepath.Join(stagedBase, "plugins", "uf-finale")
		if err := os.MkdirAll(stagedUfFinaleDir, 0o755); err != nil {
			t.Fatalf("mkdir staged uf-finale plugin dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(stagedUfFinaleDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write staged uf-finale index.ts: %v", err)
		}

		// Create installed plugin directories and files with identical content
		installedInvokeAgentDir := filepath.Join(pluginsDir, "invoke-agent")
		installedReviewDispatchDir := filepath.Join(pluginsDir, "review-dispatch")
		installedUfWorkflowDir := filepath.Join(pluginsDir, "uf-workflow")
		installedUfFinaleDir := filepath.Join(pluginsDir, "uf-finale")

		if err := os.MkdirAll(installedInvokeAgentDir, 0o755); err != nil {
			t.Fatalf("mkdir installed invoke-agent plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedReviewDispatchDir, 0o755); err != nil {
			t.Fatalf("mkdir installed review-dispatch plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedUfWorkflowDir, 0o755); err != nil {
			t.Fatalf("mkdir installed uf-workflow plugin dir: %v", err)
		}
		if err := os.MkdirAll(installedUfFinaleDir, 0o755); err != nil {
			t.Fatalf("mkdir installed uf-finale plugin dir: %v", err)
		}

		if err := os.WriteFile(filepath.Join(installedInvokeAgentDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write installed invoke-agent index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedReviewDispatchDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write installed review-dispatch index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedUfWorkflowDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write installed uf-workflow index.ts: %v", err)
		}
		if err := os.WriteFile(filepath.Join(installedUfFinaleDir, "index.ts"), []byte("// identical content"), 0o644); err != nil {
			t.Fatalf("write installed uf-finale index.ts: %v", err)
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
		if len(activated) != 4 {
			t.Errorf("expected 4 activated plugins, got %d", len(activated))
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
		ufWorkflowContent, err := os.ReadFile(filepath.Join(installedUfWorkflowDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed uf-workflow content: %v", err)
		}
		if string(ufWorkflowContent) != "// identical content" {
			t.Error("installed uf-workflow content should not have changed")
		}
		ufFinaleContent, err := os.ReadFile(filepath.Join(installedUfFinaleDir, "index.ts"))
		if err != nil {
			t.Fatalf("read installed uf-finale content: %v", err)
		}
		if string(ufFinaleContent) != "// identical content" {
			t.Error("installed uf-finale content should not have changed")
		}
	})
}

func TestPreservePluginDevDependencies_RestoresRepositoryManifest(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}
	repo, err := os.ReadFile(filepath.Join(root, ".opencode", "package.json"))
	if err != nil {
		t.Fatalf("read repository manifest: %v", err)
	}
	shipped, err := assetContent(pluginManifestAsset)
	if err != nil {
		t.Fatalf("read shipped manifest: %v", err)
	}
	// uf init over this repository must leave its dev manifest untouched.
	if got := preservePluginDevDependencies(shipped, repo); !bytes.Equal(got, repo) {
		t.Errorf("merged manifest differs from the repository manifest:\n%s", got)
	}
}

func TestPreservePluginDevDependencies_Cases(t *testing.T) {
	shipped, err := assetContent(pluginManifestAsset)
	if err != nil {
		t.Fatalf("read shipped manifest: %v", err)
	}
	for _, existing := range []string{`{"dependencies": {"zod": "4.6.6"}}`, `not json`, `{"devDependencies": {}}`} {
		if got := preservePluginDevDependencies(shipped, []byte(existing)); !bytes.Equal(got, shipped) {
			t.Errorf("existing %q: got %s, want the shipped manifest unchanged", existing, got)
		}
	}
	got := preservePluginDevDependencies(shipped, []byte(`{"devDependencies": {"z-tool": ">=1.0.0", "a-tool": "2.0.0"}}`))
	var merged struct {
		DevDependencies map[string]string `json:"devDependencies"`
		Overrides       map[string]string `json:"overrides"`
	}
	if err := json.Unmarshal(got, &merged); err != nil {
		t.Fatalf("merged manifest is not valid JSON: %v\n%s", err, got)
	}
	if merged.DevDependencies["z-tool"] != ">=1.0.0" || merged.DevDependencies["a-tool"] != "2.0.0" || len(merged.Overrides) == 0 {
		t.Errorf("merged = %+v", merged)
	}
	if !bytes.Contains(got, []byte(`"z-tool": ">=1.0.0"`)) || bytes.Index(got, []byte("a-tool")) > bytes.Index(got, []byte("z-tool")) {
		t.Errorf("devDependencies must be sorted and unescaped:\n%s", got)
	}
	if got := preservePluginDevDependencies([]byte("{}\n"), []byte(`{"devDependencies": {"a": "1"}}`)); string(got) != "{}\n" {
		t.Errorf("manifest without an overrides anchor must be returned unchanged, got %s", got)
	}
}
