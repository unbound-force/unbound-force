package doctor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// --- Test helpers ---

// stubLookPath returns a function that simulates exec.LookPath.
// Keys in the map are binary names; values are their paths.
func stubLookPath(found map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := found[name]; ok {
			return path, nil
		}
		return "", fmt.Errorf("executable %q not found", name)
	}
}

// stubLookPathSimple returns a LookPath that returns /usr/local/bin/<name>
// for found binaries.
func stubLookPathSimple(found map[string]bool) func(string) (string, error) {
	return func(name string) (string, error) {
		if found[name] {
			return "/usr/local/bin/" + name, nil
		}
		return "", fmt.Errorf("executable %q not found", name)
	}
}

// stubExecCmd returns a function that returns canned output for commands.
func stubExecCmd(outputs map[string]string, errors map[string]error) func(string, ...string) ([]byte, error) {
	return func(name string, args ...string) ([]byte, error) {
		key := name
		if len(args) > 0 {
			key = name + " " + strings.Join(args, " ")
		}
		if err, ok := errors[key]; ok {
			out := ""
			if o, ok2 := outputs[key]; ok2 {
				out = o
			}
			return []byte(out), err
		}
		if out, ok := outputs[key]; ok {
			return []byte(out), nil
		}
		return nil, fmt.Errorf("command %q not stubbed", key)
	}
}

// stubExecCmdTimeout wraps a stubExecCmd to match ExecCmdTimeout signature.
// The timeout parameter is ignored since tests control behavior via the map.
func stubExecCmdTimeout(outputs map[string]string, errors map[string]error) func(time.Duration, string, ...string) ([]byte, error) {
	inner := stubExecCmd(outputs, errors)
	return func(_ time.Duration, name string, args ...string) ([]byte, error) {
		return inner(name, args...)
	}
}

// stubEvalSymlinks returns a function that resolves paths via a map.
func stubEvalSymlinks(resolved map[string]string) func(string) (string, error) {
	return func(path string) (string, error) {
		if r, ok := resolved[path]; ok {
			return r, nil
		}
		return path, nil // identity if not in map
	}
}

// stubGetenv returns a function that reads env vars from a map.
func stubGetenv(vars map[string]string) func(string) string {
	return func(key string) string {
		return vars[key]
	}
}

// stubReadFile returns a function that reads files from a map.
func stubReadFile(files map[string][]byte) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		if data, ok := files[path]; ok {
			return data, nil
		}
		return nil, fmt.Errorf("file %q not found", path)
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

// --- Phase 2: Foundation tests (T010) ---

func TestDetectEnvironment_GoenvDetected(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPath(map[string]string{"goenv": "/opt/homebrew/bin/goenv"}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}

	env := DetectEnvironment(opts)

	found := false
	for _, m := range env.Managers {
		if m.Kind == ManagerGoenv {
			found = true
			if m.Path != "/opt/homebrew/bin/goenv" {
				t.Errorf("goenv path = %q, want /opt/homebrew/bin/goenv", m.Path)
			}
		}
	}
	if !found {
		t.Error("goenv not detected")
	}
}

func TestDetectEnvironment_NvmDetected(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPathSimple(map[string]bool{}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{"NVM_DIR": "/home/user/.nvm"}),
	}

	env := DetectEnvironment(opts)

	found := false
	for _, m := range env.Managers {
		if m.Kind == ManagerNvm {
			found = true
			if m.Path != "/home/user/.nvm" {
				t.Errorf("nvm path = %q, want /home/user/.nvm", m.Path)
			}
		}
	}
	if !found {
		t.Error("nvm not detected via NVM_DIR")
	}
}

func TestDetectEnvironment_HomebrewDetected(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPath(map[string]string{"brew": "/opt/homebrew/bin/brew"}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}

	env := DetectEnvironment(opts)

	found := false
	for _, m := range env.Managers {
		if m.Kind == ManagerHomebrew {
			found = true
		}
	}
	if !found {
		t.Error("Homebrew not detected")
	}
}

func TestDetectEnvironment_DnfDetected(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPath(map[string]string{"dnf": "/usr/bin/dnf"}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}

	env := DetectEnvironment(opts)

	found := false
	for _, m := range env.Managers {
		if m.Kind == ManagerDnf {
			found = true
			if len(m.Manages) != 1 || m.Manages[0] != "packages" {
				t.Errorf("dnf manages = %v, want [packages]", m.Manages)
			}
		}
	}
	if !found {
		t.Error("dnf not detected")
	}
}

func TestDetectEnvironment_DnfNotDetected(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPathSimple(map[string]bool{}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}

	env := DetectEnvironment(opts)

	for _, m := range env.Managers {
		if m.Kind == ManagerDnf {
			t.Error("dnf should not be detected when not in PATH")
		}
	}
}

func TestDetectEnvironment_NoManagers(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPathSimple(map[string]bool{}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}

	env := DetectEnvironment(opts)

	if env.Managers == nil {
		t.Error("Managers should be empty slice, not nil")
	}
	if len(env.Managers) != 0 {
		t.Errorf("expected 0 managers, got %d", len(env.Managers))
	}
}

func TestDetectEnvironment_Platform(t *testing.T) {
	opts := &Options{
		LookPath:     stubLookPathSimple(map[string]bool{}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
	}

	env := DetectEnvironment(opts)

	if env.Platform == "" {
		t.Error("Platform should not be empty")
	}
	if !strings.Contains(env.Platform, "/") {
		t.Errorf("Platform %q should contain /", env.Platform)
	}
}

func TestDetectProvenance_AllManagerKinds(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		resolved map[string]string
		want     ManagerKind
	}{
		{"goenv shims", "/home/user/.goenv/shims/go", nil, ManagerGoenv},
		{"goenv versions", "/home/user/.goenv/versions/1.24.3/bin/go", nil, ManagerGoenv},
		{"pyenv shims", "/home/user/.pyenv/shims/python", nil, ManagerPyenv},
		{"pyenv versions", "/home/user/.pyenv/versions/3.12/bin/python", nil, ManagerPyenv},
		{"nvm", "/home/user/.nvm/versions/node/v22.15.0/bin/node", nil, ManagerNvm},
		{"fnm multishells", "/tmp/fnm_multishells/123/bin/node", nil, ManagerFnm},
		{"fnm node-versions", "/home/user/.local/share/fnm/node-versions/v22/bin/node", nil, ManagerFnm},
		{"mise installs", "/home/user/.local/share/mise/installs/go/1.24/bin/go", nil, ManagerMise},
		{"mise shims", "/home/user/.local/share/mise/shims/go", nil, ManagerMise},
		{"bun", "/home/user/.bun/bin/bun", nil, ManagerBun},
		{"homebrew", "/usr/local/bin/gaze", map[string]string{"/usr/local/bin/gaze": "/usr/local/Cellar/gaze/0.10.0/bin/gaze"}, ManagerHomebrew},
		{"direct", "/usr/local/go/bin/go", nil, ManagerDirect},
		{"system usr/bin", "/usr/bin/python3", nil, ManagerSystem},
		{"system snap", "/snap/bin/go", nil, ManagerSystem},
		{"unknown", "/opt/custom/bin/tool", nil, ManagerUnknown},
		{"empty path", "", nil, ManagerUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{
				EvalSymlinks: stubEvalSymlinks(tt.resolved),
			}
			got := DetectProvenance(tt.path, opts)
			if got != tt.want {
				t.Errorf("DetectProvenance(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestInstallHint_GoenvDetected(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerGoenv, Path: "/opt/homebrew/bin/goenv", Manages: []string{"go"}},
		},
	}

	hint := installHint("go", env)
	if !strings.Contains(hint, "goenv install") {
		t.Errorf("hint = %q, want goenv install command", hint)
	}
}

func TestInstallHint_HomebrewFallback(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerHomebrew, Path: "/opt/homebrew/bin/brew", Manages: []string{"packages"}},
		},
	}

	hint := installHint("gaze", env)
	if !strings.Contains(hint, "brew install") {
		t.Errorf("hint = %q, want brew install command", hint)
	}
}

func TestInstallHint_NoManager(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{},
	}

	hint := installHint("opencode", env)
	if !strings.Contains(hint, "curl") {
		t.Errorf("hint = %q, want curl install command", hint)
	}
}

func TestInstallHint_NvmForNode(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerNvm, Path: "/home/user/.nvm", Manages: []string{"node"}},
		},
	}

	hint := installHint("node", env)
	if !strings.Contains(hint, "nvm install") {
		t.Errorf("hint = %q, want nvm install command", hint)
	}
}

func TestSeverity_MarshalJSON(t *testing.T) {
	tests := []struct {
		sev  Severity
		want string
	}{
		{Pass, `"pass"`},
		{Warn, `"warn"`},
		{Fail, `"fail"`},
	}

	for _, tt := range tests {
		data, err := json.Marshal(tt.sev)
		if err != nil {
			t.Fatalf("Marshal(%v): %v", tt.sev, err)
		}
		if string(data) != tt.want {
			t.Errorf("Marshal(%v) = %s, want %s", tt.sev, data, tt.want)
		}
	}
}

func TestSeverity_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Severity
	}{
		{`"pass"`, Pass},
		{`"warn"`, Warn},
		{`"fail"`, Fail},
	}

	for _, tt := range tests {
		var got Severity
		if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", tt.input, err)
		}
		if got != tt.want {
			t.Errorf("Unmarshal(%s) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSeverity_UnmarshalJSON_Invalid(t *testing.T) {
	var got Severity
	err := json.Unmarshal([]byte(`"invalid"`), &got)
	if err == nil {
		t.Error("expected error for invalid severity string")
	}
}

// --- Phase 3: User Story 1 tests (T011-T022) ---

func TestCheckCoreTools(t *testing.T) {
	opts := &Options{
		TargetDir: t.TempDir(),
		LookPath: stubLookPath(map[string]string{
			"go":         "/home/user/.goenv/shims/go",
			"gaze":       "/opt/homebrew/bin/gaze",
			"dewey":      "/opt/homebrew/bin/dewey",
			"node":       "/home/user/.nvm/versions/node/v22.15.0/bin/node",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version":     "go version go1.24.3 darwin/arm64",
				"node --version": "v22.15.0",
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(map[string]string{
			"/opt/homebrew/bin/gaze": "/opt/homebrew/Cellar/gaze/0.10.0/bin/gaze",
		}),
		Getenv:   stubGetenv(map[string]string{}),
		ReadFile: stubReadFile(nil),
	}

	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerGoenv, Path: "/opt/homebrew/bin/goenv", Manages: []string{"go"}},
			{Kind: ManagerNvm, Path: "/home/user/.nvm", Manages: []string{"node"}},
			{Kind: ManagerHomebrew, Path: "/opt/homebrew/bin/brew", Manages: []string{"packages"}},
		},
	}

	group := checkCoreTools(opts, env)

	if group.Name != "Core Tools" {
		t.Errorf("group name = %q, want Core Tools", group.Name)
	}

	// Build result map for easy lookup.
	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	// go: found via goenv, version 1.24.3
	if r, ok := results["go"]; ok {
		if r.Severity != Pass {
			t.Errorf("go severity = %v, want Pass", r.Severity)
		}
		if !strings.Contains(r.Message, "1.24.3") {
			t.Errorf("go message = %q, want version 1.24.3", r.Message)
		}
		if !strings.Contains(r.Message, "goenv") {
			t.Errorf("go message = %q, want 'via goenv'", r.Message)
		}
	} else {
		t.Error("go check not found")
	}

	// opencode: not found -> Fail
	if r, ok := results["opencode"]; ok {
		if r.Severity != Fail {
			t.Errorf("opencode severity = %v, want Fail", r.Severity)
		}
		if r.InstallHint == "" {
			t.Error("opencode should have install hint")
		}
	} else {
		t.Error("opencode check not found")
	}

	// gaze: found via Homebrew
	if r, ok := results["gaze"]; ok {
		if r.Severity != Pass {
			t.Errorf("gaze severity = %v, want Pass", r.Severity)
		}
	} else {
		t.Error("gaze check not found")
	}

	// mxf: not found -> Warn (recommended)
	if r, ok := results["mxf"]; ok {
		if r.Severity != Warn {
			t.Errorf("mxf severity = %v, want Warn", r.Severity)
		}
	} else {
		t.Error("mxf check not found")
	}

	// dewey is checked in the dedicated "Dewey Knowledge Layer" group,
	// not in Core Tools. See TestCheckDewey_* tests.

	// node: found via nvm, version 22.15.0
	if r, ok := results["node"]; ok {
		if r.Severity != Pass {
			t.Errorf("node severity = %v, want Pass", r.Severity)
		}
		if !strings.Contains(r.Message, "22.15.0") {
			t.Errorf("node message = %q, want version 22.15.0", r.Message)
		}
		if !strings.Contains(r.Message, "nvm") {
			t.Errorf("node message = %q, want 'via nvm'", r.Message)
		}
	} else {
		t.Error("node check not found")
	}

	// gh: not found -> Pass (optional, informational)
	if r, ok := results["gh"]; ok {
		if r.Severity != Pass {
			t.Errorf("gh severity = %v, want Pass (optional)", r.Severity)
		}
		if r.InstallHint == "" {
			t.Error("gh should have install hint even though optional")
		}
	} else {
		t.Error("gh check not found")
	}

	// replicator: found
	if r, ok := results["replicator"]; ok {
		if r.Severity != Pass {
			t.Errorf("replicator severity = %v, want Pass", r.Severity)
		}
	} else {
		t.Error("replicator check not found")
	}
}

func TestCheckCoreTools_UnparseableGoVersion(t *testing.T) {
	opts := &Options{
		TargetDir: t.TempDir(),
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/go/bin/go",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version": "go version devel go1.25-abcdef Sat Mar 1 00:00:00 2026 +0000 darwin/arm64",
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     stubReadFile(nil),
	}

	env := DetectedEnvironment{Managers: []ManagerInfo{}}
	group := checkCoreTools(opts, env)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	// The devel version "1.25-abcdef" should still parse the major.minor
	// but the patch part is non-numeric. parseGoVersion should extract
	// "1.25-abcdef" which starts with a digit. checkGoVersion should
	// parse 1.25 which is >= 1.24, so it should pass.
	// But if parsing fails completely, it should warn.
	r := results["go"]
	if r.Severity == Fail {
		t.Errorf("go with devel version should not fail, got severity=%v message=%q", r.Severity, r.Message)
	}
}

func TestCheckScaffoldedFiles(t *testing.T) {
	dir := t.TempDir()

	// Create scaffolded directories with files.
	createFile(t, dir, ".opencode/agents/agent1.md", "---\ndescription: test\n---\n# Agent")
	createFile(t, dir, ".opencode/agents/agent2.md", "---\ndescription: test\n---\n# Agent")
	createFile(t, dir, ".opencode/command/cmd1.md", "# Command")
	createFile(t, dir, ".opencode/uf/packs/go.md", "# Go pack")
	createFile(t, dir, ".specify/config.yaml", "# config")
	createFile(t, dir, "AGENTS.md", "# Agents")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkScaffoldedFiles(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results[".opencode/agents/"]; r.Severity != Pass {
		t.Errorf("agents severity = %v, want Pass", r.Severity)
	}
	if r := results[".opencode/command/"]; r.Severity != Pass {
		t.Errorf("command severity = %v, want Pass", r.Severity)
	}
	if r := results[".specify/"]; r.Severity != Pass {
		t.Errorf("specify severity = %v, want Pass", r.Severity)
	}
	// AGENTS.md check moved to checkAgentContext; verify it is
	// NOT in this group.
	if _, ok := results["AGENTS.md"]; ok {
		t.Error("AGENTS.md should not be checked in Scaffolded Files (moved to Agent Context)")
	}
}

func TestCheckScaffoldedFiles_Missing(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkScaffoldedFiles(opts)

	for _, r := range group.Results {
		if r.Severity == Pass {
			t.Errorf("%s should not pass in empty dir", r.Name)
		}
		if r.InstallHint == "" {
			t.Errorf("%s should have install hint", r.Name)
		}
	}
}

func TestCheckHeroAvailability(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".opencode", "agents")

	// Create agent files.
	createFile(t, dir, ".opencode/agents/muti-mind-po.md", "# agent")
	createFile(t, dir, ".opencode/agents/cobalt-crush-dev.md", "# agent")
	createFile(t, dir, ".opencode/agents/divisor-guard.md", "# agent")
	createFile(t, dir, ".opencode/agents/divisor-architect.md", "# agent")
	createFile(t, dir, ".opencode/agents/divisor-adversary.md", "# agent")
	createFile(t, dir, ".opencode/agents/divisor-sre.md", "# agent")
	createFile(t, dir, ".opencode/agents/divisor-testing.md", "# agent")
	createFile(t, dir, ".opencode/agents/mx-f-coach.md", "# agent")

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"gaze": true, "mxf": true}),
	}

	group := checkHeroAvailability(opts)

	if group.Name != "Hero Availability" {
		t.Errorf("group name = %q, want Hero Availability", group.Name)
	}

	// All 5 heroes should be available.
	for _, r := range group.Results {
		if r.Severity != Pass {
			t.Errorf("hero %q severity = %v, want Pass", r.Name, r.Severity)
		}
	}

	// Check Divisor shows persona count.
	for _, r := range group.Results {
		if strings.Contains(r.Name, "Divisor") {
			if !strings.Contains(r.Message, "+4 personas") {
				t.Errorf("Divisor message = %q, want '+4 personas'", r.Message)
			}
		}
	}

	_ = agentDir // used via dir
}

func TestCheckMCPConfig(t *testing.T) {
	dir := t.TempDir()

	ocJSON := `{
  "mcpServers": {
    "dewey": {
      "command": "dewey",
      "args": ["serve", "--vault", "."]
    }
  }
}`
	createFile(t, dir, "opencode.json", ocJSON)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"dewey": true}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["opencode.json"]; r.Severity != Pass {
		t.Errorf("opencode.json severity = %v, want Pass", r.Severity)
	}
	if r := results["dewey"]; r.Severity != Pass {
		t.Errorf("dewey severity = %v, want Pass", r.Severity)
	}
}

func TestCheckMCPConfig_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", "{invalid")

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	if len(group.Results) == 0 {
		t.Fatal("expected at least one result")
	}

	r := group.Results[0]
	if r.Severity != Warn {
		t.Errorf("severity = %v, want Warn for malformed JSON", r.Severity)
	}
	if !strings.Contains(r.Message, "could not be parsed") {
		t.Errorf("message = %q, want 'could not be parsed'", r.Message)
	}
}

func TestCheckAgentIntegrity(t *testing.T) {
	dir := t.TempDir()

	// Valid agent with frontmatter.
	createFile(t, dir, ".opencode/agents/valid.md", "---\ndescription: A valid agent\n---\n# Agent")
	// Agent with missing description.
	createFile(t, dir, ".opencode/agents/invalid.md", "---\ntitle: No description\n---\n# Agent")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkAgentSkillIntegrity(opts)

	// Should have agent result with warning about invalid.
	found := false
	for _, r := range group.Results {
		if strings.Contains(r.Name, "agents validated") {
			found = true
			if r.Severity != Warn {
				t.Errorf("agent integrity severity = %v, want Warn (1 invalid)", r.Severity)
			}
			if !strings.Contains(r.Message, "missing description") {
				t.Errorf("message = %q, want 'missing description'", r.Message)
			}
		}
	}
	if !found {
		t.Error("agent validation result not found")
	}
}

func TestCheckSkillIntegrity(t *testing.T) {
	dir := t.TempDir()

	// Valid skill.
	createFile(t, dir, ".opencode/skill/my-skill/SKILL.md",
		"---\nname: my-skill\ndescription: A valid skill\n---\n# Skill")
	// Invalid skill: name doesn't match directory.
	createFile(t, dir, ".opencode/skill/other-skill/SKILL.md",
		"---\nname: wrong-name\ndescription: Mismatched name\n---\n# Skill")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkAgentSkillIntegrity(opts)

	// Find skill results.
	var skillResults []CheckResult
	for _, r := range group.Results {
		if r.Name == "1 skill validated" || r.Name == "other-skill" || r.Name == "my-skill" {
			skillResults = append(skillResults, r)
		}
	}

	if len(skillResults) < 2 {
		t.Fatalf("expected at least 2 skill results, got %d", len(skillResults))
	}

	// One should pass, one should warn.
	passCount := 0
	warnCount := 0
	for _, r := range skillResults {
		if r.Severity == Pass {
			passCount++
		}
		if r.Severity == Warn {
			warnCount++
			if !strings.Contains(r.Message, "does not match directory") {
				t.Errorf("warn message = %q, want 'does not match directory'", r.Message)
			}
		}
	}
	if passCount == 0 {
		t.Error("expected at least one passing skill")
	}
	if warnCount == 0 {
		t.Error("expected at least one warning skill")
	}
}

func TestDoctorRun(t *testing.T) {
	dir := t.TempDir()

	// Create minimal scaffolded files.
	createFile(t, dir, ".opencode/agents/test.md", "---\ndescription: test\n---\n# Agent")
	createFile(t, dir, ".opencode/command/test.md", "# Command")
	createFile(t, dir, ".opencode/uf/packs/go.md", "# Go")
	createFile(t, dir, ".specify/config.yaml", "# config")
	createFile(t, dir, "AGENTS.md", "# Agents")
	createFile(t, dir, "opencode.json", `{"mcpServers":{}}`)

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Format:    "text",
		Stdout:    &buf,
		LookPath: stubLookPath(map[string]string{
			"go":       "/usr/local/go/bin/go",
			"opencode": "/usr/local/bin/opencode",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version": "go version go1.24.3 darwin/arm64",
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	report, err := Run(opts)
	if report == nil {
		t.Fatal("Run returned nil report")
	}
	// Some required tools are missing (e.g., gaze, mxf), so expect
	// failures (swarm not found causes Fail in Swarm Plugin group).
	if err == nil {
		t.Log("Run returned nil error (all checks passed or only warnings)")
	}

	// Verify 9 groups in correct order.
	expectedGroups := []string{
		"Detected Environment",
		"Core Tools",
		"Replicator",
		"Dewey Knowledge Layer",
		"Agent Context",
		"Scaffolded Files",
		"Hero Availability",
		"MCP Server Config",
		"Agent/Skill Integrity",
	}

	if len(report.Groups) != len(expectedGroups) {
		t.Fatalf("expected %d groups, got %d", len(expectedGroups), len(report.Groups))
	}

	for i, name := range expectedGroups {
		if report.Groups[i].Name != name {
			t.Errorf("group[%d] = %q, want %q", i, report.Groups[i].Name, name)
		}
	}

	// Verify summary counts are consistent.
	totalFromGroups := 0
	for _, g := range report.Groups {
		totalFromGroups += len(g.Results)
	}
	if report.Summary.Total != totalFromGroups {
		t.Errorf("summary total = %d, counted %d", report.Summary.Total, totalFromGroups)
	}
	if report.Summary.Passed+report.Summary.Warned+report.Summary.Failed != report.Summary.Total {
		t.Error("summary counts don't add up to total")
	}
}

func TestDoctorRun_AllPass(t *testing.T) {
	dir := t.TempDir()

	// Create all scaffolded files.
	createFile(t, dir, ".opencode/agents/test.md", "---\ndescription: test\n---\n# Agent")
	createFile(t, dir, ".opencode/command/test.md", "# Command")
	createFile(t, dir, ".opencode/uf/packs/go.md", "# Go")
	createFile(t, dir, ".specify/config.yaml", "# config")
	createFile(t, dir, "AGENTS.md", completeAGENTSmd())
	createFile(t, dir, "CLAUDE.md", "# Claude\n@AGENTS.md\n")
	createFile(t, dir, ".cursorrules", "Read AGENTS.md for conventions.\n")
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0755); err != nil {
		t.Fatalf("mkdir .uf/replicator: %v", err)
	}

	replicatorOut := "✓ OK\n"

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Format:    "text",
		Stdout:    &buf,
		LookPath: stubLookPath(map[string]string{
			"go":         "/usr/local/go/bin/go",
			"opencode":   "/usr/local/bin/opencode",
			"gaze":       "/usr/local/bin/gaze",
			"mxf":        "/usr/local/bin/mxf",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version":        "go version go1.24.3 darwin/arm64",
				"replicator doctor": replicatorOut,
			},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{
				"replicator doctor": replicatorOut,
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	report, err := Run(opts)
	if report == nil {
		t.Fatal("Run returned nil report")
	}
	if err != nil {
		t.Errorf("Run returned error when all checks should pass: %v", err)
	}
	if report.Summary.Failed != 0 {
		t.Errorf("expected 0 failures, got %d", report.Summary.Failed)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name    string
		results []CheckResult
		wantErr bool
	}{
		{
			name: "all pass",
			results: []CheckResult{
				{Name: "a", Severity: Pass, Message: "ok"},
				{Name: "b", Severity: Pass, Message: "ok"},
			},
			wantErr: false,
		},
		{
			name: "warnings only",
			results: []CheckResult{
				{Name: "a", Severity: Pass, Message: "ok"},
				{Name: "b", Severity: Warn, Message: "warn"},
			},
			wantErr: false,
		},
		{
			name: "any fail",
			results: []CheckResult{
				{Name: "a", Severity: Pass, Message: "ok"},
				{Name: "b", Severity: Fail, Message: "fail"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := []CheckGroup{{Name: "test", Results: tt.results}}
			summary := computeSummary(groups)

			hasErr := summary.Failed > 0
			if hasErr != tt.wantErr {
				t.Errorf("hasErr = %v, want %v", hasErr, tt.wantErr)
			}
		})
	}
}

func TestDoctorRun_NonGitDir(t *testing.T) {
	dir := t.TempDir()

	var buf bytes.Buffer
	opts := Options{
		TargetDir:    dir,
		Format:       "text",
		Stdout:       &buf,
		LookPath:     stubLookPathSimple(map[string]bool{}),
		ExecCmd:      stubExecCmd(nil, nil),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	report, err := Run(opts)
	if report == nil {
		t.Fatal("Run returned nil report in non-git dir")
	}
	// Expect error since required tools are missing.
	if err == nil {
		t.Log("Run returned nil error in non-git dir")
	}

	// All checks should still execute.
	if len(report.Groups) != 9 {
		t.Errorf("expected 9 groups, got %d", len(report.Groups))
	}
}

func TestDoctorRun_DirFlag(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "AGENTS.md", "# Agents")

	var buf bytes.Buffer
	opts := Options{
		TargetDir:    dir,
		Format:       "text",
		Stdout:       &buf,
		LookPath:     stubLookPathSimple(map[string]bool{}),
		ExecCmd:      stubExecCmd(nil, nil),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	report, err := Run(opts)
	if report == nil {
		t.Fatal("Run returned nil report")
	}
	// Expect error since required tools are missing.
	if err == nil {
		t.Log("Run returned nil error with --dir flag")
	}

	// AGENTS.md should be found in the specified dir.
	for _, g := range report.Groups {
		if g.Name == "Scaffolded Files" {
			for _, r := range g.Results {
				if r.Name == "AGENTS.md" && r.Severity != Pass {
					t.Errorf("AGENTS.md should pass when present in target dir")
				}
			}
		}
	}
}

// --- Replicator tests ---

func TestCheckReplicator_NotInstalled(t *testing.T) {
	opts := &Options{
		TargetDir:      t.TempDir(),
		LookPath:       stubLookPathSimple(map[string]bool{}),
		ExecCmd:        stubExecCmd(nil, nil),
		ExecCmdTimeout: stubExecCmdTimeout(nil, nil),
		EvalSymlinks:   stubEvalSymlinks(nil),
		Getenv:         stubGetenv(map[string]string{}),
		ReadFile:       os.ReadFile,
	}

	group := checkReplicator(opts)

	if len(group.Results) == 0 {
		t.Fatal("expected at least one result")
	}

	r := group.Results[0]
	if r.Severity != Warn {
		t.Errorf("replicator severity = %v, want Warn", r.Severity)
	}
	if !strings.Contains(r.InstallHint, "brew install unbound-force/tap/replicator") {
		t.Errorf("install hint = %q, want brew install command", r.InstallHint)
	}
}

func TestCheckReplicator_AllPass(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/replicator/ dir and opencode.json with mcp.replicator entry.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0755); err != nil {
		t.Fatalf("mkdir .uf/replicator: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)

	replicatorDoctorOutput := "✓ All checks passed\n"

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd: stubExecCmd(
			map[string]string{"replicator doctor": replicatorDoctorOutput},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"replicator doctor": replicatorDoctorOutput},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkReplicator(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["replicator"]; r.Severity != Pass {
		t.Errorf("replicator severity = %v, want Pass", r.Severity)
	}
	if r := results[".uf/replicator/"]; r.Severity != Pass {
		t.Errorf(".uf/replicator/ severity = %v, want Pass", r.Severity)
	}
	if r := results["MCP config"]; r.Severity != Pass {
		t.Errorf("MCP config severity = %v, want Pass", r.Severity)
	}

	// Verify embed contains replicator doctor output.
	if !strings.Contains(group.Embed, "All checks passed") {
		t.Error("embed should contain replicator doctor output")
	}
}

func TestCheckReplicator_Timeout(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd: stubExecCmd(
			nil,
			map[string]error{"replicator doctor": fmt.Errorf("context deadline exceeded: timed out")},
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			nil,
			map[string]error{"replicator doctor": fmt.Errorf("context deadline exceeded: timed out")},
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkReplicator(opts)

	for _, r := range group.Results {
		if r.Name == "replicator doctor" {
			if r.Severity != Warn {
				t.Errorf("replicator doctor severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.Message, "timed out") {
				t.Errorf("message = %q, want 'timed out'", r.Message)
			}
			return
		}
	}
	t.Error("replicator doctor result not found")
}

func TestCheckReplicator_HiveMissing(t *testing.T) {
	dir := t.TempDir()
	// No .uf/replicator/ directory.
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd: stubExecCmd(
			map[string]string{"replicator doctor": "ok"},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"replicator doctor": "ok"},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkReplicator(opts)

	for _, r := range group.Results {
		if r.Name == ".uf/replicator/" {
			if r.Severity != Warn {
				t.Errorf(".uf/replicator/ severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.InstallHint, "uf init") {
				t.Errorf("install hint = %q, want 'uf init'", r.InstallHint)
			}
			return
		}
	}
	t.Error(".uf/replicator/ result not found")
}

func TestCheckReplicator_MCPMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0755); err != nil {
		t.Fatalf("mkdir .uf/replicator: %v", err)
	}
	// opencode.json exists but no mcp.replicator.
	createFile(t, dir, "opencode.json", `{"mcp":{"dewey":{"type":"local"}}}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd: stubExecCmd(
			map[string]string{"replicator doctor": "ok"},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"replicator doctor": "ok"},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkReplicator(opts)

	for _, r := range group.Results {
		if r.Name == "MCP config" {
			if r.Severity != Warn {
				t.Errorf("MCP config severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.InstallHint, "uf init") {
				t.Errorf("install hint = %q, want 'uf init'", r.InstallHint)
			}
			return
		}
	}
	t.Error("MCP config result not found")
}

// --- Phase 6: User Story 4 tests (T063-T066) ---

func TestFormatText_NoColors(t *testing.T) {
	report := &Report{
		Groups: []CheckGroup{
			{
				Name: "Core Tools",
				Results: []CheckResult{
					{Name: "go", Severity: Pass, Message: "1.24.3"},
					{Name: "opencode", Severity: Fail, Message: "not found", InstallHint: "brew install opencode"},
					{Name: "gaze", Severity: Warn, Message: "outdated"},
				},
			},
		},
		Summary: Summary{Total: 3, Passed: 1, Warned: 1, Failed: 1},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText: %v", err)
	}

	output := buf.String()

	// When writing to a buffer (not a TTY), lipgloss should detect
	// no color support. Check for plain text indicators.
	if !strings.Contains(output, "[PASS]") && !strings.Contains(output, "✅") {
		t.Error("expected pass indicator in output")
	}
	if !strings.Contains(output, "Unbound Force Doctor") {
		t.Error("expected header in output")
	}
	if !strings.Contains(output, "🩺") {
		t.Error("expected stethoscope emoji in header")
	}
	if !strings.Contains(output, "passed") && !strings.Contains(output, "failed") {
		t.Error("expected summary counters in output")
	}
}

func TestFormatText_ReplicatorDoctorEmbed(t *testing.T) {
	report := &Report{
		Groups: []CheckGroup{
			{
				Name: "Replicator",
				Results: []CheckResult{
					{Name: "replicator", Severity: Pass, Message: "installed"},
				},
				Embed: "✓ All checks passed\n✓ Dependencies: all installed\n",
			},
		},
		Summary: Summary{Total: 1, Passed: 1},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "All checks passed") {
		t.Error("expected embedded replicator doctor output")
	}
	if !strings.Contains(output, "─") {
		t.Error("expected separator lines around embed")
	}
}

func TestFormatText_InstallHints(t *testing.T) {
	report := &Report{
		Groups: []CheckGroup{
			{
				Name: "Core Tools",
				Results: []CheckResult{
					{
						Name:        "opencode",
						Severity:    Fail,
						Message:     "not found",
						InstallHint: "brew install anomalyco/tap/opencode",
						InstallURL:  "https://opencode.ai/docs",
					},
				},
			},
		},
		Summary: Summary{Total: 1, Failed: 1},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Fix: brew install anomalyco/tap/opencode") {
		t.Error("expected fix hint in output")
	}
	if !strings.Contains(output, "Docs: https://opencode.ai/docs") {
		t.Error("expected docs URL in output")
	}
}

// --- Phase 7: User Story 5 tests (T071-T072) ---

func TestFormatJSON(t *testing.T) {
	report := &Report{
		Environment: DetectedEnvironment{
			Managers: []ManagerInfo{
				{Kind: ManagerGoenv, Path: "/opt/homebrew/bin/goenv", Manages: []string{"go"}},
			},
			Platform: "darwin/arm64",
		},
		Groups: []CheckGroup{
			{
				Name: "Core Tools",
				Results: []CheckResult{
					{Name: "go", Severity: Pass, Message: "1.24.3 via goenv", Detail: "/home/.goenv/shims/go"},
					{Name: "opencode", Severity: Fail, Message: "not found", InstallHint: "brew install opencode", InstallURL: "https://opencode.ai/docs"},
				},
			},
			{
				Name: "Replicator",
				Results: []CheckResult{
					{Name: "replicator", Severity: Pass, Message: "installed"},
				},
				Embed: "✓ All checks passed\n",
			},
		},
		Summary: Summary{Total: 3, Passed: 2, Warned: 0, Failed: 1},
	}

	var buf bytes.Buffer
	if err := FormatJSON(report, &buf); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	// Verify valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify snake_case field names.
	output := buf.String()
	if !strings.Contains(output, `"install_hint"`) {
		t.Error("expected snake_case install_hint")
	}
	if !strings.Contains(output, `"install_url"`) {
		t.Error("expected snake_case install_url")
	}

	// Verify severity as lowercase strings.
	if !strings.Contains(output, `"severity": "pass"`) {
		t.Error("expected lowercase severity 'pass'")
	}
	if !strings.Contains(output, `"severity": "fail"`) {
		t.Error("expected lowercase severity 'fail'")
	}

	// Verify all top-level fields present.
	if _, ok := parsed["environment"]; !ok {
		t.Error("missing environment field")
	}
	if _, ok := parsed["groups"]; !ok {
		t.Error("missing groups field")
	}
	if _, ok := parsed["summary"]; !ok {
		t.Error("missing summary field")
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		sev  Severity
		want string
	}{
		{Pass, "pass"},
		{Warn, "warn"},
		{Fail, "fail"},
		{Severity(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.sev.String(); got != tt.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tt.sev, got, tt.want)
		}
	}
}

func TestSeverity_MarshalJSON_Invalid(t *testing.T) {
	_, err := Severity(99).MarshalJSON()
	if err == nil {
		t.Error("expected error for invalid severity")
	}
}

func TestHasManager(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerHomebrew, Path: "/opt/homebrew/bin/brew", Manages: []string{"packages"}},
		},
	}
	if !HasManager(env, ManagerHomebrew) {
		t.Error("expected HasManager to find Homebrew")
	}
	if HasManager(env, ManagerGoenv) {
		t.Error("expected HasManager not to find goenv")
	}
}

func TestDetectEnvironment_MultipleManagers(t *testing.T) {
	opts := &Options{
		LookPath: stubLookPath(map[string]string{
			"goenv": "/opt/homebrew/bin/goenv",
			"brew":  "/opt/homebrew/bin/brew",
			"mise":  "/usr/local/bin/mise",
			"bun":   "/home/user/.bun/bin/bun",
			"fnm":   "/usr/local/bin/fnm",
		}),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv: stubGetenv(map[string]string{
			"NVM_DIR":    "/home/user/.nvm",
			"PYENV_ROOT": "/home/user/.pyenv",
		}),
	}

	env := DetectEnvironment(opts)

	// Should detect all managers.
	kinds := make(map[ManagerKind]bool)
	for _, m := range env.Managers {
		kinds[m.Kind] = true
	}

	for _, expected := range []ManagerKind{ManagerGoenv, ManagerPyenv, ManagerNvm, ManagerFnm, ManagerMise, ManagerBun, ManagerHomebrew} {
		if !kinds[expected] {
			t.Errorf("expected %s to be detected", expected)
		}
	}
}

func TestCheckDetectedEnvironment_Empty(t *testing.T) {
	env := DetectedEnvironment{Managers: []ManagerInfo{}}
	group := checkDetectedEnvironment(env)

	if len(group.Results) != 1 {
		t.Fatalf("expected 1 result for empty env, got %d", len(group.Results))
	}
	if group.Results[0].Name != "none" {
		t.Errorf("expected 'none' result, got %q", group.Results[0].Name)
	}
}

func TestCheckDetectedEnvironment_WithManagers(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerGoenv, Path: "/opt/homebrew/bin/goenv", Manages: []string{"go"}},
			{Kind: ManagerNvm, Path: "/home/user/.nvm", Manages: []string{"node"}},
		},
	}
	group := checkDetectedEnvironment(env)

	if len(group.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(group.Results))
	}
	for _, r := range group.Results {
		if r.Severity != Pass {
			t.Errorf("detected env result %q should be Pass", r.Name)
		}
	}
}

func TestParseGoVersion_Invalid(t *testing.T) {
	_, err := parseGoVersion("not a version string")
	if err == nil {
		t.Error("expected error for invalid go version")
	}
}

func TestParseNodeVersion_Invalid(t *testing.T) {
	_, err := parseNodeVersion("not a version")
	if err == nil {
		t.Error("expected error for invalid node version")
	}
}

func TestCheckGoVersion(t *testing.T) {
	tests := []struct {
		version string
		min     string
		want    bool
	}{
		{"1.24.3", "1.24", true},
		{"1.23.0", "1.24", false},
		{"2.0.0", "1.24", true},
		{"1.24", "1.24", true},
	}
	for _, tt := range tests {
		got := checkGoVersion(tt.version, tt.min)
		if got != tt.want {
			t.Errorf("checkGoVersion(%q, %q) = %v, want %v", tt.version, tt.min, got, tt.want)
		}
	}
}

func TestCheckNodeVersion(t *testing.T) {
	tests := []struct {
		version string
		min     string
		want    bool
	}{
		{"22.15.0", "18", true},
		{"16.0.0", "18", false},
		{"18.0.0", "18", true},
	}
	for _, tt := range tests {
		got := checkNodeVersion(tt.version, tt.min)
		if got != tt.want {
			t.Errorf("checkNodeVersion(%q, %q) = %v, want %v", tt.version, tt.min, got, tt.want)
		}
	}
}

func TestInstallHint_ReplicatorHomebrew(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerHomebrew, Path: "/opt/homebrew/bin/brew", Manages: []string{"packages"}},
		},
	}
	hint := installHint("replicator", env)
	if !strings.Contains(hint, "brew install unbound-force/tap/replicator") {
		t.Errorf("hint = %q, want brew install command", hint)
	}
}

func TestInstallURL(t *testing.T) {
	if url := installURL("opencode"); url == "" {
		t.Error("expected non-empty URL for opencode")
	}
	if url := installURL("unknown-tool"); url != "" {
		t.Errorf("expected empty URL for unknown tool, got %q", url)
	}
}

func TestFormatIndicator_AllCases(t *testing.T) {
	// Test with no color (plain text).
	tests := []struct {
		result CheckResult
		want   string
	}{
		{CheckResult{Severity: Pass}, "[PASS]"},
		{CheckResult{Severity: Pass, InstallHint: "hint"}, "[INFO]"},
		{CheckResult{Severity: Warn}, "[WARN]"},
		{CheckResult{Severity: Fail}, "[FAIL]"},
	}

	for _, tt := range tests {
		got := formatIndicator(tt.result, false, lipgloss.Style{}, lipgloss.Style{}, lipgloss.Style{}, lipgloss.Style{})
		if got != tt.want {
			t.Errorf("formatIndicator(%v, false) = %q, want %q", tt.result.Severity, got, tt.want)
		}
	}
}

func TestParseFrontmatter_NoDelimiter(t *testing.T) {
	_, err := parseFrontmatter([]byte("no frontmatter here"))
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseFrontmatter_NoClosing(t *testing.T) {
	_, err := parseFrontmatter([]byte("---\nkey: value\nno closing"))
	if err == nil {
		t.Error("expected error for missing closing delimiter")
	}
}

func TestParseFrontmatter_InvalidYAML(t *testing.T) {
	_, err := parseFrontmatter([]byte("---\n: invalid: yaml: [[\n---\n"))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestManagerDescription(t *testing.T) {
	tests := []struct {
		kind ManagerKind
		want string
	}{
		{ManagerGoenv, "Go version manager"},
		{ManagerPyenv, "Python version manager"},
		{ManagerNvm, "Node version manager"},
		{ManagerFnm, "Fast Node manager"},
		{ManagerMise, "Polyglot version manager"},
		{ManagerBun, "Bun JavaScript runtime"},
		{ManagerHomebrew, "Package manager"},
		{ManagerKind("custom"), "custom"},
	}
	for _, tt := range tests {
		got := managerDescription(tt.kind)
		if got != tt.want {
			t.Errorf("managerDescription(%q) = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestHomebrewInstallCmd(t *testing.T) {
	tests := []struct {
		tool string
		want string
	}{
		{"go", "brew install go"},
		{"opencode", "brew install anomalyco/tap/opencode"},
		{"gaze", "brew install unbound-force/tap/gaze"},
		{"mxf", "brew install unbound-force/tap/unbound-force (mxf is bundled)"},
		{"dewey", "brew install unbound-force/tap/dewey"},
		{"node", "brew install node"},
		{"gh", "brew install gh"},
		{"replicator", "brew install unbound-force/tap/replicator"},
		{"custom", "brew install custom"},
	}
	for _, tt := range tests {
		got := homebrewInstallCmd(tt.tool)
		if got != tt.want {
			t.Errorf("homebrewInstallCmd(%q) = %q, want %q", tt.tool, got, tt.want)
		}
	}
}

func TestGenericInstallCmd(t *testing.T) {
	tests := []struct {
		tool string
	}{
		{"opencode"},
		{"gaze"},
		{"go"},
		{"node"},
		{"replicator"},
		{"gh"},
		{"unknown"},
	}
	for _, tt := range tests {
		got := genericInstallCmd(tt.tool)
		if got == "" {
			t.Errorf("genericInstallCmd(%q) returned empty string", tt.tool)
		}
	}
}

func TestToolCategory(t *testing.T) {
	tests := []struct {
		tool string
		want string
	}{
		{"go", "go"},
		{"node", "node"},
		{"npm", "node"},
		{"python", "python"},
		{"python3", "python"},
		{"gaze", "packages"},
	}
	for _, tt := range tests {
		got := toolCategory(tt.tool)
		if got != tt.want {
			t.Errorf("toolCategory(%q) = %q, want %q", tt.tool, got, tt.want)
		}
	}
}

func TestManagerInstallCmd(t *testing.T) {
	tests := []struct {
		tool    string
		manager ManagerKind
		want    string
	}{
		{"go", ManagerGoenv, "goenv install 1.24.3 && goenv global 1.24.3"},
		{"node", ManagerNvm, "nvm install 22"},
		{"node", ManagerFnm, "fnm install 22"},
		{"go", ManagerMise, "mise install go@1.24"},
		{"node", ManagerMise, "mise install node@22"},
		{"replicator", ManagerHomebrew, "brew install unbound-force/tap/replicator"},
	}
	for _, tt := range tests {
		got := managerInstallCmd(tt.tool, tt.manager)
		if got != tt.want {
			t.Errorf("managerInstallCmd(%q, %q) = %q, want %q", tt.tool, tt.manager, got, tt.want)
		}
	}
}

func TestCheckCoreTools_GoVersionTooOld(t *testing.T) {
	opts := &Options{
		TargetDir: t.TempDir(),
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/go/bin/go",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version": "go version go1.22.5 darwin/arm64",
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     stubReadFile(nil),
	}

	env := DetectedEnvironment{Managers: []ManagerInfo{}}
	group := checkCoreTools(opts, env)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	r := results["go"]
	if r.Severity != Fail {
		t.Errorf("go 1.22 should fail version check, got severity=%v", r.Severity)
	}
	if !strings.Contains(r.Message, "requires >= 1.24") {
		t.Errorf("message = %q, want 'requires >= 1.24'", r.Message)
	}
}

func TestCheckCoreTools_GoExecFails(t *testing.T) {
	opts := &Options{
		TargetDir: t.TempDir(),
		LookPath: stubLookPath(map[string]string{
			"go": "/usr/local/go/bin/go",
		}),
		ExecCmd: stubExecCmd(
			nil,
			map[string]error{"go version": fmt.Errorf("exec failed")},
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     stubReadFile(nil),
	}

	env := DetectedEnvironment{Managers: []ManagerInfo{}}
	group := checkCoreTools(opts, env)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	r := results["go"]
	if r.Severity != Warn {
		t.Errorf("go exec failure should warn, got severity=%v", r.Severity)
	}
	if !strings.Contains(r.Message, "could not be verified") {
		t.Errorf("message = %q, want 'could not be verified'", r.Message)
	}
}

func TestValidateAgents_ReadError(t *testing.T) {
	dir := t.TempDir()
	// Create agents dir with a file that can't be read (simulate via ReadFile stub).
	agentDir := filepath.Join(dir, ".opencode", "agents")
	createFile(t, dir, ".opencode/agents/test.md", "---\ndescription: test\n---\n# Agent")

	opts := &Options{
		TargetDir: dir,
		ReadFile: func(path string) ([]byte, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	result := validateAgents(agentDir, opts)
	if result.Severity != Warn {
		t.Errorf("read error should produce Warn, got %v", result.Severity)
	}
}

func TestValidateSkills_NoSkillFile(t *testing.T) {
	dir := t.TempDir()
	// Create skill directory without SKILL.md.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "skill", "empty-skill"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	results := validateSkills(filepath.Join(dir, ".opencode", "skill"), opts)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Severity != Warn {
		t.Errorf("missing SKILL.md should produce Warn, got %v", results[0].Severity)
	}
}

func TestValidateSkills_InvalidFrontmatter(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skill/bad-skill/SKILL.md", "---\n: invalid: yaml\n---\n# Skill")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	results := validateSkills(filepath.Join(dir, ".opencode", "skill"), opts)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Severity != Warn {
		t.Errorf("invalid frontmatter should produce Warn, got %v", results[0].Severity)
	}
}

func TestValidateSkills_MissingFields(t *testing.T) {
	dir := t.TempDir()
	// Skill with no name or description.
	createFile(t, dir, ".opencode/skill/no-fields/SKILL.md", "---\ntitle: something\n---\n# Skill")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	results := validateSkills(filepath.Join(dir, ".opencode", "skill"), opts)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if results[0].Severity != Warn {
		t.Errorf("missing fields should produce Warn, got %v", results[0].Severity)
	}
	if !strings.Contains(results[0].Message, "missing name") {
		t.Errorf("message = %q, want 'missing name'", results[0].Message)
	}
}

func TestValidateSkills_InvalidNamePattern(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/skill/Bad_Name/SKILL.md", "---\nname: Bad_Name\ndescription: test\n---\n# Skill")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	results := validateSkills(filepath.Join(dir, ".opencode", "skill"), opts)
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}
	if !strings.Contains(results[0].Message, "does not match") {
		t.Errorf("message = %q, want pattern mismatch", results[0].Message)
	}
}

func TestCheckMCPConfig_MissingServerBinary(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", `{
  "mcpServers": {
    "test-server": {
      "command": "missing-binary",
      "args": []
    }
  }
}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	found := false
	for _, r := range group.Results {
		if r.Name == "test-server" {
			found = true
			if r.Severity != Warn {
				t.Errorf("missing binary should produce Warn, got %v", r.Severity)
			}
		}
	}
	if !found {
		t.Error("test-server result not found")
	}
}

// --- Dewey Knowledge Layer tests ---

func TestCheckDewey_AllPresent(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ workspace directory.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
		t.Fatalf("mkdir .uf/dewey: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME                    ID              SIZE      MODIFIED\ngranite-embedding:30m   abc123          63 MB     2 days ago\n",
			},
			nil,
		),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	if group.Name != "Dewey Knowledge Layer" {
		t.Errorf("group name = %q, want Dewey Knowledge Layer", group.Name)
	}

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["dewey binary"]; r.Severity != Pass {
		t.Errorf("dewey binary severity = %v, want Pass", r.Severity)
	}
	if r := results["embedding model"]; r.Severity != Pass {
		t.Errorf("embedding model severity = %v, want Pass", r.Severity)
	}
	if r := results["embedding capability"]; r.Severity != Pass {
		t.Errorf("embedding capability severity = %v, want Pass", r.Severity)
	}
	if r := results["workspace"]; r.Severity != Pass {
		t.Errorf("workspace severity = %v, want Pass", r.Severity)
	}
}

func TestCheckDewey_BinaryMissing(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir:  dir,
		LookPath:   stubLookPathSimple(map[string]bool{}),
		ExecCmd:    stubExecCmd(nil, nil),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	if len(group.Results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(group.Results))
	}

	// Binary not found -- severity MUST be Pass (Dewey is optional per Constitution Principle II).
	r := group.Results[0]
	if r.Name != "dewey binary" {
		t.Errorf("first result name = %q, want dewey binary", r.Name)
	}
	if r.Severity != Pass {
		t.Errorf("dewey binary severity = %v, want Pass (optional tool)", r.Severity)
	}
	if !strings.Contains(r.InstallHint, "brew install unbound-force/tap/dewey") {
		t.Errorf("install hint = %q, want brew install command", r.InstallHint)
	}

	// Remaining checks skipped -- severity MUST be Pass (skipped, not failed).
	if !strings.Contains(group.Results[1].Message, "skipped") {
		t.Errorf("embedding model should be skipped, got %q", group.Results[1].Message)
	}
	if group.Results[1].Severity != Pass {
		t.Errorf("embedding model severity = %v, want Pass (skipped)", group.Results[1].Severity)
	}
	if !strings.Contains(group.Results[2].Message, "skipped") {
		t.Errorf("embedding capability should be skipped, got %q", group.Results[2].Message)
	}
	if group.Results[2].Severity != Pass {
		t.Errorf("embedding capability severity = %v, want Pass (skipped)", group.Results[2].Severity)
	}
	if !strings.Contains(group.Results[3].Message, "skipped") {
		t.Errorf("workspace should be skipped, got %q", group.Results[3].Message)
	}
	if group.Results[3].Severity != Pass {
		t.Errorf("workspace severity = %v, want Pass (skipped)", group.Results[3].Severity)
	}
}

func TestCheckDewey_ModelMissing(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME                    ID              SIZE      MODIFIED\nllama3:latest           abc123          4.7 GB    1 day ago\n",
			},
			nil,
		),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["embedding model"]; r.Severity != Warn {
		t.Errorf("embedding model severity = %v, want Warn", r.Severity)
	}
	if r := results["embedding model"]; !strings.Contains(r.Message, "graph-only") {
		t.Errorf("embedding model message = %q, want 'graph-only'", r.Message)
	}
}

func TestCheckDewey_WorkspaceMissing(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME                    ID              SIZE      MODIFIED\ngranite-embedding:30m   abc123          63 MB     2 days ago\n",
			},
			nil,
		),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["workspace"]; r.Severity != Warn {
		t.Errorf("workspace severity = %v, want Warn", r.Severity)
	}
	if r := results["workspace"]; !strings.Contains(r.InstallHint, "dewey init") {
		t.Errorf("workspace hint = %q, want 'dewey init'", r.InstallHint)
	}
}

func TestCheckDewey_OllamaNotAvailable(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			nil,
			map[string]error{"ollama list": fmt.Errorf("ollama not found")},
		),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return fmt.Errorf("connection refused") },
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	// Embedding model check should warn when ollama is unavailable.
	if r := results["embedding model"]; r.Severity != Warn {
		t.Errorf("embedding model severity = %v, want Warn", r.Severity)
	}
	if r := results["embedding model"]; !strings.Contains(r.Message, "not available") {
		t.Errorf("embedding model message = %q, want 'not available'", r.Message)
	}
}

func TestCheckHeroAvailability_NoneAvailable(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{}),
	}

	group := checkHeroAvailability(opts)

	for _, r := range group.Results {
		if r.Severity != Warn {
			t.Errorf("hero %q should be Warn when not available, got %v", r.Name, r.Severity)
		}
	}
}

func TestCheckHeroAvailability_BinaryOnly(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath: stubLookPath(map[string]string{
			"gaze": "/usr/local/bin/gaze",
		}),
	}

	group := checkHeroAvailability(opts)

	for _, r := range group.Results {
		if strings.Contains(r.Name, "Gaze") {
			if r.Severity != Pass {
				t.Errorf("Gaze should be Pass, got %v", r.Severity)
			}
			if !strings.Contains(r.Message, "binary") {
				t.Errorf("Gaze message = %q, want 'binary'", r.Message)
			}
		}
	}
}

func TestValidateAgents_NoAgents(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".opencode", "agents")
	// Don't create the directory.

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	result := validateAgents(agentDir, opts)
	if result.Severity != Warn {
		t.Errorf("no agents dir should produce Warn, got %v", result.Severity)
	}
}

func TestValidateAgents_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	result := validateAgents(agentDir, opts)
	if result.Severity != Warn {
		t.Errorf("empty agents dir should produce Warn, got %v", result.Severity)
	}
}

func TestCheckReplicator_DoctorFails(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", `{"mcp":{"replicator":{"type":"local","command":["replicator","serve"],"enabled":true}}}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd: stubExecCmd(
			map[string]string{"replicator doctor": "✗ Not configured\n"},
			map[string]error{"replicator doctor": fmt.Errorf("exit status 1")},
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"replicator doctor": "✗ Not configured\n"},
			map[string]error{"replicator doctor": fmt.Errorf("exit status 1")},
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkReplicator(opts)

	for _, r := range group.Results {
		if r.Name == "replicator doctor" {
			if r.Severity != Warn {
				t.Errorf("replicator doctor severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.InstallHint, "uf setup") {
				t.Errorf("install hint = %q, want 'uf setup'", r.InstallHint)
			}
			if !strings.Contains(group.Embed, "Not configured") {
				t.Error("embed should contain replicator doctor output")
			}
			return
		}
	}
	t.Error("replicator doctor result not found")
}

func TestCheckMCPConfig_NoFile(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)
	if len(group.Results) == 0 {
		t.Fatal("expected at least one result")
	}
	if group.Results[0].Severity != Warn {
		t.Errorf("missing opencode.json should produce Warn, got %v", group.Results[0].Severity)
	}
}

func TestFormatJSON_EmptyReport(t *testing.T) {
	report := &Report{
		Environment: DetectedEnvironment{
			Managers: []ManagerInfo{},
			Platform: "darwin/arm64",
		},
		Groups:  []CheckGroup{},
		Summary: Summary{},
	}

	var buf bytes.Buffer
	if err := FormatJSON(report, &buf); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	// Verify valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Verify groups is empty array, not null.
	output := buf.String()
	if strings.Contains(output, `"groups": null`) {
		t.Error("groups should be empty array, not null")
	}

	// Verify summary all zeros.
	summaryRaw, ok := parsed["summary"]
	if !ok {
		t.Fatal("missing summary field")
	}
	summary, ok := summaryRaw.(map[string]interface{})
	if !ok {
		t.Fatal("summary is not an object")
	}
	for _, key := range []string{"total", "passed", "warned", "failed"} {
		val, exists := summary[key]
		if !exists {
			t.Errorf("missing summary.%s", key)
			continue
		}
		num, numOK := val.(float64)
		if !numOK {
			t.Errorf("summary.%s is not a number", key)
			continue
		}
		if num != 0 {
			t.Errorf("summary.%s = %v, want 0", key, num)
		}
	}
}

// --- Ollama check tests ---

func TestCheckOllama_InstalledWithModel(t *testing.T) {
	dir := t.TempDir()

	ollamaListOutput := "NAME                    ID              SIZE    MODIFIED\ngranite-embedding:30m    abc123  63 MB   2 days ago\nllama3:latest            def456  4.7 GB  1 week ago\n"

	opts := &Options{
		TargetDir: dir,
		LookPath: stubLookPath(map[string]string{
			"go":       "/usr/local/bin/go",
			"opencode": "/usr/local/bin/opencode",
			"ollama":   "/usr/local/bin/ollama",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version":  "go version go1.24.3 darwin/arm64",
				"ollama list": ollamaListOutput,
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	env := DetectEnvironment(opts)
	group := checkCoreTools(opts, env)

	// Find ollama result.
	for _, r := range group.Results {
		if r.Name == "ollama" {
			if r.Severity != Pass {
				t.Errorf("ollama severity = %v, want Pass", r.Severity)
			}
			if !strings.Contains(r.Message, "granite-embedding:30m model ready") {
				t.Errorf("ollama message = %q, want 'granite-embedding:30m model ready'", r.Message)
			}
			return
		}
	}
	t.Error("ollama result not found in Core Tools")
}

func TestCheckOllama_InstalledWithoutModel(t *testing.T) {
	dir := t.TempDir()

	ollamaListOutput := "NAME             ID       SIZE    MODIFIED\nllama3:latest    def456   4.7 GB  1 week ago\n"

	opts := &Options{
		TargetDir: dir,
		LookPath: stubLookPath(map[string]string{
			"go":       "/usr/local/bin/go",
			"opencode": "/usr/local/bin/opencode",
			"ollama":   "/usr/local/bin/ollama",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version":  "go version go1.24.3 darwin/arm64",
				"ollama list": ollamaListOutput,
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	env := DetectEnvironment(opts)
	group := checkCoreTools(opts, env)

	for _, r := range group.Results {
		if r.Name == "ollama" {
			if r.Severity != Pass {
				t.Errorf("ollama severity = %v, want Pass", r.Severity)
			}
			if r.InstallHint != "ollama pull granite-embedding:30m" {
				t.Errorf("ollama install hint = %q, want 'ollama pull granite-embedding:30m'", r.InstallHint)
			}
			return
		}
	}
	t.Error("ollama result not found in Core Tools")
}

func TestCheckOllama_NotInstalled(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath: stubLookPath(map[string]string{
			"go":       "/usr/local/bin/go",
			"opencode": "/usr/local/bin/opencode",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version": "go version go1.24.3 darwin/arm64",
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	env := DetectEnvironment(opts)
	group := checkCoreTools(opts, env)

	for _, r := range group.Results {
		if r.Name == "ollama" {
			if r.Severity != Pass {
				t.Errorf("ollama severity = %v, want Pass (informational)", r.Severity)
			}
			if r.InstallHint == "" {
				t.Error("expected install hint when ollama not found")
			}
			return
		}
	}
	t.Error("ollama result not found in Core Tools")
}

// TestDoctorHints_NoBareUnboundReferences is a regression guard
// for FR-006: all doctor InstallHint fields must reference `uf`
// or `unbound-force`, not bare `unbound `.
func TestDoctorHints_NoBareUnboundReferences(t *testing.T) {
	dir := t.TempDir()

	// Create minimal scaffolded files so all check groups execute.
	createFile(t, dir, ".opencode/agents/test.md", "---\ndescription: test\n---\n# Agent")
	createFile(t, dir, ".opencode/command/test.md", "# Command")
	createFile(t, dir, ".opencode/uf/packs/go.md", "# Go")
	createFile(t, dir, ".specify/config.yaml", "# config")
	createFile(t, dir, "AGENTS.md", "# Agents")
	createFile(t, dir, "opencode.json", `{"mcpServers":{}}`)

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Format:    "text",
		Stdout:    &buf,
		LookPath: stubLookPath(map[string]string{
			"go":       "/usr/local/go/bin/go",
			"opencode": "/usr/local/bin/opencode",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version": "go version go1.24.3 darwin/arm64",
			},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	report, _ := Run(opts)
	if report == nil {
		t.Fatal("Run returned nil report")
	}

	// Check all InstallHint fields across all groups.
	for _, g := range report.Groups {
		for _, r := range g.Results {
			if r.InstallHint == "" {
				continue
			}
			// Check for bare "unbound " that is NOT "unbound-force".
			hint := r.InstallHint
			// Remove all "unbound-force" occurrences to isolate bare "unbound ".
			cleaned := strings.ReplaceAll(hint, "unbound-force", "")
			if strings.Contains(cleaned, "unbound ") || strings.Contains(cleaned, "unbound\t") {
				t.Errorf("InstallHint for %q contains bare 'unbound' reference: %q (FR-006 violation)",
					r.Name, r.InstallHint)
			}
		}
	}
}

// --- MCP Config key and command format tests (Spec 017) ---

func TestCheckMCPConfig_McpKey(t *testing.T) {
	dir := t.TempDir()

	// Uses canonical "mcp" key with array-style command.
	createFile(t, dir, "opencode.json", `{
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"dewey": true}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["opencode.json"]; r.Severity != Pass {
		t.Errorf("opencode.json severity = %v, want Pass", r.Severity)
	}
	if r, ok := results["dewey"]; !ok {
		t.Error("dewey result not found")
	} else if r.Severity != Pass {
		t.Errorf("dewey severity = %v, want Pass", r.Severity)
	}
}

func TestCheckMCPConfig_McpServersKey(t *testing.T) {
	dir := t.TempDir()

	// Uses legacy "mcpServers" key.
	createFile(t, dir, "opencode.json", `{
  "mcpServers": {
    "dewey": {
      "command": "dewey",
      "args": ["serve", "--vault", "."]
    }
  }
}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"dewey": true}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r, ok := results["dewey"]; !ok {
		t.Error("dewey result not found — legacy mcpServers fallback failed")
	} else if r.Severity != Pass {
		t.Errorf("dewey severity = %v, want Pass", r.Severity)
	}
}

func TestCheckMCPConfig_ArrayCommand(t *testing.T) {
	dir := t.TempDir()

	createFile(t, dir, "opencode.json", `{
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"dewey": true}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	// Verify dewey binary was extracted from array command.
	found := false
	for _, r := range group.Results {
		if r.Name == "dewey" && r.Severity == Pass {
			found = true
			if !strings.Contains(r.Message, "dewey binary found") {
				t.Errorf("message = %q, want 'dewey binary found'", r.Message)
			}
		}
	}
	if !found {
		t.Error("dewey result not found with Pass severity")
	}
}

// --- Phase 3: User Story 1 — Embedding Capability tests (T009-T012) ---

func TestCheckDewey_EmbeddingCapability_Pass(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir .dewey: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
			},
			nil,
		),
		ReadFile: os.ReadFile,
		// Mock EmbedCheck returning nil (success).
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	r, ok := results["embedding capability"]
	if !ok {
		t.Fatal("embedding capability check not found in results")
	}
	if r.Severity != Pass {
		t.Errorf("embedding capability severity = %v, want Pass", r.Severity)
	}
	if !strings.Contains(r.Message, "granite-embedding:30m") {
		t.Errorf("embedding capability message = %q, want model name", r.Message)
	}

	// Verify JSON output includes the check (FR-007).
	var buf bytes.Buffer
	report := &Report{
		Environment: DetectedEnvironment{Managers: []ManagerInfo{}, Platform: "test"},
		Groups:      []CheckGroup{group},
		Summary:     computeSummary([]CheckGroup{group}),
	}
	if err := FormatJSON(report, &buf); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	jsonOut := buf.String()
	if !strings.Contains(jsonOut, `"embedding capability"`) {
		t.Error("JSON output should contain embedding capability check")
	}
}

func TestCheckDewey_EmbeddingCapability_Fail(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir .dewey: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
			},
			nil,
		),
		ReadFile: os.ReadFile,
		// Mock EmbedCheck returning an error.
		EmbedCheck: func(model string) error {
			return fmt.Errorf("embed request returned status 500")
		},
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	r, ok := results["embedding capability"]
	if !ok {
		t.Fatal("embedding capability check not found in results")
	}
	if r.Severity != Warn {
		t.Errorf("embedding capability severity = %v, want Warn", r.Severity)
	}
	if r.InstallHint == "" {
		t.Error("embedding capability should have actionable install hint")
	}
}

func TestCheckDewey_EmbeddingCapability_Skip(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir:  dir,
		LookPath:   stubLookPathSimple(map[string]bool{}), // dewey not found
		ExecCmd:    stubExecCmd(nil, nil),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	// Find embedding capability in skip results.
	found := false
	for _, r := range group.Results {
		if r.Name == "embedding capability" {
			found = true
			if !strings.Contains(r.Message, "skipped: dewey not installed") {
				t.Errorf("embedding capability message = %q, want 'skipped: dewey not installed'", r.Message)
			}
			if r.Severity != Pass {
				t.Errorf("embedding capability severity = %v, want Pass (skipped)", r.Severity)
			}
		}
	}
	if !found {
		t.Error("embedding capability skip result not found")
	}
}

func TestCheckDewey_EmbeddingCapability_ConnectionRefused(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir .dewey: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
			},
			nil,
		),
		ReadFile: os.ReadFile,
		// Mock EmbedCheck returning connection refused error.
		EmbedCheck: func(model string) error {
			return fmt.Errorf("embed request failed: dial tcp 127.0.0.1:11434: connection refused")
		},
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	r, ok := results["embedding capability"]
	if !ok {
		t.Fatal("embedding capability check not found in results")
	}
	if r.Severity != Warn {
		t.Errorf("embedding capability severity = %v, want Warn", r.Severity)
	}
	if !strings.Contains(r.InstallHint, "ollama serve") {
		t.Errorf("install hint = %q, want 'ollama serve'", r.InstallHint)
	}
}

// --- Phase 5: User Story 3 — Ollama Demotion test (T031) ---

func TestCheckDewey_OllamaDemotion(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir .dewey: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"ollama list": "NAME\ngranite-embedding:30m   abc123   63 MB\n",
			},
			nil,
		),
		ReadFile:   os.ReadFile,
		EmbedCheck: func(model string) error { return nil },
	}

	group := checkDewey(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	r, ok := results["embedding model"]
	if !ok {
		t.Fatal("embedding model check not found in results")
	}
	if !strings.Contains(r.Message, "(Dewey manages Ollama lifecycle)") {
		t.Errorf("embedding model message = %q, want '(Dewey manages Ollama lifecycle)' annotation", r.Message)
	}

	// Verify annotation appears in JSON output (FR-007 / T033).
	var buf bytes.Buffer
	report := &Report{
		Environment: DetectedEnvironment{Managers: []ManagerInfo{}, Platform: "test"},
		Groups:      []CheckGroup{group},
		Summary:     computeSummary([]CheckGroup{group}),
	}
	if err := FormatJSON(report, &buf); err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	jsonOut := buf.String()
	if !strings.Contains(jsonOut, "Dewey manages Ollama lifecycle") {
		t.Error("JSON output should contain Ollama demotion annotation")
	}
}

func TestCheckMCPConfig_StringCommand(t *testing.T) {
	dir := t.TempDir()

	// Uses string-style command (backward compat).
	createFile(t, dir, "opencode.json", `{
  "mcp": {
    "dewey": {
      "type": "local",
      "command": "dewey",
      "enabled": true
    }
  }
}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"dewey": true}),
		ReadFile:  os.ReadFile,
	}

	group := checkMCPConfig(opts)

	found := false
	for _, r := range group.Results {
		if r.Name == "dewey" && r.Severity == Pass {
			found = true
		}
	}
	if !found {
		t.Error("dewey result not found — string command backward compat failed")
	}
}

// --- Agent Context check group tests ---

// completeAGENTSmd returns AGENTS.md content with all Tier 1
// sections and a constitution reference.
func completeAGENTSmd() string {
	return `# AGENTS.md

## Project Overview

This is a test project.

## Build & Test Commands

` + "```" + `bash
make build
make test
` + "```" + `

## Project Structure

` + "```" + `text
project/
├── cmd/
├── internal/
` + "```" + `

## Code Conventions

- Use gofmt
- Error wrapping with fmt.Errorf

## Active Technologies

- Go 1.24+
- Cobra CLI

## Architecture

Options/Result pattern.

## Testing Conventions

Standard library testing only.

## Git & Workflow

Conventional commits, feature branches.

## Behavioral Constraints

Never modify coverage thresholds.

## Constitution (Highest Authority)

The org constitution at .specify/memory/constitution.md
defines the core principles. Speckit and OpenSpec changes
must both align with these principles regardless of framework.
`
}

func TestCheckAgentContext_NoFile(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkAgentContext(opts)

	if len(group.Results) != 1 {
		t.Fatalf("expected 1 result for missing file, got %d", len(group.Results))
	}
	r := group.Results[0]
	if r.Name != "AGENTS.md" {
		t.Errorf("name = %q, want AGENTS.md", r.Name)
	}
	if r.Severity != Fail {
		t.Errorf("severity = %v, want Fail", r.Severity)
	}
	if r.InstallHint != "Run: /agent-brief in OpenCode" {
		t.Errorf("install hint = %q, want 'Run: /agent-brief in OpenCode'", r.InstallHint)
	}
}

func TestCheckAgentContext_AllTier1Present(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "AGENTS.md", completeAGENTSmd())

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkAgentContext(opts)

	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	tier1Sections := []string{
		"Tier 1: Project Overview",
		"Tier 1: Build Commands",
		"Tier 1: Project Structure",
		"Tier 1: Code Conventions",
		"Tier 1: Technology Stack",
	}
	for _, sec := range tier1Sections {
		r, ok := results[sec]
		if !ok {
			t.Errorf("missing check result for %q", sec)
			continue
		}
		if r.Severity != Pass {
			t.Errorf("%s severity = %v, want Pass", sec, r.Severity)
		}
	}
}

func TestCheckAgentContext_MissingTier1Section(t *testing.T) {
	tests := []struct {
		name    string
		content string
		missing string
	}{
		{
			name:    "missing overview",
			content: "## Build\n## Project Structure\n## Code Conventions\n## Active Technologies\n",
			missing: "Tier 1: Project Overview",
		},
		{
			name:    "missing build",
			content: "## Project Overview\n## Project Structure\n## Code Conventions\n## Active Technologies\n",
			missing: "Tier 1: Build Commands",
		},
		{
			name:    "missing structure",
			content: "## Project Overview\n## Build\n## Code Conventions\n## Active Technologies\n",
			missing: "Tier 1: Project Structure",
		},
		{
			name:    "missing conventions",
			content: "## Project Overview\n## Build\n## Project Structure\n## Active Technologies\n",
			missing: "Tier 1: Code Conventions",
		},
		{
			name:    "missing tech stack",
			content: "## Project Overview\n## Build\n## Project Structure\n## Code Conventions\n",
			missing: "Tier 1: Technology Stack",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			createFile(t, dir, "AGENTS.md", tt.content)

			opts := &Options{
				TargetDir: dir,
				ReadFile:  os.ReadFile,
			}

			group := checkAgentContext(opts)

			results := make(map[string]CheckResult)
			for _, r := range group.Results {
				results[r.Name] = r
			}

			r, ok := results[tt.missing]
			if !ok {
				t.Fatalf("missing check result for %q", tt.missing)
			}
			if r.Severity != Fail {
				t.Errorf("%s severity = %v, want Fail", tt.missing, r.Severity)
			}
		})
	}
}

func TestCheckAgentContext_BuildCodeBlocks(t *testing.T) {
	t.Run("no code blocks", func(t *testing.T) {
		dir := t.TempDir()
		content := "## Build & Test Commands\n\nRun make build.\n\n## Project Structure\n"
		createFile(t, dir, "AGENTS.md", content)

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		r, ok := results["Build code blocks"]
		if !ok {
			t.Fatal("missing Build code blocks check")
		}
		if r.Severity != Warn {
			t.Errorf("severity = %v, want Warn", r.Severity)
		}
	})

	t.Run("with code blocks", func(t *testing.T) {
		dir := t.TempDir()
		content := "## Build & Test Commands\n\n```bash\nmake build\n```\n\n## Project Structure\n"
		createFile(t, dir, "AGENTS.md", content)

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		r, ok := results["Build code blocks"]
		if !ok {
			t.Fatal("missing Build code blocks check")
		}
		if r.Severity != Pass {
			t.Errorf("severity = %v, want Pass", r.Severity)
		}
	})
}

func TestCheckAgentContext_LineCount(t *testing.T) {
	t.Run("under threshold", func(t *testing.T) {
		dir := t.TempDir()
		content := strings.Repeat("line\n", 100)
		createFile(t, dir, "AGENTS.md", content)

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Line count"]; r.Severity != Pass {
			t.Errorf("severity = %v, want Pass for 100 lines", r.Severity)
		}
	})

	t.Run("over threshold", func(t *testing.T) {
		dir := t.TempDir()
		content := strings.Repeat("line\n", 350)
		createFile(t, dir, "AGENTS.md", content)

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Line count"]; r.Severity != Warn {
			t.Errorf("severity = %v, want Warn for 350 lines", r.Severity)
		}
	})
}

func TestCheckAgentContext_ConstitutionReference(t *testing.T) {
	t.Run("referenced", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\nThe constitution governs all work.\n")
		createFile(t, dir, ".specify/memory/constitution.md", "# Constitution\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		r, ok := results["Constitution reference"]
		if !ok {
			t.Fatal("missing Constitution reference check")
		}
		if r.Severity != Pass {
			t.Errorf("severity = %v, want Pass", r.Severity)
		}
	})

	t.Run("not referenced", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\nJust a project.\n")
		createFile(t, dir, ".specify/memory/constitution.md", "# Constitution\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		r, ok := results["Constitution reference"]
		if !ok {
			t.Fatal("missing Constitution reference check")
		}
		if r.Severity != Warn {
			t.Errorf("severity = %v, want Warn", r.Severity)
		}
	})
}

func TestCheckAgentContext_ConstitutionSkipped(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "AGENTS.md", "## Overview\nJust a project.\n")
	// No .specify/memory/constitution.md

	opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
	group := checkAgentContext(opts)

	for _, r := range group.Results {
		if r.Name == "Constitution reference" {
			t.Error("Constitution reference check should be omitted when .specify/ does not exist")
		}
	}
}

func TestCheckAgentContext_SpecFrameworkReference(t *testing.T) {
	t.Run("described", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\nThis project uses Speckit for specs.\n")
		createFile(t, dir, "specs/001-feature/spec.md", "# Spec")
		createFile(t, dir, "openspec/config.yaml", "schema: test")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		r, ok := results["Spec framework described"]
		if !ok {
			t.Fatal("missing Spec framework described check")
		}
		if r.Severity != Pass {
			t.Errorf("severity = %v, want Pass", r.Severity)
		}
	})

	t.Run("not described", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\nJust a project.\n")
		createFile(t, dir, "specs/001-feature/spec.md", "# Spec")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		r, ok := results["Spec framework described"]
		if !ok {
			t.Fatal("missing Spec framework described check")
		}
		if r.Severity != Warn {
			t.Errorf("severity = %v, want Warn", r.Severity)
		}
	})
}

func TestCheckAgentContext_SpecFrameworkSkipped(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "AGENTS.md", "## Overview\nJust a project.\n")
	// No specs/ or openspec/ directories

	opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
	group := checkAgentContext(opts)

	for _, r := range group.Results {
		if r.Name == "Spec framework described" {
			t.Error("Spec framework check should be omitted when no specs/ or openspec/ exist")
		}
	}
}

func TestCheckAgentContext_BridgeCLAUDEmd(t *testing.T) {
	t.Run("imports AGENTS.md", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\n")
		createFile(t, dir, "CLAUDE.md", "# Claude\n@AGENTS.md\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Bridge: CLAUDE.md"]; r.Severity != Pass {
			t.Errorf("severity = %v, want Pass", r.Severity)
		}
	})

	t.Run("missing", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Bridge: CLAUDE.md"]; r.Severity != Warn {
			t.Errorf("severity = %v, want Warn", r.Severity)
		}
	})

	t.Run("exists but no reference", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\n")
		createFile(t, dir, "CLAUDE.md", "# Claude\nSome content.\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Bridge: CLAUDE.md"]; r.Severity != Warn {
			t.Errorf("severity = %v, want Warn", r.Severity)
		}
	})
}

func TestCheckAgentContext_BridgeCursorrules(t *testing.T) {
	t.Run("references AGENTS.md", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\n")
		createFile(t, dir, ".cursorrules", "Read AGENTS.md for conventions.\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Bridge: .cursorrules"]; r.Severity != Pass {
			t.Errorf("severity = %v, want Pass", r.Severity)
		}
	})

	t.Run("missing", func(t *testing.T) {
		dir := t.TempDir()
		createFile(t, dir, "AGENTS.md", "## Overview\n")

		opts := &Options{TargetDir: dir, ReadFile: os.ReadFile}
		group := checkAgentContext(opts)

		results := make(map[string]CheckResult)
		for _, r := range group.Results {
			results[r.Name] = r
		}

		if r := results["Bridge: .cursorrules"]; r.Severity != Warn {
			t.Errorf("severity = %v, want Warn", r.Severity)
		}
	})
}

func TestCheckAgentContext_FullPass(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "AGENTS.md", completeAGENTSmd())
	createFile(t, dir, ".specify/memory/constitution.md", "# Constitution\n")
	createFile(t, dir, "specs/001-feature/spec.md", "# Spec")
	createFile(t, dir, "openspec/config.yaml", "schema: unbound-force")
	createFile(t, dir, "CLAUDE.md", "# Claude\n@AGENTS.md\n")
	createFile(t, dir, ".cursorrules", "Read AGENTS.md for conventions.\n")

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
	}

	group := checkAgentContext(opts)

	if group.Name != "Agent Context" {
		t.Errorf("group name = %q, want Agent Context", group.Name)
	}

	// All 12 checks should be present and passing.
	for _, r := range group.Results {
		if r.Severity != Pass {
			t.Errorf("check %q: severity = %v, want Pass (message: %s)",
				r.Name, r.Severity, r.Message)
		}
	}

	// Verify expected check count: 1 (existence) + 5 (tier1) +
	// 1 (code blocks) + 1 (line count) + 1 (constitution) +
	// 1 (spec framework) + 2 (bridges) = 12.
	if len(group.Results) != 12 {
		t.Errorf("expected 12 check results, got %d", len(group.Results))
		for _, r := range group.Results {
			t.Logf("  %s: %v — %s", r.Name, r.Severity, r.Message)
		}
	}
}
