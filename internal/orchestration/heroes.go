package orchestration

import (
	"os"
	"path/filepath"
	"strings"
)

// heroSpec defines how to detect a hero's availability.
type heroSpec struct {
	Name      string
	Role      string
	AgentFile string // relative to agentDir; empty if not agent-based
	Binary    string // binary name for exec.LookPath; empty if not binary-based
}

// heroSpecs is the canonical list of heroes and their detection methods.
// Per research.md R4: agent files in .opencode/agents/ and binaries in PATH.
var heroSpecs = []heroSpec{
	{Name: "muti-mind", Role: StageDefine, AgentFile: "muti-mind-po.md"},
	{Name: "cobalt-crush", Role: StageImplement, AgentFile: "cobalt-crush-dev.md"},
	{Name: "gaze", Role: StageValidate, Binary: "gaze"},
	{Name: "divisor", Role: StageReview, AgentFile: "divisor-guard.md"}, // any divisor-*.md
	{Name: "muti-mind", Role: StageAccept, AgentFile: "muti-mind-po.md"},
	{Name: "mx-f", Role: StageMeasure, AgentFile: "mx-f-coach.md", Binary: "mxf"},
}

// DetectHeroes checks for hero availability by looking for agent files
// in agentDir and binaries via the injected lookPath function.
// lookPath should have the same signature as exec.LookPath — it returns
// the path to the binary or an error if not found.
func DetectHeroes(agentDir string, lookPath func(string) (string, error)) ([]HeroStatus, error) {
	seen := make(map[string]bool)
	var heroes []HeroStatus

	for _, spec := range heroSpecs {
		// Deduplicate heroes that appear in multiple stages (muti-mind)
		if seen[spec.Name] {
			continue
		}
		seen[spec.Name] = true

		status := HeroStatus{
			Name:      spec.Name,
			Role:      spec.Role,
			AgentFile: spec.AgentFile,
		}

		switch {
		case spec.Name == "divisor":
			// Divisor: check for any divisor-*.md file
			status.Available = hasDivisorAgent(agentDir)
			status.DetectionMethod = "file_exists"
		case spec.AgentFile != "" && spec.Binary != "":
			// Mx F: needs both agent file and binary
			agentExists := fileExists(filepath.Join(agentDir, spec.AgentFile))
			_, binErr := lookPath(spec.Binary)
			status.Available = agentExists || binErr == nil
			if binErr == nil {
				status.DetectionMethod = "exec_lookpath"
			} else {
				status.DetectionMethod = "file_exists"
			}
		case spec.AgentFile != "":
			// Agent-only heroes (muti-mind, cobalt-crush)
			status.Available = fileExists(filepath.Join(agentDir, spec.AgentFile))
			status.DetectionMethod = "file_exists"
		case spec.Binary != "":
			// Binary-only heroes (gaze)
			_, err := lookPath(spec.Binary)
			status.Available = err == nil
			status.DetectionMethod = "exec_lookpath"
		}

		heroes = append(heroes, status)
	}

	return heroes, nil
}

// StageHeroMap returns a mapping from stage names to hero names.
func StageHeroMap() map[string]string {
	return map[string]string{
		StageDefine:    "muti-mind",
		StageImplement: "cobalt-crush",
		StageValidate:  "gaze",
		StageReview:    "divisor",
		StageAccept:    "muti-mind",
		StageMeasure:   "mx-f",
	}
}

// hasDivisorAgent checks if any divisor-*.md file exists in the agent directory.
func hasDivisorAgent(agentDir string) bool {
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "divisor-") && strings.HasSuffix(entry.Name(), ".md") {
			return true
		}
	}
	return false
}

// fileExists returns true if the path exists and is a regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
