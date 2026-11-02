package gate

import (
	"bufio"
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

// fileExists reports whether path exists and is a regular file.
func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// globFirst returns the first (alphabetical) match of
// <dir>/<base>/*/<file>, or "" when none exist.
func globFirst(dir, base, file string) string {
	matches, _ := filepath.Glob(filepath.Join(dir, base, "*", file))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

// checkArtifactExists probes both OpenSpec and Speckit locations
// for an artifact. Returns a CheckResult with the given check
// name and label. Glob errors are impossible since patterns are
// constructed from constants via filepath.Join.
func checkArtifactExists(q artifactQuery) CheckResult {
	// Branch-aware resolution: when the current branch identifies a
	// change, validate that change only. Falling back to another
	// change would let an unrelated change mask a missing artifact
	// for the active change — a false pass in multi-change repos.
	branch := detectBranch(q.dir)

	if active := activeChangeName(branch, openspecChangesBase); active != "" {
		p := filepath.Join(q.dir, openspecChangesBase, active, q.openspecFile)
		if fileExists(p) {
			return q.passAt(p)
		}
		return CheckResult{
			Name:    q.checkName,
			Passed:  false,
			Message: fmt.Sprintf("No %s artifact found for change %s (expected %s/%s/%s)", q.label, active, openspecChangesBase, active, q.openspecFile),
		}
	}

	if active := activeChangeName(branch, speckitBase); active != "" {
		p := filepath.Join(q.dir, speckitBase, active, q.speckitFile)
		if fileExists(p) {
			return q.passAt(p)
		}
		return CheckResult{
			Name:    q.checkName,
			Passed:  false,
			Message: fmt.Sprintf("No %s artifact found for change %s (expected %s/%s/%s)", q.label, active, speckitBase, active, q.speckitFile),
		}
	}

	// No branch mapping: legacy first-match behavior across all
	// changes.
	if p := globFirst(q.dir, openspecChangesBase, q.openspecFile); p != "" {
		return q.passAt(p)
	}
	if p := globFirst(q.dir, speckitBase, q.speckitFile); p != "" {
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
// a spec artifact (FR-004). Passes if found in either location.
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
// a plan/design artifact (FR-005).
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
// a tasks artifact (FR-006).
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
// coverage strategy is present (FR-006, Constitution IV).
var coverageKeywords = []string{
	"coverage target",
	"unit test",
	"integration test",
	"e2e test",
}

// checkCoverageStrategy scans plan and tasks artifacts for a
// coverage strategy per FR-006 and Constitution IV. Detection
// uses three patterns: a heading containing "coverage"
// (case-insensitive), body text containing specific keywords,
// or a line matching the coverage percentage regex.
func checkCoverageStrategy(dir string) CheckResult {
	// Collect candidate files from both OpenSpec and Speckit.
	var candidates []string

	// OpenSpec tasks and design.
	for _, name := range []string{"tasks.md", "design.md"} {
		pattern := filepath.Join(dir, "openspec", "changes", "*", name)
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			candidates = append(candidates, matches...)
		}
	}

	// Speckit tasks and plan.
	for _, name := range []string{"tasks.md", "plan.md"} {
		pattern := filepath.Join(dir, "specs", "*", name)
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			candidates = append(candidates, matches...)
		}
	}

	var readErr error
	for _, path := range candidates {
		found, err := hasCoverageStrategy(path)
		if err != nil {
			if readErr == nil {
				readErr = err
			}
			continue
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

	message := "No coverage strategy found in plan or tasks"
	if readErr != nil {
		message = fmt.Sprintf("%s (cannot read artifact: %v)", message, readErr)
	}
	return CheckResult{
		Name:    "coverage-strategy",
		Passed:  false,
		Message: message,
	}
}

// hasCoverageStrategy scans a file for coverage strategy
// indicators per FR-006.
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
// per FR-007. If any `- [ ]` lines remain, the check fails.
func checkTasksComplete(dir string) CheckResult {
	rel, content, ok, message := loadTasksFile(dir)
	if !ok {
		return CheckResult{Name: "tasks-complete", Passed: false, Message: message}
	}

	if strings.Contains(string(content), "- [ ]") {
		return CheckResult{
			Name:    "tasks-complete",
			Passed:  false,
			Message: fmt.Sprintf("Incomplete tasks remain in %s (unchecked checkboxes found)", rel),
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
// content. When the file is absent or unreadable, ok is false and
// message carries the user-facing failure text — an expected check
// outcome, not a program error.
func loadTasksFile(dir string) (rel string, content []byte, ok bool, message string) {
	path := findTasksFile(dir)
	if path == "" {
		return "", nil, false, "No tasks.md found"
	}
	rel, _ = filepath.Rel(dir, path)
	content, err := os.ReadFile(path)
	if err != nil {
		return rel, nil, false, fmt.Sprintf("Cannot read %s: %v", rel, err)
	}
	return rel, content, true, ""
}

// reviewMarkers are the HTML comment markers that indicate a
// review has been executed.
var reviewMarkers = []string{
	"<!-- spec-review: passed -->",
	"<!-- code-review: passed -->",
}

// checkReviewRun looks for any review marker in tasks.md per
// FR-007. Accepts either spec-review or code-review marker.
func checkReviewRun(dir string) CheckResult {
	rel, content, ok, message := loadTasksFile(dir)
	if !ok {
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
// tasks.md per FR-008. Unlike checkReviewRun, this requires
// specifically the code-review marker.
func checkReviewPass(dir string) CheckResult {
	rel, content, ok, message := loadTasksFile(dir)
	if !ok {
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
// per FR-008. Uses exec.Command with explicit argument
// separation (no shell invocation) per Constitution V.
func checkNotOnMain(dir string) CheckResult {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
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
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			detail = strings.TrimSpace(string(exitErr.Stderr))
		}
		return CheckResult{
			Name:    "not-on-main",
			Passed:  false,
			Message: fmt.Sprintf("Cannot determine git branch: %s (is this a git repository?)", detail),
		}
	}

	branch := strings.TrimSpace(string(out))

	// Detached HEAD returns "HEAD" — this passes since it is
	// not main (FR-008 scenario: Detached HEAD state).
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
// current branch maps to. When the branch identifies no change it
// falls back to the first alphabetical match across both layouts.
// Returns empty string if not found.
func findTasksFile(dir string) string {
	branch := detectBranch(dir)

	if active := activeChangeName(branch, openspecChangesBase); active != "" {
		p := filepath.Join(dir, openspecChangesBase, active, "tasks.md")
		if fileExists(p) {
			return p
		}
		return ""
	}

	if active := activeChangeName(branch, speckitBase); active != "" {
		p := filepath.Join(dir, speckitBase, active, "tasks.md")
		if fileExists(p) {
			return p
		}
		return ""
	}

	if p := globFirst(dir, openspecChangesBase, "tasks.md"); p != "" {
		return p
	}
	return globFirst(dir, speckitBase, "tasks.md")
}
