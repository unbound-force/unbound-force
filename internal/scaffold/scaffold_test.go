package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// findProjectRoot walks up from the current directory looking
// for go.mod to find the project root. Returns "" if not found.
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
			return ""
		}
		dir = parent
	}
}

// TestEmbeddedAssetsMatchSource verifies that every file under
// internal/scaffold/assets/ is byte-identical to the canonical
// source file at the repo root. This prevents drift between the
// embedded copies and the files developers actually use.
func TestEmbeddedAssets_MatchSource(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping drift detection")
	}

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
// the canonical source path at the repo root. Delegates to
// mapAssetPath to avoid duplicating the prefix mapping logic.
func mapAssetToSource(relPath string) string {
	return mapAssetPath(relPath)
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
	// OpenCode commands (14)
	"opencode/command/cobalt-crush.md",
	"opencode/command/finale.md",
	"opencode/command/uf-init.md",
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
	"opencode/command/unleash.md",
	// OpenCode agents — Divisor personas (5) + Cobalt-Crush (1) + Mx F coach (1) + constitution-check (1)
	"opencode/agents/cobalt-crush-dev.md",
	"opencode/agents/constitution-check.md",
	"opencode/agents/mx-f-coach.md", // Spec 007: Mx F coaching persona (user-owned, not in --divisor subset, not tool-owned)
	"opencode/agents/divisor-adversary.md",
	"opencode/agents/divisor-architect.md",
	"opencode/agents/divisor-guard.md",
	"opencode/agents/divisor-sre.md",
	"opencode/agents/divisor-testing.md",
	// OpenCode commands — includes review-council (11)
	"opencode/command/review-council.md",
	// Convention packs — shared by all heroes (6)
	"opencode/unbound/packs/default-custom.md",
	"opencode/unbound/packs/default.md",
	"opencode/unbound/packs/go-custom.md",
	"opencode/unbound/packs/go.md",
	"opencode/unbound/packs/severity.md",
	"opencode/unbound/packs/typescript-custom.md",
	"opencode/unbound/packs/typescript.md",
	// OpenSpec schema (5)
	"openspec/schemas/unbound-force/schema.yaml",
	"openspec/schemas/unbound-force/templates/proposal.md",
	"openspec/schemas/unbound-force/templates/spec.md",
	"openspec/schemas/unbound-force/templates/design.md",
	"openspec/schemas/unbound-force/templates/tasks.md",
	// OpenSpec config (1)
	"openspec/config.yaml",
	// Swarm skills (1)
	"opencode/skill/speckit-workflow/SKILL.md",
}

func TestAssetPaths_MatchExpected(t *testing.T) {
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
		".opencode/unbound/packs",
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
	if len(result.Updated) != 0 {
		t.Errorf("expected no updated files on identical re-run, got %d: %v",
			len(result.Updated), result.Updated)
	}
	if len(result.Skipped) != len(expectedAssetPaths) {
		t.Errorf("expected %d skipped files, got %d",
			len(expectedAssetPaths), len(result.Skipped))
	}

	// Verify a known tool-owned file is in Skipped
	foundToolSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "speckit.specify.md") {
			foundToolSkip = true
			break
		}
	}
	if !foundToolSkip {
		t.Error("expected tool-owned speckit.specify.md to be in Skipped list")
	}

	// Verify a known user-owned file is in Skipped
	foundUserSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "spec-template.md") {
			foundUserSkip = true
			break
		}
	}
	if !foundUserSkip {
		t.Error("expected user-owned spec-template.md to be in Skipped list")
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

	for _, relPath := range expectedAssetPaths {
		ext := filepath.Ext(relPath)
		if !markerFileExtensions[ext] {
			continue // unsupported extensions don't get markers
		}

		outRel := mapAssetPath(relPath)
		outPath := filepath.Join(dir, outRel)

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("read %s: %v", outRel, err)
			continue
		}

		marker := versionMarker("1.2.3", ext)

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

	// Version defaults to "0.0.0-dev" — check all supported files
	for _, relPath := range expectedAssetPaths {
		ext := filepath.Ext(relPath)
		if !markerFileExtensions[ext] {
			continue // unsupported extensions don't get markers
		}

		outRel := mapAssetPath(relPath)
		outPath := filepath.Join(dir, outRel)

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("read %s: %v", outRel, err)
			continue
		}

		marker := versionMarker("0.0.0-dev", ext)

		if !strings.Contains(string(content), marker) {
			t.Errorf("file %s does not contain dev version marker %q", outRel, marker)
		}
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
		// Tool-owned: all commands
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
		// Tool-owned: hypothetical future command (M1 fix)
		{"opencode/command/opsx.propose.md", true},
		// Tool-owned: OpenSpec schema
		{"openspec/schemas/unbound-force/schema.yaml", true},
		{"openspec/schemas/unbound-force/templates/proposal.md", true},
		// Tool-owned: convention packs (canonical)
		{"opencode/unbound/packs/go.md", true},
		{"opencode/unbound/packs/default.md", true},
		{"opencode/unbound/packs/typescript.md", true},
		// User-owned: convention packs (custom)
		{"opencode/unbound/packs/go-custom.md", false},
		{"opencode/unbound/packs/default-custom.md", false},
		{"opencode/unbound/packs/typescript-custom.md", false},
		// User-owned: agents (including Divisor personas and Cobalt-Crush)
		{"opencode/agents/divisor-guard.md", false},
		{"opencode/agents/divisor-architect.md", false},
		{"opencode/agents/cobalt-crush-dev.md", false},
		// User-owned: other
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
	mdMarker := "<!-- scaffolded by uf v1.0.0 -->"
	hashMarker := "# scaffolded by uf v1.0.0"

	tests := []struct {
		name     string
		input    string
		marker   string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			marker:   mdMarker,
			expected: mdMarker + "\n",
		},
		{
			name:     "no frontmatter",
			input:    "# Hello\n\nSome content.\n",
			marker:   mdMarker,
			expected: "# Hello\n\nSome content.\n" + mdMarker + "\n",
		},
		{
			name:   "with frontmatter",
			input:  "---\ntitle: Test\n---\n# Content\n",
			marker: mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n" +
				"# Content\n",
		},
		{
			name:     "unclosed frontmatter",
			input:    "---\ntitle: Test\nno closing\n",
			marker:   mdMarker,
			expected: "---\ntitle: Test\nno closing\n" + mdMarker + "\n",
		},
		{
			name:   "frontmatter with dashes in body",
			input:  "---\ntitle: Test\n---\nSome text\n---\nMore text\n",
			marker: mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n" +
				"Some text\n---\nMore text\n",
		},
		{
			name:     "bash script",
			input:    "#!/usr/bin/env bash\nset -e\n",
			marker:   hashMarker,
			expected: "#!/usr/bin/env bash\nset -e\n" + hashMarker + "\n",
		},
		{
			name:     "yaml document",
			input:    "---\nkey: value\n---\nmore: yaml\n",
			marker:   hashMarker,
			expected: "---\nkey: value\n---\n" + hashMarker + "\nmore: yaml\n",
		},
		{
			name:   "double insert on repeat call",
			input:  "# Hello\n" + mdMarker + "\n",
			marker: mdMarker,
			// insertMarkerAfterFrontmatter is not idempotent by design.
			// Run() achieves idempotency via the bytes.Equal check for
			// tool-owned files. This test documents the raw function behavior.
			expected: "# Hello\n" + mdMarker + "\n" + mdMarker + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insertMarkerAfterFrontmatter([]byte(tt.input), tt.marker)
			if string(got) != tt.expected {
				t.Errorf("got:\n%s\nexpected:\n%s", string(got), tt.expected)
			}
		})
	}
}

func TestPrintSummary_Output(t *testing.T) {
	t.Run("created_updated_skipped", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created:     []string{".specify/templates/spec-template.md", ".opencode/command/speckit.specify.md"},
			Updated:     []string{".opencode/command/speckit.plan.md"},
			Overwritten: []string{},
			Skipped:     []string{".specify/config.yaml"},
		}

		printSummary(&buf, false, false, true, r, nil)
		output := buf.String()

		// Verify total count line
		if !strings.Contains(output, "uf init: 4 files processed") {
			t.Errorf("expected total count of 4, got output:\n%s", output)
		}

		// Verify section headers
		if !strings.Contains(output, "created:     2") {
			t.Errorf("expected created count of 2, got output:\n%s", output)
		}
		if !strings.Contains(output, "updated:     1") {
			t.Errorf("expected updated count of 1, got output:\n%s", output)
		}
		if !strings.Contains(output, "skipped:     1") {
			t.Errorf("expected skipped count of 1, got output:\n%s", output)
		}

		// Verify file prefix characters
		if !strings.Contains(output, "+ .specify/templates/spec-template.md") {
			t.Errorf("expected '+' prefix for created files")
		}
		if !strings.Contains(output, "~ .opencode/command/speckit.plan.md") {
			t.Errorf("expected '~' prefix for updated files")
		}
		if !strings.Contains(output, "- .specify/config.yaml") {
			t.Errorf("expected '-' prefix for skipped files")
		}

		// Verify next-step guidance (no sub-tool results = suggest uf setup first)
		if !strings.Contains(output, "Next steps:") {
			t.Errorf("expected 'Next steps:' section")
		}
		if !strings.Contains(output, "uf setup") {
			t.Errorf("expected 'uf setup' hint when no sub-tool results")
		}
	})

	t.Run("divisor_mode", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/agents/divisor-guard.md", ".opencode/command/review-council.md"},
		}

		printSummary(&buf, true, false, true, r, nil)
		output := buf.String()

		if !strings.Contains(output, "uf init (divisor): 2 files processed") {
			t.Errorf("expected divisor label, got output:\n%s", output)
		}
		if !strings.Contains(output, "Run /review-council") {
			t.Error("expected review-council hint in divisor mode")
		}
		if strings.Contains(output, "Run /speckit.specify") {
			t.Error("speckit hint should not appear in divisor mode")
		}
		if strings.Contains(output, "Run /opsx:propose") {
			t.Error("opsx hint should not appear in divisor mode")
		}
	})

	t.Run("divisor_mode_no_lang", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/agents/divisor-guard.md"},
		}

		printSummary(&buf, true, false, false, r, nil)
		output := buf.String()

		if !strings.Contains(output, "language not detected") {
			t.Errorf("expected language detection warning, got:\n%s", output)
		}
	})

	t.Run("overwritten", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created:     []string{},
			Updated:     []string{},
			Overwritten: []string{".specify/templates/spec-template.md", ".opencode/command/speckit.specify.md"},
			Skipped:     []string{},
		}

		printSummary(&buf, false, false, true, r, nil)
		output := buf.String()

		if !strings.Contains(output, "uf init: 2 files processed") {
			t.Errorf("expected total count of 2, got output:\n%s", output)
		}
		if !strings.Contains(output, "overwritten: 2") {
			t.Errorf("expected overwritten count of 2, got output:\n%s", output)
		}
		if !strings.Contains(output, "! .specify/templates/spec-template.md") {
			t.Errorf("expected '!' prefix for overwritten files")
		}
		if !strings.Contains(output, "! .opencode/command/speckit.specify.md") {
			t.Errorf("expected '!' prefix for second overwritten file")
		}
	})
}

func TestRun_PrintSummaryIntegration(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := buf.String()
	expected := fmt.Sprintf("uf init: %d files processed", len(result.Created))
	if !strings.Contains(output, expected) {
		t.Errorf("expected summary to contain %q, got:\n%s", expected, output)
	}

	// Verify the output includes at least one specific file name
	if len(result.Created) > 0 {
		if !strings.Contains(output, result.Created[0]) {
			t.Errorf("expected output to contain file name %q", result.Created[0])
		}
	}
}

// knownNonEmbeddedFiles lists canonical source files that exist
// in .opencode/ but are intentionally NOT embedded in the unbound
// binary. These are local-only tooling files (e.g., installed by
// the Gaze scaffold) that are specific to this repository.
var knownNonEmbeddedFiles = map[string]bool{
	// Agents — local-only tooling, not scaffolded by uf init
	".opencode/agents/gaze-reporter.md":       true,
	".opencode/agents/gaze-test-generator.md": true,
	".opencode/agents/muti-mind-po.md":        true,
	// Legacy reviewer agents — superseded by divisor-* (Spec 019)
	".opencode/agents/reviewer-adversary.md": true,
	".opencode/agents/reviewer-architect.md": true,
	".opencode/agents/reviewer-guard.md":     true,
	".opencode/agents/reviewer-sre.md":       true,
	".opencode/agents/reviewer-testing.md":   true,
	// Commands — local-only tooling
	".opencode/command/cobalt-crush.md":               true,
	".opencode/command/gaze.md":                       true,
	".opencode/command/gaze-fix.md":                   true,
	".opencode/command/speckit.testreview.md":         true,
	".opencode/command/muti-mind.backlog-add.md":      true,
	".opencode/command/muti-mind.backlog-list.md":     true,
	".opencode/command/muti-mind.backlog-show.md":     true,
	".opencode/command/muti-mind.backlog-update.md":   true,
	".opencode/command/muti-mind.generate-stories.md": true,
	".opencode/command/muti-mind.init.md":             true,
	".opencode/command/muti-mind.prioritize.md":       true,
	".opencode/command/muti-mind.sync-project.md":     true,
	".opencode/command/muti-mind.sync-pull.md":        true,
	".opencode/command/muti-mind.sync-push.md":        true,
	".opencode/command/muti-mind.sync-status.md":      true,
	".opencode/command/muti-mind.sync.md":             true,
	// OpenSpec skill commands — local workflow tooling, not scaffolded by uf init
	".opencode/command/opsx-apply.md":   true,
	".opencode/command/opsx-archive.md": true,
	".opencode/command/opsx-explore.md": true,
	".opencode/command/opsx-propose.md": true,
	// Workflow commands — Spec 008 swarm orchestration, local-only
	".opencode/command/workflow-start.md":   true,
	".opencode/command/workflow-status.md":  true,
	".opencode/command/workflow-list.md":    true,
	".opencode/command/workflow-advance.md": true,
	".opencode/command/workflow-seed.md":    true,
	// Swarm skills — Spec 008, local-only
	".opencode/skill/unbound-force-heroes/SKILL.md": true,
}

func TestCanonicalSources_AreEmbedded(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping reverse drift detection")
	}

	// Build a set of embedded asset paths (mapped to source paths)
	embeddedSet := make(map[string]bool)
	for _, p := range expectedAssetPaths {
		srcRel := mapAssetToSource(p)
		embeddedSet[srcRel] = true
	}

	// Walk canonical source directories and check each file
	canonicalDirs := []string{
		".opencode/command",
		".opencode/agents",
		".opencode/unbound/packs",
		".specify/templates",
		".specify/scripts/bash",
		"openspec/schemas",
	}

	for _, dir := range canonicalDirs {
		fullDir := filepath.Join(root, dir)
		if _, err := os.Stat(fullDir); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(fullDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			relPath, _ := filepath.Rel(root, path)
			if knownNonEmbeddedFiles[relPath] {
				return nil // Explicitly excluded
			}
			if !embeddedSet[relPath] {
				t.Errorf("canonical source %s is not embedded and not in knownNonEmbeddedFiles exclusion list", relPath)
			}
			return nil
		})
		if err != nil {
			t.Errorf("walk %s: %v", dir, err)
		}
	}

	// Also check the two standalone config files
	for _, f := range []string{".specify/config.yaml", "openspec/config.yaml"} {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			if !embeddedSet[f] {
				t.Errorf("canonical source %s is not embedded", f)
			}
		}
	}
}

func TestMapAssetPath_Prefixes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"specify/templates/spec-template.md", ".specify/templates/spec-template.md"},
		{"opencode/command/speckit.specify.md", ".opencode/command/speckit.specify.md"},
		{"openspec/config.yaml", "openspec/config.yaml"},
		{"openspec/schemas/unbound-force/schema.yaml", "openspec/schemas/unbound-force/schema.yaml"},
		// Unknown prefix passes through unchanged (default branch)
		{"scripts/validate.sh", "scripts/validate.sh"},
	}

	for _, tt := range tests {
		got := mapAssetPath(tt.input)
		if got != tt.expected {
			t.Errorf("mapAssetPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsDivisorAsset(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Divisor agents
		{"opencode/agents/divisor-guard.md", true},
		{"opencode/agents/divisor-architect.md", true},
		{"opencode/agents/divisor-adversary.md", true},
		{"opencode/agents/divisor-sre.md", true},
		{"opencode/agents/divisor-testing.md", true},
		// Divisor command
		{"opencode/command/review-council.md", true},
		// Divisor convention packs
		{"opencode/unbound/packs/go.md", true},
		{"opencode/unbound/packs/default.md", true},
		{"opencode/unbound/packs/go-custom.md", true},
		{"opencode/unbound/packs/severity.md", true},
		// Non-Divisor assets
		{"opencode/agents/constitution-check.md", false},
		{"opencode/command/speckit.specify.md", false},
		{"opencode/command/speckit.plan.md", false},
		{"specify/templates/spec-template.md", false},
		{"openspec/config.yaml", false},
		// Non-Divisor: Cobalt-Crush agent
		{"opencode/agents/cobalt-crush-dev.md", false},
	}

	for _, tt := range tests {
		got := isDivisorAsset(tt.path)
		if got != tt.expected {
			t.Errorf("isDivisorAsset(%q) = %v, want %v",
				tt.path, got, tt.expected)
		}
	}
}

func TestDetectLang(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"go project", []string{"go.mod"}, "go"},
		{"typescript tsconfig", []string{"tsconfig.json"}, "typescript"},
		{"typescript package.json", []string{"package.json"}, "typescript"},
		{"python project", []string{"pyproject.toml"}, "python"},
		{"rust project", []string{"Cargo.toml"}, "rust"},
		{"no markers", []string{}, ""},
		{"go takes priority over ts", []string{"go.mod", "package.json"}, "go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0o644); err != nil {
					t.Fatalf("create marker %s: %v", f, err)
				}
			}
			got := detectLang(dir)
			if got != tt.expected {
				t.Errorf("detectLang() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestShouldDeployPack(t *testing.T) {
	tests := []struct {
		relPath  string
		lang     string
		expected bool
	}{
		// Non-pack files always pass
		{"opencode/agents/divisor-guard.md", "go", true},
		{"opencode/command/review-council.md", "go", true},
		// Default and severity packs always deploy (language-agnostic)
		{"opencode/unbound/packs/default.md", "go", true},
		{"opencode/unbound/packs/default-custom.md", "go", true},
		{"opencode/unbound/packs/default.md", "typescript", true},
		{"opencode/unbound/packs/severity.md", "go", true},
		{"opencode/unbound/packs/severity.md", "typescript", true},
		{"opencode/unbound/packs/severity.md", "default", true},
		// Matching language packs deploy
		{"opencode/unbound/packs/go.md", "go", true},
		{"opencode/unbound/packs/go-custom.md", "go", true},
		{"opencode/unbound/packs/typescript.md", "typescript", true},
		{"opencode/unbound/packs/typescript-custom.md", "typescript", true},
		// Non-matching language packs do NOT deploy
		{"opencode/unbound/packs/typescript.md", "go", false},
		{"opencode/unbound/packs/typescript-custom.md", "go", false},
		{"opencode/unbound/packs/go.md", "typescript", false},
		{"opencode/unbound/packs/go-custom.md", "typescript", false},
		// Default lang gets only default packs
		{"opencode/unbound/packs/go.md", "default", false},
		{"opencode/unbound/packs/default.md", "default", true},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s_lang=%s", filepath.Base(tt.relPath), tt.lang)
		t.Run(name, func(t *testing.T) {
			got := shouldDeployPack(tt.relPath, tt.lang)
			if got != tt.expected {
				t.Errorf("shouldDeployPack(%q, %q) = %v, want %v",
					tt.relPath, tt.lang, got, tt.expected)
			}
		})
	}
}

func TestRun_DivisorSubset(t *testing.T) {
	dir := t.TempDir()
	// Create a go.mod to trigger Go language detection
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("create go.mod: %v", err)
	}

	var buf bytes.Buffer
	result, err := Run(Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Version:     "1.0.0",
		Stdout:      &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Should create only Divisor files
	if len(result.Created) == 0 {
		t.Fatal("expected Divisor files to be created")
	}

	// Verify no speckit/openspec files created
	for _, f := range result.Created {
		if strings.HasPrefix(f, ".specify/") || strings.HasPrefix(f, "openspec/") {
			t.Errorf("DivisorOnly should not create %s", f)
		}
		if strings.Contains(f, "reviewer-") {
			t.Errorf("DivisorOnly should not create legacy reviewer files: %s", f)
		}
		if strings.Contains(f, "speckit.") {
			t.Errorf("DivisorOnly should not create speckit commands: %s", f)
		}
		if strings.Contains(f, "cobalt-crush") {
			t.Errorf("DivisorOnly should not create cobalt-crush files: %s", f)
		}
	}

	// Verify Divisor agents exist
	for _, agent := range []string{"divisor-guard.md", "divisor-architect.md", "divisor-adversary.md", "divisor-sre.md", "divisor-testing.md"} {
		found := false
		for _, f := range result.Created {
			if strings.HasSuffix(f, agent) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be created", agent)
		}
	}

	// Verify Go convention pack deployed (auto-detected)
	foundGoPack := false
	for _, f := range result.Created {
		if strings.HasSuffix(f, "go.md") && strings.Contains(f, "unbound/packs") {
			foundGoPack = true
			break
		}
	}
	if !foundGoPack {
		t.Error("expected Go convention pack to be deployed")
	}

	// Verify no openspec empty dirs
	specsDir := filepath.Join(dir, "openspec", "specs")
	if _, err := os.Stat(specsDir); !os.IsNotExist(err) {
		t.Error("DivisorOnly should not create openspec/specs directory")
	}

	// Verify summary mentions divisor
	output := buf.String()
	if !strings.Contains(output, "divisor") {
		t.Errorf("expected summary to mention divisor, got:\n%s", output)
	}
	if !strings.Contains(output, "review-council") {
		t.Errorf("expected summary to mention review-council hint")
	}
}

func TestRun_DivisorSubset_WithLangFlag(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := Run(Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Lang:        "typescript",
		Version:     "1.0.0",
		Stdout:      &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify TypeScript pack deployed
	foundTSPack := false
	for _, f := range result.Created {
		if strings.HasSuffix(f, "typescript.md") && strings.Contains(f, "unbound/packs") {
			foundTSPack = true
		}
	}
	if !foundTSPack {
		t.Error("expected TypeScript convention pack to be deployed")
	}

	// Verify Go pack NOT deployed
	for _, f := range result.Created {
		if strings.HasSuffix(f, "/go.md") && strings.Contains(f, "unbound/packs") {
			t.Error("Go convention pack should not be deployed when lang=typescript")
		}
	}

	// All 5 agent files still created
	agentCount := 0
	for _, f := range result.Created {
		if strings.Contains(f, "agents/divisor-") {
			agentCount++
		}
	}
	if agentCount != 5 {
		t.Errorf("expected 5 Divisor agent files, got %d", agentCount)
	}
}

func TestRun_DivisorSubset_DefaultFallback(t *testing.T) {
	dir := t.TempDir() // Empty — no language markers
	var buf bytes.Buffer

	result, err := Run(Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Version:     "1.0.0",
		Stdout:      &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify only default and severity packs deployed (language-agnostic)
	for _, f := range result.Created {
		if strings.Contains(f, "unbound/packs") {
			base := filepath.Base(f)
			if !strings.HasPrefix(base, "default") && base != "severity.md" {
				t.Errorf("expected only default/severity packs, got %s", f)
			}
		}
	}

	// Verify default.md and default-custom.md exist
	foundDefault := false
	foundDefaultCustom := false
	for _, f := range result.Created {
		if strings.HasSuffix(f, "default.md") {
			foundDefault = true
		}
		if strings.HasSuffix(f, "default-custom.md") {
			foundDefaultCustom = true
		}
	}
	if !foundDefault {
		t.Error("expected default.md pack to be deployed")
	}
	if !foundDefaultCustom {
		t.Error("expected default-custom.md pack to be deployed")
	}

	// Verify language detection warning in output
	output := buf.String()
	if !strings.Contains(output, "language not detected") {
		t.Errorf("expected language detection warning, got:\n%s", output)
	}
}

// TestAssetPaths_KnownPrefixes verifies all embedded assets use
// a recognized top-level prefix. Catches new directories added
// without updating mapAssetPath.
func TestAssetPaths_KnownPrefixes(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	for _, p := range paths {
		found := false
		for _, prefix := range knownAssetPrefixes {
			if strings.HasPrefix(p, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("asset %q does not match any known prefix %v — update mapAssetPath and knownAssetPrefixes",
				p, knownAssetPrefixes)
		}
	}
}

// TestScaffoldOutput_NoGraphthulhuReferences is a regression guard
// for FR-001/FR-002/SC-001: scaffolded files must not contain any
// graphthulhu or knowledge-graph references. Dewey replaces
// graphthulhu as the knowledge layer.
func TestScaffoldOutput_NoGraphthulhuReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Stale patterns that must NOT appear in scaffolded output.
	stalePatterns := []string{
		"graphthulhu",
		"knowledge-graph_",
		"knowledge-graph",
	}

	// Walk all generated files and search for stale patterns.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)
		relPath, _ := filepath.Rel(dir, path)

		for _, pattern := range stalePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains stale %q reference (SC-001 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestScaffoldOutput_NoBareUnboundReferences is a regression guard
// for FR-015/SC-003: scaffolded files must not contain bare
// `unbound init`, `unbound doctor`, `unbound setup`, or
// `unbound version` CLI references.
func TestScaffoldOutput_NoBareUnboundReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Bare patterns that must NOT appear in scaffolded output.
	// These are the old CLI command names before the rename.
	barePatterns := []string{
		"unbound init",
		"unbound doctor",
		"unbound setup",
		"unbound version",
	}

	// Walk all generated files and search for bare patterns.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)
		relPath, _ := filepath.Rel(dir, path)

		for _, pattern := range barePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains bare %q reference (FR-015 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestDivisorAgents_NoBareFRReferences is a regression guard for
// Spec 019 FR-008: all FR references in Divisor agent files must
// use the qualified "per Spec NNN FR-XXX" format.
func TestDivisorAgents_NoBareFRReferences(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	// Regex: bare "FR-NNN" not preceded by "per Spec NNN "
	// We check for any "FR-" followed by digits that is NOT
	// preceded by "per Spec" on the same line.
	for _, relPath := range paths {
		if !strings.HasPrefix(relPath, "opencode/agents/divisor-") {
			continue
		}

		content, readErr := assetContent(relPath)
		if readErr != nil {
			t.Errorf("read %s: %v", relPath, readErr)
			continue
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			// Skip lines that don't contain FR- references
			if !strings.Contains(line, "FR-") {
				continue
			}
			// Check that every FR- reference has "per Spec" qualifier
			// Find all FR-NNN occurrences and verify each is qualified
			idx := 0
			for {
				pos := strings.Index(line[idx:], "FR-")
				if pos < 0 {
					break
				}
				absPos := idx + pos
				// Check if "per Spec" (case-insensitive) appears before this FR- on the same line
				prefix := strings.ToLower(line[:absPos])
				if !strings.Contains(prefix, "per spec") {
					t.Errorf("%s:%d: bare FR reference without 'per Spec' qualifier: %s",
						relPath, i+1, strings.TrimSpace(line))
					break
				}
				idx = absPos + 3
			}
		}
	}
}

// TestRun_LegacyFileWarning verifies that uf init warns about
// previously scaffolded reviewer-*.md files per Spec 019 FR-003a.
func TestRun_LegacyFileWarning(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create legacy reviewer files.
	legacyFiles := []string{
		"reviewer-adversary.md",
		"reviewer-architect.md",
		"reviewer-guard.md",
		"reviewer-sre.md",
		"reviewer-testing.md",
	}
	for _, f := range legacyFiles {
		if err := os.WriteFile(filepath.Join(agentsDir, f), []byte("legacy"), 0o644); err != nil {
			t.Fatalf("create %s: %v", f, err)
		}
	}

	var buf bytes.Buffer
	_, err := Run(Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := buf.String()

	// Verify warning is printed.
	if !strings.Contains(output, "Legacy reviewer agents detected") {
		t.Errorf("expected legacy warning, got:\n%s", output)
	}

	// Verify file names are listed.
	for _, f := range legacyFiles {
		if !strings.Contains(output, f) {
			t.Errorf("expected %s in warning, got:\n%s", f, output)
		}
	}

	// Verify removal command is suggested.
	if !strings.Contains(output, "rm .opencode/agents/reviewer-*.md") {
		t.Errorf("expected removal command in warning, got:\n%s", output)
	}

	// Verify legacy files are NOT deleted (FR-003a).
	for _, f := range legacyFiles {
		if _, err := os.Stat(filepath.Join(agentsDir, f)); os.IsNotExist(err) {
			t.Errorf("legacy file %s should NOT be deleted", f)
		}
	}
}

// --- Sub-tool initialization tests ---

// stubScaffoldLookPath returns a function that simulates exec.LookPath.
func stubScaffoldLookPath(found map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := found[name]; ok {
			return path, nil
		}
		return "", fmt.Errorf("executable %q not found", name)
	}
}

// scaffoldCmdRecorder records ExecCmd calls for scaffold tests.
type scaffoldCmdRecorder struct {
	calls  []string
	errors map[string]error
}

func (r *scaffoldCmdRecorder) execCmd(name string, args ...string) ([]byte, error) {
	key := name
	if len(args) > 0 {
		key = name + " " + strings.Join(args, " ")
	}
	r.calls = append(r.calls, key)

	if err, ok := r.errors[key]; ok {
		return nil, err
	}
	return []byte(""), nil
}

func TestInitSubTools_DeweyAvailable(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 4 results: config.yaml + dewey init + dewey index + opencode.json.
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d: %v", len(results), results)
	}

	if results[0].name != ".unbound-force/config.yaml" || results[0].action != "initialized" {
		t.Errorf("expected .unbound-force/config.yaml initialized, got %s %s", results[0].name, results[0].action)
	}
	if results[1].name != ".dewey/" || results[1].action != "initialized" {
		t.Errorf("expected .dewey/ initialized, got %s %s", results[1].name, results[1].action)
	}
	if results[2].name != "dewey index" || results[2].action != "completed" {
		t.Errorf("expected dewey index completed, got %s %s", results[2].name, results[2].action)
	}
	if results[3].name != "opencode.json" || results[3].action != "created" {
		t.Errorf("expected opencode.json created, got %s %s", results[3].name, results[3].action)
	}

	// Verify commands were called.
	expectedCalls := []string{"dewey init", "dewey index"}
	for _, expected := range expectedCalls {
		found := false
		for _, call := range rec.calls {
			if call == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q, got calls: %v", expected, rec.calls)
		}
	}
}

func TestInitSubTools_DeweyNotAvailable(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No dewey
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 2 results: config.yaml initialized + opencode.json skipped.
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d: %v", len(results), results)
	}
	if len(results) > 0 && results[0].name != ".unbound-force/config.yaml" {
		t.Errorf("expected config.yaml result, got %s", results[0].name)
	}
	if len(results) > 1 && results[1].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[1].name)
	}

	// No commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands, got: %v", rec.calls)
	}
}

func TestInitSubTools_DeweyAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	// Create .dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 2 results: config.yaml initialized + opencode.json created
	// (.dewey/ already exists, dewey in PATH → mcp.dewey added).
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d: %v", len(results), results)
	}
	if len(results) > 0 && results[0].name != ".unbound-force/config.yaml" {
		t.Errorf("expected config.yaml result, got %s", results[0].name)
	}
	if len(results) > 1 && results[1].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[1].name)
	}

	// dewey init should NOT have been called.
	for _, call := range rec.calls {
		if call == "dewey init" {
			t.Error("dewey init should NOT be called when .dewey/ already exists")
		}
	}
}

func TestInitSubTools_DeweyInitFails(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"dewey init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 3 results: config.yaml initialized + dewey init failed + opencode.json created.
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}

	if results[0].name != ".unbound-force/config.yaml" || results[0].action != "initialized" {
		t.Errorf("expected config.yaml initialized, got %s %s", results[0].name, results[0].action)
	}
	if results[1].name != ".dewey/" || results[1].action != "failed" {
		t.Errorf("expected .dewey/ failed, got %s %s", results[1].name, results[1].action)
	}
	if results[2].name != "opencode.json" || results[2].action != "created" {
		t.Errorf("expected opencode.json created, got %s %s", results[2].name, results[2].action)
	}

	// dewey index should NOT have been called.
	for _, call := range rec.calls {
		if call == "dewey index" {
			t.Error("dewey index should NOT be called when dewey init fails")
		}
	}
}

func TestInitSubTools_DivisorOnly(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir:   dir,
		DivisorOnly: true,
		LookPath:    stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:     rec.execCmd,
	}

	results := initSubTools(opts)

	// Should return nil — DivisorOnly skips all sub-tool init.
	if results != nil {
		t.Errorf("expected nil results in DivisorOnly mode, got %v", results)
	}

	// No commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands in DivisorOnly mode, got: %v", rec.calls)
	}
}

func TestPrintSummary_NextSteps(t *testing.T) {
	t.Run("with_sub_tools", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/command/speckit.specify.md"},
		}
		subResults := []subToolResult{
			{name: ".dewey/", action: "initialized"},
			{name: "dewey index", action: "completed"},
		}

		printSummary(&buf, false, false, true, r, subResults)
		output := buf.String()

		// Should show sub-tool results.
		if !strings.Contains(output, "Sub-tool initialization:") {
			t.Errorf("expected sub-tool section, got:\n%s", output)
		}
		if !strings.Contains(output, ".dewey/ initialized") {
			t.Errorf("expected dewey init result, got:\n%s", output)
		}
		if !strings.Contains(output, "dewey index completed") {
			t.Errorf("expected dewey index result, got:\n%s", output)
		}

		// Should show full next steps (not uf setup first).
		if !strings.Contains(output, "Next steps:") {
			t.Errorf("expected 'Next steps:' section")
		}
		if !strings.Contains(output, "/speckit.constitution") {
			t.Errorf("expected constitution hint")
		}
		if !strings.Contains(output, "uf doctor") {
			t.Errorf("expected doctor hint")
		}
	})

	t.Run("without_sub_tools", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/command/speckit.specify.md"},
		}

		printSummary(&buf, false, false, true, r, nil)
		output := buf.String()

		// Should suggest uf setup as first step.
		if !strings.Contains(output, "uf setup") {
			t.Errorf("expected 'uf setup' hint when no sub-tool results, got:\n%s", output)
		}
	})

	t.Run("sub_tool_failure", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/command/speckit.specify.md"},
		}
		subResults := []subToolResult{
			{name: ".dewey/", action: "failed", detail: "dewey init failed"},
		}

		printSummary(&buf, false, false, true, r, subResults)
		output := buf.String()

		// Should show failure with ✗ symbol.
		if !strings.Contains(output, "✗") {
			t.Errorf("expected failure symbol, got:\n%s", output)
		}
		if !strings.Contains(output, "failed") {
			t.Errorf("expected 'failed' in output, got:\n%s", output)
		}
	})
}

// --- Workflow config file scaffold tests ---

func TestInitSubTools_CreatesWorkflowConfig(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No dewey
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 1 result: config.yaml initialized.
	foundConfig := false
	for _, r := range results {
		if r.name == ".unbound-force/config.yaml" && r.action == "initialized" {
			foundConfig = true
		}
	}
	if !foundConfig {
		t.Errorf("expected .unbound-force/config.yaml initialized, got %v", results)
	}

	// Verify file exists with commented content.
	configPath := filepath.Join(dir, ".unbound-force", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config.yaml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "# workflow:") {
		t.Error("config.yaml should contain commented-out workflow section")
	}
	if !strings.Contains(text, "#   execution_modes:") {
		t.Error("config.yaml should contain commented-out execution_modes")
	}
	if !strings.Contains(text, "#     define: swarm") {
		t.Error("config.yaml should contain commented-out define: swarm example")
	}
	if !strings.Contains(text, "#   spec_review: false") {
		t.Error("config.yaml should contain commented-out spec_review")
	}
}

// --- Dewey auto-sources tests ---

func TestGenerateDeweySources_SiblingsDetected(t *testing.T) {
	// Create a parent dir with the "current" project and 3 sibling repos.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write default sources.yaml (single source entry).
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(currentDir, ".dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	// Create 3 sibling repos with .git/ directories.
	for _, name := range []string{"gaze", "dewey", "website"} {
		sibDir := filepath.Join(parentDir, name)
		if err := os.MkdirAll(filepath.Join(sibDir, ".git"), 0o755); err != nil {
			t.Fatalf("mkdir sibling %s: %v", name, err)
		}
	}

	// Create a non-repo directory (no .git/) — should be ignored.
	if err := os.MkdirAll(filepath.Join(parentDir, "not-a-repo"), 0o755); err != nil {
		t.Fatalf("mkdir not-a-repo: %v", err)
	}

	// Stub ExecCmd to return a GitHub SSH remote.
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			key := name
			if len(args) > 0 {
				key = name + " " + strings.Join(args, " ")
			}
			if key == "git remote get-url origin" {
				return []byte("git@github.com:unbound-force/my-project.git\n"), nil
			}
			return rec.execCmd(name, args...)
		},
	}

	result := generateDeweySources(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "completed" {
		t.Errorf("expected action 'completed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "4 repos detected") {
		t.Errorf("expected '4 repos detected' in detail, got %q", result.detail)
	}

	// Read the generated sources.yaml and verify content.
	content, err := os.ReadFile(filepath.Join(currentDir, ".dewey", "sources.yaml"))
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	text := string(content)

	// Verify per-repo disk sources.
	if !strings.Contains(text, "- id: disk-local") {
		t.Error("expected disk-local source")
	}
	if !strings.Contains(text, "- id: disk-gaze") {
		t.Error("expected disk-gaze source")
	}
	if !strings.Contains(text, "- id: disk-dewey") {
		t.Error("expected disk-dewey source")
	}
	if !strings.Contains(text, "- id: disk-website") {
		t.Error("expected disk-website source")
	}

	// Verify disk-org source.
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}

	// Verify GitHub source with repos list.
	if !strings.Contains(text, "- id: github-unbound-force") {
		t.Error("expected github-unbound-force source")
	}
	if !strings.Contains(text, "org: unbound-force") {
		t.Error("expected org: unbound-force in GitHub config")
	}
	// Verify repos list includes current + siblings.
	if !strings.Contains(text, "        - my-project") {
		t.Error("expected my-project in repos list")
	}
	if !strings.Contains(text, "        - gaze") {
		t.Error("expected gaze in repos list")
	}

	// Verify non-repo directory was NOT included.
	if strings.Contains(text, "not-a-repo") {
		t.Error("non-repo directory should not appear in sources")
	}

	// Verify sibling paths use relative notation.
	if !strings.Contains(text, "path: \"../gaze\"") {
		t.Error("expected relative path for gaze sibling")
	}
}

func TestGenerateDeweySources_NoSiblings(t *testing.T) {
	// Create a parent dir with only the current project.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "lonely-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write default sources.yaml.
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(currentDir, ".dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	// No ExecCmd stub needed — extractGitHubOrg will fail gracefully.
	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("no remote")
		},
	}

	result := generateDeweySources(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "completed" {
		t.Errorf("expected action 'completed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "1 repos detected") {
		t.Errorf("expected '1 repos detected', got %q", result.detail)
	}

	// Read generated sources.yaml.
	content, err := os.ReadFile(filepath.Join(currentDir, ".dewey", "sources.yaml"))
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	text := string(content)

	// Should have disk-local + disk-org only.
	if !strings.Contains(text, "- id: disk-local") {
		t.Error("expected disk-local source")
	}
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}

	// Should NOT have GitHub source (no remote).
	if strings.Contains(text, "type: github") {
		t.Error("should not have GitHub source when no remote")
	}
}

func TestGenerateDeweySources_AlreadyCustomized(t *testing.T) {
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write a customized sources.yaml with 3 source entries.
	customSources := `sources:
  - id: disk-local
    type: disk
    config:
      path: "."
  - id: disk-other
    type: disk
    config:
      path: "../other"
  - id: github-myorg
    type: github
    config:
      org: myorg
`
	sourcesPath := filepath.Join(currentDir, ".dewey", "sources.yaml")
	if err := os.WriteFile(sourcesPath, []byte(customSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("should not be called")
		},
	}

	result := generateDeweySources(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "skipped" {
		t.Errorf("expected action 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "already customized") {
		t.Errorf("expected 'already customized' in detail, got %q", result.detail)
	}

	// Verify file was NOT overwritten.
	content, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	if string(content) != customSources {
		t.Error("customized sources.yaml should not have been overwritten")
	}
}

func TestExtractGitHubOrg_SSH(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("git@github.com:unbound-force/repo.git\n"), nil
		},
	}

	org := extractGitHubOrg(opts)
	if org != "unbound-force" {
		t.Errorf("expected 'unbound-force', got %q", org)
	}
}

func TestExtractGitHubOrg_HTTPS(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("https://github.com/unbound-force/repo.git\n"), nil
		},
	}

	org := extractGitHubOrg(opts)
	if org != "unbound-force" {
		t.Errorf("expected 'unbound-force', got %q", org)
	}
}

func TestExtractGitHubOrg_NonGitHub(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("https://gitlab.com/myorg/repo.git\n"), nil
		},
	}

	org := extractGitHubOrg(opts)
	if org != "" {
		t.Errorf("expected empty string for non-GitHub remote, got %q", org)
	}
}

func TestExtractGitHubOrg_NoRemote(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("fatal: No such remote 'origin'")
		},
	}

	org := extractGitHubOrg(opts)
	if org != "" {
		t.Errorf("expected empty string when no remote, got %q", org)
	}
}

func TestIsDefaultSourcesConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "default single source",
			input:    "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n",
			expected: true,
		},
		{
			name:     "empty file",
			input:    "",
			expected: true,
		},
		{
			name:     "no sources at all",
			input:    "# empty config\n",
			expected: true,
		},
		{
			name: "customized with 3 sources",
			input: `sources:
  - id: disk-local
    type: disk
  - id: disk-other
    type: disk
  - id: github-org
    type: github
`,
			expected: false,
		},
		{
			name: "customized with 2 sources",
			input: `sources:
  - id: disk-local
    type: disk
  - id: disk-other
    type: disk
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDefaultSourcesConfig([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("isDefaultSourcesConfig() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInitSubTools_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	// Create existing config with custom content.
	ufDir := filepath.Join(dir, ".unbound-force")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	customContent := "workflow:\n  execution_modes:\n    define: swarm\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(customContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have a config result — file already exists.
	for _, r := range results {
		if r.name == ".unbound-force/config.yaml" {
			t.Errorf("expected no config result (file exists), got %s %s", r.name, r.action)
		}
	}

	// Verify file was NOT overwritten.
	content, err := os.ReadFile(filepath.Join(ufDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(content) != customContent {
		t.Error("existing config.yaml should not have been overwritten")
	}
}

// --- configureOpencodeJSON tests ---

// parseOpencodeJSON is a test helper that parses opencode.json from a dir.
func parseOpencodeJSON(t *testing.T, dir string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse opencode.json: %v", err)
	}
	return m
}

// getMCPDewey extracts the mcp.dewey entry from a parsed opencode.json.
func getMCPDewey(t *testing.T, ocMap map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key not found")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	deweyRaw, ok := mcpMap["dewey"]
	if !ok {
		t.Fatal("mcp.dewey not found")
	}
	var dewey map[string]json.RawMessage
	if err := json.Unmarshal(deweyRaw, &dewey); err != nil {
		t.Fatalf("parse mcp.dewey: %v", err)
	}
	return dewey
}

// getPlugins extracts the plugin array from a parsed opencode.json.
func getPlugins(t *testing.T, ocMap map[string]json.RawMessage) []string {
	t.Helper()
	pluginRaw, ok := ocMap["plugin"]
	if !ok {
		t.Fatal("plugin key not found")
	}
	var plugins []string
	if err := json.Unmarshal(pluginRaw, &plugins); err != nil {
		t.Fatalf("parse plugin: %v", err)
	}
	return plugins
}

// Phase 3: US1 tests — Fresh Repo Init

func TestConfigureOpencodeJSON_Create(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(results), results)
	}
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify $schema.
	var schema string
	if err := json.Unmarshal(ocMap["$schema"], &schema); err != nil {
		t.Fatalf("parse $schema: %v", err)
	}
	if schema != "https://opencode.ai/config.json" {
		t.Errorf("$schema = %q, want opencode.ai URL", schema)
	}

	// Verify mcp.dewey entry.
	dewey := getMCPDewey(t, ocMap)
	var deweyType string
	_ = json.Unmarshal(dewey["type"], &deweyType) //nolint:errcheck // test helper
	if deweyType != "local" {
		t.Errorf("mcp.dewey.type = %q, want 'local'", deweyType)
	}
	var cmd []string
	_ = json.Unmarshal(dewey["command"], &cmd) //nolint:errcheck // test helper
	expectedCmd := []string{"dewey", "serve", "--vault", "."}
	if len(cmd) != len(expectedCmd) {
		t.Errorf("mcp.dewey.command = %v, want %v", cmd, expectedCmd)
	} else {
		for i := range cmd {
			if cmd[i] != expectedCmd[i] {
				t.Errorf("mcp.dewey.command[%d] = %q, want %q", i, cmd[i], expectedCmd[i])
			}
		}
	}
	var enabled bool
	_ = json.Unmarshal(dewey["enabled"], &enabled) //nolint:errcheck // test helper
	if !enabled {
		t.Error("mcp.dewey.enabled should be true")
	}

	// Verify plugin array.
	plugins := getPlugins(t, ocMap)
	if len(plugins) != 1 || plugins[0] != "opencode-swarm-plugin" {
		t.Errorf("plugin = %v, want [opencode-swarm-plugin]", plugins)
	}
}

func TestConfigureOpencodeJSON_DeweyOnly(t *testing.T) {
	dir := t.TempDir()
	// No .hive/ directory.

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should have mcp.dewey.
	if _, ok := ocMap["mcp"]; !ok {
		t.Fatal("mcp key should exist")
	}

	// Should NOT have plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist when no .hive/")
	}
}

func TestConfigureOpencodeJSON_HiveOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No dewey
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should NOT have mcp key.
	if _, ok := ocMap["mcp"]; ok {
		t.Error("mcp key should not exist when dewey not available")
	}

	// Should have plugin array.
	plugins := getPlugins(t, ocMap)
	if len(plugins) != 1 || plugins[0] != "opencode-swarm-plugin" {
		t.Errorf("plugin = %v, want [opencode-swarm-plugin]", plugins)
	}
}

func TestConfigureOpencodeJSON_Neither(t *testing.T) {
	dir := t.TempDir()
	// No dewey, no .hive/.

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "skipped" {
		t.Errorf("expected action 'skipped', got %q", results[0].action)
	}
	if results[0].detail != "nothing to configure" {
		t.Errorf("expected detail 'nothing to configure', got %q", results[0].detail)
	}

	// No file should be created.
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Error("opencode.json should not be created when nothing to configure")
	}
}

// Phase 4: US2 tests — Idempotent Re-run

func TestConfigureOpencodeJSON_Idempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	// Create opencode.json with both entries already present.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true,
      "type": "local"
    }
  },
  "plugin": [
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", results[0].action)
	}

	// Verify file is unchanged.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != existing {
		t.Error("file should be byte-identical when already configured")
	}
}

func TestConfigureOpencodeJSON_IdempotentWithOtherPlugins(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  },
  "plugin": [
    "other-plugin",
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", results[0].action)
	}

	// Verify file is unchanged — both plugins preserved.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != existing {
		t.Error("file should be byte-identical when already configured")
	}
}

func TestConfigureOpencodeJSON_AddMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	// Has plugin but no mcp.dewey.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.dewey was added.
	getMCPDewey(t, ocMap)

	// Verify existing plugin array is preserved.
	plugins := getPlugins(t, ocMap)
	if len(plugins) != 1 || plugins[0] != "opencode-swarm-plugin" {
		t.Errorf("plugin = %v, want [opencode-swarm-plugin]", plugins)
	}
}

func TestConfigureOpencodeJSON_PreserveCustom(t *testing.T) {
	dir := t.TempDir()

	// Has a custom MCP server.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "my-custom-server": {
      "type": "local",
      "command": ["my-server", "start"],
      "enabled": true
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify custom server preserved.
	mcpRaw := ocMap["mcp"]
	var mcpMap map[string]json.RawMessage
	_ = json.Unmarshal(mcpRaw, &mcpMap) //nolint:errcheck // test helper
	if _, ok := mcpMap["my-custom-server"]; !ok {
		t.Error("custom MCP server should be preserved")
	}
	if _, ok := mcpMap["dewey"]; !ok {
		t.Error("mcp.dewey should be added alongside custom server")
	}
}

func TestConfigureOpencodeJSON_LegacyMcpServers(t *testing.T) {
	dir := t.TempDir()

	// Uses legacy mcpServers key.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcpServers": {
    "dewey": {
      "command": "dewey",
      "args": ["serve", "--vault", "."]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", results[0].action)
	}

	// Verify no duplicate mcp.dewey added.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != existing {
		t.Error("file should be unchanged when legacy mcpServers.dewey exists")
	}
}

func TestConfigureOpencodeJSON_Malformed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if results[0].detail != "malformed JSON" {
		t.Errorf("expected detail 'malformed JSON', got %q", results[0].detail)
	}

	// Verify file not modified.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != "{invalid json" {
		t.Error("malformed file should not be modified")
	}
}

func TestConfigureOpencodeJSON_ReadPermissionDenied(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ReadFile: func(path string) ([]byte, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if !strings.Contains(results[0].detail, "read failed") {
		t.Errorf("expected detail to contain 'read failed', got %q", results[0].detail)
	}
}

func TestConfigureOpencodeJSON_WriteFail(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		WriteFile: func(path string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("disk full")
		},
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "failed" {
		t.Errorf("expected action 'failed', got %q", results[0].action)
	}
	if !strings.Contains(results[0].detail, "write failed") {
		t.Errorf("expected detail to contain 'write failed', got %q", results[0].detail)
	}
}

func TestConfigureOpencodeJSON_ByteIdentical(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	// First run — creates the file.
	configureOpencodeJSON(opts)
	data1, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))

	// Second run — should be "already configured" and file unchanged.
	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected 'already configured' on second run, got %q", results[0].action)
	}

	data2, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if !bytes.Equal(data1, data2) {
		t.Error("output should be byte-identical on re-run (FR-016)")
	}
}

// Phase 5: US3 tests — Force Overwrite

func TestConfigureOpencodeJSON_Force(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}

	// Stale mcp.dewey with --include-hidden flag.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--include-hidden", "--vault", "."],
      "enabled": true
    }
  },
  "plugin": [
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "overwritten" {
		t.Errorf("expected action 'overwritten', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.dewey was overwritten with correct command (no --include-hidden).
	dewey := getMCPDewey(t, ocMap)
	var cmd []string
	_ = json.Unmarshal(dewey["command"], &cmd) //nolint:errcheck // test helper
	expectedCmd := []string{"dewey", "serve", "--vault", "."}
	if len(cmd) != len(expectedCmd) {
		t.Fatalf("command = %v, want %v", cmd, expectedCmd)
	}
	for i := range cmd {
		if cmd[i] != expectedCmd[i] {
			t.Errorf("command[%d] = %q, want %q", i, cmd[i], expectedCmd[i])
		}
	}

	// Verify plugin array is NOT duplicated — still exactly one entry.
	plugins := getPlugins(t, ocMap)
	count := 0
	for _, p := range plugins {
		if p == "opencode-swarm-plugin" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 opencode-swarm-plugin entry, got %d", count)
	}
}

func TestConfigureOpencodeJSON_ForceCorrect(t *testing.T) {
	dir := t.TempDir()

	// Correct mcp.dewey — force should still overwrite.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "overwritten" {
		t.Errorf("expected action 'overwritten', got %q", results[0].action)
	}
}

func TestConfigureOpencodeJSON_DryRun(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		DryRun:    true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "dry-run" {
		t.Errorf("expected action 'dry-run', got %q", results[0].action)
	}

	// Verify no file was written.
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Error("no file should be written in dry-run mode")
	}
}

// --- Dewey force re-index tests ---

func TestInitSubTools_DeweyForceReindex(t *testing.T) {
	dir := t.TempDir()
	// Create .dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have dewey index re-indexed result.
	foundReindex := false
	for _, r := range results {
		if r.name == "dewey index" && r.action == "re-indexed" {
			foundReindex = true
		}
	}
	if !foundReindex {
		t.Errorf("expected dewey index re-indexed result, got %v", results)
	}

	// Verify dewey index was called.
	indexCalled := false
	for _, call := range rec.calls {
		if call == "dewey index" {
			indexCalled = true
		}
	}
	if !indexCalled {
		t.Errorf("expected dewey index command, got calls: %v", rec.calls)
	}

	// Verify dewey init was NOT called (.dewey/ already exists).
	for _, call := range rec.calls {
		if call == "dewey init" {
			t.Error("dewey init should NOT be called when .dewey/ already exists")
		}
	}
}

func TestInitSubTools_DeweyExistsNoForce(t *testing.T) {
	dir := t.TempDir()
	// Create .dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		Force:     false,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have any dewey-related results (skipped silently).
	for _, r := range results {
		if r.name == ".dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when Force=false and .dewey/ exists: %s %s", r.name, r.action)
		}
	}

	// Verify no dewey commands were called.
	for _, call := range rec.calls {
		if call == "dewey init" || call == "dewey index" {
			t.Errorf("unexpected dewey command when Force=false: %s", call)
		}
	}
}

// Phase 8: Integration test (T032a)

func TestInitSubTools_OpencodeJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".hive"), 0o755); err != nil {
		t.Fatalf("mkdir .hive: %v", err)
	}
	// Create .dewey/ so dewey init is skipped.
	if err := os.MkdirAll(filepath.Join(dir, ".dewey"), 0o755); err != nil {
		t.Fatalf("mkdir .dewey: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Find the opencode.json result.
	var ocResult *subToolResult
	for i := range results {
		if results[i].name == "opencode.json" {
			ocResult = &results[i]
			break
		}
	}
	if ocResult == nil {
		t.Fatal("opencode.json result not found in initSubTools output")
	}
	if ocResult.action != "created" {
		t.Errorf("expected opencode.json action 'created', got %q", ocResult.action)
	}

	// Verify file exists with expected content.
	ocMap := parseOpencodeJSON(t, dir)
	getMCPDewey(t, ocMap)
	plugins := getPlugins(t, ocMap)
	if len(plugins) != 1 || plugins[0] != "opencode-swarm-plugin" {
		t.Errorf("plugin = %v, want [opencode-swarm-plugin]", plugins)
	}
}
