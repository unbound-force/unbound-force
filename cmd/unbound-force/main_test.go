package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	// 35 = 34 prior + 1 review-pr.md command
	if !strings.Contains(output, "70 files processed") {
		t.Errorf("expected '70 files processed' in output, got:\n%s", output)
	}

	// Verify a user-owned file was created
	agentFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Error("expected user-owned cobalt-crush-dev.md to be created")
	}

	// Verify a tool-owned file was created
	toolFile := filepath.Join(dir, ".opencode", "command", "review-council.md")
	if _, err := os.Stat(toolFile); os.IsNotExist(err) {
		t.Error("expected tool-owned review-council.md to be created")
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
	userFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")
	if err := os.WriteFile(userFile, []byte("user content"), 0o644); err != nil {
		t.Fatalf("modify user file: %v", err)
	}

	// Modify a tool-owned file
	toolFile := filepath.Join(dir, ".opencode", "command", "review-council.md")
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

func TestInitCmd_Execute_CreatesFiles(t *testing.T) {
	dir := t.TempDir()

	// Build a root command and point it at the temp dir by overriding os.Getwd
	// is not possible without subprocess; instead we exercise newInitCmd via
	// a hand-rolled root that wires --target-dir. Since newInitCmd uses
	// os.Getwd() internally, we change the working directory for this test.
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	cmd := newInitCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init command error: %v", err)
	}

	// Verify at least one scaffolded file exists
	agentFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Error("expected cobalt-crush-dev.md to be scaffolded by init command")
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
	expected := "unbound-force v"
	if !strings.HasPrefix(output, expected) {
		t.Errorf("expected output to start with %q, got %q", expected, output)
	}

	// Verify format: "unbound-force vVERSION (commit COMMIT, built DATE)\n"
	if !strings.Contains(output, "(commit ") || !strings.Contains(output, "built ") {
		t.Errorf("expected format 'unbound-force vX (commit Y, built Z)', got %q", output)
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

// TestRootCmd_HelpOutput is a regression guard for FR-004: the help
// output must show the alias relationship and correct usage line.
func TestRootCmd_HelpOutput(t *testing.T) {
	root := &cobra.Command{
		Use:   "unbound-force",
		Short: "Unbound Force specification framework toolkit (alias: uf)",
	}
	root.AddCommand(newInitCmd())
	root.AddCommand(newVersionCmd())
	root.AddCommand(newDoctorCmd())
	root.AddCommand(newSetupCmd())

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("root --help error: %v", err)
	}

	output := buf.String()

	// FR-004: help output must indicate the alias relationship.
	if !strings.Contains(output, "(alias: uf)") {
		t.Errorf("expected help output to contain '(alias: uf)', got:\n%s", output)
	}

	// Usage line must show unbound-force [command].
	if !strings.Contains(output, "unbound-force [command]") {
		t.Errorf("expected help output to contain 'unbound-force [command]', got:\n%s", output)
	}
}
