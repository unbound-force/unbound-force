package homebrew

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseWorkflow_HomebrewPublicationIsOrderedAndFailClosed(t *testing.T) {
	t.Parallel()

	workflow := releaseWorkflowRunBlock(t)
	if !strings.HasPrefix(strings.TrimSpace(workflow), "set -euo pipefail") {
		t.Fatal("Homebrew publication run block must begin with set -euo pipefail")
	}

	stages := []struct {
		name    string
		command string
	}{
		{"Cask checksum patching", "patch_checksums \"$CASK_FILE\""},
		{"transformer archive checksum validation", "TRANSFORMER_ACTUAL_SHA=$(shasum -a 256 \"$TRANSFORMER_ARCHIVE\" | awk '{print $1}')"},
		{"Cask transformation", "\"$TRANSFORMER_DIR/unbound-force\" transform-homebrew-cask"},
		{"semantic validation", "postflight_steps_count=$(grep -Fxc '  postflight_steps do' \"$CASK_FILE\" || true)"},
		{"candidate Cask copy", "cp \"$CASK_FILE\" \"$STAGED_TAP/Casks/unbound-force.rb\""},
		{"candidate Cask staging", "git -C \"$STAGED_TAP\" add Casks/unbound-force.rb"},
		{"candidate Cask commit", "commit -m \"test: stage release candidate Cask\""},
		{"tap registration for audit", "brew tap unbound-force/staging \"$STAGED_TAP\""},
		{"registered Cask verification", "cmp \"$CASK_FILE\" \"$REGISTERED_CASK\""},
		{"tap trust for cask operations", "brew trust unbound-force/staging"},
		{"Homebrew static validation", "brew audit --cask --strict unbound-force/staging/unbound-force"},
		{"dependency state capture", "if brew list --cask dewey >/dev/null 2>&1; then"},
		{"staged Cask smoke test", "brew install --cask unbound-force/staging/unbound-force"},
		{"smoke test dependency cleanup", "remove_smoke_test_casks\n          SMOKE_TEST_CASK_INSTALLED=false"},
		{"tap cleanup after smoke test", "SMOKE_TEST_CASK_INSTALLED=false\n          brew untap unbound-force/staging"},
		{"canonical tap copy", "cp \"$CASK_FILE\" tap/Casks/unbound-force.rb"},
		{"tap publication", "git push"},
	}

	previous := -1
	for _, stage := range stages {
		position := strings.Index(workflow, stage.command)
		if position == -1 {
			t.Fatalf("Homebrew publication run block is missing %s command %q", stage.name, stage.command)
		}
		if position <= previous {
			t.Fatalf("%s must run after the preceding publication gate", stage.name)
		}
		previous = position
	}

	for _, command := range []string{
		"\"$TRANSFORMER_DIR/unbound-force\" transform-homebrew-cask \\",
		"cmp \"$CASK_FILE\" \"$REGISTERED_CASK\"",
		"brew audit --cask --strict unbound-force/staging/unbound-force",
		"brew install --cask unbound-force/staging/unbound-force",
		"test -x \"$BREW_PREFIX/bin/unbound-force\"",
		"test -L \"$BREW_PREFIX/bin/uf\"",
		"test -x \"$BREW_PREFIX/bin/uf\"",
	} {
		assertFailFastCommand(t, workflow, command)
	}
	for _, validation := range []string{
		"if [ -z \"$TRANSFORMER_SHA\" ] || [ \"$TRANSFORMER_SHA\" != \"$TRANSFORMER_ACTUAL_SHA\" ]; then",
		"if [ \"$postflight_steps_count\" -ne 1 ]; then",
		"if grep -Fqx '  postflight do' \"$CASK_FILE\"; then",
		"if ! grep -Fqx '      run \"/usr/bin/xattr\",' \"$CASK_FILE\" || \\",
	} {
		assertValidationFailureExits(t, workflow, validation)
	}
	for _, cleanup := range []string{
		"if [ \"$DEWEY_CASK_WAS_INSTALLED\" = false ] && brew list --cask dewey >/dev/null 2>&1; then",
		"brew uninstall --cask --force dewey || cleanup_status=$?",
		"if [ \"$OLLAMA_APP_CASK_WAS_INSTALLED\" = false ] && brew list --cask ollama-app >/dev/null 2>&1; then",
		"brew uninstall --cask --force ollama-app || cleanup_status=$?",
	} {
		if !strings.Contains(workflow, cleanup) {
			t.Errorf("Homebrew publication cleanup missing %q", cleanup)
		}
	}
}

func TestReleaseWorkflow_ValidationMatchesTransformerOutput(t *testing.T) {
	t.Parallel()

	workflow := releaseWorkflowRunBlock(t)
	legacyLines := strings.Split(legacyPostflightHook, "\n")
	declarativeLines := strings.Split(declarativePostflightSteps, "\n")
	checks := []string{
		"postflight_steps_count=$(grep -Fxc '" + declarativeLines[0] + "' \"$CASK_FILE\" || true)",
		"if grep -Fqx '" + legacyLines[0] + "' \"$CASK_FILE\"; then",
		"grep -Fqx '" + declarativeLines[2] + "' \"$CASK_FILE\"",
		"grep -Fqx '" + declarativeLines[3] + "' \"$CASK_FILE\"",
		"grep -Fqx '" + declarativeLines[4] + "' \"$CASK_FILE\"",
	}
	for _, check := range checks {
		if !strings.Contains(workflow, check) {
			t.Errorf("release workflow validation missing transformer output check %q", check)
		}
	}
}

func TestReleaseWorkflow_SmokeCleanupRemovesOnlyIntroducedDependencies(t *testing.T) {
	t.Parallel()

	workflow := releaseWorkflowRunBlock(t)
	functionStart := strings.Index(workflow, "          remove_smoke_test_casks() {")
	if functionStart == -1 {
		t.Fatal("release workflow is missing remove_smoke_test_casks")
	}
	functionEnd := strings.Index(workflow[functionStart:], "\n          }\n")
	if functionEnd == -1 {
		t.Fatal("release workflow remove_smoke_test_casks function is incomplete")
	}
	functionEnd += functionStart + len("\n          }")
	cleanupFunction := workflow[functionStart:functionEnd]

	tests := []struct {
		name               string
		deweyWasInstalled  bool
		ollamaWasInstalled bool
		wantRemoved        []string
		wantPreserved      []string
	}{
		{
			name:          "introduced dependencies",
			wantRemoved:   []string{"unbound-force", "dewey", "ollama-app"},
			wantPreserved: nil,
		},
		{
			name:               "pre-existing dependencies",
			deweyWasInstalled:  true,
			ollamaWasInstalled: true,
			wantRemoved:        []string{"unbound-force"},
			wantPreserved:      []string{"dewey", "ollama-app"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tempDir := t.TempDir()
			logPath := filepath.Join(tempDir, "brew.log")
			binDir := filepath.Join(tempDir, "bin")
			if err := os.Mkdir(binDir, 0o755); err != nil {
				t.Fatalf("create fake Homebrew prefix: %v", err)
			}

			script := "BREW_PREFIX=\"$1\"\nLOG=\"$2\"\n" +
				"DEWEY_CASK_WAS_INSTALLED=" + boolString(test.deweyWasInstalled) + "\n" +
				"OLLAMA_APP_CASK_WAS_INSTALLED=" + boolString(test.ollamaWasInstalled) + "\n" +
				"brew() {\n" +
				"  printf '%s\\n' \"$*\" >> \"$LOG\"\n" +
				"  return 0\n" +
				"}\n" + cleanupFunction + "\nremove_smoke_test_casks\n"
			command := exec.Command("bash", "-c", script, "smoke-cleanup", tempDir, logPath)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("run smoke cleanup: %v\n%s", err, output)
			}

			logContent, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("read fake brew log: %v", err)
			}
			logOutput := string(logContent)
			for _, cask := range test.wantRemoved {
				command := "uninstall --cask --force " + cask
				if !strings.Contains(logOutput, command) {
					t.Errorf("cleanup did not run %q\n%s", command, logOutput)
				}
			}
			for _, cask := range test.wantPreserved {
				command := "uninstall --cask --force " + cask
				if strings.Contains(logOutput, command) {
					t.Errorf("cleanup unexpectedly ran %q\n%s", command, logOutput)
				}
			}
		})
	}
}

func TestReleaseWorkflow_ChecksumPatcherSupportsGeneratedLayouts(t *testing.T) {
	t.Parallel()

	workflow := releaseWorkflowRunBlock(t)
	functionStart := strings.Index(workflow, "          patch_checksums() {")
	if functionStart == -1 {
		t.Fatal("release workflow is missing patch_checksums")
	}
	functionEnd := strings.Index(workflow[functionStart:], "\n          }\n")
	if functionEnd == -1 {
		t.Fatal("release workflow patch_checksums function is incomplete")
	}
	functionEnd += functionStart + len("\n          }")

	const (
		arm64SHA = "26dd4fe6a21582b3506cec2c3e00fb632fcd0dededbe3a11bdfafb669a0c6cd0"
		amd64SHA = "c52c4d33e0bbb37e20e848ec31306f568407218a4f1865b864362299414de777"
		linuxSHA = "91d86f661882fce1969c3f19b215c885967905c91e9aba2943107f56a3856a08"
	)
	fixtures := map[string]string{
		"cask.rb": "" +
			"    on_arm do\n" +
			"      sha256 \"old-arm64\"\n" +
			"      url \"artifact_darwin_arm64.tar.gz\"\n" +
			"    end\n" +
			"    on_intel do\n" +
			"      sha256 \"old-amd64\"\n" +
			"      url \"artifact_darwin_amd64.tar.gz\"\n" +
			"    end\n" +
			"    on_arm do\n" +
			"      sha256 \"" + linuxSHA + "\"\n" +
			"      url \"artifact_linux_arm64.tar.gz\"\n" +
			"    end\n",
		"formula.rb": "" +
			"    if Hardware::CPU.intel?\n" +
			"      url \"artifact_darwin_amd64.tar.gz\"\n" +
			"      sha256 \"old-amd64\"\n" +
			"      define_method(:install) do\n" +
			"      end\n" +
			"    end\n" +
			"    if Hardware::CPU.arm?\n" +
			"      url \"artifact_darwin_arm64.tar.gz\"\n" +
			"      sha256 \"old-arm64\"\n" +
			"    end\n" +
			"    if Hardware::CPU.arm?\n" +
			"      url \"artifact_linux_arm64.tar.gz\"\n" +
			"      sha256 \"" + linuxSHA + "\"\n" +
			"    end\n",
	}

	tempDir := t.TempDir()
	paths := make([]string, 0, len(fixtures))
	for name, content := range fixtures {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s fixture: %v", name, err)
		}
		paths = append(paths, path)
	}

	script := "ARM64_SHA=" + arm64SHA + "\nAMD64_SHA=" + amd64SHA + "\n" +
		workflow[functionStart:functionEnd] + "\npatch_checksums \"$1\"\npatch_checksums \"$2\"\n"
	command := exec.Command("bash", "-c", script, "patch-checksums", paths[0], paths[1])
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run patch_checksums: %v\n%s", err, output)
	}

	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read patched fixture: %v", err)
		}
		patched := string(content)
		for name, checksum := range map[string]string{
			"arm64": arm64SHA,
			"amd64": amd64SHA,
			"linux": linuxSHA,
		} {
			if count := strings.Count(patched, checksum); count != 1 {
				t.Errorf("%s contains %s checksum %d times, want 1\n%s", filepath.Base(path), name, count, patched)
			}
		}
		if strings.Contains(patched, "old-") {
			t.Errorf("%s retained an unsigned Darwin checksum\n%s", filepath.Base(path), patched)
		}
	}
}

func releaseWorkflowRunBlock(t *testing.T) string {
	t.Helper()
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate workflow regression test")
	}
	workflowPath := filepath.Join(filepath.Dir(testFile), "..", "..", ".github", "workflows", "release.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	const step = "      - name: Update Homebrew cask and formula\n        run: |\n"
	start := strings.Index(string(content), step)
	if start == -1 {
		t.Fatal("release workflow is missing the Homebrew publication step")
	}
	runBlock := string(content)[start+len(step):]
	if end := strings.Index(runBlock, "        env:\n"); end != -1 {
		runBlock = runBlock[:end]
	}
	return runBlock
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func assertFailFastCommand(t *testing.T, workflow, command string) {
	t.Helper()
	position := strings.Index(workflow, command)
	if position == -1 {
		t.Fatalf("Homebrew publication run block is missing command %q", command)
	}
	lineStart := strings.LastIndex(workflow[:position], "\n") + 1
	lineEnd := strings.Index(workflow[position:], "\n")
	if lineEnd == -1 {
		lineEnd = len(workflow)
	} else {
		lineEnd += position
	}
	if strings.TrimSpace(workflow[lineStart:lineEnd]) != command {
		t.Fatalf("command %q must be a standalone fail-fast command", command)
	}
}

func assertValidationFailureExits(t *testing.T, workflow, validation string) {
	t.Helper()
	start := strings.Index(workflow, validation)
	if start == -1 {
		t.Fatalf("Homebrew publication run block is missing validation %q", validation)
	}
	blockEnd := strings.Index(workflow[start:], "\n          fi")
	if blockEnd == -1 {
		t.Fatalf("validation %q is missing its closing conditional", validation)
	}
	block := strings.TrimSpace(workflow[start : start+blockEnd])
	lastLine := block[strings.LastIndex(block, "\n")+1:]
	if strings.TrimSpace(lastLine) != "exit 1" {
		t.Fatalf("validation %q must exit before publication", validation)
	}
}
