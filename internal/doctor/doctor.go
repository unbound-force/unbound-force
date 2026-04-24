package doctor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Options configures a doctor run. All external dependencies are
// injected as function fields for testability per Constitution
// Principle IV. Zero-value fields are filled with production
// implementations by defaults().
type Options struct {
	// TargetDir is the project directory to check.
	TargetDir string

	// Format is the output format: "text" or "json".
	Format string

	// Stdout is the writer for output.
	Stdout io.Writer

	// LookPath finds a binary in PATH (like exec.LookPath).
	LookPath func(string) (string, error)

	// ExecCmd runs a command and returns combined output.
	// Arguments: name, args.
	ExecCmd func(name string, args ...string) ([]byte, error)

	// ExecCmdTimeout runs a command with a timeout and returns combined
	// output. Used for subprocess calls that may hang (e.g., swarm doctor).
	// Defaults to exec.CommandContext with the given timeout.
	ExecCmdTimeout func(timeout time.Duration, name string, args ...string) ([]byte, error)

	// EvalSymlinks resolves symlinks (like filepath.EvalSymlinks).
	EvalSymlinks func(string) (string, error)

	// Getenv reads an environment variable (like os.Getenv).
	Getenv func(string) string

	// ReadFile reads a file's contents (like os.ReadFile).
	ReadFile func(string) ([]byte, error)

	// EmbedCheck tests whether the embedding model can generate
	// embeddings. Returns nil on success or an error describing
	// the failure. Injected for testability per Constitution
	// Principle IV.
	//
	// Production implementation sends a POST request to the Ollama
	// /api/embed endpoint with a minimal test input. The endpoint
	// URL is derived from OLLAMA_HOST env var (default:
	// http://localhost:11434).
	EmbedCheck func(model string) error

	// SkipChecks lists check group names to skip entirely.
	// Populated from config.Doctor.Skip. Empty means run all.
	SkipChecks []string

	// ToolSeverities maps tool names to severity overrides.
	// Populated from config.Doctor.Tools.
	ToolSeverities map[string]string

	// EmbeddingModel is the embedding model name from config.
	// Used instead of the compiled default when non-empty.
	EmbeddingModel string
}

// defaults fills zero-value fields with production implementations.
func (o *Options) defaults() {
	if o.TargetDir == "" {
		o.TargetDir, _ = os.Getwd()
	}
	if o.Format == "" {
		o.Format = "text"
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.LookPath == nil {
		o.LookPath = exec.LookPath
	}
	if o.ExecCmd == nil {
		o.ExecCmd = defaultExecCmd
	}
	if o.ExecCmdTimeout == nil {
		o.ExecCmdTimeout = defaultExecCmdTimeout
	}
	if o.EvalSymlinks == nil {
		o.EvalSymlinks = filepath.EvalSymlinks
	}
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.ReadFile == nil {
		o.ReadFile = os.ReadFile
	}
	if o.EmbedCheck == nil {
		o.EmbedCheck = defaultEmbedCheck(o.Getenv)
	}
}

// defaultExecCmd is the production implementation of ExecCmd.
func defaultExecCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// defaultExecCmdTimeout is the production implementation of ExecCmdTimeout.
// It uses exec.CommandContext with a context deadline for FR-009.
func defaultExecCmdTimeout(timeout time.Duration, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// Run executes all doctor checks and returns a Report.
// Returns an error if the platform is unsupported or if any
// check has Fail severity (for exit code 1).
func Run(opts Options) (*Report, error) {
	opts.defaults()

	// Platform guard: Windows is not supported (FR-037).
	if runtime.GOOS == "windows" {
		return nil, fmt.Errorf("platform not supported: doctor and setup require macOS or Linux")
	}

	env := DetectEnvironment(&opts)

	allGroups := []CheckGroup{
		checkDetectedEnvironment(env),
		checkCoreTools(&opts, env),
		checkReplicator(&opts),
		checkDewey(&opts),
		checkConfiguration(&opts),
		checkScaffoldedFiles(&opts),
		checkHeroAvailability(&opts),
		checkMCPConfig(&opts),
		checkAgentSkillIntegrity(&opts),
	}

	// Apply SkipChecks filter: remove check groups or
	// individual results whose name matches a skip entry.
	groups := filterSkippedChecks(allGroups, opts.SkipChecks)

	summary := computeSummary(groups)

	report := &Report{
		Environment: env,
		Groups:      groups,
		Summary:     summary,
	}

	// Return error for exit code 1 when any check failed.
	if summary.Failed > 0 {
		return report, fmt.Errorf("%d check(s) failed", summary.Failed)
	}

	return report, nil
}

// filterSkippedChecks removes check groups or individual results
// whose name matches a skip entry. Names are compared case-insensitively.
func filterSkippedChecks(groups []CheckGroup, skipChecks []string) []CheckGroup {
	if len(skipChecks) == 0 {
		return groups
	}

	skipSet := make(map[string]bool, len(skipChecks))
	for _, s := range skipChecks {
		skipSet[strings.ToLower(s)] = true
	}

	var filtered []CheckGroup
	for _, g := range groups {
		// Skip entire group if its name matches.
		if skipSet[strings.ToLower(g.Name)] {
			continue
		}

		// Filter individual results within the group.
		var results []CheckResult
		for _, r := range g.Results {
			if !skipSet[strings.ToLower(r.Name)] {
				results = append(results, r)
			}
		}
		if len(results) > 0 {
			g.Results = results
			filtered = append(filtered, g)
		}
	}
	return filtered
}

// computeSummary aggregates check result counts across all groups.
func computeSummary(groups []CheckGroup) Summary {
	var s Summary
	for _, g := range groups {
		for _, r := range g.Results {
			s.Total++
			switch r.Severity {
			case Pass:
				s.Passed++
			case Warn:
				s.Warned++
			case Fail:
				s.Failed++
			}
		}
	}
	return s
}
