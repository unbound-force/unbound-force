package gate

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// --- helpers ---

// scaffoldOpenSpec creates an OpenSpec artifact tree inside dir
// under the change "test-change".
func scaffoldOpenSpec(t *testing.T, dir string, files map[string]string) {
	scaffoldOpenSpecChange(t, dir, "test-change", files)
}

// scaffoldOpenSpecChange creates an OpenSpec artifact tree for the
// named change inside dir.
func scaffoldOpenSpecChange(t *testing.T, dir, change string, files map[string]string) {
	t.Helper()
	pinGitCeiling(t, dir)
	changeDir := filepath.Join(dir, "openspec", "changes", change)
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

// scaffoldSpeckit creates a Speckit artifact tree inside dir under
// the spec "001-test-feature".
func scaffoldSpeckit(t *testing.T, dir string, files map[string]string) {
	scaffoldSpeckitChange(t, dir, "001-test-feature", files)
}

// scaffoldSpeckitChange creates a Speckit artifact tree for the
// named spec inside dir.
func scaffoldSpeckitChange(t *testing.T, dir, spec string, files map[string]string) {
	t.Helper()
	pinGitCeiling(t, dir)
	specDir := filepath.Join(dir, "specs", spec)
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
// Repo-local config pins the test environment: user identity,
// no GPG signing, and a hooks path pointing at an empty local
// directory, so ambient global/system git config cannot change
// commit behavior.
func initGitRepo(t *testing.T, dir, branch string) {
	t.Helper()
	// Git fixtures spawn several subprocesses per test; skip them
	// under -short so quick iterations run without them (TC-011).
	if testing.Short() {
		t.Skip("skipping git subprocess fixture in short mode")
	}
	runGit(t, dir, "init", "--initial-branch=main")
	runGit(t, dir, "config", "user.email", "test@test.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "config", "commit.gpgsign", "false")
	hooksDir := filepath.Join(dir, ".test-hooks")
	if err := os.MkdirAll(hooksDir, 0o755); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}
	runGit(t, dir, "config", "core.hooksPath", hooksDir)
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

// pinGitCeiling stops git repository discovery from walking above
// dir, so fallback-path fixtures do not depend on ambient repository
// state — a TMPDIR inside a repository on an opsx/* or NNN-* branch
// would otherwise map fixtures to an unrelated change.
func pinGitCeiling(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
}

// testCtx builds a run context for dir, resolving the change the
// way Run does: git-scaffolded fixtures exercise branch-mapped
// resolution, plain fixtures exercise the unmapped fallback.
func testCtx(dir string) runContext {
	return runContext{dir: dir, res: resolveChange(dir)}
}

// --- checkSpecExists tests ---

func TestCheckSpecExists_OpenSpec(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal\nSome content",
	})

	r := checkSpecExists(testCtx(dir))
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

	r := checkSpecExists(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckSpecExists_Missing(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkSpecExists(testCtx(dir))
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

	r := checkPlanExists(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckPlanExists_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"plan.md": "# Plan\nSome content",
	})

	r := checkPlanExists(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckPlanExists_Missing(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkPlanExists(testCtx(dir))
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

	r := checkTasksExist(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckTasksExist_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [ ] Task 1",
	})

	r := checkTasksExist(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckTasksExist_Missing(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkTasksExist(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail, got pass")
	}
}

// --- checkCoverageStrategy tests ---

func TestCheckCoverageStrategy_Heading(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		// Heading-only fixture: no keyword or percentage match,
		// so only heading detection can pass this test.
		"tasks.md": "# Tasks\n\n## Test coverage plan\n\nTests are organized by package.",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for coverage heading, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_Keyword(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		// Keyword-only fixture: no percentage, so only keyword
		// detection can pass this test.
		"tasks.md": "# Tasks\n\ncoverage target: high\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for coverage keyword, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_UnitTestKeyword(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\nWe need a unit test for each function.\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for 'unit test' keyword, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_Missing(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\n- [ ] Implement feature\n- [ ] Write docs\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail for missing coverage strategy, got pass")
	}
}

func TestCheckCoverageStrategy_NoFiles(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail when no task/plan files exist, got pass")
	}
}

func TestCheckCoverageStrategy_LongLine(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		// A single line longer than bufio.Scanner's 64KiB default
		// token limit must be ordinary check input, not an internal
		// error: the strategy appearing after it is still detected.
		"tasks.md": "# Tasks\n" + strings.Repeat("x", 128*1024) + "\n## Coverage Strategy\nUnit tests.\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass with coverage strategy after a long line, got fail: %s", r.Message)
	}
}

// --- checkTasksComplete tests ---

func TestCheckTasksComplete_AllDone(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [x] Task 2\n",
	})

	r := checkTasksComplete(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass, got fail: %s", r.Message)
	}
}

func TestCheckTasksComplete_Incomplete(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n- [ ] Task 2\n",
	})

	r := checkTasksComplete(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail for incomplete tasks, got pass")
	}
}

func TestCheckTasksComplete_ProseLiteral(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Complete task\n\nUse `- [ ]` for incomplete tasks.\n",
	})

	r := checkTasksComplete(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected prose literal to be ignored, got fail: %s", r.Message)
	}
}

func TestCheckTasksComplete_IndentedSubtask(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] 1.1 Done\n  - [ ] 1.1.1 Nested incomplete\n",
	})

	r := checkTasksComplete(testCtx(dir))
	if r.Passed {
		t.Errorf("expected indented unchecked subtask to fail, got pass")
	}
}

func TestCheckTasksComplete_NoFile(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkTasksComplete(testCtx(dir))
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

	r := checkReviewRun(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for spec-review marker, got fail: %s", r.Message)
	}
}

func TestCheckReviewRun_CodeReview(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- code-review: passed -->",
	})

	r := checkReviewRun(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for code-review marker, got fail: %s", r.Message)
	}
}

func TestCheckReviewRun_NoMarker(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n",
	})

	r := checkReviewRun(testCtx(dir))
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

	r := checkReviewPass(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for code-review marker, got fail: %s", r.Message)
	}
}

func TestCheckReviewPass_OnlySpecReview(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- spec-review: passed -->",
	})

	r := checkReviewPass(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail when only spec-review marker present, got pass")
	}
}

func TestCheckReviewPass_NoMarker(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n",
	})

	r := checkReviewPass(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail when no marker, got pass")
	}
}

func TestCheckReviewPass_NoFile(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkReviewPass(testCtx(dir))
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

	r := checkReviewRun(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for speckit layout, got fail: %s", r.Message)
	}
}

func TestCheckReviewRun_NoFile(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkReviewRun(testCtx(dir))
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

	r := checkTasksComplete(testCtx(dir))
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

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for speckit plan, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_PercentagePattern(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\nTarget: coverage >= 90%\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass for percentage pattern, got fail: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_ActiveChangeMissingCoverage_Fails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "aaa-other", map[string]string{
		"tasks.md": "# Tasks\n\n## Coverage Strategy\n\nUnit tests required.\n",
	})
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"tasks.md": "# Tasks\n\n- [ ] Implement feature\n",
	})
	initGitRepo(t, dir, "opsx/zzz-active")

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected active change without coverage strategy to fail, got: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_SpeckitBranchChange_MissingCoverage_Fails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// Another Speckit feature carries the coverage strategy; the
	// branch-mapped one does not. Strict resolution must not let
	// 001-other satisfy 002-active's check.
	scaffoldSpeckitChange(t, dir, "001-other", map[string]string{
		"plan.md": "# Plan\n\n## Coverage Strategy\n\nUnit tests.\n",
	})
	scaffoldSpeckitChange(t, dir, "002-active", map[string]string{
		"plan.md": "# Plan\n",
	})
	initGitRepo(t, dir, "002-active")

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected active Speckit change without coverage strategy to fail, got: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_SpeckitBranchChange_Passes(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldSpeckitChange(t, dir, "001-other", map[string]string{
		"plan.md": "# Plan\n",
	})
	scaffoldSpeckitChange(t, dir, "002-active", map[string]string{
		"plan.md": "# Plan\n\n## Coverage Strategy\n\nIntegration tests.\n",
	})
	initGitRepo(t, dir, "002-active")

	r := checkCoverageStrategy(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected coverage strategy from branch-mapped Speckit change, got: %s", r.Message)
	}
}

// --- quick-branch strict resolution tests ---

func TestCheckSpecExists_QuickBranch_Fails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})
	initGitRepo(t, dir, "quick/some-change")

	r := checkSpecExists(testCtx(dir))
	if r.Passed {
		t.Errorf("expected quick branch to fail spec-exists, got: %s", r.Message)
	}
	if !strings.Contains(r.Message, "quick") {
		t.Errorf("expected quick-branch explanation in message, got: %s", r.Message)
	}
}

func TestCheckCoverageStrategy_QuickBranch_Fails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\n## Coverage Strategy\n\nUnit tests required.\n",
	})
	initGitRepo(t, dir, "quick/some-change")

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected quick branch to fail coverage-strategy, got: %s", r.Message)
	}
	if !strings.Contains(r.Message, "quick") {
		t.Errorf("expected quick-branch explanation in message, got: %s", r.Message)
	}
}

// TestCheckTasksComplete_QuickBranch pins the quick-branch
// explanation for the tasks-consuming checks: the failure names
// the tier limitation, not a missing artifact.
func TestCheckTasksComplete_QuickBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] done\n",
	})
	initGitRepo(t, dir, "quick/some-change")

	r := checkTasksComplete(testCtx(dir))
	if r.Passed {
		t.Fatalf("expected quick branch to fail tasks-complete, got: %s", r.Message)
	}
	if !strings.Contains(r.Message, "quick") {
		t.Errorf("expected quick-branch explanation in message, got: %s", r.Message)
	}
}

func TestFindTasksFile_QuickBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks",
	})
	initGitRepo(t, dir, "quick/some-change")

	path, err := findTasksFile(testCtx(dir))
	if err != nil {
		t.Fatalf("find tasks file: %v", err)
	}
	if path != "" {
		t.Errorf("expected no tasks file on quick branch, got %s", path)
	}
}

// --- checkNotOnMain tests ---

func TestCheckNotOnMain_FeatureBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	initGitRepo(t, dir, "opsx/test-feature")

	r := checkNotOnMain(testCtx(dir))
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

	r := checkNotOnMain(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail on main branch, got pass")
	}
}

func TestCheckNotOnMain_NonGitDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	pinGitCeiling(t, dir)

	r := checkNotOnMain(testCtx(dir))
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
	// Detach HEAD. Pin the CI-ref recovery env so an ambient
	// GITHUB_REF_NAME (for example "main" on a push event) cannot
	// flip the resolution on CI runners.
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "")
	t.Setenv("GITHUB_REF_NAME", "")

	r := checkNotOnMain(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected pass on detached HEAD, got fail: %s", r.Message)
	}
}

func TestCheckNotOnMain_GitUnavailable(t *testing.T) {
	dir := t.TempDir()
	// Empty PATH so exec.LookPath("git") fails (GATE-FR-008
	// "Git not available" scenario).
	t.Setenv("PATH", t.TempDir())

	r := checkNotOnMain(testCtx(dir))
	if r.Passed {
		t.Error("expected fail when git is unavailable")
	}
	if !strings.Contains(r.Message, "uf doctor") {
		t.Errorf("expected 'uf doctor' hint in message, got: %s", r.Message)
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
	pinGitCeiling(t, dir)
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

func TestRun_ImplementPhase_MultiChange_ResolvesBranchChange(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// aaa-other is complete; the branch-mapped zzz-active is not.
	// Strict resolution must fail the run on zzz-active's gaps
	// without letting aaa-other's artifacts satisfy any check.
	scaffoldOpenSpecChange(t, dir, "aaa-other", map[string]string{
		"proposal.md": "# Proposal",
		"design.md":   "# Design",
		"tasks.md":    "# Tasks\n\n## Coverage Strategy\n\nUnit tests.\n",
	})
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"proposal.md": "# Proposal",
	})
	initGitRepo(t, dir, "opsx/zzz-active")

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "implement",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err == nil {
		t.Fatal("expected error for incomplete active change")
	}
	if report == nil {
		t.Fatal("expected report on check failure")
	}
	if report.Passed {
		t.Error("expected report.Passed=false")
	}
	if report.Branch != "opsx/zzz-active" {
		t.Errorf("expected provenance branch opsx/zzz-active, got %q", report.Branch)
	}
	if report.Summary != (Summary{Total: 4, Passed: 1, Failed: 3}) {
		t.Errorf("unexpected summary: %+v", report.Summary)
	}
	for _, c := range report.Checks {
		if strings.Contains(c.Message, "aaa-other") {
			t.Errorf("check %s resolved another change's artifacts: %s", c.Name, c.Message)
		}
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
	base := t.TempDir()
	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: filepath.Join(base, "nonexistent"),
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
	// GATE-FR-010: the error names the missing directory.
	if !strings.Contains(err.Error(), "directory does not exist") {
		t.Errorf("expected 'directory does not exist' message, got: %v", err)
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
	if _, perr := time.Parse(time.RFC3339, report.Timestamp); perr != nil {
		t.Errorf("timestamp %q is not RFC3339: %v", report.Timestamp, perr)
	}
	if report.Phase != "specify" {
		t.Errorf("expected phase specify, got %s", report.Phase)
	}
	if report.Branch != "" {
		t.Errorf("expected empty branch for non-git dir, got %s", report.Branch)
	}
	if report.Commit != "" {
		t.Errorf("expected empty commit for non-git dir, got %s", report.Commit)
	}
	if report.ToolVersion != "dev" {
		t.Errorf("expected default tool version dev, got %s", report.ToolVersion)
	}
}

// --- FormatJSON tests ---

func TestFormatJSON_ValidOutput(t *testing.T) {
	report := &GateReport{
		Version:     SchemaVersion,
		ToolVersion: "v1.2.3",
		Producer:    ProducerName,
		Timestamp:   "2026-09-08T12:00:00Z",
		Branch:      "opsx/test",
		Commit:      "0123456789abcdef",
		Phase:       "specify",
		Passed:      true,
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
	if decoded.Producer != ProducerName {
		t.Errorf("expected producer %s, got %s", ProducerName, decoded.Producer)
	}
	if decoded.ToolVersion != "v1.2.3" {
		t.Errorf("expected tool version v1.2.3, got %s", decoded.ToolVersion)
	}
	if decoded.Branch != "opsx/test" {
		t.Errorf("expected branch opsx/test, got %s", decoded.Branch)
	}
	if decoded.Commit != "0123456789abcdef" {
		t.Errorf("expected commit provenance, got %s", decoded.Commit)
	}
	if _, err := time.Parse(time.RFC3339, decoded.Timestamp); err != nil {
		t.Errorf("expected RFC3339 timestamp %q, parse error: %v", decoded.Timestamp, err)
	}
	if decoded.Summary != (Summary{Total: 1, Passed: 1, Failed: 0}) {
		t.Errorf("unexpected summary: %+v", decoded.Summary)
	}
	if len(decoded.Checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(decoded.Checks))
	}
	if decoded.Checks[0].Name != "spec-exists" || !decoded.Checks[0].Passed || decoded.Checks[0].Message != "Spec found" {
		t.Errorf("unexpected check: %+v", decoded.Checks[0])
	}
}

// TestFormatJSON_WireContractKeys asserts the raw JSON wire keys of
// the gate report. Struct round-trips (TestFormatJSON_ValidOutput)
// cannot detect json-tag drift — both sides use the same struct — so
// this test pins the machine-readable contract GATE-FR-003 declares
// stable across minor versions.
func TestFormatJSON_WireContractKeys(t *testing.T) {
	report := &GateReport{
		Version:     SchemaVersion,
		ToolVersion: "v1.2.3",
		Producer:    ProducerName,
		Timestamp:   "2026-09-08T12:00:00Z",
		Branch:      "opsx/test",
		Commit:      "0123456789abcdef",
		Phase:       "specify",
		Passed:      true,
		Checks: []CheckResult{
			{Name: "spec-exists", Passed: true, Message: "Spec found"},
		},
		Summary: Summary{Total: 1, Passed: 1, Failed: 0},
	}

	var buf bytes.Buffer
	if err := FormatJSON(report, &buf); err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	wantKeys := []string{
		"version", "tool_version", "producer", "timestamp",
		"branch", "commit", "phase", "passed", "checks", "summary",
	}
	for _, key := range wantKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("wire output missing key %q: %s", key, buf.String())
		}
	}
	if len(raw) != len(wantKeys) {
		t.Errorf("expected exactly %d top-level keys, got %d: %s",
			len(wantKeys), len(raw), buf.String())
	}

	var checks []map[string]json.RawMessage
	if err := json.Unmarshal(raw["checks"], &checks); err != nil {
		t.Fatalf("invalid checks array: %v", err)
	}
	for _, key := range []string{"name", "passed", "message"} {
		if _, ok := checks[0][key]; !ok {
			t.Errorf("checks[0] missing key %q: %s", key, buf.String())
		}
	}
	if len(checks[0]) != 3 {
		t.Errorf("expected exactly 3 keys in checks[0], got %d: %s",
			len(checks[0]), buf.String())
	}

	var summary map[string]json.RawMessage
	if err := json.Unmarshal(raw["summary"], &summary); err != nil {
		t.Fatalf("invalid summary object: %v", err)
	}
	for _, key := range []string{"total", "passed", "failed"} {
		if _, ok := summary[key]; !ok {
			t.Errorf("summary missing key %q: %s", key, buf.String())
		}
	}
	if len(summary) != 3 {
		t.Errorf("expected exactly 3 keys in summary, got %d: %s",
			len(summary), buf.String())
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

func TestFormatText_ProvenanceLine(t *testing.T) {
	report := &GateReport{
		Version:     SchemaVersion,
		ToolVersion: "v1.2.3",
		Producer:    ProducerName,
		Timestamp:   "2026-09-08T12:00:00Z",
		Branch:      "opsx/test",
		Commit:      "0123456789abcdef",
		Phase:       "specify",
		Passed:      true,
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
	for _, want := range []string{"branch: opsx/test", "commit: 0123456789abcdef", ProducerName, "v1.2.3"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in provenance output, got: %s", want, output)
		}
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

func TestValidPhases_SyncWithPhaseChecks(t *testing.T) {
	// Verify every valid phase has a corresponding entry in
	// phaseChecks. This is also enforced by the init() guard
	// in checks.go, but a test documents the expectation and
	// catches map-only additions that skip validPhases.
	for _, phase := range validPhases {
		if _, ok := phaseChecks[phase]; !ok {
			t.Errorf("validPhases contains %q but phaseChecks has no entry", phase)
		}
	}
	for phase := range phaseChecks {
		found := false
		for _, p := range validPhases {
			if p == phase {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("phaseChecks contains %q but validPhases does not", phase)
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
	// Branch must map to the scaffolded change (opsx/<change>) so
	// branch-aware artifact resolution selects it.
	initGitRepo(t, dir, "opsx/test-change")
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
	if report.Branch != "opsx/test-change" {
		t.Errorf("expected branch opsx/test-change, got %s", report.Branch)
	}
	if report.Commit == "" {
		t.Error("expected commit provenance for git repository")
	}
	if report.ToolVersion != "dev" {
		t.Errorf("expected default tool version dev, got %s", report.ToolVersion)
	}
	if report.Summary.Total != 2 || report.Summary.Failed != 0 {
		t.Errorf("unexpected summary: %+v", report.Summary)
	}
	for _, want := range []string{"review-pass", "not-on-main"} {
		found := false
		for _, c := range report.Checks {
			if c.Name == want && c.Passed {
				found = true
			}
		}
		if !found {
			t.Errorf("expected passing check %q, got: %+v", want, report.Checks)
		}
	}
}

// TestRun_QuickBranch_FailsArtifactChecks covers the quick-branch
// resolution mode through Run() itself (GATE-FR-012): the quick tier
// carries no spec artifacts, so an unrelated change's artifacts must
// not satisfy its checks, and the report provenance must name the
// quick branch.
func TestRun_QuickBranch_FailsArtifactChecks(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"proposal.md": "# Proposal",
	})
	initGitRepo(t, dir, "quick/some-change")
	t.Setenv("GITHUB_HEAD_REF", "")
	t.Setenv("GITHUB_REF_NAME", "")

	report, err := Run(Options{TargetDir: dir, Phase: "specify"})
	if err == nil {
		t.Fatal("expected error for failing checks")
	}
	if report == nil {
		t.Fatal("expected report even on failure")
	}
	if report.Passed {
		t.Error("expected report.Passed=false on quick branch")
	}
	if report.Branch != "quick/some-change" {
		t.Errorf("expected branch provenance quick/some-change, got %q", report.Branch)
	}
	specExists := false
	for _, c := range report.Checks {
		if c.Name == "spec-exists" {
			specExists = !c.Passed
		}
	}
	if !specExists {
		t.Errorf("expected spec-exists to fail on quick branch, got: %+v", report.Checks)
	}
}

// TestRun_DetachedHead_RecoversCIRef covers the detached-HEAD
// CI-recovery resolution mode through Run() itself (GATE-FR-012):
// the branch recovered from GITHUB_HEAD_REF drives both artifact
// resolution and the report's branch provenance.
func TestRun_DetachedHead_RecoversCIRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	// aaa-first lacks the artifact; zzz-active has it. Passing
	// proves the run resolved the CI-named change, not the
	// alphabetical first.
	scaffoldOpenSpecChange(t, dir, "aaa-first", map[string]string{
		"notes.md": "# Notes",
	})
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"proposal.md": "# Proposal",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "opsx/zzz-active")
	t.Setenv("GITHUB_REF_NAME", "123/merge")

	report, err := Run(Options{TargetDir: dir, Phase: "specify"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.Passed {
		t.Errorf("expected pass via CI-recovered change, got: %+v", report.Checks)
	}
	if report.Branch != "opsx/zzz-active" {
		t.Errorf("expected CI-recovered branch provenance opsx/zzz-active, got %q", report.Branch)
	}
}

// TestRun_PRPhase_DetachedHead_CIRefMain_Fails pins the branch
// protection semantics in detached CI contexts: when the recovered
// CI ref is main, the not-on-main check must fail (previously the
// check re-ran git itself and only ever saw the literal "HEAD").
func TestRun_PRPhase_DetachedHead_CIRefMain_Fails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] Task 1\n<!-- code-review: passed -->",
	})
	initGitRepo(t, dir, "opsx/test-change")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "main")
	t.Setenv("GITHUB_REF_NAME", "")

	report, err := Run(Options{TargetDir: dir, Phase: "pr"})
	if err == nil {
		t.Fatal("expected error for not-on-main failure")
	}
	if report == nil {
		t.Fatal("expected report even on failure")
	}
	notOnMain := false
	for _, c := range report.Checks {
		if c.Name == "not-on-main" {
			notOnMain = !c.Passed
		}
	}
	if !notOnMain {
		t.Errorf("expected not-on-main to fail on recovered main ref, got: %+v", report.Checks)
	}
	if report.Branch != "main" {
		t.Errorf("expected branch provenance main, got %q", report.Branch)
	}
}

// --- findTasksFile tests ---

func TestFindTasksFile_OpenSpec(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks",
	})

	path, err := findTasksFile(testCtx(dir))
	if err != nil {
		t.Fatalf("find tasks file: %v", err)
	}
	if path == "" {
		t.Error("expected to find tasks.md in OpenSpec layout")
	}
}

func TestFindTasksFile_Speckit(t *testing.T) {
	dir := t.TempDir()
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks",
	})

	path, err := findTasksFile(testCtx(dir))
	if err != nil {
		t.Fatalf("find tasks file: %v", err)
	}
	if path == "" {
		t.Error("expected to find tasks.md in Speckit layout")
	}
}

func TestFindTasksFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)

	path, err := findTasksFile(testCtx(dir))
	if err != nil {
		t.Fatalf("find tasks file: %v", err)
	}
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

// --- symlinked target dir (GATE-FR-010) ---

func TestRun_SymlinkedTargetDir(t *testing.T) {
	real := t.TempDir()
	scaffoldOpenSpec(t, real, map[string]string{
		"proposal.md": "# Proposal",
	})
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink not supported: %v", err)
	}

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: link,
		Phase:     "specify",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err != nil {
		t.Fatalf("unexpected error through symlink: %v", err)
	}
	if !report.Passed {
		t.Errorf("expected pass through symlinked dir, got: %+v", report.Checks)
	}
}

// --- branch-aware multi-change resolution ---

func TestFindTasksFile_MultipleChanges_SelectsBranchChange(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// Two concurrent changes. The alphabetically first is NOT the
	// active one; the branch must drive selection.
	scaffoldOpenSpecChange(t, dir, "aaa-first", map[string]string{
		"tasks.md": "# Tasks\n- [x] a\n",
	})
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"tasks.md": "# Tasks\n- [x] z\n",
	})
	initGitRepo(t, dir, "opsx/zzz-active")

	path, err := findTasksFile(testCtx(dir))
	if err != nil {
		t.Fatalf("find tasks file: %v", err)
	}
	want := filepath.Join(dir, "openspec", "changes", "zzz-active", "tasks.md")
	if path != want {
		t.Errorf("expected branch-mapped tasks file %s, got %s", want, path)
	}
}

func TestCheckSpecExists_ActiveChangeMissing_Fails(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// An unrelated change has the artifact; the active change does
	// not. Strict branch-aware resolution must not fall back to the
	// unrelated change (false-pass prevention).
	scaffoldOpenSpecChange(t, dir, "aaa-other", map[string]string{
		"proposal.md": "# Proposal",
	})
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"tasks.md": "# Tasks\n",
	})
	initGitRepo(t, dir, "opsx/zzz-active")

	r := checkSpecExists(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail: active change has no proposal.md; got pass: %s", r.Message)
	}
}

func TestFindTasksFile_SpeckitBranchChange(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// Speckit layout: the change directory is the full branch name.
	scaffoldSpeckit(t, dir, map[string]string{
		"tasks.md": "# Tasks\n",
	})
	initGitRepo(t, dir, "001-test-feature")

	path, err := findTasksFile(testCtx(dir))
	if err != nil {
		t.Fatalf("find tasks file: %v", err)
	}
	want := filepath.Join(dir, "specs", "001-test-feature", "tasks.md")
	if path != want {
		t.Errorf("expected branch-mapped tasks file %s, got %s", want, path)
	}
}

// --- read-error branches ---

func TestTaskChecks_UnreadableTasksFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; permission bits do not apply")
	}
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] done\n<!-- code-review: passed -->",
	})
	tasksPath := filepath.Join(dir, "openspec", "changes", "test-change", "tasks.md")
	if err := os.Chmod(tasksPath, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	for _, tc := range []struct {
		name string
		fn   func(runContext) CheckResult
	}{
		{"tasks-complete", checkTasksComplete},
		{"review-run", checkReviewRun},
		{"review-pass", checkReviewPass},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.fn(testCtx(dir))
			if r.Passed {
				t.Errorf("expected fail for unreadable tasks.md")
			}
			if r.internalErr == nil {
				t.Errorf("expected internal error, got: %s", r.Message)
			}
		})
	}
}

func TestCheckCoverageStrategy_UnreadableArtifact(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; permission bits do not apply")
	}
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\n## Coverage Strategy\n\n>= 80%\n",
	})
	tasksPath := filepath.Join(dir, "openspec", "changes", "test-change", "tasks.md")
	if err := os.Chmod(tasksPath, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail for unreadable artifact")
	}
	if r.internalErr == nil {
		t.Errorf("expected internal error, got: %s", r.Message)
	}
}

// --- InternalError contract ---

func TestInternalError_ErrorUnwrapExitCode(t *testing.T) {
	inner := errors.New("boom")
	e := &InternalError{Err: inner}
	if e.Error() != "boom" {
		t.Errorf("expected 'boom', got %q", e.Error())
	}
	if e.Unwrap() != inner {
		t.Errorf("Unwrap did not return the wrapped error")
	}
	if e.ExitCode() != 2 {
		t.Errorf("expected exit code 2, got %d", e.ExitCode())
	}
}

// --- strict-mode missing branches ---

func TestFindTasksFile_ActiveChangeMissing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// The branch maps to a change that has no tasks.md. No fallback
	// to other changes is allowed.
	scaffoldOpenSpecChange(t, dir, "other-change", map[string]string{
		"tasks.md": "# Tasks\n",
	})
	initGitRepo(t, dir, "opsx/zzz-missing")

	if path, err := findTasksFile(testCtx(dir)); err != nil {
		t.Fatalf("find tasks file: %v", err)
	} else if path != "" {
		t.Errorf("expected no tasks file for missing active change, got %s", path)
	}
}

func TestCheckSpecExists_SpeckitBranch_Missing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	// A different Speckit feature has the artifact; the active one
	// does not.
	scaffoldSpeckitChange(t, dir, "001-other", map[string]string{
		"spec.md": "# Spec",
	})
	initGitRepo(t, dir, "002-active")

	r := checkSpecExists(testCtx(dir))
	if r.Passed {
		t.Errorf("expected fail for active Speckit branch without spec.md; got pass: %s", r.Message)
	}
}

// --- detached-HEAD CI branch recovery ---

func TestResolveChange_DetachedHead_RecoversCIRef(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "aaa-other", map[string]string{
		"tasks.md": "# Tasks\n",
	})
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"tasks.md": "# Tasks\n\n## Coverage Strategy\n\nUnit tests.\n",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "opsx/zzz-active")
	t.Setenv("GITHUB_REF_NAME", "")

	res := resolveChange(dir)
	if !res.Strict || res.Name != "zzz-active" {
		t.Errorf("expected strict resolution to zzz-active via CI ref, got %+v", res)
	}
	if res.Branch != "opsx/zzz-active" {
		t.Errorf("expected effective branch opsx/zzz-active, got %q", res.Branch)
	}

	// The coverage check must read the CI-named change, not the
	// alphabetical first.
	r := checkCoverageStrategy(runContext{dir: dir, res: res})
	if !r.Passed {
		t.Errorf("expected coverage strategy from CI-resolved change, got: %s", r.Message)
	}
}

func TestResolveChange_DetachedHead_RefNameFallback(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"tasks.md": "# Tasks\n",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "")
	t.Setenv("GITHUB_REF_NAME", "opsx/zzz-active")

	res := resolveChange(dir)
	if !res.Strict || res.Name != "zzz-active" {
		t.Errorf("expected strict resolution to zzz-active via GITHUB_REF_NAME, got %+v", res)
	}
}

func TestResolveChange_DetachedHead_NoCIRef_FallsBack(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "aaa-first", map[string]string{
		"tasks.md": "# Tasks\n",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "")
	t.Setenv("GITHUB_REF_NAME", "")

	res := resolveChange(dir)
	if res.Strict || res.Quick || res.Name != "aaa-first" {
		t.Errorf("expected unmapped fallback to resolve aaa-first without CI refs, got %+v", res)
	}
	if res.Branch != "HEAD" {
		t.Errorf("expected literal HEAD branch without CI refs, got %q", res.Branch)
	}
	if path, err := findTasksFile(runContext{dir: dir, res: res}); err != nil {
		t.Fatalf("find tasks file: %v", err)
	} else if !strings.HasSuffix(path, filepath.Join("aaa-first", "tasks.md")) {
		t.Errorf("expected alphabetical fallback to aaa-first, got %s", path)
	}
}

// TestResolveChange_DetachedHead_HeadRefPrecedence covers the
// combination GATE-FR-012's ordering exists for: both CI variables
// populated with different values, as on a pull_request event where
// GITHUB_REF_NAME is the merge ref. GITHUB_HEAD_REF must win.
func TestResolveChange_DetachedHead_HeadRefPrecedence(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "zzz-active", map[string]string{
		"proposal.md": "# Proposal",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "opsx/zzz-active")
	t.Setenv("GITHUB_REF_NAME", "123/merge")

	res := resolveChange(dir)
	if !res.Strict || res.Name != "zzz-active" {
		t.Errorf("expected GITHUB_HEAD_REF to win over the merge ref, got %+v", res)
	}
	if res.Branch != "opsx/zzz-active" {
		t.Errorf("expected effective branch opsx/zzz-active, got %q", res.Branch)
	}
}

// TestResolveChange_DetachedHead_MergeRefFallsBack pins that a
// merge-style GITHUB_REF_NAME (no CI head ref) names no change and
// maps to the unmapped fallback, as documented in ciBranchRef.
func TestResolveChange_DetachedHead_MergeRefFallsBack(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "aaa-first", map[string]string{
		"tasks.md": "# Tasks",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")
	t.Setenv("GITHUB_HEAD_REF", "")
	t.Setenv("GITHUB_REF_NAME", "123/merge")

	res := resolveChange(dir)
	if res.Strict || res.Quick || res.Name != "aaa-first" {
		t.Errorf("expected unmapped fallback to resolve aaa-first for merge ref, got %+v", res)
	}
	if res.Branch != "123/merge" {
		t.Errorf("expected effective branch 123/merge, got %q", res.Branch)
	}
}

// TestResolveChange_DetachedHead_UnsafeCIRefRejected verifies the
// Constitution V input validation on CI-provided refs: traversal,
// absolute-path, and control-character values are ignored rather
// than flowing into artifact path construction or messages.
func TestResolveChange_DetachedHead_UnsafeCIRefRejected(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	scaffoldOpenSpecChange(t, dir, "aaa-first", map[string]string{
		"tasks.md": "# Tasks",
	})
	initGitRepo(t, dir, "main")
	runGit(t, dir, "checkout", "--detach")

	for _, ref := range []string{
		"../../evil", "/abs/path", `opsx\evil`, "opsx//evil", "opsx/./evil",
		"opsx/x\nPASS: 2 of 2 checks passed", "opsx/x\ttab", "opsx/x\x07bel",
		"opsx/x\x7fdel", "opsx/x\u0085nb",
	} {
		t.Setenv("GITHUB_HEAD_REF", ref)
		t.Setenv("GITHUB_REF_NAME", "")
		res := resolveChange(dir)
		if res.Branch != "HEAD" || res.Strict || res.Name != "aaa-first" {
			t.Errorf("expected unsafe CI ref %q to be ignored (safe alphabetical fallback), got %+v", ref, res)
		}
	}
}

// --- unmapped fallback resolution tests (GATE-FR-012) ---

// TestCheckSpecExists_UnmappedFallback_SelectsFirstChange pins
// that the unmapped fallback resolves one first-alphabetical
// change for the whole run: a later change's artifact cannot
// satisfy the phase.
func TestCheckSpecExists_UnmappedFallback_SelectsFirstChange(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)
	scaffoldOpenSpecChange(t, dir, "aaa-first", map[string]string{
		"tasks.md": "# Tasks",
	})
	scaffoldOpenSpecChange(t, dir, "bbb-second", map[string]string{
		"proposal.md": "# Proposal",
	})

	r := checkSpecExists(testCtx(dir))
	if r.Passed {
		t.Fatalf("expected fail: fallback must resolve only aaa-first, got pass")
	}
	if !strings.Contains(r.Message, "aaa-first") {
		t.Errorf("expected message to name aaa-first, got: %s", r.Message)
	}
	if strings.Contains(r.Message, "bbb-second") {
		t.Errorf("expected no reference to bbb-second, got: %s", r.Message)
	}
}

// TestCheckSpecExists_UnmappedFallback_PrefersOpenSpecLayout pins
// GATE-FR-012's layout-precedence scenario: the fallback probes
// openspec/changes/ before specs/, so an OpenSpec change wins even
// when a Speckit change directory sorts earlier.
func TestCheckSpecExists_UnmappedFallback_PrefersOpenSpecLayout(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)
	scaffoldOpenSpecChange(t, dir, "zzz-change", map[string]string{
		"proposal.md": "# Proposal",
	})
	scaffoldSpeckitChange(t, dir, "001-feature", map[string]string{
		"spec.md": "# Spec",
	})

	r := checkSpecExists(testCtx(dir))
	if !r.Passed {
		t.Fatalf("expected pass against zzz-change, got: %s", r.Message)
	}
	if !strings.Contains(r.Message, "zzz-change") {
		t.Errorf("expected message to name zzz-change, got: %s", r.Message)
	}
	if strings.Contains(r.Message, "001-feature") {
		t.Errorf("expected no reference to 001-feature, got: %s", r.Message)
	}
}

// TestRun_ImplementPhase_UnmappedFallback_DoesNotCombineChanges
// pins GATE-FR-012's no-union scenario: the implement phase must
// not pass by combining spec/tasks from one change with
// plan/coverage from another.
func TestRun_ImplementPhase_UnmappedFallback_DoesNotCombineChanges(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)
	scaffoldOpenSpecChange(t, dir, "aaa-nostrat", map[string]string{
		"proposal.md": "# Proposal",
		"tasks.md":    "# Tasks\n- [x] done\n",
	})
	scaffoldOpenSpecChange(t, dir, "bbb-cov", map[string]string{
		"design.md": "# Design\n\ncoverage target: >= 80%\n",
	})

	var stdout, stderr bytes.Buffer
	report, err := Run(Options{
		TargetDir: dir,
		Phase:     "implement",
		Format:    "json",
		Stdout:    &stdout,
		Stderr:    &stderr,
	})
	if err == nil {
		t.Fatal("expected implement failure: artifacts of two changes must not combine")
	}
	if report == nil || report.Passed {
		t.Fatalf("expected failed report, got %+v", report)
	}
	for _, r := range report.Checks {
		if (r.Name == "plan-exists" || r.Name == "coverage-strategy") && r.Passed {
			t.Errorf("%s must fail: bbb-cov artifacts must not satisfy aaa-nostrat", r.Name)
		}
		if strings.Contains(r.Message, "bbb-cov") {
			t.Errorf("no check message may reference bbb-cov, got: %s", r.Message)
		}
	}
}

// TestCheckSpecExists_UnmappedFallback_FollowsSymlinkedChangeDir
// pins that symlinked change directories resolve in the fallback,
// matching strict resolution (which stats through symlinks).
func TestCheckSpecExists_UnmappedFallback_FollowsSymlinkedChangeDir(t *testing.T) {
	dir := t.TempDir()
	pinGitCeiling(t, dir)
	target := filepath.Join(dir, "elsewhere", "real-change")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(target, "proposal.md"), []byte("# Proposal"), 0o644); err != nil {
		t.Fatalf("write proposal: %v", err)
	}
	changes := filepath.Join(dir, "openspec", "changes")
	if err := os.MkdirAll(changes, 0o755); err != nil {
		t.Fatalf("mkdir changes: %v", err)
	}
	if err := os.Symlink(filepath.Join("..", "..", "elsewhere", "real-change"), filepath.Join(changes, "aaa-link")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	r := checkSpecExists(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected symlinked change dir to resolve in fallback, got fail: %s", r.Message)
	}
}

// TestCheckCoverageStrategy_NegatedHeading pins GATE-FR-006's
// negation exclusion: a heading that denies a coverage strategy
// must not satisfy the Constitution IV check.
func TestCheckCoverageStrategy_NegatedHeading(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"design.md": "# Design\n\n## No coverage strategy is planned\n\nnone\n",
		"tasks.md":  "# Tasks\n- [x] done\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Errorf("expected negated coverage heading to fail, got pass")
	}
}

// TestCheckTasksComplete_FencedExample pins GATE-FR-007's fenced
// code block exclusion: an example checkbox inside a fence is
// documentation, not an incomplete task.
func TestCheckTasksComplete_FencedExample(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n- [x] done\n\nExample syntax:\n\n```\n- [ ] like this\n```\n",
	})

	r := checkTasksComplete(testCtx(dir))
	if !r.Passed {
		t.Errorf("expected fenced example to be ignored, got fail: %s", r.Message)
	}
}

// TestCheckTasksComplete_MixedMarkerFence pins GATE-FR-007's fence
// matching: a fence closes only on its own marker character, so a
// tilde line inside a backtick fence is content and cannot mask
// task lines after the true closer.
func TestCheckTasksComplete_MixedMarkerFence(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"tasks.md": "# Tasks\n\n```\n~~~\n```\n\n- [ ] still open\n",
	})

	r := checkTasksComplete(testCtx(dir))
	if r.Passed {
		t.Error("expected task after a mixed-marker fence to fail, got pass")
	}
}

// TestFenceLine pins the CommonMark delimiter classification the
// scanners rely on: markers must repeat at least three times, a
// closing fence carries no info string, and a backtick fence's
// info string must not contain a backtick.
func TestFenceLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		marker  byte
		length  int
		closing bool
	}{
		{name: "opening with info", line: "```go", marker: '`', length: 3, closing: false},
		{name: "plain backticks", line: "```", marker: '`', length: 3, closing: true},
		{name: "indented longer tildes", line: "  ~~~~  ", marker: '~', length: 4, closing: true},
		{name: "tilde info string", line: "~~~ruby", marker: '~', length: 3, closing: false},
		{name: "two markers only", line: "``", marker: 0, length: 0, closing: false},
		{name: "task line", line: "- [ ] task", marker: 0, length: 0, closing: false},
		{name: "backtick info with backticks", line: "```a`b", marker: 0, length: 0, closing: false},
	}
	for _, tt := range tests {
		marker, length, closing := fenceLine(tt.line)
		if marker != tt.marker || length != tt.length || closing != tt.closing {
			t.Errorf("%s: fenceLine(%q) = (%d, %d, %v), want (%d, %d, %v)",
				tt.name, tt.line, marker, length, closing, tt.marker, tt.length, tt.closing)
		}
	}
}

// TestCheckCoverageStrategy_FencedExample pins GATE-FR-006's fenced
// code block exclusion: coverage syntax documented inside a fence
// is not a strategy.
func TestCheckCoverageStrategy_FencedExample(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"design.md": "# Design\n\n```\n## Coverage Strategy\n```\n",
		"tasks.md":  "# Tasks\n- [x] done\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Error("expected fenced coverage heading to fail, got pass")
	}
}

// TestCheckCoverageStrategy_NegatedBody pins GATE-FR-006's body
// negation exclusion: body text denying tests is not a strategy.
func TestCheckCoverageStrategy_NegatedBody(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"design.md": "# Design\n\nNo unit tests are planned for this change.\n",
		"tasks.md":  "# Tasks\n- [x] done\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Error("expected negated body text to fail, got pass")
	}
}

// TestCheckCoverageStrategy_ZeroPercentage pins GATE-FR-006's
// zero-percent exclusion: a measured coverage of zero denies a
// strategy rather than stating one.
func TestCheckCoverageStrategy_ZeroPercentage(t *testing.T) {
	dir := t.TempDir()
	scaffoldOpenSpec(t, dir, map[string]string{
		"design.md": "# Design\n\nCurrent coverage: 0%.\n",
		"tasks.md":  "# Tasks\n- [x] done\n",
	})

	r := checkCoverageStrategy(testCtx(dir))
	if r.Passed {
		t.Error("expected zero-percent line to fail, got pass")
	}
}
