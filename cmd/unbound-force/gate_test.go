package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/unbound-force/unbound-force/internal/gate"
)

// scaffoldGateFixture creates an OpenSpec artifact tree for testing.
func scaffoldGateFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	// Stop git discovery above the fixture so fallback-path tests do
	// not depend on ambient repository state.
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
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
	if exitCode(err) != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode(err))
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
	if exitCode(err) != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode(err))
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
	if exitCode(err) != 1 {
		t.Errorf("expected exit code 1 for invalid phase, got %d", exitCode(err))
	}
}

func TestRunGate_UnrecognizedFormat_FallsToText(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	// Note: format validation lives in RunE (Cobra level), covered
	// by TestNewGateCmd_InvalidFormat. runGate accepts unvalidated
	// formats, so "xml" here falls through to the default text
	// formatter. With an empty dir the check fails (exit 1), not
	// an internal error.
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
	if exitCode(err) != 1 {
		t.Errorf("expected exit code 1 for check failure, got %d", exitCode(err))
	}
	if !strings.Contains(stdout.String(), "Gate: specify") {
		t.Errorf("expected text formatter output, got: %s", stdout.String())
	}
}

func TestRunGate_NonExistentDir(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: filepath.Join(dir, "nonexistent"),
		phase:     "specify",
		format:    "text",
		stdout:    &stdout,
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
	if exitCode(err) != 1 {
		t.Errorf("expected exit code 1 for non-existent dir, got %d", exitCode(err))
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
	if exitCode(err) != 1 {
		t.Errorf("expected exit code 1 for check failure, got %d", exitCode(err))
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
	if exitCode(err) != 1 {
		t.Errorf("expected exit code 1 for file path, got %d", exitCode(err))
	}
}

func TestNewGateCmd_InvalidFormat(t *testing.T) {
	cmd := newGateCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--phase", "specify", "--format", "xml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("expected 'invalid format' in error, got: %v", err)
	}
	// GATE-FR-003: the error message goes to stderr, keeping stdout clean.
	if !strings.Contains(errBuf.String(), "invalid format") {
		t.Errorf("expected 'invalid format' on stderr, got: %q", errBuf.String())
	}
	if out.String() != "" {
		t.Errorf("expected empty stdout on validation error, got: %q", out.String())
	}
}

func TestNewGateCmd_Help(t *testing.T) {
	var out bytes.Buffer
	cmd := newGateCmd()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help execution: %v", err)
	}
	text := out.String()
	for _, want := range []string{"specify", "plan", "implement", "review", "pr", "--format", "Exit codes"} {
		if !strings.Contains(text, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}

func TestNewGateCmd_MissingPhase(t *testing.T) {
	cmd := newGateCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing phase")
	}
	if !strings.Contains(err.Error(), "--phase is required") {
		t.Errorf("expected '--phase is required' in error, got: %v", err)
	}
}

func TestExitCode_MapsErrorsToProcessCodes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil error", nil, 0},
		{"plain error", fmt.Errorf("check failed"), 1},
		{"internal error", &gate.InternalError{
			Err: fmt.Errorf("boom"),
		}, 2},
		{"wrapped internal error", fmt.Errorf(
			"wrap: %w",
			&gate.InternalError{Err: fmt.Errorf("boom")},
		), 2},
		// *exec.ExitError satisfies exitCoder structurally, but a
		// subprocess failure must not leak its child status into the
		// CLI's exit code: raw and wrapped exec errors exit 1.
		{"exec exit error", &exec.ExitError{}, 1},
		{"wrapped exec exit error", fmt.Errorf(
			"check container: %w",
			&exec.ExitError{},
		), 1},
		// An explicit opt-in wins even when it wraps an exec error.
		{"internal error wrapping exec exit error", &gate.InternalError{
			Err: &exec.ExitError{},
		}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := exitCode(tt.err); got != tt.want {
				t.Errorf("exitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNewGateCmd_Execute(t *testing.T) {
	dir := t.TempDir()
	scaffoldGateFixture(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	cmd := newGateCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--phase", "specify", "--dir", dir, "--format", "json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var report gate.GateReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, out.String())
	}
	if !report.Passed {
		t.Error("expected report.Passed=true")
	}
	// On success, stderr stays clean — the report is stdout-only.
	if errBuf.Len() != 0 {
		t.Errorf("expected empty stderr on success, got: %q", errBuf.String())
	}
}

func TestNewGateCmd_Execute_DefaultDir(t *testing.T) {
	fixture := t.TempDir()
	scaffoldGateFixture(t, fixture, map[string]string{
		"proposal.md": "# Proposal",
	})
	t.Chdir(fixture)

	cmd := newGateCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--phase", "specify", "--format", "text"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "Gate: specify") {
		t.Errorf("expected text output, got: %s", out.String())
	}
}

// failingWriter always fails on Write, simulating a broken
// stdout to exercise the format-failure internal error path
// (both FormatJSON and FormatText propagate write errors).
type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("simulated write failure")
}

func TestRunGate_FormatFailure_ExitCode2(t *testing.T) {
	dir := t.TempDir()
	scaffoldGateFixture(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "json",
		stdout:    failingWriter{},
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for format write failure")
	}
	if exitCode(err) != 2 {
		t.Errorf("expected exit code 2 for internal error, got %d", exitCode(err))
	}
}

func TestRunGate_ArtifactReadFailure_ExitCode2(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	openspecDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(openspecDir, 0o755); err != nil {
		t.Fatalf("mkdir openspec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(openspecDir, "changes"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write malformed artifact root: %v", err)
	}

	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "json",
		stdout:    &bytes.Buffer{},
		stderr:    &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("expected error for artifact read failure")
	}
	if exitCode(err) != 2 {
		t.Errorf("expected exit code 2 for artifact read failure, got %d", exitCode(err))
	}
}

func TestRunGate_TextFormatFailure_ExitCode2(t *testing.T) {
	dir := t.TempDir()
	scaffoldGateFixture(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stderr bytes.Buffer
	err := runGate(gateParams{
		targetDir: dir,
		phase:     "specify",
		format:    "text",
		stdout:    failingWriter{},
		stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for text format write failure")
	}
	if exitCode(err) != 2 {
		t.Errorf("expected exit code 2 for internal error, got %d", exitCode(err))
	}
}

func TestNewGateCmd_ValidationError_StreamSeparation(t *testing.T) {
	dir := t.TempDir()
	cmd := newGateCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--phase", "deploy", "--dir", dir})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid phase")
	}

	// Validation errors must reach stderr, not stdout.
	if !strings.Contains(errBuf.String(), "invalid") {
		t.Errorf("expected validation error on stderr, got: %q", errBuf.String())
	}
	if strings.Contains(out.String(), "invalid") {
		t.Errorf("validation error leaked to stdout: %q", out.String())
	}
	// Stdout must not carry a report on a validation error.
	var report gate.GateReport
	if json.Unmarshal(out.Bytes(), &report) == nil {
		t.Errorf("stdout should not contain a gate report on validation error: %q", out.String())
	}
}

func TestNewGateCmd_CheckFailure_StreamSeparation(t *testing.T) {
	// Empty directory — the specify check fails.
	dir := t.TempDir()
	cmd := newGateCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--phase", "specify", "--dir", dir, "--format", "json"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for failing check")
	}

	// The report is the only stdout content: pure, parseable JSON.
	var report gate.GateReport
	if jsonErr := json.Unmarshal(out.Bytes(), &report); jsonErr != nil {
		t.Fatalf("stdout should be parseable JSON on check failure: %v\nraw: %s", jsonErr, out.String())
	}
	if report.Passed {
		t.Error("expected report.Passed=false")
	}
	// The error line goes to stderr.
	if !strings.Contains(errBuf.String(), "Error:") {
		t.Errorf("expected error line on stderr, got: %q", errBuf.String())
	}
}

func TestNewGateCmd_InternalError_StreamSeparation(t *testing.T) {
	// A file where the openspec/changes directory belongs makes the
	// artifact scan fail with an internal error (exit code 2).
	dir := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	openspecDir := filepath.Join(dir, "openspec")
	if err := os.MkdirAll(openspecDir, 0o755); err != nil {
		t.Fatalf("mkdir openspec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(openspecDir, "changes"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write malformed artifact root: %v", err)
	}

	cmd := newGateCmd()
	var out, errBuf bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errBuf)
	cmd.SetArgs([]string{"--phase", "specify", "--dir", dir})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for internal failure")
	}
	if exitCode(err) != 2 {
		t.Errorf("expected exit code 2 for internal error, got %d", exitCode(err))
	}
	// GATE-FR-009: the internal-error message reaches stderr.
	if !strings.Contains(errBuf.String(), "Error:") {
		t.Errorf("expected error line on stderr, got: %q", errBuf.String())
	}
	if strings.Contains(out.String(), "Error:") {
		t.Errorf("error leaked to stdout: %q", out.String())
	}
}
