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

// Run executes the full setup workflow per FR-021/030/032/034/035.
func Run(opts Options) error {
	opts.defaults()

	// Platform guard: Windows is not supported (FR-037).
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Platform not supported: doctor and setup require macOS or Linux")
	}

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

	// Step 3: Ensure Node.js (FR-024).
	nodeResult := ensureNodeJS(&opts, env)
	results = append(results, nodeResult)
	nodeAvailable := nodeResult.err == nil && nodeResult.action != "failed"

	// Steps 4-7: Swarm-related (require Node.js).
	if nodeAvailable {
		// Step 4: Install Swarm plugin (FR-025).
		swarmResult := installSwarmPlugin(&opts, env)
		results = append(results, swarmResult)

		if swarmResult.err == nil && swarmResult.action != "failed" && swarmResult.action != "skipped" {
			// Step 5: Run swarm setup (FR-026).
			results = append(results, runSwarmSetup(&opts))

			// Step 6: Configure opencode.json (FR-027/027a/028).
			results = append(results, configureOpencodeJSON(&opts))

			// Step 7: Initialize .hive/ (FR-029).
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
		results = append(results, stepResult{name: "Swarm plugin", action: "skipped", detail: "no Node.js"})
		results = append(results, stepResult{name: "swarm setup", action: "skipped", detail: "no swarm"})
		results = append(results, stepResult{name: "opencode.json", action: "skipped", detail: "no swarm"})
		results = append(results, stepResult{name: ".hive/", action: "skipped", detail: "no swarm"})
	}

	// Step 8: Install Dewey (after Swarm, before uf init).
	results = append(results, installDewey(&opts, env))

	// Step 9: Run uf init (FR-033).
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

	// Ollama tip: suggest installation for enhanced semantic memory.
	if _, ollamaErr := opts.LookPath("ollama"); ollamaErr != nil {
		fmt.Fprintln(opts.Stdout)
		fmt.Fprintln(opts.Stdout, "Tip: Install Ollama for enhanced semantic memory:")
		fmt.Fprintln(opts.Stdout, "  brew install ollama && ollama pull granite-embedding:30m")
		fmt.Fprintln(opts.Stdout, "  (Without Ollama, Dewey uses full-text search only)")
	}

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
		return stepResult{name: "Dewey", action: "installed", detail: "via Homebrew (embedding model pull failed)"}
	}

	return stepResult{name: "Dewey", action: "installed", detail: "via Homebrew"}
}

// pullEmbeddingModel pulls the granite-embedding:30m model via Ollama.
func pullEmbeddingModel(opts *Options) stepResult {
	if _, err := opts.LookPath("ollama"); err != nil {
		return stepResult{name: "Dewey", action: "already installed", detail: "embedding model requires ollama"}
	}

	if opts.DryRun {
		return stepResult{name: "Dewey", action: "dry-run", detail: "Would run: ollama pull granite-embedding:30m"}
	}

	// Check if model is already pulled.
	output, err := opts.ExecCmd("ollama", "list")
	if err == nil && strings.Contains(string(output), "granite-embedding") {
		return stepResult{name: "Dewey", action: "already installed", detail: "embedding model ready"}
	}

	if _, err := opts.ExecCmd("ollama", "pull", "granite-embedding:30m"); err != nil {
		return stepResult{name: "Dewey", action: "failed", detail: "ollama pull failed", err: err}
	}

	return stepResult{name: "Dewey", action: "installed", detail: "embedding model pulled"}
}

// runUnboundInit runs `uf init` if .opencode/ doesn't exist per FR-033.
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
	result, err := scaffold.Run(scaffold.Options{
		TargetDir: opts.TargetDir,
		Stdout:    opts.Stdout,
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
