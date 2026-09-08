package gate

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// --- helpers ---

// scaffoldOpenSpec creates an OpenSpec artifact tree inside dir.
func scaffoldOpenSpec(t *testing.T, dir string, files map[string]string) {
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

// scaffoldSpeckit creates a Speckit artifact tree inside dir.
func scaffoldSpeckit(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	specDir := filepath.Join(dir, "specs", "001-test-feature")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(specDir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

// initGitRepo creates a git repository in dir with a branch.
func initGitRepo(t *testing.T, dir, branch string) {
	t.Helper()
	runGit(t, dir, "init", "--initial-branch=main")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	// Create initial commit so branch operations work.
	dummy := filepath.Join(dir, ".gitkeep")
	if err := os.WriteFile(dummy, []byte(""), 0o644); err != nil {
		t.Fatalf("write .gitkeep: %v", err)
	}
	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "init")
	if branch != "main" && branch != "" {
		runGit(t, dir, "checkout", "-b", branch)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// --- checkSpecExists tests ---

func TestCheckSpecExists_OpenSpec(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal\nSome content",
	})

	r := checkSpecExists(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
	if r.Name != "spec-exists" {
		t.Errorf("expected name spec-exists, got %s", r.Name)
	}
}

func TestCheckSpecExists_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"spec.md": "# Spec\nSome content",
	})

	r := checkSpecExists(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckSpecExists_Missing(t *testing.T) {
	dir := t.TempDir()

	r := checkSpecExists(dir)
	if r.Passed {
		t.Errorf("expected fail, got pass")
	}
}

// --- checkPlanExists tests ---

func TestCheckPlanExists_OpenSpec(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"design.md": "# Design\nSome content",
	})

	r := checkPlanExists(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckPlanExists_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"plan.md": "# Plan\nSome content",
	})

	r := checkPlanExists(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckPlanExists_Missing(t *testing.T) {
	dir := t.TempDir()

	r := checkPlanExists(dir)
	if r.Passed {
		t.Errorf("expected fail, got pass")
	}
}

// --- checkTasksExist tests ---

func TestCheckTasksExist_OpenSpec(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [ ] Task 1",
	})

	r := checkTasksExist(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckTasksExist_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [ ] Task 1",
	})

	r := checkTasksExist(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckTasksExist_Missing(t *testing.T) {
	dir := t.TempDir()

	r := checkTasksExist(dir)
	if r.Passed {
		t.Errorf("expected fail, got pass")
	}
}

// --- checkCoverageStrategy tests ---

func TestCheckCoverageStrategy_Heading(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\n## Coverage strategy\n\nUnit tests for all functions.",
	})

	r := checkCoverageStrategy(dir)
	if !r.Passed {
		t.Errorf("expected pass for coverage heading, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_Keyword(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\ncoverage target: >= 80%\n",
	})

	r := checkCoverageStrategy(dir)
	if !r.Passed {
		t.Errorf("expected pass for coverage keyword, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_UnitTestKeyword(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\nWe need a unit test for each function.\n",
	})

	r := checkCoverageStrategy(dir)
	if !r.Passed {
		t.Errorf("expected pass for 'unit test' keyword, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_Missing(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\n- [ ] Implement feature\n- [ ] Write docs\n",
	})

	r := checkCoverageStrategy(dir)
	if r.Passed {
		t.Errorf("expected fail for missing coverage strategy, got pass")
	}
}

func TestCheckCoverageStrategy_NoFiles(t *testing.T) {
	dir := t.TempDir()

	r := checkCoverageStrategy(dir)
	if r.Passed {
		t.Errorf("expected fail when no task/plan files exist, got pass")
	}
}

// --- checkTasksComplete tests ---

func TestCheckTasksComplete_AllDone(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [x] Task 2\n",
	})

	r := checkTasksComplete(dir)
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckTasksComplete_Incomplete(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [ ] Task 2\n",
	})

	r := checkTasksComplete(dir)
	if r.Passed {
		t.Errorf("expected fail for incomplete tasks, got pass")
	}
}

func TestCheckTasksComplete_NoFile(t *testing.T) {
	dir := t.TempDir()

	r := checkTasksComplete(dir)
	if r.Passed {
		t.Errorf("expected fail when no tasks.md, got pass")
	}
}

// --- checkReviewRun tests ---

func TestCheckReviewRun_SpecReview(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- spec-review: passed -->",
	})

	r := checkReviewRun(dir)
	if !r.Passed {
		t.Errorf("expected pass for spec-review marker, got fail: %s", r.Message)
	}
}

func TestCheckReviewRun_CodeReview(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- code-review: passed -->",
	})

	r := checkReviewRun(dir)
	if !r.Passed {
		t.Errorf("expected pass for code-review marker, got fail: %s", r.Message)
	}
}

func TestCheckReviewRun_NoMarker(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n",
	})

	r := checkReviewRun(dir)
	if r.Passed {
		t.Errorf("expected fail when no review marker, got pass")
	}
}

// --- checkReviewPass tests ---

func TestCheckReviewPass_Present(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- code-review: passed -->",
	})

	r := checkReviewPass(dir)
	if !r.Passed {
		t.Errorf("expected pass for code-review marker, got fail: %s", r.Message)
	}
}

func TestCheckReviewPass_OnlySpecReview(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- spec-review: passed -->",
	})

	r := checkReviewPass(dir)
	if r.Passed {
		t.Errorf("expected fail when only spec-review marker present, got pass")
	}
}

func TestCheckReviewPass_NoMarker(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n",
	})

	r := checkReviewPass(dir)
	if r.Passed {
		t.Errorf("expected fail when no marker, got pass")
	}
}

func TestCheckReviewPass_NoFile(t *testing.T) {
	dir := t.TempDir()

	r := checkReviewPass(dir)
	if r.Passed {
		t.Errorf("expected fail when no tasks.md, got pass")
	}
	if r.Name != "review-pass" {
		t.Errorf("expected name review-pass, got %s", r.Name)
	}
}

// --- checkReviewRun with Speckit layout ---

func TestCheckReviewRun_SpeckitLayout(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- spec-review: passed -->",
	})

	r := checkReviewRun(dir)
	if !r.Passed {
		t.Errorf("expected pass for speckit layout, got fail: %s", r.Message)
	}
}

func TestCheckReviewRun_NoFile(t *testing.T) {
	dir := t.TempDir()

	r := checkReviewRun(dir)
	if r.Passed {
		t.Errorf("expected fail when no tasks.md, got pass")
	}
}

// --- checkTasksComplete with Speckit layout ---

func TestCheckTasksComplete_SpeckitLayout(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [x] Task 2\n",
	})

	r := checkTasksComplete(dir)
	if !r.Passed {
		t.Errorf("expected pass for speckit layout, got fail: %s", r.Message)
	}
}

// --- checkCoverageStrategy with Speckit layout ---

func TestCheckCoverageStrategy_SpeckitPlan(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"plan.md": "# Plan\n\n## Coverage Strategy\n\nUnit tests for all.\n",
	})

	r := checkCoverageStrategy(dir)
	if !r.Passed {
		t.Errorf("expected pass for speckit plan, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_PercentagePattern(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\nTarget: coverage >= 90%\n",
	})

	r := checkCoverageStrategy(dir)
	if !r.Passed {
		t.Errorf("expected pass for percentage pattern, got fail: %s", r.Message)
	}
}

// --- checkNotOnMain tests ---

func TestCheckNotOnMain_FeatureBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir, "opsx/test-feature")

	r := checkNotOnMain(dir)
	if !r.Passed {
		t.Errorf("expected pass on feature branch, got fail: %s", r.Message)
	}
}

func TestCheckNotOnMain_OnMain(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir, "main")

	r := checkNotOnMain(dir)
	if r.Passed {
		t.Errorf("expected fail on main branch, got pass")
	}
}

func TestCheckNotOnMain_NonGitDir(t *testing.T) {
	dir := t.TempDir()

	r := checkNotOnMain(dir)
	if r.Passed {
		t.Errorf("expected fail in non-git directory, got pass")
	}
	if !strings.Contains(r.Message, "git repository") {
		t.Errorf("expected message about git repository, got: %s", r.Message)
	}
}

func TestCheckNotOnMain_DetachedHead(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir, "main")
	// Detach HEAD.
	runGit(t, dir, "checkout", "--detach")

	r := checkNotOnMain(dir)
	if !r.Passed {
		t.Errorf("expected pass on detached HEAD, got fail: %s", r.Message)
	}
}

// --- Run() integration tests ---

func TestRun_SpecifyPhase_Pass(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "specify",
		Format:    "text",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if !report.Passed {
		t.Errorf("expected report.Passed=true, got false")
	}
	if report.Summary.Total != 1 {
		t.Errorf("expected 1 check, got %d", report.Summary.Total)
	}
}

func TestRun_ImplementPhase_Fail(t *testing.T) {
	dir := t.TempDir()
	// Empty directory — all checks should fail.

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "implement",
		Format:    "json",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for failing checks")
	}
	if report == nil {
		t.Fatal("expected report even on failure")
	}
	if report.Passed {
		t.Errorf("expected report.Passed=false, got true")
	}
	if report.Summary.Failed != 4 {
		t.Errorf("expected 4 failures, got %d", report.Summary.Failed)
	}
	if report.Summary.Total != 4 {
		t.Errorf("expected 4 total checks, got %d", report.Summary.Total)
	}
}

func TestRun_PlanPhase_Pass(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
		"design.md":   "# Design",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "plan",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Passed {
		t.Errorf("expected report.Passed=true, got false")
	}
	if report.Summary.Total != 2 {
		t.Errorf("expected 2 checks, got %d", report.Summary.Total)
	}
}

func TestRun_PlanPhase_FailMissingPlan(t *testing.T) {
	dir := t.TempDir()
	// Only spec, no plan/design.
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "plan",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for failing check")
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", report.Summary.Failed)
	}
	if report.Summary.Passed != 1 {
		t.Errorf("expected 1 pass, got %d", report.Summary.Passed)
	}
}

func TestRun_ImplementPhase_Pass(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
		"design.md":   "# Design",
		"tasks.md":    "# Tasks\n\n## Coverage Strategy\n\n>= 90% line coverage\n",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "implement",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Passed {
		t.Errorf("expected report.Passed=true, got false")
	}
	if report.Summary.Total != 4 {
		t.Errorf("expected 4 checks, got %d", report.Summary.Total)
	}
	if report.Summary.Passed != 4 {
		t.Errorf("expected 4 passed, got %d", report.Summary.Passed)
	}
}

func TestRun_InvalidPhase(t *testing.T) {
	dir := t.TempDir()

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "deploy",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for invalid phase")
	}
	if report != nil {
		t.Errorf("expected nil report for invalid phase, got %+v", report)
	}
}

func TestRun_NonExistentDir(t *testing.T) {
	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: "/tmp/nonexistent-gate-test-dir-xyz",
		Phase:     "specify",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
	if report != nil {
		t.Errorf("expected nil report for non-existent dir")
	}
}

func TestRun_Provenance(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "specify",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Version != SchemaVersion {
		t.Errorf("expected version %s, got %s", SchemaVersion, report.Version)
	}
	if report.Producer != ProducerName {
		t.Errorf("expected producer %s, got %s", ProducerName, report.Producer)
	}
	if report.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
	if report.Phase != "specify" {
		t.Errorf("expected phase specify, got %s", report.Phase)
	}
}

// --- FormatJSON tests ---

func TestFormatJSON_ValidOutput(t *testing.T) {
	report := &GateReport{
		Version:   SchemaVersion,
		Producer:  ProducerName,
		Timestamp: "2026-09-08T12:00:00Z",
		Branch:    "opsx/test",
		Phase:     "specify",
		Passed:    true,
		Checks: []CheckResult{
			{Name: "spec-exists", Passed: true, Message: "Spec found"},
		},
		Summary: Summary{Total: 1, Passed: 1, Failed: 0},
	}

	var buf bytes.Buffer
	if err := FormatJSON(report, &buf); err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Verify it's valid JSON.
	var decoded GateReport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if decoded.Phase != "specify" {
		t.Errorf("expected phase specify, got %s", decoded.Phase)
	}
	if !decoded.Passed {
		t.Error("expected passed=true")
	}
	if decoded.Version != SchemaVersion {
		t.Errorf("expected version %s, got %s", SchemaVersion, decoded.Version)
	}
}

// --- FormatText tests ---

func TestFormatText_PassOutput(t *testing.T) {
	report := &GateReport{
		Version:   SchemaVersion,
		Producer:  ProducerName,
		Timestamp: "2026-09-08T12:00:00Z",
		Phase:     "specify",
		Passed:    true,
		Checks: []CheckResult{
			{Name: "spec-exists", Passed: true, Message: "Spec found"},
		},
		Summary: Summary{Total: 1, Passed: 1, Failed: 0},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Gate: specify") {
		t.Errorf("expected 'Gate: specify' in output, got: %s", output)
	}
	if !strings.Contains(output, "PASS") {
		t.Errorf("expected 'PASS' in output, got: %s", output)
	}
}

func TestFormatText_FailOutput(t *testing.T) {
	report := &GateReport{
		Version:   SchemaVersion,
		Producer:  ProducerName,
		Timestamp: "2026-09-08T12:00:00Z",
		Phase:     "implement",
		Passed:    false,
		Checks: []CheckResult{
			{Name: "spec-exists", Passed: true, Message: "Spec found"},
			{Name: "coverage-strategy", Passed: false, Message: "No coverage strategy found"},
		},
		Summary: Summary{Total: 2, Passed: 1, Failed: 1},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "FAIL") {
		t.Errorf("expected 'FAIL' in output, got: %s", output)
	}
}

// --- FormatText additional tests ---

func TestFormatText_NoBranch(t *testing.T) {
	report := &GateReport{
		Version:   SchemaVersion,
		Producer:  ProducerName,
		Timestamp: "2026-09-08T12:00:00Z",
		Branch:    "",
		Phase:     "specify",
		Passed:    true,
		Checks: []CheckResult{
			{Name: "spec-exists", Passed: true, Message: "Spec found"},
		},
		Summary: Summary{Total: 1, Passed: 1, Failed: 0},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	// Should NOT contain branch info when Branch is empty.
	if strings.Contains(output, "branch:") {
		t.Errorf("expected no branch line when Branch is empty, got: %s", output)
	}
}

func TestFormatText_FailedCheckWithMessage(t *testing.T) {
	report := &GateReport{
		Version:   SchemaVersion,
		Producer:  ProducerName,
		Timestamp: "2026-09-08T12:00:00Z",
		Phase:     "implement",
		Passed:    false,
		Checks: []CheckResult{
			{Name: "spec-exists", Passed: false, Message: "No spec found"},
		},
		Summary: Summary{Total: 1, Passed: 0, Failed: 1},
	}

	var buf bytes.Buffer
	if err := FormatText(report, &buf); err != nil {
		t.Fatalf("FormatText error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No spec found") {
		t.Errorf("expected failure message in output, got: %s", output)
	}
}

// --- gateIndicator tests ---

func TestGateIndicator_NoColor(t *testing.T) {
	renderer := lipgloss.NewRenderer(os.Stderr)
	passStyle := renderer.NewStyle()
	failStyle := renderer.NewStyle()

	passResult := CheckResult{Name: "test", Passed: true}
	failResult := CheckResult{Name: "test", Passed: false}

	passIndicator := gateIndicator(passResult, false, passStyle, failStyle)
	if passIndicator != "[PASS]" {
		t.Errorf("expected [PASS], got %s", passIndicator)
	}

	failIndicator := gateIndicator(failResult, false, passStyle, failStyle)
	if failIndicator != "[FAIL]" {
		t.Errorf("expected [FAIL], got %s", failIndicator)
	}
}

func TestGateIndicator_WithColor(t *testing.T) {
	renderer := lipgloss.NewRenderer(os.Stderr)
	passStyle := renderer.NewStyle().Foreground(lipgloss.Color("10"))
	failStyle := renderer.NewStyle().Foreground(lipgloss.Color("9"))

	passResult := CheckResult{Name: "test", Passed: true}
	failResult := CheckResult{Name: "test", Passed: false}

	passIndicator := gateIndicator(passResult, true, passStyle, failStyle)
	if passIndicator == "[PASS]" {
		t.Error("expected styled output for color mode, got plain [PASS]")
	}

	failIndicator := gateIndicator(failResult, true, passStyle, failStyle)
	if failIndicator == "[FAIL]" {
		t.Error("expected styled output for color mode, got plain [FAIL]")
	}
}

// --- ValidPhases tests ---

func TestValidPhases_ReturnsAllPhases(t *testing.T) {
	phases := ValidPhases()
	if len(phases) != 5 {
		t.Errorf("expected 5 phases, got %d", len(phases))
	}

	expected := []string{"specify", "plan", "implement", "review", "pr"}
	for i, p := range expected {
		if phases[i] != p {
			t.Errorf("expected phase[%d]=%s, got %s", i, p, phases[i])
		}
	}
}

func TestValidPhases_ReturnsCopy(t *testing.T) {
	phases := ValidPhases()
	phases[0] = "mutated"

	fresh := ValidPhases()
	if fresh[0] == "mutated" {
		t.Error("ValidPhases should return a copy, not the original slice")
	}
}

// --- defaults tests ---

func TestDefaults_FillsZeroValues(t *testing.T) {
	opts := Options{
		Phase: "specify",
	}
	opts.defaults()

	if opts.TargetDir == "" {
		t.Error("expected TargetDir to be filled")
	}
	if opts.Format != "text" {
		t.Errorf("expected Format=text, got %s", opts.Format)
	}
	if opts.Stdout == nil {
		t.Error("expected Stdout to be filled")
	}
	if opts.Stderr == nil {
		t.Error("expected Stderr to be filled")
	}
}

func TestDefaults_PreservesSetValues(t *testing.T) {
	var buf bytes.Buffer
	opts := Options{
		TargetDir: "/custom",
		Phase:     "plan",
		Format:    "json",
		Stdout:    &buf,
		Stderr:    &buf,
	}
	opts.defaults()

	if opts.TargetDir != "/custom" {
		t.Errorf("expected /custom, got %s", opts.TargetDir)
	}
	if opts.Format != "json" {
		t.Errorf("expected json, got %s", opts.Format)
	}
	if opts.Stdout != &buf {
		t.Error("expected Stdout to be preserved")
	}
	if opts.Stderr != &buf {
		t.Error("expected Stderr to be preserved")
	}
}

// --- Run with file path that is not a directory ---

func TestRun_PathIsFile(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "somefile.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: filePath,
		Phase:     "specify",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for file path")
	}
	if report != nil {
		t.Errorf("expected nil report for file path")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("expected 'not a directory' error, got: %v", err)
	}
}

// --- Review phase integration ---

func TestRun_ReviewPhase_Pass(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [x] Task 2\n<!-- code-review: passed -->",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "review",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Passed {
		t.Error("expected review phase to pass")
	}
	if report.Summary.Total != 2 {
		t.Errorf("expected 2 checks for review phase, got %d", report.Summary.Total)
	}
}

func TestRun_ReviewPhase_FailIncomplete(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [ ] Task 2\n<!-- code-review: passed -->",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "review",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error for incomplete tasks")
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", report.Summary.Failed)
	}
}

// --- PR phase integration ---

func TestRun_PRPhase_Pass(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir, "opsx/test-feature")
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- code-review: passed -->",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "pr",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Passed {
		t.Error("expected PR phase to pass")
	}
}

// --- findTasksFile tests ---

func TestFindTasksFile_OpenSpec(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks",
	})

	path := findTasksFile(dir)
	if path == "" {
		t.Error("expected to find tasks.md in OpenSpec layout")
	}
}

func TestFindTasksFile_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks",
	})

	path := findTasksFile(dir)
	if path == "" {
		t.Error("expected to find tasks.md in Speckit layout")
	}
}

func TestFindTasksFile_NotFound(t *testing.T) {
	dir := t.TempDir()

	path := findTasksFile(dir)
	if path != "" {
		t.Errorf("expected empty string, got %s", path)
	}
}

func TestRun_PRPhase_FailOnMain(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir, "main")
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- code-review: passed -->",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "pr",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})

	if err == nil {
		t.Fatal("expected error on main branch")
	}
	if report.Summary.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", report.Summary.Failed)
	}
}
