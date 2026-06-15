// Package setup implements automated tool chain installation for
// the Unbound Force development environment. It detects existing
// version managers, installs missing tools through the appropriate
// manager, configures Replicator, and scaffolds project files.
// All external dependencies are injected for testability per
// Constitution Principle IV.
package setup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/unbound-force/unbound-force/internal/config"
	"github.com/unbound-force/unbound-force/internal/doctor"
)

// Options configures a setup run. All external dependencies are
// injected as function fields for testability.
type Options struct {
	// TargetDir is the project directory to set up.
	TargetDir string

	// DryRun prints what would be done without executing.
	DryRun bool

	// YesFlag skips curl|bash confirmation prompts.
	YesFlag bool

	// IsTTY returns whether stdout is a terminal (for interactive prompts).
	IsTTY func() bool

	// Stdout is the writer for output.
	Stdout io.Writer

	// Stderr is the writer for progress messages.
	Stderr io.Writer

	// LookPath finds a binary in PATH.
	LookPath func(string) (string, error)

	// ExecCmd runs a command and returns combined output.
	ExecCmd func(name string, args ...string) ([]byte, error)

	// EvalSymlinks resolves symlinks.
	EvalSymlinks func(string) (string, error)

	// Getenv reads an environment variable.
	Getenv func(string) string

	// ReadFile reads a file's contents.
	ReadFile func(string) ([]byte, error)

	// WriteFile writes data to a file atomically.
	WriteFile func(string, []byte, os.FileMode) error

	// GOOS overrides the detected operating system for testability.
	// Defaults to runtime.GOOS when empty.
	GOOS string

	// Version is the current binary version (e.g., "0.12.0"),
	// used to construct GitHub Release RPM URLs. Set by the CLI
	// from the build-time version variable.
	Version string

	// PackageManager is the preferred package manager from config.
	// Valid: "auto", "homebrew", "dnf", "apt", "manual".
	PackageManager string

	// SkipTools lists tool names to skip during setup.
	SkipTools []string

	// ToolMethods provides per-tool install method overrides from config.
	ToolMethods map[string]config.ToolConfig

	// EmbeddingModel is the embedding model name from config.
	// Defaults to "granite-embedding:30m".
	EmbeddingModel string

	// EmbeddingDimensions is the embedding vector dimension from config.
	// Defaults to 256.
	EmbeddingDimensions int
}

// defaults fills zero-value fields with production implementations.
func (o *Options) defaults() {
	if o.TargetDir == "" {
		o.TargetDir, _ = os.Getwd()
	}
	if o.Stdout == nil {
		o.Stdout = os.Stdout
	}
	if o.Stderr == nil {
		o.Stderr = os.Stderr
	}
	if o.LookPath == nil {
		o.LookPath = exec.LookPath
	}
	if o.ExecCmd == nil {
		o.ExecCmd = defaultExecCmd
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
	if o.WriteFile == nil {
		o.WriteFile = atomicWriteFile
	}
	if o.IsTTY == nil {
		o.IsTTY = func() bool { return false }
	}
	if o.GOOS == "" {
		o.GOOS = runtime.GOOS
	}
}

// defaultExecCmd is the production implementation of ExecCmd.
func defaultExecCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// stepResult tracks the outcome of a setup step.
// The output field captures stdout+stderr from CombinedOutput()
// for diagnostic reporting when a step fails.
type stepResult struct {
	name   string
	action string // "installed", "already installed", "skipped", "failed"
	detail string
	err    error
	output []byte // captured stdout+stderr from CombinedOutput()
}

// Default embedding model constants — used when config does not
// override. IBM Granite, Apache 2.0, permissibly licensed training data.
const (
	defaultEmbeddingModel = "granite-embedding:30m"
	defaultEmbeddingDim   = "256"
)

// embeddingModel returns the configured or default embedding model name.
func (o *Options) embeddingModel() string {
	if o.EmbeddingModel != "" {
		return o.EmbeddingModel
	}
	return defaultEmbeddingModel
}

// embeddingDim returns the configured or default embedding dimension as a string.
func (o *Options) embeddingDim() string {
	if o.EmbeddingDimensions > 0 {
		return strconv.Itoa(o.EmbeddingDimensions)
	}
	return defaultEmbeddingDim
}

// shouldSkipTool returns true if the tool should be skipped
// based on the config skip list or per-tool method override.
func (o *Options) shouldSkipTool(toolName string) bool {
	for _, s := range o.SkipTools {
		if s == toolName {
			return true
		}
	}
	if o.ToolMethods != nil {
		if tc, ok := o.ToolMethods[toolName]; ok && tc.Method == "skip" {
			return true
		}
	}
	if o.PackageManager == "manual" {
		// In manual mode, skip tools with auto method (no per-tool override).
		if o.ToolMethods == nil {
			return true
		}
		if tc, ok := o.ToolMethods[toolName]; !ok || tc.Method == "" || tc.Method == "auto" {
			return true
		}
	}
	return false
}

// toolMethod returns the configured install method for a tool,
// or "auto" if no override is set.
func (o *Options) toolMethod(toolName string) string {
	if o.ToolMethods != nil {
		if tc, ok := o.ToolMethods[toolName]; ok && tc.Method != "" {
			return tc.Method
		}
	}
	return "auto"
}

// resolveMethod resolves the effective install method for a tool by
// applying the precedence chain: per-tool override > global
// PackageManager preference > environment-aware auto detection.
// The fallbacks parameter defines the tool's preferred install
// methods in priority order (e.g., ["homebrew", "dnf", "go"]).
// For auto mode, resolveMethod returns the first fallback whose
// package manager is detected in the environment. This eliminates
// duplicated auto-mode fallback blocks in each install function.
func (o *Options) resolveMethod(toolName string, env doctor.DetectedEnvironment, fallbacks ...string) string {
	// (1) Per-tool override takes precedence.
	if method := o.toolMethod(toolName); method != "auto" {
		return method
	}

	// (2) Global PackageManager preference.
	switch o.PackageManager {
	case "homebrew", "dnf":
		return o.PackageManager
	case "apt":
		// apt support not yet implemented — inform the user and
		// fall through to auto-mode (D1, D5).
		fmt.Fprintf(o.Stderr, "apt support not yet implemented, using auto-mode fallback\n")
	}

	// (3) Auto-mode: return the first fallback whose manager is
	// detected in the environment.
	managerMap := map[string]doctor.ManagerKind{
		"homebrew": doctor.ManagerHomebrew,
		"dnf":      doctor.ManagerDnf,
	}
	for _, fb := range fallbacks {
		mgr, needsManager := managerMap[fb]
		if !needsManager {
			// Methods like "go" don't need a package manager —
			// they're always available as a fallback.
			return fb
		}
		if doctor.HasManager(env, mgr) {
			return fb
		}
	}

	return "skip"
}

// installViaBrew installs a tool using Homebrew. Returns a stepResult
// with the outcome. Uses opts.ExecCmd for testability per Constitution
// Principle IV.
func installViaBrew(opts *Options, toolName string, brewFormula string) stepResult {
	if opts.DryRun {
		return stepResult{
			name:   toolName,
			action: "dry-run",
			detail: "Would install: brew install " + brewFormula,
		}
	}
	if out, err := opts.ExecCmd("brew", "install", brewFormula); err != nil {
		return stepResult{
			name:   toolName,
			action: "failed",
			detail: "brew install failed",
			err:    fmt.Errorf("brew install %s: %w", brewFormula, err),
			output: out,
		}
	}
	return stepResult{name: toolName, action: "installed", detail: "via Homebrew"}
}

// installViaGo installs a Go binary tool using `go install`. Returns
// a stepResult with the outcome. Uses opts.LookPath and opts.ExecCmd
// for testability per Constitution Principle IV (D3).
func installViaGo(opts *Options, toolName, goModule string) stepResult {
	// Check if Go is available in PATH.
	if _, err := opts.LookPath("go"); err != nil {
		return stepResult{
			name:   toolName,
			action: "skipped",
			detail: "Go not available. Install Go or use Homebrew/dnf.",
		}
	}

	if opts.DryRun {
		return stepResult{
			name:   toolName,
			action: "dry-run",
			detail: "Would install: go install " + goModule + "@latest",
		}
	}

	if out, err := opts.ExecCmd("go", "install", goModule+"@latest"); err != nil {
		return stepResult{
			name:   toolName,
			action: "failed",
			detail: "go install failed — try: go install " + goModule + "@latest",
			err:    fmt.Errorf("go install %s: %w", goModule, err),
			output: out,
		}
	}
	return stepResult{name: toolName, action: "installed", detail: "via go install"}
}

// stepDef defines a single setup step for data-driven dispatch.
// Each step has a display name, a skip-tool key, an install function,
// and optional gate and effect callbacks. This struct replaces the
// repetitive if/else blocks in Run(), reducing cyclomatic complexity
// per CS-004 (DRY) and CS-010 (single responsibility).
type stepDef struct {
	// name is the display name shown in progress output (e.g., "OpenCode").
	name string

	// tool is the key passed to shouldSkipTool (e.g., "opencode").
	// Empty string means the step is never skipped by tool config.
	tool string

	// install executes the step and returns its result.
	install func(*Options, doctor.DetectedEnvironment) stepResult

	// gate returns true if the step should run. When non-nil and
	// returning false, the step is skipped with "prerequisite not met".
	gate func() bool

	// gateDetail overrides the default "prerequisite not met" skip
	// detail when the gate returns false.
	gateDetail string

	// effect is called after a successful install to update mutable
	// state (e.g., setting nodeAvailable based on the result).
	effect func(stepResult)
}

// Run executes the full setup workflow per FR-021/030/032/034/035.
func Run(opts Options) error {
	opts.defaults()

	// Platform guard: Windows is not supported (FR-037).
	if runtime.GOOS == "windows" {
		return fmt.Errorf("platform not supported: doctor and setup require macOS or Linux")
	}

	// Set Ollama env vars so all embedding consumers use the same
	// embedding model. These are inherited by child processes
	// (replicator setup, dewey serve). Values come from config
	// or compiled defaults.
	_ = os.Setenv("OLLAMA_MODEL", opts.embeddingModel())
	_ = os.Setenv("OLLAMA_EMBED_DIM", opts.embeddingDim())

	// Detect environment (reuse from doctor package).
	doctorOpts := &doctor.Options{
		TargetDir:    opts.TargetDir,
		LookPath:     opts.LookPath,
		ExecCmd:      opts.ExecCmd,
		EvalSymlinks: opts.EvalSymlinks,
		Getenv:       opts.Getenv,
		ReadFile:     opts.ReadFile,
	}
	env := doctor.DetectEnvironment(doctorOpts)

	printHeader(&opts, env)

	// Mutable state tracked across steps — gate functions close
	// over these variables to enforce inter-step dependencies.
	var nodeAvailable, uvAvailable bool
	replicatorAvailable := true

	// Define all 16 steps as data. Order matters — later steps
	// may gate on state set by earlier steps' effect callbacks.
	steps := buildSteps(&opts, &nodeAvailable, &uvAvailable, &replicatorAvailable)

	results := executeSteps(&opts, env, steps)

	return printSummary(&opts, results)
}

// buildSteps constructs the ordered step definitions. Gate and effect
// closures capture mutable state pointers so inter-step dependencies
// are tracked without branching in Run().
func buildSteps(opts *Options, nodeAvailable, uvAvailable, replicatorAvailable *bool) []stepDef {
	return []stepDef{
		{name: "OpenCode", tool: "opencode", install: installOpenCode},
		{name: "Gaze", tool: "gaze", install: installGaze},
		{name: "GitHub CLI", tool: "gh", install: installGH},
		{
			name: "Node.js", tool: "node", install: ensureNodeJS,
			effect: func(r stepResult) {
				*nodeAvailable = r.err == nil && r.action != "failed"
			},
		},
		{
			name: "OpenSpec CLI", tool: "openspec", install: installOpenSpec,
			gate:       func() bool { return *nodeAvailable },
			gateDetail: "no Node.js",
		},
		{
			name: "uv", tool: "uv", install: installUV,
			effect: func(r stepResult) {
				*uvAvailable = r.err == nil && r.action != "failed"
			},
		},
		{
			name: "Specify CLI", tool: "specify", install: installSpecify,
			gate:       func() bool { return *uvAvailable },
			gateDetail: "no uv",
		},
		{
			name: "Replicator", tool: "replicator", install: installReplicator,
			effect: func(r stepResult) {
				*replicatorAvailable = r.err == nil && r.action != "failed" && r.action != "skipped"
			},
		},
		{
			name: "replicator setup",
			// No tool key — skip logic is handled entirely by the gate.
			install: func(o *Options, _ doctor.DetectedEnvironment) stepResult {
				return runReplicatorSetup(o)
			},
			gate:       func() bool { return *replicatorAvailable },
			gateDetail: "no replicator",
		},
		{name: "Ollama", tool: "ollama", install: installOllama},
		{name: "Podman", tool: "podman", install: installPodman},
		{name: "DevPod", tool: "devpod", install: installDevPod},
		{
			name: "DevPod provider",
			install: func(o *Options, _ doctor.DetectedEnvironment) stepResult {
				return configureDevPodProvider(o)
			},
		},
		{name: "Dewey", tool: "dewey", install: installDewey},
		{name: "golangci-lint", tool: "golangci-lint", install: installGolangciLint},
		{name: "govulncheck", tool: "govulncheck", install: installGovulncheck},
	}
}

// executeSteps iterates through step definitions, applying skip/gate
// logic and collecting results. This is the core dispatch loop that
// replaces the 16 repetitive if/else blocks.
func executeSteps(opts *Options, env doctor.DetectedEnvironment, steps []stepDef) []stepResult {
	total := len(steps)
	results := make([]stepResult, 0, total)

	for i, step := range steps {
		fmt.Fprintf(opts.Stdout, "  [%d/%d] %s...\n", i+1, total, step.name)

		// Check tool skip list first.
		if step.tool != "" && opts.shouldSkipTool(step.tool) {
			r := stepResult{name: step.name, action: "skipped", detail: "excluded by config"}
			results = append(results, r)
			if step.effect != nil {
				step.effect(r)
			}
			continue
		}

		// Check prerequisite gate.
		if step.gate != nil && !step.gate() {
			detail := "prerequisite not met"
			if step.gateDetail != "" {
				detail = step.gateDetail
			}
			results = append(results, stepResult{name: step.name, action: "skipped", detail: detail})
			continue
		}

		r := step.install(opts, env)
		results = append(results, r)

		if step.effect != nil {
			step.effect(r)
		}
	}

	return results
}

// printHeader writes the setup banner and detected environment.
func printHeader(opts *Options, env doctor.DetectedEnvironment) {
	fmt.Fprintln(opts.Stdout, "Unbound Force Setup")
	fmt.Fprintln(opts.Stdout, "===================")
	fmt.Fprintln(opts.Stdout)

	fmt.Fprintln(opts.Stdout, "Detected Environment")
	if len(env.Managers) > 0 {
		var parts []string
		for _, m := range env.Managers {
			parts = append(parts, fmt.Sprintf("  %s (%s)", m.Kind, strings.Join(m.Manages, ", ")))
		}
		fmt.Fprintln(opts.Stdout, strings.Join(parts, "\n"))
	} else {
		fmt.Fprintln(opts.Stdout, "  No version managers detected")
	}
	fmt.Fprintln(opts.Stdout)

	if opts.DryRun {
		fmt.Fprintln(opts.Stdout, "Dry run mode — no changes will be made")
		fmt.Fprintln(opts.Stdout)
	}

	fmt.Fprintln(opts.Stdout, "Installing...")
}

// printSummary writes step results and the completion message.
// Returns an error if any steps failed.
func printSummary(opts *Options, results []stepResult) error {
	for _, r := range results {
		printStepResult(opts.Stdout, r)
	}

	fmt.Fprintln(opts.Stdout)

	failCount := 0
	for _, r := range results {
		if r.action == "failed" {
			failCount++
		}
	}

	if failCount > 0 {
		fmt.Fprintln(opts.Stdout, "Setup partially complete. Fix the errors above, then re-run `uf setup`.")
		return fmt.Errorf("%d step(s) failed", failCount)
	}

	fmt.Fprintln(opts.Stdout, "Setup complete! Run `uf doctor` to verify.")

	// Embedding model alignment note.
	fmt.Fprintln(opts.Stdout)
	fmt.Fprintln(opts.Stdout, "Note: Replicator and Dewey are configured to use "+opts.embeddingModel()+".")
	fmt.Fprintln(opts.Stdout, "  Add to your shell profile for consistent behavior:")
	fmt.Fprintln(opts.Stdout, "  export OLLAMA_MODEL="+opts.embeddingModel())
	fmt.Fprintln(opts.Stdout, "  export OLLAMA_EMBED_DIM="+opts.embeddingDim())

	return nil
}

// installOpenCode installs OpenCode if missing per FR-022/FR-036.
func installOpenCode(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("opencode"); err == nil {
		return stepResult{name: "OpenCode", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "OpenCode", action: "dry-run", detail: "Would install: brew install anomalyco/tap/opencode"}
		}
		return stepResult{name: "OpenCode", action: "dry-run", detail: "Would install: curl -fsSL https://opencode.ai/install | bash"}
	}

	// Try Homebrew first.
	if doctor.HasManager(env, doctor.ManagerHomebrew) {
		if out, err := opts.ExecCmd("brew", "install", "anomalyco/tap/opencode"); err != nil {
			return stepResult{name: "OpenCode", action: "failed", detail: "brew install failed", err: err, output: out}
		}
		return stepResult{name: "OpenCode", action: "installed", detail: "via Homebrew"}
	}

	// Fallback to curl|bash — requires --yes or TTY confirmation (FR-036).
	if !opts.YesFlag && !opts.IsTTY() {
		return stepResult{
			name:   "OpenCode",
			action: "skipped",
			detail: "curl|bash install requires --yes flag or interactive terminal",
		}
	}

	if out, err := opts.ExecCmd("bash", "-c", "curl -fsSL https://opencode.ai/install | bash"); err != nil {
		return stepResult{name: "OpenCode", action: "failed", detail: "curl install failed", err: err, output: out}
	}
	return stepResult{name: "OpenCode", action: "installed", detail: "via curl"}
}

// installGH installs the GitHub CLI if missing.
// Fallback chain (D2): Homebrew → dnf → skip (no go install — GH
// CLI is not a simple Go binary). GH CLI has its own dnf repo,
// so dnf install uses the package name directly, not installViaRpm (D4).
func installGH(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("gh"); err == nil {
		return stepResult{name: "GitHub CLI", action: "already installed"}
	}

	// Fallback chain (D2): Homebrew → dnf → skip.
	method := opts.resolveMethod("gh", env, "homebrew", "dnf")
	switch method {
	case "homebrew":
		return installViaBrew(opts, "GitHub CLI", "gh")
	case "dnf":
		return installGHViaDnf(opts)
	default:
		return ghSkipResult(opts)
	}
}

// installGHViaDnf installs GitHub CLI via dnf. On failure, returns a
// graceful skip (not a hard failure) with an actionable repo setup
// link (D4). GH CLI uses its own dnf repository, not GoReleaser RPMs.
func installGHViaDnf(opts *Options) stepResult {
	if opts.DryRun {
		return stepResult{name: "GitHub CLI", action: "dry-run", detail: "Would install: dnf install -y gh"}
	}
	if out, err := opts.ExecCmd("dnf", "install", "-y", "gh"); err != nil {
		_ = out // dnf gh failure is a graceful skip, not a hard failure
		return stepResult{
			name:   "GitHub CLI",
			action: "skipped",
			detail: "dnf install failed — configure the GitHub CLI repo: https://github.com/cli/cli/blob/trunk/docs/install_linux.md or download from https://cli.github.com",
		}
	}
	return stepResult{name: "GitHub CLI", action: "installed", detail: "via dnf"}
}

// ghSkipResult returns the appropriate skip/dry-run result when no
// package manager is available for GitHub CLI.
func ghSkipResult(opts *Options) stepResult {
	if opts.DryRun {
		return stepResult{
			name:   "GitHub CLI",
			action: "dry-run",
			detail: "Would install: download from https://cli.github.com",
		}
	}
	return stepResult{
		name:   "GitHub CLI",
		action: "skipped",
		detail: "Homebrew not available. Download from https://cli.github.com",
	}
}

// installOpenSpec installs the OpenSpec CLI if missing.
// Uses npm as the sole installation method (FR-004).
func installOpenSpec(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("openspec"); err == nil {
		return stepResult{name: "OpenSpec CLI", action: "already installed"}
	}

	if opts.DryRun {
		return stepResult{name: "OpenSpec CLI", action: "dry-run", detail: "Would install: npm install -g @fission-ai/openspec@latest"}
	}

	if out, err := opts.ExecCmd("npm", "install", "-g", "@fission-ai/openspec@latest"); err != nil {
		return stepResult{
			name:   "OpenSpec CLI",
			action: "failed",
			detail: "npm install failed — fix npm permissions (see https://docs.npmjs.com/resolving-eacces-permissions-errors)",
			err:    err,
			output: out,
		}
	}
	return stepResult{name: "OpenSpec CLI", action: "installed", detail: "via npm"}
}

// installGaze installs Gaze if missing per FR-023.
// Fallback chain (D2): Homebrew → dnf (RPM) → go install → skip.
func installGaze(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("gaze"); err == nil {
		return stepResult{name: "Gaze", action: "already installed"}
	}

	// Fallback chain (D2): Homebrew → dnf → go install → skip.
	skipResult := stepResult{
		name:   "Gaze",
		action: "skipped",
		detail: "Homebrew not available. Download from https://github.com/unbound-force/gaze/releases",
	}
	method := opts.resolveMethod("gaze", env, "homebrew", "dnf", "go")
	switch method {
	case "homebrew":
		return installViaBrew(opts, "Gaze", "unbound-force/tap/gaze")
	case "rpm", "dnf":
		return installViaRpm(opts, "Gaze", "unbound-force/gaze", opts.Version)
	case "go":
		result := installViaGo(opts, "Gaze", "github.com/unbound-force/gaze/cmd/gaze")
		if result.action != "skipped" {
			return result
		}
		return skipResult
	default:
		return skipResult
	}
}

// ensureNodeJS checks for Node.js >= 18 and installs if needed per FR-024.
func ensureNodeJS(opts *Options, env doctor.DetectedEnvironment) stepResult {
	// Check if node is already available.
	if _, err := opts.LookPath("node"); err == nil {
		output, execErr := opts.ExecCmd("node", "--version")
		if execErr == nil {
			version := strings.TrimSpace(strings.TrimPrefix(string(output), "v"))
			// Verify version >= 18 per FR-024.
			if major, parseErr := parseNodeMajor(version); parseErr == nil {
				if major < 18 {
					// Node.js found but too old -- attempt upgrade.
					return installNodeJS(opts, env, fmt.Sprintf("version %s is below minimum 18", version))
				}
			}
			return stepResult{name: "Node.js", action: "already installed", detail: version}
		}
	}

	// Node.js not found in PATH — attempt install.
	return installNodeJS(opts, env, "not found")
}

// parseNodeMajor extracts the major version number from a Node.js version string.
// Accepts formats like "22.15.0" or "22".
func parseNodeMajor(version string) (int, error) {
	parts := strings.SplitN(version, ".", 2)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty version string")
	}
	return strconv.Atoi(parts[0])
}

// installNodeJS attempts to install Node.js through detected managers.
// Called when Node.js is either missing or below the minimum version.
func installNodeJS(opts *Options, env doctor.DetectedEnvironment, reason string) stepResult {
	if opts.DryRun {
		nvmDir := opts.Getenv("NVM_DIR")
		if nvmDir != "" {
			return stepResult{name: "Node.js", action: "dry-run", detail: fmt.Sprintf("%s. Would install: nvm install 22", reason)}
		}
		if doctor.HasManager(env, doctor.ManagerFnm) {
			return stepResult{name: "Node.js", action: "dry-run", detail: fmt.Sprintf("%s. Would install: fnm install 22", reason)}
		}
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Node.js", action: "dry-run", detail: fmt.Sprintf("%s. Would install: brew install node", reason)}
		}
		return stepResult{name: "Node.js", action: "dry-run", detail: fmt.Sprintf("%s. No Node.js manager detected", reason)}
	}

	// Try nvm first (bash function, not binary).
	nvmDir := opts.Getenv("NVM_DIR")
	if nvmDir != "" {
		cmd := fmt.Sprintf("source %s/nvm.sh && nvm install 22", nvmDir)
		if out, err := opts.ExecCmd("bash", "-c", cmd); err != nil {
			_ = out // nvm failure falls through to next manager
			fmt.Fprintf(opts.Stderr, "nvm install failed: %v\n", err)
			fmt.Fprintf(opts.Stderr, "Manual install: source %s/nvm.sh && nvm install 22\n", nvmDir)
		} else {
			return stepResult{name: "Node.js", action: "installed", detail: "via nvm"}
		}
	}

	// Try fnm.
	if doctor.HasManager(env, doctor.ManagerFnm) {
		if out, err := opts.ExecCmd("fnm", "install", "22"); err != nil {
			return stepResult{name: "Node.js", action: "failed", detail: "fnm install failed", err: err, output: out}
		}
		return stepResult{name: "Node.js", action: "installed", detail: "via fnm"}
	}

	// Try Homebrew.
	if doctor.HasManager(env, doctor.ManagerHomebrew) {
		if out, err := opts.ExecCmd("brew", "install", "node"); err != nil {
			return stepResult{name: "Node.js", action: "failed", detail: "brew install failed", err: err, output: out}
		}
		return stepResult{name: "Node.js", action: "installed", detail: "via Homebrew"}
	}

	return stepResult{
		name:   "Node.js",
		action: "failed",
		detail: fmt.Sprintf("%s. Install: brew install node or https://nodejs.org/", reason),
		err:    fmt.Errorf("node.js not available"),
	}
}

// installUV installs the uv Python package manager if missing.
// Follows the installOpenCode() pattern: Homebrew-first with curl
// fallback and interactive guard for curl|bash.
func installUV(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("uv"); err == nil {
		return stepResult{name: "uv", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "uv", action: "dry-run", detail: "Would install: brew install uv"}
		}
		return stepResult{name: "uv", action: "dry-run", detail: "Would install: curl -LsSf https://astral.sh/uv/install.sh | sh"}
	}

	// Try Homebrew first.
	if doctor.HasManager(env, doctor.ManagerHomebrew) {
		if out, err := opts.ExecCmd("brew", "install", "uv"); err != nil {
			return stepResult{name: "uv", action: "failed", detail: "brew install failed", err: err, output: out}
		}
		return stepResult{name: "uv", action: "installed", detail: "via Homebrew"}
	}

	// Fallback to curl|bash — requires --yes or TTY confirmation.
	if !opts.YesFlag && !opts.IsTTY() {
		return stepResult{
			name:   "uv",
			action: "skipped",
			detail: "curl|bash install requires --yes flag or interactive terminal",
		}
	}

	if out, err := opts.ExecCmd("bash", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh"); err != nil {
		return stepResult{name: "uv", action: "failed", detail: "curl install failed", err: err, output: out}
	}
	return stepResult{name: "uv", action: "installed", detail: "via curl"}
}

// installSpecify installs the Specify CLI via uv tool install.
// Gated by uv availability — if uv is not in PATH, the step is
// skipped. Follows the installOpenSpec() pattern (single install
// method, gated by package manager availability).
func installSpecify(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("specify"); err == nil {
		return stepResult{name: "Specify CLI", action: "already installed"}
	}

	if opts.DryRun {
		return stepResult{name: "Specify CLI", action: "dry-run", detail: "Would install: uv tool install specify-cli"}
	}

	// Check uv availability.
	if _, err := opts.LookPath("uv"); err != nil {
		return stepResult{
			name:   "Specify CLI",
			action: "skipped",
			detail: "uv not available — install uv first",
		}
	}

	if out, err := opts.ExecCmd("uv", "tool", "install", "specify-cli"); err != nil {
		return stepResult{
			name:   "Specify CLI",
			action: "failed",
			detail: "uv tool install failed — try: uv tool install specify-cli",
			err:    err,
			output: out,
		}
	}
	return stepResult{name: "Specify CLI", action: "installed", detail: "via uv"}
}

// installReplicator installs Replicator if missing per FR-001.
// Fallback chain (D2): Homebrew → dnf (RPM) → go install → skip.
func installReplicator(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("replicator"); err == nil {
		return stepResult{name: "Replicator", action: "already installed"}
	}

	// Fallback chain (D2): Homebrew → dnf → go install → skip.
	skipResult := stepResult{
		name:   "Replicator",
		action: "skipped",
		detail: "Homebrew not available. Download from https://github.com/unbound-force/replicator/releases",
	}
	method := opts.resolveMethod("replicator", env, "homebrew", "dnf", "go")
	switch method {
	case "homebrew":
		return installViaBrew(opts, "Replicator", "unbound-force/tap/replicator")
	case "rpm", "dnf":
		return installViaRpm(opts, "Replicator", "unbound-force/replicator", opts.Version)
	case "go":
		result := installViaGo(opts, "Replicator", "github.com/unbound-force/replicator/cmd/replicator")
		if result.action != "skipped" {
			return result
		}
		return skipResult
	default:
		return skipResult
	}
}

// runReplicatorSetup runs `replicator setup` per FR-002.
// Interactive guard prevents unattended execution.
func runReplicatorSetup(opts *Options) stepResult {
	if opts.DryRun {
		return stepResult{name: "replicator setup", action: "dry-run", detail: "Would run: replicator setup"}
	}

	if !opts.YesFlag && !opts.IsTTY() {
		return stepResult{
			name:   "replicator setup",
			action: "skipped",
			detail: "interactive — run `replicator setup` manually or use --yes",
		}
	}

	if out, err := opts.ExecCmd("replicator", "setup"); err != nil {
		return stepResult{name: "replicator setup", action: "failed", detail: "replicator setup failed", err: err, output: out}
	}
	return stepResult{name: "replicator setup", action: "completed"}
}

// installGolangciLint installs golangci-lint if missing per Spec 019
// FR-012. Uses go install as primary method with Homebrew fallback.
func installGolangciLint(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("golangci-lint"); err == nil {
		return stepResult{name: "golangci-lint", action: "already installed"}
	}

	if opts.DryRun {
		return stepResult{name: "golangci-lint", action: "dry-run", detail: "Would install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"}
	}

	// Try go install first (Go is already a prerequisite).
	if _, err := opts.ExecCmd("go", "install", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"); err == nil {
		return stepResult{name: "golangci-lint", action: "installed", detail: "via go install"}
	}

	// Fallback to Homebrew.
	if doctor.HasManager(env, doctor.ManagerHomebrew) {
		if out, err := opts.ExecCmd("brew", "install", "golangci-lint"); err != nil {
			return stepResult{name: "golangci-lint", action: "failed", detail: "brew install failed", err: err, output: out}
		}
		return stepResult{name: "golangci-lint", action: "installed", detail: "via Homebrew"}
	}

	return stepResult{
		name:   "golangci-lint",
		action: "failed",
		detail: "Install: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest",
		err:    fmt.Errorf("golangci-lint not available"),
	}
}

// installGovulncheck installs govulncheck if missing per Spec 019
// FR-012. Uses go install (the only installation method).
func installGovulncheck(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("govulncheck"); err == nil {
		return stepResult{name: "govulncheck", action: "already installed"}
	}

	if opts.DryRun {
		return stepResult{name: "govulncheck", action: "dry-run", detail: "Would install: go install golang.org/x/vuln/cmd/govulncheck@latest"}
	}

	if out, err := opts.ExecCmd("go", "install", "golang.org/x/vuln/cmd/govulncheck@latest"); err != nil {
		return stepResult{name: "govulncheck", action: "failed", detail: "go install failed", err: err, output: out}
	}
	return stepResult{name: "govulncheck", action: "installed", detail: "via go install"}
}

// rpmURL constructs the GitHub Release RPM download URL for a tool.
// The URL pattern follows GoReleaser's nfpms naming convention.
func rpmURL(repo, version, arch string) string {
	return fmt.Sprintf(
		"https://github.com/%s/releases/download/v%s/%s_%s_linux_%s.rpm",
		repo,
		version,
		repoName(repo),
		version,
		arch,
	)
}

// repoName extracts the repository name from a "owner/repo" string.
func repoName(repo string) string {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return repo
}

// rpmArch returns the RPM architecture string for the current
// Go architecture.
func rpmArch() string {
	switch runtime.GOARCH {
	case "arm64":
		return "arm64"
	default:
		return "amd64"
	}
}

// installViaRpm installs a tool from a GitHub Release RPM URL
// using dnf. Returns a stepResult with the outcome.
func installViaRpm(opts *Options, toolName, repo, version string) stepResult {
	if version == "" {
		return stepResult{
			name:   toolName,
			action: "skipped",
			detail: "version unknown — cannot construct RPM URL",
		}
	}

	url := rpmURL(repo, version, rpmArch())

	if opts.DryRun {
		return stepResult{
			name:   toolName,
			action: "dry-run",
			detail: "Would install: dnf install -y " + url,
		}
	}

	if out, err := opts.ExecCmd("dnf", "install", "-y", url); err != nil {
		return stepResult{
			name:   toolName,
			action: "failed",
			detail: "dnf install failed — try: dnf install " + url,
			err:    err,
			output: out,
		}
	}
	return stepResult{name: toolName, action: "installed", detail: "via dnf (RPM)"}
}

// ollamaBrew returns the brew command arguments for installing
// Ollama on the given OS. macOS uses the cask (ollama-app) for
// .app bundle with auto-updates. Linux uses the formula (ollama)
// because Homebrew casks are macOS-only.
func ollamaBrew(goos string) []string {
	if goos == "darwin" {
		return []string{"brew", "install", "--cask", "ollama-app"}
	}
	return []string{"brew", "install", "ollama"}
}

// installOllama installs Ollama if missing. Ollama is the local
// model runtime used by both Dewey (semantic search embeddings)
// and Replicator (semantic memory). OS-aware: uses cask on macOS,
// formula on Linux. Skips with download link if no Homebrew.
func installOllama(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("ollama"); err == nil {
		return stepResult{name: "Ollama", action: "already installed"}
	}

	// Determine the Homebrew install method based on OS.
	// macOS: cask (ollama-app) for .app bundle with auto-updates.
	// Linux: formula (ollama) — casks are macOS-only.
	brewArgs := ollamaBrew(opts.GOOS)

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Ollama", action: "dry-run", detail: "Would install: brew install " + strings.Join(brewArgs[1:], " ")}
		}
		return stepResult{name: "Ollama", action: "dry-run", detail: "Would install: download from https://ollama.com/download"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "Ollama",
			action: "skipped",
			detail: "Homebrew not available. Download from https://ollama.com/download",
		}
	}

	if out, err := opts.ExecCmd(brewArgs[0], brewArgs[1:]...); err != nil {
		return stepResult{name: "Ollama", action: "failed", detail: "brew install failed", err: err, output: out}
	}
	return stepResult{name: "Ollama", action: "installed", detail: "via Homebrew"}
}

// podmanMachineTimeout is the timeout for podman machine init,
// which downloads a VM image (~300MB) and can be slow on
// constrained networks. 180 seconds balances patience with
// preventing indefinite hangs.
const podmanMachineTimeout = "180"

// installPodman installs Podman if missing. Podman is the container
// runtime used by the sandbox for isolated agent sessions. On macOS,
// after installation, a Podman machine is initialized and started
// (best-effort — failures are reported but do not block the step).
// A smoke test via `podman info` verifies the installation is
// functional.
func installPodman(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("podman"); err == nil {
		return stepResult{name: "Podman", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Podman", action: "dry-run", detail: "Would install: brew install podman"}
		}
		return stepResult{name: "Podman", action: "dry-run", detail: "Would install: download from https://podman.io/docs/installation"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "Podman",
			action: "skipped",
			detail: "Homebrew not available. Download from https://podman.io/docs/installation",
		}
	}

	if out, err := opts.ExecCmd("brew", "install", "podman"); err != nil {
		return stepResult{name: "Podman", action: "failed", detail: "brew install failed", err: err, output: out}
	}

	// macOS post-install: initialize and start a Podman machine.
	// Podman on macOS requires a VM to run Linux containers.
	detail := "via Homebrew"
	if opts.GOOS == "darwin" {
		detail = podmanMachineInit(opts, detail)
	}

	// Smoke test: verify Podman is functional.
	if _, err := opts.ExecCmd("podman", "info"); err != nil {
		detail += "; podman info failed"
	} else {
		detail += "; verified"
	}

	return stepResult{name: "Podman", action: "installed", detail: detail}
}

// podmanMachineInit checks for an existing Podman machine on macOS
// and initializes one if none exists. Returns the updated detail
// string with machine status appended. Machine failures are
// best-effort — they are reported but do not fail the step (D6).
func podmanMachineInit(opts *Options, detail string) string {
	// Check if a machine already exists.
	output, err := opts.ExecCmd("podman", "machine", "list", "--format", "{{.Name}}")
	if err == nil && strings.TrimSpace(string(output)) != "" {
		// Machine already exists — no init needed.
		return detail
	}

	// No machine exists — initialize one with timeout to prevent
	// indefinite hangs on slow networks (D6). Try gtimeout (macOS
	// coreutils) first, then timeout (Linux), then fall back to
	// running without a timeout wrapper.
	if _, err := initMachineWithTimeout(opts); err != nil {
		return detail + "; machine init failed"
	}

	// Start the machine.
	if _, err := opts.ExecCmd("podman", "machine", "start"); err != nil {
		return detail + "; machine start failed"
	}

	return detail
}

// initMachineWithTimeout runs podman machine init with a timeout
// wrapper. Tries gtimeout (macOS Homebrew coreutils), then timeout
// (Linux/GNU), then falls back to running without a timeout.
func initMachineWithTimeout(opts *Options) ([]byte, error) {
	// macOS: gtimeout from brew install coreutils.
	if _, err := opts.LookPath("gtimeout"); err == nil {
		return opts.ExecCmd("gtimeout", podmanMachineTimeout, "podman", "machine", "init")
	}
	// Linux: GNU timeout is standard.
	if _, err := opts.LookPath("timeout"); err == nil {
		return opts.ExecCmd("timeout", podmanMachineTimeout, "podman", "machine", "init")
	}
	// Fallback: no timeout wrapper available.
	return opts.ExecCmd("podman", "machine", "init")
}

// installDevPod installs DevPod if missing. DevPod provides
// persistent workspace management on top of Podman. Follows the
// installGaze() pattern: Homebrew only, skip with download URL
// if no Homebrew.
func installDevPod(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("devpod"); err == nil {
		return stepResult{name: "DevPod", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "DevPod", action: "dry-run", detail: "Would install: brew install devpod"}
		}
		return stepResult{name: "DevPod", action: "dry-run", detail: "Would install: download from https://devpod.sh/docs/getting-started/install"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "DevPod",
			action: "skipped",
			detail: "Homebrew not available. Download from https://devpod.sh/docs/getting-started/install",
		}
	}

	if out, err := opts.ExecCmd("brew", "install", "devpod"); err != nil {
		return stepResult{name: "DevPod", action: "failed", detail: "brew install failed", err: err, output: out}
	}
	return stepResult{name: "DevPod", action: "installed", detail: "via Homebrew"}
}

// configureDevPodProvider configures the DevPod Podman provider
// alias. The standalone podman provider was removed from DevPod;
// users must alias the Docker provider with DOCKER_PATH=podman
// (D3). This step is gated on both devpod and podman being
// available in PATH. Provider detection uses exact first-column
// name matching on `devpod provider list` output (D5).
func configureDevPodProvider(opts *Options) stepResult {
	// Gate: both devpod and podman must be available.
	if _, err := opts.LookPath("devpod"); err != nil {
		return stepResult{name: "DevPod provider", action: "skipped", detail: "no devpod"}
	}
	if _, err := opts.LookPath("podman"); err != nil {
		return stepResult{name: "DevPod provider", action: "skipped", detail: "no podman"}
	}

	if opts.DryRun {
		return stepResult{
			name:   "DevPod provider",
			action: "dry-run",
			detail: "Would run: devpod provider add docker --name podman -o DOCKER_PATH=podman",
		}
	}

	// Check if provider is already registered.
	output, err := opts.ExecCmd("devpod", "provider", "list")
	if err != nil {
		return stepResult{
			name:   "DevPod provider",
			action: "skipped",
			detail: "devpod provider list failed — check devpod installation",
		}
	}

	// Parse provider list output: exact first-column name matching
	// to avoid false positives from providers like "podman-custom" (D5).
	if hasProvider(string(output), "podman") {
		return stepResult{name: "DevPod provider", action: "already installed"}
	}

	// Provider missing — add the Docker provider aliased to Podman.
	addCmd := "devpod provider add docker --name podman -o DOCKER_PATH=podman"
	if out, err := opts.ExecCmd("devpod", "provider", "add", "docker", "--name", "podman", "-o", "DOCKER_PATH=podman"); err != nil {
		return stepResult{
			name:   "DevPod provider",
			action: "failed",
			detail: "Run manually: " + addCmd,
			err:    err,
			output: out,
		}
	}
	return stepResult{name: "DevPod provider", action: "installed", detail: "podman provider configured"}
}

// hasProvider checks if a provider name appears as an exact match
// in the first column of `devpod provider list` output. Uses exact
// matching to avoid false positives from providers with similar
// names (e.g., "podman-custom" should not match "podman").
func hasProvider(output, name string) bool {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return true
		}
	}
	return false
}

// installDewey installs Dewey and pulls the embedding model.
// Position: after Replicator, before golangci-lint.
// Design decision: Dewey is optional (Constitution Principle II —
// Composability First), so installation failures produce warnings
// rather than hard failures. Note: brew install and ollama pull are
// non-interactive (no stdin prompts), so no interactive guard is
// needed here (unlike swarm setup which may prompt for input).
// Fallback chain (D2): Homebrew → go install → skip (no RPMs).
func installDewey(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("dewey"); err == nil {
		// Dewey already installed — check embedding model.
		return pullEmbeddingModel(opts)
	}

	// Fallback chain (D2): Homebrew → go install → skip.
	// Dewey has no RPMs, so dnf is not in the fallback list.
	method := opts.resolveMethod("dewey", env, "homebrew", "go")
	switch method {
	case "homebrew":
		result := installViaBrew(opts, "Dewey", "unbound-force/tap/dewey")
		if result.action == "installed" {
			return deweyPostInstall(opts, result.detail)
		}
		return result
	case "dnf", "go":
		// Dewey has no RPMs, so "dnf" falls through to go install
		// per D5 design: tools without dnf support gracefully
		// degrade to go install when Go is available.
		result := installViaGo(opts, "Dewey", "github.com/unbound-force/dewey/cmd/dewey")
		if result.action == "installed" {
			return deweyPostInstall(opts, result.detail)
		}
		if result.action != "skipped" {
			return result
		}
		return stepResult{
			name:   "Dewey",
			action: "skipped",
			detail: "Homebrew not available. Install from https://github.com/unbound-force/dewey",
		}
	default:
		return stepResult{
			name:   "Dewey",
			action: "skipped",
			detail: "Homebrew not available. Install from https://github.com/unbound-force/dewey",
		}
	}
}

// deweyPostInstall pulls the embedding model after a successful Dewey
// install. Extracted to avoid duplicating the post-install logic
// across multiple install paths (DRY per CS-004).
func deweyPostInstall(opts *Options, installDetail string) stepResult {
	modelResult := pullEmbeddingModel(opts)
	if modelResult.action == "failed" {
		return stepResult{
			name:   "Dewey",
			action: "installed",
			detail: installDetail + " (model pull failed — run 'ollama serve' then 'ollama pull " + opts.embeddingModel() + "')",
		}
	}
	return stepResult{name: "Dewey", action: "installed", detail: installDetail}
}

// pullEmbeddingModel pulls the enterprise-grade embedding model
// via Ollama. Used by both Dewey and Replicator for consistent
// semantic search across the toolchain.
func pullEmbeddingModel(opts *Options) stepResult {
	if _, err := opts.LookPath("ollama"); err != nil {
		return stepResult{name: "Dewey", action: "skipped", detail: "embedding model requires ollama (install from https://ollama.com/download)"}
	}

	if opts.DryRun {
		return stepResult{name: "Dewey", action: "dry-run", detail: "Would run: ollama pull " + opts.embeddingModel()}
	}

	// Check if model is already pulled.
	output, err := opts.ExecCmd("ollama", "list")
	if err == nil && strings.Contains(string(output), "granite-embedding") {
		return stepResult{name: "Dewey", action: "already installed", detail: "embedding model ready"}
	}

	if out, err := opts.ExecCmd("ollama", "pull", opts.embeddingModel()); err != nil {
		return stepResult{
			name:   "Dewey",
			action: "failed",
			detail: "ollama pull failed — ensure the Ollama server is running (ollama serve), then run: ollama pull " + opts.embeddingModel(),
			err:    err,
			output: out,
		}
	}

	return stepResult{name: "Dewey", action: "installed", detail: "embedding model pulled"}
}

// truncateOutput returns a string representation of command output,
// truncated to the last maxLines/2 lines if the output exceeds
// maxLines lines. Empty output returns an empty string.
func truncateOutput(output []byte, maxLines int) string {
	s := strings.TrimSpace(string(output))
	if s == "" {
		return ""
	}
	lines := strings.Split(s, "\n")
	if len(lines) <= maxLines {
		return s
	}
	tail := maxLines / 2
	omitted := len(lines) - tail
	return fmt.Sprintf("... (%d lines omitted)\n%s", omitted, strings.Join(lines[len(lines)-tail:], "\n"))
}

// printStepResult prints a formatted step result.
func printStepResult(w io.Writer, r stepResult) {
	symbol := "✓"
	switch r.action {
	case "failed":
		symbol = "✗"
	case "skipped":
		symbol = "-"
	case "dry-run":
		symbol = "~"
	}

	line := fmt.Sprintf("  %s %-16s %s", symbol, r.name, r.action)
	if r.detail != "" {
		line += " (" + r.detail + ")"
	}
	fmt.Fprintln(w, line)

	if r.err != nil {
		fmt.Fprintf(w, "                     Error: %v\n", r.err)
	}

	if len(r.output) > 0 && r.action == "failed" {
		if trimmed := truncateOutput(r.output, 20); trimmed != "" {
			for _, line := range strings.Split(trimmed, "\n") {
				fmt.Fprintf(w, "                     %s\n", line)
			}
		}
	}
}

// FormatSetupText renders setup output with symbols per US4/T069.
// This is called by printStepResult during Run() — the setup
// command formats output inline as steps execute.
func FormatSetupText(w io.Writer, results []stepResult) {
	for _, r := range results {
		printStepResult(w, r)
	}
}

// atomicWriteFile writes data to a file atomically using
// write-to-temp-then-rename per FR-027a.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".unbound-setup-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// Clean up temp file on error.
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err = os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename temp to target: %w", err)
	}

	return nil
}
