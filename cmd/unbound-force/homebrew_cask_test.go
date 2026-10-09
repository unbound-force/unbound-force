package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const generatedHomebrewCaskFixture = `cask "unbound-force" do
  version "1.2.3"

  binary "unbound-force"

  postflight do
    if OS.mac?
      system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/unbound-force"]
      # Create uf symlink alias for daily-use convenience (FR-002/FR-009).
      system_command "/bin/ln", args: ["-sf", "#{staged_path}/unbound-force", "#{HOMEBREW_PREFIX}/bin/uf"]
    end
  end
end
`

func TestRunHomebrewCask_TransformsAndPrintsOutputPath(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "input.rb")
	outputPath := filepath.Join(directory, "output.rb")
	if err := os.WriteFile(inputPath, []byte(generatedHomebrewCaskFixture), 0o644); err != nil {
		t.Fatalf("write input Cask: %v", err)
	}

	var stdout bytes.Buffer
	if err := runHomebrewCask(homebrewCaskParams{
		inputPath:  inputPath,
		outputPath: outputPath,
		stdout:     &stdout,
	}); err != nil {
		t.Fatalf("runHomebrewCask() error = %v", err)
	}

	wantOutput := "Transformed Homebrew Cask: " + outputPath + "\n"
	if stdout.String() != wantOutput {
		t.Errorf("runHomebrewCask() output = %q, want %q", stdout.String(), wantOutput)
	}
	transformed, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read transformed Cask: %v", err)
	}
	if !strings.Contains(string(transformed), "  postflight_steps do") {
		t.Errorf("transformed Cask missing postflight_steps block:\n%s", transformed)
	}
}

func TestRunHomebrewCask_MissingInputReturnsError(t *testing.T) {
	directory := t.TempDir()
	err := runHomebrewCask(homebrewCaskParams{
		inputPath:  filepath.Join(directory, "missing.rb"),
		outputPath: filepath.Join(directory, "output.rb"),
		stdout:     &bytes.Buffer{},
	})
	if err == nil {
		t.Fatal("runHomebrewCask() error = nil, want missing input error")
	}
	if !strings.Contains(err.Error(), "read generated Homebrew Cask") {
		t.Errorf("runHomebrewCask() error = %q, want read context", err)
	}
}

func TestNewHomebrewCaskCmd_Execute(t *testing.T) {
	directory := t.TempDir()
	inputPath := filepath.Join(directory, "input.rb")
	outputPath := filepath.Join(directory, "output.rb")
	if err := os.WriteFile(inputPath, []byte(generatedHomebrewCaskFixture), 0o644); err != nil {
		t.Fatalf("write input Cask: %v", err)
	}

	command := newHomebrewCaskCmd()
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stdout)
	command.SetArgs([]string{"--input", inputPath, "--output", outputPath})

	if err := command.Execute(); err != nil {
		t.Fatalf("Homebrew Cask command error: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("transformed Cask state = %v, want output file", err)
	}
	if !strings.Contains(stdout.String(), outputPath) {
		t.Errorf("command output = %q, want output path %q", stdout.String(), outputPath)
	}
}
