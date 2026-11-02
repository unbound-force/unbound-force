package gate

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckFunc is the signature for all gate check functions.
// Each function accepts a project root directory and returns
// a structured result. Check functions are pure — they read
// the filesystem but produce no side effects (Constitution IV).
type CheckFunc func(dir string) CheckResult

// validPhases lists the accepted values for --phase per D2.
var validPhases = []string{
	"specify", "plan", "implement", "review", "pr",
}

// ValidPhases returns a copy of the valid phase names for use
// in CLI help text and error messages.
func ValidPhases() []string {
	out := make([]string, len(validPhases))
	copy(out, validPhases)
	return out
}

// phaseChecks maps each workflow phase to the check functions
// that must pass before the phase is allowed per D2.
var phaseChecks = map[string][]CheckFunc{
	"specify":   {checkSpecExists},
	"plan":      {checkSpecExists, checkPlanExists},
	"implement": {checkSpecExists, checkPlanExists, checkTasksExist, checkCoverageStrategy},
	"review":    {checkTasksComplete, checkReviewRun},
	"pr":        {checkReviewPass, checkNotOnMain},
}

func init() {
	for _, phase := range validPhases {
		if _, ok := phaseChecks[phase]; !ok {
			panic(fmt.Sprintf(
				"gate: validPhases contains %q but phaseChecks has no entry for it",
				phase))
		}
	}
}

// artifactQuery describes the artifact to probe across both
// OpenSpec and Speckit locations.
type artifactQuery struct {
	dir          string // project root directory
	checkName    string // check identifier for the result
	label        string // human-readable artifact label
	openspecFile string // filename under openspec/changes/*/
	speckitFile  string // filename under specs/*/
}

// Base directories of the two supported artifact layouts.
const (
	openspecChangesBase = "openspec/changes"
	speckitBase         = "specs"
)

// speckitBranchRe matches Speckit feature branches (NNN-<name>).
var speckitBranchRe = regexp.MustCompile(`^[0-9]{3,}-`)

// activeChangeName returns the change directory name that the given
// branch maps to under base, or "" when the branch does not identify
// a change in that layout:
//
//	opsx/<name> under openspec/changes -> <name>
//	NNN-<name> under specs             -> NNN-<name> (full branch)
func activeChangeName(branch, base string) string {
	switch base {
	case openspecChangesBase:
		if name, ok := strings.CutPrefix(branch, "opsx/"); ok && name != "" {
			return name
		}
	case speckitBase:
		if speckitBranchRe.MatchString(branch) {
			return branch
		}
	}
	return ""
}

// changeResolution describes the change a gate run evaluates.
type changeResolution struct {
	// Branch is the git branch name examined ("" when git is
	// unavailable or the directory is not a repository).
	Branch string

	// Name is the change directory the branch maps to ("" when the
	// branch names no change).
	Name string

	// Base is the layout base directory for Name ("" when unmapped).
	Base string

	// Strict reports whether cross-change fallback is forbidden:
	// true whenever the branch names a change or workflow tier.
	Strict bool

	// Quick reports a quick/<name> branch. The quick tier carries no
	// spec artifacts, so no artifact can satisfy its checks.
	Quick bool
}

// resolveChange maps the current git branch to the change it names.
// Branches that map to a change (opsx/<name> under openspec/changes,
// NNN-<name> under specs) resolve strictly to that change: another
// change must never mask a missing artifact for the active change —
// a false pass in multi-change repos (design R1). quick/<name>
// branches resolve strictly to no change: the quick tier has no
// spec artifacts, so unrelated changes must not satisfy its checks
// either. Unmapped branches keep the documented alphabetical
// fallback.
func resolveChange(dir string) changeResolution {
	branch := detectBranch(dir)
	if active := activeChangeName(branch, openspecChangesBase); active != "" {
		return changeResolution{Branch: branch, Name: active, Base: openspecChangesBase, Strict: true}
	}
	if active := activeChangeName(branch, speckitBase); active != "" {
		return changeResolution{Branch: branch, Name: active, Base: speckitBase, Strict: true}
	}
	if name, ok := strings.CutPrefix(branch, "quick/"); ok && name != "" {
		return changeResolution{Branch: branch, Strict: true, Quick: true}
	}
	return changeResolution{Branch: branch}
}

// regularFile reports whether path exists and is a regular file.
func regularFile(path string) (bool, error) {
	st, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return st.Mode().IsRegular(), nil
}

// firstArtifactFile returns the first alphabetical artifact below base.
// It reads literal directory names so --dir values cannot alter discovery
// through filepath.Glob metacharacters.
func firstArtifactFile(dir, base, file string) (string, error) {
	entries, err := os.ReadDir(filepath.Join(dir, base))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, base, entry.Name(), file)
		exists, statErr := regularFile(path)
		if statErr != nil {
			return "", statErr
		}
		if exists {
			return path, nil
		}
	}
	return "", nil
}

// internalCheck builds a failed CheckResult carrying an internal
// error; Run escalates it to exit code 2.
func internalCheck(name string, err error) CheckResult {
	return CheckResult{
		Name:        name,
		Message:     fmt.Sprintf("cannot execute %s check: %v", name, err),
		internalErr: fmt.Errorf("%s check: %w", name, err),
	}
}

// checkArtifactExists probes both OpenSpec and Speckit locations
// for an artifact. Returns a CheckResult with the given check
// name and label. Resolution follows resolveChange: branch-mapped
// and quick branches never fall back to another change's artifacts.
func checkArtifactExists(q artifactQuery) CheckResult {
	res := resolveChange(q.dir)

	if res.Quick {
		return CheckResult{
			Name:    q.checkName,
			Passed:  false,
			Message: fmt.Sprintf("No %s artifact can satisfy quick branch %s: quick changes carry no spec artifacts", q.label, res.Branch),
		}
	}

	if res.Strict {
		file := q.openspecFile
		if res.Base == speckitBase {
			file = q.speckitFile
		}
		p := filepath.Join(q.dir, res.Base, res.Name, file)
		exists, err := regularFile(p)
		if err != nil {
			return internalCheck(q.checkName, err)
		}
		if exists {
			return q.passAt(p)
		}
		return CheckResult{
			Name:    q.checkName,
			Passed:  false,
			Message: fmt.Sprintf("No %s artifact found for change %s (expected %s/%s/%s)", q.label, res.Name, res.Base, res.Name, file),
		}
	}

	// No branch mapping: legacy first-match behavior across all
	// changes.
	if p, err := firstArtifactFile(q.dir, openspecChangesBase, q.openspecFile); err != nil {
		return internalCheck(q.checkName, err)
	} else if p != "" {
		return q.passAt(p)
	}
	if p, err := firstArtifactFile(q.dir, speckitBase, q.speckitFile); err != nil {
		return internalCheck(q.checkName, err)
	} else if p != "" {
		return q.passAt(p)
	}

	return CheckResult{
		Name:    q.checkName,
		Passed:  false,
		Message: fmt.Sprintf("No %s artifact found (expected %s/*/%s or %s/*/%s)", q.label, openspecChangesBase, q.openspecFile, speckitBase, q.speckitFile),
	}
}

// passAt builds the passing CheckResult for a resolved artifact path.
func (q artifactQuery) passAt(path string) CheckResult {
	rel, _ := filepath.Rel(q.dir, path)
	return CheckResult{
		Name:    q.checkName,
		Passed:  true,
		Message: fmt.Sprintf("%s found at %s", q.label, rel),
	}
}

// checkSpecExists probes both OpenSpec and Speckit locations for
// a spec artifact (GATE-FR-004). Passes if found in either location.
func checkSpecExists(dir string) CheckResult {
	return checkArtifactExists(artifactQuery{
		dir:          dir,
		checkName:    "spec-exists",
		label:        "Spec",
		openspecFile: "proposal.md",
		speckitFile:  "spec.md",
	})
}

// checkPlanExists probes both OpenSpec and Speckit locations for
// a plan/design artifact (GATE-FR-005).
func checkPlanExists(dir string) CheckResult {
	return checkArtifactExists(artifactQuery{
		dir:          dir,
		checkName:    "plan-exists",
		label:        "Plan",
		openspecFile: "design.md",
		speckitFile:  "plan.md",
	})
}

// checkTasksExist probes both OpenSpec and Speckit locations for
// a tasks artifact (GATE-FR-006).
func checkTasksExist(dir string) CheckResult {
	return checkArtifactExists(artifactQuery{
		dir:          dir,
		checkName:    "tasks-exist",
		label:        "Tasks",
		openspecFile: "tasks.md",
		speckitFile:  "tasks.md",
	})
}

// coveragePctRe matches lines containing a coverage percentage
// target (e.g., "coverage target: >= 80%").
var coveragePctRe = regexp.MustCompile(`(?i)coverage.*%`)

// coverageKeywords are the body-text terms that indicate a
// coverage strategy is present (GATE-FR-006, Constitution IV).
var coverageKeywords = []string{
	"coverage target",
	"unit test",
	"integration test",
	"e2e test",
}

// checkCoverageStrategy scans plan and tasks artifacts for a
// coverage strategy per GATE-FR-006 and Constitution IV. Detection
// uses three patterns: a heading containing "coverage"
// (case-insensitive), body text containing specific keywords,
// or a line matching the coverage percentage regex.
func checkCoverageStrategy(dir string) CheckResult {
	candidates, err := coverageArtifacts(dir)
	if err != nil {
		return internalCheck("coverage-strategy", err)
	}
	for _, path := range candidates {
		found, err := hasCoverageStrategy(path)
		if err != nil {
			return internalCheck("coverage-strategy", err)
		}
		if found {
			rel, _ := filepath.Rel(dir, path)
			return CheckResult{
				Name:    "coverage-strategy",
				Passed:  true,
				Message: fmt.Sprintf("Coverage strategy found in %s", rel),
			}
		}
	}

	return CheckResult{
		Name:    "coverage-strategy",
		Passed:  false,
		Message: "No coverage strategy found in plan or tasks",
	}
}

// coverageArtifacts returns the plan and tasks artifacts selected for
// the current branch (resolveChange): the branch-mapped change's
// artifacts only, or the documented alphabetical fallback used by the
// other artifact checks on unmapped branches. quick branches resolve
// to no candidates.
func coverageArtifacts(dir string) ([]string, error) {
	res := resolveChange(dir)

	if res.Quick {
		return nil, nil
	}

	if res.Strict {
		names := []string{"tasks.md", "design.md"}
		if res.Base == speckitBase {
			names = []string{"tasks.md", "plan.md"}
		}
		return changeArtifacts(dir, res.Base, res.Name, names)
	}

	var candidates []string
	for _, name := range []string{"tasks.md", "design.md"} {
		path, err := firstArtifactFile(dir, openspecChangesBase, name)
		if err != nil {
			return nil, err
		}
		if path != "" {
			candidates = append(candidates, path)
		}
	}
	for _, name := range []string{"tasks.md", "plan.md"} {
		path, err := firstArtifactFile(dir, speckitBase, name)
		if err != nil {
			return nil, err
		}
		if path != "" {
			candidates = append(candidates, path)
		}
	}
	return candidates, nil
}

func changeArtifacts(dir, base, change string, names []string) ([]string, error) {
	var candidates []string
	for _, name := range names {
		path := filepath.Join(dir, base, change, name)
		exists, err := regularFile(path)
		if err != nil {
			return nil, err
		}
		if exists {
			candidates = append(candidates, path)
		}
	}
	return candidates, nil
}

// hasCoverageStrategy scans a file for coverage strategy
// indicators per GATE-FR-006.
func hasCoverageStrategy(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)

		// Heading containing "coverage" (case-insensitive).
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(lower, "coverage") {
			return true, nil
		}

		// Body text keywords.
		for _, kw := range coverageKeywords {
			if strings.Contains(lower, kw) {
				return true, nil
			}
		}

		// Coverage percentage pattern.
		if coveragePctRe.MatchString(line) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// checkTasksComplete scans tasks.md for unchecked task boxes
// per GATE-FR-007. If any `- [ ]` checkbox lines remain — including
// indented subtasks — the check fails. Prose mentioning the
// literal mid-line does not count.
func checkTasksComplete(dir string) CheckResult {
	rel, content, message, err := loadTasksFile(dir)
	if err != nil {
		return internalCheck("tasks-complete", err)
	}
	if message != "" {
		return CheckResult{Name: "tasks-complete", Passed: false, Message: message}
	}

	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "- [ ]") {
			return CheckResult{
				Name:    "tasks-complete",
				Passed:  false,
				Message: fmt.Sprintf("Incomplete tasks remain in %s (unchecked checkboxes found)", rel),
			}
		}
	}

	return CheckResult{
		Name:    "tasks-complete",
		Passed:  true,
		Message: fmt.Sprintf("All task checkboxes are complete in %s", rel),
	}
}

// loadTasksFile resolves the active change's tasks.md and reads it.
// On success it returns a relative path (for messages) and the file
// content. A missing file is an expected check outcome; a read failure is
// returned so Run can classify it as an internal error.
func loadTasksFile(dir string) (rel string, content []byte, message string, readErr error) {
	path, err := findTasksFile(dir)
	if err != nil {
		return "", nil, "", err
	}
	if path == "" {
		return "", nil, "No tasks.md found", nil
	}
	rel, _ = filepath.Rel(dir, path)
	content, err = os.ReadFile(path)
	if err != nil {
		return rel, nil, "", fmt.Errorf("read %s: %w", rel, err)
	}
	return rel, content, "", nil
}

// reviewMarkers are the HTML comment markers that indicate a
// review has been executed.
var reviewMarkers = []string{
	"<!-- spec-review: passed -->",
	"<!-- code-review: passed -->",
}

// checkReviewRun looks for any review marker in tasks.md per
// GATE-FR-007. Accepts either spec-review or code-review marker.
func checkReviewRun(dir string) CheckResult {
	rel, content, message, err := loadTasksFile(dir)
	if err != nil {
		return internalCheck("review-run", err)
	}
	if message != "" {
		return CheckResult{Name: "review-run", Passed: false, Message: message}
	}

	text := string(content)
	for _, marker := range reviewMarkers {
		if strings.Contains(text, marker) {
			return CheckResult{
				Name:    "review-run",
				Passed:  true,
				Message: fmt.Sprintf("Review marker found in %s", rel),
			}
		}
	}

	return CheckResult{
		Name:    "review-run",
		Passed:  false,
		Message: fmt.Sprintf("No review has been recorded in %s (no spec-review or code-review marker found)", rel),
	}
}

// codeReviewMarker is the specific marker required for the PR
// phase's review-pass check.
const codeReviewMarker = "<!-- code-review: passed -->"

// checkReviewPass looks for the code-review passed marker in
// tasks.md per GATE-FR-008. Unlike checkReviewRun, this requires
// specifically the code-review marker.
func checkReviewPass(dir string) CheckResult {
	rel, content, message, err := loadTasksFile(dir)
	if err != nil {
		return internalCheck("review-pass", err)
	}
	if message != "" {
		return CheckResult{Name: "review-pass", Passed: false, Message: message}
	}

	if strings.Contains(string(content), codeReviewMarker) {
		return CheckResult{
			Name:    "review-pass",
			Passed:  true,
			Message: fmt.Sprintf("Code review passed marker found in %s", rel),
		}
	}

	return CheckResult{
		Name:    "review-pass",
		Passed:  false,
		Message: fmt.Sprintf("No code-review passed marker found in %s", rel),
	}
}

// checkNotOnMain verifies the current git branch is not main
// per GATE-FR-008. Uses exec.CommandContext with a bounded timeout and
// explicit argument separation (no shell invocation) per
// Constitution V.
func checkNotOnMain(dir string) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), gitCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		// Distinguish between git not found and other errors.
		if _, lookErr := exec.LookPath("git"); lookErr != nil {
			return CheckResult{
				Name:    "not-on-main",
				Passed:  false,
				Message: "git is not available — run 'uf doctor' to check tool availability",
			}
		}
		detail := err.Error()
		if ctx.Err() != nil {
			detail = fmt.Sprintf("git did not respond within %s", gitCommandTimeout)
		} else {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
				detail = strings.TrimSpace(string(exitErr.Stderr))
			}
		}
		return CheckResult{
			Name:    "not-on-main",
			Passed:  false,
			Message: fmt.Sprintf("Cannot determine git branch: %s (is this a git repository?)", detail),
		}
	}

	branch := strings.TrimSpace(string(out))

	// Detached HEAD returns "HEAD" — this passes since it is
	// not main (GATE-FR-008 scenario: Detached HEAD state).
	if branch == "main" {
		return CheckResult{
			Name:    "not-on-main",
			Passed:  false,
			Message: "Current branch is main — direct commits to main are not allowed",
		}
	}

	return CheckResult{
		Name:    "not-on-main",
		Passed:  true,
		Message: fmt.Sprintf("Current branch is %s", branch),
	}
}

// findTasksFile locates the tasks.md file for the change that the
// current branch maps to (resolveChange): the branch-mapped change's
// tasks.md only, or the first alphabetical match across both layouts
// on unmapped branches. quick branches resolve to no tasks file.
// Returns an empty path if not found.
func findTasksFile(dir string) (string, error) {
	res := resolveChange(dir)

	if res.Quick {
		return "", nil
	}

	if res.Strict {
		p := filepath.Join(dir, res.Base, res.Name, "tasks.md")
		exists, err := regularFile(p)
		if err != nil {
			return "", err
		}
		if exists {
			return p, nil
		}
		return "", nil
	}

	if p, err := firstArtifactFile(dir, openspecChangesBase, "tasks.md"); err != nil {
		return "", err
	} else if p != "" {
		return p, nil
	}
	return firstArtifactFile(dir, speckitBase, "tasks.md")
}
