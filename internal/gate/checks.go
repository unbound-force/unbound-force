package gate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckFunc is the signature for all gate check functions.
// Each function accepts a run context — the project directory
// plus the change resolution resolved once per run — and returns
// a structured result. Check functions are pure — they read
// the filesystem but produce no side effects (Constitution IV).
type CheckFunc func(ctx runContext) CheckResult

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

	// BranchErr records why branch detection failed ("" means
	// detection succeeded). Checks that consume branch identity
	// (not-on-main) report this failure; artifact resolution
	// proceeds as if the branch were unmapped.
	BranchErr error

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

	// ResolveErr records why fallback change discovery failed (""
	// means discovery succeeded or was not needed). Run escalates
	// it to an internal error (exit 2) before any check executes.
	ResolveErr error
}

// runContext carries the project directory and the change
// resolution for one gate run. Run resolves the context once so
// every check sees a single branch snapshot and report provenance
// matches the resolution the checks used.
type runContext struct {
	// dir is the canonicalized project root directory.
	dir string

	// res is the change resolution for the run.
	res changeResolution
}

// resolveChange maps the current git branch to the change it names.
// Branches that map to a change (opsx/<name> under openspec/changes,
// NNN-<name> under specs) resolve strictly to that change: another
// change must never mask a missing artifact for the active change —
// a false pass in multi-change repos (design R1). quick/<name>
// branches resolve strictly to no change: the quick tier has no
// spec artifacts, so unrelated changes must not satisfy its checks
// either. Unmapped branches resolve to one first-alphabetical change
// directory for the whole run, so artifacts of different changes are
// never combined to satisfy a phase (design R1).
//
// A detached HEAD makes git report the literal branch "HEAD", which
// names no change. Because CI runners check out commits in a
// detached state, the resolver first recovers the branch from
// CI-provided refs (GITHUB_HEAD_REF on pull request events, then
// GITHUB_REF_NAME on push events) before applying the fallback, so
// headless CI runs keep strict branch-keyed resolution (design R1).
func resolveChange(dir string) changeResolution {
	branch, branchErr := detectBranch(dir)
	if branchErr == nil && branch == "HEAD" {
		if ref := ciBranchRef(); ref != "" {
			branch = ref
		}
	}
	if active := activeChangeName(branch, openspecChangesBase); active != "" {
		return changeResolution{Branch: branch, BranchErr: branchErr, Name: active, Base: openspecChangesBase, Strict: true}
	}
	if active := activeChangeName(branch, speckitBase); active != "" {
		return changeResolution{Branch: branch, BranchErr: branchErr, Name: active, Base: speckitBase, Strict: true}
	}
	if name, ok := strings.CutPrefix(branch, "quick/"); ok && name != "" {
		return changeResolution{Branch: branch, BranchErr: branchErr, Strict: true, Quick: true}
	}
	res := changeResolution{Branch: branch, BranchErr: branchErr}
	if name, base, err := firstChangeDir(dir); err != nil {
		res.ResolveErr = err
	} else {
		res.Name, res.Base = name, base
	}
	return res
}

// ciBranchRef returns the branch name a CI runner reports for this
// run, or "". GitHub Actions checks out a detached HEAD; pull
// request events expose the head branch through GITHUB_HEAD_REF and
// push events expose the pushed branch through GITHUB_REF_NAME.
// Refs that name no change (for example PR merge refs like
// "123/merge") simply map to the unmapped fallback. The value is
// external input (Constitution V): refs that could alter path
// resolution when joined into artifact paths are rejected rather
// than trusted.
func ciBranchRef() string {
	ref := strings.TrimSpace(os.Getenv("GITHUB_HEAD_REF"))
	if ref == "" {
		ref = strings.TrimSpace(os.Getenv("GITHUB_REF_NAME"))
	}
	if !validCIRef(ref) {
		return ""
	}
	return ref
}

// validCIRef reports whether ref is safe to use for change
// resolution: no absolute paths, backslashes, traversal, empty
// path components, or control characters that could escape the
// artifact base directories or inject content into messages.
func validCIRef(ref string) bool {
	if ref == "" {
		return false
	}
	if filepath.IsAbs(ref) || strings.Contains(ref, "\\") {
		return false
	}
	if strings.ContainsFunc(ref, isControlRune) {
		return false
	}
	for _, part := range strings.Split(ref, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

// isControlRune reports whether r is a control character. Git
// refnames cannot contain control characters, so their presence
// in a CI-provided ref is injected content, never a branch name.
func isControlRune(r rune) bool {
	return r < 0x20 || (r >= 0x7f && r <= 0x9f)
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

// firstChangeDir returns the first alphabetical change directory
// below the supported layouts — openspec/changes first, then
// specs. Symlinked change directories are followed so fallback
// resolution matches strict resolution. It returns empty names
// when no change directory exists in either layout.
func firstChangeDir(dir string) (name, base string, err error) {
	for _, b := range []string{openspecChangesBase, speckitBase} {
		entries, readErr := os.ReadDir(filepath.Join(dir, b))
		if errors.Is(readErr, os.ErrNotExist) {
			continue
		}
		if readErr != nil {
			return "", "", readErr
		}
		for _, entry := range entries {
			info, statErr := os.Stat(filepath.Join(dir, b, entry.Name()))
			if errors.Is(statErr, os.ErrNotExist) {
				continue
			}
			if statErr != nil {
				return "", "", statErr
			}
			if info.IsDir() {
				return entry.Name(), b, nil
			}
		}
	}
	return "", "", nil
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

// checkArtifactExists probes the resolved change for an artifact
// across both OpenSpec and Speckit locations. Returns a CheckResult
// with the given check name and label. Resolution follows the run
// context: branch-mapped and quick branches never fall back to
// another change's artifacts; unmapped branches resolve one
// first-alphabetical change for the whole run.
func checkArtifactExists(ctx runContext, q artifactQuery) CheckResult {
	res := ctx.res

	if res.Quick {
		return CheckResult{
			Name:    q.checkName,
			Passed:  false,
			Message: fmt.Sprintf("No %s artifact can satisfy quick branch %s: quick changes carry no spec artifacts", q.label, res.Branch),
		}
	}

	if res.Name == "" {
		return CheckResult{
			Name:    q.checkName,
			Passed:  false,
			Message: fmt.Sprintf("No change directory found under %s or %s", openspecChangesBase, speckitBase),
		}
	}

	file := q.openspecFile
	if res.Base == speckitBase {
		file = q.speckitFile
	}
	p := filepath.Join(ctx.dir, res.Base, res.Name, file)
	exists, err := regularFile(p)
	if err != nil {
		return internalCheck(q.checkName, err)
	}
	if exists {
		return q.passAt(ctx.dir, p)
	}
	return CheckResult{
		Name:    q.checkName,
		Passed:  false,
		Message: fmt.Sprintf("No %s artifact found for change %s (expected %s/%s/%s)", q.label, res.Name, res.Base, res.Name, file),
	}
}

// passAt builds the passing CheckResult for a resolved artifact path.
func (q artifactQuery) passAt(dir, path string) CheckResult {
	rel, _ := filepath.Rel(dir, path)
	return CheckResult{
		Name:    q.checkName,
		Passed:  true,
		Message: fmt.Sprintf("%s found at %s", q.label, rel),
	}
}

// checkSpecExists probes both OpenSpec and Speckit locations for
// a spec artifact (GATE-FR-004). Passes if found in either location.
func checkSpecExists(ctx runContext) CheckResult {
	return checkArtifactExists(ctx, artifactQuery{
		checkName:    "spec-exists",
		label:        "Spec",
		openspecFile: "proposal.md",
		speckitFile:  "spec.md",
	})
}

// checkPlanExists probes both OpenSpec and Speckit locations for
// a plan/design artifact (GATE-FR-005).
func checkPlanExists(ctx runContext) CheckResult {
	return checkArtifactExists(ctx, artifactQuery{
		checkName:    "plan-exists",
		label:        "Plan",
		openspecFile: "design.md",
		speckitFile:  "plan.md",
	})
}

// checkTasksExist probes both OpenSpec and Speckit locations for
// a tasks artifact (GATE-FR-006).
func checkTasksExist(ctx runContext) CheckResult {
	return checkArtifactExists(ctx, artifactQuery{
		checkName:    "tasks-exists",
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
func checkCoverageStrategy(ctx runContext) CheckResult {
	// By-design resolution failures carry the same explanation the
	// artifact-existence checks emit, so a quick-branch or no-change
	// log reads as a tier limitation rather than a missing artifact.
	if ctx.res.Quick {
		return CheckResult{
			Name:    "coverage-strategy",
			Passed:  false,
			Message: fmt.Sprintf("No coverage strategy can satisfy quick branch %s: quick changes carry no spec artifacts", ctx.res.Branch),
		}
	}
	if ctx.res.Name == "" {
		return CheckResult{
			Name:    "coverage-strategy",
			Passed:  false,
			Message: fmt.Sprintf("No change directory found under %s or %s", openspecChangesBase, speckitBase),
		}
	}
	candidates, err := coverageArtifacts(ctx)
	if err != nil {
		return internalCheck("coverage-strategy", err)
	}
	for _, path := range candidates {
		found, err := hasCoverageStrategy(path)
		if err != nil {
			return internalCheck("coverage-strategy", err)
		}
		if found {
			rel, _ := filepath.Rel(ctx.dir, path)
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

// coverageArtifacts returns the plan and tasks artifacts of the
// run's resolved change: the branch-mapped change's artifacts
// only, or the one first-alphabetical change resolved for the
// whole run on unmapped branches — coverage text in another
// change cannot satisfy the active resolution. quick branches
// and runs with no change directory resolve to no candidates.
func coverageArtifacts(ctx runContext) ([]string, error) {
	res := ctx.res

	if res.Quick || res.Name == "" {
		return nil, nil
	}

	names := []string{"tasks.md", "design.md"}
	if res.Base == speckitBase {
		names = []string{"tasks.md", "plan.md"}
	}
	return changeArtifacts(ctx.dir, res.Base, res.Name, names)
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

// negatedCoverageRe matches headings that mention coverage in a
// negated form: "No coverage strategy is planned" contains the
// word "coverage" but denies a strategy, so the heading signal
// must not count it (GATE-FR-006).
var negatedCoverageRe = regexp.MustCompile(`(?i)\b(?:no|without|missing|lacking)\s+coverage\b|\bcoverage\s+strategy\s+is\s+(?:not|absent|none)\b`)

// negatedCoverageBodyRe matches body text and percentage lines that
// deny a coverage strategy rather than state one: "No unit tests
// are planned", "without integration tests", or a measured coverage
// of zero percent (GATE-FR-006).
var negatedCoverageBodyRe = regexp.MustCompile(`(?i)\b(?:no|without|missing|lacking)\s+(?:unit\s+|integration\s+|e2e\s+)?tests?\b|\bcoverage\b[^%\n]{0,40}?\b0\s*%`)

// hasCoverageStrategy scans a file for coverage strategy
// indicators per GATE-FR-006: a heading containing "coverage"
// that does not negate it, non-negated body keywords, or a
// coverage percentage line that does not report a measured
// zero. Signals inside fenced code blocks are documentation,
// not strategy, and do not count. The file is read whole,
// matching the sibling checks that consume the same artifacts:
// artifact content is check input, never an internal error.
func hasCoverageStrategy(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	fence := fenceTracker{}
	for _, line := range strings.Split(string(content), "\n") {
		if fence.update(line) {
			continue
		}
		lower := strings.ToLower(line)

		// Heading containing "coverage" (case-insensitive), unless
		// the heading negates it (GATE-FR-006).
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(lower, "coverage") && !negatedCoverageRe.MatchString(line) {
			return true, nil
		}

		// Body text keywords and percentage lines, unless the
		// line negates them (GATE-FR-006).
		if !negatedCoverageBodyRe.MatchString(line) {
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
	}
	return false, nil
}

// fenceTracker implements CommonMark code-fence matching for the
// artifact scanners: a fence opens with three or more backticks or
// tildes and closes only on a line of the same marker character,
// at least as long as the opening fence, carrying nothing but the
// marker. A line using the other marker inside an open fence is
// content, so mixed-marker markdown cannot flip the state and mask
// subsequent task lines or coverage signals (GATE-FR-006,
// GATE-FR-007).
type fenceTracker struct {
	marker byte // 0 when no fence is open; '`' or '~' otherwise
	length int  // marker repetitions of the opening fence
}

// fenced reports whether the tracker currently sits inside a fence.
func (f *fenceTracker) fenced() bool { return f.marker != 0 }

// update feeds one line to the tracker and reports whether the line
// belongs to a fenced code block — a delimiter or content. Callers
// skip such lines when scanning for task and coverage signals.
func (f *fenceTracker) update(line string) bool {
	marker, length, closing := fenceLine(line)
	if marker == 0 {
		return f.fenced()
	}
	if !f.fenced() {
		f.marker, f.length = marker, length
		return true
	}
	if closing && marker == f.marker && length >= f.length {
		f.marker, f.length = 0, 0
	}
	return true
}

// fenceLine classifies a line as a code-fence delimiter. It returns
// the marker character (0 for non-fence lines), the marker length,
// and whether the line can close a fence — marker characters only,
// with no info string. An opening fence may carry an info string,
// but a backtick fence's info string must not contain a backtick.
func fenceLine(line string) (marker byte, length int, closing bool) {
	s := strings.TrimSpace(line)
	if s == "" {
		return 0, 0, false
	}
	c := s[0]
	if c != '`' && c != '~' {
		return 0, 0, false
	}
	n := 0
	for n < len(s) && s[n] == c {
		n++
	}
	if n < 3 {
		return 0, 0, false
	}
	info := strings.TrimSpace(s[n:])
	if info == "" {
		return c, n, true
	}
	if c == '`' && strings.Contains(info, "`") {
		return 0, 0, false
	}
	return c, n, false
}

// checkTasksComplete scans tasks.md for unchecked task boxes
// per GATE-FR-007. If any `- [ ]` checkbox lines remain — including
// indented subtasks — the check fails. Prose mentioning the
// literal mid-line and examples inside fenced code blocks (which
// close only on their own marker and length, per CommonMark) do
// not count.
func checkTasksComplete(ctx runContext) CheckResult {
	rel, content, message, err := loadTasksFile(ctx)
	if err != nil {
		return internalCheck("tasks-complete", err)
	}
	if message != "" {
		return CheckResult{Name: "tasks-complete", Passed: false, Message: message}
	}

	fence := fenceTracker{}
	for _, line := range strings.Split(string(content), "\n") {
		// Fenced code blocks can document task syntax without
		// declaring tasks; their delimiters and contents are not
		// task lines (GATE-FR-007).
		if fence.update(line) {
			continue
		}
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
func loadTasksFile(ctx runContext) (rel string, content []byte, message string, readErr error) {
	// By-design resolution failures carry the same explanation the
	// artifact-existence checks emit, so a quick-branch or no-change
	// log reads as a tier limitation rather than a missing artifact.
	if ctx.res.Quick {
		return "", nil, fmt.Sprintf("No tasks.md can satisfy quick branch %s: quick changes carry no spec artifacts", ctx.res.Branch), nil
	}
	if ctx.res.Name == "" {
		return "", nil, fmt.Sprintf("No change directory found under %s or %s", openspecChangesBase, speckitBase), nil
	}
	path, err := findTasksFile(ctx)
	if err != nil {
		return "", nil, "", err
	}
	if path == "" {
		return "", nil, "No tasks.md found", nil
	}
	rel, _ = filepath.Rel(ctx.dir, path)
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
	codeReviewMarker,
}

// checkReviewRun looks for any review marker in tasks.md per
// GATE-FR-007. Accepts either spec-review or code-review marker.
func checkReviewRun(ctx runContext) CheckResult {
	rel, content, message, err := loadTasksFile(ctx)
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
func checkReviewPass(ctx runContext) CheckResult {
	rel, content, message, err := loadTasksFile(ctx)
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

// checkNotOnMain verifies the run's resolved branch is not main per
// GATE-FR-008. It consumes the once-per-run branch resolution
// (design D2, GATE-FR-012) so the check, the report provenance,
// and every other check observe the same branch — including
// detached CI checkouts whose branch was recovered from
// GITHUB_HEAD_REF or GITHUB_REF_NAME.
func checkNotOnMain(ctx runContext) CheckResult {
	if ctx.res.BranchErr != nil {
		return CheckResult{
			Name:    "not-on-main",
			Passed:  false,
			Message: ctx.res.BranchErr.Error(),
		}
	}

	// A detached HEAD without a CI ref resolves to the literal
	// "HEAD" — this passes since it is not main (GATE-FR-008
	// scenario: Detached HEAD state).
	if ctx.res.Branch == "main" {
		return CheckResult{
			Name:    "not-on-main",
			Passed:  false,
			Message: "Current branch is main — direct commits to main are not allowed",
		}
	}

	return CheckResult{
		Name:    "not-on-main",
		Passed:  true,
		Message: fmt.Sprintf("Current branch is %s", ctx.res.Branch),
	}
}

// findTasksFile locates the tasks.md file of the run's resolved
// change: the branch-mapped change's tasks.md only, or the one
// first-alphabetical change resolved for the whole run on
// unmapped branches. quick branches and runs with no change
// directory resolve to no tasks file. Returns an empty path if
// not found.
func findTasksFile(ctx runContext) (string, error) {
	res := ctx.res

	if res.Quick || res.Name == "" {
		return "", nil
	}

	p := filepath.Join(ctx.dir, res.Base, res.Name, "tasks.md")
	exists, err := regularFile(p)
	if err != nil {
		return "", err
	}
	if exists {
		return p, nil
	}
	return "", nil
}
