package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// findProjectRoot walks up from the current directory looking
// for go.mod to find the project root.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// TestEmbeddedAssetsMatchSource verifies that every file under
// internal/scaffold/assets/ is byte-identical to the canonical
// source file at the repo root. This prevents drift between the
// embedded copies and the files developers actually use.
func TestEmbeddedAssetsMatchSource(t *testing.T) {
	root := findProjectRoot(t)

	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("no embedded assets found")
	}

	for _, relPath := range paths {
		// Map asset path to canonical source path
		srcRel := mapAssetToSource(relPath)
		srcPath := filepath.Join(root, srcRel)

		embedded, err := assetContent(relPath)
		if err != nil {
			t.Errorf("read embedded %s: %v", relPath, err)
			continue
		}

		source, err := os.ReadFile(srcPath)
		if err != nil {
			t.Errorf("read source %s: %v (expected canonical source at %s)", relPath, err, srcPath)
			continue
		}

		if !bytes.Equal(embedded, source) {
			t.Errorf("drift detected: internal/scaffold/assets/%s differs from %s\n"+
				"Run: cp %s internal/scaffold/assets/%s",
				relPath, srcRel, srcRel, relPath)
		}
	}
}

// mapAssetToSource converts an embedded asset relative path to
// the canonical source path at the repo root.
//
//	specify/   -> .specify/
//	opencode/  -> .opencode/
//	openspec/  -> openspec/
func mapAssetToSource(relPath string) string {
	switch {
	case strings.HasPrefix(relPath, "specify/"):
		return "." + relPath
	case strings.HasPrefix(relPath, "opencode/"):
		return "." + relPath
	default:
		return relPath
	}
}

// expectedAssetPaths is the canonical list of embedded assets.
// Update this list when adding or removing assets.
var expectedAssetPaths = []string{
	// Speckit templates (6)
	"specify/templates/agent-file-template.md",
	"specify/templates/checklist-template.md",
	"specify/templates/constitution-template.md",
	"specify/templates/plan-template.md",
	"specify/templates/spec-template.md",
	"specify/templates/tasks-template.md",
	// Speckit config (1)
	"specify/config.yaml",
	// Speckit scripts (5)
	"specify/scripts/bash/check-prerequisites.sh",
	"specify/scripts/bash/common.sh",
	"specify/scripts/bash/create-new-feature.sh",
	"specify/scripts/bash/setup-plan.sh",
	"specify/scripts/bash/update-agent-context.sh",
	// OpenCode commands (10)
	"opencode/command/constitution-check.md",
	"opencode/command/speckit.analyze.md",
	"opencode/command/speckit.checklist.md",
	"opencode/command/speckit.clarify.md",
	"opencode/command/speckit.constitution.md",
	"opencode/command/speckit.implement.md",
	"opencode/command/speckit.plan.md",
	"opencode/command/speckit.specify.md",
	"opencode/command/speckit.tasks.md",
	"opencode/command/speckit.taskstoissues.md",
	// OpenCode agents (1)
	"opencode/agents/constitution-check.md",
	// OpenSpec schema (5)
	"openspec/schemas/unbound-force/schema.yaml",
	"openspec/schemas/unbound-force/templates/proposal.md",
	"openspec/schemas/unbound-force/templates/spec.md",
	"openspec/schemas/unbound-force/templates/design.md",
	"openspec/schemas/unbound-force/templates/tasks.md",
	// OpenSpec config (1)
	"openspec/config.yaml",
}

func TestAssetPaths(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	sort.Strings(paths)
	expected := make([]string, len(expectedAssetPaths))
	copy(expected, expectedAssetPaths)
	sort.Strings(expected)

	if len(paths) != len(expected) {
		t.Errorf("expected %d assets, got %d", len(expected), len(paths))
		t.Logf("expected: %v", expected)
		t.Logf("got:      %v", paths)
		return
	}

	for i := range paths {
		if paths[i] != expected[i] {
			t.Errorf("asset mismatch at index %d: expected %q, got %q", i, expected[i], paths[i])
		}
	}
}

func TestRun_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// All files should be created on first run
	if len(result.Created) == 0 {
		t.Error("expected files to be created")
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected no skipped files, got %d", len(result.Skipped))
	}
	if len(result.Overwritten) != 0 {
		t.Errorf("expected no overwritten files, got %d", len(result.Overwritten))
	}
	if len(result.Updated) != 0 {
		t.Errorf("expected no updated files, got %d", len(result.Updated))
	}

	// Verify expected directory structure
	expectedDirs := []string{
		".specify/templates",
		".specify/scripts/bash",
		".opencode/command",
		".opencode/agents",
		"openspec/specs",
		"openspec/changes",
	}
	for _, d := range expectedDirs {
		full := filepath.Join(dir, d)
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}

	// Verify created file count matches expected assets
	if len(result.Created) != len(expectedAssetPaths) {
		t.Errorf("expected %d created files, got %d", len(expectedAssetPaths), len(result.Created))
	}
}

func TestRun_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run creates everything
	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Second run should skip user-owned, skip identical tool-owned
	buf.Reset()
	result, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	if len(result.Created) != 0 {
		t.Errorf("expected no created files on second run, got %d", len(result.Created))
	}
	// All files should be skipped (user-owned skipped, tool-owned
	// skipped because content is identical)
	totalSkipped := len(result.Skipped)
	totalUpdated := len(result.Updated)
	if totalSkipped+totalUpdated != len(expectedAssetPaths) {
		t.Errorf("expected %d skipped+updated files, got skipped=%d updated=%d",
			len(expectedAssetPaths), totalSkipped, totalUpdated)
	}
}

func TestRun_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run
	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Second run with --force
	buf.Reset()
	result, err := Run(Options{
		TargetDir: dir,
		Force:     true,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("force Run() error: %v", err)
	}

	if len(result.Overwritten) != len(expectedAssetPaths) {
		t.Errorf("expected %d overwritten files, got %d",
			len(expectedAssetPaths), len(result.Overwritten))
	}
	if len(result.Created) != 0 {
		t.Errorf("expected no created files with force, got %d", len(result.Created))
	}
}

func TestRun_VersionMarker(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.2.3",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	marker := "<!-- scaffolded by unbound v1.2.3 -->"

	for _, relPath := range expectedAssetPaths {
		outRel := mapAssetPath(relPath)
		outPath := filepath.Join(dir, outRel)

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("read %s: %v", outRel, err)
			continue
		}

		if !strings.Contains(string(content), marker) {
			t.Errorf("file %s does not contain version marker %q", outRel, marker)
		}
	}
}

func TestRun_VersionMarkerDev(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := Run(Options{
		TargetDir: dir,
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Version defaults to "dev"
	marker := "<!-- scaffolded by unbound vdev -->"

	// Check at least one file
	outPath := filepath.Join(dir, ".specify", "templates", "spec-template.md")
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read spec-template.md: %v", err)
	}

	if !strings.Contains(string(content), marker) {
		t.Errorf("file does not contain dev version marker %q", marker)
	}
}

func TestRun_OverwriteOnDiff_ToolOwned(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run
	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Modify a tool-owned file on disk
	toolFile := filepath.Join(dir, ".opencode", "command", "speckit.specify.md")
	if err := os.WriteFile(toolFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("modify tool-owned file: %v", err)
	}

	// Modify a user-owned file on disk
	userFile := filepath.Join(dir, ".specify", "templates", "spec-template.md")
	if err := os.WriteFile(userFile, []byte("user modified"), 0o644); err != nil {
		t.Fatalf("modify user-owned file: %v", err)
	}

	// Re-run
	buf.Reset()
	result, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Tool-owned file should be updated
	if len(result.Updated) == 0 {
		t.Error("expected at least one updated file (tool-owned)")
	}

	foundToolUpdate := false
	for _, f := range result.Updated {
		if strings.Contains(f, "speckit.specify.md") {
			foundToolUpdate = true
			break
		}
	}
	if !foundToolUpdate {
		t.Error("expected speckit.specify.md to be in Updated list")
	}

	// User-owned file should be skipped
	foundUserSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "spec-template.md") {
			foundUserSkip = true
			break
		}
	}
	if !foundUserSkip {
		t.Error("expected spec-template.md to be in Skipped list")
	}

	// Verify tool-owned file content was restored
	restored, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("read restored tool file: %v", err)
	}
	if string(restored) == "modified content" {
		t.Error("tool-owned file was not restored to canonical content")
	}

	// Verify user-owned file was NOT overwritten
	preserved, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("read preserved user file: %v", err)
	}
	if string(preserved) != "user modified" {
		t.Error("user-owned file should not have been overwritten")
	}
}

func TestRun_OverwriteOnDiff_SkipsIdentical(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run
	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Re-run without any modifications
	buf.Reset()
	result, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Tool-owned files with identical content should be skipped
	if len(result.Updated) != 0 {
		t.Errorf("expected no updated files when content is identical, got %d: %v",
			len(result.Updated), result.Updated)
	}
}

func TestIsToolOwned(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Tool-owned
		{"opencode/command/speckit.specify.md", true},
		{"opencode/command/speckit.plan.md", true},
		{"opencode/command/speckit.tasks.md", true},
		{"opencode/command/speckit.clarify.md", true},
		{"opencode/command/speckit.analyze.md", true},
		{"opencode/command/speckit.checklist.md", true},
		{"opencode/command/speckit.implement.md", true},
		{"opencode/command/speckit.constitution.md", true},
		{"opencode/command/speckit.taskstoissues.md", true},
		{"opencode/command/constitution-check.md", true},
		{"openspec/schemas/unbound-force/schema.yaml", true},
		{"openspec/schemas/unbound-force/templates/proposal.md", true},
		// User-owned
		{"specify/templates/spec-template.md", false},
		{"specify/templates/plan-template.md", false},
		{"specify/scripts/bash/common.sh", false},
		{"opencode/agents/constitution-check.md", false},
		{"openspec/config.yaml", false},
		{"specify/config.yaml", false},
	}

	for _, tt := range tests {
		got := isToolOwned(tt.path)
		if got != tt.expected {
			t.Errorf("isToolOwned(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestRun_SchemaDistribution(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run creates everything
	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Modify a schema file (tool-owned) and config (user-owned)
	schemaFile := filepath.Join(dir, "openspec", "schemas",
		"unbound-force", "schema.yaml")
	configFile := filepath.Join(dir, "openspec", "config.yaml")

	if err := os.WriteFile(schemaFile, []byte("modified schema"), 0o644); err != nil {
		t.Fatalf("modify schema file: %v", err)
	}
	if err := os.WriteFile(configFile, []byte("user config"), 0o644); err != nil {
		t.Fatalf("modify config file: %v", err)
	}

	// Re-run without --force
	buf.Reset()
	result, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Schema file (tool-owned) should be updated
	foundSchemaUpdate := false
	for _, f := range result.Updated {
		if strings.Contains(f, "schema.yaml") {
			foundSchemaUpdate = true
			break
		}
	}
	if !foundSchemaUpdate {
		t.Error("expected schema.yaml to be in Updated list")
	}

	// Config file (user-owned) should be skipped
	foundConfigSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "config.yaml") {
			foundConfigSkip = true
			break
		}
	}
	if !foundConfigSkip {
		t.Error("expected config.yaml to be in Skipped list")
	}

	// Verify schema was restored
	restored, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("read restored schema: %v", err)
	}
	if string(restored) == "modified schema" {
		t.Error("schema file was not restored to canonical content")
	}

	// Verify config was preserved
	preserved, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("read preserved config: %v", err)
	}
	if string(preserved) != "user config" {
		t.Error("config file should not have been overwritten")
	}
}

func TestInsertMarkerAfterFrontmatter(t *testing.T) {
	marker := "<!-- scaffolded by unbound v1.0.0 -->"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			expected: marker + "\n",
		},
		{
			name:     "no frontmatter",
			input:    "# Hello\n\nSome content.\n",
			expected: "# Hello\n\nSome content.\n" + marker + "\n",
		},
		{
			name:  "with frontmatter",
			input: "---\ntitle: Test\n---\n# Content\n",
			expected: "---\ntitle: Test\n---\n" + marker + "\n" +
				"# Content\n",
		},
		{
			name:     "unclosed frontmatter",
			input:    "---\ntitle: Test\nno closing\n",
			expected: "---\ntitle: Test\nno closing\n" + marker + "\n",
		},
		{
			name:  "frontmatter with dashes in body",
			input: "---\ntitle: Test\n---\nSome text\n---\nMore text\n",
			expected: "---\ntitle: Test\n---\n" + marker + "\n" +
				"Some text\n---\nMore text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insertMarkerAfterFrontmatter([]byte(tt.input), marker)
			if string(got) != tt.expected {
				t.Errorf("got:\n%s\nexpected:\n%s", string(got), tt.expected)
			}
		})
	}
}
