package doctor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/unbound-force/unbound-force/internal/orchestration"
	"gopkg.in/yaml.v3"
)

// checkDetectedEnvironment builds the "Detected Environment" group
// listing all detected managers per FR-000a. All items are Pass
// severity — this section is informational only.
func checkDetectedEnvironment(env DetectedEnvironment) CheckGroup {
	group := CheckGroup{
		Name:    "Detected Environment",
		Results: []CheckResult{},
	}

	for _, m := range env.Managers {
		group.Results = append(group.Results, CheckResult{
			Name:     string(m.Kind),
			Severity: Pass,
			Message:  managerDescription(m.Kind) + " (" + strings.Join(m.Manages, ", ") + ")",
			Detail:   m.Path,
		})
	}

	if len(group.Results) == 0 {
		group.Results = append(group.Results, CheckResult{
			Name:     "none",
			Severity: Pass,
			Message:  "No version managers detected",
			Detail:   "Using system defaults",
		})
	}

	return group
}

// managerDescription returns a human-readable description for a
// manager kind.
func managerDescription(kind ManagerKind) string {
	switch kind {
	case ManagerGoenv:
		return "Go version manager"
	case ManagerPyenv:
		return "Python version manager"
	case ManagerNvm:
		return "Node version manager"
	case ManagerFnm:
		return "Fast Node manager"
	case ManagerMise:
		return "Polyglot version manager"
	case ManagerBun:
		return "Bun JavaScript runtime"
	case ManagerHomebrew:
		return "Package manager"
	default:
		return string(kind)
	}
}

// toolSpec defines how to check a binary tool.
type toolSpec struct {
	name         string
	required     bool // true=Fail if missing, false=Warn or Pass(info)
	recommended  bool // true=Warn if missing (recommended tools)
	versionCmd   []string
	versionParse func(output string) (string, error)
	minVersion   string
	versionCheck func(version string, min string) bool
}

// coreToolSpecs defines the 8 binaries to check per FR-001/002/003.
var coreToolSpecs = []toolSpec{
	{
		name:         "go",
		required:     true,
		versionCmd:   []string{"go", "version"},
		versionParse: parseGoVersion,
		minVersion:   "1.24",
		versionCheck: checkGoVersion,
	},
	{
		name:     "opencode",
		required: true,
	},
	{
		name:        "gaze",
		recommended: true,
	},
	{
		name:        "mxf",
		recommended: true,
	},
	{
		name:         "node",
		versionCmd:   []string{"node", "--version"},
		versionParse: parseNodeVersion,
		minVersion:   "18",
		versionCheck: checkNodeVersion,
	},
	{
		name: "gh",
	},
	{
		name: "swarm",
	},
	{
		name: "ollama",
	},
}

// checkCoreTools checks the core binaries per FR-001/002/003.
func checkCoreTools(opts *Options, env DetectedEnvironment) CheckGroup {
	group := CheckGroup{
		Name:    "Core Tools",
		Results: []CheckResult{},
	}

	for _, spec := range coreToolSpecs {
		result := checkOneTool(spec, opts, env)
		group.Results = append(group.Results, result)

		// Ollama post-check: when ollama is found, verify
		// the granite-embedding:30m model is pulled.
		if spec.name == "ollama" && result.Severity == Pass && result.Message != "not found" {
			result = checkOllamaModel(opts, result)
			// Replace the last result with the enriched one.
			group.Results[len(group.Results)-1] = result
		}
	}

	return group
}

// checkOllamaModel checks whether the granite-embedding:30m model
// is available in the local Ollama installation. Enriches the
// existing CheckResult with model status.
func checkOllamaModel(opts *Options, base CheckResult) CheckResult {
	output, err := opts.ExecCmd("ollama", "list")
	if err != nil {
		// ollama list failed — keep existing result, add hint.
		base.InstallHint = "ollama pull granite-embedding:30m"
		return base
	}

	if strings.Contains(string(output), "granite-embedding") {
		base.Message = base.Message + " (granite-embedding:30m model ready)"
		return base
	}

	// Model not pulled.
	base.InstallHint = "ollama pull granite-embedding:30m"
	base.Message = base.Message + " (model not pulled)"
	return base
}

// checkOneTool checks a single tool binary.
func checkOneTool(spec toolSpec, opts *Options, env DetectedEnvironment) CheckResult {
	path, err := opts.LookPath(spec.name)
	if err != nil {
		// Tool not found — determine severity based on classification.
		sev := Pass // optional: informational
		if spec.required {
			sev = Fail
		} else if spec.recommended {
			sev = Warn
		}

		return CheckResult{
			Name:        spec.name,
			Severity:    sev,
			Message:     "not found",
			InstallHint: installHint(spec.name, env),
			InstallURL:  installURL(spec.name),
		}
	}

	// Tool found — detect provenance and version.
	manager := DetectProvenance(path, opts)
	viaStr := ""
	if manager != ManagerUnknown {
		viaStr = " via " + string(manager)
	}

	// If there's a version command, run it.
	if len(spec.versionCmd) > 0 && spec.versionParse != nil {
		output, execErr := opts.ExecCmd(spec.versionCmd[0], spec.versionCmd[1:]...)
		if execErr != nil {
			// Command failed — pass with warning about version.
			return CheckResult{
				Name:     spec.name,
				Severity: Warn,
				Message:  "installed, version could not be verified" + viaStr,
				Detail:   path,
			}
		}

		version, parseErr := spec.versionParse(string(output))
		if parseErr != nil {
			// Unparseable version output — pass with warning per edge case.
			return CheckResult{
				Name:     spec.name,
				Severity: Warn,
				Message:  "installed, version could not be verified" + viaStr,
				Detail:   path,
			}
		}

		// Check minimum version if specified.
		if spec.minVersion != "" && spec.versionCheck != nil {
			if !spec.versionCheck(version, spec.minVersion) {
				hint := installHint(spec.name, env)
				return CheckResult{
					Name:        spec.name,
					Severity:    Fail,
					Message:     version + viaStr + " (requires >= " + spec.minVersion + ")",
					Detail:      path,
					InstallHint: hint,
					InstallURL:  installURL(spec.name),
				}
			}
		}

		return CheckResult{
			Name:     spec.name,
			Severity: Pass,
			Message:  version + viaStr,
			Detail:   path,
		}
	}

	// No version command — just report as installed.
	return CheckResult{
		Name:     spec.name,
		Severity: Pass,
		Message:  "installed" + viaStr,
		Detail:   path,
	}
}

// parseGoVersion extracts the version from `go version` output.
// Expected format: "go version go1.24.3 darwin/arm64"
func parseGoVersion(output string) (string, error) {
	parts := strings.Fields(output)
	for _, p := range parts {
		if strings.HasPrefix(p, "go") && len(p) > 2 {
			ver := strings.TrimPrefix(p, "go")
			// Verify it looks like a version number.
			if len(ver) > 0 && (ver[0] >= '0' && ver[0] <= '9') {
				return ver, nil
			}
		}
	}
	return "", fmt.Errorf("could not parse go version from: %s", output)
}

// checkGoVersion verifies Go version >= minimum.
func checkGoVersion(version, min string) bool {
	vMajor, vMinor := parseVersionParts(version)
	mMajor, mMinor := parseVersionParts(min)

	if vMajor != mMajor {
		return vMajor > mMajor
	}
	return vMinor >= mMinor
}

// parseVersionParts extracts major.minor from a version string.
// Handles non-numeric suffixes like "25-abcdef" by extracting
// the leading numeric portion.
func parseVersionParts(version string) (int, int) {
	parts := strings.SplitN(version, ".", 3)
	major := 0
	minor := 0
	if len(parts) >= 1 {
		major, _ = strconv.Atoi(extractLeadingDigits(parts[0]))
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(extractLeadingDigits(parts[1]))
	}
	return major, minor
}

// extractLeadingDigits returns the leading numeric portion of a
// string. E.g., "25-abcdef" -> "25", "3" -> "3".
func extractLeadingDigits(s string) string {
	for i, c := range s {
		if c < '0' || c > '9' {
			return s[:i]
		}
	}
	return s
}

// parseNodeVersion extracts the version from `node --version` output.
// Expected format: "v22.15.0"
func parseNodeVersion(output string) (string, error) {
	trimmed := strings.TrimSpace(output)
	if strings.HasPrefix(trimmed, "v") {
		return strings.TrimPrefix(trimmed, "v"), nil
	}
	return "", fmt.Errorf("could not parse node version from: %s", output)
}

// checkNodeVersion verifies Node.js version >= minimum.
func checkNodeVersion(version, min string) bool {
	vMajor, _ := parseVersionParts(version)
	mMajor, _ := parseVersionParts(min)
	return vMajor >= mMajor
}

// checkSwarmPlugin checks the Swarm plugin installation, runs
// `swarm doctor`, checks .hive/ and plugin config per
// FR-008/009/010/012.
func checkSwarmPlugin(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Swarm Plugin",
		Results: []CheckResult{},
	}

	// Check swarm binary.
	swarmPath, err := opts.LookPath("swarm")
	if err != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "swarm",
			Severity:    Fail,
			Message:     "not found",
			InstallHint: "npm install -g opencode-swarm-plugin@latest",
			InstallURL:  "https://www.swarmtools.ai/",
		})
		return group
	}

	group.Results = append(group.Results, CheckResult{
		Name:     "swarm",
		Severity: Pass,
		Message:  "installed",
		Detail:   swarmPath,
	})

	// Run swarm doctor with 10-second timeout per FR-009.
	// Uses ExecCmdTimeout which defaults to exec.CommandContext
	// with context.WithTimeout in production.
	output, swarmErr := opts.ExecCmdTimeout(10*time.Second, "swarm", "doctor")
	if swarmErr != nil {
		errMsg := swarmErr.Error()
		if strings.Contains(errMsg, "timed out") || strings.Contains(errMsg, "deadline exceeded") {
			group.Results = append(group.Results, CheckResult{
				Name:        "swarm doctor",
				Severity:    Warn,
				Message:     "swarm doctor timed out",
				InstallHint: "Run swarm doctor manually",
			})
		} else {
			group.Embed = string(output)
			group.Results = append(group.Results, CheckResult{
				Name:        "swarm doctor",
				Severity:    Warn,
				Message:     "swarm doctor reported issues",
				InstallHint: "Run: uf setup",
			})
		}
	} else {
		group.Embed = string(output)
	}

	// Check .hive/ existence (FR-010).
	hivePath := filepath.Join(opts.TargetDir, ".hive")
	if info, statErr := os.Stat(hivePath); statErr == nil && info.IsDir() {
		group.Results = append(group.Results, CheckResult{
			Name:     ".hive/",
			Severity: Pass,
			Message:  "initialized",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        ".hive/",
			Severity:    Warn,
			Message:     "not initialized",
			InstallHint: "Run: swarm init",
		})
	}

	// Check plugin config in opencode.json (FR-012).
	ocPath := filepath.Join(opts.TargetDir, "opencode.json")
	ocData, readErr := opts.ReadFile(ocPath)
	if readErr != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "plugin config",
			Severity:    Warn,
			Message:     "opencode.json not found",
			InstallHint: "Run: uf setup",
		})
	} else {
		var ocMap map[string]json.RawMessage
		if jsonErr := json.Unmarshal(ocData, &ocMap); jsonErr != nil {
			group.Results = append(group.Results, CheckResult{
				Name:        "plugin config",
				Severity:    Warn,
				Message:     "opencode.json could not be parsed",
				InstallHint: "Fix JSON syntax in opencode.json",
			})
		} else if pluginRaw, ok := ocMap["plugin"]; ok {
			var plugins []string
			if pErr := json.Unmarshal(pluginRaw, &plugins); pErr == nil {
				found := false
				for _, p := range plugins {
					if p == "opencode-swarm-plugin" {
						found = true
						break
					}
				}
				if found {
					group.Results = append(group.Results, CheckResult{
						Name:     "plugin config",
						Severity: Pass,
						Message:  "opencode-swarm-plugin in opencode.json",
					})
				} else {
					group.Results = append(group.Results, CheckResult{
						Name:        "plugin config",
						Severity:    Warn,
						Message:     "opencode-swarm-plugin not in plugin array",
						InstallHint: "Fix: uf setup",
					})
				}
			} else {
				group.Results = append(group.Results, CheckResult{
					Name:        "plugin config",
					Severity:    Warn,
					Message:     "plugin array could not be parsed",
					InstallHint: "Fix: uf setup",
				})
			}
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "plugin config",
				Severity:    Warn,
				Message:     "no plugin key in opencode.json",
				InstallHint: "Fix: uf setup",
			})
		}
	}

	return group
}

// checkScaffoldedFiles verifies that uf init files exist
// per FR-006.
func checkScaffoldedFiles(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Scaffolded Files",
		Results: []CheckResult{},
	}

	// Check .opencode/agents/ with file count.
	agentsDir := filepath.Join(opts.TargetDir, ".opencode", "agents")
	group.Results = append(group.Results, checkDirWithCount(agentsDir, ".opencode/agents/", "agent files", ".md"))

	// Check .opencode/command/ with file count.
	commandDir := filepath.Join(opts.TargetDir, ".opencode", "command")
	group.Results = append(group.Results, checkDirWithCount(commandDir, ".opencode/command/", "command files", ".md"))

	// Check .opencode/unbound/packs/ for convention packs.
	packsDir := filepath.Join(opts.TargetDir, ".opencode", "unbound", "packs")
	group.Results = append(group.Results, checkDirWithCount(packsDir, ".opencode/unbound/packs/", "convention packs", ".md"))

	// Check .specify/ existence.
	specifyDir := filepath.Join(opts.TargetDir, ".specify")
	if info, err := os.Stat(specifyDir); err == nil && info.IsDir() {
		group.Results = append(group.Results, CheckResult{
			Name:     ".specify/",
			Severity: Pass,
			Message:  "present",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        ".specify/",
			Severity:    Fail,
			Message:     "not found",
			InstallHint: "Run: uf init",
		})
	}

	// Check AGENTS.md existence.
	agentsMd := filepath.Join(opts.TargetDir, "AGENTS.md")
	if info, err := os.Stat(agentsMd); err == nil && !info.IsDir() {
		group.Results = append(group.Results, CheckResult{
			Name:     "AGENTS.md",
			Severity: Pass,
			Message:  "present",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        "AGENTS.md",
			Severity:    Fail,
			Message:     "not found",
			InstallHint: "Run: uf init",
		})
	}

	return group
}

// checkDirWithCount checks a directory exists and counts files
// with the given extension.
func checkDirWithCount(dir, name, label, ext string) CheckResult {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return CheckResult{
			Name:        name,
			Severity:    Fail,
			Message:     "not found",
			InstallHint: "Run: uf init",
		}
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ext) {
			count++
		}
	}

	if count == 0 {
		return CheckResult{
			Name:        name,
			Severity:    Warn,
			Message:     "directory exists but no " + label,
			InstallHint: "Run: uf init",
		}
	}

	return CheckResult{
		Name:     name,
		Severity: Pass,
		Message:  fmt.Sprintf("%d %s", count, label),
	}
}

// checkHeroAvailability checks for all 5 heroes per FR-007,
// reusing orchestration.DetectHeroes.
func checkHeroAvailability(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Hero Availability",
		Results: []CheckResult{},
	}

	agentDir := filepath.Join(opts.TargetDir, ".opencode", "agents")
	heroes, err := orchestration.DetectHeroes(agentDir, opts.LookPath)
	if err != nil {
		group.Results = append(group.Results, CheckResult{
			Name:     "detection",
			Severity: Warn,
			Message:  fmt.Sprintf("hero detection failed: %v", err),
		})
		return group
	}

	// Map hero names to human-readable display names.
	displayNames := map[string]string{
		"muti-mind":    "Muti-Mind (PO)",
		"cobalt-crush": "Cobalt-Crush (Dev)",
		"gaze":         "Gaze (Tester)",
		"divisor":      "The Divisor (Reviewer)",
		"mx-f":         "Mx F (Manager)",
	}

	for _, h := range heroes {
		displayName := displayNames[h.Name]
		if displayName == "" {
			displayName = h.Name
		}

		if h.Available {
			method := "agent: " + h.AgentFile
			if h.DetectionMethod == "exec_lookpath" {
				method = "binary"
			}
			// Special case: Divisor shows persona count.
			if h.Name == "divisor" {
				count := countDivisorPersonas(agentDir)
				if count > 1 {
					method = fmt.Sprintf("agent: %s (+%d personas)", h.AgentFile, count-1)
				}
			}
			group.Results = append(group.Results, CheckResult{
				Name:     displayName,
				Severity: Pass,
				Message:  method,
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        displayName,
				Severity:    Warn,
				Message:     "not available",
				InstallHint: "Run: uf init",
			})
		}
	}

	return group
}

// countDivisorPersonas counts divisor-*.md files in the agent dir.
func countDivisorPersonas(agentDir string) int {
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "divisor-") && strings.HasSuffix(e.Name(), ".md") {
			count++
		}
	}
	return count
}

// checkMCPConfig parses opencode.json and checks MCP server
// binaries per FR-011.
func checkMCPConfig(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "MCP Server Config",
		Results: []CheckResult{},
	}

	ocPath := filepath.Join(opts.TargetDir, "opencode.json")
	data, err := opts.ReadFile(ocPath)
	if err != nil {
		group.Results = append(group.Results, CheckResult{
			Name:     "opencode.json",
			Severity: Warn,
			Message:  "not found",
		})
		return group
	}

	var ocMap map[string]json.RawMessage
	if jsonErr := json.Unmarshal(data, &ocMap); jsonErr != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "opencode.json",
			Severity:    Warn,
			Message:     "could not be parsed",
			InstallHint: "Fix JSON syntax in opencode.json",
		})
		return group
	}

	group.Results = append(group.Results, CheckResult{
		Name:     "opencode.json",
		Severity: Pass,
		Message:  "valid",
	})

	// Check MCP servers — prefer canonical "mcp" key, fall back to
	// legacy "mcpServers" key (FR-012).
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		mcpRaw, ok = ocMap["mcpServers"]
		if !ok {
			return group
		}
	}

	var servers map[string]json.RawMessage
	if sErr := json.Unmarshal(mcpRaw, &servers); sErr != nil {
		return group
	}

	for name, serverRaw := range servers {
		// Extract the binary name from the command field.
		// Handles both string-style ("command": "dewey") and
		// array-style ("command": ["dewey", "serve", "--vault", "."]).
		binary := extractMCPBinary(serverRaw)
		if binary == "" {
			continue
		}

		// Check if the command binary exists.
		if _, lookErr := opts.LookPath(binary); lookErr != nil {
			group.Results = append(group.Results, CheckResult{
				Name:        name,
				Severity:    Warn,
				Message:     fmt.Sprintf("%s binary not found", binary),
				InstallHint: installURL(binary),
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:     name,
				Severity: Pass,
				Message:  binary + " binary found",
			})
		}
	}

	return group
}

// extractMCPBinary extracts the binary name from an MCP server
// definition's command field. Handles both string-style
// ("command": "dewey") and array-style ("command": ["dewey",
// "serve", "--vault", "."]) formats (FR-014). For array-style,
// the first element is the binary name.
func extractMCPBinary(serverRaw json.RawMessage) string {
	// Try parsing with string command first (legacy format).
	var stringDef struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(serverRaw, &stringDef); err == nil && stringDef.Command != "" {
		return stringDef.Command
	}

	// Try parsing with array command (canonical format).
	var arrayDef struct {
		Command []string `json:"command"`
	}
	if err := json.Unmarshal(serverRaw, &arrayDef); err == nil && len(arrayDef.Command) > 0 {
		return arrayDef.Command[0]
	}

	return ""
}

// checkAgentSkillIntegrity validates YAML frontmatter in agent
// and skill files per FR-013/FR-014.
func checkAgentSkillIntegrity(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Agent/Skill Integrity",
		Results: []CheckResult{},
	}

	// Validate agents (FR-013).
	agentDir := filepath.Join(opts.TargetDir, ".opencode", "agents")
	agentResult := validateAgents(agentDir, opts)
	group.Results = append(group.Results, agentResult)

	// Validate skills (FR-014) — check both skill/ and skills/ dirs.
	for _, skillBase := range []string{"skill", "skills"} {
		skillDir := filepath.Join(opts.TargetDir, ".opencode", skillBase)
		if info, err := os.Stat(skillDir); err == nil && info.IsDir() {
			skillResults := validateSkills(skillDir, opts)
			group.Results = append(group.Results, skillResults...)
		}
	}

	return group
}

// validateAgents walks .opencode/agents/*.md and validates YAML
// frontmatter per FR-013.
func validateAgents(agentDir string, opts *Options) CheckResult {
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return CheckResult{
			Name:        "agents",
			Severity:    Warn,
			Message:     "agents directory not found",
			InstallHint: "Run: uf init",
		}
	}

	total := 0
	invalid := 0
	var issues []string

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		total++

		data, readErr := opts.ReadFile(filepath.Join(agentDir, e.Name()))
		if readErr != nil {
			invalid++
			issues = append(issues, e.Name()+": could not read")
			continue
		}

		fm, parseErr := parseFrontmatter(data)
		if parseErr != nil {
			invalid++
			issues = append(issues, e.Name()+": invalid frontmatter")
			continue
		}

		desc, _ := fm["description"].(string)
		if desc == "" {
			invalid++
			issues = append(issues, e.Name()+": missing description")
		}
	}

	if total == 0 {
		return CheckResult{
			Name:        "agents",
			Severity:    Warn,
			Message:     "no agent files found",
			InstallHint: "Run: uf init",
		}
	}

	if invalid > 0 {
		return CheckResult{
			Name:        fmt.Sprintf("%d agents validated", total),
			Severity:    Warn,
			Message:     fmt.Sprintf("%d with issues: %s", invalid, strings.Join(issues, "; ")),
			InstallHint: "Fix frontmatter in agent files",
		}
	}

	return CheckResult{
		Name:     fmt.Sprintf("%d agents validated", total),
		Severity: Pass,
		Message:  "all frontmatter valid",
	}
}

// skillNameRegex validates skill names per FR-014.
var skillNameRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validateSkills walks skill directories and validates SKILL.md
// frontmatter per FR-014.
func validateSkills(skillDir string, opts *Options) []CheckResult {
	var results []CheckResult

	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return results
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		skillFile := filepath.Join(skillDir, e.Name(), "SKILL.md")
		data, readErr := opts.ReadFile(skillFile)
		if readErr != nil {
			results = append(results, CheckResult{
				Name:        e.Name(),
				Severity:    Warn,
				Message:     "SKILL.md not found",
				InstallHint: "Create SKILL.md with name and description frontmatter",
			})
			continue
		}

		fm, parseErr := parseFrontmatter(data)
		if parseErr != nil {
			results = append(results, CheckResult{
				Name:        e.Name(),
				Severity:    Warn,
				Message:     "invalid frontmatter in SKILL.md",
				InstallHint: "Fix YAML frontmatter in SKILL.md",
			})
			continue
		}

		name, _ := fm["name"].(string)
		desc, _ := fm["description"].(string)

		var issues []string
		if name == "" {
			issues = append(issues, "missing name")
		} else {
			if !skillNameRegex.MatchString(name) {
				issues = append(issues, fmt.Sprintf("name %q does not match ^[a-z0-9]+(-[a-z0-9]+)*$", name))
			}
			if name != e.Name() {
				issues = append(issues, fmt.Sprintf("name %q does not match directory %q", name, e.Name()))
			}
		}
		if desc == "" {
			issues = append(issues, "missing description")
		}

		if len(issues) > 0 {
			results = append(results, CheckResult{
				Name:        e.Name(),
				Severity:    Warn,
				Message:     strings.Join(issues, "; "),
				InstallHint: "Fix frontmatter in SKILL.md",
			})
		} else {
			results = append(results, CheckResult{
				Name:     fmt.Sprintf("1 skill validated"),
				Severity: Pass,
				Message:  name,
			})
		}
	}

	return results
}

// checkDewey checks the Dewey knowledge layer components:
// binary, embedding model, and workspace directory.
// Design decision: Dewey checks are a separate group (not part of
// Core Tools) because Dewey has multiple interdependent components
// that should be reported together. When the dewey binary is absent,
// remaining checks are skipped per the contract.
func checkDewey(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Dewey Knowledge Layer",
		Results: []CheckResult{},
	}

	// Check 1: dewey binary.
	deweyPath, err := opts.LookPath("dewey")
	if err != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "dewey binary",
			Severity:    Pass,
			Message:     "not found",
			InstallHint: "brew install unbound-force/tap/dewey",
		})
		// Skip remaining checks when dewey is not installed.
		group.Results = append(group.Results, CheckResult{
			Name:     "embedding model",
			Severity: Pass,
			Message:  "skipped: dewey not installed",
		})
		group.Results = append(group.Results, CheckResult{
			Name:     "workspace",
			Severity: Pass,
			Message:  "skipped: dewey not installed",
		})
		return group
	}

	group.Results = append(group.Results, CheckResult{
		Name:     "dewey binary",
		Severity: Pass,
		Message:  "found",
		Detail:   deweyPath,
	})

	// Check 2: embedding model via Ollama.
	ollamaOutput, ollamaErr := opts.ExecCmd("ollama", "list")
	if ollamaErr != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "embedding model",
			Severity:    Warn,
			Message:     "could not check (ollama not available)",
			InstallHint: "ollama pull granite-embedding:30m",
		})
	} else if strings.Contains(string(ollamaOutput), "granite-embedding") {
		group.Results = append(group.Results, CheckResult{
			Name:     "embedding model",
			Severity: Pass,
			Message:  "granite-embedding:30m installed",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        "embedding model",
			Severity:    Warn,
			Message:     "not pulled (graph-only mode available)",
			InstallHint: "ollama pull granite-embedding:30m",
		})
	}

	// Check 3: .dewey/ workspace directory.
	deweyDir := filepath.Join(opts.TargetDir, ".dewey")
	if info, statErr := os.Stat(deweyDir); statErr == nil && info.IsDir() {
		group.Results = append(group.Results, CheckResult{
			Name:     "workspace",
			Severity: Pass,
			Message:  "initialized",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        "workspace",
			Severity:    Warn,
			Message:     "not initialized",
			InstallHint: "dewey init",
		})
	}

	return group
}

// parseFrontmatter extracts YAML frontmatter from a Markdown file.
// Per research.md R6: split on --- delimiters, unmarshal with yaml.v3.
func parseFrontmatter(data []byte) (map[string]interface{}, error) {
	content := string(data)

	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("no frontmatter delimiter found")
	}

	// Find the closing --- delimiter.
	rest := content[3:]
	// Skip the newline after opening ---
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, fmt.Errorf("no closing frontmatter delimiter found")
	}

	yamlContent := rest[:endIdx]

	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	return fm, nil
}
