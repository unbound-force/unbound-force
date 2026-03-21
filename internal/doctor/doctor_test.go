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
			"go":          "/home/user/.goenv/shims/go",
			"gaze":        "/opt/homebrew/bin/gaze",
			"graphthulhu": "/opt/homebrew/bin/graphthulhu",
			"node":        "/home/user/.nvm/versions/node/v22.15.0/bin/node",
			"swarm":       "/usr/local/bin/swarm",
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

	// graphthulhu: found -> Pass
	if r, ok := results["graphthulhu"]; ok {
		if r.Severity != Pass {
			t.Errorf("graphthulhu severity = %v, want Pass", r.Severity)
		}
	} else {
		t.Error("graphthulhu check not found")
	}

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

	// swarm: found
	if r, ok := results["swarm"]; ok {
		if r.Severity != Pass {
			t.Errorf("swarm severity = %v, want Pass", r.Severity)
		}
	} else {
		t.Error("swarm check not found")
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
	createFile(t, dir, ".opencode/unbound/packs/go.md", "# Go pack")
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
	if r := results["AGENTS.md"]; r.Severity != Pass {
		t.Errorf("AGENTS.md severity = %v, want Pass", r.Severity)
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
    "knowledge-graph": {
      "command": "graphthulhu",
      "args": ["--port", "3000"]
    }
  }
}`
	createFile(t, dir, "opencode.json", ocJSON)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPathSimple(map[string]bool{"graphthulhu": true}),
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
	if r := results["knowledge-graph"]; r.Severity != Pass {
		t.Errorf("knowledge-graph severity = %v, want Pass", r.Severity)
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
	createFile(t, dir, ".opencode/unbound/packs/go.md", "# Go")
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

	// Verify 7 groups in correct order.
	expectedGroups := []string{
		"Detected Environment",
		"Core Tools",
		"Swarm Plugin",
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
	createFile(t, dir, ".opencode/unbound/packs/go.md", "# Go")
	createFile(t, dir, ".specify/config.yaml", "# config")
	createFile(t, dir, "AGENTS.md", "# Agents")
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"],"mcpServers":{}}`)
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	swarmOut := "✓ OK\n"

	var buf bytes.Buffer
	opts := Options{
		TargetDir: dir,
		Format:    "text",
		Stdout:    &buf,
		LookPath: stubLookPath(map[string]string{
			"go":       "/usr/local/go/bin/go",
			"opencode": "/usr/local/bin/opencode",
			"gaze":     "/usr/local/bin/gaze",
			"mxf":      "/usr/local/bin/mxf",
			"swarm":    "/usr/local/bin/swarm",
		}),
		ExecCmd: stubExecCmd(
			map[string]string{
				"go version":   "go version go1.24.3 darwin/arm64",
				"swarm doctor": swarmOut,
			},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{
				"swarm doctor": swarmOut,
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
	if len(report.Groups) != 7 {
		t.Errorf("expected 7 groups, got %d", len(report.Groups))
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

// --- Phase 4: User Story 2 tests (T032-T036) ---

func TestCheckSwarmPlugin_NotInstalled(t *testing.T) {
	opts := &Options{
		TargetDir:    t.TempDir(),
		LookPath:     stubLookPathSimple(map[string]bool{}),
		ExecCmd:      stubExecCmd(nil, nil),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkSwarmPlugin(opts)

	if len(group.Results) == 0 {
		t.Fatal("expected at least one result")
	}

	r := group.Results[0]
	if r.Severity != Fail {
		t.Errorf("swarm severity = %v, want Fail", r.Severity)
	}
	if !strings.Contains(r.InstallHint, "npm install -g opencode-swarm-plugin@latest") {
		t.Errorf("install hint = %q, want npm install command", r.InstallHint)
	}
}

func TestCheckSwarmPlugin_MissingPluginConfig(t *testing.T) {
	dir := t.TempDir()
	// opencode.json exists but no plugin array.
	createFile(t, dir, "opencode.json", `{"mcpServers":{}}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"swarm": "/usr/local/bin/swarm"}),
		ExecCmd: stubExecCmd(
			map[string]string{},
			map[string]error{"swarm doctor": fmt.Errorf("not configured")},
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{},
			map[string]error{"swarm doctor": fmt.Errorf("not configured")},
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkSwarmPlugin(opts)

	// Find plugin config result.
	for _, r := range group.Results {
		if r.Name == "plugin config" {
			if r.Severity != Warn {
				t.Errorf("plugin config severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.InstallHint, "unbound setup") {
				t.Errorf("install hint = %q, want 'unbound setup'", r.InstallHint)
			}
			return
		}
	}
	t.Error("plugin config result not found")
}

func TestCheckSwarmPlugin_Installed(t *testing.T) {
	dir := t.TempDir()

	// Create .hive/ dir and opencode.json with plugin entry.
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

	swarmDoctorOutput := "✓ OpenCode plugin configured\n✓ Hive storage: libSQL (embedded SQLite)\n✓ Semantic memory: ready\n✓ Dependencies: all installed\n"

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"swarm": "/usr/local/bin/swarm"}),
		ExecCmd: stubExecCmd(
			map[string]string{"swarm doctor": swarmDoctorOutput},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"swarm doctor": swarmDoctorOutput},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkSwarmPlugin(opts)

	// Verify swarm=Pass.
	results := make(map[string]CheckResult)
	for _, r := range group.Results {
		results[r.Name] = r
	}

	if r := results["swarm"]; r.Severity != Pass {
		t.Errorf("swarm severity = %v, want Pass", r.Severity)
	}

	// Verify embed contains swarm doctor output.
	if !strings.Contains(group.Embed, "OpenCode plugin configured") {
		t.Error("embed should contain swarm doctor output")
	}

	// Verify .hive/=Pass.
	if r := results[".hive/"]; r.Severity != Pass {
		t.Errorf(".hive/ severity = %v, want Pass", r.Severity)
	}

	// Verify plugin config=Pass.
	if r := results["plugin config"]; r.Severity != Pass {
		t.Errorf("plugin config severity = %v, want Pass", r.Severity)
	}
}

func TestCheckSwarmPlugin_DoctorFails(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"swarm": "/usr/local/bin/swarm"}),
		ExecCmd: stubExecCmd(
			map[string]string{"swarm doctor": "✗ Plugin not configured\n"},
			map[string]error{"swarm doctor": fmt.Errorf("exit status 1")},
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"swarm doctor": "✗ Plugin not configured\n"},
			map[string]error{"swarm doctor": fmt.Errorf("exit status 1")},
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkSwarmPlugin(opts)

	// Find swarm doctor result.
	for _, r := range group.Results {
		if r.Name == "swarm doctor" {
			if r.Severity != Warn {
				t.Errorf("swarm doctor severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.InstallHint, "unbound setup") {
				t.Errorf("install hint = %q, want 'unbound setup'", r.InstallHint)
			}
			// Verify stderr is embedded.
			if !strings.Contains(group.Embed, "Plugin not configured") {
				t.Error("embed should contain swarm doctor stderr output")
			}
			return
		}
	}
	t.Error("swarm doctor result not found")
}

func TestCheckSwarmPlugin_Timeout(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, "opencode.json", `{"plugin":["opencode-swarm-plugin"]}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"swarm": "/usr/local/bin/swarm"}),
		ExecCmd: stubExecCmd(
			nil,
			map[string]error{"swarm doctor": fmt.Errorf("context deadline exceeded: timed out")},
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			nil,
			map[string]error{"swarm doctor": fmt.Errorf("context deadline exceeded: timed out")},
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkSwarmPlugin(opts)

	// Find swarm doctor result.
	for _, r := range group.Results {
		if r.Name == "swarm doctor" {
			if r.Severity != Warn {
				t.Errorf("swarm doctor severity = %v, want Warn", r.Severity)
			}
			if !strings.Contains(r.Message, "timed out") {
				t.Errorf("message = %q, want 'timed out'", r.Message)
			}
			return
		}
	}
	t.Error("swarm doctor result not found")
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
	if !strings.Contains(output, "[PASS]") && !strings.Contains(output, "✓") {
		t.Error("expected pass indicator in output")
	}
	if !strings.Contains(output, "Unbound Force Doctor") {
		t.Error("expected header in output")
	}
	if !strings.Contains(output, "Summary:") {
		t.Error("expected summary in output")
	}
}

func TestFormatText_SwarmDoctorEmbed(t *testing.T) {
	report := &Report{
		Groups: []CheckGroup{
			{
				Name: "Swarm Plugin",
				Results: []CheckResult{
					{Name: "swarm", Severity: Pass, Message: "installed"},
				},
				Embed: "✓ OpenCode plugin configured\n✓ Dependencies: all installed\n",
			},
		},
		Summary: Summary{Total: 1, Passed: 1},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "OpenCode plugin configured") {
		t.Error("expected embedded swarm doctor output")
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
	if !strings.Contains(output, "Install: brew install anomalyco/tap/opencode") {
		t.Error("expected install hint in output")
	}
	if !strings.Contains(output, "Docs: https://opencode.ai/docs") {
		t.Error("expected install URL in output")
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
				Name: "Swarm Plugin",
				Results: []CheckResult{
					{Name: "swarm", Severity: Pass, Message: "installed"},
				},
				Embed: "✓ OpenCode plugin configured\n",
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

func TestInstallHint_BunForSwarm(t *testing.T) {
	env := DetectedEnvironment{
		Managers: []ManagerInfo{
			{Kind: ManagerBun, Path: "/home/user/.bun/bin/bun", Manages: []string{"node", "packages"}},
		},
	}
	hint := installHint("swarm", env)
	if !strings.Contains(hint, "bun") {
		t.Errorf("hint = %q, want bun command", hint)
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
		{"mxf", "brew install unbound-force/tap/mxf"},
		{"graphthulhu", "brew install unbound-force/tap/graphthulhu"},
		{"node", "brew install node"},
		{"gh", "brew install gh"},
		{"swarm", "npm install -g opencode-swarm-plugin@latest"},
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
		{"swarm"},
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
		{"swarm", ManagerBun, "bun add -g opencode-swarm-plugin@latest"},
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

func TestCheckSwarmPlugin_PluginArrayParseError(t *testing.T) {
	dir := t.TempDir()
	// plugin key exists but is not an array.
	createFile(t, dir, "opencode.json", `{"plugin": "not-an-array"}`)

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubLookPath(map[string]string{"swarm": "/usr/local/bin/swarm"}),
		ExecCmd: stubExecCmd(
			map[string]string{"swarm doctor": "ok"},
			nil,
		),
		ExecCmdTimeout: stubExecCmdTimeout(
			map[string]string{"swarm doctor": "ok"},
			nil,
		),
		EvalSymlinks: stubEvalSymlinks(nil),
		Getenv:       stubGetenv(map[string]string{}),
		ReadFile:     os.ReadFile,
	}

	group := checkSwarmPlugin(opts)

	for _, r := range group.Results {
		if r.Name == "plugin config" {
			if r.Severity != Warn {
				t.Errorf("unparseable plugin array should produce Warn, got %v", r.Severity)
			}
			return
		}
	}
	t.Error("plugin config result not found")
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

	ollamaListOutput := "NAME                    ID              SIZE    MODIFIED\nmxbai-embed-large:latest abc123  1.2 GB  2 days ago\nllama3:latest            def456  4.7 GB  1 week ago\n"

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
			if !strings.Contains(r.Message, "mxbai-embed-large model ready") {
				t.Errorf("ollama message = %q, want 'mxbai-embed-large model ready'", r.Message)
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
			if r.InstallHint != "ollama pull mxbai-embed-large" {
				t.Errorf("ollama install hint = %q, want 'ollama pull mxbai-embed-large'", r.InstallHint)
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
