// Package setup implements automated tool chain installation for
// the Unbound Force development environment. It detects existing
// version managers, installs missing tools through the appropriate
// manager, configures the Swarm plugin, and scaffolds project files.
// All external dependencies are injected for testability per
// Constitution Principle IV.
package setup

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/unbound-force/unbound-force/internal/doctor"
	"github.com/unbound-force/unbound-force/internal/scaffold"
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
// Dewey and Swarm. IBM Granite, Apache 2.0, permissibly licensed
// training data. Setting these env vars aligns Swarm's Hivemind
// with Dewey's embedding model.
const (
	graniteModel    = "granite-embedding:30m"
	graniteEmbedDim = "256"
)

// Run executes the full setup workflow per FR-021/030/032/034/035.
func Run(opts Options) error {
	opts.defaults()

	// Platform guard: Windows is not supported (FR-037).
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Platform not supported: doctor and setup require macOS or Linux")
	}

	// Set Ollama env vars so Swarm's Hivemind uses the same
	// enterprise-grade embedding model as Dewey. These are
	// inherited by child processes (swarm setup, swarm init).
	os.Setenv("OLLAMA_MODEL", graniteModel)
	os.Setenv("OLLAMA_EMBED_DIM", graniteEmbedDim)

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
	results = append(results, installOpenCode(&opts, env))

	// Step 2: Install Gaze (FR-023).
	results = append(results, installGaze(&opts, env))

	// Step 3: Install Mx F Manager hero.
	results = append(results, installMxF(&opts, env))

	// Step 4: Install GitHub CLI.
	results = append(results, installGH(&opts, env))

	// Step 5: Ensure Node.js (FR-024).
	nodeResult := ensureNodeJS(&opts, env)
	results = append(results, nodeResult)
	nodeAvailable := nodeResult.err == nil && nodeResult.action != "failed"

	// Steps 5-11: Node.js-dependent tools (inside nodeAvailable block).
	if nodeAvailable {
		// Step 6: Ensure bun is available (prerequisite for swarm setup).
		bunResult := ensureBun(&opts, env)
		results = append(results, bunResult)

		// Step 7: Install OpenSpec CLI.
		results = append(results, installOpenSpec(&opts, env))

		// Step 8: Install Swarm plugin (FR-025).
		swarmResult := installSwarmPlugin(&opts, env)
		results = append(results, swarmResult)

		if swarmResult.err == nil && swarmResult.action != "failed" && swarmResult.action != "skipped" {
			// Step 9: Run swarm setup (FR-026).
			results = append(results, runSwarmSetup(&opts))

			// Step 10: Configure opencode.json (FR-027/027a/028).
			results = append(results, configureOpencodeJSON(&opts))

			// Step 11: Initialize .hive/ (FR-029).
			results = append(results, initializeHive(&opts))
		} else if swarmResult.action == "already installed" {
			// Swarm already installed — still configure.
			results = append(results, configureOpencodeJSON(&opts))
			results = append(results, initializeHive(&opts))
		} else {
			results = append(results, stepResult{name: "swarm setup", action: "skipped", detail: "no swarm"})
			results = append(results, stepResult{name: "opencode.json", action: "skipped", detail: "no swarm"})
			results = append(results, stepResult{name: ".hive/", action: "skipped", detail: "no swarm"})
		}
	} else {
		results = append(results, stepResult{name: "OpenSpec CLI", action: "skipped", detail: "no Node.js"})
		results = append(results, stepResult{name: "Swarm plugin", action: "skipped", detail: "no Node.js"})
		results = append(results, stepResult{name: "swarm setup", action: "skipped", detail: "no swarm"})
		results = append(results, stepResult{name: "opencode.json", action: "skipped", detail: "no swarm"})
		results = append(results, stepResult{name: ".hive/", action: "skipped", detail: "no swarm"})
	}

	// Step 12: Install Ollama (prerequisite for Dewey + Swarm embeddings).
	results = append(results, installOllama(&opts, env))

	// Step 13: Install Dewey (after Ollama, before uf init).
	results = append(results, installDewey(&opts, env))

	// Step 14: Initialize .dewey/ workspace.
	deweyInitResult := initDewey(&opts)
	results = append(results, deweyInitResult)

	// Step 15: Build Dewey index (skip if init failed).
	if deweyInitResult.action != "failed" {
		results = append(results, indexDewey(&opts))
	} else {
		results = append(results, stepResult{name: "dewey index", action: "skipped", detail: "dewey init failed"})
	}

	// Step 16: Run uf init (FR-033).
	results = append(results, runUnboundInit(&opts))

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
	fmt.Fprintln(opts.Stdout, "Note: Swarm and Dewey are configured to use "+graniteModel+".")
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
// Follows the installSwarmPlugin() pattern with bun preference:
// try bun first, fall back to npm.
func installOpenSpec(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("openspec"); err == nil {
		return stepResult{name: "OpenSpec CLI", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerBun) {
			return stepResult{name: "OpenSpec CLI", action: "dry-run", detail: "Would install: bun add -g @fission-ai/openspec@latest"}
		}
		return stepResult{name: "OpenSpec CLI", action: "dry-run", detail: "Would install: npm install -g @fission-ai/openspec@latest"}
	}

	// Prefer bun, fall back to npm (enhanced from installSwarmPlugin pattern
	// — falls through to npm on bun failure for resilience).
	if doctor.HasManager(env, doctor.ManagerBun) {
		if _, bunErr := opts.ExecCmd("bun", "add", "-g", "@fission-ai/openspec@latest"); bunErr == nil {
			return stepResult{name: "OpenSpec CLI", action: "installed", detail: "via bun"}
		}
		// bun failed — fall through to npm (log for diagnostics).
		fmt.Fprintln(opts.Stderr, "  bun install failed, trying npm...")
	}

	if _, err := opts.ExecCmd("npm", "install", "-g", "@fission-ai/openspec@latest"); err != nil {
		return stepResult{
			name:   "OpenSpec CLI",
			action: "failed",
			detail: "npm install failed — fix npm permissions (see https://docs.npmjs.com/resolving-eacces-permissions-errors) or install via bun",
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
		err:    fmt.Errorf("Node.js not available"),
	}
}

// installSwarmPlugin installs the Swarm plugin per FR-025.
func installSwarmPlugin(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("swarm"); err == nil {
		return stepResult{name: "Swarm plugin", action: "already installed"}
	}

	if opts.DryRun {
		if doctor.HasManager(env, doctor.ManagerBun) {
			return stepResult{name: "Swarm plugin", action: "dry-run", detail: "Would install: bun add -g opencode-swarm-plugin@latest"}
		}
		return stepResult{name: "Swarm plugin", action: "dry-run", detail: "Would install: npm install -g opencode-swarm-plugin@latest"}
	}

	// Prefer bun if available (FR-025).
	if doctor.HasManager(env, doctor.ManagerBun) {
		if _, err := opts.ExecCmd("bun", "add", "-g", "opencode-swarm-plugin@latest"); err != nil {
			return stepResult{name: "Swarm plugin", action: "failed", detail: "bun install failed", err: err}
		}
		return stepResult{name: "Swarm plugin", action: "installed", detail: "via bun"}
	}

	// Default to npm.
	if _, err := opts.ExecCmd("npm", "install", "-g", "opencode-swarm-plugin@latest"); err != nil {
		return stepResult{name: "Swarm plugin", action: "failed", detail: "npm install failed", err: err}
	}
	return stepResult{name: "Swarm plugin", action: "installed", detail: "via npm"}
}

// ensureBun installs bun if not present. Bun is a prerequisite
// for swarm setup -- the swarm plugin requires bun at runtime.
func ensureBun(opts *Options, env doctor.DetectedEnvironment) stepResult {
	if _, err := opts.LookPath("bun"); err == nil {
		return stepResult{name: "Bun", action: "already installed"}
	}

	if opts.DryRun {
		return stepResult{name: "Bun", action: "dry-run", detail: "Would install: npm install -g bun"}
	}

	if _, err := opts.ExecCmd("npm", "install", "-g", "bun"); err != nil {
		return stepResult{name: "Bun", action: "failed", detail: "npm install -g bun failed", err: err}
	}
	return stepResult{name: "Bun", action: "installed", detail: "via npm"}
}

// runSwarmSetup runs `swarm setup` per FR-026.
// swarm setup may prompt for user input, so it requires either
// the --yes flag or an interactive terminal. Without these,
// CombinedOutput() would hang waiting for stdin that never arrives.
func runSwarmSetup(opts *Options) stepResult {
	if opts.DryRun {
		return stepResult{name: "swarm setup", action: "dry-run", detail: "Would run: swarm setup"}
	}

	if !opts.YesFlag && !opts.IsTTY() {
		return stepResult{
			name:   "swarm setup",
			action: "skipped",
			detail: "interactive — run `swarm setup` manually or use --yes",
		}
	}

	if _, err := opts.ExecCmd("swarm", "setup"); err != nil {
		return stepResult{name: "swarm setup", action: "failed", detail: "swarm setup failed", err: err}
	}
	return stepResult{name: "swarm setup", action: "completed"}
}

// configureOpencodeJSON adds the Swarm plugin to opencode.json
// per FR-027/027a/028.
func configureOpencodeJSON(opts *Options) stepResult {
	if opts.DryRun {
		return stepResult{name: "opencode.json", action: "dry-run", detail: "Would configure plugin entry"}
	}

	ocPath := filepath.Join(opts.TargetDir, "opencode.json")
	data, err := opts.ReadFile(ocPath)

	var ocMap map[string]json.RawMessage

	if err != nil {
		// No opencode.json — create minimal one (FR-028).
		ocMap = map[string]json.RawMessage{
			"$schema": json.RawMessage(`"https://opencode.ai/config.json"`),
		}
	} else {
		// Parse existing file.
		if jsonErr := json.Unmarshal(data, &ocMap); jsonErr != nil {
			// Malformed JSON — refuse to modify (edge case).
			return stepResult{
				name:   "opencode.json",
				action: "skipped",
				detail: "malformed JSON — fix manually",
			}
		}
	}

	// Check if plugin array already has the entry.
	var plugins []string
	if pluginRaw, ok := ocMap["plugin"]; ok {
		if pErr := json.Unmarshal(pluginRaw, &plugins); pErr != nil {
			plugins = []string{}
		}
	}

	// Check if already configured.
	for _, p := range plugins {
		if p == "opencode-swarm-plugin" {
			return stepResult{name: "opencode.json", action: "already configured"}
		}
	}

	// Add the plugin.
	plugins = append(plugins, "opencode-swarm-plugin")
	pluginJSON, _ := json.Marshal(plugins)
	ocMap["plugin"] = json.RawMessage(pluginJSON)

	// Marshal with indentation.
	output, marshalErr := json.MarshalIndent(ocMap, "", "  ")
	if marshalErr != nil {
		return stepResult{name: "opencode.json", action: "failed", detail: "marshal failed", err: marshalErr}
	}
	output = append(output, '\n')

	// Write atomically (FR-027a): temp file + rename.
	if writeErr := opts.WriteFile(ocPath, output, 0644); writeErr != nil {
		return stepResult{name: "opencode.json", action: "failed", detail: "write failed", err: writeErr}
	}

	if err != nil {
		return stepResult{name: "opencode.json", action: "created", detail: "with plugin entry"}
	}
	return stepResult{name: "opencode.json", action: "configured", detail: "plugin added"}
}

// initializeHive runs `swarm init` if .hive/ doesn't exist per FR-029.
// swarm init may prompt for user input, so it requires either
// the --yes flag or an interactive terminal.
func initializeHive(opts *Options) stepResult {
	hivePath := filepath.Join(opts.TargetDir, ".hive")
	if info, err := os.Stat(hivePath); err == nil && info.IsDir() {
		return stepResult{name: ".hive/", action: "already initialized"}
	}

	if opts.DryRun {
		return stepResult{name: ".hive/", action: "dry-run", detail: "Would run: swarm init"}
	}

	if !opts.YesFlag && !opts.IsTTY() {
		return stepResult{
			name:   ".hive/",
			action: "skipped",
			detail: "interactive — run `swarm init` manually or use --yes",
		}
	}

	if _, err := opts.ExecCmd("swarm", "init"); err != nil {
		return stepResult{name: ".hive/", action: "failed", detail: "swarm init failed", err: err}
	}
	return stepResult{name: ".hive/", action: "initialized"}
}

// initDewey runs `dewey init` if `.dewey/` doesn't exist.
// Follows the runSwarmSetup() precedent: takes only opts (no env),
// since no Homebrew/version manager logic is needed.
func initDewey(opts *Options) stepResult {
	deweyDir := filepath.Join(opts.TargetDir, ".dewey")
	if info, err := os.Stat(deweyDir); err == nil && info.IsDir() {
		return stepResult{name: ".dewey/", action: "already initialized"}
	}

	if _, err := opts.LookPath("dewey"); err != nil {
		return stepResult{name: ".dewey/", action: "skipped", detail: "dewey not installed"}
	}

	if opts.DryRun {
		return stepResult{name: ".dewey/", action: "dry-run", detail: "Would run: dewey init"}
	}

	if _, err := opts.ExecCmd("dewey", "init"); err != nil {
		return stepResult{name: ".dewey/", action: "failed", detail: "dewey init failed", err: err}
	}
	return stepResult{name: ".dewey/", action: "initialized"}
}

// indexDewey runs `dewey index` if `.dewey/` exists.
// Follows the runSwarmSetup() precedent: takes only opts (no env).
func indexDewey(opts *Options) stepResult {
	deweyDir := filepath.Join(opts.TargetDir, ".dewey")
	if _, err := os.Stat(deweyDir); os.IsNotExist(err) {
		return stepResult{name: "dewey index", action: "skipped", detail: "no .dewey/ workspace"}
	}

	if _, err := opts.LookPath("dewey"); err != nil {
		return stepResult{name: "dewey index", action: "skipped", detail: "dewey not installed"}
	}

	if opts.DryRun {
		return stepResult{name: "dewey index", action: "dry-run", detail: "Would run: dewey index"}
	}

	if _, err := opts.ExecCmd("dewey", "index"); err != nil {
		return stepResult{
			name:   "dewey index",
			action: "failed",
			detail: "dewey index failed — ensure Ollama server is running (ollama serve)",
			err:    err,
		}
	}
	return stepResult{name: "dewey index", action: "completed"}
}

// installOllama installs Ollama if missing. Ollama is the local
// model runtime used by both Dewey (semantic search embeddings)
// and Swarm (semantic memory). Follows the installGaze() pattern:
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
// Position: after Swarm plugin, before uf init.
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
// via Ollama. Used by both Dewey and Swarm for consistent
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

// runUnboundInit runs `uf init` if .opencode/ doesn't exist per FR-033.
// Note: scaffold.Run() calls initSubTools() which attempts dewey init + dewey index.
// When called from uf setup, steps 14-15 (initDewey/indexDewey) already ran, so
// initSubTools' .dewey/ existence check causes it to skip — no double execution.
// When called standalone (uf init), initSubTools runs dewey init for the first time.
func runUnboundInit(opts *Options) stepResult {
	ocDir := filepath.Join(opts.TargetDir, ".opencode")
	if info, err := os.Stat(ocDir); err == nil && info.IsDir() {
		return stepResult{name: "uf init", action: "already scaffolded"}
	}

	if opts.DryRun {
		return stepResult{name: "uf init", action: "dry-run", detail: "Would run: uf init"}
	}

	// Call scaffold.Run() directly (same binary, no subprocess needed).
	// Design decision: Direct function call avoids subprocess overhead
	// and ensures consistent behavior. Per DRY principle, reuse the
	// existing scaffold engine rather than shelling out.
	// Forward LookPath/ExecCmd to maintain the testability injection chain.
	result, err := scaffold.Run(scaffold.Options{
		TargetDir: opts.TargetDir,
		Stdout:    opts.Stdout,
		LookPath:  opts.LookPath,
		ExecCmd:   opts.ExecCmd,
	})
	if err != nil {
		return stepResult{name: "uf init", action: "failed", detail: "scaffold failed", err: err}
	}

	return stepResult{
		name:   "uf init",
		action: "scaffolded",
		detail: fmt.Sprintf("%d files", len(result.Created)+len(result.Updated)),
	}
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
			os.Remove(tmpPath)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
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
