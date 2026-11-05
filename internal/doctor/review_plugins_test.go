package doctor

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reviewPluginManifestFixture mirrors the scaffolded .opencode/package.json
// declared dependencies (design D13): two runtime deps, no dev deps.
const reviewPluginManifestFixture = `{
  "name": "unbound-force-opencode-plugins",
  "private": true,
  "type": "module",
  "dependencies": {
    "@opencode-ai/plugin": "1.4.10",
    "zod": "4.1.8"
  }
}`

// reviewPluginInstalledVersions are the versions the fixture installs under
// node_modules; they match reviewPluginManifestFixture's pins.
var reviewPluginInstalledVersions = map[string]string{
	"@opencode-ai/plugin": "1.4.10",
	"zod":                 "4.1.8",
}

// writeInstalledPackages installs every review plugin package under
// .opencode/node_modules at its pinned version.
func writeInstalledPackages(t *testing.T, dir string) {
	t.Helper()
	for _, pkg := range reviewPluginPackages {
		createFile(t, dir, filepath.Join(".opencode", "node_modules", pkg, "package.json"),
			`{"version": "`+reviewPluginInstalledVersions[pkg]+`"}`)
	}
}

// reviewPluginOpencodeJSONFixture mirrors the scaffolded opencode.json
// plugin-array entries for the two review plugins.
const reviewPluginOpencodeJSONFixture = `{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "./.opencode/plugins/invoke-agent/index.ts",
    "./.opencode/plugins/review-dispatch/index.ts"
  ]
}`

const invokeAgentSourceFixture = "export default async () => ({ tool: { invoke_agent: {} } })"

const reviewDispatchSourceFixture = `export default async () => ({
  tool: {
    plan_review_dispatch: {},
    finalize_review_dispatch: {},
    acquire_sibling_evidence: {},
    prepare_lesson_learning: {},
  },
})`

// writeReviewPluginScaffold creates a minimal, consistent review plugin
// scaffold under dir: package.json, node_modules for the
// declared runtime packages, both plugin sources, and the opencode.json
// registration entries.
func writeReviewPluginScaffold(t *testing.T, dir string) {
	t.Helper()
	createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
	writeInstalledPackages(t, dir)
	createFile(t, dir, ".opencode/plugins/invoke-agent/index.ts", invokeAgentSourceFixture)
	createFile(t, dir, ".opencode/plugins/review-dispatch/index.ts", reviewDispatchSourceFixture)
	createFile(t, dir, "opencode.json", reviewPluginOpencodeJSONFixture)
}

func reviewPluginDoctorOpts(dir string, lookPath map[string]string, execCmd map[string]string) *Options {
	return &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		LookPath:  stubLookPath(lookPath),
		ExecCmd:   stubExecCmd(execCmd, nil),
	}
}

func TestParseExactSemver_AcceptsAnchoredASCII(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		allowVPrefix bool
		wantMajor    uint32
		wantMinor    uint32
		wantPatch    uint32
	}{
		{name: "node v prefix", output: "v20.0.0", allowVPrefix: true, wantMajor: 20},
		{name: "node without v prefix", output: "24.1.2\n", allowVPrefix: true, wantMajor: 24, wantMinor: 1, wantPatch: 2},
		{name: "npm CRLF and edge trim", output: "\t 11.6.2 \t\r\n", wantMajor: 11, wantMinor: 6, wantPatch: 2},
		{name: "uint32 boundary", output: "24.4294967295.0", allowVPrefix: true, wantMajor: 24, wantMinor: 4294967295},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExactSemver(tt.output, tt.allowVPrefix)
			if err != nil {
				t.Fatalf("parseExactSemver(%q) error: %v", tt.output, err)
			}
			if got.Major != tt.wantMajor || got.Minor != tt.wantMinor || got.Patch != tt.wantPatch {
				t.Errorf("parseExactSemver(%q) = %+v, want %d.%d.%d", tt.output, got, tt.wantMajor, tt.wantMinor, tt.wantPatch)
			}
		})
	}
}

func TestParseExactSemver_RejectsMalformedAndOverflow(t *testing.T) {
	tests := []struct {
		name         string
		output       string
		allowVPrefix bool
	}{
		{name: "explicit npm v prefix", output: "v10.11.0"},
		{name: "major leading zero", output: "020.1.1", allowVPrefix: true},
		{name: "minor leading zero", output: "20.01.1", allowVPrefix: true},
		{name: "patch leading zero", output: "20.1.01", allowVPrefix: true},
		{name: "Unicode digits", output: "２０.1.1", allowVPrefix: true},
		{name: "positive sign", output: "+20.1.1", allowVPrefix: true},
		{name: "negative sign", output: "-20.1.1", allowVPrefix: true},
		{name: "embedded space", output: "20. 1.1", allowVPrefix: true},
		{name: "embedded tab", output: "20.\t1.1", allowVPrefix: true},
		{name: "prerelease", output: "20.1.1-rc.1", allowVPrefix: true},
		{name: "build metadata", output: "20.1.1+build", allowVPrefix: true},
		{name: "bare carriage return", output: "20.1.1\r", allowVPrefix: true},
		{name: "two terminal line feeds", output: "20.1.1\n\n", allowVPrefix: true},
		{name: "extra line", output: "20.1.1\nextra", allowVPrefix: true},
		{name: "uint32 overflow", output: "20.4294967296.0", allowVPrefix: true},
		{name: "four components", output: "20.1.1.1", allowVPrefix: true},
		{name: "two components", output: "20.1", allowVPrefix: true},
		{name: "uppercase V prefix", output: "V20.1.1", allowVPrefix: true},
		{name: "empty", output: "", allowVPrefix: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := parseExactSemver(tt.output, tt.allowVPrefix); err == nil {
				t.Errorf("parseExactSemver(%q) = %+v, want error", tt.output, got)
			}
		})
	}
}

func TestCheckReviewPlugins_NilWithoutManifest(t *testing.T) {
	opts := &Options{TargetDir: t.TempDir(), ReadFile: os.ReadFile}
	if group := checkReviewPlugins(opts); group != nil {
		t.Fatalf("checkReviewPlugins() = %+v, want nil without .opencode/package.json", group)
	}
}

func TestCheckReviewPlugins_AllPass(t *testing.T) {
	dir := t.TempDir()
	writeReviewPluginScaffold(t, dir)
	opts := reviewPluginDoctorOpts(dir,
		map[string]string{"node": "/usr/local/bin/node", "npm": "/usr/local/bin/npm"},
		map[string]string{"node --version": "v22.15.0", "npm --version": "10.9.2"},
	)

	group := checkReviewPlugins(opts)
	if group == nil {
		t.Fatal("checkReviewPlugins() = nil, want group")
	}
	byName := make(map[string]CheckResult)
	for _, r := range group.Results {
		byName[r.Name] = r
	}
	for _, name := range []string{"manifest-install", "dependencies", "node version", "npm version", "registration", "plugin tools"} {
		r, ok := byName[name]
		if !ok {
			t.Errorf("missing result %q in group", name)
			continue
		}
		if r.Severity != Pass {
			t.Errorf("%s severity = %v, want Pass (message=%q)", name, r.Severity, r.Message)
		}
	}
	if r := byName["node version"]; r.Message != "22.15.0" {
		t.Errorf("node version message = %q, want precise %q", r.Message, "22.15.0")
	}
	if r := byName["npm version"]; r.Message != "10.9.2" {
		t.Errorf("npm version message = %q, want precise %q", r.Message, "10.9.2")
	}
}

func TestCheckManifestInstallConsistency_Mismatch(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
	writeInstalledPackages(t, dir)
	createFile(t, dir, ".opencode/node_modules/zod/package.json", `{"version": "4.2.0"}`)

	opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
	result := checkManifestInstallConsistency(opts)
	if result.Severity != Fail {
		t.Errorf("severity = %v, want Fail (hard version mismatch)", result.Severity)
	}
	if !strings.Contains(result.Message, "zod") {
		t.Errorf("message %q should identify the mismatched package zod", result.Message)
	}
}

func TestCheckManifestInstallConsistency_NotInstalledPasses(t *testing.T) {
	// Packages that are not installed are the dependencies check's
	// concern; with no lock shipped there is nothing else to compare.
	dir := t.TempDir()
	createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
	opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}

	result := checkManifestInstallConsistency(opts)
	if result.Severity != Pass {
		t.Errorf("severity = %v, want Pass (message=%q)", result.Severity, result.Message)
	}
}

func TestCheckManifestInstallConsistency_MissingManifest(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}

	result := checkManifestInstallConsistency(opts)
	if result.Severity != Warn {
		t.Errorf("severity = %v, want Warn (repairable missing manifest)", result.Severity)
	}
	if result.InstallHint == "" {
		t.Error("missing-manifest result should carry a remediation hint")
	}
}

func TestCheckManifestInstallConsistency_UnreadableInstall(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
	writeInstalledPackages(t, dir)
	createFile(t, dir, ".opencode/node_modules/zod/package.json", "{}")

	result := checkManifestInstallConsistency(&Options{TargetDir: dir, ReadFile: os.ReadFile})
	if result.Severity != Fail || !strings.Contains(result.Message, "zod") {
		t.Errorf("result = %v %q, want Fail naming zod", result.Severity, result.Message)
	}
}

func TestCheckManifestInstallConsistency_ReadErrorFails(t *testing.T) {
	// A present-but-unreadable install (EACCES, EISDIR) must not pass
	// silently as if the package were absent.
	dir := t.TempDir()
	createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
	writeInstalledPackages(t, dir)
	readFile := func(path string) ([]byte, error) {
		if strings.Contains(path, filepath.Join("node_modules", "zod")) {
			return nil, fs.ErrPermission
		}
		return os.ReadFile(path)
	}

	result := checkManifestInstallConsistency(&Options{TargetDir: dir, ReadFile: readFile})
	if result.Severity != Fail || !strings.Contains(result.Message, "zod: installed package.json unreadable") {
		t.Errorf("result = %v %q, want Fail naming unreadable zod", result.Severity, result.Message)
	}
}

func TestCheckManifestInstallConsistency_Undeclared(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/package.json", strings.Replace(reviewPluginManifestFixture, `"zod": "4.1.8"`, `"other": "1.0.0"`, 1))

	result := checkManifestInstallConsistency(&Options{TargetDir: dir, ReadFile: os.ReadFile})
	if result.Severity != Fail || !strings.Contains(result.Message, "zod not declared") {
		t.Errorf("result = %v %q, want Fail naming undeclared zod", result.Severity, result.Message)
	}
}

func TestCheckReviewPluginDependencies_Missing(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
	// Only one of the two runtime packages is installed.
	createFile(t, dir, ".opencode/node_modules/zod/package.json", "{}")

	opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
	result := checkReviewPluginDependencies(opts)
	if result.Severity != Warn {
		t.Errorf("severity = %v, want Warn", result.Severity)
	}
	if strings.Contains(result.Message, "vitest") {
		t.Errorf("message %q must not require devDependencies the scaffold never installs", result.Message)
	}
	for _, missing := range []string{"@opencode-ai/plugin"} {
		if !strings.Contains(result.Message, missing) {
			t.Errorf("message %q should report missing %q", result.Message, missing)
		}
	}
}

func TestCheckReviewPluginRuntime_Boundaries(t *testing.T) {
	tests := []struct {
		name      string
		tool      string
		allowV    bool
		output    string
		wantSever Severity
	}{
		{name: "node 19 rejected", tool: "node", allowV: true, output: "v19.9.0", wantSever: Fail},
		{name: "node 20 accepted", tool: "node", allowV: true, output: "v20.0.0", wantSever: Pass},
		{name: "node 24 accepted", tool: "node", allowV: true, output: "v24.9.9", wantSever: Pass},
		{name: "node 25 rejected", tool: "node", allowV: true, output: "v25.0.0", wantSever: Fail},
		{name: "npm 9 rejected", tool: "npm", output: "9.9.9", wantSever: Fail},
		{name: "npm 10 accepted", tool: "npm", output: "10.0.0", wantSever: Pass},
		{name: "npm 11 accepted", tool: "npm", output: "11.0.0", wantSever: Pass},
		{name: "npm 12 rejected", tool: "npm", output: "12.0.0", wantSever: Fail},
		{name: "npm v prefix rejected", tool: "npm", output: "v10.11.0", wantSever: Warn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := reviewPluginDoctorOpts(t.TempDir(),
				map[string]string{tt.tool: "/usr/local/bin/" + tt.tool},
				map[string]string{tt.tool + " --version": tt.output},
			)
			result := checkReviewPluginRuntime(opts, tt.tool, tt.allowV)
			if result.Severity != tt.wantSever {
				t.Errorf("severity = %v, want %v (message=%q)", result.Severity, tt.wantSever, result.Message)
			}
		})
	}
}

func TestCheckReviewPluginRuntime_MissingBinary(t *testing.T) {
	opts := reviewPluginDoctorOpts(t.TempDir(), map[string]string{}, map[string]string{})
	result := checkReviewPluginRuntime(opts, "node", true)
	if result.Severity != Warn {
		t.Errorf("severity = %v, want Warn for missing binary", result.Severity)
	}
	if result.InstallHint == "" {
		t.Error("missing-binary result should carry a runtime remediation hint")
	}
}

func TestCheckPluginRegistration_States(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, dir string)
		wantSever  Severity
		wantSubstr string
	}{
		{
			name: "registered",
			setup: func(t *testing.T, dir string) {
				createFile(t, dir, ".opencode/plugins/invoke-agent/index.ts", invokeAgentSourceFixture)
				createFile(t, dir, ".opencode/plugins/review-dispatch/index.ts", reviewDispatchSourceFixture)
				createFile(t, dir, "opencode.json", reviewPluginOpencodeJSONFixture)
			},
			wantSever:  Pass,
			wantSubstr: "registered",
		},
		{
			name: "present but not registered",
			setup: func(t *testing.T, dir string) {
				createFile(t, dir, ".opencode/plugins/invoke-agent/index.ts", invokeAgentSourceFixture)
				createFile(t, dir, ".opencode/plugins/review-dispatch/index.ts", reviewDispatchSourceFixture)
			},
			wantSever:  Fail,
			wantSubstr: "not registered",
		},
		{
			name:       "inactive",
			setup:      func(t *testing.T, dir string) {},
			wantSever:  Warn,
			wantSubstr: "inactive",
		},
		{
			name: "partial one source",
			setup: func(t *testing.T, dir string) {
				createFile(t, dir, ".opencode/plugins/invoke-agent/index.ts", invokeAgentSourceFixture)
			},
			wantSever:  Warn,
			wantSubstr: "partial activation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)
			result := checkPluginRegistration(&Options{TargetDir: dir, ReadFile: os.ReadFile})
			if result.Severity != tt.wantSever {
				t.Errorf("severity = %v, want %v (message=%q)", result.Severity, tt.wantSever, result.Message)
			}
			if !strings.Contains(result.Message, tt.wantSubstr) {
				t.Errorf("message %q does not contain %q", result.Message, tt.wantSubstr)
			}
		})
	}
}

func TestCheckPluginToolRegistration(t *testing.T) {
	t.Run("registered", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, ".opencode/plugins/invoke-agent/index.ts", invokeAgentSourceFixture)
		createFile(t, dir, ".opencode/plugins/review-dispatch/index.ts", reviewDispatchSourceFixture)
		result := checkPluginToolRegistration(&Options{TargetDir: dir, ReadFile: os.ReadFile})
		if result.Severity != Pass {
			t.Errorf("severity = %v, want Pass (message=%q)", result.Severity, result.Message)
		}
	})

	t.Run("missing tool breaks", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, ".opencode/plugins/invoke-agent/index.ts", invokeAgentSourceFixture)
		// review-dispatch source is missing plan_review_dispatch.
		createFile(t, dir, ".opencode/plugins/review-dispatch/index.ts", `export default async () => ({ tool: {} })`)
		result := checkPluginToolRegistration(&Options{TargetDir: dir, ReadFile: os.ReadFile})
		if result.Severity != Fail {
			t.Errorf("severity = %v, want Fail (broken activation)", result.Severity)
		}
		if !strings.Contains(result.Message, "plan_review_dispatch") {
			t.Errorf("message %q should name the missing tool", result.Message)
		}
	})

	t.Run("no sources inactive", func(t *testing.T) {
		dir := t.TempDir()
		result := checkPluginToolRegistration(&Options{TargetDir: dir, ReadFile: os.ReadFile})
		if result.Severity != Warn {
			t.Errorf("severity = %v, want Warn (inactive)", result.Severity)
		}
	})
}

func TestCheckReviewPlugins_RepairableVsBroken(t *testing.T) {
	t.Run("repairable inactive", func(t *testing.T) {
		// Manifest and node_modules are consistent, but the plugin
		// sources were never activated: recoverable partial.
		dir := t.TempDir()
		createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
		writeInstalledPackages(t, dir)
		opts := reviewPluginDoctorOpts(dir,
			map[string]string{"node": "/usr/local/bin/node", "npm": "/usr/local/bin/npm"},
			map[string]string{"node --version": "v22.0.0", "npm --version": "10.0.0"},
		)
		group := checkReviewPlugins(opts)
		byName := resultMap(group)
		if r := byName["registration"]; r.Severity != Warn {
			t.Errorf("registration severity = %v, want Warn (repairable)", r.Severity)
		}
		if r := byName["plugin tools"]; r.Severity != Warn {
			t.Errorf("plugin tools severity = %v, want Warn (repairable)", r.Severity)
		}
		// Everything else that is repairable stays non-Fail.
		if r := byName["manifest-install"]; r.Severity == Fail {
			t.Errorf("manifest-install severity = Fail, want non-Fail in repairable state")
		}
	})

	t.Run("broken mismatched install", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, ".opencode/package.json", reviewPluginManifestFixture)
		writeInstalledPackages(t, dir)
		createFile(t, dir, ".opencode/node_modules/@opencode-ai/plugin/package.json", `{"version": "9.9.9"}`)
		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkReviewPlugins(opts)
		byName := resultMap(group)
		if r := byName["manifest-install"]; r.Severity != Fail {
			t.Errorf("manifest-install severity = %v, want Fail (broken)", r.Severity)
		}
	})
}

// resultMap indexes a CheckGroup's results by name for assertion ergonomics.
func resultMap(group *CheckGroup) map[string]CheckResult {
	m := make(map[string]CheckResult)
	if group == nil {
		return m
	}
	for _, r := range group.Results {
		m[r.Name] = r
	}
	return m
}
