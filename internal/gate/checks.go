package gate

import (
	"bufio"
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

// checkArtifactExists probes both OpenSpec and Speckit locations
// for an artifact. Returns a CheckResult with the given check
// name and label. Glob errors are impossible since patterns are
// constructed from constants via filepath.Join.
func checkArtifactExists(dir, checkName, label, openspecFile, speckitFile string) CheckResult {
	// OpenSpec: openspec/changes/*/<file>
	pattern := filepath.Join(dir, "openspec", "changes", "*", openspecFile)
	if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
		rel, _ := filepath.Rel(dir, matches[0])
		return CheckResult{
			Name:    checkName,
			Passed:  true,
			Message: fmt.Sprintf("%s found at %s", label, rel),
		}
	}

	// Speckit: specs/*/<file>
	pattern = filepath.Join(dir, "specs", "*", speckitFile)
	if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
		rel, _ := filepath.Rel(dir, matches[0])
		return CheckResult{
			Name:    checkName,
			Passed:  true,
			Message: fmt.Sprintf("%s found at %s", label, rel),
		}
	}

	return CheckResult{
		Name:    checkName,
		Passed:  false,
		Message: fmt.Sprintf("No %s artifact found (expected openspec/changes/*/%s or specs/*/%s)", label, openspecFile, speckitFile),
	}
}

// checkSpecExists probes both OpenSpec and Speckit locations for
// a spec artifact (FR-004). Passes if found in either location.
func checkSpecExists(dir string) CheckResult {
	return checkArtifactExists(dir, "spec-exists", "Spec", "proposal.md", "spec.md")
}

// checkPlanExists probes both OpenSpec and Speckit locations for
// a plan/design artifact (FR-005).
func checkPlanExists(dir string) CheckResult {
	return checkArtifactExists(dir, "plan-exists", "Plan", "design.md", "plan.md")
}

// checkTasksExist probes both OpenSpec and Speckit locations for
// a tasks artifact (FR-006).
func checkTasksExist(dir string) CheckResult {
	return checkArtifactExists(dir, "tasks-exist", "Tasks", "tasks.md", "tasks.md")
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

	for _, path := range candidates {
		if hasCoverageStrategy(path) {
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

// hasCoverageStrategy scans a file for coverage strategy
// indicators per FR-006.
func hasCoverageStrategy(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)

		// Heading containing "coverage" (case-insensitive).
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(lower, "coverage") {
			return true
		}

		// Body text keywords.
		for _, kw := range coverageKeywords {
			if strings.Contains(lower, kw) {
				return true
			}
		}

		// Coverage percentage pattern.
		if coveragePctRe.MatchString(line) {
			return true
		}
	}

	return false
}

// checkTasksComplete scans tasks.md for unchecked task boxes
// per FR-007. If any `- [ ]` lines remain, the check fails.
func checkTasksComplete(dir string) CheckResult {
	path := findTasksFile(dir)
	if path == "" {
		return CheckResult{
			Name:    "tasks-complete",
			Passed:  false,
			Message: "No tasks.md found",
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return CheckResult{
			Name:    "tasks-complete",
			Passed:  false,
			Message: fmt.Sprintf("Cannot read tasks.md: %v", err),
		}
	}

	if strings.Contains(string(content), "- [ ]") {
		return CheckResult{
			Name:    "tasks-complete",
			Passed:  false,
			Message: "Incomplete tasks remain (unchecked checkboxes found)",
		}
	}

	return CheckResult{
		Name:    "tasks-complete",
		Passed:  true,
		Message: "All task checkboxes are complete",
	}
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
	path := findTasksFile(dir)
	if path == "" {
		return CheckResult{
			Name:    "review-run",
			Passed:  false,
			Message: "No tasks.md found",
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return CheckResult{
			Name:    "review-run",
			Passed:  false,
			Message: fmt.Sprintf("Cannot read tasks.md: %v", err),
		}
	}

	text := string(content)
	for _, marker := range reviewMarkers {
		if strings.Contains(text, marker) {
			return CheckResult{
				Name:    "review-run",
				Passed:  true,
				Message: "Review marker found in tasks.md",
			}
		}
	}

	return CheckResult{
		Name:    "review-run",
		Passed:  false,
		Message: "No review has been recorded (no spec-review or code-review marker found)",
	}
}

// codeReviewMarker is the specific marker required for the PR
// phase's review-pass check.
const codeReviewMarker = "<!-- code-review: passed -->"

// checkReviewPass looks for the code-review passed marker in
// tasks.md per FR-008. Unlike checkReviewRun, this requires
// specifically the code-review marker.
func checkReviewPass(dir string) CheckResult {
	path := findTasksFile(dir)
	if path == "" {
		return CheckResult{
			Name:    "review-pass",
			Passed:  false,
			Message: "No tasks.md found",
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return CheckResult{
			Name:    "review-pass",
			Passed:  false,
			Message: fmt.Sprintf("Cannot read tasks.md: %v", err),
		}
	}

	if strings.Contains(string(content), codeReviewMarker) {
		return CheckResult{
			Name:    "review-pass",
			Passed:  true,
			Message: "Code review passed marker found",
		}
	}

	return CheckResult{
		Name:    "review-pass",
		Passed:  false,
		Message: "No code-review passed marker found in tasks.md",
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
		return CheckResult{
			Name:    "not-on-main",
			Passed:  false,
			Message: fmt.Sprintf("Cannot determine git branch: %v (is this a git repository?)", err),
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

// findTasksFile locates the tasks.md file, checking both
// OpenSpec and Speckit locations. Returns empty string if
// not found.
func findTasksFile(dir string) string {
	// OpenSpec: openspec/changes/*/tasks.md
	pattern := filepath.Join(dir, "openspec", "changes", "*", "tasks.md")
	if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
		return matches[0]
	}

	// Speckit: specs/NNN-*/tasks.md
	pattern = filepath.Join(dir, "specs", "*", "tasks.md")
	if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
		return matches[0]
	}

	return ""
}
