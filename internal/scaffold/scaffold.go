package scaffold

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/unbound-force/unbound-force/internal/config"
	"github.com/unbound-force/unbound-force/internal/textutil"
)

// markerFileExtensions defines which file types receive version markers.
// Files with extensions not in this set are written without markers.
var markerFileExtensions = map[string]bool{
	".md":   true,
	".yaml": true,
	".yml":  true,
	".sh":   true,
}

//go:embed assets
var assets embed.FS

// Options configures a scaffold run.
type Options struct {
	TargetDir   string                                  // Root dir to scaffold into (default: cwd)
	Force       bool                                    // Overwrite existing files when true
	DivisorOnly bool                                    // Deploy only Divisor agents, command, and packs
	DryRun      bool                                    // When true, configureOpencodeJSON() skips writing
	Lang        string                                  // Language for convention pack selection (auto-detect if empty)
	Version     string                                  // Version string for marker comment (default: "dev")
	Stdout      io.Writer                               // Writer for summary output (default: os.Stdout)
	LookPath    func(string) (string, error)            // Finds a binary in PATH (default: exec.LookPath)
	ExecCmd     func(string, ...string) ([]byte, error) // Runs a command (default: exec.Command wrapper)
	ReadFile    func(string) ([]byte, error)            // Reads a file (default: os.ReadFile)
	WriteFile   func(string, []byte, os.FileMode) error // Writes a file (default: os.WriteFile)
}

// Result tracks the disposition of each scaffolded file.
type Result struct {
	Created     []string // Files written for the first time
	Skipped     []string // Files that existed and were not overwritten
	Overwritten []string // Files that existed and were replaced (Force=true)
	Updated     []string // Tool-owned files overwritten via overwrite-on-diff
}

// defaultExecCmd is the production implementation of ExecCmd.
func defaultExecCmd(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

// Run walks the embedded assets and writes them to the target directory.
// It applies file ownership rules and version markers.
func Run(opts Options) (*Result, error) {
	// Default LookPath and ExecCmd FIRST — before any code path
	// that calls initSubTools() can execute.
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}
	if opts.ExecCmd == nil {
		opts.ExecCmd = defaultExecCmd
	}
	if opts.ReadFile == nil {
		opts.ReadFile = os.ReadFile
	}
	if opts.WriteFile == nil {
		opts.WriteFile = os.WriteFile
	}

	if opts.TargetDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		opts.TargetDir = cwd
	}
	if opts.Version == "" {
		opts.Version = "0.0.0-dev"
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}

	// Resolve language for convention pack selection
	lang := opts.Lang
	langExplicit := lang != ""
	if lang == "" {
		lang = detectLang(opts.TargetDir)
	}
	langDetected := lang != ""
	if lang == "" {
		lang = "default"
	}

	result := &Result{}

	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Strip "assets/" prefix to get the relative path
		relPath := strings.TrimPrefix(path, "assets/")

		// devcontainer/ assets are OS-specific and generated
		// per-user by uf sandbox init. Skip during uf init.
		if strings.HasPrefix(relPath, "devcontainer/") {
			return nil
		}

		// DivisorOnly mode: skip non-Divisor assets
		if opts.DivisorOnly && !isDivisorAsset(relPath) {
			return nil
		}

		// Convention pack language filter (DivisorOnly mode only;
		// full scaffold deploys all packs)
		if opts.DivisorOnly && !shouldDeployPack(relPath, lang) {
			return nil
		}

		// Map asset paths to output paths:
		//   opencode/   -> .opencode/
		//   openspec/   -> openspec/
		outRel := mapAssetPath(relPath)
		outPath := filepath.Join(opts.TargetDir, outRel)

		// Read embedded content
		content, err := assets.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", path, err)
		}

		// Insert format-appropriate version marker for supported file types
		ext := filepath.Ext(relPath)
		var out []byte
		if markerFileExtensions[ext] {
			marker := versionMarker(opts.Version, ext)
			out = insertMarkerAfterFrontmatter(content, marker)
		} else {
			out = content
		}

		// Create parent directories
		dir := filepath.Dir(outPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}

		// Check if file already exists
		existing, readErr := os.ReadFile(outPath)
		fileExists := readErr == nil

		if !fileExists {
			// New file -- create it
			if err := os.WriteFile(outPath, out, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", outPath, err)
			}
			result.Created = append(result.Created, outRel)
			return nil
		}

		// File exists
		if opts.Force {
			// Force mode -- overwrite everything
			if err := os.WriteFile(outPath, out, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", outPath, err)
			}
			result.Overwritten = append(result.Overwritten, outRel)
			return nil
		}

		if isToolOwned(relPath) {
			// Tool-owned -- overwrite if content differs
			if bytes.Equal(existing, out) {
				result.Skipped = append(result.Skipped, outRel)
			} else {
				if err := os.WriteFile(outPath, out, 0o644); err != nil {
					return fmt.Errorf("write %s: %w", outPath, err)
				}
				result.Updated = append(result.Updated, outRel)
			}
			return nil
		}

		// User-owned -- skip
		result.Skipped = append(result.Skipped, outRel)
		return nil
	})

	if err != nil {
		printSummary(opts.Stdout, opts.DivisorOnly, langExplicit, langDetected, result, nil)
		return result, err
	}

	// Create empty directories for user content (skip in DivisorOnly mode)
	if !opts.DivisorOnly {
		emptyDirs := []string{
			filepath.Join(opts.TargetDir, "openspec", "specs"),
			filepath.Join(opts.TargetDir, "openspec", "changes"),
		}
		for _, d := range emptyDirs {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return nil, fmt.Errorf("create directory %s: %w", d, err)
			}
		}
	}

	// Detect legacy reviewer-*.md files in the target directory and
	// warn the user. Per Spec 019 FR-003a: warn but do NOT delete.
	warnLegacyReviewerFiles(opts.Stdout, opts.TargetDir)

	// Ensure .gitignore has the standard UF ignore block.
	// Called after file scaffolding but before sub-tool delegation
	// so that .gitignore is ready before sub-tools create runtime files.
	giResult := ensureGitignore(&opts)

	// Ensure cross-tool bridge files exist so Claude Code and Cursor
	// users discover convention packs out of the box.
	agentsResult := ensureAGENTSmdPackSection(&opts, lang)
	claudeResult := ensureCLAUDEmd(&opts, lang)
	cursorResult := ensureCursorrules(&opts, lang)

	// Initialize sub-tools after file scaffolding, before summary.
	subResults := append([]subToolResult{giResult, agentsResult, claudeResult, cursorResult}, initSubTools(&opts)...)

	// Migrate legacy .opencode/command/ to .opencode/commands/.
	// Runs after initSubTools() so files created by specify init,
	// gaze init, etc. in the old directory are caught and moved.
	if migResult := migrateCommandDir(&opts); migResult != nil {
		subResults = append(subResults, *migResult)
	}

	printSummary(opts.Stdout, opts.DivisorOnly, langExplicit, langDetected, result, subResults)
	return result, nil
}

// warnLegacyReviewerFiles checks for previously scaffolded reviewer-*.md
// files in the target's .opencode/agents/ directory. If found, prints a
// warning listing each file and suggests a removal command. Per Spec 019
// FR-003a: warn but do NOT delete the files.
func warnLegacyReviewerFiles(w io.Writer, targetDir string) {
	pattern := filepath.Join(targetDir, ".opencode", "agents", "reviewer-*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "⚠  Legacy reviewer agents detected:")
	for _, m := range matches {
		_, _ = fmt.Fprintf(w, "    %s\n", filepath.Base(m))
	}
	_, _ = fmt.Fprintln(w, "  These have been superseded by divisor-* agents.")
	_, _ = fmt.Fprintln(w, "  Remove with: rm .opencode/agents/reviewer-*.md")
}

// knownAssetPrefixes enumerates the valid top-level prefixes
// in the embedded assets directory. Used by mapAssetPath to
// detect assets added under unexpected directories.
var knownAssetPrefixes = []string{"opencode/", "openspec/", "devcontainer/"}

// mapAssetPath converts an embedded asset relative path to the
// output path in the target directory. The assets/ directory
// structure mirrors the target with these prefix mappings:
//
//	opencode/ -> .opencode/
//	openspec/ -> openspec/  (no dot prefix)
func mapAssetPath(relPath string) string {
	switch {
	case strings.HasPrefix(relPath, "opencode/"):
		return "." + relPath
	case strings.HasPrefix(relPath, "openspec/"):
		// openspec/ paths pass through without dot prefix
		return relPath
	case strings.HasPrefix(relPath, "devcontainer/"):
		// devcontainer/ assets map to .devcontainer/ in the
		// target directory. Deployed by both uf init and
		// uf sandbox init (D7).
		return "." + relPath
	default:
		// Unknown prefix — pass through unchanged but this
		// indicates a new asset directory was added without
		// updating the mapping. The TestMapAssetPath test
		// should be extended to cover the new prefix.
		return relPath
	}
}

// isToolOwned returns true if the file is maintained by the
// unbound tool and should be overwritten when content differs.
// Tool-owned files: all OpenCode commands, OpenSpec schema
// files, and canonical convention packs (but NOT custom packs).
// Agent files (including Divisor personas) are user-owned and
// fall through to the default return false.
func isToolOwned(relPath string) bool {
	if strings.HasPrefix(relPath, "openspec/schemas/") {
		return true
	}
	if strings.HasPrefix(relPath, "opencode/commands/") {
		return true
	}
	// Skill files are tool-owned (maintained by unbound init).
	if strings.HasPrefix(relPath, "opencode/skills/") {
		return true
	}
	// Convention packs: canonical packs are tool-owned,
	// custom packs (-custom.md) are user-owned
	if isConventionPack(relPath) {
		base := filepath.Base(relPath)
		return !strings.Contains(base, "-custom")
	}
	return false
}

// isDivisorAsset returns true if the asset belongs to the
// Divisor PR Reviewer Council subset. Used to filter assets
// when DivisorOnly mode is active. Convention packs at the
// shared opencode/uf/packs/ location are included via
// isConventionPack() since they are essential for Divisor
// personas to function.
func isDivisorAsset(relPath string) bool {
	if strings.HasPrefix(relPath, "opencode/agents/divisor-") {
		return true
	}
	if relPath == "opencode/commands/review-council.md" {
		return true
	}
	if isConventionPack(relPath) {
		return true
	}
	return false
}

// isConventionPack returns true if the asset is a convention
// pack file under opencode/uf/packs/.
func isConventionPack(relPath string) bool {
	return strings.HasPrefix(relPath, "opencode/uf/packs/")
}

// shouldDeployPack returns true if the convention pack file
// should be deployed for the given resolved language. Always
// deploys default packs. For language-specific packs, only
// deploys the matching language. Non-pack files always return
// true.
func shouldDeployPack(relPath, lang string) bool {
	if !isConventionPack(relPath) {
		return true // Not a pack file — always deploy
	}
	base := filepath.Base(relPath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Always deploy default, severity, and content packs (language-agnostic)
	if name == "default" || name == "default-custom" || name == "severity" ||
		name == "content" || name == "content-custom" {
		return true
	}
	// Deploy language-specific pack and its custom extension
	if name == lang || name == lang+"-custom" {
		return true
	}
	return false
}

// detectLang auto-detects the project language by checking for
// well-known marker files in the target directory. Returns ""
// if no language can be detected.
func detectLang(targetDir string) string {
	markers := []struct {
		file string
		lang string
	}{
		{"go.mod", "go"},
		{"tsconfig.json", "typescript"},
		{"package.json", "typescript"},
		{"pyproject.toml", "python"},
		{"setup.py", "python"},
		{"setup.cfg", "python"},
		{"requirements.txt", "python"},
		{"tox.ini", "python"},
		{"Pipfile", "python"},
		{"Cargo.toml", "rust"},
	}
	for _, m := range markers {
		if _, err := os.Stat(filepath.Join(targetDir, m.file)); err == nil {
			return m.lang
		}
	}
	return ""
}

// versionMarker returns the provenance marker formatted for the
// given file extension. Markdown files use HTML comments; YAML
// and shell scripts use hash comments.
func versionMarker(version string, ext string) string {
	switch ext {
	case ".yaml", ".yml", ".sh":
		return fmt.Sprintf("# scaffolded by uf v%s", version)
	default:
		return fmt.Sprintf("<!-- scaffolded by uf v%s -->", version)
	}
}

// stripExistingMarkers removes all scaffold provenance marker
// lines from content, regardless of version or comment format.
// Marker lines are identified by the prefixes
// "<!-- scaffolded by uf " (HTML comment) and
// "# scaffolded by uf " (hash comment).
func stripExistingMarkers(s string) string {
	var kept []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<!-- scaffolded by uf ") ||
			strings.HasPrefix(trimmed, "# scaffolded by uf ") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// insertMarkerAfterFrontmatter inserts the version marker after
// YAML frontmatter (if present) or appends it at the end.
// Frontmatter is delimited by "---\n" at the start and a
// matching "---\n" line.
//
// The function is idempotent: existing scaffold markers are
// stripped before the new marker is inserted, so the output
// always contains exactly one marker regardless of input state.
func insertMarkerAfterFrontmatter(content []byte, marker string) []byte {
	s := stripExistingMarkers(string(content))

	// Check for YAML frontmatter: must start with "---\n"
	if !strings.HasPrefix(s, "---\n") {
		// No frontmatter -- append marker at the end
		if len(s) > 0 && !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		return []byte(s + marker + "\n")
	}

	// Find closing "---\n" delimiter (after the opening one)
	rest := s[4:] // skip opening "---\n"
	idx := strings.Index(rest, "\n---\n")
	if idx < 0 {
		// Unclosed frontmatter -- append marker at end
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		return []byte(s + marker + "\n")
	}

	// Insert marker after closing "---\n"
	insertPos := 4 + idx + len("\n---\n")
	before := s[:insertPos]
	after := s[insertPos:]

	return []byte(before + marker + "\n" + after)
}

// subToolResult tracks the outcome of a sub-tool initialization step.
// Action values: "initialized", "completed", "failed", "skipped",
// "created", "configured", "already configured", "overwritten",
// "error", "dry-run".
// The err field captures the underlying error from ExecCmd when a
// command fails. The output field captures stdout+stderr from
// CombinedOutput() for diagnostic reporting.
type subToolResult struct {
	name   string
	action string
	detail string
	err    error  // underlying error from ExecCmd
	output []byte // captured stdout+stderr from CombinedOutput()
}

// configureOpencodeJSON creates or updates opencode.json with the Dewey
// MCP server entry (when dewey is in PATH) and the Replicator MCP entry
// (when replicator is in PATH). Respects setup.tools.*.method: skip in
// .uf/config.yaml — tools configured as "skip" are treated as absent.
// Migrates legacy opencode-swarm-plugin entries from the plugin array.
// Idempotent by default; Force overwrites stale mcp.dewey entries.
// Returns a subToolResult describing the outcome.
//
// Design decision: Uses map[string]json.RawMessage to preserve unknown
// user keys (custom MCP servers, custom config). Per SOLID Open/Closed
// Principle — the function adds managed entries without disturbing
// user-owned entries.
func configureOpencodeJSON(opts *Options) []subToolResult {
	// Default file I/O for direct callers (tests) that bypass Run()/initSubTools().
	if opts.ReadFile == nil {
		opts.ReadFile = os.ReadFile
	}
	if opts.WriteFile == nil {
		opts.WriteFile = os.WriteFile
	}

	if opts.DryRun {
		return []subToolResult{{
			name:   "opencode.json",
			action: "dry-run",
		}}
	}

	// Load config to respect setup.tools.*.method: skip.
	// Error ignored: Load returns compiled defaults on missing or
	// malformed config files — graceful degradation by design.
	cfg, _ := config.Load(config.LoadOptions{
		ProjectDir: opts.TargetDir,
		ReadFile:   opts.ReadFile,
	})

	// Detect what needs to be configured.
	hasDewey := false
	if _, err := opts.LookPath("dewey"); err == nil {
		hasDewey = true
	}
	if tc, ok := cfg.Setup.Tools["dewey"]; ok && tc.Method == "skip" {
		hasDewey = false
	}
	hasReplicator := false
	if _, err := opts.LookPath("replicator"); err == nil {
		hasReplicator = true
	}
	if tc, ok := cfg.Setup.Tools["replicator"]; ok && tc.Method == "skip" {
		hasReplicator = false
	}

	// Nothing to configure — skip.
	if !hasDewey && !hasReplicator {
		return []subToolResult{{
			name:   "opencode.json",
			action: "skipped",
			detail: "nothing to configure",
		}}
	}

	ocPath := filepath.Join(opts.TargetDir, "opencode.json")
	data, readErr := opts.ReadFile(ocPath)

	var ocMap map[string]json.RawMessage
	fileExisted := false

	if readErr != nil {
		if !os.IsNotExist(readErr) {
			// Non-"not found" read error (e.g., permission denied).
			return []subToolResult{{
				name:   "opencode.json",
				action: "error",
				detail: fmt.Sprintf("read failed: %v", readErr),
			}}
		}
		// File does not exist — create a new map.
		ocMap = map[string]json.RawMessage{
			"$schema": json.RawMessage(`"https://opencode.ai/config.json"`),
		}
	} else {
		fileExisted = true
		if jsonErr := json.Unmarshal(data, &ocMap); jsonErr != nil {
			return []subToolResult{{
				name:   "opencode.json",
				action: "error",
				detail: "malformed JSON",
			}}
		}
	}

	// Track whether we made any changes.
	changed := false
	forceOverwritten := false

	// --- MCP dewey entry ---
	if hasDewey {
		deweyEntry := json.RawMessage(`{
    "type": "local",
    "command": ["dewey", "serve", "--vault", "."],
    "enabled": true
  }`)

		// Check for existing mcp.dewey or legacy mcpServers.dewey.
		alreadyHasDewey := false

		// Check canonical "mcp" key.
		if mcpRaw, ok := ocMap["mcp"]; ok {
			var mcpMap map[string]json.RawMessage
			if json.Unmarshal(mcpRaw, &mcpMap) == nil {
				if _, hasDeweyKey := mcpMap["dewey"]; hasDeweyKey {
					alreadyHasDewey = true
				}
			}
		}

		// Check legacy "mcpServers" key.
		if !alreadyHasDewey {
			if mcpServersRaw, ok := ocMap["mcpServers"]; ok {
				var mcpServersMap map[string]json.RawMessage
				if json.Unmarshal(mcpServersRaw, &mcpServersMap) == nil {
					if _, hasDeweyKey := mcpServersMap["dewey"]; hasDeweyKey {
						alreadyHasDewey = true
					}
				}
			}
		}

		if !alreadyHasDewey || opts.Force {
			// Get or create the mcp map.
			var mcpMap map[string]json.RawMessage
			if mcpRaw, ok := ocMap["mcp"]; ok {
				if json.Unmarshal(mcpRaw, &mcpMap) != nil {
					mcpMap = make(map[string]json.RawMessage)
				}
			} else {
				mcpMap = make(map[string]json.RawMessage)
			}

			if opts.Force && alreadyHasDewey {
				forceOverwritten = true
			}

			mcpMap["dewey"] = deweyEntry
			mcpJSON, _ := json.Marshal(mcpMap)
			ocMap["mcp"] = json.RawMessage(mcpJSON)
			changed = true
		}
	}

	// --- Replicator MCP entry ---
	if hasReplicator {
		replicatorEntry := json.RawMessage(`{
    "type": "local",
    "command": ["replicator", "serve"],
    "enabled": true
  }`)

		// Check for existing mcp.replicator.
		alreadyHasReplicator := false
		if mcpRaw, ok := ocMap["mcp"]; ok {
			var mcpMap map[string]json.RawMessage
			if json.Unmarshal(mcpRaw, &mcpMap) == nil {
				if _, hasKey := mcpMap["replicator"]; hasKey {
					alreadyHasReplicator = true
				}
			}
		}

		if !alreadyHasReplicator || opts.Force {
			// Get or create the mcp map.
			var mcpMap map[string]json.RawMessage
			if mcpRaw, ok := ocMap["mcp"]; ok {
				if json.Unmarshal(mcpRaw, &mcpMap) != nil {
					mcpMap = make(map[string]json.RawMessage)
				}
			} else {
				mcpMap = make(map[string]json.RawMessage)
			}

			mcpMap["replicator"] = replicatorEntry
			mcpJSON, _ := json.Marshal(mcpMap)
			ocMap["mcp"] = json.RawMessage(mcpJSON)
			changed = true
		}
	}

	// --- Legacy plugin migration ---
	// Remove opencode-swarm-plugin from plugin array if present.
	if pluginRaw, ok := ocMap["plugin"]; ok {
		var plugins []string
		if json.Unmarshal(pluginRaw, &plugins) == nil {
			var filtered []string
			removed := false
			for _, p := range plugins {
				if p == "opencode-swarm-plugin" {
					removed = true
					continue
				}
				filtered = append(filtered, p)
			}
			if removed {
				if len(filtered) == 0 {
					// Empty plugin array — remove the key entirely.
					delete(ocMap, "plugin")
				} else {
					pluginJSON, _ := json.Marshal(filtered)
					ocMap["plugin"] = json.RawMessage(pluginJSON)
				}
				changed = true
			}
		}
	}

	// Nothing changed — already configured.
	if !changed {
		return []subToolResult{{
			name:   "opencode.json",
			action: "already configured",
		}}
	}

	// Marshal with 2-space indent + trailing newline (FR-016: deterministic output).
	output, marshalErr := json.MarshalIndent(ocMap, "", "  ")
	if marshalErr != nil {
		return []subToolResult{{
			name:   "opencode.json",
			action: "error",
			detail: fmt.Sprintf("marshal failed: %v", marshalErr),
		}}
	}
	output = append(output, '\n')

	// Write the file.
	if writeErr := opts.WriteFile(ocPath, output, 0o644); writeErr != nil {
		return []subToolResult{{
			name:   "opencode.json",
			action: "failed",
			detail: fmt.Sprintf("write failed: %v", writeErr),
		}}
	}

	// Determine the action based on what happened.
	action := "created"
	if fileExisted {
		if forceOverwritten {
			action = "overwritten"
		} else {
			action = "configured"
		}
	}

	return []subToolResult{{
		name:   "opencode.json",
		action: action,
	}}
}

// gitignoreBlock is the standard Unbound Force ignore block appended
// to .gitignore by ensureGitignore(). The marker comment on the first
// line is used for idempotency detection — if it already exists in
// the file, the block is not appended again.
const gitignoreBlock = `# Unbound Force — managed by uf init
# Runtime data under .uf/ (databases, caches, locks, logs)
.uf/workflows/
.uf/artifacts/
.uf/dewey/graph.db
.uf/dewey/graph.db-shm
.uf/dewey/graph.db-wal
.uf/dewey/*.lock
.uf/dewey/cache/
.uf/dewey/dewey.log
.uf/replicator/*.db
.uf/replicator/*.db-shm
.uf/replicator/*.db-wal
.uf/replicator/*.lock
.uf/muti-mind/artifacts/
.uf/mx-f/data/
# Devcontainer — OS-specific, generated by uf sandbox init
.devcontainer/
# Legacy tool directories (renamed to .uf/ in Spec 025)
.dewey/
.hive/
.unbound-force/
.muti-mind/
.mx-f/
`

// gitignoreMarker is the sentinel string used to detect whether the
// UF ignore block has already been appended. Extracted as a constant
// so the marker is defined in exactly one place (DRY).
const gitignoreMarker = "# Unbound Force — managed by uf init"

// ensureGitignore appends the standard UF ignore block to .gitignore
// in targetDir. Idempotent: if the marker comment is already present,
// the file is not modified. Creates .gitignore if it does not exist.
// Uses opts.ReadFile/WriteFile for testability (dependency injection).
func ensureGitignore(opts *Options) subToolResult {
	giPath := filepath.Join(opts.TargetDir, ".gitignore")

	existing, readErr := opts.ReadFile(giPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		// Non-"not found" error (e.g., permission denied).
		return subToolResult{
			name:   ".gitignore",
			action: "failed",
			detail: fmt.Sprintf("read failed: %v", readErr),
		}
	}

	// Idempotency check: if the marker already exists, skip.
	if readErr == nil && strings.Contains(string(existing), gitignoreMarker) {
		return subToolResult{
			name:   ".gitignore",
			action: "already configured",
		}
	}

	// Build the new content: existing content + blank line separator + UF block.
	var content string
	if readErr == nil {
		content = string(existing)
		// Ensure a blank line separates existing content from the UF block.
		if len(content) > 0 && !strings.HasSuffix(content, "\n\n") {
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n"
		}
	}
	content += gitignoreBlock

	if writeErr := opts.WriteFile(giPath, []byte(content), 0o644); writeErr != nil {
		return subToolResult{
			name:   ".gitignore",
			action: "failed",
			detail: fmt.Sprintf("write failed: %v", writeErr),
		}
	}

	return subToolResult{
		name:   ".gitignore",
		action: "configured",
	}
}

// migrateCommandDir moves files from the legacy .opencode/command/
// directory to the canonical .opencode/commands/ directory. Handles
// three cases: atomic rename (only old exists), per-file merge (both
// exist), and no-op (old does not exist). Idempotent and re-runnable.
// Returns nil when there is nothing to migrate (silent no-op).
// Skipped in DivisorOnly mode (subset deployment should not rename
// directories in a foreign repo).
func migrateCommandDir(opts *Options) *subToolResult {
	if opts.DivisorOnly {
		return nil
	}

	oldDir := filepath.Join(opts.TargetDir, ".opencode", "command")
	newDir := filepath.Join(opts.TargetDir, ".opencode", "commands")

	// Check if old dir exists. Use Lstat to detect symlinks.
	oldInfo, err := os.Lstat(oldDir)
	if err != nil {
		// Old dir does not exist — nothing to migrate.
		return nil
	}

	// Symlink guard: do not migrate symlinked directories.
	if oldInfo.Mode()&os.ModeSymlink != 0 {
		return &subToolResult{
			name:   ".opencode/command/",
			action: "skipped",
			detail: "symlink detected; manual migration required",
		}
	}

	// Case: only old dir exists — atomic rename.
	if _, statErr := os.Stat(newDir); os.IsNotExist(statErr) {
		if renameErr := os.Rename(oldDir, newDir); renameErr != nil {
			return &subToolResult{
				name:   ".opencode/command/ → commands/",
				action: "failed",
				detail: fmt.Sprintf("rename: %v", renameErr),
			}
		}
		count := countMDFiles(newDir)
		return &subToolResult{
			name:   ".opencode/command/ → commands/",
			action: "migrated",
			detail: fmt.Sprintf("%d files renamed", count),
		}
	}

	// Case: both dirs exist — per-file merge.
	var moved, skipped, warned int
	entries, readErr := os.ReadDir(oldDir)
	if readErr != nil {
		return &subToolResult{
			name:   ".opencode/command/ → commands/",
			action: "failed",
			detail: fmt.Sprintf("read old dir: %v", readErr),
		}
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		oldPath := filepath.Join(oldDir, name)
		newPath := filepath.Join(newDir, name)

		if _, statErr := os.Stat(newPath); statErr == nil {
			// File exists in both dirs.
			oldContent, rErr := opts.ReadFile(oldPath)
			if rErr != nil {
				_, _ = fmt.Fprintf(opts.Stdout,
					"  ⚠ %s: read failed: %v\n", name, rErr)
				warned++
				continue
			}
			newContent, rErr := opts.ReadFile(newPath)
			if rErr != nil {
				_, _ = fmt.Fprintf(opts.Stdout,
					"  ⚠ %s: read failed: %v\n", name, rErr)
				warned++
				continue
			}
			if !bytes.Equal(oldContent, newContent) {
				_, _ = fmt.Fprintf(opts.Stdout,
					"  ⚠ %s: conflict — kept commands/ version"+
						" (run /uf-init for AI-assisted resolution)\n", name)
				warned++
			}
			// Remove old copy (keep commands/ version).
			_ = os.Remove(oldPath)
			skipped++
		} else {
			// File only in old dir — move it.
			if moveErr := moveFile(oldPath, newPath, opts); moveErr != nil {
				_, _ = fmt.Fprintf(opts.Stdout,
					"  ⚠ %s: move failed: %v\n", name, moveErr)
				warned++
				continue
			}
			moved++
		}
	}

	// Try to remove old dir if empty.
	_ = os.Remove(oldDir)

	return &subToolResult{
		name:   ".opencode/command/ → commands/",
		action: "migrated",
		detail: fmt.Sprintf("moved %d, skipped %d duplicates, %d warnings",
			moved, skipped, warned),
	}
}

// moveFile moves a file from src to dst. Uses os.Rename first
// (fast, atomic on same filesystem). On failure, falls back to
// read → write → remove.
func moveFile(src, dst string, opts *Options) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	// Fallback: read → write → remove.
	content, err := opts.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	if err := opts.WriteFile(dst, content, 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return os.Remove(src)
}

// countMDFiles counts .md files in the given directory.
func countMDFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			count++
		}
	}
	return count
}

// agentsmdPackMarker is the heading used to detect whether the
// Convention Packs section has already been appended to AGENTS.md.
const agentsmdPackMarker = "## Convention Packs"

// ensureAGENTSmdPackSection appends a "Convention Packs" section to
// AGENTS.md listing the deployed convention packs. Idempotent: if the
// heading already exists, the file is not modified. Skips if AGENTS.md
// does not exist (nothing to append to).
// Uses opts.ReadFile/WriteFile for testability (dependency injection).
func ensureAGENTSmdPackSection(opts *Options, lang string) subToolResult {
	agentsPath := filepath.Join(opts.TargetDir, "AGENTS.md")

	existing, readErr := opts.ReadFile(agentsPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return subToolResult{
				name:   "AGENTS.md pack section",
				action: "skipped (no AGENTS.md)",
			}
		}
		return subToolResult{
			name:   "AGENTS.md pack section",
			action: "failed",
			detail: fmt.Sprintf("read failed: %v", readErr),
		}
	}

	if strings.Contains(string(existing), agentsmdPackMarker) {
		return subToolResult{
			name:   "AGENTS.md pack section",
			action: "already configured",
		}
	}

	packs := collectDeployedPacks(lang, opts.TargetDir)
	var section strings.Builder
	section.WriteString("\n" + agentsmdPackMarker + "\n\n")
	section.WriteString("This repository uses convention packs scaffolded by\n")
	section.WriteString("unbound-force. Agents MUST read the applicable pack(s)\n")
	section.WriteString("before writing or reviewing code.\n\n")
	for _, p := range packs {
		section.WriteString("- `.opencode/uf/packs/" + p + "`\n")
	}

	content := string(existing)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += section.String()

	if writeErr := opts.WriteFile(agentsPath, []byte(content), 0o644); writeErr != nil {
		return subToolResult{
			name:   "AGENTS.md pack section",
			action: "failed",
			detail: fmt.Sprintf("write failed: %v", writeErr),
		}
	}

	return subToolResult{
		name:   "AGENTS.md pack section",
		action: "configured",
	}
}

// replaceManagedBlock replaces the managed block (from marker to EOF)
// in content with newBlock. Returns the updated content and whether
// a change was made. If the managed block is already identical to
// newBlock, the original content is returned unchanged.
func replaceManagedBlock(content, marker, newBlock string) (string, bool) {
	idx := strings.Index(content, marker)
	if idx < 0 {
		return content, false
	}
	prefix := content[:idx]
	if prefix+newBlock == content {
		return content, false
	}
	return prefix + newBlock, true
}

// claudemdMarker is the sentinel string used to detect the managed
// block in CLAUDE.md. Same marker pattern as gitignoreMarker.
const claudemdMarker = "# Unbound Force — managed by uf init"

// buildCLAUDEmdBlock generates the managed block for CLAUDE.md.
// root is the project root directory; when non-empty, empty custom
// packs are excluded from the generated @-import lines (see
// collectDeployedPacks).
func buildCLAUDEmdBlock(lang, root string) string {
	packs := collectDeployedPacks(lang, root)
	var block strings.Builder
	block.WriteString(claudemdMarker + "\n\n")
	block.WriteString("@AGENTS.md\n")
	block.WriteString("@.opencode/agents/cobalt-crush-dev.md\n\n")
	block.WriteString("## Convention Packs\n\n")
	for _, p := range packs {
		block.WriteString("@.opencode/uf/packs/" + p + "\n")
	}
	block.WriteString("\n## Review Agents (read on-demand)\n\n")
	block.WriteString("When performing code review, read the applicable\n")
	block.WriteString("Divisor agent from .opencode/agents/:\n")
	block.WriteString("- divisor-guard.md — intent drift, constitution\n")
	block.WriteString("- divisor-architect.md — structure, patterns, DRY\n")
	block.WriteString("- divisor-adversary.md — security, error handling\n")
	block.WriteString("- divisor-testing.md — test quality, assertions\n")
	block.WriteString("- divisor-sre.md — operations, performance\n")
	return block.String()
}

// ensureCLAUDEmd creates or appends a managed block to CLAUDE.md with
// @imports for AGENTS.md and deployed convention packs. Idempotent:
// if the managed block already matches the expected content, the file
// is not modified. Stale managed blocks are replaced in-place.
// Uses opts.ReadFile/WriteFile for testability (dependency injection).
func ensureCLAUDEmd(opts *Options, lang string) subToolResult {
	claudePath := filepath.Join(opts.TargetDir, "CLAUDE.md")

	existing, readErr := opts.ReadFile(claudePath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return subToolResult{
			name:   "CLAUDE.md",
			action: "failed",
			detail: fmt.Sprintf("read failed: %v", readErr),
		}
	}

	newBlock := buildCLAUDEmdBlock(lang, opts.TargetDir)

	// Marker present -- check whether managed block needs updating.
	if readErr == nil && strings.Contains(string(existing), claudemdMarker) {
		updated, changed := replaceManagedBlock(
			string(existing), claudemdMarker, newBlock,
		)
		if !changed {
			return subToolResult{
				name:   "CLAUDE.md",
				action: "already configured",
			}
		}
		if writeErr := opts.WriteFile(claudePath, []byte(updated), 0o644); writeErr != nil {
			return subToolResult{
				name:   "CLAUDE.md",
				action: "failed",
				detail: fmt.Sprintf("write failed: %v", writeErr),
			}
		}
		return subToolResult{
			name:   "CLAUDE.md",
			action: "updated",
		}
	}

	// No marker -- create or append.
	var content string
	if readErr == nil {
		content = string(existing)
		if len(content) > 0 && !strings.HasSuffix(content, "\n\n") {
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n"
		}
	}
	content += newBlock

	if writeErr := opts.WriteFile(claudePath, []byte(content), 0o644); writeErr != nil {
		return subToolResult{
			name:   "CLAUDE.md",
			action: "failed",
			detail: fmt.Sprintf("write failed: %v", writeErr),
		}
	}

	action := "configured"
	if readErr == nil {
		action = "appended"
	}
	return subToolResult{
		name:   "CLAUDE.md",
		action: action,
	}
}

// cursorrulesMarker is the sentinel string used to detect the managed
// block in .cursorrules. Same marker pattern as claudemdMarker.
const cursorrulesMarker = claudemdMarker

// buildCursorrulesBlock generates the managed block for .cursorrules.
// root is the project root directory; when non-empty, empty custom
// packs are excluded (see collectDeployedPacks).
func buildCursorrulesBlock(lang, root string) string {
	packs := collectDeployedPacks(lang, root)
	var block strings.Builder
	block.WriteString(cursorrulesMarker + "\n\n")
	block.WriteString("This project follows coding conventions defined in\n")
	block.WriteString("AGENTS.md and enforced through convention packs. Before\n")
	block.WriteString("writing or reviewing code, read the applicable convention\n")
	block.WriteString("pack(s) from .opencode/uf/packs/ and apply all rules\n")
	block.WriteString("marked [MUST].\n\n")
	block.WriteString("Available packs:\n")
	for _, p := range packs {
		block.WriteString("- .opencode/uf/packs/" + p + "\n")
	}
	block.WriteString("\nFor engineering philosophy and coding principles, read\n")
	block.WriteString(".opencode/agents/cobalt-crush-dev.md.\n\n")
	block.WriteString("When reviewing code, consult the applicable reviewer\n")
	block.WriteString("checklist from .opencode/agents/:\n")
	block.WriteString("- divisor-guard.md — intent drift, constitution\n")
	block.WriteString("- divisor-architect.md — structure, patterns, DRY\n")
	block.WriteString("- divisor-adversary.md — security, error handling\n")
	block.WriteString("- divisor-testing.md — test quality, assertions\n")
	block.WriteString("- divisor-sre.md — operations, performance\n")
	return block.String()
}

// ensureCursorrules creates or appends a managed block to .cursorrules
// with instructions to read AGENTS.md and convention packs. Idempotent:
// if the managed block already matches the expected content, the file
// is not modified. Stale managed blocks are replaced in-place.
// Uses opts.ReadFile/WriteFile for testability (dependency injection).
func ensureCursorrules(opts *Options, lang string) subToolResult {
	rulesPath := filepath.Join(opts.TargetDir, ".cursorrules")

	existing, readErr := opts.ReadFile(rulesPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return subToolResult{
			name:   ".cursorrules",
			action: "failed",
			detail: fmt.Sprintf("read failed: %v", readErr),
		}
	}

	newBlock := buildCursorrulesBlock(lang, opts.TargetDir)

	// Marker present -- check whether managed block needs updating.
	if readErr == nil && strings.Contains(string(existing), cursorrulesMarker) {
		updated, changed := replaceManagedBlock(
			string(existing), cursorrulesMarker, newBlock,
		)
		if !changed {
			return subToolResult{
				name:   ".cursorrules",
				action: "already configured",
			}
		}
		if writeErr := opts.WriteFile(rulesPath, []byte(updated), 0o644); writeErr != nil {
			return subToolResult{
				name:   ".cursorrules",
				action: "failed",
				detail: fmt.Sprintf("write failed: %v", writeErr),
			}
		}
		return subToolResult{
			name:   ".cursorrules",
			action: "updated",
		}
	}

	// No marker -- create or append.
	var content string
	if readErr == nil {
		content = string(existing)
		if len(content) > 0 && !strings.HasSuffix(content, "\n\n") {
			if !strings.HasSuffix(content, "\n") {
				content += "\n"
			}
			content += "\n"
		}
	}
	content += newBlock

	if writeErr := opts.WriteFile(rulesPath, []byte(content), 0o644); writeErr != nil {
		return subToolResult{
			name:   ".cursorrules",
			action: "failed",
			detail: fmt.Sprintf("write failed: %v", writeErr),
		}
	}

	action := "configured"
	if readErr == nil {
		action = "appended"
	}
	return subToolResult{
		name:   ".cursorrules",
		action: action,
	}
}

// hasRuleContent reports whether the convention pack file at path
// contains actual rule content — defined as at least one non-whitespace
// character after the last occurrence of the placeholder sentinel
// comment. If the file cannot be read (missing, permission error,
// etc.), hasRuleContent returns true so that the caller fails open and
// does not silently drop an import the user may have populated.
func hasRuleContent(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return true // fail-open: unknown → include
	}
	const sentinel = "<!-- Add project-specific rules below this line -->"
	content := string(data)
	idx := strings.LastIndex(content, sentinel)
	if idx < 0 {
		return true // no sentinel → assume non-empty scaffold
	}
	after := content[idx+len(sentinel):]
	return strings.TrimSpace(after) != ""
}

// collectDeployedPacks returns the list of convention pack filenames
// that would be deployed for the given resolved language. The list
// always includes default.md, default-custom.md, severity.md,
// content.md, and content-custom.md. Language-specific packs are
// added when lang is not "default".
//
// When root is non-empty, each *-custom.md entry is checked with
// hasRuleContent against the file on disk at
// <root>/.opencode/uf/packs/<name>. Empty stubs are omitted from the
// returned list. Non-custom packs are always included regardless of
// root. When root is empty the function behaves as before (all packs
// returned, no filesystem access).
func collectDeployedPacks(lang, root string) []string {
	candidates := []string{
		"default.md",
		"default-custom.md",
		"severity.md",
		"content.md",
		"content-custom.md",
	}
	if lang != "" && lang != "default" {
		candidates = append(candidates, lang+".md", lang+"-custom.md")
	}
	return filterEmptyCustomPacks(candidates, root)
}

// filterEmptyCustomPacks removes *-custom.md entries whose on-disk
// files are empty stubs (no rule content after the sentinel comment).
// When root is empty, all candidates are returned unchanged (no
// filesystem access).
func filterEmptyCustomPacks(candidates []string, root string) []string {
	if root == "" {
		return candidates
	}
	packs := make([]string, 0, len(candidates))
	for _, name := range candidates {
		if strings.HasSuffix(name, "-custom.md") {
			path := filepath.Join(root, ".opencode", "uf", "packs", name)
			if !hasRuleContent(path) {
				continue // skip empty stub
			}
		}
		packs = append(packs, name)
	}
	return packs
}

// initSubTools initializes sub-tools after file scaffolding.
// Errors are captured and reported as warnings in printSummary,
// not hard failures (per Constitution Principle II — Composability First).
// Skips in DivisorOnly mode (deploying reviewer assets to an
// external repo should not initialize Dewey).
func initSubTools(opts *Options) []subToolResult {
	if opts.DivisorOnly {
		return nil
	}

	// Default Stdout and file I/O for direct callers (tests) that bypass Run().
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.ReadFile == nil {
		opts.ReadFile = os.ReadFile
	}
	if opts.WriteFile == nil {
		opts.WriteFile = os.WriteFile
	}

	// NOTE: .uf/config.yaml is no longer created by uf init.
	// Users create it via `uf config init` when they need
	// customization. See internal/config/ package.

	// Load config to respect setup.tools.*.method: skip.
	// Must happen here (not just in configureOpencodeJSON) so that
	// skipped tools are never init'd or indexed.
	// Error ignored: Load returns compiled defaults on missing or
	// malformed config files — graceful degradation by design.
	cfg, _ := config.Load(config.LoadOptions{
		ProjectDir: opts.TargetDir,
		ReadFile:   opts.ReadFile,
	})

	// shouldSkipTool returns true when the user has configured
	// setup.tools.<name>.method: skip in .uf/config.yaml (or
	// the user-level config).
	shouldSkipTool := func(name string) bool {
		if tc, ok := cfg.Setup.Tools[name]; ok && tc.Method == "skip" {
			return true
		}
		return false
	}

	// Concurrent sub-tool initialization (D1, D2 from design.md).
	//
	// Group A (Dewey): dewey init -> generateDeweySources ->
	//   dewey index — sequential within its goroutine.
	// Group B (others): replicator, specify, openspec, gaze —
	//   each in its own goroutine, all concurrent with Group A.
	//
	// sync.WaitGroup coordinates completion; sync.Mutex guards
	// the shared results slice. Individual tool failures are
	// non-fatal (FR-005) — all tools attempt init regardless.

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results []subToolResult
	)

	// collect safely appends results from any goroutine.
	collect := func(rs ...subToolResult) {
		mu.Lock()
		results = append(results, rs...)
		mu.Unlock()
	}

	// logf safely writes to opts.Stdout from any goroutine.
	// opts.Stdout may be a bytes.Buffer in tests, which is
	// not goroutine-safe.
	logf := func(format string, a ...interface{}) {
		mu.Lock()
		_, _ = fmt.Fprintf(opts.Stdout, format, a...)
		mu.Unlock()
	}

	// --- Group A: Dewey (sequential within goroutine) ---

	if !shouldSkipTool("dewey") {
		if _, err := opts.LookPath("dewey"); err == nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				collect(initDewey(opts, logf)...)
			}()
		}
	}

	// --- Group B: Independent tools (each in own goroutine) ---

	// simpleTools defines Group B tools. Each entry maps a tool
	// name to its sentinel path (relative to TargetDir), result
	// name, display label, and optional extra args for ExecCmd.
	type simpleTool struct {
		name     string   // config + LookPath key
		sentinel string   // path to check (skip init if exists)
		result   string   // subToolResult.name
		label    string   // display text for logf
		args     []string // extra args after "init"
	}

	simpleTools := []simpleTool{
		{"replicator", ".uf/replicator", ".uf/replicator/",
			"Replicator workspace", nil},
		{"specify", ".specify", ".specify/",
			"Speckit framework", nil},
		{"openspec", filepath.Join("openspec", "config.yaml"),
			"openspec/", "OpenSpec framework",
			[]string{"--tools", "opencode"}},
		{"gaze", filepath.Join(".opencode", "agents", "gaze-reporter.md"),
			"gaze", "Gaze integration", nil},
	}

	for _, tool := range simpleTools {
		if shouldSkipTool(tool.name) {
			continue
		}
		if _, err := opts.LookPath(tool.name); err != nil {
			continue
		}
		// Capture loop variable for goroutine closure.
		tool := tool
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := initSimpleTool(opts, tool.name, tool.sentinel,
				tool.result, tool.label, tool.args, logf)
			if r != nil {
				collect(*r)
			}
		}()
	}

	// Wait for all concurrent groups to finish before
	// configuring opencode.json (FR-004, D4).
	wg.Wait()

	// Configure opencode.json with Dewey MCP server and Replicator MCP
	// entries. Runs after all sub-tool initialization steps.
	collect(configureOpencodeJSON(opts)...)

	return results
}

// initDewey runs the Dewey initialization sequence: init ->
// generateDeweySources -> index. The three steps MUST execute
// in order (FR-002). Returns results for all steps attempted.
func initDewey(opts *Options, logf func(string, ...interface{})) []subToolResult {
	var results []subToolResult
	deweyDir := filepath.Join(opts.TargetDir, ".uf", "dewey")

	if _, statErr := os.Stat(deweyDir); os.IsNotExist(statErr) {
		// First run: initialize workspace and build index.
		logf("  Initializing Dewey workspace...\n")
		if out, initErr := opts.ExecCmd("dewey", "init"); initErr != nil {
			return []subToolResult{{
				name: ".uf/dewey/", action: "failed",
				detail:  fmt.Sprintf("dewey init: %s", initErr),
				err:     initErr,
				output:  out,
			}}
		}
		results = append(results, subToolResult{
			name: ".uf/dewey/", action: "initialized"})

		// Auto-detect sibling repos for Dewey sources config.
		if sr := generateDeweySources(opts, false); sr != nil {
			results = append(results, *sr)
		}

		logf("  Indexing Dewey sources (this may take a moment)...\n")
		if out, idxErr := opts.ExecCmd("dewey", "index"); idxErr != nil {
			results = append(results, subToolResult{
				name: "dewey index", action: "failed",
				detail:  fmt.Sprintf("dewey index: %s", idxErr),
				err:     idxErr,
				output:  out,
			})
		} else {
			results = append(results, subToolResult{
				name: "dewey index", action: "completed"})
		}
		return results
	}

	if opts.Force {
		// Force: regenerate sources.yaml + re-index.
		if sr := generateDeweySources(opts, true); sr != nil {
			results = append(results, *sr)
		}
		logf("  Re-indexing Dewey sources...\n")
		if out, idxErr := opts.ExecCmd("dewey", "index"); idxErr != nil {
			results = append(results, subToolResult{
				name: "dewey index", action: "failed",
				detail:  fmt.Sprintf("dewey index: %s", idxErr),
				err:     idxErr,
				output:  out,
			})
		} else {
			results = append(results, subToolResult{
				name: "dewey index", action: "re-indexed"})
		}
	}
	return results
}

// initSimpleTool initializes a single sub-tool by checking a
// sentinel path and running "<name> init [args...]". Returns
// nil if the sentinel already exists (tool already initialized).
func initSimpleTool(opts *Options, name, sentinel, resultName, label string, extraArgs []string, logf func(string, ...interface{})) *subToolResult {
	sentinelPath := filepath.Join(opts.TargetDir, sentinel)
	if _, statErr := os.Stat(sentinelPath); !os.IsNotExist(statErr) {
		return nil // Already initialized.
	}

	logf("  Initializing %s...\n", label)
	args := append([]string{"init"}, extraArgs...)
	if out, initErr := opts.ExecCmd(name, args...); initErr != nil {
		return &subToolResult{
			name: resultName, action: "failed",
			detail:  fmt.Sprintf("%s init: %s", name, initErr),
			err:     initErr,
			output:  out,
		}
	}
	return &subToolResult{
		name: resultName, action: "initialized"}
}

// subToolSymbol returns the display symbol for a sub-tool result action.
// FR-021: created/configured/already configured/overwritten → ✓;
// skipped/dry-run → —; error/failed → ✗.
func subToolSymbol(action string) string {
	switch action {
	case "error", "failed":
		return "✗"
	case "skipped", "dry-run":
		return "—"
	default:
		// "initialized", "completed", "created", "configured",
		// "already configured", "overwritten"
		return "✓"
	}
}

// Next-step hint commands shown after scaffold summary.
const (
	hintDivisor = "Run /review-council to start a code review."
)

// printFileCategories writes per-category file lists (created,
// updated, overwritten, skipped) to w.
func printFileCategories(w io.Writer, r *Result) {
	if len(r.Created) > 0 {
		_, _ = fmt.Fprintf(w, "  created:     %d\n", len(r.Created))
		for _, f := range r.Created {
			_, _ = fmt.Fprintf(w, "    + %s\n", f)
		}
	}
	if len(r.Updated) > 0 {
		_, _ = fmt.Fprintf(w, "  updated:     %d\n", len(r.Updated))
		for _, f := range r.Updated {
			_, _ = fmt.Fprintf(w, "    ~ %s\n", f)
		}
	}
	if len(r.Overwritten) > 0 {
		_, _ = fmt.Fprintf(w, "  overwritten: %d\n", len(r.Overwritten))
		for _, f := range r.Overwritten {
			_, _ = fmt.Fprintf(w, "    ! %s\n", f)
		}
	}
	if len(r.Skipped) > 0 {
		_, _ = fmt.Fprintf(w, "  skipped:     %d (use --force to overwrite)\n", len(r.Skipped))
		for _, f := range r.Skipped {
			_, _ = fmt.Fprintf(w, "    - %s\n", f)
		}
	}
}

// printSubToolResults writes the sub-tool initialization section,
// including error messages and truncated output for failures.
func printSubToolResults(w io.Writer, subResults []subToolResult) {
	if len(subResults) == 0 {
		return
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Sub-tool initialization:")
	for _, sr := range subResults {
		symbol := subToolSymbol(sr.action)
		line := fmt.Sprintf("  %s %s %s", symbol, sr.name, sr.action)
		if sr.detail != "" {
			line += " (" + sr.detail + ")"
		}
		_, _ = fmt.Fprintln(w, line)
		// Show error and captured output for failed/error results.
		if sr.err != nil {
			_, _ = fmt.Fprintf(w, "                     Error: %v\n", sr.err)
		}
		if len(sr.output) > 0 && (sr.action == "failed" || sr.action == "error") {
			if trimmed := textutil.TruncateOutput(sr.output, 20); trimmed != "" {
				for _, line := range strings.Split(trimmed, "\n") {
					_, _ = fmt.Fprintf(w, "                     %s\n", line)
				}
			}
		}
	}
}

// printNextSteps writes context-aware footer guidance: language
// detection warnings, divisor hints, or numbered next steps.
func printNextSteps(w io.Writer, divisorOnly, langExplicit, langDetected bool, subResults []subToolResult) {
	_, _ = fmt.Fprintln(w)
	if divisorOnly && !langExplicit && !langDetected {
		_, _ = fmt.Fprintln(w, "  note: language not detected; deployed default convention pack only. Use --lang to specify.")
		_, _ = fmt.Fprintln(w)
	}
	if divisorOnly {
		_, _ = fmt.Fprintln(w, hintDivisor)
		return
	}
	// Show context-aware next steps.
	_, _ = fmt.Fprintln(w, "Next steps:")
	// Check if key tools are available to determine guidance.
	hasDewey := false
	for _, sr := range subResults {
		if sr.name == ".uf/dewey/" && (sr.action == "initialized" || sr.action == "completed") {
			hasDewey = true
		}
	}
	if !hasDewey && len(subResults) == 0 {
		_, _ = fmt.Fprintln(w, "  1. Run uf setup to install the full toolchain")
		_, _ = fmt.Fprintln(w, "  2. Run /speckit.constitution to create your project constitution")
		_, _ = fmt.Fprintln(w, "  3. Run uf doctor to verify your environment")
	} else {
		_, _ = fmt.Fprintln(w, "  1. Run /speckit.constitution to create your project constitution")
		_, _ = fmt.Fprintln(w, "  2. Run uf doctor to verify your environment")
		_, _ = fmt.Fprintln(w, "  3. Run /speckit.specify to start a strategic spec")
		_, _ = fmt.Fprintln(w, "  4. Run /opsx:propose to start a tactical change")
	}
}

// printSummary writes a human-readable summary of the scaffold
// result to the given writer. When divisorOnly is true, shows
// Divisor-specific hints instead of the standard hints.
// langExplicit indicates --lang was set; langDetected indicates
// auto-detection found a language. subResults reports sub-tool
// initialization outcomes (may be nil).
func printSummary(w io.Writer, divisorOnly, langExplicit, langDetected bool, r *Result, subResults []subToolResult) {
	total := len(r.Created) + len(r.Skipped) + len(r.Overwritten) + len(r.Updated)

	label := "uf init"
	if divisorOnly {
		label = "uf init (divisor)"
	}
	_, _ = fmt.Fprintf(w, "\n%s: %d files processed\n\n", label, total)

	printFileCategories(w, r)
	printSubToolResults(w, subResults)
	printNextSteps(w, divisorOnly, langExplicit, langDetected, subResults)
}

// assetPaths returns all relative paths of embedded assets.
// Used by tests to verify the asset manifest.
func assetPaths() ([]string, error) {
	var paths []string
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		paths = append(paths, strings.TrimPrefix(path, "assets/"))
		return nil
	})
	return paths, err
}

// assetContent returns the raw content of an embedded asset.
// Used by the drift detection test.
func assetContent(relPath string) ([]byte, error) {
	return assets.ReadFile("assets/" + relPath)
}

// generateDeweySources detects sibling repos and generates a
// multi-repo Dewey sources configuration. Called from initSubTools
// after `dewey init` creates the default sources.yaml and before
// `dewey index`. Skips if sources.yaml doesn't exist, or if the
// user has already customized it (> 1 source entry).
//
// Design decision: user-owned after creation. Once the user adds
// sources, uf init never overwrites. Detection uses simple
// `- id:` counting per Composability First — no YAML parsing
// dependency needed.
func generateDeweySources(opts *Options, force bool) *subToolResult {
	sourcesPath := filepath.Join(opts.TargetDir, ".uf", "dewey", "sources.yaml")

	// Skip if sources.yaml doesn't exist (dewey init didn't run
	// or was cleaned up).
	data, err := os.ReadFile(sourcesPath)
	if err != nil {
		return nil
	}

	// Skip if user has customized the file (more than the default
	// single-source config) — unless force is true (regenerate
	// even if customized).
	if !force && !isDefaultSourcesConfig(data) {
		return &subToolResult{
			name:   "dewey sources",
			action: "skipped",
			detail: "already customized",
		}
	}

	// Detect sibling repos: directories with .git/ in the parent dir.
	parentDir := filepath.Dir(opts.TargetDir)
	currentName := filepath.Base(opts.TargetDir)
	entries, readErr := os.ReadDir(parentDir)
	if readErr != nil {
		return nil
	}

	var siblings []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if e.Name() == currentName {
			continue
		}
		// Check for .git/ directory — indicates a git repo.
		gitDir := filepath.Join(parentDir, e.Name(), ".git")
		if info, statErr := os.Stat(gitDir); statErr == nil && info.IsDir() {
			siblings = append(siblings, e.Name())
		}
	}
	sort.Strings(siblings)

	// Extract GitHub org from git remote URL.
	org := extractGitHubOrg(opts)

	// Generate and write the multi-repo sources config.
	if writeErr := writeSourcesConfig(sourcesPath, currentName, siblings, parentDir, org); writeErr != nil {
		return &subToolResult{
			name:   "dewey sources",
			action: "failed",
			detail: writeErr.Error(),
		}
	}

	repoCount := 1 + len(siblings) // current + siblings
	return &subToolResult{
		name:   "dewey sources",
		action: "completed",
		detail: fmt.Sprintf("%d repos detected", repoCount),
	}
}

// isDefaultSourcesConfig returns true if the sources.yaml content
// has exactly 1 source entry (the default from `dewey init`).
// Uses simple `- id:` occurrence counting — if the user has added
// sources (> 1 entry), we treat the file as customized and skip
// overwriting.
func isDefaultSourcesConfig(data []byte) bool {
	return strings.Count(string(data), "- id:") <= 1
}

// extractGitHubOrg parses the GitHub organization name from the
// current repo's git remote URL. Supports both SSH and HTTPS
// formats. Returns empty string on any failure (non-GitHub remote,
// no remote configured, exec error) — graceful degradation per
// Constitution Principle II (Composability First).
func extractGitHubOrg(opts *Options) string {
	output, err := opts.ExecCmd("git", "remote", "get-url", "origin")
	if err != nil {
		return ""
	}

	url := strings.TrimSpace(string(output))

	// SSH format: git@github.com:ORG/repo.git
	if strings.HasPrefix(url, "git@github.com:") {
		trimmed := strings.TrimPrefix(url, "git@github.com:")
		trimmed = strings.TrimSuffix(trimmed, ".git")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0]
		}
		return ""
	}

	// HTTPS format: https://github.com/ORG/repo.git
	if strings.Contains(url, "github.com/") {
		idx := strings.Index(url, "github.com/")
		trimmed := url[idx+len("github.com/"):]
		trimmed = strings.TrimSuffix(trimmed, ".git")
		parts := strings.SplitN(trimmed, "/", 2)
		if len(parts) >= 1 && parts[0] != "" {
			return parts[0]
		}
		return ""
	}

	// Not a GitHub remote — omit GitHub source.
	return ""
}

// DevcontainerContent returns the raw content of the embedded
// devcontainer.json template. Used by uf sandbox init to write
// the template to .devcontainer/devcontainer.json. The template
// is embedded but NOT deployed by uf init (per design D7).
func DevcontainerContent() ([]byte, error) {
	return assets.ReadFile("assets/devcontainer/devcontainer.json")
}

// writeSourcesConfig generates a multi-repo Dewey sources.yaml
// with per-repo disk sources, a disk-org source for the parent
// directory, and optionally a GitHub API source if the org name
// was detected. The generated YAML is hand-crafted (not marshalled)
// to produce clean, commented output.
func writeSourcesConfig(path, currentName string, siblings []string, parentDir, org string) error {
	var b strings.Builder

	b.WriteString("# Auto-generated by uf init. Customize as needed.\n")
	b.WriteString("# This file is user-owned -- uf init will not\n")
	b.WriteString("# overwrite it after initial creation.\n")
	b.WriteString("\n")
	b.WriteString("sources:\n")

	// Per-repo disk sources (fine-grained provenance).
	b.WriteString("  # Per-repo disk sources (fine-grained provenance)\n")

	// Current repo first.
	b.WriteString("  - id: disk-local\n")
	b.WriteString("    type: disk\n")
	_, _ = fmt.Fprintf(&b, "    name: %s\n", currentName)
	b.WriteString("    config:\n")
	b.WriteString("      path: \".\"\n")

	// Sibling repos.
	for _, sib := range siblings {
		b.WriteString("\n")
		_, _ = fmt.Fprintf(&b, "  - id: disk-%s\n", sib)
		b.WriteString("    type: disk\n")
		_, _ = fmt.Fprintf(&b, "    name: %s\n", sib)
		b.WriteString("    config:\n")
		_, _ = fmt.Fprintf(&b, "      path: \"../%s\"\n", sib)
	}

	// Org-level disk source.
	b.WriteString("\n")
	b.WriteString("  # Org-level files (design papers, plans)\n")
	b.WriteString("  - id: disk-org\n")
	b.WriteString("    type: disk\n")
	b.WriteString("    name: org-workspace\n")
	b.WriteString("    config:\n")
	b.WriteString("      path: \"../\"\n")
	b.WriteString("      recursive: false\n")

	// GitHub API source (optional — only if org was detected).
	if org != "" {
		b.WriteString("\n")
		b.WriteString("  # GitHub API (issues, PRs, READMEs)\n")
		_, _ = fmt.Fprintf(&b, "  - id: github-%s\n", org)
		b.WriteString("    type: github\n")
		_, _ = fmt.Fprintf(&b, "    name: %s org\n", org)
		b.WriteString("    config:\n")
		_, _ = fmt.Fprintf(&b, "      org: %s\n", org)
		b.WriteString("      repos:\n")
		_, _ = fmt.Fprintf(&b, "        - %s\n", currentName)
		for _, sib := range siblings {
			_, _ = fmt.Fprintf(&b, "        - %s\n", sib)
		}
		b.WriteString("    refresh_interval: daily\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}
