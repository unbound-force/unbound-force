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
}

// defaultExecCmd is the production implementation of ExecCmd.
func defaultExecCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// stepResult tracks the outcome of a setup step.
type stepResult struct {
	name   string
	action string // "installed", "already installed", "skipped", "failed"
	detail string
	err    error
}

// graniteModel is the enterprise-grade embedding model used by both
// Dewey and Replicator. IBM Granite, Apache 2.0, permissibly licensed
// training data. Setting these env vars aligns all tools
// with the same embedding model.
const (
	graniteModel    = "granite-embedding:30m"
	graniteEmbedDim = "256"
)

// Run executes the full setup workflow per FR-021/030/032/034/035.
func Run(opts Options) error {
	opts.defaults()

	// Platform guard: Windows is not supported (FR-037).
	if runtime.GOOS == "windows" {
		return fmt.Errorf("platform not supported: doctor and setup require macOS or Linux")
	}

	// Set Ollama env vars so all embedding consumers use the same
	// enterprise-grade embedding model. These are inherited by
	// child processes (replicator setup, dewey serve).
	_ = os.Setenv("OLLAMA_MODEL", graniteModel)
	_ = os.Setenv("OLLAMA_EMBED_DIM", graniteEmbedDim)

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

	var results []stepResult

	// Print header.
	fmt.Fprintln(opts.Stdout, "Unbound Force Setup")
	fmt.Fprintln(opts.Stdout, "===================")
	fmt.Fprintln(opts.Stdout)

	// Print detected environment.
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

	// Step 1: Install OpenCode (FR-022).
	fmt.Fprintf(opts.Stdout, "  [1/14] OpenCode...\n")
	results = append(results, installOpenCode(&opts, env))

	// Step 2: Install Gaze (FR-023).
	fmt.Fprintf(opts.Stdout, "  [2/14] Gaze...\n")
	results = append(results, installGaze(&opts, env))

	// Step 3: Install Mx F Manager hero.
	fmt.Fprintf(opts.Stdout, "  [3/14] Mx F...\n")
	results = append(results, installMxF(&opts, env))

	// Step 4: Install GitHub CLI.
	fmt.Fprintf(opts.Stdout, "  [4/14] GitHub CLI...\n")
	results = append(results, installGH(&opts, env))

	// Step 5: Ensure Node.js (FR-024).
	fmt.Fprintf(opts.Stdout, "  [5/14] Node.js...\n")
	nodeResult := ensureNodeJS(&opts, env)
	results = append(results, nodeResult)
	nodeAvailable := nodeResult.err == nil && nodeResult.action != "failed"

	// Step 6: Install OpenSpec CLI (Node.js-dependent).
	if nodeAvailable {
		fmt.Fprintf(opts.Stdout, "  [6/14] OpenSpec CLI...\n")
		results = append(results, installOpenSpec(&opts, env))
	} else {
		results = append(results, stepResult{name: "OpenSpec CLI", action: "skipped", detail: "no Node.js"})
	}

	// Step 7: Install uv (Python package manager for Specify CLI).
	fmt.Fprintf(opts.Stdout, "  [7/14] uv...\n")
	uvResult := installUV(&opts, env)
	results = append(results, uvResult)
	uvAvailable := uvResult.err == nil && uvResult.action != "failed"

	// Step 8: Install Specify CLI (uv-dependent).
	if uvAvailable {
		fmt.Fprintf(opts.Stdout, "  [8/14] Specify CLI...\n")
		results = append(results, installSpecify(&opts, env))
	} else {
		results = append(results, stepResult{name: "Specify CLI", action: "skipped", detail: "no uv"})
	}

	// Step 9: Install Replicator (Homebrew, replaces Swarm plugin).
	fmt.Fprintf(opts.Stdout, "  [9/14] Replicator...\n")
	replicatorResult := installReplicator(&opts, env)
	results = append(results, replicatorResult)

	// Step 10: Run replicator setup.
	if replicatorResult.err == nil && replicatorResult.action != "failed" && replicatorResult.action != "skipped" {
		fmt.Fprintf(opts.Stdout, "  [10/14] Replicator setup...\n")
		results = append(results, runReplicatorSetup(&opts))
	} else {
		results = append(results, stepResult{name: "replicator setup", action: "skipped", detail: "no replicator"})
	}

	// Step 11: Install Ollama (prerequisite for Dewey + Replicator embeddings).
	fmt.Fprintf(opts.Stdout, "  [11/14] Ollama...\n")
	results = append(results, installOllama(&opts, env))

	// Step 12: Install Dewey (after Ollama).
	fmt.Fprintf(opts.Stdout, "  [12/14] Dewey...\n")
	results = append(results, installDewey(&opts, env))

	// Step 13: Install golangci-lint (Spec 019 FR-012).
	fmt.Fprintf(opts.Stdout, "  [13/14] golangci-lint...\n")
	results = append(results, installGolangciLint(&opts, env))

	// Step 14: Install govulncheck (Spec 019 FR-012).
	fmt.Fprintf(opts.Stdout, "  [14/14] govulncheck...\n")
	results = append(results, installGovulncheck(&opts, env))

	// Print results.
	for _, r := range results {
		printStepResult(opts.Stdout, r)
	}

	// Print completion summary (FR-034).
	fmt.Fprintln(opts.Stdout)
	hasFailures := false
	for _, r := range results {
		if r.action == "failed" {
			hasFailures = true
			break
		}
	}

	if hasFailures {
		failCount := 0
		for _, r := range results {
			if r.action == "failed" {
				failCount++
			}
		}
		fmt.Fprintln(opts.Stdout, "Setup partially complete. Fix the errors above, then re-run `uf setup`.")
		return fmt.Errorf("%d step(s) failed", failCount)
	}

	fmt.Fprintln(opts.Stdout, "Setup complete! Run `uf doctor` to verify.")

	// Embedding model alignment note.
	fmt.Fprintln(opts.Stdout)
	fmt.Fprintln(opts.Stdout, "Note: Replicator and Dewey are configured to use "+graniteModel+".")
	fmt.Fprintln(opts.Stdout, "  Add to your shell profile for consistent behavior:")
	fmt.Fprintln(opts.Stdout, "  export OLLAMA_MODEL="+graniteModel)
	fmt.Fprintln(opts.Stdout, "  export OLLAMA_EMBED_DIM="+graniteEmbedDim)

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
		if _, err := opts.ExecCmd("brew", "install", "anomalyco/tap/opencode"); err != nil {
			return stepResult{name: "OpenCode", action: "failed", detail: "brew install failed", err: err}
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

	if _, err := opts.ExecCmd("bash", "-c", "curl -fsSL https://opencode.ai/install | bash"); err != nil {
		return stepResult{name: "OpenCode", action: "failed", detail: "curl install failed", err: err}
	}
	return stepResult{name: "OpenCode", action: "installed", detail: "via curl"}
}

// installMxF installs the Mx F Manager hero if missing.
// Follows the installGaze() pattern: Homebrew only, skip with
// GitHub releases link if no Homebrew.
func installMxF(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("mxf"); err == nil {
		return stepResult{name: "Mx F", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Mx F", action: "dry-run", detail: "Would install: brew install unbound-force/tap/mxf"}
		}
		return stepResult{name: "Mx F", action: "dry-run", detail: "Would install: download from GitHub releases"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "Mx F",
			action: "skipped",
			detail: "Homebrew not available. Download from https://github.com/unbound-force/unbound-force/releases",
		}
	}

	if _, err := opts.ExecCmd("brew", "install", "unbound-force/tap/mxf"); err != nil {
		return stepResult{name: "Mx F", action: "failed", detail: "brew install failed", err: err}
	}
	return stepResult{name: "Mx F", action: "installed", detail: "via Homebrew"}
}

// installGH installs the GitHub CLI if missing.
// Follows the installGaze() pattern: Homebrew only, skip with
// download link if no Homebrew.
func installGH(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("gh"); err == nil {
		return stepResult{name: "GitHub CLI", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "GitHub CLI", action: "dry-run", detail: "Would install: brew install gh"}
		}
		return stepResult{name: "GitHub CLI", action: "dry-run", detail: "Would install: download from https://cli.github.com"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "GitHub CLI",
			action: "skipped",
			detail: "Homebrew not available. Download from https://cli.github.com",
		}
	}

	if _, err := opts.ExecCmd("brew", "install", "gh"); err != nil {
		return stepResult{name: "GitHub CLI", action: "failed", detail: "brew install failed", err: err}
	}
	return stepResult{name: "GitHub CLI", action: "installed", detail: "via Homebrew"}
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

	if _, err := opts.ExecCmd("npm", "install", "-g", "@fission-ai/openspec@latest"); err != nil {
		return stepResult{
			name:   "OpenSpec CLI",
			action: "failed",
			detail: "npm install failed — fix npm permissions (see https://docs.npmjs.com/resolving-eacces-permissions-errors)",
			err:    err,
		}
	}
	return stepResult{name: "OpenSpec CLI", action: "installed", detail: "via npm"}
}

// installGaze installs Gaze if missing per FR-023.
func installGaze(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("gaze"); err == nil {
		return stepResult{name: "Gaze", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Gaze", action: "dry-run", detail: "Would install: brew install unbound-force/tap/gaze"}
		}
		return stepResult{name: "Gaze", action: "dry-run", detail: "Would install: download from GitHub releases"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "Gaze",
			action: "skipped",
			detail: "Homebrew not available. Download from https://github.com/unbound-force/gaze/releases",
		}
	}

	if _, err := opts.ExecCmd("brew", "install", "unbound-force/tap/gaze"); err != nil {
		return stepResult{name: "Gaze", action: "failed", detail: "brew install failed", err: err}
	}
	return stepResult{name: "Gaze", action: "installed", detail: "via Homebrew"}
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
		if _, err := opts.ExecCmd("bash", "-c", cmd); err != nil {
			fmt.Fprintf(opts.Stderr, "nvm install failed: %v\n", err)
			fmt.Fprintf(opts.Stderr, "Manual install: source %s/nvm.sh && nvm install 22\n", nvmDir)
		} else {
			return stepResult{name: "Node.js", action: "installed", detail: "via nvm"}
		}
	}

	// Try fnm.
	if doctor.HasManager(env, doctor.ManagerFnm) {
		if _, err := opts.ExecCmd("fnm", "install", "22"); err != nil {
			return stepResult{name: "Node.js", action: "failed", detail: "fnm install failed", err: err}
		}
		return stepResult{name: "Node.js", action: "installed", detail: "via fnm"}
	}

	// Try Homebrew.
	if doctor.HasManager(env, doctor.ManagerHomebrew) {
		if _, err := opts.ExecCmd("brew", "install", "node"); err != nil {
			return stepResult{name: "Node.js", action: "failed", detail: "brew install failed", err: err}
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
		if _, err := opts.ExecCmd("brew", "install", "uv"); err != nil {
			return stepResult{name: "uv", action: "failed", detail: "brew install failed", err: err}
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

	if _, err := opts.ExecCmd("bash", "-c", "curl -LsSf https://astral.sh/uv/install.sh | sh"); err != nil {
		return stepResult{name: "uv", action: "failed", detail: "curl install failed", err: err}
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

	if _, err := opts.ExecCmd("uv", "tool", "install", "specify-cli"); err != nil {
		return stepResult{
			name:   "Specify CLI",
			action: "failed",
			detail: "uv tool install failed — try: uv tool install specify-cli",
			err:    err,
		}
	}
	return stepResult{name: "Specify CLI", action: "installed", detail: "via uv"}
}

// installReplicator installs Replicator if missing per FR-001.
// Follows the installGaze() pattern: Homebrew only, skip with
// GitHub releases link if no Homebrew.
func installReplicator(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("replicator"); err == nil {
		return stepResult{name: "Replicator", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Replicator", action: "dry-run", detail: "Would install: brew install unbound-force/tap/replicator"}
		}
		return stepResult{name: "Replicator", action: "dry-run", detail: "Would install: download from GitHub releases"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "Replicator",
			action: "skipped",
			detail: "Homebrew not available. Download from https://github.com/unbound-force/replicator/releases",
		}
	}

	if _, err := opts.ExecCmd("brew", "install", "unbound-force/tap/replicator"); err != nil {
		return stepResult{name: "Replicator", action: "failed", detail: "brew install failed", err: err}
	}
	return stepResult{name: "Replicator", action: "installed", detail: "via Homebrew"}
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

	if _, err := opts.ExecCmd("replicator", "setup"); err != nil {
		return stepResult{name: "replicator setup", action: "failed", detail: "replicator setup failed", err: err}
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
		if _, err := opts.ExecCmd("brew", "install", "golangci-lint"); err != nil {
			return stepResult{name: "golangci-lint", action: "failed", detail: "brew install failed", err: err}
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

	if _, err := opts.ExecCmd("go", "install", "golang.org/x/vuln/cmd/govulncheck@latest"); err != nil {
		return stepResult{name: "govulncheck", action: "failed", detail: "go install failed", err: err}
	}
	return stepResult{name: "govulncheck", action: "installed", detail: "via go install"}
}

// installOllama installs Ollama if missing. Ollama is the local
// model runtime used by both Dewey (semantic search embeddings)
// and Replicator (semantic memory). Follows the installGaze() pattern:
// Homebrew only, skip with download link if no Homebrew.
func installOllama(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("ollama"); err == nil {
		return stepResult{name: "Ollama", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Ollama", action: "dry-run", detail: "Would install: brew install --cask ollama-app"}
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

	if _, err := opts.ExecCmd("brew", "install", "--cask", "ollama-app"); err != nil {
		return stepResult{name: "Ollama", action: "failed", detail: "brew install failed", err: err}
	}
	return stepResult{name: "Ollama", action: "installed", detail: "via Homebrew"}
}

// installDewey installs Dewey and pulls the embedding model.
// Position: after Replicator, before golangci-lint.
// Design decision: Dewey is optional (Constitution Principle II —
// Composability First), so installation failures produce warnings
// rather than hard failures. Note: brew install and ollama pull are
// non-interactive (no stdin prompts), so no interactive guard is
// needed here (unlike swarm setup which may prompt for input).
func installDewey(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("dewey"); err == nil {
		// Dewey already installed — check embedding model.
		return pullEmbeddingModel(opts)
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerHomebrew) {
			return stepResult{name: "Dewey", action: "dry-run", detail: "Would install: brew install unbound-force/tap/dewey"}
		}
		return stepResult{name: "Dewey", action: "skipped", detail: "Homebrew not available"}
	}

	if !doctor.HasManager(env, doctor.ManagerHomebrew) {
		return stepResult{
			name:   "Dewey",
			action: "skipped",
			detail: "Homebrew not available. Install from https://github.com/unbound-force/dewey",
		}
	}

	if _, err := opts.ExecCmd("brew", "install", "unbound-force/tap/dewey"); err != nil {
		return stepResult{name: "Dewey", action: "failed", detail: "brew install failed", err: err}
	}

	// After installing, pull the embedding model.
	modelResult := pullEmbeddingModel(opts)
	if modelResult.action == "failed" {
		return stepResult{name: "Dewey", action: "installed", detail: "via Homebrew (model pull failed — run 'ollama serve' then 'ollama pull " + graniteModel + "')"}
	}

	return stepResult{name: "Dewey", action: "installed", detail: "via Homebrew"}
}

// pullEmbeddingModel pulls the enterprise-grade embedding model
// via Ollama. Used by both Dewey and Replicator for consistent
// semantic search across the toolchain.
func pullEmbeddingModel(opts *Options) stepResult {
	if _, err := opts.LookPath("ollama"); err != nil {
		return stepResult{name: "Dewey", action: "skipped", detail: "embedding model requires ollama (install from https://ollama.com/download)"}
	}

	if opts.DryRun {
		return stepResult{name: "Dewey", action: "dry-run", detail: "Would run: ollama pull " + graniteModel}
	}

	// Check if model is already pulled.
	output, err := opts.ExecCmd("ollama", "list")
	if err == nil && strings.Contains(string(output), "granite-embedding") {
		return stepResult{name: "Dewey", action: "already installed", detail: "embedding model ready"}
	}

	if _, err := opts.ExecCmd("ollama", "pull", graniteModel); err != nil {
		return stepResult{
			name:   "Dewey",
			action: "failed",
			detail: "ollama pull failed — ensure the Ollama server is running (ollama serve), then run: ollama pull " + graniteModel,
			err:    err,
		}
	}

	return stepResult{name: "Dewey", action: "installed", detail: "embedding model pulled"}
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
