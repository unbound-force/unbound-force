package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	Lang        string                                  // Language for convention pack selection (auto-detect if empty)
	Version     string                                  // Version string for marker comment (default: "dev")
	Stdout      io.Writer                               // Writer for summary output (default: os.Stdout)
	LookPath    func(string) (string, error)            // Finds a binary in PATH (default: exec.LookPath)
	ExecCmd     func(string, ...string) ([]byte, error) // Runs a command (default: exec.Command wrapper)
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
		//   specify/    -> .specify/
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

	// Initialize sub-tools after file scaffolding, before summary.
	subResults := initSubTools(&opts)

	printSummary(opts.Stdout, opts.DivisorOnly, langExplicit, langDetected, result, subResults)
	return result, nil
}

// knownAssetPrefixes enumerates the valid top-level prefixes
// in the embedded assets directory. Used by mapAssetPath to
// detect assets added under unexpected directories.
var knownAssetPrefixes = []string{"specify/", "opencode/", "openspec/"}

// mapAssetPath converts an embedded asset relative path to the
// output path in the target directory. The assets/ directory
// structure mirrors the target with these prefix mappings:
//
//	specify/  -> .specify/
//	opencode/ -> .opencode/
//	openspec/ -> openspec/  (no dot prefix)
func mapAssetPath(relPath string) string {
	switch {
	case strings.HasPrefix(relPath, "specify/"):
		return "." + relPath
	case strings.HasPrefix(relPath, "opencode/"):
		return "." + relPath
	case strings.HasPrefix(relPath, "openspec/"):
		// openspec/ paths pass through without dot prefix
		return relPath
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
	if strings.HasPrefix(relPath, "opencode/command/") {
		return true
	}
	// Skill files are tool-owned (maintained by unbound init).
	if strings.HasPrefix(relPath, "opencode/skill/") {
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
// shared opencode/unbound/packs/ location are included via
// isConventionPack() since they are essential for Divisor
// personas to function.
func isDivisorAsset(relPath string) bool {
	if strings.HasPrefix(relPath, "opencode/agents/divisor-") {
		return true
	}
	if relPath == "opencode/command/review-council.md" {
		return true
	}
	if isConventionPack(relPath) {
		return true
	}
	return false
}

// isConventionPack returns true if the asset is a convention
// pack file under opencode/unbound/packs/.
func isConventionPack(relPath string) bool {
	return strings.HasPrefix(relPath, "opencode/unbound/packs/")
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

	// Always deploy default packs
	if name == "default" || name == "default-custom" {
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

// insertMarkerAfterFrontmatter inserts the version marker after
// YAML frontmatter (if present) or appends it at the end.
// Frontmatter is delimited by "---\n" at the start and a
// matching "---\n" line.
func insertMarkerAfterFrontmatter(content []byte, marker string) []byte {
	s := string(content)

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
type subToolResult struct {
	name   string
	action string // "initialized", "completed", "failed", "skipped"
	detail string
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

	var results []subToolResult

	// Dewey: init + index if binary available and workspace absent.
	if _, err := opts.LookPath("dewey"); err == nil {
		deweyDir := filepath.Join(opts.TargetDir, ".dewey")
		if _, statErr := os.Stat(deweyDir); os.IsNotExist(statErr) {
			if _, initErr := opts.ExecCmd("dewey", "init"); initErr != nil {
				results = append(results, subToolResult{
					name: ".dewey/", action: "failed",
					detail: "dewey init failed"})
				return results // skip index if init failed
			}
			results = append(results, subToolResult{
				name: ".dewey/", action: "initialized"})

			if _, idxErr := opts.ExecCmd("dewey", "index"); idxErr != nil {
				results = append(results, subToolResult{
					name: "dewey index", action: "failed",
					detail: "dewey index failed"})
			} else {
				results = append(results, subToolResult{
					name: "dewey index", action: "completed"})
			}
		}
	}

	return results
}

// Next-step hint commands shown after scaffold summary.
const (
	hintDivisor = "Run /review-council to start a code review."
)

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
	fmt.Fprintf(w, "\n%s: %d files processed\n\n", label, total)

	if len(r.Created) > 0 {
		fmt.Fprintf(w, "  created:     %d\n", len(r.Created))
		for _, f := range r.Created {
			fmt.Fprintf(w, "    + %s\n", f)
		}
	}
	if len(r.Updated) > 0 {
		fmt.Fprintf(w, "  updated:     %d\n", len(r.Updated))
		for _, f := range r.Updated {
			fmt.Fprintf(w, "    ~ %s\n", f)
		}
	}
	if len(r.Overwritten) > 0 {
		fmt.Fprintf(w, "  overwritten: %d\n", len(r.Overwritten))
		for _, f := range r.Overwritten {
			fmt.Fprintf(w, "    ! %s\n", f)
		}
	}
	if len(r.Skipped) > 0 {
		fmt.Fprintf(w, "  skipped:     %d (use --force to overwrite)\n", len(r.Skipped))
		for _, f := range r.Skipped {
			fmt.Fprintf(w, "    - %s\n", f)
		}
	}

	// Sub-tool initialization results.
	if len(subResults) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Sub-tool initialization:")
		for _, sr := range subResults {
			symbol := "✓"
			if sr.action == "failed" {
				symbol = "✗"
			}
			line := fmt.Sprintf("  %s %s %s", symbol, sr.name, sr.action)
			if sr.detail != "" {
				line += " (" + sr.detail + ")"
			}
			fmt.Fprintln(w, line)
		}
	}

	fmt.Fprintln(w)
	if divisorOnly && !langExplicit && !langDetected {
		fmt.Fprintln(w, "  note: language not detected; deployed default convention pack only. Use --lang to specify.")
		fmt.Fprintln(w)
	}
	if divisorOnly {
		fmt.Fprintln(w, hintDivisor)
	} else {
		// Show context-aware next steps.
		fmt.Fprintln(w, "Next steps:")
		// Check if key tools are available to determine guidance.
		hasDewey := false
		if r != nil {
			// Use the opts passed to Run() — we check via the
			// presence of sub-tool results and file creation.
			// Since printSummary doesn't have direct access to
			// LookPath, we infer from subResults and created files.
			for _, sr := range subResults {
				if sr.name == ".dewey/" && (sr.action == "initialized" || sr.action == "completed") {
					hasDewey = true
				}
			}
			// If no sub-tool results but .dewey/ wasn't created,
			// tools may still be available — check if dewey was
			// already initialized (subResults would be empty).
			if len(subResults) == 0 {
				// No sub-tool actions means either DivisorOnly (handled above)
				// or dewey was already initialized or not available.
				// Default to showing uf setup as first step.
				hasDewey = false
			}
		}
		if !hasDewey && len(subResults) == 0 {
			fmt.Fprintln(w, "  1. Run uf setup to install the full toolchain")
			fmt.Fprintln(w, "  2. Run /speckit.constitution to create your project constitution")
			fmt.Fprintln(w, "  3. Run uf doctor to verify your environment")
		} else {
			fmt.Fprintln(w, "  1. Run /speckit.constitution to create your project constitution")
			fmt.Fprintln(w, "  2. Run uf doctor to verify your environment")
			fmt.Fprintln(w, "  3. Run /speckit.specify to start a strategic spec")
			fmt.Fprintln(w, "  4. Run /opsx:propose to start a tactical change")
		}
	}
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
