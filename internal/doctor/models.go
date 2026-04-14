// Package doctor implements environment health checking for the
// Unbound Force development tool chain. It diagnoses tool
// availability, version compatibility, scaffolded file integrity,
// hero availability, MCP server configuration, and agent/skill
// frontmatter validity. Results are reported as structured data
// suitable for both colored terminal display and JSON output.
package doctor

import (
	"encoding/json"
	"fmt"
)

// Severity represents the diagnostic result severity level.
// JSON serialization uses lowercase strings: "pass", "warn", "fail".
type Severity int

const (
	// Pass indicates the check succeeded.
	Pass Severity = iota
	// Warn indicates a non-critical issue or optional item missing.
	Warn
	// Fail indicates a required item missing or invalid.
	Fail
)

// severityStrings maps Severity values to their lowercase string
// representations for JSON serialization and display.
var severityStrings = map[Severity]string{
	Pass: "pass",
	Warn: "warn",
	Fail: "fail",
}

// strToSeverity maps lowercase strings to Severity values for
// JSON deserialization.
var strToSeverity = map[string]Severity{
	"pass": Pass,
	"warn": Warn,
	"fail": Fail,
}

// String returns the lowercase string representation of the severity.
func (s Severity) String() string {
	if str, ok := severityStrings[s]; ok {
		return str
	}
	return "unknown"
}

// MarshalJSON serializes Severity as a lowercase JSON string.
func (s Severity) MarshalJSON() ([]byte, error) {
	str, ok := severityStrings[s]
	if !ok {
		return nil, fmt.Errorf("unknown severity value: %d", s)
	}
	return json.Marshal(str)
}

// UnmarshalJSON deserializes a lowercase JSON string to Severity.
func (s *Severity) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	val, ok := strToSeverity[str]
	if !ok {
		return fmt.Errorf("unknown severity string: %q", str)
	}
	*s = val
	return nil
}

// ManagerKind identifies a version or package manager.
type ManagerKind string

const (
	ManagerGoenv    ManagerKind = "goenv"
	ManagerNvm      ManagerKind = "nvm"
	ManagerFnm      ManagerKind = "fnm"
	ManagerPyenv    ManagerKind = "pyenv"
	ManagerMise     ManagerKind = "mise"
	ManagerHomebrew ManagerKind = "homebrew"
	ManagerDnf      ManagerKind = "dnf"
	ManagerBun      ManagerKind = "bun"
	ManagerSystem   ManagerKind = "system"
	ManagerDirect   ManagerKind = "direct"
	ManagerUnknown  ManagerKind = "unknown"
)

// ManagerInfo describes a detected version or package manager.
type ManagerInfo struct {
	Kind    ManagerKind `json:"kind"`
	Path    string      `json:"path"`
	Manages []string    `json:"manages"`
}

// DetectedEnvironment captures the developer's tool management
// landscape, detected once at startup and shared between doctor
// and setup.
type DetectedEnvironment struct {
	Managers []ManagerInfo `json:"managers"`
	Platform string        `json:"platform"`
}

// ToolProvenance records how a specific tool binary was installed.
// This type is internal-only — provenance information is encoded
// in CheckResult.Message and CheckResult.Detail for output.
type ToolProvenance struct {
	Manager ManagerKind
	Version string
	Path    string
}

// CheckResult represents a single diagnostic finding.
type CheckResult struct {
	Name        string   `json:"name"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
	Detail      string   `json:"detail,omitempty"`
	InstallHint string   `json:"install_hint,omitempty"`
	InstallURL  string   `json:"install_url,omitempty"`
}

// CheckGroup is a named collection of related check results.
type CheckGroup struct {
	Name    string        `json:"name"`
	Results []CheckResult `json:"results"`
	Embed   string        `json:"embed,omitempty"`
}

// Summary provides aggregate counts of check results by severity.
type Summary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Warned int `json:"warned"`
	Failed int `json:"failed"`
}

// Report is the complete diagnostic output from a doctor run.
type Report struct {
	Environment DetectedEnvironment `json:"environment"`
	Groups      []CheckGroup        `json:"groups"`
	Summary     Summary             `json:"summary"`
}
