package scaffold

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed assets
var assets embed.FS

// Options configures a scaffold run.
type Options struct {
	TargetDir string    // Root dir to scaffold into (default: cwd)
	Force     bool      // Overwrite existing files when true
	Version   string    // Version string for marker comment (default: "dev")
	Stdout    io.Writer // Writer for summary output (default: os.Stdout)
}

// Result tracks the disposition of each scaffolded file.
type Result struct {
	Created     []string // Files written for the first time
	Skipped     []string // Files that existed and were not overwritten
	Overwritten []string // Files that existed and were replaced (Force=true)
	Updated     []string // Tool-owned files overwritten via overwrite-on-diff
}

// Run walks the embedded assets and writes them to the target directory.
// It applies file ownership rules and version markers.
func Run(opts Options) (*Result, error) {
	if opts.TargetDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		opts.TargetDir = cwd
	}
	if opts.Version == "" {
		opts.Version = "dev"
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}

	result := &Result{}
	marker := versionMarker(opts.Version)

	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Strip "assets/" prefix to get the relative path
		relPath := strings.TrimPrefix(path, "assets/")

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

		// Insert version marker
		out := insertMarkerAfterFrontmatter(content, marker)

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
		return nil, err
	}

	// Create empty directories for user content
	emptyDirs := []string{
		filepath.Join(opts.TargetDir, "openspec", "specs"),
		filepath.Join(opts.TargetDir, "openspec", "changes"),
	}
	for _, d := range emptyDirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, fmt.Errorf("create directory %s: %w", d, err)
		}
	}

	printSummary(opts.Stdout, result)
	return result, nil
}

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
	default:
		// openspec/ and any other paths pass through unchanged
		return relPath
	}
}

// isToolOwned returns true if the file is maintained by the
// unbound tool and should be overwritten when content differs.
// Tool-owned files: speckit commands, constitution-check
// command, and OpenSpec schema files.
func isToolOwned(relPath string) bool {
	if strings.HasPrefix(relPath, "openspec/schemas/") {
		return true
	}
	switch {
	case strings.HasPrefix(relPath, "opencode/command/speckit."):
		return true
	case relPath == "opencode/command/constitution-check.md":
		return true
	}
	return false
}

// versionMarker returns the HTML comment marker for provenance.
func versionMarker(version string) string {
	return fmt.Sprintf("<!-- scaffolded by unbound v%s -->", version)
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

// printSummary writes a human-readable summary of the scaffold
// result to the given writer.
func printSummary(w io.Writer, r *Result) {
	total := len(r.Created) + len(r.Skipped) + len(r.Overwritten) + len(r.Updated)
	fmt.Fprintf(w, "\nunbound init: %d files processed\n\n", total)

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

	fmt.Fprintln(w)
	fmt.Fprintln(w, "Run /speckit.specify to start a strategic spec.")
	fmt.Fprintln(w, "Run /opsx:propose to start a tactical change.")
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
