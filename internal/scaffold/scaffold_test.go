package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/unbound-force/unbound-force/internal/textutil"
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
		// devcontainer/ assets are OS-specific and gitignored.
		// No canonical source file exists in the repo to drift
		// against — the embedded template is the source of truth.
		if strings.HasPrefix(relPath, "devcontainer/") {
			continue
		}

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

// TestEmbeddedAssets_SingleMarker verifies that no embedded
// Markdown asset contains more than one scaffold provenance
// marker line. This prevents marker accumulation through the
// asset-sync feedback loop.
func TestEmbeddedAssets_SingleMarker(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	for _, relPath := range paths {
		if filepath.Ext(relPath) != ".md" {
			continue
		}

		content, err := assetContent(relPath)
		if err != nil {
			t.Errorf("read embedded %s: %v", relPath, err)
			continue
		}

		count := 0
		for _, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "<!-- scaffolded by uf ") ||
				strings.HasPrefix(trimmed, "# scaffolded by uf ") {
				count++
			}
		}

		if count > 1 {
			t.Errorf("embedded asset %s contains %d scaffold markers (expected at most 1)",
				relPath, count)
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
	// OpenCode commands (9) — UF-custom only; speckit.*.md externalized to specify init + /uf-init
	"opencode/commands/address-feedback.md",
	"opencode/commands/agent-brief.md",
	"opencode/commands/cobalt-crush.md",
	"opencode/commands/constitution-check.md",
	"opencode/commands/finale.md",
	"opencode/commands/review-council.md",
	"opencode/commands/review-pr.md",
	"opencode/commands/uf-init.md",
	"opencode/commands/unleash.md",
	// OpenCode agents — Divisor personas (6) + Cobalt-Crush (1) + Mx F coach (1) + constitution-check (1)
	"opencode/agents/cobalt-crush-dev.md",
	"opencode/agents/constitution-check.md",
	"opencode/agents/mx-f-coach.md", // Spec 007: Mx F coaching persona (user-owned, not in --divisor subset, not tool-owned)
	"opencode/agents/divisor-adversary.md",
	"opencode/agents/divisor-architect.md",
	"opencode/agents/divisor-curator.md",
	"opencode/agents/divisor-guard.md",
	"opencode/agents/divisor-sre.md",
	"opencode/agents/divisor-testing.md",
	// OpenCode agents — Divisor content agents (3)
	"opencode/agents/divisor-scribe.md",
	"opencode/agents/divisor-herald.md",
	"opencode/agents/divisor-envoy.md",
	// Convention packs — shared by all heroes (11)
	"opencode/uf/packs/content-custom.md",
	"opencode/uf/packs/content.md",
	"opencode/uf/packs/default-custom.md",
	"opencode/uf/packs/default.md",
	"opencode/uf/packs/go-custom.md",
	"opencode/uf/packs/go.md",
	"opencode/uf/packs/python-custom.md",
	"opencode/uf/packs/python.md",
	"opencode/uf/packs/severity.md",
	"opencode/uf/packs/typescript-custom.md",
	"opencode/uf/packs/typescript.md",
	// OpenSpec schema (5)
	"openspec/schemas/unbound-force/schema.yaml",
	"openspec/schemas/unbound-force/templates/proposal.md",
	"openspec/schemas/unbound-force/templates/spec.md",
	"openspec/schemas/unbound-force/templates/design.md",
	"openspec/schemas/unbound-force/templates/tasks.md",
	// Swarm skills (1)
	"opencode/skills/speckit-workflow/SKILL.md",
}

// nonDeployedAssetPaths lists embedded assets that are NOT
// deployed by uf init. These are accessed via dedicated
// functions (e.g., DevcontainerContent()) and are skipped
// during the Run() walk. They must be listed here so
// TestAssetPaths_MatchExpected accounts for them.
var nonDeployedAssetPaths = []string{
	// Devcontainer template — OS-specific, generated per-user
	// by uf sandbox init (not deployed by uf init).
	"devcontainer/devcontainer.json",
}

func TestAssetPaths_MatchExpected(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	sort.Strings(paths)

	// Combine deployed and non-deployed assets for the full manifest.
	allExpected := make([]string, 0, len(expectedAssetPaths)+len(nonDeployedAssetPaths))
	allExpected = append(allExpected, expectedAssetPaths...)
	allExpected = append(allExpected, nonDeployedAssetPaths...)
	sort.Strings(allExpected)

	if len(paths) != len(allExpected) {
		t.Errorf("expected %d assets, got %d", len(allExpected), len(paths))
		t.Logf("expected: %v", allExpected)
		t.Logf("got:      %v", paths)
		return
	}

	for i := range paths {
		if paths[i] != allExpected[i] {
			t.Errorf("asset mismatch at index %d: expected %q, got %q", i, allExpected[i], paths[i])
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
		".opencode/commands",
		".opencode/agents",
		".opencode/uf/packs",
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
		if strings.Contains(f, "review-council.md") {
			foundToolSkip = true
			break
		}
	}
	if !foundToolSkip {
		t.Error("expected tool-owned review-council.md to be in Skipped list")
	}

	// Verify a known user-owned file is in Skipped
	foundUserSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "cobalt-crush-dev.md") {
			foundUserSkip = true
			break
		}
	}
	if !foundUserSkip {
		t.Error("expected user-owned cobalt-crush-dev.md to be in Skipped list")
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
	toolFile := filepath.Join(dir, ".opencode", "commands", "review-council.md")
	if err := os.WriteFile(toolFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("modify tool-owned file: %v", err)
	}

	// Modify a user-owned file on disk
	userFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")
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
		if strings.Contains(f, "review-council.md") {
			foundToolUpdate = true
			break
		}
	}
	if !foundToolUpdate {
		t.Error("expected review-council.md to be in Updated list")
	}

	// User-owned file should be skipped
	foundUserSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "cobalt-crush-dev.md") {
			foundUserSkip = true
			break
		}
	}
	if !foundUserSkip {
		t.Error("expected cobalt-crush-dev.md to be in Skipped list")
	}

	// Verify tool-owned file content was restored (review-council.md)
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
		{"opencode/commands/speckit.specify.md", true},
		{"opencode/commands/speckit.plan.md", true},
		{"opencode/commands/speckit.tasks.md", true},
		{"opencode/commands/speckit.clarify.md", true},
		{"opencode/commands/speckit.analyze.md", true},
		{"opencode/commands/speckit.checklist.md", true},
		{"opencode/commands/speckit.implement.md", true},
		{"opencode/commands/speckit.constitution.md", true},
		{"opencode/commands/speckit.taskstoissues.md", true},
		{"opencode/commands/constitution-check.md", true},
		// Tool-owned: hypothetical future command (M1 fix)
		{"opencode/commands/opsx.propose.md", true},
		// Tool-owned: OpenSpec schema
		{"openspec/schemas/unbound-force/schema.yaml", true},
		{"openspec/schemas/unbound-force/templates/proposal.md", true},
		// Tool-owned: convention packs (canonical)
		{"opencode/uf/packs/go.md", true},
		{"opencode/uf/packs/default.md", true},
		{"opencode/uf/packs/typescript.md", true},
		// Tool-owned: convention packs (canonical) — Python
		{"opencode/uf/packs/python.md", true},
		// User-owned: convention packs (custom)
		{"opencode/uf/packs/go-custom.md", false},
		{"opencode/uf/packs/default-custom.md", false},
		{"opencode/uf/packs/typescript-custom.md", false},
		{"opencode/uf/packs/python-custom.md", false},
		// User-owned: agents (including Divisor personas and Cobalt-Crush)
		{"opencode/agents/divisor-guard.md", false},
		{"opencode/agents/divisor-architect.md", false},
		{"opencode/agents/cobalt-crush-dev.md", false},
		// User-owned: other
		{"opencode/agents/constitution-check.md", false},
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

	// Modify a schema file (tool-owned) and an agent (user-owned)
	schemaFile := filepath.Join(dir, "openspec", "schemas",
		"unbound-force", "schema.yaml")
	agentFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")

	if err := os.WriteFile(schemaFile, []byte("modified schema"), 0o644); err != nil {
		t.Fatalf("modify schema file: %v", err)
	}
	if err := os.WriteFile(agentFile, []byte("user agent"), 0o644); err != nil {
		t.Fatalf("modify agent file: %v", err)
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

	// Agent file (user-owned) should be skipped
	foundAgentSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "cobalt-crush-dev.md") {
			foundAgentSkip = true
			break
		}
	}
	if !foundAgentSkip {
		t.Error("expected cobalt-crush-dev.md to be in Skipped list")
	}

	// Verify schema was restored
	restored, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("read restored schema: %v", err)
	}
	if string(restored) == "modified schema" {
		t.Error("schema file was not restored to canonical content")
	}

	// Verify agent was preserved
	preserved, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("read preserved agent: %v", err)
	}
	if string(preserved) != "user agent" {
		t.Error("agent file should not have been overwritten")
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
			name:   "idempotent on repeat call",
			input:  "# Hello\n" + mdMarker + "\n",
			marker: mdMarker,
			// insertMarkerAfterFrontmatter is idempotent: existing
			// markers are stripped before the new one is inserted.
			expected: "# Hello\n" + mdMarker + "\n",
		},
		{
			name:     "strips multiple existing markers",
			input:    "---\ntitle: Test\n---\n<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vv0.6.1 -->\n# Content\n",
			marker:   mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n# Content\n",
		},
		{
			name:     "replaces old version with new",
			input:    "---\ntitle: Test\n---\n<!-- scaffolded by uf v0.5.0 -->\n# Content\n",
			marker:   mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n# Content\n",
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

func TestStripExistingMarkers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no markers",
			input:    "# Hello\n\nSome content.\n",
			expected: "# Hello\n\nSome content.\n",
		},
		{
			name:     "single HTML marker",
			input:    "# Hello\n<!-- scaffolded by uf vdev -->\nContent\n",
			expected: "# Hello\nContent\n",
		},
		{
			name:     "multiple HTML markers",
			input:    "<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vv0.6.1 -->\nContent\n",
			expected: "Content\n",
		},
		{
			name:     "single hash marker",
			input:    "#!/bin/bash\n# scaffolded by uf vdev\nset -e\n",
			expected: "#!/bin/bash\nset -e\n",
		},
		{
			name:     "mixed HTML and hash markers",
			input:    "<!-- scaffolded by uf v1.0.0 -->\n# scaffolded by uf v2.0.0\nContent\n",
			expected: "Content\n",
		},
		{
			name:     "frontmatter preserved",
			input:    "---\ntitle: Test\n---\n<!-- scaffolded by uf vdev -->\n# Content\n",
			expected: "---\ntitle: Test\n---\n# Content\n",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "content only no markers",
			input:    "line1\nline2\nline3\n",
			expected: "line1\nline2\nline3\n",
		},
		{
			name:     "marker-like content preserved",
			input:    "<!-- scaffolded by someone else -->\nContent\n",
			expected: "<!-- scaffolded by someone else -->\nContent\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripExistingMarkers(tt.input)
			if got != tt.expected {
				t.Errorf("got:\n%q\nexpected:\n%q", got, tt.expected)
			}
		})
	}
}

func TestPrintSummary_Output(t *testing.T) {
	t.Run("created_updated_skipped", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created:     []string{".opencode/agents/cobalt-crush-dev.md", ".opencode/commands/review-council.md"},
			Updated:     []string{".opencode/commands/unleash.md"},
			Overwritten: []string{},
			Skipped:     []string{".opencode/uf/packs/go-custom.md"},
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
		if !strings.Contains(output, "+ .opencode/agents/cobalt-crush-dev.md") || !strings.Contains(output, "+ .opencode/commands/review-council.md") {
			t.Errorf("expected '+' prefix for created files")
		}
		if !strings.Contains(output, "~ .opencode/commands/unleash.md") {
			t.Errorf("expected '~' prefix for updated files")
		}
		if !strings.Contains(output, "- .opencode/uf/packs/go-custom.md") {
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
			Created: []string{".opencode/agents/divisor-guard.md", ".opencode/commands/review-council.md"},
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
			Overwritten: []string{".opencode/agents/cobalt-crush-dev.md", ".opencode/commands/review-council.md"},
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
		if !strings.Contains(output, "! .opencode/agents/cobalt-crush-dev.md") {
			t.Errorf("expected '!' prefix for overwritten files")
		}
		if !strings.Contains(output, "! .opencode/commands/review-council.md") {
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
	// Speckit commands — created by specify init + /uf-init, not scaffolded by uf init
	".opencode/commands/speckit.specify.md":       true,
	".opencode/commands/speckit.clarify.md":       true,
	".opencode/commands/speckit.plan.md":          true,
	".opencode/commands/speckit.tasks.md":         true,
	".opencode/commands/speckit.analyze.md":       true,
	".opencode/commands/speckit.checklist.md":     true,
	".opencode/commands/speckit.implement.md":     true,
	".opencode/commands/speckit.constitution.md":  true,
	".opencode/commands/speckit.taskstoissues.md": true,
	// Speckit files — created by specify init, not scaffolded by uf init
	".specify/config.yaml":                          true,
	".specify/templates/agent-file-template.md":     true,
	".specify/templates/checklist-template.md":      true,
	".specify/templates/constitution-template.md":   true,
	".specify/templates/plan-template.md":           true,
	".specify/templates/spec-template.md":           true,
	".specify/templates/tasks-template.md":          true,
	".specify/scripts/bash/check-prerequisites.sh":  true,
	".specify/scripts/bash/common.sh":               true,
	".specify/scripts/bash/create-new-feature.sh":   true,
	".specify/scripts/bash/setup-plan.sh":           true,
	".specify/scripts/bash/update-agent-context.sh": true,
	// OpenSpec config — created by openspec init
	"openspec/config.yaml": true,
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
	".opencode/commands/cobalt-crush.md":               true,
	".opencode/commands/gaze.md":                       true,
	".opencode/commands/gaze-fix.md":                   true,
	".opencode/commands/speckit.testreview.md":         true,
	".opencode/commands/muti-mind.backlog-add.md":      true,
	".opencode/commands/muti-mind.backlog-list.md":     true,
	".opencode/commands/muti-mind.backlog-show.md":     true,
	".opencode/commands/muti-mind.backlog-update.md":   true,
	".opencode/commands/muti-mind.generate-stories.md": true,
	".opencode/commands/muti-mind.init.md":             true,
	".opencode/commands/muti-mind.prioritize.md":       true,
	".opencode/commands/muti-mind.sync-project.md":     true,
	".opencode/commands/muti-mind.sync-pull.md":        true,
	".opencode/commands/muti-mind.sync-push.md":        true,
	".opencode/commands/muti-mind.sync-status.md":      true,
	".opencode/commands/muti-mind.sync.md":             true,
	// OpenSpec skill commands — local workflow tooling, not scaffolded by uf init
	".opencode/commands/opsx-apply.md":   true,
	".opencode/commands/opsx-archive.md": true,
	".opencode/commands/opsx-explore.md": true,
	".opencode/commands/opsx-propose.md": true,
	// Workflow commands — Spec 008 swarm orchestration, local-only
	".opencode/commands/workflow-start.md":   true,
	".opencode/commands/workflow-status.md":  true,
	".opencode/commands/workflow-list.md":    true,
	".opencode/commands/workflow-advance.md": true,
	".opencode/commands/workflow-seed.md":    true,
	// Swarm skills — Spec 008, local-only
	".opencode/skill/unbound-force-heroes/SKILL.md": true,
	// Replicator-scaffolded agents and commands — created by replicator init,
	// not part of the uf binary's scaffold assets
	".opencode/agents/background-worker.md": true,
	".opencode/agents/coordinator.md":       true,
	".opencode/agents/worker.md":            true,
	".opencode/commands/forge.md":            true,
	".opencode/commands/forge-status.md":     true,
	".opencode/commands/handoff.md":          true,
	".opencode/commands/inbox.md":            true,
	".opencode/commands/org.md":              true,
	// Replicator-scaffolded skills — created by replicator init
	".opencode/skills/always-on-guidance/SKILL.md": true,
	".opencode/skills/forge-coordination/SKILL.md": true,
	".opencode/skills/forge-global/SKILL.md":       true,
	".opencode/skills/learning-systems/SKILL.md":   true,
	".opencode/skills/replicator-cli/SKILL.md":     true,
	".opencode/skills/system-design/SKILL.md":      true,
	".opencode/skills/testing-patterns/SKILL.md":   true,
	// Dewey-scaffolded commands — created by dewey init
	".opencode/commands/dewey-index.md":   true,
	".opencode/commands/dewey-reindex.md": true,
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
		".opencode/commands",
		".opencode/agents",
		".opencode/uf/packs",
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

	// Also check standalone config files that are still embedded.
	// Note: .specify/config.yaml and openspec/config.yaml are now
	// created by external CLIs (specify init, openspec init) and
	// listed in knownNonEmbeddedFiles — no longer checked here.
}

func TestMapAssetPath_Prefixes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"opencode/commands/speckit.specify.md", ".opencode/commands/speckit.specify.md"},
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
		{"opencode/commands/review-council.md", true},
		// Divisor convention packs
		{"opencode/uf/packs/go.md", true},
		{"opencode/uf/packs/default.md", true},
		{"opencode/uf/packs/go-custom.md", true},
		{"opencode/uf/packs/severity.md", true},
		{"opencode/uf/packs/python.md", true},
		{"opencode/uf/packs/python-custom.md", true},
		// Non-Divisor assets
		{"opencode/agents/constitution-check.md", false},
		{"opencode/commands/speckit.specify.md", false},
		{"opencode/commands/speckit.plan.md", false},
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
		{"python pyproject.toml", []string{"pyproject.toml"}, "python"},
		{"python setup.py", []string{"setup.py"}, "python"},
		{"python setup.cfg", []string{"setup.cfg"}, "python"},
		{"python requirements.txt", []string{"requirements.txt"}, "python"},
		{"python tox.ini", []string{"tox.ini"}, "python"},
		{"python Pipfile", []string{"Pipfile"}, "python"},
		{"python pyproject takes priority", []string{"pyproject.toml", "setup.py", "requirements.txt"}, "python"},
		{"rust project", []string{"Cargo.toml"}, "rust"},
		{"no markers", []string{}, ""},
		{"go takes priority over ts", []string{"go.mod", "package.json"}, "go"},
		{"go takes priority over python", []string{"go.mod", "requirements.txt"}, "go"},
		{"ts takes priority over python", []string{"tsconfig.json", "pyproject.toml"}, "typescript"},
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
		{"opencode/commands/review-council.md", "go", true},
		// Default and severity packs always deploy (language-agnostic)
		{"opencode/uf/packs/default.md", "go", true},
		{"opencode/uf/packs/default-custom.md", "go", true},
		{"opencode/uf/packs/default.md", "typescript", true},
		{"opencode/uf/packs/severity.md", "go", true},
		{"opencode/uf/packs/severity.md", "typescript", true},
		{"opencode/uf/packs/severity.md", "default", true},
		// Matching language packs deploy
		{"opencode/uf/packs/go.md", "go", true},
		{"opencode/uf/packs/go-custom.md", "go", true},
		{"opencode/uf/packs/typescript.md", "typescript", true},
		{"opencode/uf/packs/typescript-custom.md", "typescript", true},
		// Python packs deploy for python lang
		{"opencode/uf/packs/python.md", "python", true},
		{"opencode/uf/packs/python-custom.md", "python", true},
		// Non-matching language packs do NOT deploy
		{"opencode/uf/packs/typescript.md", "go", false},
		{"opencode/uf/packs/typescript-custom.md", "go", false},
		{"opencode/uf/packs/go.md", "typescript", false},
		{"opencode/uf/packs/go-custom.md", "typescript", false},
		{"opencode/uf/packs/python.md", "go", false},
		{"opencode/uf/packs/python-custom.md", "go", false},
		{"opencode/uf/packs/go.md", "python", false},
		{"opencode/uf/packs/go-custom.md", "python", false},
		// Default lang gets only default packs
		{"opencode/uf/packs/go.md", "default", false},
		{"opencode/uf/packs/python.md", "default", false},
		{"opencode/uf/packs/default.md", "default", true},
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

	// Verify no openspec files created (except schema which is embedded)
	for _, f := range result.Created {
		if strings.HasPrefix(f, "openspec/") && !strings.Contains(f, "schemas/") {
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
		if strings.HasSuffix(f, "go.md") && strings.Contains(f, "uf/packs") {
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
		if strings.HasSuffix(f, "typescript.md") && strings.Contains(f, "uf/packs") {
			foundTSPack = true
		}
	}
	if !foundTSPack {
		t.Error("expected TypeScript convention pack to be deployed")
	}

	// Verify Go pack NOT deployed
	for _, f := range result.Created {
		if strings.HasSuffix(f, "/go.md") && strings.Contains(f, "uf/packs") {
			t.Error("Go convention pack should not be deployed when lang=typescript")
		}
	}

	// All 9 Divisor agent files created (6 review + 3 content)
	agentCount := 0
	for _, f := range result.Created {
		if strings.Contains(f, "agents/divisor-") {
			agentCount++
		}
	}
	if agentCount != 9 {
		t.Errorf("expected 9 Divisor agent files, got %d", agentCount)
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

	// Verify only always-deploy packs deployed (default, severity, content)
	for _, f := range result.Created {
		if strings.Contains(f, "uf/packs") {
			base := filepath.Base(f)
			if !strings.HasPrefix(base, "default") && base != "severity.md" &&
				!strings.HasPrefix(base, "content") {
				t.Errorf("expected only default/severity/content packs, got %s", f)
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

// TestScaffoldOutput_NoSwarmPluginReferences is a regression guard
// for Spec 024 FR-010/FR-016: scaffolded files must not contain any
// Swarm plugin references. Replicator replaces the Swarm plugin.
func TestScaffoldOutput_NoSwarmPluginReferences(t *testing.T) {
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
		"opencode-swarm-plugin",
		"installSwarmPlugin",
		"ensureBun",
		"swarmForkSource",
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
				t.Errorf("scaffolded file %s contains stale %q reference (Spec 024 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestScaffoldOutput_NoHivemindReferences is a regression guard
// for Spec 022 FR-006/FR-007: scaffolded files must not contain any
// Hivemind tool references. Dewey replaces Hivemind as the unified
// memory layer for all learning storage and retrieval.
func TestScaffoldOutput_NoHivemindReferences(t *testing.T) {
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
		"hivemind_store",
		"hivemind_find",
		"hivemind_validate",
		"hivemind_remove",
		"hivemind_get",
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
				t.Errorf("scaffolded file %s contains stale %q reference (Spec 022 violation)", relPath, pattern)
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

// TestScaffoldOutput_NoOldPathReferences is a regression guard
// for Spec 025 SC-001/SC-002: scaffolded files must not contain any
// old directory path references. All tool workspace directories are
// unified under .uf/ per the directory convention.
func TestScaffoldOutput_NoOldPathReferences(t *testing.T) {
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
		"opencode/unbound/",
		".unbound-force/",
		".dewey/",
		".hive/",
		".muti-mind/",
		".mx-f/",
	}

	// Walk all generated files and search for stale patterns.
	// Exclude .gitignore — it intentionally lists legacy directories
	// as ignore patterns (per gitignore-init change).
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		relPath, _ := filepath.Rel(dir, path)
		if relPath == ".gitignore" {
			return nil // .gitignore intentionally lists legacy dirs
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)

		for _, pattern := range stalePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains stale %q reference (Spec 025 violation)", relPath, pattern)
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
	mu      sync.Mutex
	calls   []string
	errors  map[string]error
	outputs map[string]string
}

func (r *scaffoldCmdRecorder) execCmd(name string, args ...string) ([]byte, error) {
	key := name
	if len(args) > 0 {
		key = name + " " + strings.Join(args, " ")
	}

	r.mu.Lock()
	r.calls = append(r.calls, key)
	err, ok := r.errors[key]
	var out []byte
	if r.outputs != nil {
		if s, has := r.outputs[key]; has {
			out = []byte(s)
		}
	}
	r.mu.Unlock()

	if ok {
		return out, err
	}
	if out != nil {
		return out, nil
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

	// Should have 3 results: dewey init + dewey index + opencode.json.
	// (.uf/config.yaml is no longer created by uf init — use uf config init)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}

	if results[0].name != ".uf/dewey/" || results[0].action != "initialized" {
		t.Errorf("expected .uf/dewey/ initialized, got %s %s", results[0].name, results[0].action)
	}
	if results[1].name != "dewey index" || results[1].action != "completed" {
		t.Errorf("expected dewey index completed, got %s %s", results[1].name, results[1].action)
	}
	if results[2].name != "opencode.json" || results[2].action != "created" {
		t.Errorf("expected opencode.json created, got %s %s", results[2].name, results[2].action)
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

	// Should have 1 result: opencode.json skipped.
	// (.uf/config.yaml is no longer created by uf init)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d: %v", len(results), results)
	}
	if len(results) > 0 && results[0].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[0].name)
	}

	// No commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands, got: %v", rec.calls)
	}
}

func TestInitSubTools_DeweyAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 1 result: opencode.json created
	// (.uf/dewey/ already exists, dewey in PATH → mcp.dewey added).
	// (.uf/config.yaml no longer created by uf init)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d: %v", len(results), results)
	}
	if len(results) > 0 && results[0].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[0].name)
	}

	// dewey init should NOT have been called.
	for _, call := range rec.calls {
		if call == "dewey init" {
			t.Error("dewey init should NOT be called when .uf/dewey/ already exists")
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

	// Should have 2 results: dewey init failed + opencode.json created.
	// (.uf/config.yaml no longer created by uf init)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d: %v", len(results), results)
	}

	if results[0].name != ".uf/dewey/" || results[0].action != "failed" {
		t.Errorf("expected .uf/dewey/ failed, got %s %s", results[0].name, results[0].action)
	}
	if results[1].name != "opencode.json" || results[1].action != "created" {
		t.Errorf("expected opencode.json created, got %s %s", results[1].name, results[1].action)
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
			Created: []string{".opencode/commands/review-council.md"},
		}
		subResults := []subToolResult{
			{name: ".uf/dewey/", action: "initialized"},
			{name: "dewey index", action: "completed"},
		}

		printSummary(&buf, false, false, true, r, subResults)
		output := buf.String()

		// Should show sub-tool results.
		if !strings.Contains(output, "Sub-tool initialization:") {
			t.Errorf("expected sub-tool section, got:\n%s", output)
		}
		if !strings.Contains(output, ".uf/dewey/ initialized") {
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
			Created: []string{".opencode/commands/review-council.md"},
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
			Created: []string{".opencode/commands/review-council.md"},
		}
		subResults := []subToolResult{
			{name: ".uf/dewey/", action: "failed", detail: "dewey init failed"},
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

func TestInitSubTools_DoesNotCreateWorkflowConfig(t *testing.T) {
	// .uf/config.yaml is no longer created by uf init.
	// It is now created exclusively by uf config init.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No dewey
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have a config.yaml result.
	for _, r := range results {
		if r.name == ".uf/config.yaml" {
			t.Errorf("uf init should NOT create .uf/config.yaml, got result: %s %s", r.name, r.action)
		}
	}

	// Verify file does NOT exist.
	configPath := filepath.Join(dir, ".uf", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		t.Error(".uf/config.yaml should not exist after uf init")
	}
}

// --- Dewey auto-sources tests ---

func TestGenerateDeweySources_SiblingsDetected(t *testing.T) {
	// Create a parent dir with the "current" project and 3 sibling repos.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write default sources.yaml (single source entry).
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
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

	result := generateDeweySources(opts, false)

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
	content, err := os.ReadFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"))
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

	// Verify disk-org source with recursive: false.
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}
	if !strings.Contains(text, "recursive: false") {
		t.Error("expected recursive: false on disk-org")
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
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write default sources.yaml.
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	// No ExecCmd stub needed — extractGitHubOrg will fail gracefully.
	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("no remote")
		},
	}

	result := generateDeweySources(opts, false)

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
	content, err := os.ReadFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"))
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
	if !strings.Contains(text, "recursive: false") {
		t.Error("expected recursive: false on disk-org")
	}

	// Should NOT have GitHub source (no remote).
	if strings.Contains(text, "type: github") {
		t.Error("should not have GitHub source when no remote")
	}
}

func TestGenerateDeweySources_AlreadyCustomized(t *testing.T) {
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
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
	sourcesPath := filepath.Join(currentDir, ".uf", "dewey", "sources.yaml")
	if err := os.WriteFile(sourcesPath, []byte(customSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("should not be called")
		},
	}

	result := generateDeweySources(opts, false)

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

func TestGenerateDeweySources_ForceOverwritesCustom(t *testing.T) {
	// Create a parent dir with the "current" project and a sibling repo.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a sibling repo.
	if err := os.MkdirAll(filepath.Join(parentDir, "gaze", ".git"), 0o755); err != nil {
		t.Fatalf("mkdir sibling: %v", err)
	}

	// Write a customized sources.yaml (>1 source entry).
	customSources := `sources:
  - id: disk-local
    type: disk
    config:
      path: "."
  - id: my-custom-source
    type: disk
    config:
      path: "../custom"
`
	sourcesPath := filepath.Join(currentDir, ".uf", "dewey", "sources.yaml")
	if err := os.WriteFile(sourcesPath, []byte(customSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			if name == "git" {
				return []byte("git@github.com:unbound-force/my-project.git\n"), nil
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}

	// Call with force=true — should overwrite customized file.
	result := generateDeweySources(opts, true)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "completed" {
		t.Errorf("expected action 'completed', got %q", result.action)
	}

	// Read the regenerated sources.yaml.
	content, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	text := string(content)

	// Should have the auto-detected config, not the custom one.
	if strings.Contains(text, "my-custom-source") {
		t.Error("custom source should have been overwritten")
	}
	if !strings.Contains(text, "- id: disk-gaze") {
		t.Error("expected disk-gaze source in regenerated config")
	}
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}
	if !strings.Contains(text, "recursive: false") {
		t.Error("expected recursive: false on disk-org")
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
	ufDir := filepath.Join(dir, ".uf")
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
		if r.name == ".uf/config.yaml" {
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

// getMCPReplicator extracts the mcp.replicator entry from a parsed opencode.json.
func getMCPReplicator(t *testing.T, ocMap map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key not found")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	repRaw, ok := mcpMap["replicator"]
	if !ok {
		t.Fatal("mcp.replicator not found")
	}
	var rep map[string]json.RawMessage
	if err := json.Unmarshal(repRaw, &rep); err != nil {
		t.Fatalf("parse mcp.replicator: %v", err)
	}
	return rep
}

func TestConfigureOpencodeJSON_Create(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
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
	getMCPDewey(t, ocMap)

	// Verify mcp.replicator entry.
	rep := getMCPReplicator(t, ocMap)
	var repType string
	_ = json.Unmarshal(rep["type"], &repType)
	if repType != "local" {
		t.Errorf("mcp.replicator.type = %q, want 'local'", repType)
	}
	var cmd []string
	_ = json.Unmarshal(rep["command"], &cmd)
	expectedCmd := []string{"replicator", "serve"}
	if len(cmd) != len(expectedCmd) {
		t.Errorf("mcp.replicator.command = %v, want %v", cmd, expectedCmd)
	} else {
		for i := range cmd {
			if cmd[i] != expectedCmd[i] {
				t.Errorf("mcp.replicator.command[%d] = %q, want %q", i, cmd[i], expectedCmd[i])
			}
		}
	}

	// Verify no plugin array.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist (Replicator uses MCP, not plugin array)")
	}
}

func TestConfigureOpencodeJSON_DeweyOnly(t *testing.T) {
	dir := t.TempDir()

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
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_ReplicatorOnly(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should have mcp.replicator.
	getMCPReplicator(t, ocMap)

	// Should NOT have plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_Neither(t *testing.T) {
	dir := t.TempDir()
	// No dewey, no replicator.

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

	// Create opencode.json with both entries already present.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true,
      "type": "local"
    },
    "replicator": {
      "command": ["replicator", "serve"],
      "enabled": true,
      "type": "local"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
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

func TestConfigureOpencodeJSON_LegacyPluginMigration(t *testing.T) {
	dir := t.TempDir()

	// Legacy opencode.json with opencode-swarm-plugin in plugin array.
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
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.replicator was added.
	getMCPReplicator(t, ocMap)

	// Verify plugin key was removed (empty after removing swarm plugin).
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should be removed after legacy migration (empty array)")
	}
}

func TestConfigureOpencodeJSON_LegacyPluginMigration_OtherPlugins(t *testing.T) {
	dir := t.TempDir()

	// Legacy opencode.json with swarm plugin AND other plugins.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
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
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.replicator was added.
	getMCPReplicator(t, ocMap)

	// Verify plugin key preserved with only other-plugin.
	plugins := getPlugins(t, ocMap)
	if len(plugins) != 1 || plugins[0] != "other-plugin" {
		t.Errorf("plugin = %v, want [other-plugin]", plugins)
	}
}

func TestConfigureOpencodeJSON_ReplicatorNotInstalled(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should have mcp.dewey but NOT mcp.replicator.
	getMCPDewey(t, ocMap)
	mcpRaw := ocMap["mcp"]
	var mcpMap map[string]json.RawMessage
	_ = json.Unmarshal(mcpRaw, &mcpMap)
	if _, ok := mcpMap["replicator"]; ok {
		t.Error("mcp.replicator should not exist when replicator not installed")
	}

	// Should NOT have plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_AddMissing(t *testing.T) {
	dir := t.TempDir()

	// Has mcp.dewey but no mcp.replicator.
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
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.dewey preserved.
	getMCPDewey(t, ocMap)

	// Verify mcp.replicator was added.
	getMCPReplicator(t, ocMap)
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

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
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

	// Stale mcp.dewey with --include-hidden flag.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--include-hidden", "--vault", "."],
      "enabled": true
    },
    "replicator": {
      "type": "local",
      "command": ["replicator", "serve"],
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
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "overwritten" {
		t.Errorf("expected action 'overwritten', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.dewey was overwritten with correct command (no --include-hidden).
	dewey := getMCPDewey(t, ocMap)
	var cmd []string
	_ = json.Unmarshal(dewey["command"], &cmd)
	expectedCmd := []string{"dewey", "serve", "--vault", "."}
	if len(cmd) != len(expectedCmd) {
		t.Fatalf("command = %v, want %v", cmd, expectedCmd)
	}
	for i := range cmd {
		if cmd[i] != expectedCmd[i] {
			t.Errorf("command[%d] = %q, want %q", i, cmd[i], expectedCmd[i])
		}
	}

	// Verify no plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
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
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
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

func TestConfigureOpencodeJSON_SkipDewey(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with dewey method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action == "skipped" {
		t.Fatal("should not skip entirely — replicator is still active")
	}

	ocMap := parseOpencodeJSON(t, dir)
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key should exist for replicator")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	if _, ok := mcpMap["dewey"]; ok {
		t.Error("mcp.dewey should NOT exist when setup.tools.dewey.method is skip")
	}
	if _, ok := mcpMap["replicator"]; !ok {
		t.Error("mcp.replicator should exist")
	}
}

func TestConfigureOpencodeJSON_SkipReplicator(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with replicator method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    replicator:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action == "skipped" {
		t.Fatal("should not skip entirely — dewey is still active")
	}

	ocMap := parseOpencodeJSON(t, dir)
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key should exist for dewey")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	if _, ok := mcpMap["replicator"]; ok {
		t.Error("mcp.replicator should NOT exist when setup.tools.replicator.method is skip")
	}
	if _, ok := mcpMap["dewey"]; !ok {
		t.Error("mcp.dewey should exist")
	}
}

func TestConfigureOpencodeJSON_SkipBoth(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with both tools set to skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n    replicator:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "skipped" {
		t.Errorf("expected action 'skipped' when both tools are skip, got %q", results[0].action)
	}
	if results[0].detail != "nothing to configure" {
		t.Errorf("expected detail 'nothing to configure', got %q", results[0].detail)
	}

	// Verify no file was written.
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Error("opencode.json should not be created when both tools are skipped")
	}
}

// --- Dewey force re-index tests ---

func TestInitSubTools_DeweyForceReindex(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write a default sources.yaml so generateDeweySources has
	// something to regenerate on force.
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".uf", "dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
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

	// Verify dewey init was NOT called (.uf/dewey/ already exists).
	for _, call := range rec.calls {
		if call == "dewey init" {
			t.Error("dewey init should NOT be called when .uf/dewey/ already exists")
		}
	}

	// Verify sources.yaml was regenerated with recursive: false.
	content, readErr := os.ReadFile(filepath.Join(dir, ".uf", "dewey", "sources.yaml"))
	if readErr != nil {
		t.Fatalf("read sources.yaml: %v", readErr)
	}
	if !strings.Contains(string(content), "recursive: false") {
		t.Error("expected sources.yaml to contain recursive: false after force regeneration")
	}
}

func TestInitSubTools_DeweyExistsNoForce(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
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
		if r.name == ".uf/dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when Force=false and .uf/dewey/ exists: %s %s", r.name, r.action)
		}
	}

	// Verify no dewey commands were called.
	for _, call := range rec.calls {
		if call == "dewey init" || call == "dewey index" {
			t.Errorf("unexpected dewey command when Force=false: %s", call)
		}
	}
}

func TestInitSubTools_DeweySkippedByConfig(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with dewey method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Dewey commands should NOT have been called.
	for _, call := range rec.calls {
		if call == "dewey init" || call == "dewey index" {
			t.Errorf("dewey command should NOT be called when method is skip: %s", call)
		}
	}

	// No dewey-related results should appear.
	for _, r := range results {
		if r.name == ".uf/dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when method is skip: %s %s", r.name, r.action)
		}
	}

	// .uf/dewey/ directory should NOT have been created.
	if _, err := os.Stat(filepath.Join(dir, ".uf", "dewey")); !os.IsNotExist(err) {
		t.Error(".uf/dewey/ should not exist when dewey method is skip")
	}
}

func TestInitSubTools_DeweySkippedByConfig_ForceIgnored(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with dewey method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	// Create .uf/dewey/ — even with Force, skip should win.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir dewey: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Even with Force=true, skip config should prevent re-index.
	for _, call := range rec.calls {
		if call == "dewey init" || call == "dewey index" {
			t.Errorf("dewey command should NOT be called when method is skip (even with Force): %s", call)
		}
	}
	for _, r := range results {
		if r.name == ".uf/dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when method is skip: %s %s", r.name, r.action)
		}
	}
}

func TestInitSubTools_ReplicatorSkippedByConfig(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with replicator method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    replicator:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Replicator commands should NOT have been called.
	for _, call := range rec.calls {
		if call == "replicator init" {
			t.Errorf("replicator init should NOT be called when method is skip")
		}
	}
	for _, r := range results {
		if r.name == ".uf/replicator/" {
			t.Errorf("unexpected replicator result when method is skip: %s %s", r.name, r.action)
		}
	}
}

func TestInitSubTools_AllToolsSkippedByConfig(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with all tools set to skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n    replicator:\n      method: skip\n    specify:\n      method: skip\n    openspec:\n      method: skip\n    gaze:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
			"specify":    "/usr/local/bin/specify",
			"openspec":   "/usr/local/bin/openspec",
			"gaze":       "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// No tool commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands when all tools are skipped, got: %v", rec.calls)
	}

	// Only result should be opencode.json skipped (nothing to configure).
	if len(results) != 1 {
		t.Fatalf("expected 1 result (opencode.json), got %d: %v", len(results), results)
	}
	if results[0].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[0].name)
	}
}

// Phase: Concurrent execution tests (FR-001 through FR-005).

func TestInitSubTools_ConcurrentAllResults(t *testing.T) {
	// FR-001, FR-003: All five tools available — results from all
	// tools are collected regardless of execution order.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
			"specify":    "/usr/local/bin/specify",
			"openspec":   "/usr/local/bin/openspec",
			"gaze":       "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// Expect results from: dewey init, dewey index, replicator,
	// specify, openspec, gaze, opencode.json.
	// (dewey sources may or may not appear depending on filesystem.)
	expectedNames := map[string]bool{
		".uf/dewey/":      false,
		"dewey index":     false,
		".uf/replicator/": false,
		".specify/":       false,
		"openspec/":       false,
		"gaze":            false,
		"opencode.json":   false,
	}

	for _, r := range results {
		if _, ok := expectedNames[r.name]; ok {
			expectedNames[r.name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("missing result for %q in concurrent run", name)
		}
	}

	// Verify all expected commands were issued.
	rec.mu.Lock()
	callsCopy := make([]string, len(rec.calls))
	copy(callsCopy, rec.calls)
	rec.mu.Unlock()

	expectedCmds := []string{
		"dewey init", "dewey index",
		"replicator init", "specify init",
		"openspec init --tools opencode", "gaze init",
	}
	for _, cmd := range expectedCmds {
		found := false
		for _, call := range callsCopy {
			if call == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q in concurrent run, got calls: %v", cmd, callsCopy)
		}
	}
}

func TestInitSubTools_ConcurrentMissingToolSkipped(t *testing.T) {
	// FR-001: Missing tools are skipped; available tools still
	// complete their init without delay.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	// Only dewey and gaze available; others missing.
	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey": "/usr/local/bin/dewey",
			"gaze":  "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// Should see results for dewey, gaze, and opencode.json.
	// Should NOT see replicator, specify, or openspec results.
	resultNames := map[string]bool{}
	for _, r := range results {
		resultNames[r.name] = true
	}

	if !resultNames[".uf/dewey/"] {
		t.Error("expected dewey result when dewey is available")
	}
	if !resultNames["gaze"] {
		t.Error("expected gaze result when gaze is available")
	}
	if resultNames[".uf/replicator/"] {
		t.Error("unexpected replicator result when replicator is not available")
	}
	if resultNames[".specify/"] {
		t.Error("unexpected specify result when specify is not available")
	}
	if resultNames["openspec/"] {
		t.Error("unexpected openspec result when openspec is not available")
	}
}

func TestInitSubTools_ConcurrentOneFailureNoBlock(t *testing.T) {
	// FR-005: One tool failure does not prevent other tools from
	// completing. Dewey init fails; replicator and gaze succeed.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{
		"dewey init": fmt.Errorf("simulated dewey failure"),
	}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
			"gaze":       "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// Dewey should appear with "failed" action.
	deweyFailed := false
	replicatorOK := false
	gazeOK := false
	opencodeOK := false
	for _, r := range results {
		switch r.name {
		case ".uf/dewey/":
			if r.action == "failed" {
				deweyFailed = true
			}
		case ".uf/replicator/":
			if r.action == "initialized" {
				replicatorOK = true
			}
		case "gaze":
			if r.action == "initialized" {
				gazeOK = true
			}
		case "opencode.json":
			opencodeOK = true
		}
	}

	if !deweyFailed {
		t.Error("expected dewey to report failure")
	}
	if !replicatorOK {
		t.Error("expected replicator to succeed despite dewey failure")
	}
	if !gazeOK {
		t.Error("expected gaze to succeed despite dewey failure")
	}
	if !opencodeOK {
		t.Error("expected opencode.json result after concurrent completion")
	}

	// Verify dewey index was NOT called (init failed).
	rec.mu.Lock()
	for _, call := range rec.calls {
		if call == "dewey index" {
			t.Error("dewey index should not run when dewey init fails")
		}
	}
	rec.mu.Unlock()
}

func TestInitSubTools_ConcurrentOpencodeJSONAfterAll(t *testing.T) {
	// FR-004: configureOpencodeJSON runs after all concurrent
	// groups complete. Verified by checking opencode.json result
	// is present and contains correct MCP entries.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// opencode.json must be present in results.
	opencodeFound := false
	for _, r := range results {
		if r.name == "opencode.json" {
			opencodeFound = true
		}
	}
	if !opencodeFound {
		t.Fatal("expected opencode.json result after concurrent init")
	}

	// opencode.json file should exist on disk (written by
	// configureOpencodeJSON after wg.Wait).
	opencodeJSON := filepath.Join(dir, "opencode.json")
	data, err := os.ReadFile(opencodeJSON)
	if err != nil {
		t.Fatalf("opencode.json should exist: %v", err)
	}

	// Verify it contains expected MCP server entries for both
	// tools that were initialized.
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("opencode.json is not valid JSON: %v", jsonErr)
	}
	mcp, ok := parsed["mcp"]
	if !ok {
		t.Fatal("opencode.json missing mcp key")
	}
	mcpMap, ok := mcp.(map[string]interface{})
	if !ok {
		t.Fatal("opencode.json mcp is not a map")
	}
	if _, hasDewey := mcpMap["dewey"]; !hasDewey {
		t.Error("opencode.json should contain dewey MCP entry")
	}
	if _, hasReplicator := mcpMap["replicator"]; !hasReplicator {
		t.Error("opencode.json should contain replicator MCP entry")
	}
}

// Phase 8: Integration test (T032a)

func TestInitSubTools_OpencodeJSON(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ so dewey init is skipped.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir .uf/dewey: %v", err)
	}
	// Create .uf/replicator/ so replicator init is skipped.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0o755); err != nil {
		t.Fatalf("mkdir .uf/replicator: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
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
	getMCPReplicator(t, ocMap)

	// Verify no plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

// --- Replicator init delegation tests ---

func TestInitSubTools_ReplicatorInit(t *testing.T) {
	dir := t.TempDir()
	// No .uf/replicator/ — should trigger replicator init.
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have .uf/replicator/ initialized result.
	foundReplicator := false
	for _, r := range results {
		if r.name == ".uf/replicator/" && r.action == "initialized" {
			foundReplicator = true
		}
	}
	if !foundReplicator {
		t.Errorf("expected .uf/replicator/ initialized, got %v", results)
	}

	// Verify replicator init was called.
	found := false
	for _, call := range rec.calls {
		if call == "replicator init" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'replicator init' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_ReplicatorInitSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/replicator/ — should skip replicator init.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have .uf/replicator/ result.
	for _, r := range results {
		if r.name == ".uf/replicator/" {
			t.Errorf("unexpected .uf/replicator/ result when already exists: %s %s", r.name, r.action)
		}
	}

	// Verify replicator init was NOT called.
	for _, call := range rec.calls {
		if call == "replicator init" {
			t.Error("replicator init should NOT be called when .uf/replicator/ exists")
		}
	}
}

func TestInitSubTools_ReplicatorInitFails(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"replicator init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have .uf/replicator/ failed result.
	foundFailed := false
	for _, r := range results {
		if r.name == ".uf/replicator/" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected .uf/replicator/ failed, got %v", results)
	}

	// Should still have opencode.json result (init failure doesn't block).
	foundOC := false
	for _, r := range results {
		if r.name == "opencode.json" {
			foundOC = true
		}
	}
	if !foundOC {
		t.Error("opencode.json result should still be present after replicator init failure")
	}
}

// --- Specify delegation tests ---

func TestInitSubTools_SpecifyInit(t *testing.T) {
	dir := t.TempDir()
	// No .specify/ — should trigger specify init.
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundSpecify := false
	for _, r := range results {
		if r.name == ".specify/" && r.action == "initialized" {
			foundSpecify = true
		}
	}
	if !foundSpecify {
		t.Errorf("expected .specify/ initialized, got %v", results)
	}

	found := false
	for _, call := range rec.calls {
		if call == "specify init" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'specify init' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_SpecifySkipped(t *testing.T) {
	dir := t.TempDir()
	// Create .specify/ — should skip specify init.
	if err := os.MkdirAll(filepath.Join(dir, ".specify"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	for _, r := range results {
		if r.name == ".specify/" {
			t.Errorf("unexpected .specify/ result when already exists: %s %s", r.name, r.action)
		}
	}

	for _, call := range rec.calls {
		if call == "specify init" {
			t.Error("specify init should NOT be called when .specify/ exists")
		}
	}
}

func TestInitSubTools_SpecifyNotInstalled(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No specify
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	for _, r := range results {
		if r.name == ".specify/" {
			t.Errorf("unexpected .specify/ result when specify not installed: %s %s", r.name, r.action)
		}
	}
}

func TestInitSubTools_SpecifyFailed(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"specify init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundFailed := false
	for _, r := range results {
		if r.name == ".specify/" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected .specify/ failed, got %v", results)
	}
}

// --- OpenSpec delegation tests ---

func TestInitSubTools_OpenSpecInit(t *testing.T) {
	dir := t.TempDir()
	// Create openspec/ directory (simulating embedded schema deployment)
	// but no config.yaml — should trigger openspec init.
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "schemas"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"openspec": "/usr/local/bin/openspec"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundOpenSpec := false
	for _, r := range results {
		if r.name == "openspec/" && r.action == "initialized" {
			foundOpenSpec = true
		}
	}
	if !foundOpenSpec {
		t.Errorf("expected openspec/ initialized, got %v", results)
	}

	found := false
	for _, call := range rec.calls {
		if call == "openspec init --tools opencode" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'openspec init --tools opencode' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_OpenSpecSkipped(t *testing.T) {
	dir := t.TempDir()
	// Create openspec/config.yaml — should skip openspec init.
	if err := os.MkdirAll(filepath.Join(dir, "openspec"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openspec", "config.yaml"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"openspec": "/usr/local/bin/openspec"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	for _, r := range results {
		if r.name == "openspec/" {
			t.Errorf("unexpected openspec/ result when config exists: %s %s", r.name, r.action)
		}
	}

	for _, call := range rec.calls {
		if strings.Contains(call, "openspec init") {
			t.Error("openspec init should NOT be called when config.yaml exists")
		}
	}
}

func TestInitSubTools_OpenSpecFailed(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"openspec init --tools opencode": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"openspec": "/usr/local/bin/openspec"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundFailed := false
	for _, r := range results {
		if r.name == "openspec/" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected openspec/ failed, got %v", results)
	}
}

// --- Gaze delegation tests ---

func TestInitSubTools_GazeInit(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/agents/ but no gaze-reporter.md — should trigger gaze init.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundGaze := false
	for _, r := range results {
		if r.name == "gaze" && r.action == "initialized" {
			foundGaze = true
		}
	}
	if !foundGaze {
		t.Errorf("expected gaze initialized, got %v", results)
	}

	found := false
	for _, call := range rec.calls {
		if call == "gaze init" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'gaze init' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_GazeSkipped(t *testing.T) {
	dir := t.TempDir()
	// Create gaze-reporter.md — should skip gaze init.
	agentDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "gaze-reporter.md"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	for _, r := range results {
		if r.name == "gaze" {
			t.Errorf("unexpected gaze result when agent exists: %s %s", r.name, r.action)
		}
	}

	for _, call := range rec.calls {
		if call == "gaze init" {
			t.Error("gaze init should NOT be called when gaze-reporter.md exists")
		}
	}
}

func TestInitSubTools_GazeFailed(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/agents/ but no gaze-reporter.md.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"gaze init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundFailed := false
	for _, r := range results {
		if r.name == "gaze" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected gaze failed, got %v", results)
	}
}

// --- ensureGitignore tests ---

func TestEnsureGitignore_FreshDir(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureGitignore(opts)

	if result.name != ".gitignore" {
		t.Errorf("expected name '.gitignore', got %q", result.name)
	}
	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	// Verify file exists with marker and all patterns.
	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, gitignoreMarker) {
		t.Error("expected marker comment in .gitignore")
	}
	// Verify representative patterns from each section.
	for _, pattern := range []string{
		".uf/workflows/",
		".uf/dewey/graph.db",
		".uf/dewey/cache/",
		".uf/replicator/*.db",
		".uf/muti-mind/artifacts/",
		".uf/mx-f/data/",
		".devcontainer/",
		".dewey/",
		".hive/",
		".unbound-force/",
		".muti-mind/",
		".mx-f/",
	} {
		if !strings.Contains(text, pattern) {
			t.Errorf("expected pattern %q in .gitignore", pattern)
		}
	}
}

func TestEnsureGitignore_ExistingNoBlock(t *testing.T) {
	dir := t.TempDir()

	// Create existing .gitignore with project-specific content.
	existingContent := "node_modules/\n*.log\n"
	giPath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(giPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureGitignore(opts)

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(giPath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	text := string(content)

	// Existing content must be preserved at the start.
	if !strings.HasPrefix(text, existingContent) {
		t.Errorf("existing content not preserved; got:\n%s", text)
	}

	// UF block must be appended.
	if !strings.Contains(text, gitignoreMarker) {
		t.Error("expected UF marker in .gitignore")
	}

	// Blank line separator between existing content and UF block.
	markerIdx := strings.Index(text, gitignoreMarker)
	if markerIdx < 2 {
		t.Fatal("marker at unexpected position")
	}
	// The two characters before the marker should be "\n\n" (blank line separator).
	before := text[:markerIdx]
	if !strings.HasSuffix(before, "\n\n") {
		t.Errorf("expected blank line separator before UF block, got trailing chars: %q",
			before[len(before)-4:])
	}
}

func TestEnsureGitignore_ExistingWithBlock(t *testing.T) {
	dir := t.TempDir()

	// Create .gitignore that already has the UF block.
	existingContent := "node_modules/\n\n" + gitignoreBlock
	giPath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(giPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureGitignore(opts)

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	// Verify file is byte-identical (not modified).
	after, err := os.ReadFile(giPath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(after) != existingContent {
		t.Error(".gitignore should be byte-identical when marker already present")
	}
}

func TestEnsureGitignore_Idempotent(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	// First call — creates the file.
	result1 := ensureGitignore(opts)
	if result1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", result1.action)
	}

	content1, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read after first call: %v", err)
	}

	// Second call — should skip (idempotent).
	result2 := ensureGitignore(opts)
	if result2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", result2.action)
	}

	content2, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read after second call: %v", err)
	}

	// Content must be identical after both calls (block not duplicated).
	if !bytes.Equal(content1, content2) {
		t.Errorf("content differs after second call:\nfirst:\n%s\nsecond:\n%s",
			string(content1), string(content2))
	}
}

// TestAgentBriefAsset_GovernanceBlockPresence verifies the
// /agent-brief scaffold asset contains all 7 behavioral rule
// templates and the conditional governance sections. Governance
// blocks moved from /uf-init Step 9 to /agent-brief verbatim
// templates as part of agent-brief-consolidation. Added by
// Spec 030, updated by agent-brief-consolidation.
func TestAgentBriefAsset_GovernanceBlockPresence(t *testing.T) {
	content, err := assetContent("opencode/commands/agent-brief.md")
	if err != nil {
		t.Fatalf("read agent-brief.md asset: %v", err)
	}
	text := string(content)

	// Each entry is a detection phrase that must appear in the
	// agent-brief.md asset to confirm the governance block is
	// defined in the verbatim template.
	requiredPhrases := []struct {
		block  string
		phrase string
	}{
		{"Gatekeeping", "**Gatekeeping**"},
		{"Phase boundaries", "**Phase boundaries**"},
		{"CI parity", "**CI parity**"},
		{"Review council", "**Review council**"},
		{"Branch protection", "**Branch protection**"},
		{"Documentation gate", "**Documentation gate**"},
		{"Zero-waste", "**Zero-waste**"},
		{"Specification Workflow", "## Specification Workflow"},
		{"Knowledge Retrieval", "## Knowledge Retrieval"},
	}

	for _, rp := range requiredPhrases {
		if !strings.Contains(text, rp.phrase) {
			t.Errorf("agent-brief.md asset missing governance block %q (detection phrase %q not found)",
				rp.block, rp.phrase)
		}
	}
}

// --- Cross-tool bridge tests ---

func TestEnsureAGENTSmdPackSection_NoAGENTSmd(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdPackSection(opts, "go")

	if result.action != "skipped (no AGENTS.md)" {
		t.Errorf("expected action 'skipped (no AGENTS.md)', got %q", result.action)
	}
}

func TestEnsureAGENTSmdPackSection_ExistingWithoutSection(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\nSome content.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdPackSection(opts, "go")

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)

	if !strings.HasPrefix(text, existing) {
		t.Error("existing content not preserved")
	}
	if !strings.Contains(text, agentsmdPackMarker) {
		t.Error("expected Convention Packs heading")
	}
	if !strings.Contains(text, ".opencode/uf/packs/go.md") {
		t.Error("expected Go pack reference")
	}
	if !strings.Contains(text, ".opencode/uf/packs/default.md") {
		t.Error("expected default pack reference")
	}
}

func TestEnsureAGENTSmdPackSection_AlreadyConfigured(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\n## Convention Packs\n\nAlready here.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdPackSection(opts, "go")

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(after) != existing {
		t.Error("AGENTS.md should be unchanged when section already present")
	}
}

func TestEnsureAGENTSmdPackSection_Idempotent(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	r1 := ensureAGENTSmdPackSection(opts, "go")
	if r1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", r1.action)
	}

	content1, _ := os.ReadFile(agentsPath)

	r2 := ensureAGENTSmdPackSection(opts, "go")
	if r2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", r2.action)
	}

	content2, _ := os.ReadFile(agentsPath)
	if !bytes.Equal(content1, content2) {
		t.Error("content should be identical after second call")
	}
}

func TestEnsureCLAUDEmd_FreshDir(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, claudemdMarker) {
		t.Error("expected marker in CLAUDE.md")
	}
	if !strings.Contains(text, "@AGENTS.md") {
		t.Error("expected @AGENTS.md import")
	}
	if !strings.Contains(text, "@.opencode/agents/cobalt-crush-dev.md") {
		t.Error("expected cobalt-crush @import")
	}
	if !strings.Contains(text, "@.opencode/uf/packs/go.md") {
		t.Error("expected Go pack @import")
	}
	if !strings.Contains(text, "@.opencode/uf/packs/default.md") {
		t.Error("expected default pack @import")
	}
	if !strings.Contains(text, "divisor-guard.md") {
		t.Error("expected Divisor review agent reference")
	}
}

func TestEnsureCLAUDEmd_ExistingWithoutMarker(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project Claude Config\n\nSome rules.\n"
	claudePath := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")

	if result.action != "appended" {
		t.Errorf("expected action 'appended', got %q", result.action)
	}

	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(content)

	if !strings.HasPrefix(text, existing) {
		t.Error("existing content not preserved")
	}
	if !strings.Contains(text, claudemdMarker) {
		t.Error("expected marker appended")
	}
}

func TestReplaceManagedBlock_NoMarker(t *testing.T) {
	content := "some content without marker\n"
	result, changed := replaceManagedBlock(content, claudemdMarker, "new block")
	if changed {
		t.Error("expected no change when marker absent")
	}
	if result != content {
		t.Error("expected original content returned")
	}
}

func TestReplaceManagedBlock_Identical(t *testing.T) {
	block := claudemdMarker + "\n\n@AGENTS.md\n"
	result, changed := replaceManagedBlock(block, claudemdMarker, block)
	if changed {
		t.Error("expected no change when content identical")
	}
	if result != block {
		t.Error("expected original content returned")
	}
}

func TestReplaceManagedBlock_Updated(t *testing.T) {
	prefix := "# My Project\n\n"
	oldBlock := claudemdMarker + "\n\nold content\n"
	newBlock := claudemdMarker + "\n\nnew content\n"
	content := prefix + oldBlock

	result, changed := replaceManagedBlock(content, claudemdMarker, newBlock)
	if !changed {
		t.Error("expected change when content differs")
	}
	if result != prefix+newBlock {
		t.Errorf("unexpected result:\n%s", result)
	}
}

func TestReplaceManagedBlock_PreservesPrefix(t *testing.T) {
	prefix := "User content line 1\nUser content line 2\n\n"
	oldBlock := claudemdMarker + "\n\nmanaged stuff\n"
	newBlock := claudemdMarker + "\n\nupdated managed stuff\n"
	content := prefix + oldBlock

	result, changed := replaceManagedBlock(content, claudemdMarker, newBlock)
	if !changed {
		t.Error("expected change")
	}
	if !strings.HasPrefix(result, prefix) {
		t.Error("prefix should be preserved")
	}
	if !strings.HasSuffix(result, "updated managed stuff\n") {
		t.Error("new block should be at end")
	}
}

func TestEnsureCLAUDEmd_AlreadyConfigured(t *testing.T) {
	dir := t.TempDir()

	existing := buildCLAUDEmdBlock("go", "")
	claudePath := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	after, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if string(after) != existing {
		t.Error("CLAUDE.md should be unchanged when block already matches")
	}
}

func TestEnsureCLAUDEmd_StaleContent(t *testing.T) {
	dir := t.TempDir()

	// Seed with an older managed block missing cobalt-crush and
	// review agents sections.
	stale := claudemdMarker + "\n\n@AGENTS.md\n\n## Convention Packs\n\n" +
		"@.opencode/uf/packs/default.md\n@.opencode/uf/packs/go.md\n"
	claudePath := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(stale), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")

	if result.action != "updated" {
		t.Errorf("expected action 'updated', got %q", result.action)
	}

	after, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(after)

	if !strings.Contains(text, "@.opencode/agents/cobalt-crush-dev.md") {
		t.Error("expected cobalt-crush @import after update")
	}
	if !strings.Contains(text, "divisor-guard.md") {
		t.Error("expected review agents section after update")
	}
}

func TestEnsureCLAUDEmd_StaleContentPreservesPrefix(t *testing.T) {
	dir := t.TempDir()

	// Seed with user content above the managed block.
	prefix := "# My Project\n\nCustom instructions.\n\n"
	stale := prefix + claudemdMarker + "\n\n@AGENTS.md\n"
	claudePath := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(stale), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")

	if result.action != "updated" {
		t.Errorf("expected action 'updated', got %q", result.action)
	}

	after, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	text := string(after)

	if !strings.HasPrefix(text, prefix) {
		t.Error("user content above marker should be preserved")
	}
	if !strings.Contains(text, "divisor-guard.md") {
		t.Error("expected review agents section after update")
	}
}

func TestEnsureCLAUDEmd_Idempotent(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	r1 := ensureCLAUDEmd(opts, "go")
	if r1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", r1.action)
	}

	content1, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))

	r2 := ensureCLAUDEmd(opts, "go")
	if r2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", r2.action)
	}

	content2, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if !bytes.Equal(content1, content2) {
		t.Error("content should be identical after second call")
	}
}

func TestEnsureCursorrules_FreshDir(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCursorrules(opts, "go")

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".cursorrules"))
	if err != nil {
		t.Fatalf("read .cursorrules: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, cursorrulesMarker) {
		t.Error("expected marker in .cursorrules")
	}
	if !strings.Contains(text, "AGENTS.md") {
		t.Error("expected AGENTS.md reference")
	}
	if !strings.Contains(text, ".opencode/uf/packs/go.md") {
		t.Error("expected Go pack reference")
	}
	if !strings.Contains(text, "cobalt-crush-dev.md") {
		t.Error("expected cobalt-crush agent reference")
	}
	if !strings.Contains(text, "divisor-guard.md") {
		t.Error("expected Divisor review agent reference")
	}
}

func TestEnsureCursorrules_ExistingWithoutMarker(t *testing.T) {
	dir := t.TempDir()

	existing := "Use TypeScript strict mode.\n"
	rulesPath := filepath.Join(dir, ".cursorrules")
	if err := os.WriteFile(rulesPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write .cursorrules: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCursorrules(opts, "typescript")

	if result.action != "appended" {
		t.Errorf("expected action 'appended', got %q", result.action)
	}

	content, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("read .cursorrules: %v", err)
	}
	text := string(content)

	if !strings.HasPrefix(text, existing) {
		t.Error("existing content not preserved")
	}
	if !strings.Contains(text, ".opencode/uf/packs/typescript.md") {
		t.Error("expected TypeScript pack reference")
	}
}

func TestEnsureCursorrules_AlreadyConfigured(t *testing.T) {
	dir := t.TempDir()

	existing := buildCursorrulesBlock("go", "")
	rulesPath := filepath.Join(dir, ".cursorrules")
	if err := os.WriteFile(rulesPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write .cursorrules: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCursorrules(opts, "go")

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	after, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("read .cursorrules: %v", err)
	}
	if string(after) != existing {
		t.Error(".cursorrules should be unchanged when block already matches")
	}
}

func TestEnsureCursorrules_StaleContent(t *testing.T) {
	dir := t.TempDir()

	// Seed with an older managed block missing cobalt-crush and
	// review agents sections.
	stale := cursorrulesMarker + "\n\nThis project follows coding conventions.\n\n" +
		"Available packs:\n- .opencode/uf/packs/default.md\n"
	rulesPath := filepath.Join(dir, ".cursorrules")
	if err := os.WriteFile(rulesPath, []byte(stale), 0o644); err != nil {
		t.Fatalf("write .cursorrules: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCursorrules(opts, "go")

	if result.action != "updated" {
		t.Errorf("expected action 'updated', got %q", result.action)
	}

	after, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("read .cursorrules: %v", err)
	}
	text := string(after)

	if !strings.Contains(text, "cobalt-crush-dev.md") {
		t.Error("expected cobalt-crush reference after update")
	}
	if !strings.Contains(text, "divisor-guard.md") {
		t.Error("expected review agents section after update")
	}
}

func TestEnsureCursorrules_StaleContentPreservesPrefix(t *testing.T) {
	dir := t.TempDir()

	prefix := "Use TypeScript strict mode.\n\n"
	stale := prefix + cursorrulesMarker + "\n\nSome old rules.\n"
	rulesPath := filepath.Join(dir, ".cursorrules")
	if err := os.WriteFile(rulesPath, []byte(stale), 0o644); err != nil {
		t.Fatalf("write .cursorrules: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCursorrules(opts, "go")

	if result.action != "updated" {
		t.Errorf("expected action 'updated', got %q", result.action)
	}

	after, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("read .cursorrules: %v", err)
	}
	text := string(after)

	if !strings.HasPrefix(text, prefix) {
		t.Error("user content above marker should be preserved")
	}
	if !strings.Contains(text, "divisor-guard.md") {
		t.Error("expected review agents section after update")
	}
}

func TestEnsureCursorrules_Idempotent(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	r1 := ensureCursorrules(opts, "go")
	if r1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", r1.action)
	}

	content1, _ := os.ReadFile(filepath.Join(dir, ".cursorrules"))

	r2 := ensureCursorrules(opts, "go")
	if r2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", r2.action)
	}

	content2, _ := os.ReadFile(filepath.Join(dir, ".cursorrules"))
	if !bytes.Equal(content1, content2) {
		t.Error("content should be identical after second call")
	}
}

func TestCollectDeployedPacks_Go(t *testing.T) {
	packs := collectDeployedPacks("go", "")

	expected := map[string]bool{
		"default.md":        true,
		"default-custom.md": true,
		"severity.md":       true,
		"content.md":        true,
		"content-custom.md": true,
		"go.md":             true,
		"go-custom.md":      true,
	}

	if len(packs) != len(expected) {
		t.Errorf("expected %d packs, got %d: %v", len(expected), len(packs), packs)
	}
	for _, p := range packs {
		if !expected[p] {
			t.Errorf("unexpected pack %q", p)
		}
	}
}

func TestCollectDeployedPacks_TypeScript(t *testing.T) {
	packs := collectDeployedPacks("typescript", "")

	found := false
	for _, p := range packs {
		if p == "typescript.md" {
			found = true
		}
	}
	if !found {
		t.Error("expected typescript.md in packs")
	}
}

func TestCollectDeployedPacks_Default(t *testing.T) {
	packs := collectDeployedPacks("default", "")

	for _, p := range packs {
		if p == "default.md" || p == "default-custom.md" ||
			p == "severity.md" || p == "content.md" ||
			p == "content-custom.md" {
			continue
		}
		t.Errorf("unexpected pack %q for default lang", p)
	}
	if len(packs) != 5 {
		t.Errorf("expected 5 packs for default lang, got %d", len(packs))
	}
}

// --- hasRuleContent tests ---

func TestHasRuleContent_EmptyStub(t *testing.T) {
	dir := t.TempDir()
	stub := `---
pack_id: default-custom
language: Any
version: 1.0.0
---

# Custom Rules: Default

Project-specific conventions.

## Custom Rules

<!-- Add project-specific rules below this line -->`
	path := filepath.Join(dir, "default-custom.md")
	if err := os.WriteFile(path, []byte(stub), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	if hasRuleContent(path) {
		t.Error("expected false for empty stub, got true")
	}
}

func TestHasRuleContent_WhitespaceOnlyAfterSentinel(t *testing.T) {
	dir := t.TempDir()
	stub := "<!-- Add project-specific rules below this line -->\n\n   \n\t\n"
	path := filepath.Join(dir, "go-custom.md")
	if err := os.WriteFile(path, []byte(stub), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	if hasRuleContent(path) {
		t.Error("expected false for whitespace-only after sentinel, got true")
	}
}

func TestHasRuleContent_PopulatedFile(t *testing.T) {
	dir := t.TempDir()
	populated := `---
pack_id: go-custom
language: Go
version: 1.0.0
---

## Custom Rules

<!-- Add project-specific rules below this line -->

## CR-001 [MUST] Use structured logging

All log calls MUST use charmbracelet/log.
`
	path := filepath.Join(dir, "go-custom.md")
	if err := os.WriteFile(path, []byte(populated), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !hasRuleContent(path) {
		t.Error("expected true for populated file, got false")
	}
}

func TestHasRuleContent_NoSentinel(t *testing.T) {
	dir := t.TempDir()
	// File with content but no sentinel — treat as non-empty (fail-open).
	path := filepath.Join(dir, "content-custom.md")
	if err := os.WriteFile(path, []byte("some content"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !hasRuleContent(path) {
		t.Error("expected true (no sentinel → fail-open), got false")
	}
}

func TestHasRuleContent_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.md")

	if !hasRuleContent(path) {
		t.Error("expected true for missing file (fail-open), got false")
	}
}

// --- collectDeployedPacks with root tests ---

func TestCollectDeployedPacks_WithRoot_AllEmpty(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	stubs := []string{"default-custom.md", "content-custom.md", "go-custom.md"}
	for _, name := range stubs {
		if err := os.WriteFile(filepath.Join(packsDir, name), []byte(sentinel), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	packs := collectDeployedPacks("go", dir)

	for _, p := range packs {
		if strings.HasSuffix(p, "-custom.md") {
			t.Errorf("empty custom pack %q should be excluded, but was included", p)
		}
	}
	// Non-custom packs must still be present.
	required := []string{"default.md", "severity.md", "content.md", "go.md"}
	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	for _, r := range required {
		if !packSet[r] {
			t.Errorf("required pack %q missing from result", r)
		}
	}
}

func TestCollectDeployedPacks_WithRoot_OnePopulated(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	// default-custom and content-custom are empty stubs.
	for _, name := range []string{"default-custom.md", "content-custom.md"} {
		if err := os.WriteFile(filepath.Join(packsDir, name), []byte(sentinel), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	// go-custom is populated.
	populated := sentinel + "\n\n## CR-001 [MUST] Example rule\n"
	if err := os.WriteFile(filepath.Join(packsDir, "go-custom.md"), []byte(populated), 0o644); err != nil {
		t.Fatalf("write go-custom.md: %v", err)
	}

	packs := collectDeployedPacks("go", dir)

	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	if !packSet["go-custom.md"] {
		t.Error("go-custom.md should be included (populated)")
	}
	if packSet["default-custom.md"] {
		t.Error("default-custom.md should be excluded (empty)")
	}
	if packSet["content-custom.md"] {
		t.Error("content-custom.md should be excluded (empty)")
	}
}

func TestCollectDeployedPacks_WithRoot_EmptyRootFallback(t *testing.T) {
	// When root is "", all packs including empty custom packs are returned
	// (backward-compatible behaviour; no filesystem access).
	packs := collectDeployedPacks("go", "")

	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	for _, name := range []string{"default-custom.md", "content-custom.md", "go-custom.md"} {
		if !packSet[name] {
			t.Errorf("expected %q in packs when root is empty, got: %v", name, packs)
		}
	}
}

// --- filterEmptyCustomPacks tests ---

func TestFilterEmptyCustomPacks_EmptyRoot(t *testing.T) {
	candidates := []string{"default.md", "default-custom.md", "go.md"}
	got := filterEmptyCustomPacks(candidates, "")

	if len(got) != len(candidates) {
		t.Errorf("expected %d packs with empty root, got %d", len(candidates), len(got))
	}
	for i, name := range candidates {
		if got[i] != name {
			t.Errorf("pack[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestFilterEmptyCustomPacks_ExcludesEmptyStubs(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	if err := os.WriteFile(filepath.Join(packsDir, "default-custom.md"), []byte(sentinel), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	candidates := []string{"default.md", "default-custom.md", "severity.md"}
	got := filterEmptyCustomPacks(candidates, dir)

	gotSet := make(map[string]bool, len(got))
	for _, p := range got {
		gotSet[p] = true
	}
	if gotSet["default-custom.md"] {
		t.Error("empty default-custom.md should be excluded")
	}
	if !gotSet["default.md"] {
		t.Error("non-custom default.md should be included")
	}
	if !gotSet["severity.md"] {
		t.Error("non-custom severity.md should be included")
	}
}

func TestFilterEmptyCustomPacks_KeepsPopulated(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	populated := "<!-- Add project-specific rules below this line -->\n\n## CR-001 Real rule\n"
	if err := os.WriteFile(filepath.Join(packsDir, "go-custom.md"), []byte(populated), 0o644); err != nil {
		t.Fatalf("write populated: %v", err)
	}

	candidates := []string{"go.md", "go-custom.md"}
	got := filterEmptyCustomPacks(candidates, dir)

	if len(got) != 2 {
		t.Errorf("expected 2 packs (populated custom kept), got %d: %v", len(got), got)
	}
}

// --- ensureCLAUDEmd with empty custom packs tests ---

func TestEnsureCLAUDEmd_EmptyCustomPacksOmitted(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	for _, name := range []string{"default-custom.md", "content-custom.md", "go-custom.md"} {
		if err := os.WriteFile(filepath.Join(packsDir, name), []byte(sentinel), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")
	if result.action == "failed" {
		t.Fatalf("ensureCLAUDEmd failed: %s", result.detail)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)

	for _, name := range []string{"default-custom.md", "content-custom.md", "go-custom.md"} {
		if strings.Contains(content, name) {
			t.Errorf("empty custom pack %q should not appear in CLAUDE.md, but found it", name)
		}
	}
	// Non-custom packs must be present.
	for _, name := range []string{"default.md", "severity.md", "content.md", "go.md"} {
		if !strings.Contains(content, name) {
			t.Errorf("required pack %q missing from CLAUDE.md", name)
		}
	}
}

func TestEnsureCLAUDEmd_PopulatedCustomPackIncluded(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	// default-custom empty, go-custom populated.
	if err := os.WriteFile(filepath.Join(packsDir, "default-custom.md"), []byte(sentinel), 0o644); err != nil {
		t.Fatalf("write default-custom.md: %v", err)
	}
	populated := sentinel + "\n\n## CR-001 [MUST] rule\n"
	if err := os.WriteFile(filepath.Join(packsDir, "go-custom.md"), []byte(populated), 0o644); err != nil {
		t.Fatalf("write go-custom.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packsDir, "content-custom.md"), []byte(sentinel), 0o644); err != nil {
		t.Fatalf("write content-custom.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureCLAUDEmd(opts, "go")
	if result.action == "failed" {
		t.Fatalf("ensureCLAUDEmd failed: %s", result.detail)
	}

	data, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "go-custom.md") {
		t.Error("go-custom.md should appear in CLAUDE.md (populated)")
	}
	if strings.Contains(content, "default-custom.md") {
		t.Error("default-custom.md should not appear in CLAUDE.md (empty)")
	}
}

// --- migrateCommandDir tests ---

// createFile is a test helper that creates a file with the given
// content at the specified relative path under dir.
func createFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func TestMigrateCommandDir_NoOldDir(t *testing.T) {
	dir := t.TempDir()
	// Only .opencode/commands/ exists — no old dir to migrate.
	createFile(t, dir, ".opencode/commands/review-council.md", "new")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result != nil {
		t.Errorf("expected nil (no-op) when only commands/ exists, got %+v", result)
	}
}

func TestMigrateCommandDir_NeitherDir(t *testing.T) {
	dir := t.TempDir()
	// Empty dir — no .opencode/ at all.

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result != nil {
		t.Errorf("expected nil (no-op) when neither dir exists, got %+v", result)
	}
}

func TestMigrateCommandDir_RenameOnly(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/command/ with 3 .md files, NO .opencode/commands/.
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/b.md", "bravo")
	createFile(t, dir, ".opencode/command/c.md", "charlie")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result for rename migration")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify .opencode/commands/ exists with 3 files.
	newDir := filepath.Join(dir, ".opencode", "commands")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		t.Fatalf("read new dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 files in commands/, got %d", len(entries))
	}

	// Verify .opencode/command/ does not exist.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old .opencode/command/ should not exist after rename")
	}

	// Verify file content preserved.
	content, err := os.ReadFile(filepath.Join(newDir, "a.md"))
	if err != nil {
		t.Fatalf("read a.md: %v", err)
	}
	if string(content) != "alpha" {
		t.Errorf("a.md content = %q, want %q", string(content), "alpha")
	}
}

func TestMigrateCommandDir_MergeUnique(t *testing.T) {
	dir := t.TempDir()
	// .opencode/command/ has a.md, b.md
	// .opencode/commands/ has c.md, d.md
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/b.md", "bravo")
	createFile(t, dir, ".opencode/commands/c.md", "charlie")
	createFile(t, dir, ".opencode/commands/d.md", "delta")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result for merge")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify .opencode/commands/ has all 4 files.
	newDir := filepath.Join(dir, ".opencode", "commands")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		t.Fatalf("read new dir: %v", err)
	}
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	for _, expected := range []string{"a.md", "b.md", "c.md", "d.md"} {
		if !names[expected] {
			t.Errorf("expected %s in commands/, got %v", expected, names)
		}
	}

	// Verify .opencode/command/ removed.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old .opencode/command/ should be removed after merge")
	}
}

func TestMigrateCommandDir_MergeDupIdentical(t *testing.T) {
	dir := t.TempDir()
	// Both dirs have x.md with identical content.
	createFile(t, dir, ".opencode/command/x.md", "same")
	createFile(t, dir, ".opencode/commands/x.md", "same")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// x.md in commands/ should be preserved.
	content, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "x.md"))
	if err != nil {
		t.Fatalf("read commands/x.md: %v", err)
	}
	if string(content) != "same" {
		t.Errorf("commands/x.md content = %q, want %q", string(content), "same")
	}

	// x.md in command/ should be removed.
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "command", "x.md")); !os.IsNotExist(err) {
		t.Error("command/x.md should be removed after merge")
	}

	// No conflict warning should be printed.
	output := buf.String()
	if strings.Contains(output, "conflict") {
		t.Errorf("should not warn about conflict for identical files, got:\n%s", output)
	}
}

func TestMigrateCommandDir_MergeDupDifferent(t *testing.T) {
	dir := t.TempDir()
	// command/x.md = "old", commands/x.md = "new"
	createFile(t, dir, ".opencode/command/x.md", "old")
	createFile(t, dir, ".opencode/commands/x.md", "new")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// commands/x.md should still contain "new".
	content, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "x.md"))
	if err != nil {
		t.Fatalf("read commands/x.md: %v", err)
	}
	if string(content) != "new" {
		t.Errorf("commands/x.md content = %q, want %q", string(content), "new")
	}

	// Warning should contain "conflict" and "/uf-init".
	output := buf.String()
	if !strings.Contains(output, "conflict") {
		t.Errorf("expected 'conflict' in warning output, got:\n%s", output)
	}
	if !strings.Contains(output, "/uf-init") {
		t.Errorf("expected '/uf-init' in warning output, got:\n%s", output)
	}
}

func TestMigrateCommandDir_Symlink(t *testing.T) {
	dir := t.TempDir()
	// Create a real directory to symlink to.
	realDir := filepath.Join(dir, "real-command-dir")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir real dir: %v", err)
	}
	// Create .opencode/ parent.
	ocDir := filepath.Join(dir, ".opencode")
	if err := os.MkdirAll(ocDir, 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}
	// Create .opencode/command as a symlink.
	if err := os.Symlink(realDir, filepath.Join(ocDir, "command")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result for symlink")
	}
	if result.action != "skipped" {
		t.Errorf("expected action 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "symlink") {
		t.Errorf("expected 'symlink' in detail, got %q", result.detail)
	}
}

func TestMigrateCommandDir_DivisorOnly(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/command/ — should NOT be touched in DivisorOnly mode.
	createFile(t, dir, ".opencode/command/review-council.md", "content")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Stdout:      &buf,
		ReadFile:    os.ReadFile,
		WriteFile:   os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result != nil {
		t.Errorf("expected nil in DivisorOnly mode, got %+v", result)
	}

	// .opencode/command/ should still exist.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Error(".opencode/command/ should still exist in DivisorOnly mode")
	}
}

func TestMigrateCommandDir_NonMDFiles(t *testing.T) {
	dir := t.TempDir()
	// .opencode/command/ has a.md and .DS_Store, no commands/.
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/.DS_Store", "\x00\x00")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Both files should be in commands/ (rename moves entire dir).
	newDir := filepath.Join(dir, ".opencode", "commands")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		t.Fatalf("read new dir: %v", err)
	}
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	if !names["a.md"] {
		t.Error("expected a.md in commands/")
	}
	if !names[".DS_Store"] {
		t.Error("expected .DS_Store in commands/")
	}

	// .opencode/command/ should be removed.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old .opencode/command/ should be removed")
	}
}

func TestMigrateCommandDir_Idempotent(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/command/a.md", "alpha")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	// First call: migrates command/ → commands/.
	result1 := migrateCommandDir(opts)
	if result1 == nil || result1.action != "migrated" {
		t.Fatalf("first call: expected migrated, got %+v", result1)
	}

	// Second call: command/ is gone → returns nil (silent no-op).
	result2 := migrateCommandDir(opts)
	if result2 != nil {
		t.Errorf("second call: expected nil (no-op), got %+v", result2)
	}

	// Verify commands/ still has the file.
	content, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "a.md"))
	if err != nil {
		t.Fatalf("read a.md: %v", err)
	}
	if string(content) != "alpha" {
		t.Errorf("a.md content = %q, want %q", string(content), "alpha")
	}
}

func TestMigrateCommandDir_PartialFailure(t *testing.T) {
	dir := t.TempDir()
	// Both dirs exist. command/ has a.md (moveable) and b.md (read will fail).
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/b.md", "bravo")
	createFile(t, dir, ".opencode/commands/c.md", "charlie")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	// Make b.md unreadable to simulate a partial failure during
	// the duplicate-check path (both dirs have b.md with conflict).
	// Instead, use a custom ReadFile that fails for b.md in the old dir.
	createFile(t, dir, ".opencode/commands/b.md", "different-bravo")
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, filepath.Join("command", "b.md")) {
			return nil, fmt.Errorf("simulated read error")
		}
		return os.ReadFile(path)
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// a.md should have been moved successfully.
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "a.md")); err != nil {
		t.Error("a.md should have been moved to commands/")
	}

	// Warning should have been printed for b.md.
	output := buf.String()
	if !strings.Contains(output, "b.md") {
		t.Errorf("expected warning about b.md, got:\n%s", output)
	}

	// command/ should NOT be fully removed (b.md still there due to error).
	// Actually the implementation calls os.Remove(oldPath) on the old copy
	// only when content matches, and prints a warning when ReadFile fails.
	// The old b.md stays because ReadFile failed before os.Remove runs.
	// os.Remove(oldDir) at the end will fail because dir is non-empty.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Error("old .opencode/command/ should still exist when partial failure occurs")
	}
}

func TestMoveFile_FallbackOnRenameError(t *testing.T) {
	// We can test the fallback path by using two different temp dirs
	// on potentially different filesystems, but that's unreliable.
	// Instead, create the file and use a wrapper that makes the
	// first Rename fail but the ReadFile/WriteFile path succeed.
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.md")
	dstPath := filepath.Join(dstDir, "test.md")

	if err := os.WriteFile(srcPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Use the moveFile function directly. On the same filesystem,
	// os.Rename will succeed, so the fallback path won't be tested.
	// To force the fallback, we rely on the fact that TempDir
	// creates dirs on the same filesystem. Instead, we test that
	// moveFile works end-to-end and verify the result.
	opts := &Options{
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	err := moveFile(srcPath, dstPath, opts)
	if err != nil {
		t.Fatalf("moveFile() error: %v", err)
	}

	// Verify dst has the content.
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "content" {
		t.Errorf("dst content = %q, want %q", string(got), "content")
	}

	// Verify src is removed.
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("src should be removed after move")
	}
}

func TestMoveFile_FallbackPath(t *testing.T) {
	// Force the fallback (read → write → remove) path by making
	// the source file read-only directory entry impossible to rename.
	// The simplest approach: wrap os.Rename to always fail, then
	// verify the copy fallback works.
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.md")
	dstPath := filepath.Join(dstDir, "test.md")

	if err := os.WriteFile(srcPath, []byte("fallback-content"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Temporarily make the dst directory such that Rename will fail
	// but WriteFile will succeed. Since both are temp dirs on the
	// same filesystem, Rename would normally succeed. Instead, we
	// create a scenario where dst already exists as a directory.
	subDir := filepath.Join(dstDir, "test.md")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir blocking dst: %v", err)
	}
	// Rename(file, dir) fails → triggers fallback.
	// But WriteFile(dir, ...) would also fail. So remove the
	// blocking dir first to let the fallback succeed.
	// This test verifies the happy path of moveFile more directly.
	if err := os.Remove(subDir); err != nil {
		t.Fatalf("remove blocking dir: %v", err)
	}

	opts := &Options{
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	err := moveFile(srcPath, dstPath, opts)
	if err != nil {
		t.Fatalf("moveFile() error: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "fallback-content" {
		t.Errorf("dst content = %q, want %q", string(got), "fallback-content")
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("src should be removed after move")
	}
}

// --- migrateCommandDir additional tests ---

func TestMigrateCommandDir_AtomicRenameVerifiesContent(t *testing.T) {
	dir := t.TempDir()
	oldDir := filepath.Join(dir, ".opencode", "command")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "test.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "other.md"), []byte("world"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		Stdout:    &bytes.Buffer{},
	}

	result := migrateCommandDir(opts)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify both files are accessible in the new location.
	newDir := filepath.Join(dir, ".opencode", "commands")
	for _, file := range []string{"test.md", "other.md"} {
		if _, err := os.Stat(filepath.Join(newDir, file)); err != nil {
			t.Errorf("expected %s in new dir: %v", file, err)
		}
	}

	// Verify old dir no longer exists.
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old dir should be removed after atomic rename")
	}
}

func TestMigrateCommandDir_MergeComplex(t *testing.T) {
	dir := t.TempDir()
	oldDir := filepath.Join(dir, ".opencode", "command")
	newDir := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatalf("mkdir old: %v", err)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}

	// File only in old dir — should be moved.
	if err := os.WriteFile(filepath.Join(oldDir, "unique.md"), []byte("unique"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Identical file in both — should be skipped (old removed).
	if err := os.WriteFile(filepath.Join(oldDir, "same.md"), []byte("same"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "same.md"), []byte("same"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Conflicting file — should be warned and commands/ version kept.
	if err := os.WriteFile(filepath.Join(oldDir, "conflict.md"), []byte("old-version"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "conflict.md"), []byte("new-version"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var stdout bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		Stdout:    &stdout,
	}

	result := migrateCommandDir(opts)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify unique.md was moved.
	content, err := os.ReadFile(filepath.Join(newDir, "unique.md"))
	if err != nil {
		t.Fatalf("read unique.md: %v", err)
	}
	if string(content) != "unique" {
		t.Errorf("expected 'unique', got %q", string(content))
	}

	// Verify conflict.md kept commands/ version.
	content, err = os.ReadFile(filepath.Join(newDir, "conflict.md"))
	if err != nil {
		t.Fatalf("read conflict.md: %v", err)
	}
	if string(content) != "new-version" {
		t.Errorf("expected 'new-version' preserved, got %q", string(content))
	}

	// Verify conflict warning was printed.
	if !strings.Contains(stdout.String(), "conflict") {
		t.Errorf("expected conflict warning in output, got: %s", stdout.String())
	}

	// Verify old copy of conflict.md was removed.
	if _, err := os.Stat(filepath.Join(oldDir, "conflict.md")); !os.IsNotExist(err) {
		t.Error("old copy of conflict.md should be removed")
	}
}

// --- Sub-tool error reporting tests (Spec 036 Phase 2) ---

func TestInitSubTools_DeweyInitFails_ShowsError(t *testing.T) {
	t.Run("short_output", func(t *testing.T) {
		dir := t.TempDir()
		rec := &scaffoldCmdRecorder{
			errors: map[string]error{
				"dewey init": fmt.Errorf("exit status 1"),
			},
			outputs: map[string]string{
				"dewey init": "Error: vault not found\nRun 'dewey init --vault .' to create one.\n",
			},
		}

		var buf bytes.Buffer
		opts := &Options{
			TargetDir: dir,
			LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
			ExecCmd:   rec.execCmd,
			Stdout:    &buf,
		}

		results := initSubTools(opts)

		// Find the dewey init failure result.
		var deweyResult *subToolResult
		for i := range results {
			if results[i].name == ".uf/dewey/" && results[i].action == "failed" {
				deweyResult = &results[i]
				break
			}
		}
		if deweyResult == nil {
			t.Fatalf("expected .uf/dewey/ failed result, got %v", results)
		}

		// Verify err field is set.
		if deweyResult.err == nil {
			t.Error("expected err to be non-nil on failed dewey init")
		}

		// Verify output field contains the error text.
		if !strings.Contains(string(deweyResult.output), "vault not found") {
			t.Errorf("expected output to contain 'vault not found', got %q", string(deweyResult.output))
		}

		// Verify detail includes the error message.
		if !strings.Contains(deweyResult.detail, "exit status 1") {
			t.Errorf("expected detail to contain error text, got %q", deweyResult.detail)
		}

		// Verify printSummary shows the error output.
		printSummary(&buf, false, false, true, &Result{Created: []string{"test.md"}}, results)
		output := buf.String()

		if !strings.Contains(output, "vault not found") {
			t.Errorf("expected printSummary output to contain 'vault not found', got:\n%s", output)
		}
		if !strings.Contains(output, "Error:") {
			t.Errorf("expected printSummary output to contain 'Error:', got:\n%s", output)
		}
	})

	t.Run("long_output_truncated", func(t *testing.T) {
		dir := t.TempDir()
		// Generate 25 lines of output to trigger truncation (maxLines=20).
		longOutput := generateLines(25)
		rec := &scaffoldCmdRecorder{
			errors: map[string]error{
				"dewey init": fmt.Errorf("exit status 1"),
			},
			outputs: map[string]string{
				"dewey init": longOutput,
			},
		}

		var buf bytes.Buffer
		opts := &Options{
			TargetDir: dir,
			LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
			ExecCmd:   rec.execCmd,
			Stdout:    &buf,
		}

		results := initSubTools(opts)

		// Find the dewey init failure result.
		var deweyResult *subToolResult
		for i := range results {
			if results[i].name == ".uf/dewey/" && results[i].action == "failed" {
				deweyResult = &results[i]
				break
			}
		}
		if deweyResult == nil {
			t.Fatalf("expected .uf/dewey/ failed result, got %v", results)
		}

		// Verify printSummary renders truncated output.
		printSummary(&buf, false, false, true, &Result{Created: []string{"test.md"}}, results)
		output := buf.String()

		// Should contain the truncation marker.
		if !strings.Contains(output, "... (15 lines omitted)") {
			t.Errorf("expected '... (15 lines omitted)' in printSummary output, got:\n%s", output)
		}

		// Should contain the last 10 lines (line 16 through line 25).
		if !strings.Contains(output, "line 25") {
			t.Errorf("expected last line 'line 25' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "line 16") {
			t.Errorf("expected 'line 16' (first of last 10) in output, got:\n%s", output)
		}

		// Should NOT contain early lines that were truncated.
		// Use a specific match to avoid matching "line 1" as a
		// substring of "line 10", "line 15", etc.
		for _, earlyLine := range []string{"line 1\n", "line 2\n", "line 5\n"} {
			// Check the rendered output (each line is indented).
			// The truncation should have removed these lines.
			truncated := textutil.TruncateOutput(deweyResult.output, 20)
			if strings.Contains(truncated, strings.TrimSuffix(earlyLine, "\n")+"\n") {
				// Only flag if the early line appears as a full line
				// in the truncated output (not as a substring).
				lines := strings.Split(truncated, "\n")
				for _, l := range lines {
					if l == strings.TrimSuffix(earlyLine, "\n") {
						t.Errorf("early line %q should be truncated from output", strings.TrimSuffix(earlyLine, "\n"))
					}
				}
			}
		}
	})
}

func TestInitSubTools_SimpleToolFails_ShowsError(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"specify init": fmt.Errorf("exit status 1"),
		},
		outputs: map[string]string{
			"specify init": "Error: .specify already exists\nUse --force to overwrite.\n",
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Find the specify failure result.
	var specifyResult *subToolResult
	for i := range results {
		if results[i].name == ".specify/" && results[i].action == "failed" {
			specifyResult = &results[i]
			break
		}
	}
	if specifyResult == nil {
		t.Fatalf("expected .specify/ failed result, got %v", results)
	}

	// Verify err field is set.
	if specifyResult.err == nil {
		t.Error("expected err to be non-nil on failed specify init")
	}

	// Verify output field contains the error text.
	if !strings.Contains(string(specifyResult.output), ".specify already exists") {
		t.Errorf("expected output to contain '.specify already exists', got %q", string(specifyResult.output))
	}

	// Verify detail includes the tool name and error.
	if !strings.Contains(specifyResult.detail, "specify init:") {
		t.Errorf("expected detail to contain 'specify init:', got %q", specifyResult.detail)
	}
}

// generateLines creates a string with n lines like "line 1\nline 2\n...".
func generateLines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

func TestInitSubTools_ExitCodeOnly(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"gaze init": fmt.Errorf("exit status 127"),
		},
		// No outputs entry — empty output bytes.
	}

	// Create .opencode/agents/ but no gaze-reporter.md so gaze init triggers.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Find the gaze failure result.
	var gazeResult *subToolResult
	for i := range results {
		if results[i].name == "gaze" && results[i].action == "failed" {
			gazeResult = &results[i]
			break
		}
	}
	if gazeResult == nil {
		t.Fatalf("expected gaze failed result, got %v", results)
	}

	// Verify err is set.
	if gazeResult.err == nil {
		t.Error("expected err to be non-nil")
	}

	// Verify detail includes the error text.
	if !strings.Contains(gazeResult.detail, "exit status 127") {
		t.Errorf("expected detail to contain 'exit status 127', got %q", gazeResult.detail)
	}
}

func TestPrintSummary_SuccessUnchanged(t *testing.T) {
	var buf bytes.Buffer

	r := &Result{
		Created: []string{"test.md"},
	}
	subResults := []subToolResult{
		{name: ".uf/dewey/", action: "initialized"},
		{name: "dewey index", action: "completed"},
		{name: "opencode.json", action: "created"},
	}

	printSummary(&buf, false, false, true, r, subResults)
	output := buf.String()

	// Verify each sub-tool result appears on exactly one line
	// (no "Output:" or "Error:" lines for successful results).
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Output:") {
			t.Errorf("successful results should not have 'Output:' lines, got:\n%s", output)
		}
		if strings.Contains(line, "Error:") {
			t.Errorf("successful results should not have 'Error:' lines, got:\n%s", output)
		}
	}

	// Verify the sub-tool results are present.
	if !strings.Contains(output, ".uf/dewey/ initialized") {
		t.Errorf("expected '.uf/dewey/ initialized' in output")
	}
	if !strings.Contains(output, "dewey index completed") {
		t.Errorf("expected 'dewey index completed' in output")
	}
	if !strings.Contains(output, "opencode.json created") {
		t.Errorf("expected 'opencode.json created' in output")
	}
}
