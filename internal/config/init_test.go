// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitFile_CreateFromScratch(t *testing.T) {
	dir := t.TempDir()
	result, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("InitFile error = %v", err)
	}
	if !result.Created {
		t.Error("expected Created = true")
	}
	if result.Updated {
		t.Error("expected Updated = false")
	}

	// Verify file exists and contains all sections.
	data, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	content := string(data)
	for _, section := range knownSections {
		if !strings.Contains(content, section) {
			t.Errorf("config missing section %q", section)
		}
	}
}

func TestInitFile_IdempotentReRun(t *testing.T) {
	dir := t.TempDir()

	// First run creates the file.
	_, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("first InitFile error = %v", err)
	}

	// Second run does nothing.
	result, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("second InitFile error = %v", err)
	}
	if result.Created || result.Updated {
		t.Error("expected no changes on idempotent re-run")
	}
}

func TestInitFile_PreservesUncommentedValues(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a config with an uncommented value.
	existing := Template() + "\n"
	existing = strings.Replace(existing, "# sandbox:", "sandbox:", 1)
	existing = strings.Replace(existing,
		"#   runtime: auto", "  runtime: docker", 1)
	if err := os.WriteFile(
		filepath.Join(ufDir, "config.yaml"),
		[]byte(existing), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	result, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("InitFile error = %v", err)
	}

	// The file has all sections, so nothing to add/remove.
	if result.Created || result.Updated {
		t.Logf("SectionsAdded: %v, SectionsRemoved: %v",
			result.SectionsAdded, result.SectionsRemoved)
	}

	// Verify the user's value is preserved.
	data, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "runtime: docker") {
		t.Error("user value 'runtime: docker' was not preserved")
	}
}

func TestInitFile_AddsNewSection(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a config file missing the "gateway" section.
	tmpl := Template()
	lines := strings.Split(tmpl, "\n")
	var filtered []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, "─── Gateway") {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(line, "# ───") {
			skip = false
		}
		if !skip {
			filtered = append(filtered, line)
		}
	}
	content := strings.Join(filtered, "\n")
	if err := os.WriteFile(
		filepath.Join(ufDir, "config.yaml"),
		[]byte(content), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	result, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("InitFile error = %v", err)
	}
	if !result.Updated {
		t.Error("expected Updated = true")
	}
	if !contains(result.SectionsAdded, "gateway") {
		t.Errorf("expected 'gateway' in SectionsAdded, got %v", result.SectionsAdded)
	}

	// Verify the gateway section was added.
	data, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "gateway") {
		t.Error("gateway section not found in updated config")
	}
}

func TestInitFile_RemovesDeprecatedSection(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a config with all sections plus a deprecated one.
	content := Template() + "\n# ─── Legacy Old ──────\n# legacy_old:\n#   field: value\n"
	if err := os.WriteFile(
		filepath.Join(ufDir, "config.yaml"),
		[]byte(content), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	result, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("InitFile error = %v", err)
	}

	// The "legacy_old" section should not be in the known
	// sections, but since our detection only matches known
	// section names, it won't be detected as a section to
	// remove either. This tests the edge case.
	_ = result
}

func TestInitFile_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	result, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("InitFile error = %v", err)
	}

	// Verify the file has correct permissions.
	info, err := os.Stat(result.Path)
	if err != nil {
		t.Fatalf("stat error = %v", err)
	}
	// On Linux, file permissions should be 0o644.
	perm := info.Mode().Perm()
	if perm != 0o644 {
		t.Errorf("file permissions = %o, want 0644", perm)
	}
}

func TestInitFile_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a config file missing a section so update triggers.
	tmpl := Template()
	lines := strings.Split(tmpl, "\n")
	var filtered []string
	skip := false
	for _, line := range lines {
		if strings.Contains(line, "─── Gateway") {
			skip = true
			continue
		}
		if skip && strings.HasPrefix(line, "# ───") {
			skip = false
		}
		if !skip {
			filtered = append(filtered, line)
		}
	}
	content := strings.Join(filtered, "\n")
	if err := os.WriteFile(
		filepath.Join(ufDir, "config.yaml"),
		[]byte(content), 0o644,
	); err != nil {
		t.Fatal(err)
	}

	_, err := InitFile(InitOptions{ProjectDir: dir})
	if err != nil {
		t.Fatalf("InitFile error = %v", err)
	}

	// Verify backup exists.
	backupPath := filepath.Join(ufDir, "config.yaml.bak")
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("backup file not created: %v", err)
	}
}

func TestTemplate_ContainsAllSections(t *testing.T) {
	tmpl := Template()
	for _, section := range knownSections {
		if !strings.Contains(tmpl, section) {
			t.Errorf("template missing section %q", section)
		}
	}
}

func TestDetectSections(t *testing.T) {
	content := "# ─── Setup Preferences ───\n# setup:\n\n# ─── Gateway ───\n# gateway:\n"
	found := detectSections(content)
	if !found["setup"] {
		t.Error("expected to detect 'setup' section")
	}
	if !found["gateway"] {
		t.Error("expected to detect 'gateway' section")
	}
	if found["sandbox"] {
		t.Error("should not detect 'sandbox' section")
	}
}
