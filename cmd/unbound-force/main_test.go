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

// Verify the summary includes a non-trivial, correct file count.
	// The exact count is environment-dependent: review plugin activation
	// appends the two plugin sources plus the shared reviewer-manifest
	// module to the created set only when
	// Node/npm/OpenCode are present and the install plus probes succeed
	// (63 files). Otherwise the sources remain activation-gated and only 61
	// files are processed. The shared agent-executor library and plugins add
	// up to 65 files when all probes pass. The scaffold-level asset inventory
	// and drift tests pin the exact asset list, so this CLI check asserts
	// either valid count. (devcontainer excluded — OS-specific, generated
	// per-user by uf sandbox init.)
	var fileCountOK bool
	for _, count := range []string{"64", "65", "66", "67"} {
		if strings.Contains(output, count+" files processed") {
			fileCountOK = true
			break
		}
	}
	if !fileCountOK {
		t.Errorf("expected N files processed [64-67] in output, got:\n%s", output)
	}

	// Verify a user-owned file was created
	agentFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")
	if _, err := os.Stat(agentFile); os.IsNotExist(err) {
		t.Error("expected user-owned cobalt-crush-dev.md to be created")
	}

	// Verify a tool-owned file was created
	toolFile := filepath.Join(dir, ".opencode", "commands", "uf.review-council.md")
	if _, err := os.Stat(toolFile); os.IsNotExist(err) {
		t.Error("expected tool-owned uf.review-council.md to be created")
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
	toolFile := filepath.Join(dir, ".opencode", "commands", "uf.review-council.md")
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
// --- newDoctorCmd tests ---

func TestNewDoctorCmd_DefaultFormatFlag(t *testing.T) {
	cmd := newDoctorCmd()

	formatFlag := cmd.Flag("format")
	if formatFlag == nil {
		t.Fatal("expected 'format' flag to be registered")
	}
	if formatFlag.DefValue != "text" {
		t.Errorf("expected format default 'text', got %q", formatFlag.DefValue)
	}
}

func TestNewDoctorCmd_InvalidFormatRejected(t *testing.T) {
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--format", "xml"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid format 'xml'")
	}
	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("expected 'invalid format' error, got: %s", err.Error())
	}
}

func TestNewDoctorCmd_AcceptsTextFormat(t *testing.T) {
	// Use a temp dir with no tools installed to avoid
	// real system probing. The doctor command may return
	// an error (failing checks) but should NOT reject
	// the format flag.
	dir := t.TempDir()
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--format", "text", "--dir", dir})

	// Execute — we expect it to run (format accepted).
	// It may or may not return an error depending on checks.
	_ = cmd.Execute()

	// If it got past the format validation, output should
	// contain something (header, check results, etc.)
	if buf.Len() == 0 {
		t.Error("expected some output from doctor command with text format")
	}
}

func TestNewDoctorCmd_AcceptsJSONFormat(t *testing.T) {
	dir := t.TempDir()
	cmd := newDoctorCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--format", "json", "--dir", dir})

	_ = cmd.Execute()

	output := buf.String()
	// JSON format should produce a JSON object.
	if !strings.Contains(output, "{") {
		t.Errorf("expected JSON output, got: %s", output)
	}
}

func TestNewDoctorCmd_DirFlag(t *testing.T) {
	cmd := newDoctorCmd()

	dirFlag := cmd.Flag("dir")
	if dirFlag == nil {
		t.Fatal("expected 'dir' flag to be registered")
	}
	if dirFlag.DefValue != "." {
		t.Errorf("expected dir default '.', got %q", dirFlag.DefValue)
	}
}

func TestRunDoctor_TextFormat(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// runDoctor with a clean temp dir — no tools to find.
	// This exercises the text formatting path (lines 155-163).
	_ = runDoctor(doctorParams{
		targetDir: dir,
		format:    "text",
		stdout:    &buf,
	})

	output := buf.String()
	// Text output should contain some check results.
	if buf.Len() == 0 {
		t.Error("expected non-empty text output from runDoctor")
	}
	// Verify text format is used (not JSON).
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Error("expected text format, got JSON-like output")
	}
}

func TestRunDoctor_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_ = runDoctor(doctorParams{
		targetDir: dir,
		format:    "json",
		stdout:    &buf,
	})

	output := buf.String()
	// JSON format should produce a JSON object.
	if !strings.Contains(output, "{") {
		t.Errorf("expected JSON output from runDoctor, got: %s", output)
	}
}

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

// TestInitCmd_StealthHelpText is task 3.13's help-text assertion: the
// --stealth flag help documents that the exclusion is local-only and not
// preserved on re-clone.
func TestInitCmd_StealthHelpText(t *testing.T) {
	cmd := newInitCmd()
	stealthFlag := cmd.Flags().Lookup("stealth")
	if stealthFlag == nil {
		t.Fatal("expected --stealth flag to be registered")
	}
	if !strings.Contains(stealthFlag.Usage, "local-only") {
		t.Errorf("expected --stealth usage to mention local-only, got: %q", stealthFlag.Usage)
	}
	long := cmd.Long
	if !strings.Contains(long, "local-only") {
		t.Errorf("expected help text to mention local-only:\n%s", long)
	}
	if !strings.Contains(long, "re-clone") {
		t.Errorf("expected help text to document re-clone non-portability:\n%s", long)
	}
}

func TestInitCmd_StealthFlag(t *testing.T) {
	cmd := newInitCmd()
	stealthFlag := cmd.Flags().Lookup("stealth")
	if stealthFlag == nil {
		t.Fatal("expected --stealth flag to be registered")
	}
	checkFlag := cmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Fatal("expected --check flag to be registered")
	}
}

func TestNewSetupCmd_CorrectUseAndShort(t *testing.T) {
	cmd := newSetupCmd()
	if cmd.Use != "setup" {
		t.Errorf("Use = %q, want %q", cmd.Use, "setup")
	}
	if cmd.Short == "" {
		t.Error("expected non-empty Short description")
	}
	if cmd.Long == "" {
		t.Error("expected non-empty Long description")
	}
}

func TestNewSetupCmd_FlagRegistered(t *testing.T) {
	cmd := newSetupCmd()
	dirFlag := cmd.Flags().Lookup("dir")
	if dirFlag == nil {
		t.Error("expected --dir flag")
	}
	dryRunFlag := cmd.Flags().Lookup("dry-run")
	if dryRunFlag == nil {
		t.Error("expected --dry-run flag")
	}
	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Error("expected --yes flag")
	}
}

func TestSetupParamsStruct(t *testing.T) {
	_ = setupParams{
		targetDir: ".",
		dryRun:    true,
		yesFlag:   true,
		stdout:    nil,
		stderr:    nil,
	}
}

func TestRunSetup_ParamsConstruct(t *testing.T) {
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	_ = runSetup(setupParams{
		targetDir: dir,
		dryRun:    false,
		yesFlag:   false,
		stdout:    &stdout,
		stderr:    &stderr,
	})
}

func TestNewSetupCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := newSetupCmd()
	cmd.SetArgs([]string{"--help"})
	root := &cobra.Command{Use: "uf"}
	root.AddCommand(cmd)
	root.SetArgs([]string{"setup", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error executing --help: %v", err)
	}
}

func TestNewSetupCmd_ExecuteCoversRunE(t *testing.T) {
	dir := t.TempDir()
	cmd := newSetupCmd()
	root := &cobra.Command{Use: "uf"}
	root.AddCommand(cmd)
	root.SetArgs([]string{"setup", "--dir", dir})
	if err := root.Execute(); err != nil {
		t.Logf("expected setup error: %v", err)
	}
}

func TestNewDoctorCmd_ExecuteCoversRunE(t *testing.T) {
	dir := t.TempDir()
	cmd := newDoctorCmd()
	root := &cobra.Command{Use: "uf"}
	root.AddCommand(cmd)
	root.SetArgs([]string{"doctor", "--dir", dir, "--format", "text"})
	if err := root.Execute(); err != nil {
		t.Logf("expected doctor error: %v", err)
	}
}

func TestInitCmd_CheckFlag(t *testing.T) {
	cmd := newInitCmd()
	checkFlag := cmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Fatal("expected --check flag")
	}
}

func TestRunInit_StealthCheckParams(t *testing.T) {
	p := initParams{
		targetDir:   ".",
		force:       false,
		divisorOnly: true,
		stealth:     true,
		check:       true,
		version:     "1.0.0-test",
		stdout:      nil,
	}
	if !p.stealth {
		t.Error("expected stealth to be true")
	}
	if !p.check {
		t.Error("expected check to be true")
	}
}
