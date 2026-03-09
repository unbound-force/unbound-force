package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInit_FreshDir(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	err := runInit(initParams{
		targetDir: dir,
		force:     false,
		version:   "1.0.0-test",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("runInit() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "files processed") {
		t.Errorf("expected output to contain 'files processed', got:\n%s", output)
	}

	// Verify the summary includes a non-trivial file count
	if !strings.Contains(output, "32 files processed") {
		t.Errorf("expected '32 files processed' in output, got:\n%s", output)
	}

	// Verify a user-owned file was created
	specTemplate := filepath.Join(dir, ".specify", "templates", "spec-template.md")
	if _, err := os.Stat(specTemplate); os.IsNotExist(err) {
		t.Error("expected user-owned spec-template.md to be created")
	}

	// Verify a tool-owned file was created
	toolFile := filepath.Join(dir, ".opencode", "command", "speckit.specify.md")
	if _, err := os.Stat(toolFile); os.IsNotExist(err) {
		t.Error("expected tool-owned speckit.specify.md to be created")
	}
}

func TestRunInit_ForceFlag(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run
	err := runInit(initParams{
		targetDir: dir,
		force:     false,
		version:   "1.0.0",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first runInit() error: %v", err)
	}

	// Modify a user-owned file
	userFile := filepath.Join(dir, ".specify", "templates", "spec-template.md")
	if err := os.WriteFile(userFile, []byte("user content"), 0o644); err != nil {
		t.Fatalf("modify user file: %v", err)
	}

	// Modify a tool-owned file
	toolFile := filepath.Join(dir, ".opencode", "command", "speckit.specify.md")
	if err := os.WriteFile(toolFile, []byte("tool content"), 0o644); err != nil {
		t.Fatalf("modify tool file: %v", err)
	}

	// Second run with force
	buf.Reset()
	err = runInit(initParams{
		targetDir: dir,
		force:     true,
		version:   "1.0.0",
		stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("force runInit() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "overwritten:") {
		t.Errorf("expected 'overwritten:' in force output, got:\n%s", output)
	}

	// Verify the user-owned file was overwritten
	content, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("read user file: %v", err)
	}
	if string(content) == "user content" {
		t.Error("expected user-owned file to be overwritten with --force")
	}

	// Verify the tool-owned file was overwritten
	content, err = os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("read tool file: %v", err)
	}
	if string(content) == "tool content" {
		t.Error("expected tool-owned file to be overwritten with --force")
	}
}

func TestVersionCmd_Output(t *testing.T) {
	cmd := newVersionCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version command error: %v", err)
	}

	output := buf.String()
	expected := "unbound v"
	if !strings.HasPrefix(output, expected) {
		t.Errorf("expected output to start with %q, got %q", expected, output)
	}

	// Verify format: "unbound vVERSION (commit COMMIT, built DATE)\n"
	if !strings.Contains(output, "(commit ") || !strings.Contains(output, "built ") {
		t.Errorf("expected format 'unbound vX (commit Y, built Z)', got %q", output)
	}

	// Verify the actual variable values are interpolated
	// Note: version var defaults to "dev" (set by ldflags in release builds)
	if !strings.Contains(output, "vdev") {
		t.Errorf("expected version 'vdev' in output, got %q", output)
	}
	if !strings.Contains(output, "commit none") {
		t.Errorf("expected 'commit none' in output, got %q", output)
	}
	if !strings.Contains(output, "built unknown") {
		t.Errorf("expected 'built unknown' in output, got %q", output)
	}
}
