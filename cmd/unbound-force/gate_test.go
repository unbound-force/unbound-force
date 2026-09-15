package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/gate"
)

// scaffoldGateFixture creates an OpenSpec artifact tree for testing.
func scaffoldGateFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	changeDir := filepath.Join(dir, "openspec", "changes", "test-change")
	if err := os.MkdirAll(changeDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(changeDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func TestRunGate_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	scaffoldGateFixture(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "json",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gateExitCode(err) != 0 {
		t.Errorf("expected exit code 0, got %d", gateExitCode(err))
	}

	// Verify valid JSON.
	var report gate.GateReport
	if jsonErr := json.Unmarshal(stdout.Bytes(), &report); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, stdout.String())
	}
	if !report.Passed {
		t.Error("expected report.Passed=true")
	}
	if report.Version != gate.SchemaVersion {
		t.Errorf("expected version %s, got %s", gate.SchemaVersion, report.Version)
	}
}

func TestRunGate_TextOutput(t *testing.T) {
	dir := t.TempDir()
	scaffoldGateFixture(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "text",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gateExitCode(err) != 0 {
		t.Errorf("expected exit code 0, got %d", gateExitCode(err))
	}

	output := stdout.String()
	if !strings.Contains(output, "Gate: specify") {
		t.Errorf("expected 'Gate: specify' in output, got: %s", output)
	}
}

func TestRunGate_InvalidPhase(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "deploy",
		format:    "text",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for invalid phase")
	}
	if gateExitCode(err) != 1 {
		t.Errorf("expected exit code 1 for invalid phase, got %d", gateExitCode(err))
	}
}

func TestRunGate_InvalidFormat(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	// Note: format validation moved to RunE (Cobra level).
	// runGate now receives only validated formats, so passing
	// "xml" here falls through to the default text formatter
	// since gate.Run still accepts the options. This test
	// verifies the internal error wrapping for invalid phase
	// scenarios instead. The format validation is tested via
	// the Cobra command in TestNewGateCmd_InvalidFormat.
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "xml",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	// With an empty dir, the check will fail (exit 1), not
	// an internal error — since "xml" falls to default branch.
	if err == nil {
		t.Fatal("expected error for failing check")
	}
}

func TestRunGate_NonExistentDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: "/tmp/nonexistent-gate-cli-test-dir-xyz",
		phase:     "specify",
		format:    "text",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
	if gateExitCode(err) != 1 {
		t.Errorf("expected exit code 1 for non-existent dir, got %d", gateExitCode(err))
	}
}

func TestRunGate_CheckFailure_ExitCode1(t *testing.T) {
	dir := t.TempDir()
	// Empty directory — specify phase will fail.

	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "json",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for failing check")
	}
	if gateExitCode(err) != 1 {
		t.Errorf("expected exit code 1 for check failure, got %d", gateExitCode(err))
	}

	// Verify JSON output was still produced.
	var report gate.GateReport
	if jsonErr := json.Unmarshal(stdout.Bytes(), &report); jsonErr != nil {
		t.Fatalf("expected JSON output even on failure: %v\nraw: %s", jsonErr, stdout.String())
	}
	if report.Passed {
		t.Error("expected report.Passed=false")
	}
}

func TestRunGate_PathIsFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "somefile.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: filePath,
		phase:     "specify",
		format:    "text",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for file path")
	}
	if gateExitCode(err) != 1 {
		t.Errorf("expected exit code 1 for file path, got %d", gateExitCode(err))
	}
}

func TestNewGateCmd_InvalidFormat(t *testing.T) {
	cmd := newGateCmd()
	cmd.SetArgs([]string{"--phase", "specify", "--format", "xml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("expected 'invalid format' in error, got: %v", err)
	}
}
