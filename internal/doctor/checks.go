package doctor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/unbound-force/unbound-force/internal/orchestration"
	"gopkg.in/yaml.v3"
)

// graniteModel is the enterprise-grade embedding model used by
// Dewey for semantic search. Defined locally to avoid a circular
// dependency on internal/setup.
const graniteModel = "granite-embedding:30m"

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
		name: "replicator",
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

// checkReplicator checks the Replicator installation, runs
// `replicator doctor`, checks .uf/replicator/ and MCP config per FR-011.
func checkReplicator(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Replicator",
		Results: []CheckResult{},
	}

	// Check 1: replicator binary.
	replicatorPath, err := opts.LookPath("replicator")
	if err != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "replicator",
			Severity:    Warn,
			Message:     "not found",
			InstallHint: "brew install unbound-force/tap/replicator",
			InstallURL:  "https://github.com/unbound-force/replicator",
		})
		return group
	}

	group.Results = append(group.Results, CheckResult{
		Name:     "replicator",
		Severity: Pass,
		Message:  "installed",
		Detail:   replicatorPath,
	})

	// Check 2: replicator doctor delegation with 10-second timeout.
	output, repErr := opts.ExecCmdTimeout(10*time.Second, "replicator", "doctor")
	if repErr != nil {
		errMsg := repErr.Error()
		if strings.Contains(errMsg, "timed out") || strings.Contains(errMsg, "deadline exceeded") {
			group.Results = append(group.Results, CheckResult{
				Name:        "replicator doctor",
				Severity:    Warn,
				Message:     "replicator doctor timed out",
				InstallHint: "Run replicator doctor manually",
			})
		} else {
			group.Embed = string(output)
			group.Results = append(group.Results, CheckResult{
				Name:        "replicator doctor",
				Severity:    Warn,
				Message:     "replicator doctor reported issues",
				InstallHint: "Run: uf setup",
			})
		}
	} else {
		group.Embed = string(output)
	}

	// Check 3: .uf/replicator/ existence.
	replicatorDirPath := filepath.Join(opts.TargetDir, ".uf", "replicator")
	if info, statErr := os.Stat(replicatorDirPath); statErr == nil && info.IsDir() {
		group.Results = append(group.Results, CheckResult{
			Name:     ".uf/replicator/",
			Severity: Pass,
			Message:  "initialized",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        ".uf/replicator/",
			Severity:    Warn,
			Message:     "not initialized",
			InstallHint: "Run: uf init",
		})
	}

	// Check 4: MCP config — check for mcp.replicator in opencode.json.
	ocPath := filepath.Join(opts.TargetDir, "opencode.json")
	ocData, readErr := opts.ReadFile(ocPath)
	if readErr != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "MCP config",
			Severity:    Warn,
			Message:     "opencode.json not found",
			InstallHint: "Run: uf init",
		})
	} else {
		var ocMap map[string]json.RawMessage
		if jsonErr := json.Unmarshal(ocData, &ocMap); jsonErr != nil {
			group.Results = append(group.Results, CheckResult{
				Name:        "MCP config",
				Severity:    Warn,
				Message:     "opencode.json could not be parsed",
				InstallHint: "Fix JSON syntax in opencode.json",
			})
		} else {
			// Check canonical "mcp" key for replicator entry.
			found := false
			if mcpRaw, ok := ocMap["mcp"]; ok {
				var mcpMap map[string]json.RawMessage
				if json.Unmarshal(mcpRaw, &mcpMap) == nil {
					if _, hasKey := mcpMap["replicator"]; hasKey {
						found = true
					}
				}
			}
			if found {
				group.Results = append(group.Results, CheckResult{
					Name:     "MCP config",
					Severity: Pass,
					Message:  "mcp.replicator in opencode.json",
				})
			} else {
				group.Results = append(group.Results, CheckResult{
					Name:        "MCP config",
					Severity:    Warn,
					Message:     "mcp.replicator not in opencode.json",
					InstallHint: "Run: uf init",
				})
			}
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

	// Check .opencode/uf/packs/ for convention packs.
	packsDir := filepath.Join(opts.TargetDir, ".opencode", "uf", "packs")
	group.Results = append(group.Results, checkDirWithCount(packsDir, ".opencode/uf/packs/", "convention packs", ".md"))

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
				Name:     "1 skill validated",
				Severity: Pass,
				Message:  name,
			})
		}
	}

	return results
}

// defaultEmbedCheck returns a function that tests embedding
// generation by sending a POST to Ollama's /api/embed endpoint.
// Uses OLLAMA_HOST env var (default http://localhost:11434) and
// a 5-second timeout. Returns nil on success or a descriptive
// error on failure per contracts/doctor-checks.md.
func defaultEmbedCheck(getenv func(string) string) func(model string) error {
	return func(model string) error {
		host := getenv("OLLAMA_HOST")
		if host == "" {
			host = "http://localhost:11434"
		}

		url := host + "/api/embed"
		body := fmt.Sprintf(`{"model": %q, "input": "test"}`, model)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(url, "application/json", strings.NewReader(body))
		if err != nil {
			return fmt.Errorf("embed request failed: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			// Read body for error details.
			var errResp struct {
				Error string `json:"error"`
			}
			if decErr := json.NewDecoder(resp.Body).Decode(&errResp); decErr == nil && errResp.Error != "" {
				return fmt.Errorf("%s", errResp.Error)
			}
			return fmt.Errorf("embed request returned status %d", resp.StatusCode)
		}

		// Parse response to verify embeddings were generated.
		var result struct {
			Embeddings [][]float64 `json:"embeddings"`
		}
		if decErr := json.NewDecoder(resp.Body).Decode(&result); decErr != nil {
			return fmt.Errorf("could not parse embed response: %w", decErr)
		}
		if len(result.Embeddings) == 0 {
			return fmt.Errorf("empty embeddings returned")
		}

		return nil
	}
}

// checkEmbeddingCapability tests whether the embedding model can
// generate embeddings end-to-end by calling opts.EmbedCheck.
// Returns Pass on success, Warn with categorized hints on failure
// per contracts/doctor-checks.md behavior matrix.
func checkEmbeddingCapability(opts *Options) CheckResult {
	err := opts.EmbedCheck(graniteModel)
	if err == nil {
		return CheckResult{
			Name:     "embedding capability",
			Severity: Pass,
			Message:  graniteModel + " generating embeddings",
		}
	}

	errMsg := err.Error()

	// Categorize error for actionable hints.
	if strings.Contains(errMsg, "connection refused") {
		return CheckResult{
			Name:        "embedding capability",
			Severity:    Warn,
			Message:     "cannot generate embeddings (Ollama not running)",
			InstallHint: "Start Ollama: ollama serve",
		}
	}
	if strings.Contains(errMsg, "not found") {
		return CheckResult{
			Name:        "embedding capability",
			Severity:    Warn,
			Message:     "cannot generate embeddings (model not loaded)",
			InstallHint: "ollama pull " + graniteModel,
		}
	}

	// Other errors (timeout, parse failure, etc.) — combined hint.
	return CheckResult{
		Name:        "embedding capability",
		Severity:    Warn,
		Message:     "cannot generate embeddings",
		InstallHint: "Start Ollama: ollama serve, then: ollama pull " + graniteModel,
	}
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
			Name:     "embedding capability",
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
			InstallHint: "ollama pull " + graniteModel,
		})
	} else if strings.Contains(string(ollamaOutput), "granite-embedding") {
		// Annotate with Ollama demotion per US3 — Dewey manages
		// the Ollama lifecycle, so direct Ollama status is
		// informational rather than actionable.
		group.Results = append(group.Results, CheckResult{
			Name:     "embedding model",
			Severity: Pass,
			Message:  graniteModel + " installed (Dewey manages Ollama lifecycle)",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        "embedding model",
			Severity:    Warn,
			Message:     "not pulled (graph-only mode available)",
			InstallHint: "ollama pull " + graniteModel,
		})
	}

	// Check 3: embedding capability — end-to-end verification.
	group.Results = append(group.Results, checkEmbeddingCapability(opts))

	// Check 4: .uf/dewey/ workspace directory.
	deweyDir := filepath.Join(opts.TargetDir, ".uf", "dewey")
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

// agentContextSection describes a required section in AGENTS.md
// with its display name and detection patterns.
type agentContextSection struct {
	name     string
	patterns []*regexp.Regexp
}

// agentContextTier1Sections are the essential sections every
// AGENTS.md must have.
var agentContextTier1Sections = []agentContextSection{
	{
		name: "Project Overview",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^##\s+.*overview`),
			regexp.MustCompile(`(?i)^##\s+about`),
		},
	},
	{
		name: "Build Commands",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^##\s+.*build`),
			regexp.MustCompile(`(?i)^##\s+development`),
		},
	},
	{
		name: "Project Structure",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^##\s+.*structure`),
			regexp.MustCompile(`(?i)^##\s+.*layout`),
			regexp.MustCompile(`(?i)^##\s+.*directory`),
		},
	},
	{
		name: "Code Conventions",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^##\s+.*convention`),
			regexp.MustCompile(`(?i)^##\s+.*coding\s+standard`),
			regexp.MustCompile(`(?i)^##\s+.*style\s+guide`),
			regexp.MustCompile(`(?i)^##\s+.*coding\s+convention`),
		},
	},
	{
		name: "Technology Stack",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^##\s+.*technolog`),
			regexp.MustCompile(`(?i)^##\s+.*tech\s+stack`),
			regexp.MustCompile(`(?i)^##\s+.*active\s+technologies`),
		},
	},
}

// agentContextLineCountThreshold is the line count above which
// a warning is emitted suggesting the file be condensed.
const agentContextLineCountThreshold = 300

// detectAGENTSmdSections scans AGENTS.md content and returns
// which sections are present. Keys are section display names.
func detectAGENTSmdSections(content []byte) map[string]bool {
	found := make(map[string]bool)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		for _, sec := range agentContextTier1Sections {
			if found[sec.name] {
				continue
			}
			for _, p := range sec.patterns {
				if p.MatchString(line) {
					found[sec.name] = true
					break
				}
			}
		}
	}
	return found
}

// hasBuildCodeBlocks checks whether the Build section of
// AGENTS.md contains at least one fenced code block.
func hasBuildCodeBlocks(content []byte) bool {
	lines := strings.Split(string(content), "\n")
	inBuild := false
	buildPattern := regexp.MustCompile(`(?i)^##\s+.*build`)
	nextSection := regexp.MustCompile(`^##\s+`)

	for _, line := range lines {
		if !inBuild && buildPattern.MatchString(line) {
			inBuild = true
			continue
		}
		if inBuild && nextSection.MatchString(line) {
			break
		}
		if inBuild && strings.HasPrefix(strings.TrimSpace(line), "```") {
			return true
		}
	}
	return false
}

// hasSpecNumberedDirs checks whether a specs/ directory contains
// any numbered subdirectories matching the NNN-* pattern.
func hasSpecNumberedDirs(specsDir string) bool {
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return false
	}
	specDirPattern := regexp.MustCompile(`^\d{3}-`)
	for _, e := range entries {
		if e.IsDir() && specDirPattern.MatchString(e.Name()) {
			return true
		}
	}
	return false
}

// checkAgentContext validates AGENTS.md content quality with a
// context-sensitive section taxonomy. Checks file existence,
// Tier 1 section headers, build code blocks, line count,
// constitution reference, spec framework description, and
// bridge files (CLAUDE.md, .cursorrules).
func checkAgentContext(opts *Options) CheckGroup {
	group := CheckGroup{
		Name:    "Agent Context",
		Results: []CheckResult{},
	}

	// Check #1: AGENTS.md existence.
	agentsMdPath := filepath.Join(opts.TargetDir, "AGENTS.md")
	content, readErr := opts.ReadFile(agentsMdPath)
	if readErr != nil {
		group.Results = append(group.Results, CheckResult{
			Name:        "AGENTS.md",
			Severity:    Fail,
			Message:     "not found",
			InstallHint: "Run: /agent-brief in OpenCode",
		})
		return group
	}

	lineCount := strings.Count(string(content), "\n") + 1
	group.Results = append(group.Results, CheckResult{
		Name:     "AGENTS.md",
		Severity: Pass,
		Message:  fmt.Sprintf("present (%d lines)", lineCount),
	})

	// Checks #2-6: Tier 1 section presence.
	sections := detectAGENTSmdSections(content)
	for _, sec := range agentContextTier1Sections {
		if sections[sec.name] {
			group.Results = append(group.Results, CheckResult{
				Name:     "Tier 1: " + sec.name,
				Severity: Pass,
				Message:  "found",
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "Tier 1: " + sec.name,
				Severity:    Fail,
				Message:     "not found",
				InstallHint: "Run: /agent-brief in OpenCode",
			})
		}
	}

	// Check #7: Build section has code blocks.
	if sections["Build Commands"] {
		if hasBuildCodeBlocks(content) {
			group.Results = append(group.Results, CheckResult{
				Name:     "Build code blocks",
				Severity: Pass,
				Message:  "found",
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "Build code blocks",
				Severity:    Warn,
				Message:     "no code blocks in Build section",
				InstallHint: "Add fenced code blocks with build/test commands",
			})
		}
	}

	// Check #8: Line count.
	if lineCount > agentContextLineCountThreshold {
		group.Results = append(group.Results, CheckResult{
			Name:        "Line count",
			Severity:    Warn,
			Message:     fmt.Sprintf("%d lines (threshold: %d)", lineCount, agentContextLineCountThreshold),
			InstallHint: "Run: /agent-brief in OpenCode for condensing suggestions",
		})
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:     "Line count",
			Severity: Pass,
			Message:  fmt.Sprintf("%d lines", lineCount),
		})
	}

	// Check #9: Constitution reference (context-sensitive).
	constitutionPath := filepath.Join(opts.TargetDir,
		".specify", "memory", "constitution.md")
	if _, err := os.Stat(constitutionPath); err == nil {
		contentStr := strings.ToLower(string(content))
		if strings.Contains(contentStr, "constitution") {
			group.Results = append(group.Results, CheckResult{
				Name:     "Constitution reference",
				Severity: Pass,
				Message:  "found (.specify/ detected)",
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "Constitution reference",
				Severity:    Warn,
				Message:     "not referenced (.specify/ detected)",
				InstallHint: "Run: /agent-brief in OpenCode",
			})
		}
	}

	// Check #10: Spec framework description (context-sensitive).
	specsDir := filepath.Join(opts.TargetDir, "specs")
	openspecConfig := filepath.Join(opts.TargetDir, "openspec", "config.yaml")
	hasSpeckit := hasSpecNumberedDirs(specsDir)
	_, openspecErr := os.Stat(openspecConfig)
	hasOpenspec := openspecErr == nil

	if hasSpeckit || hasOpenspec {
		specPattern := regexp.MustCompile(
			`(?i)(speckit|openspec|spec\s*(ification)?\s*framework)`)
		if specPattern.Match(content) {
			group.Results = append(group.Results, CheckResult{
				Name:     "Spec framework described",
				Severity: Pass,
				Message:  "found",
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "Spec framework described",
				Severity:    Warn,
				Message:     "not described (specs/ or openspec/ detected)",
				InstallHint: "Run: /agent-brief in OpenCode",
			})
		}
	}

	// Check #11: CLAUDE.md bridge.
	claudePath := filepath.Join(opts.TargetDir, "CLAUDE.md")
	claudeContent, claudeErr := opts.ReadFile(claudePath)
	if claudeErr == nil {
		if strings.Contains(string(claudeContent), "AGENTS.md") {
			group.Results = append(group.Results, CheckResult{
				Name:     "Bridge: CLAUDE.md",
				Severity: Pass,
				Message:  "imports AGENTS.md",
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "Bridge: CLAUDE.md",
				Severity:    Warn,
				Message:     "exists but does not reference AGENTS.md",
				InstallHint: "Run: /agent-brief in OpenCode",
			})
		}
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        "Bridge: CLAUDE.md",
			Severity:    Warn,
			Message:     "not found",
			InstallHint: "Run: /agent-brief in OpenCode",
		})
	}

	// Check #12: .cursorrules bridge.
	cursorPath := filepath.Join(opts.TargetDir, ".cursorrules")
	cursorContent, cursorErr := opts.ReadFile(cursorPath)
	if cursorErr == nil {
		if strings.Contains(string(cursorContent), "AGENTS.md") {
			group.Results = append(group.Results, CheckResult{
				Name:     "Bridge: .cursorrules",
				Severity: Pass,
				Message:  "references AGENTS.md",
			})
		} else {
			group.Results = append(group.Results, CheckResult{
				Name:        "Bridge: .cursorrules",
				Severity:    Warn,
				Message:     "exists but does not reference AGENTS.md",
				InstallHint: "Run: /agent-brief in OpenCode",
			})
		}
	} else {
		group.Results = append(group.Results, CheckResult{
			Name:        "Bridge: .cursorrules",
			Severity:    Warn,
			Message:     "not found",
			InstallHint: "Run: /agent-brief in OpenCode",
		})
	}

	return group
}
