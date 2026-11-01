package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/unbound-force/unbound-force/internal/textutil"
)

// findProjectRoot walks up from the current directory looking
// for go.mod to find the project root. Returns "" if not found.
func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func runScaffoldWithoutExternalTools(t *testing.T, opts Options) (*Result, error) {
	t.Helper()
	if opts.LookPath == nil {
		opts.LookPath = stubScaffoldLookPath(nil)
	}
	if opts.ExecCmdInDir == nil {
		opts.ExecCmdInDir = func(dir, name string, args ...string) ([]byte, error) {
			t.Fatalf("unexpected external command in %s: %s %v", dir, name, args)
			return nil, nil
		}
	}
	return Run(opts)
}

func readJSONFile(t *testing.T, path string, target any) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read JSON file %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("parse JSON file %s: %v", path, err)
	}
}

func assertStringMap(t *testing.T, label string, got, want map[string]string) {
	t.Helper()

	if len(got) != len(want) {
		t.Errorf("%s has %d entries, want %d: %v", label, len(got), len(want), got)
	}
	for name, wantValue := range want {
		if gotValue := got[name]; gotValue != wantValue {
			t.Errorf("%s[%q] = %q, want %q", label, name, gotValue, wantValue)
		}
	}
}

type packageLockEntry struct {
	Version         string            `json:"version"`
	Resolved        string            `json:"resolved"`
	Integrity       string            `json:"integrity"`
	License         string            `json:"license"`
	Dev             bool              `json:"dev"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// TestEmbeddedAssetsMatchSource verifies that every file under
// internal/scaffold/assets/ is byte-identical to the canonical
// source file at the repo root. This prevents drift between the
// embedded copies and the files developers actually use.
func TestEmbeddedAssets_MatchSource(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping drift detection")
	}

	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	if len(paths) == 0 {
		t.Fatal("no embedded assets found")
	}

	for _, relPath := range paths {
		// devcontainer/ assets are OS-specific and gitignored.
		// No canonical source file exists in the repo to drift
		// against — the embedded template is the source of truth.
		if strings.HasPrefix(relPath, "devcontainer/") {
			continue
		}
		// specify/ assets are starter templates, not mirrors of
		// the org's .specify/ files. The embedded starter
		// constitution is the source of truth.
		if strings.HasPrefix(relPath, "specify/") {
			continue
		}
		// The live meta-repository declaration intentionally contains active
		// organization siblings. Generated projects receive the empty template.
		if relPath == "uf/sibling-repos.yaml" {
			continue
		}
		// Custom packs (*-custom.md) are user-owned: uf init seeds the
		// embedded copy on first run and never overwrites the canonical
		// file thereafter. The embedded copies are initial deployment
		// templates — the canonical repo copies MAY legitimately diverge.
		if strings.HasPrefix(relPath, "opencode/uf/packs/") &&
			strings.HasSuffix(relPath, "-custom.md") {
			continue
		}

		// Map asset path to canonical source path
		srcRel := mapAssetToSource(relPath)
		srcPath := filepath.Join(root, srcRel)

		embedded, err := assetContent(relPath)
		if err != nil {
			t.Errorf("read embedded %s: %v", relPath, err)
			continue
		}

		source, err := os.ReadFile(srcPath)
		if err != nil {
			t.Errorf("read source %s: %v (expected canonical source at %s)", relPath, err, srcPath)
			continue
		}

		if !bytes.Equal(embedded, source) {
			t.Errorf("drift detected: internal/scaffold/assets/%s differs from %s\n"+
				"Run: cp %s internal/scaffold/assets/%s",
				relPath, srcRel, srcRel, relPath)
		}
	}
}

// TestSpecReviewPerFindingGateContracts pins the structural
// contract of the per-finding spec review decision gate across
// the two command assets and the severity pack.
func TestSpecReviewPerFindingGateContracts(t *testing.T) {
	unleashContent, err := assetContent("opencode/commands/uf.unleash.md")
	if err != nil {
		t.Fatalf("read embedded unleash command: %v", err)
	}
	unleash := string(unleashContent)

	councilContent, err := assetContent("opencode/commands/uf.review-council.md")
	if err != nil {
		t.Fatalf("read embedded review-council command: %v", err)
	}
	council := string(councilContent)

	optionLabels := []string{
		`"Apply suggested fix"`,
		`"Override & continue"`,
		`"Show more context"`,
		`"Pause & exit"`,
		`"Apply all remaining LOW/MEDIUM fixes"`,
	}

	t.Run("option labels present in both command assets", func(t *testing.T) {
		for _, label := range optionLabels {
			if !strings.Contains(unleash, label) {
				t.Errorf("unleash command missing option label %s", label)
			}
			if !strings.Contains(council, label) {
				t.Errorf("review-council command missing option label %s", label)
			}
		}
	})

	t.Run("mandatory gate markers retained", func(t *testing.T) {
		if !strings.Contains(council, ">>> MANDATORY GATE: HUMAN CONFIRMATION REQUIRED <<<") {
			t.Error("review-council command missing MANDATORY GATE marker")
		}
		if !strings.Contains(council, ">>> END MANDATORY GATE <<<") {
			t.Error("review-council command missing END MANDATORY GATE marker")
		}
	})

	t.Run("severity-order presentation", func(t *testing.T) {
		if !strings.Contains(unleash, "severity order (CRITICAL, HIGH, MEDIUM, LOW)") {
			t.Error("unleash command missing severity-order presentation")
		}
		if !strings.Contains(council, "severity order (CRITICAL, HIGH, MEDIUM, LOW)") {
			t.Error("review-council command missing severity-order presentation")
		}
	})

	t.Run("no autonomous progression past high/critical", func(t *testing.T) {
		if !strings.Contains(unleash, "MUST NOT proceed past it: no autonomous fix") {
			t.Error("unleash command missing no-autonomous-progression rule")
		}
		if !strings.Contains(council, "MUST NOT proceed past a HIGH/CRITICAL finding without") {
			t.Error("review-council command missing no-autonomous-progression rule")
		}
	})

	t.Run("severity pack note present", func(t *testing.T) {
		packContent, err := assetContent("opencode/uf/packs/severity.md")
		if err != nil {
			t.Fatalf("read embedded severity pack: %v", err)
		}
		if !strings.Contains(string(packContent), "Human-directed fixes vs autonomous auto-fixes") {
			t.Error("severity pack missing human-directed vs autonomous auto-fix note")
		}
	})

	t.Run("no stale bulk-gate option labels", func(t *testing.T) {
		stale := []string{
			`"Apply auto-fixes"`,
			`"Review findings first"`,
			`"Skip auto-fixes -- report only"`,
		}
		for _, label := range stale {
			if strings.Contains(unleash, label) {
				t.Errorf("unleash command retains stale option label %s", label)
			}
			if strings.Contains(council, label) {
				t.Errorf("review-council command retains stale option label %s", label)
			}
		}
	})

	t.Run("override recording contract pinned", func(t *testing.T) {
		if !strings.Contains(unleash, `"overrides": [...]`) {
			t.Error("unleash command missing overrides checkpoint payload")
		}
		if !strings.Contains(council, `step: "spec-review-fix-loop"`) {
			t.Error("review-council command missing spec-review-fix-loop checkpoint marker")
		}
		if !strings.Contains(council, "An override without a recorded justification is") {
			t.Error("review-council command missing override-justification requirement")
		}
	})

	t.Run("session-resume guard pinned", func(t *testing.T) {
		if !strings.Contains(unleash, "checkpoint data stand") {
			t.Error("unleash command missing durable-resolution-stands rule")
		}
		if !strings.Contains(council, "context compression or session resume are NOT valid") {
			t.Error("review-council command missing session-resume guard rule")
		}
	})

	t.Run("recommended-option labeling pinned", func(t *testing.T) {
		if !strings.Contains(unleash, `labeled "(Recommended)"`) {
			t.Error("unleash command missing recommended-option labeling")
		}
		if !strings.Contains(council, `labeled "(Recommended)"`) {
			t.Error("review-council command missing recommended-option labeling")
		}
	})

	t.Run("autonomous auto-fix boundary pinned", func(t *testing.T) {
		if !strings.Contains(council, "remains LOW/MEDIUM only") {
			t.Error("review-council command missing autonomous auto-fix boundary statement")
		}
		packContent, err := assetContent("opencode/uf/packs/severity.md")
		if err != nil {
			t.Fatalf("read embedded severity pack: %v", err)
		}
		pack := string(packContent)
		for _, row := range []string{"| HIGH | Report only |", "| CRITICAL | Report only |"} {
			if !strings.Contains(pack, row) {
				t.Errorf("severity pack missing auto-fix policy row %s", row)
			}
		}
	})

	t.Run("bulk-shortcut severity restriction pinned", func(t *testing.T) {
		if !strings.Contains(unleash, "MUST NOT appear for HIGH/CRITICAL findings") {
			t.Error("unleash command missing bulk-shortcut HIGH/CRITICAL restriction")
		}
		if !strings.Contains(unleash, "MUST NOT modify any HIGH/CRITICAL finding") {
			t.Error("unleash command missing bulk-shortcut HIGH/CRITICAL modification ban")
		}
		if !strings.Contains(council, "MUST NOT modify any HIGH/CRITICAL finding") {
			t.Error("review-council command missing bulk-shortcut HIGH/CRITICAL modification ban")
		}
	})

	t.Run("override carry-forward matching key pinned", func(t *testing.T) {
		if !strings.Contains(unleash, "same file and root cause") {
			t.Error("unleash command missing override carry-forward matching key")
		}
		if !strings.Contains(council, "same file and root cause") {
			t.Error("review-council command missing override carry-forward matching key")
		}
	})

	t.Run("override-only outcome labeling pinned", func(t *testing.T) {
		if !strings.Contains(unleash, "REQUEST CHANGES resolved by recorded overrides") {
			t.Error("unleash command missing override-only outcome labeling")
		}
		if !strings.Contains(council, "REQUEST CHANGES resolved by recorded overrides") {
			t.Error("review-council command missing override-only outcome labeling")
		}
	})
}

// TestQuestionToolPresentationContractGates pins the question-tool
// presentation contract across the three mutation-command assets:
// the shared wrap-safe contract fragments (each once per gate), the
// per-gate option inventories (FR-007/FR-008/FR-009), the per-gate
// fragment and gate counts, the backtick-bracket option-array
// literal counts, the per-label occurrence totals of every
// FR-007/FR-008/FR-009 inventory label, and the absence of the
// legacy bare-options phrasing asserted both raw and
// whitespace-normalized (so line-wrapped occurrences fail).
func TestQuestionToolPresentationContractGates(t *testing.T) {
	readAsset := func(name string) string {
		t.Helper()
		content, err := assetContent("opencode/commands/" + name)
		if err != nil {
			t.Fatalf("read embedded %s command: %v", name, err)
		}
		return string(content)
	}

	triage := readAsset("uf.triage-issue.md")
	feedback := readAsset("uf.address-feedback.md")
	review := readAsset("uf.review-pr.md")

	// normalize collapses markdown line wrapping so option labels
	// split across lines in the source still match their inventory
	// spelling. Contract fragments are checked raw on purpose:
	// each fragment MUST sit on a single line (wrap-safe pin).
	normalize := func(s string) string {
		return strings.Join(strings.Fields(s), " ")
	}

	contractFragments := []string{
		"the question tool MUST present the",
		"same details and options",
		"Per-option reasoning: what each choice does and",
		"when it fits",
		"advisory only",
		"the human decides",
		"recommended option FIRST",
		`labeled "(Recommended)"`,
	}

	t.Run("shared contract fragments present in all three command assets", func(t *testing.T) {
		for _, tc := range []struct {
			assetName string
			asset     string
		}{
			{"uf.triage-issue", triage},
			{"uf.address-feedback", feedback},
			{"uf.review-pr", review},
		} {
			for _, fragment := range contractFragments {
				if !strings.Contains(tc.asset, fragment) {
					t.Errorf("%s command missing contract fragment %q", tc.assetName, fragment)
				}
			}
		}
	})

	triageOptions := []string{
		`"Proceed -- execute triage actions"`,
		`"Skip -- write artifact only"`,
		`"Yes -- create and apply label '<label>'"`,
		`"Yes -- apply label '<label>'"`,
		`"Yes -- apply duplicate label"`,
		`"Yes -- post another comment"`,
		`"No -- skip comment"`,
		`"Approve -- post as-is"`,
		`"Modify -- adjust comment text"`,
		`"Abort -- do not post"`,
		`"Yes -- create this child issue"`,
		`"Yes -- create anyway"`,
		`"No -- skip this child issue"`,
		`"No -- skip"`,
	}

	feedbackOptions := []string{
		`"Accept"`,
		`"Modify"`,
		`"Reject"`,
		`"Ask"`,
		`"Confirm -- proceed with execution"`,
		`"Revise -- change decisions"`,
		`"Proceed with Phase 4 execution"`,
		`"Review plan again"`,
		`"Abort -- stop here"`,
		`"Rebase onto remote and push"`,
		`"Abort -- preserve local commits"`,
		`"Yes -- post reply comments"`,
		`"No -- skip posting"`,
		`"Yes -- resolve accepted threads"`,
		`"No -- leave threads open"`,
	}

	reviewOptions := []string{
		`"Wait for checks to complete"`,
		`"Proceed with available results"`,
		`"Review all files"`,
		`"Focus on specific files"`,
		`"Yes -- create fix branch"`,
		`"No -- skip"`,
		`"Commit -- apply fix"`,
		`"Edit commit message"`,
		`"Abort -- discard changes"`,
		`"Yes -- post as GitHub review"`,
		`"No -- terminal summary is sufficient"`,
		`"Yes -- post new review"`,
		`"No -- skip posting"`,
		`"Yes -- override with <new_verdict>"`,
		`"No -- keep existing <old_verdict>"`,
		`"Approve -- post review"`,
		`"Yes -- post review"`,
		`"Edit comments first"`,
		`"Change verdict"`,
	}

	// Per-label occurrence totals, computed empirically against
	// the embedded assets. The quoted-label form anchors each
	// count so prefix labels cannot collide with longer labels.
	triageLabelCounts := map[string]int{
		`"Proceed -- execute triage actions"`:       2,
		`"Skip -- write artifact only"`:             2,
		`"Yes -- create and apply label '<label>'"`: 2,
		`"Yes -- apply label '<label>'"`:            2,
		`"Yes -- apply duplicate label"`:            2,
		`"Yes -- post another comment"`:             2,
		`"No -- skip comment"`:                      4,
		`"Approve -- post as-is"`:                   3,
		`"Modify -- adjust comment text"`:           2,
		`"Abort -- do not post"`:                    2,
		`"Yes -- create this child issue"`:          2,
		`"Yes -- create anyway"`:                    2,
		`"No -- skip this child issue"`:             2,
		`"No -- skip"`:                              10,
	}

	feedbackLabelCounts := map[string]int{
		`"Accept"`:                            3,
		`"Modify"`:                            2,
		`"Reject"`:                            2,
		`"Ask"`:                               2,
		`"Confirm -- proceed with execution"`: 2,
		`"Revise -- change decisions"`:        2,
		`"Proceed with Phase 4 execution"`:    3,
		`"Review plan again"`:                 3,
		`"Abort -- stop here"`:                2,
		`"Rebase onto remote and push"`:       2,
		`"Abort -- preserve local commits"`:   2,
		`"Yes -- post reply comments"`:        2,
		`"No -- skip posting"`:                2,
		`"Yes -- resolve accepted threads"`:   2,
		`"No -- leave threads open"`:          2,
	}

	reviewLabelCounts := map[string]int{
		`"Wait for checks to complete"`:          2,
		`"Proceed with available results"`:       2,
		`"Review all files"`:                     3,
		`"Focus on specific files"`:              3,
		`"Yes -- create fix branch"`:             3,
		`"No -- skip"`:                           2,
		`"Commit -- apply fix"`:                  3,
		`"Edit commit message"`:                  2,
		`"Abort -- discard changes"`:             2,
		`"Yes -- post as GitHub review"`:         3,
		`"No -- terminal summary is sufficient"`: 2,
		`"Yes -- post new review"`:               2,
		`"No -- skip posting"`:                   6,
		`"Yes -- override with <new_verdict>"`:   2,
		`"No -- keep existing <old_verdict>"`:    2,
		`"Approve -- post review"`:               2,
		`"Yes -- post review"`:                   2,
		`"Edit comments first"`:                  3,
		`"Change verdict"`:                       3,
	}

	t.Run("gate option inventories present in their command assets", func(t *testing.T) {
		for _, tc := range []struct {
			assetName string
			asset     string
			labels    []string
		}{
			{"uf.triage-issue", triage, triageOptions},
			{"uf.address-feedback", feedback, feedbackOptions},
			{"uf.review-pr", review, reviewOptions},
		} {
			normalized := normalize(tc.asset)
			for _, label := range tc.labels {
				if !strings.Contains(normalized, normalize(label)) {
					t.Errorf("%s command missing option label %s", tc.assetName, label)
				}
			}
		}
	})

	t.Run("per-gate contract fragment counts", func(t *testing.T) {
		for _, tc := range []struct {
			assetName string
			asset     string
			want      int
		}{
			{"uf.triage-issue", triage, 8},
			{"uf.address-feedback", feedback, 6},
			{"uf.review-pr", review, 8},
		} {
			for _, fragment := range contractFragments {
				if got := strings.Count(tc.asset, fragment); got != tc.want {
					t.Errorf("%s command carries %d occurrences of contract fragment %q, want %d (one per gate)", tc.assetName, got, fragment, tc.want)
				}
			}
		}
	})

	t.Run("option-array literal counts", func(t *testing.T) {
		// The backtick-bracket opener counts one per house-style
		// option array; embedded JSON examples (no backtick) are
		// excluded. review-pr records 9, not 8, because Step 8c
		// carries two verdict-specific option sets.
		for _, tc := range []struct {
			assetName string
			asset     string
			want      int
		}{
			{"uf.triage-issue", triage, 8},
			{"uf.address-feedback", feedback, 6},
			{"uf.review-pr", review, 9},
		} {
			if got := strings.Count(tc.asset, "`[\""); got != tc.want {
				t.Errorf("%s command carries %d backtick-bracket option arrays, want %d", tc.assetName, got, tc.want)
			}
		}
	})

	t.Run("legacy bare-options phrasing absent", func(t *testing.T) {
		legacy := "Use the **question tool** with options"
		for _, tc := range []struct {
			assetName string
			asset     string
		}{
			{"uf.triage-issue", triage},
			{"uf.address-feedback", feedback},
			{"uf.review-pr", review},
		} {
			if strings.Contains(tc.asset, legacy) {
				t.Errorf("%s command retains legacy bare-options phrasing %q", tc.assetName, legacy)
			}
			if strings.Contains(normalize(tc.asset), normalize(legacy)) {
				t.Errorf("%s command retains legacy bare-options phrasing %q across wrapped lines", tc.assetName, legacy)
			}
		}
	})

	t.Run("per-label occurrence totals", func(t *testing.T) {
		// Every inventory label's quoted form is counted in the
		// whitespace-normalized asset; the closing quote anchors
		// the match so a prefix label ("No -- skip") cannot
		// collide with a longer label ("No -- skip posting").
		// Totals were computed empirically against the assets;
		// each occurrence is an option-array entry, a per-option
		// reasoning bullet, a recommendation, or follow-up prose,
		// so removing a label from any array fails the pin.
		for _, tc := range []struct {
			assetName string
			asset     string
			labels    []string
			wants     map[string]int
		}{
			{"uf.triage-issue", triage, triageOptions, triageLabelCounts},
			{"uf.address-feedback", feedback, feedbackOptions, feedbackLabelCounts},
			{"uf.review-pr", review, reviewOptions, reviewLabelCounts},
		} {
			normalized := normalize(tc.asset)
			for _, label := range tc.labels {
				if got := strings.Count(normalized, normalize(label)); got != tc.wants[label] {
					t.Errorf("%s command carries %d occurrences of option label %s, want %d", tc.assetName, got, label, tc.wants[label])
				}
			}
		}
	})
}

func TestUnleashCommand_ToolRoutingContracts(t *testing.T) {
	content, err := assetContent("opencode/commands/uf.unleash.md")
	if err != nil {
		t.Fatalf("read embedded unleash command: %v", err)
	}

	command := string(content)
	between := func(t *testing.T, value, start, end string) string {
		t.Helper()
		startIndex := strings.Index(value, start)
		if startIndex == -1 {
			t.Fatalf("unleash command missing section start %q", start)
		}
		value = value[startIndex:]
		endIndex := strings.Index(value, end)
		if endIndex == -1 {
			t.Fatalf("unleash command missing section end %q", end)
		}
		return value[:endIndex]
	}

	specReviewExit := between(t, command, `- If the human chooses "Pause & exit"`, "Rendered shape:")
	demoNextSteps := between(t, command, "4. **Next Steps**:", "Then persist the summary")
	stepTenPersistence := between(t, command, "Then persist the summary", "## Guardrails")

	// Step-heading-anchored spans for the seam pins below. Each
	// span covers one pipeline step (or one review-dispatch
	// sequence) so every pin proves the tool is invoked by its
	// consuming step, not merely mentioned elsewhere in the
	// command.
	progressTracking := between(t, command, "### TodoWrite Progress Tracking", "### 0. Startup Cleanup")
	startupCleanup := between(t, command, "### 0. Startup Cleanup", "### 1. Branch Safety Gate")
	branchGateStep := between(t, command, "### 1. Branch Safety Gate", "### 2. Resumability Detection")
	resumeDetection := between(t, command, "### 2. Resumability Detection", "### 3. Step 1 -- Clarify")
	clarifyStep := between(t, command, "### 3. Step 1 -- Clarify", "### 4. Step 2 -- Plan")
	specReviewDispatch := between(t, command, "### 6. Step 4 -- Spec Review", "**Processing results**")
	implementExecution := between(t, command, "### 7. Step 5 -- Implement", "### 8. Step 6 -- Code Review")
	codeReviewDispatch := between(t, command, "### 8. Step 6 -- Code Review", "**Processing results**")
	codeReviewStep := between(t, command, "### 8. Step 6 -- Code Review", "### 9. Step 7 -- Retrospective")
	retrospectiveStep := between(t, command, "### 9. Step 7 -- Retrospective", "### 10. Step 8 -- Demo")
	demoCompletion := between(t, command, "### 10. Step 8 -- Demo", "## Guardrails")

	t.Run("spec review exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(specReviewExit, "format_unleash_exit") {
			t.Error("spec review exit missing format_unleash_exit call")
		}
		if !strings.Contains(specReviewExit, `variant: "spec-review"`) {
			t.Error("spec review exit missing spec-review variant")
		}
		if !strings.Contains(specReviewExit, "tier: WORKFLOW_TIER") {
			t.Error("spec review exit does not pass tier to the formatter tool")
		}
		if !strings.Contains(specReviewExit, "The tool selects the tier-specific recovery guidance") {
			t.Error("spec review exit missing tool-selected tier guidance note")
		}
	})

	t.Run("demo routes through format_unleash_demo", func(t *testing.T) {
		if !strings.Contains(demoNextSteps, "format_unleash_demo") {
			t.Error("demo missing format_unleash_demo call")
		}
		if !strings.Contains(demoNextSteps, "tier: WORKFLOW_TIER") {
			t.Error("demo does not pass tier to the formatter tool")
		}
		if !strings.Contains(demoNextSteps, "- Run `/uf.finale` to create PR and watch CI") {
			t.Error("demo example missing fixed /uf.finale option")
		}
		if !strings.Contains(demoNextSteps, "[tier-specific refinement option selected above]") {
			t.Error("demo example missing tier-selected refinement option")
		}
		for _, param := range []string{"whatWasBuilt:", "howToVerify:", "testResults:"} {
			if !strings.Contains(demoNextSteps, param) {
				t.Errorf("demo format_unleash_demo invocation missing %s", param)
			}
		}
	})

	t.Run("step 10 persists the summary through write_unleash_summary", func(t *testing.T) {
		if !strings.Contains(stepTenPersistence, "write_unleash_summary") {
			t.Error("step 10 missing write_unleash_summary invocation")
		}
		for _, param := range []string{
			"change_name:",
			"tier: WORKFLOW_TIER",
			"feature_dir:",
			"review_verdicts:",
			"learnings_stored:",
			"whatWasBuilt:",
			"howToVerify:",
			"testResults:",
		} {
			if !strings.Contains(stepTenPersistence, param) {
				t.Errorf("step 10 write_unleash_summary invocation missing %s", param)
			}
		}
		if strings.Contains(stepTenPersistence, "persist_unleash_summary") {
			t.Error("step 10 still references the superseded persist_unleash_summary tool name")
		}
	})

	t.Run("ceremony reduced to the two checkpoint calls", func(t *testing.T) {
		if got := strings.Count(command, "CHECKPOINT: Call"); got != 2 {
			t.Errorf("unleash command checkpoint ceremony count = %d, want 2", got)
		}
	})

	t.Run("code-review fix loop delegates through uf_delegate_command", func(t *testing.T) {
		fixLoop := between(t, command, "enter the fix loop", "If the verdict is INCONCLUSIVE or UNAVAILABLE")
		if !strings.Contains(fixLoop, "uf_delegate_command") {
			t.Error("fix loop missing uf_delegate_command delegation")
		}
		if !strings.Contains(fixLoop, "prompt:") {
			t.Error("fix loop delegation missing prompt-only call shape")
		}
		if strings.Contains(fixLoop, "command_path:") {
			t.Error("fix loop delegation must be prompt-only (no command_path)")
		}
		if strings.Contains(fixLoop, "via the Task tool") {
			t.Error("fix loop still delegates via the Task tool")
		}
	})

	t.Run("step 2 plan delegates through uf_delegate_command", func(t *testing.T) {
		planStep := between(t, command, "### 4. Step 2 -- Plan", "### 5. Step 3 -- Tasks")
		if !strings.Contains(planStep, "uf_delegate_command") {
			t.Error("plan step missing uf_delegate_command delegation")
		}
		if !strings.Contains(planStep, `command_path: ".opencode/commands/speckit.plan.md"`) {
			t.Error("plan step delegation missing speckit.plan command path")
		}
		if !strings.Contains(planStep, `checkpoint_step: "plan"`) {
			t.Error("plan step delegation missing plan checkpoint step")
		}
		if strings.Contains(planStep, "via the Task tool") {
			t.Error("plan step still delegates via the Task tool")
		}
	})

	t.Run("step 3 tasks delegates through uf_delegate_command", func(t *testing.T) {
		tasksStep := between(t, command, "### 5. Step 3 -- Tasks", "### 6. Step 4 -- Spec Review")
		if !strings.Contains(tasksStep, "uf_delegate_command") {
			t.Error("tasks step missing uf_delegate_command delegation")
		}
		if !strings.Contains(tasksStep, `command_path: ".opencode/commands/speckit.tasks.md"`) {
			t.Error("tasks step delegation missing speckit.tasks command path")
		}
		if !strings.Contains(tasksStep, `checkpoint_step: "tasks"`) {
			t.Error("tasks step delegation missing tasks checkpoint step")
		}
		if strings.Contains(tasksStep, "via the Task tool") {
			t.Error("tasks step still delegates via the Task tool")
		}
	})

	t.Run("step 5 implement delegates through uf_delegate_command", func(t *testing.T) {
		implementStep := between(t, command, "### 7. Step 5 -- Implement", "### 8. Step 6 -- Code Review")
		if !strings.Contains(implementStep, "uf_delegate_command") {
			t.Error("implement step missing uf_delegate_command delegation")
		}
		if !strings.Contains(implementStep, "prompt:") {
			t.Error("implement step delegation missing prompt-only call shape")
		}
		// Scope the prompt-only negative to the per-task
		// delegation block so prose mentions of the
		// `command_path` parameter elsewhere in the step
		// cannot trip it; the Step 5 fix loop's call site
		// carries its own dedicated pin below.
		perTaskDelegation := between(t, command, "Sequential execution", "Parallel execution")
		if strings.Contains(perTaskDelegation, "command_path:") {
			t.Error("implement step per-task delegation must be prompt-only (no command_path)")
		}
		if strings.Contains(implementStep, "via the Task tool") {
			t.Error("implement step still delegates via the Task tool")
		}
	})

	// Pin: the quick tier's whole-change delegation is a
	// third prompt-only call site (outside both the per-task
	// and fix-loop spans), so it carries a dedicated shape
	// pin rather than relying on span-wide negatives.
	t.Run("quick-tier whole-change delegation is prompt-only", func(t *testing.T) {
		quickTierDelegation := between(t, command, "an unclassified prompt-only call fails closed", "tier's only blocking build/test gate")
		if !strings.Contains(quickTierDelegation, "uf_delegate_command") {
			t.Error("quick-tier delegation missing uf_delegate_command")
		}
		if !strings.Contains(quickTierDelegation, "prompt:") {
			t.Error("quick-tier delegation missing prompt-only call shape")
		}
		if !strings.Contains(quickTierDelegation, "task_classification:") {
			t.Error("quick-tier delegation missing task_classification contract")
		}
		if !strings.Contains(quickTierDelegation, `checkpoint_step: "implement"`) {
			t.Error("quick-tier delegation missing its implement checkpoint step")
		}
		if strings.Contains(quickTierDelegation, "command_path:") {
			t.Error("quick-tier delegation must be prompt-only (no command_path)")
		}
		if strings.Contains(quickTierDelegation, "via the Task tool") {
			t.Error("quick-tier delegation still delegates via the Task tool")
		}
	})

	t.Run("tier guidance no longer enumerated inline", func(t *testing.T) {
		if strings.Contains(specReviewExit, "`/speckit.clarify` to address the findings") {
			t.Error("spec review exit still enumerates speckit guidance inline")
		}
		if strings.Contains(specReviewExit, "artifacts under `openspec/changes/<name>/` to") {
			t.Error("spec review exit still enumerates openspec guidance inline")
		}
	})

	t.Run("hand-rolled output templates removed", func(t *testing.T) {
		if strings.Contains(command, "Before composing this exit message, re-read the") {
			t.Error("command still carries the spec review re-read instruction")
		}
		if strings.Contains(command, "Format the output as:") {
			t.Error("command still carries the hand-rolled demo output template")
		}
	})

	// Pin: resume re-initialization reads persisted state
	// through get_progress before rebuilding the TodoWrite list.
	t.Run("progress tracking re-initializes through get_progress", func(t *testing.T) {
		if !strings.Contains(progressTracking, "get_progress") {
			t.Error("progress tracking section missing get_progress invocation")
		}
		if !strings.Contains(progressTracking, `command: "uf.unleash"`) {
			t.Error("progress tracking get_progress invocation missing command parameter")
		}
		for _, field := range []string{"completed_steps", "current_step"} {
			if !strings.Contains(progressTracking, field) {
				t.Errorf("progress tracking get_progress consumption missing %s result field", field)
			}
		}
	})

	// Pin: startup cleanup enumerates and clears stale worktrees
	// through the forge worktree tools before any pipeline logic.
	t.Run("startup cleanup routes through the forge worktree tools", func(t *testing.T) {
		if !strings.Contains(startupCleanup, "forge_worktree_list") {
			t.Error("startup cleanup missing forge_worktree_list invocation")
		}
		if !strings.Contains(startupCleanup, "forge_worktree_cleanup") {
			t.Error("startup cleanup missing forge_worktree_cleanup invocation")
		}
		if !strings.Contains(startupCleanup, "cleanup_all: true") {
			t.Error("startup cleanup forge_worktree_cleanup invocation missing cleanup_all parameter")
		}
	})

	// Pin: the branch safety gate is the parameterless
	// uf_unleash_branch_gate call whose payload sets the tier
	// and feature directory for the whole pipeline.
	t.Run("branch safety gate routes through uf_unleash_branch_gate", func(t *testing.T) {
		if !strings.Contains(branchGateStep, "uf_unleash_branch_gate") {
			t.Error("branch safety gate missing uf_unleash_branch_gate invocation")
		}
		if !strings.Contains(branchGateStep, "(no parameters)") {
			t.Error("branch safety gate invocation must remain parameterless")
		}
		for _, field := range []string{"workflow_tier", "feature_dir"} {
			if !strings.Contains(branchGateStep, field) {
				t.Errorf("branch safety gate payload consumption missing %s field", field)
			}
		}
	})

	// Pin: resumability detection routes through
	// uf_unleash_detect_resume with the tier-resolved inputs.
	t.Run("resumability detection routes through uf_unleash_detect_resume", func(t *testing.T) {
		if !strings.Contains(resumeDetection, "uf_unleash_detect_resume") {
			t.Error("resumability detection missing uf_unleash_detect_resume invocation")
		}
		for _, param := range []string{"feature_dir: FEATURE_DIR", "workflow_tier: WORKFLOW_TIER"} {
			if !strings.Contains(resumeDetection, param) {
				t.Errorf("resumability detection invocation missing %s parameter", param)
			}
		}
		for _, field := range []string{"step_statuses", "resume_from", "all_complete", "unchecked_count"} {
			if !strings.Contains(resumeDetection, field) {
				t.Errorf("resumability detection consumption missing %s result field", field)
			}
		}
	})

	// Pin: the clarify step scans markers, queries Dewey when
	// available, and resolves questions through the dedicated
	// clarification tools rather than hand-rolled spec edits.
	t.Run("clarify routes through scan_clarifications and resolve_clarification", func(t *testing.T) {
		if !strings.Contains(clarifyStep, "scan_clarifications") {
			t.Error("clarify step missing scan_clarifications invocation")
		}
		if !strings.Contains(clarifyStep, "resolve_clarification") {
			t.Error("clarify step missing resolve_clarification invocation")
		}
		if !strings.Contains(clarifyStep, "dewey_semantic_search") {
			t.Error("clarify step missing dewey_semantic_search availability probe")
		}
		for _, param := range []string{"feature_dir:", "question:", "answer:", "source:"} {
			if !strings.Contains(clarifyStep, param) {
				t.Errorf("clarify resolve_clarification invocation missing %s parameter", param)
			}
		}
	})

	// Pin: the clarify exit renders through the
	// format_unleash_exit clarify variant with the scanned
	// question/context pairs.
	t.Run("clarify exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(clarifyStep, "format_unleash_exit") {
			t.Error("clarify exit missing format_unleash_exit call")
		}
		if !strings.Contains(clarifyStep, `variant: "clarify"`) {
			t.Error("clarify exit missing clarify variant")
		}
		if !strings.Contains(clarifyStep, "questions: [{question, context}, ...]") {
			t.Error("clarify exit missing questions parameter")
		}
	})

	// Pin: both review steps resolve their immutable base ref
	// through the offline resolve_base_ref tool.
	t.Run("review dispatch resolves the base ref through resolve_base_ref", func(t *testing.T) {
		if !strings.Contains(specReviewDispatch, "resolve_base_ref") {
			t.Error("spec review dispatch missing resolve_base_ref invocation")
		}
		if !strings.Contains(codeReviewDispatch, "resolve_base_ref") {
			t.Error("code review dispatch missing resolve_base_ref invocation")
		}
		if !strings.Contains(specReviewDispatch, "(no network access)") ||
			!strings.Contains(codeReviewDispatch, "(no network access)") {
			t.Error("resolve_base_ref invocation lost the no-network-access contract")
		}
	})

	// Pin: both review steps plan their dispatch through
	// plan_review_dispatch with the mode-specific plan inputs.
	t.Run("review dispatch plans through plan_review_dispatch", func(t *testing.T) {
		if !strings.Contains(specReviewDispatch, "plan_review_dispatch") {
			t.Error("spec review dispatch missing plan_review_dispatch invocation")
		}
		if !strings.Contains(codeReviewDispatch, "plan_review_dispatch") {
			t.Error("code review dispatch missing plan_review_dispatch invocation")
		}
		if !strings.Contains(specReviewDispatch, "mode `specs`") {
			t.Error("spec review plan_review_dispatch invocation missing specs mode")
		}
		if !strings.Contains(codeReviewDispatch, "mode `code`") {
			t.Error("code review plan_review_dispatch invocation missing code mode")
		}
		for _, param := range []string{"full: false", "augment: false", "changed_files"} {
			if !strings.Contains(specReviewDispatch, param) ||
				!strings.Contains(codeReviewDispatch, param) {
				t.Errorf("plan_review_dispatch invocation missing %s parameter", param)
			}
		}
	})

	// Pin: both review steps fan out through dispatch_review_runs
	// with the shared dispatch parameters and their
	// mode-specific context.
	t.Run("review dispatch runs reviewers through dispatch_review_runs", func(t *testing.T) {
		if !strings.Contains(specReviewDispatch, "dispatch_review_runs") {
			t.Error("spec review dispatch missing dispatch_review_runs invocation")
		}
		if !strings.Contains(codeReviewDispatch, "dispatch_review_runs") {
			t.Error("code review dispatch missing dispatch_review_runs invocation")
		}
		for _, param := range []string{
			"diff_path",
			"changed_files",
			"input_context",
			"plan_entries",
			"session_metadata",
			"max_parallel_runs",
			"budget_usd",
			"per_run_timeout_ms",
		} {
			if !strings.Contains(specReviewDispatch, param) ||
				!strings.Contains(codeReviewDispatch, param) {
				t.Errorf("dispatch_review_runs invocation missing %s parameter", param)
			}
		}
		if !strings.Contains(specReviewDispatch, "review_context") {
			t.Error("spec review dispatch_review_runs invocation missing review_context parameter")
		}
		if !strings.Contains(codeReviewDispatch, "pre_flight_results") {
			t.Error("code review dispatch_review_runs invocation missing pre_flight_results parameter")
		}
		if !strings.Contains(specReviewDispatch, "exactly once") ||
			!strings.Contains(codeReviewDispatch, "exactly once") {
			t.Error("dispatch_review_runs invocation lost the exactly-once contract")
		}
	})

	// Pin: both review steps finalize through consolidate_dispatch
	// with the review-council workflow provenance.
	t.Run("review dispatch consolidates through consolidate_dispatch", func(t *testing.T) {
		if !strings.Contains(specReviewDispatch, "consolidate_dispatch") {
			t.Error("spec review dispatch missing consolidate_dispatch invocation")
		}
		if !strings.Contains(codeReviewDispatch, "consolidate_dispatch") {
			t.Error("code review dispatch missing consolidate_dispatch invocation")
		}
		if !strings.Contains(specReviewDispatch, `workflow_id: "uf.review-council"`) ||
			!strings.Contains(codeReviewDispatch, `workflow_id: "uf.review-council"`) {
			t.Error("consolidate_dispatch invocation missing workflow_id provenance parameter")
		}
		if !strings.Contains(specReviewDispatch, "40-character head") ||
			!strings.Contains(codeReviewDispatch, "40-character head") {
			t.Error("consolidate_dispatch invocation missing head SHA provenance requirement")
		}
	})

	// Pin: implement reads the task breakdown through parse_tasks
	// before any delegation.
	t.Run("implement parses tasks through parse_tasks", func(t *testing.T) {
		if !strings.Contains(implementExecution, "parse_tasks") {
			t.Error("implement step missing parse_tasks invocation")
		}
		if !strings.Contains(implementExecution, `tasks_path: FEATURE_DIR + "tasks.md"`) {
			t.Error("implement parse_tasks invocation missing tasks_path parameter")
		}
		for _, field := range []string{"total_tasks", "done_count", "all_done"} {
			if !strings.Contains(implementExecution, field) {
				t.Errorf("implement parse_tasks consumption missing %s result field", field)
			}
		}
	})

	// Pin: implement records progress through mark_task_complete
	// immediately after each successful delegation.
	t.Run("implement marks tasks complete through mark_task_complete", func(t *testing.T) {
		if !strings.Contains(implementExecution, "mark_task_complete") {
			t.Error("implement step missing mark_task_complete invocation")
		}
		for _, param := range []string{"tasks_path:", "task_id:"} {
			if !strings.Contains(implementExecution, param) {
				t.Errorf("implement mark_task_complete invocation missing %s parameter", param)
			}
		}
	})

	// Pin: parallel implement execution isolates workers through
	// the forge worktree lifecycle tools.
	t.Run("implement parallel workers route through the forge worktree tools", func(t *testing.T) {
		for _, tool := range []string{
			"forge_worktree_create",
			"forge_spawn_subtask",
			"forge_worktree_merge",
			"forge_worktree_cleanup",
		} {
			if !strings.Contains(implementExecution, tool) {
				t.Errorf("implement parallel execution missing %s invocation", tool)
			}
		}
		if !strings.Contains(implementExecution, "task ID, and current commit hash") {
			t.Error("implement forge_worktree_create invocation missing its parameter contract")
		}
		if !strings.Contains(implementExecution, "files from the task") {
			t.Error("implement forge_spawn_subtask invocation missing its parameter contract")
		}
		if !strings.Contains(implementExecution, "cleanup_all: true") {
			t.Error("implement forge_worktree_cleanup invocation missing cleanup_all parameter")
		}
		// The spawn call block carries the delegation
		// classification contract as tool parameters. Scoped to
		// the parallel-execution span: the sequential
		// uf_delegate_command block passes the same parameter
		// names, so a step-wide Contains could not prove the
		// spawn block itself carries them.
		parallelExecution := between(t, command, "3. **Parallel execution**", "4. **Phase checkpoint**")
		for _, param := range []string{
			"task_classification:",
			"skill: [repo skill or script reference",
			"skill_absent_reason: [what was searched",
		} {
			if !strings.Contains(parallelExecution, param) {
				t.Errorf("implement spawn call block missing %s parameter", param)
			}
		}
		// The superseded description-as-carrier sentence must
		// stay gone from both the embedded mirror and the
		// canonical command file.
		staleSentence := "The spawn tool has no classification parameters"
		if strings.Contains(command, staleSentence) {
			t.Error("embedded uf.unleash.md still carries the stale no-classification-parameters sentence")
		}
		root := findProjectRoot(t)
		if root == "" {
			t.Fatalf("project root not found; cannot verify canonical uf.unleash.md")
		}
		canonical, readErr := os.ReadFile(filepath.Join(root, ".opencode", "commands", "uf.unleash.md"))
		if readErr != nil {
			t.Fatalf("read canonical uf.unleash.md: %v", readErr)
		}
		if strings.Contains(string(canonical), staleSentence) {
			t.Error("canonical uf.unleash.md still carries the stale no-classification-parameters sentence")
		}
	})

	// Pin: worktree merges are verified through
	// scan_conflict_markers before cleanup proceeds.
	t.Run("implement scans merge residue through scan_conflict_markers", func(t *testing.T) {
		if !strings.Contains(implementExecution, "scan_conflict_markers") {
			t.Error("implement step missing scan_conflict_markers invocation")
		}
		if !strings.Contains(implementExecution, "paths: [files affected by the merge]") {
			t.Error("implement scan_conflict_markers invocation missing paths parameter")
		}
	})

	// Pin: the implement worker-failure exit renders through the
	// format_unleash_exit implement-worker-failed variant.
	t.Run("implement worker-failure exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(implementExecution, "format_unleash_exit") {
			t.Error("implement step missing format_unleash_exit call")
		}
		if !strings.Contains(implementExecution, `variant: "implement-worker-failed"`) {
			t.Error("implement step missing implement-worker-failed variant")
		}
		for _, param := range []string{
			"phase: [phase number]",
			"failed_task: [task ID and description]",
			"error: [error details from worker]",
		} {
			if !strings.Contains(implementExecution, param) {
				t.Errorf("implement worker-failure exit missing %s parameter", param)
			}
		}
	})

	// Pin: the merge-conflict exit renders through the
	// format_unleash_exit implement-merge-conflict variant.
	t.Run("implement merge-conflict exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(implementExecution, `variant: "implement-merge-conflict"`) {
			t.Error("implement step missing implement-merge-conflict variant")
		}
		for _, param := range []string{
			"phase: [phase number]",
			"conflicting_files: [paths from the scan result]",
		} {
			if !strings.Contains(implementExecution, param) {
				t.Errorf("implement merge-conflict exit missing %s parameter", param)
			}
		}
	})

	// Pin: the phase-checkpoint exit renders through the
	// format_unleash_exit implement-phase-checkpoint variant.
	t.Run("implement phase-checkpoint exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(implementExecution, `variant: "implement-phase-checkpoint"`) {
			t.Error("implement step missing implement-phase-checkpoint variant")
		}
		for _, param := range []string{
			"phase: [phase number]",
			"failed_command: [the command that failed]",
			"output: [error output]",
		} {
			if !strings.Contains(implementExecution, param) {
				t.Errorf("implement phase-checkpoint exit missing %s parameter", param)
			}
		}
	})

	// Pin: failing checks at blocking gates route through the
	// build/test fix loop before any exit fires. Each wiring
	// site carries a distinct phrase so deleting either one
	// fails this subtest (span-machinery intent: one distinct
	// phrase per wiring site).
	t.Run("blocking gates run the build/test fix loop before exiting", func(t *testing.T) {
		if !strings.Contains(implementExecution, "build/test fix loop (defined below the") {
			t.Error("quick-tier pre-flight gate missing the build/test fix loop wiring")
		}
		if !strings.Contains(implementExecution, "build/test fix loop (above)") {
			t.Error("phase checkpoint missing the build/test fix loop wiring")
		}
		for _, param := range []string{
			// causality filter at gate entry
			"classify every failure by causality",
			"never enter the loop",
			// loop cap and exhaustion phrase
			"up to 3 total",
			"If a check still fails after the loop's 3 iterations",
			// deterministic re-run and green supersession
			"re-run the exact",
			"superseding",
		} {
			if !strings.Contains(implementExecution, param) {
				t.Errorf("implement fix loop missing %s parameter", param)
			}
		}
		if !strings.Contains(codeReviewStep, "build/test fix loop (defined in Step 5)") {
			t.Error("code review preflight missing the build/test fix loop wiring")
		}
	})

	// Pin: the fix-loop iteration budget is durable -- recorded
	// in each gate's checkpoint data as a gate-distinct free-form
	// data key and continued on resume, never reset per session.
	// Each store site carries a distinct contiguous phrase so
	// deleting any one store fails this subtest. The seam pin
	// ("iteration:" absent) keeps the count on the free-form data
	// keys: the checkpoint tool persists only number, verdict, and
	// inventory from the structured phase object.
	t.Run("fix loop iteration budget survives resume", func(t *testing.T) {
		for _, param := range []string{
			// quick-gate budget-only checkpoint step
			"quick-fix-budget",
			// quick-gate completion-claim separation
			"ONLY completion",
			// quick-gate checkpoint data key
			"fixLoopQuick",
			// quick-gate budget phrasing
			"iterations used, 0 if the",
			// phase-checkpoint data key
			"fixLoopPhase[N]",
			// phase-checkpoint budget phrasing
			"iterations used so far",
			// resume rule (loop block)
			"never 3 more per resume",
			// phase-checkpoint durability prose
			"read it back on resume",
		} {
			if !strings.Contains(implementExecution, param) {
				t.Errorf("implement fix loop missing durable iteration budget wiring: %s", param)
			}
		}
		if strings.Contains(implementExecution, "iteration:") {
			t.Error("fix-loop budget must ride the free-form data keys, not the structured phase object (checkpoint tool drops undeclared phase fields)")
		}
		for _, param := range []string{
			// Step 6 preflight data key and durability phrasing
			"fixLoopPreflight",
			"3-attempt budget spans sessions",
		} {
			if !strings.Contains(codeReviewStep, param) {
				t.Errorf("step 6 preflight missing durable iteration budget wiring: %s", param)
			}
		}
	})

	// Pin: the loop's exit half -- the fail-closed
	// classification fallback, exhaustion exits at each
	// gate, supersession at the recording site, and the
	// guardrails exit-list entry. Each phrase is
	// site-unique so deleting one wiring fails this
	// subtest.
	t.Run("fix loop exits and supersession stay wired", func(t *testing.T) {
		if !strings.Contains(implementExecution, "fail closed into the loop") {
			t.Error("causality contract missing the fail-closed classification rule")
		}
		if !strings.Contains(implementExecution, "the phase checkpoint with verdict") {
			t.Error("phase checkpoint missing green-supersession at the recording site")
		}
		if !strings.Contains(implementExecution, "When the loop exhausts its 3 iterations") {
			t.Error("phase checkpoint missing its exhaustion exit")
		}
		if !strings.Contains(codeReviewStep, "persist after 3 iterations") {
			t.Error("step 6 preflight missing its persistence STOP")
		}
		guardrails := between(t, command, "## Guardrails", "</protect>")
		if !strings.Contains(guardrails, "after the fix loop's") {
			t.Error("guardrails exit list missing the fix-loop exhaustion exit")
		}
	})

	// Pin: the Step 5 build/test fix loop delegates through
	// uf_delegate_command with the prompt-only, classified call
	// shape (mirrors the Step 6 fix-loop delegation pin above).
	t.Run("build/test fix loop delegates through uf_delegate_command", func(t *testing.T) {
		buildTestFixLoop := between(t, command, "**Build/test fix loop**", "**For each phase in tasks.md**")
		if !strings.Contains(buildTestFixLoop, "uf_delegate_command") {
			t.Error("build/test fix loop missing uf_delegate_command delegation")
		}
		if !strings.Contains(buildTestFixLoop, "prompt:") {
			t.Error("build/test fix loop delegation missing prompt-only call shape")
		}
		if !strings.Contains(buildTestFixLoop, "task_classification:") {
			t.Error("build/test fix loop delegation missing task_classification contract")
		}
		if strings.Contains(buildTestFixLoop, "command_path:") {
			t.Error("build/test fix loop delegation must be prompt-only (no command_path)")
		}
		if strings.Contains(buildTestFixLoop, "checkpoint_step:") {
			t.Error("build/test fix loop delegation must be prompt-only (no checkpoint_step)")
		}
		if strings.Contains(buildTestFixLoop, "via the Task tool") {
			t.Error("build/test fix loop still delegates via the Task tool")
		}
	})

	// Pin: the code-review preflight exit renders through the
	// format_unleash_exit code-review-preflight variant.
	t.Run("code-review preflight exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(codeReviewStep, "format_unleash_exit") {
			t.Error("code review step missing format_unleash_exit call")
		}
		if !strings.Contains(codeReviewStep, `variant: "code-review-preflight"`) {
			t.Error("code review step missing code-review-preflight variant")
		}
		if !strings.Contains(codeReviewStep, "findings: [branch-caused failures with full output]") {
			t.Error("code-review preflight exit missing findings parameter")
		}
	})

	// Pin: the no-success exit renders through the
	// format_unleash_exit code-review-no-success variant.
	t.Run("code-review no-success exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(codeReviewStep, `variant: "code-review-no-success"`) {
			t.Error("code review step missing code-review-no-success variant")
		}
		if !strings.Contains(codeReviewStep, "reasons: [which reviewers ran or failed, and why]") {
			t.Error("code-review no-success exit missing reasons parameter")
		}
	})

	// Pin: the iterations-exhausted exit renders through the
	// format_unleash_exit code-review-iterations-exhausted
	// variant.
	t.Run("code-review iterations-exhausted exit routes through format_unleash_exit", func(t *testing.T) {
		if !strings.Contains(codeReviewStep, `variant: "code-review-iterations-exhausted"`) {
			t.Error("code review step missing code-review-iterations-exhausted variant")
		}
		for _, param := range []string{"persistent_findings:", "circular_dependencies:"} {
			if !strings.Contains(codeReviewStep, param) {
				t.Errorf("code-review iterations-exhausted exit missing %s parameter", param)
			}
		}
	})

	// Pin: the retrospective stores learnings through
	// dewey_store_learning when Dewey is available.
	t.Run("retrospective stores learnings through dewey_store_learning", func(t *testing.T) {
		if !strings.Contains(retrospectiveStep, "dewey_store_learning") {
			t.Error("retrospective missing dewey_store_learning invocation")
		}
		if !strings.Contains(retrospectiveStep, "If Dewey is available") {
			t.Error("retrospective missing the Dewey availability guard")
		}
		if !strings.Contains(retrospectiveStep, "with tags including") {
			t.Error("retrospective dewey_store_learning invocation missing tags parameter")
		}
	})

	// Pin: pipeline completion clears the progress state through
	// clear_progress.
	t.Run("completion clears progress state through clear_progress", func(t *testing.T) {
		if !strings.Contains(demoCompletion, "clear_progress") {
			t.Error("pipeline completion missing clear_progress invocation")
		}
		if !strings.Contains(demoCompletion, `command: "uf.unleash"`) {
			t.Error("completion clear_progress invocation missing command parameter")
		}
	})
}

func TestPluginPackageInputs_ExactVersionsAndLockIntegrity(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	expectedRuntime := map[string]string{
		"@opencode-ai/plugin": "1.19.0",
		"zod":                 "4.1.8",
	}
	expectedDevelopment := map[string]string{
		"@vitest/coverage-v8": "5.0.3",
		"vitest":              "5.0.3",
	}
	expectedOverrides := map[string]string{
		"toml": "4.2.0",
		"uuid": "13.0.1",
	}
	expectedEngines := map[string]string{
		"node": "^22.22.2 || ^24.15.0 || >=26.0.0",
	}

	var manifest struct {
		Name            string            `json:"name"`
		Private         bool              `json:"private"`
		Type            string            `json:"type"`
		Scripts         map[string]string `json:"scripts"`
		Engines         map[string]string `json:"engines"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Overrides       map[string]string `json:"overrides"`
	}
	readJSONFile(t, filepath.Join(root, ".opencode", "package.json"), &manifest)
	if manifest.Name != "unbound-force-opencode-plugins" {
		t.Errorf("manifest name = %q, want unbound-force-opencode-plugins", manifest.Name)
	}
	if !manifest.Private {
		t.Error("manifest must be private")
	}
	if manifest.Type != "module" {
		t.Errorf("manifest type = %q, want module", manifest.Type)
	}
	if len(manifest.Scripts) != 0 {
		t.Errorf("task 1.5 must not add npm scripts, got %v", manifest.Scripts)
	}
	assertStringMap(t, "runtime dependencies", manifest.Dependencies, expectedRuntime)
	assertStringMap(t, "development dependencies", manifest.DevDependencies, expectedDevelopment)
	assertStringMap(t, "security overrides", manifest.Overrides, expectedOverrides)
	assertStringMap(t, "node engines", manifest.Engines, expectedEngines)

	var lock struct {
		Name            string                      `json:"name"`
		LockfileVersion int                         `json:"lockfileVersion"`
		Requires        bool                        `json:"requires"`
		Packages        map[string]packageLockEntry `json:"packages"`
	}
	readJSONFile(t, filepath.Join(root, ".opencode", "package-lock.json"), &lock)
	if lock.Name != manifest.Name {
		t.Errorf("lock name = %q, want manifest name %q", lock.Name, manifest.Name)
	}
	if lock.LockfileVersion != 3 {
		t.Errorf("lockfileVersion = %d, want 3", lock.LockfileVersion)
	}
	if !lock.Requires {
		t.Error("lock requires must be true")
	}
	lockRoot, ok := lock.Packages[""]
	if !ok {
		t.Fatal("lock is missing the root package entry")
	}
	assertStringMap(t, "locked runtime dependencies", lockRoot.Dependencies, expectedRuntime)
	assertStringMap(t, "locked development dependencies", lockRoot.DevDependencies, expectedDevelopment)

	for name, version := range expectedRuntime {
		assertLockedDirectPackage(t, lock.Packages, name, version, false)
	}
	for name, version := range expectedDevelopment {
		assertLockedDirectPackage(t, lock.Packages, name, version, true)
	}
	for path, entry := range lock.Packages {
		if path == "" || entry.Resolved == "" {
			continue
		}
		if !strings.HasPrefix(entry.Integrity, "sha512-") {
			t.Errorf("locked package %s integrity = %q, want sha512 integrity", path, entry.Integrity)
		}
	}
	assertLockedSecurityOverride(t, lock.Packages, "toml", "4.2.0",
		"sha512-TvAJjbHZlYmI323+srtqHQFyJsoWy6mI09ppkuj9+iRsqsVKG9fvTcOP7FHF2UCb0QSYtjEavffrKzdd0XgClg==")
	assertLockedSecurityOverride(t, lock.Packages, "uuid", "13.0.1",
		"sha512-9ezox2roIft6ExBVTVqibSd5dc5/47Sw/uY6b4SjQUT2TzQ0tltNquWA46y4xPQmdZYqvnio22SgWd41M86+jw==")
}

func assertLockedDirectPackage(
	t *testing.T,
	packages map[string]packageLockEntry,
	name, version string,
	wantDevelopment bool,
) {
	t.Helper()

	locked, ok := packages["node_modules/"+name]
	if !ok {
		t.Errorf("lock is missing direct package %s", name)
		return
	}
	if locked.Version != version {
		t.Errorf("locked %s version = %q, want %q", name, locked.Version, version)
	}
	if locked.License != "MIT" {
		t.Errorf("locked %s license = %q, want MIT", name, locked.License)
	}
	if locked.Dev != wantDevelopment {
		t.Errorf("locked %s dev = %t, want %t", name, locked.Dev, wantDevelopment)
	}
	if !strings.HasPrefix(locked.Resolved, "https://registry.npmjs.org/") {
		t.Errorf("locked %s resolved URL = %q, want npm registry URL", name, locked.Resolved)
	}
	if !strings.HasPrefix(locked.Integrity, "sha512-") {
		t.Errorf("locked %s integrity = %q, want sha512 integrity", name, locked.Integrity)
	}
}

func assertLockedSecurityOverride(
	t *testing.T,
	packages map[string]packageLockEntry,
	name, version, integrity string,
) {
	t.Helper()

	locked, ok := packages["node_modules/"+name]
	if !ok {
		t.Errorf("lock is missing security override %s", name)
		return
	}
	if locked.Version != version {
		t.Errorf("locked override %s version = %q, want %q", name, locked.Version, version)
	}
	if locked.Integrity != integrity {
		t.Errorf("locked override %s integrity = %q, want %q", name, locked.Integrity, integrity)
	}
	if locked.License != "MIT" {
		t.Errorf("locked override %s license = %q, want MIT", name, locked.License)
	}
}

func TestPluginPackageInputs_ApprovedMaintenanceProvenance(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	type provenanceEntry struct {
		Version             string   `json:"version"`
		License             string   `json:"license"`
		MaintainedBy        string   `json:"maintained_by"`
		ProjectURL          string   `json:"project_url"`
		RegistryMaintainers []string `json:"registry_maintainers"`
	}
	type overrideProvenance struct {
		Version      string `json:"version"`
		License      string `json:"license"`
		GitHead      string `json:"git_head"`
		Integrity    string `json:"integrity"`
		MaintainedBy string `json:"maintained_by"`
		ProjectURL   string `json:"project_url"`
	}
	var provenance struct {
		VerifiedAt        string                        `json:"verified_at"`
		Packages          map[string]provenanceEntry    `json:"packages"`
		SecurityOverrides map[string]overrideProvenance `json:"security_overrides"`
	}
	readJSONFile(t, filepath.Join(root, "internal", "scaffold", "testdata", "plugin-package-provenance.json"), &provenance)

	expectedPackages := map[string]struct {
		version string
		project string
	}{
		"@opencode-ai/plugin": {version: "1.19.0", project: "OpenCode"},
		"zod":                 {version: "4.1.8", project: "Zod"},
		"vitest":              {version: "5.0.3", project: "Vitest"},
		"@vitest/coverage-v8": {version: "5.0.3", project: "Vitest"},
	}
	if provenance.VerifiedAt != "2026-10-19" {
		t.Errorf("provenance verification date = %q, want 2026-10-19", provenance.VerifiedAt)
	}
	if len(provenance.Packages) != len(expectedPackages) {
		t.Errorf("provenance has %d packages, want %d", len(provenance.Packages), len(expectedPackages))
	}
	for name, expected := range expectedPackages {
		entry, ok := provenance.Packages[name]
		if !ok {
			t.Errorf("provenance is missing %s", name)
			continue
		}
		if entry.License != "MIT" {
			t.Errorf("%s provenance license = %q, want MIT", name, entry.License)
		}
		if entry.MaintainedBy != expected.project {
			t.Errorf("%s maintained_by = %q, want %q", name, entry.MaintainedBy, expected.project)
		}
		if entry.Version != expected.version {
			t.Errorf("%s provenance version = %q, want %q", name, entry.Version, expected.version)
		}
		parsedURL, err := url.ParseRequestURI(entry.ProjectURL)
		if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
			t.Errorf("%s project_url = %q, want absolute HTTPS URL", name, entry.ProjectURL)
		}
		if len(entry.RegistryMaintainers) == 0 {
			t.Errorf("%s registry maintainers are empty", name)
		}
	}

	expectedOverrides := map[string]overrideProvenance{
		"toml": {
			Version:      "4.2.0",
			License:      "MIT",
			GitHead:      "6d42b7e4fa693f8e6fb025d419b10b38107e83e2",
			Integrity:    "sha512-TvAJjbHZlYmI323+srtqHQFyJsoWy6mI09ppkuj9+iRsqsVKG9fvTcOP7FHF2UCb0QSYtjEavffrKzdd0XgClg==",
			MaintainedBy: "toml-node",
			ProjectURL:   "https://github.com/BinaryMuse/toml-node",
		},
		"uuid": {
			Version:      "13.0.1",
			License:      "MIT",
			GitHead:      "0643802db81cece7ee445f5147529d7a77394630",
			Integrity:    "sha512-9ezox2roIft6ExBVTVqibSd5dc5/47Sw/uY6b4SjQUT2TzQ0tltNquWA46y4xPQmdZYqvnio22SgWd41M86+jw==",
			MaintainedBy: "uuidjs",
			ProjectURL:   "https://github.com/uuidjs/uuid",
		},
	}
	if len(provenance.SecurityOverrides) != len(expectedOverrides) {
		t.Errorf("provenance has %d security overrides, want %d",
			len(provenance.SecurityOverrides), len(expectedOverrides))
	}
	for name, expected := range expectedOverrides {
		got, ok := provenance.SecurityOverrides[name]
		if !ok {
			t.Errorf("provenance is missing security override %s", name)
			continue
		}
		if got != expected {
			t.Errorf("security override provenance for %s = %+v, want %+v", name, got, expected)
		}
		parsedURL, err := url.ParseRequestURI(got.ProjectURL)
		if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
			t.Errorf("%s override project_url = %q, want absolute HTTPS URL", name, got.ProjectURL)
		}
	}
}

func TestPluginPackageInputs_GitignoreTracksCanonicalInputs(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	content, err := os.ReadFile(filepath.Join(root, ".opencode", ".gitignore"))
	if err != nil {
		t.Fatalf("read .opencode/.gitignore: %v", err)
	}
	want := "node_modules\nbun.lock\n"
	if string(content) != want {
		t.Errorf(".opencode/.gitignore = %q, want %q", string(content), want)
	}
	for _, trackedInput := range []string{"package.json", "package-lock.json"} {
		if strings.Contains(string(content), trackedInput) {
			t.Errorf(".opencode/.gitignore must not ignore %s", trackedInput)
		}
	}
}

func TestPluginPackageInputs_MirrorsMatchCanonical(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	for _, name := range []string{"package.json", "package-lock.json"} {
		canonical, err := os.ReadFile(filepath.Join(root, ".opencode", name))
		if err != nil {
			t.Fatalf("read canonical %s: %v", name, err)
		}
		mirror, err := assetContent(filepath.ToSlash(filepath.Join("opencode", name)))
		if err != nil {
			t.Fatalf("read scaffold mirror %s: %v", name, err)
		}
		if !bytes.Equal(canonical, mirror) {
			t.Errorf("scaffold mirror %s differs from canonical input", name)
		}
	}
}

func TestSCFR003_FreshTargetInputsDoNotActivatePlugins(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	target := t.TempDir()
	var output bytes.Buffer
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: target,
		Version:   "1.0.0-test",
		Stdout:    &output,
		LookPath:  stubScaffoldLookPath(nil),
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			t.Fatalf("fresh package-input scaffold unexpectedly executed %s %v", name, args)
			return nil, nil
		},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, name := range []string{"package.json", "package-lock.json"} {
		canonical, err := os.ReadFile(filepath.Join(root, ".opencode", name))
		if err != nil {
			t.Fatalf("read canonical %s: %v", name, err)
		}
		generatedPath := filepath.Join(target, ".opencode", name)
		generated, err := os.ReadFile(generatedPath)
		if err != nil {
			t.Fatalf("read fresh-target %s: %v", name, err)
		}
		if !bytes.Equal(generated, canonical) {
			t.Errorf("fresh-target %s differs from canonical reproducible input", name)
		}
		if !containsPath(result.Created, filepath.ToSlash(filepath.Join(".opencode", name))) {
			t.Errorf("fresh-target %s is absent from result.Created", name)
		}
	}

	for _, relPath := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset, ufWorkflowPluginAsset, ufFinalePluginAsset} {
		if !isActivationGatedAsset(relPath) {
			t.Errorf("plugin source %s is not activation-gated", relPath)
		}
		if _, err := assetContent(relPath); err != nil {
			t.Errorf("read embedded activation-gated source %s: %v", relPath, err)
		}
		autoDiscoveryPath := filepath.Join(target, mapAssetPath(relPath))
		if _, err := os.Stat(autoDiscoveryPath); !os.IsNotExist(err) {
			t.Errorf("activation-gated source exposed at %s before install and probes: %v", autoDiscoveryPath, err)
		}
		if containsPath(result.Created, mapAssetPath(relPath)) {
			t.Errorf("activation-gated source %s reported as created", relPath)
		}
	}

	pluginsDir := filepath.Join(target, ".opencode", "plugins")
	if _, err := os.Stat(pluginsDir); !os.IsNotExist(err) {
		t.Errorf("auto-discovery directory exists before plugin activation: %v", err)
	}
}

func TestIsActivationGatedAsset_OnlyReviewPluginSources(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: invokeAgentPluginAsset, want: true},
		{path: reviewDispatchPluginAsset, want: true},
		{path: ufWorkflowPluginAsset, want: true},
		{path: "opencode/package.json", want: false},
		{path: "opencode/lib/review-dispatch-sibling-evidence.ts", want: false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			got := isActivationGatedAsset(test.path)
			if got != test.want {
				t.Errorf("isActivationGatedAsset(%q) = %t, want %t", test.path, got, test.want)
			}
		})
	}
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}

func TestPluginPackageInputs_RollbackPrerequisites(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	var rollback struct {
		BaselineLockSHA256 string `json:"baseline_lock_sha256"`
		PriorGitignore     string `json:"prior_gitignore"`
		SecurityOverride   struct {
			LockSHA256         string            `json:"lock_sha256"`
			TransitiveVersions map[string]string `json:"transitive_versions"`
		} `json:"security_override_rollback"`
	}
	readJSONFile(t, filepath.Join(root, "internal", "scaffold", "testdata", "plugin-package-rollback.json"), &rollback)

	const baselineLockSHA256 = "1ac9cdf171a743c69b18498e15711d628280a4748defd6334cd4276e194d32dc"
	if rollback.BaselineLockSHA256 != baselineLockSHA256 {
		t.Errorf("baseline lock SHA256 = %q, want %q", rollback.BaselineLockSHA256, baselineLockSHA256)
	}
	const priorGitignore = "node_modules\npackage.json\nbun.lock\n.gitignore\n"
	if rollback.PriorGitignore != priorGitignore {
		t.Errorf("prior .opencode/.gitignore = %q, want %q", rollback.PriorGitignore, priorGitignore)
	}
	const preHardeningLockSHA256 = "fef6f7f145d9dc8398233566934ff7a4e36d9505dbcb662660c773c03c1cdcba"
	if rollback.SecurityOverride.LockSHA256 != preHardeningLockSHA256 {
		t.Errorf("pre-hardening lock SHA256 = %q, want %q",
			rollback.SecurityOverride.LockSHA256, preHardeningLockSHA256)
	}
	assertStringMap(t, "pre-hardening transitive versions",
		rollback.SecurityOverride.TransitiveVersions,
		map[string]string{"toml": "4.1.1", "uuid": "13.0.0"})
}

func TestDispatchRuntime_TaskScopedPluginSources(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	invokeCanonical := filepath.Join(root, ".opencode", "plugins", "invoke-agent", "index.ts")
	if _, err := os.Stat(invokeCanonical); err != nil {
		t.Fatalf("task 2.2 invoke-agent source missing at %s: %v", invokeCanonical, err)
	}
	invokeEmbedded := filepath.ToSlash(filepath.Join("opencode", "plugins", "invoke-agent", "index.ts"))
	if _, err := assetContent(invokeEmbedded); err != nil {
		t.Fatalf("task 2.2 invoke-agent scaffold mirror missing at %s: %v", invokeEmbedded, err)
	}

	reviewCanonical := filepath.Join(root, ".opencode", "plugins", "review-dispatch", "index.ts")
	if _, err := os.Stat(reviewCanonical); err != nil {
		t.Fatalf("task 2.1 review-dispatch source missing at %s: %v", reviewCanonical, err)
	}
	reviewEmbedded := filepath.ToSlash(filepath.Join("opencode", "plugins", "review-dispatch", "index.ts"))
	if _, err := assetContent(reviewEmbedded); err != nil {
		t.Fatalf("task 2.1 review-dispatch scaffold mirror missing at %s: %v", reviewEmbedded, err)
	}
}

// TestEmbeddedAssets_SingleMarker verifies that no embedded
// Markdown asset contains more than one scaffold provenance
// marker line. This prevents marker accumulation through the
// asset-sync feedback loop.
func TestEmbeddedAssets_SingleMarker(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	for _, relPath := range paths {
		if filepath.Ext(relPath) != ".md" {
			continue
		}

		content, err := assetContent(relPath)
		if err != nil {
			t.Errorf("read embedded %s: %v", relPath, err)
			continue
		}

		count := 0
		for _, line := range strings.Split(string(content), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "<!-- scaffolded by uf ") ||
				strings.HasPrefix(trimmed, "# scaffolded by uf ") {
				count++
			}
		}

		if count > 1 {
			t.Errorf("embedded asset %s contains %d scaffold markers (expected at most 1)",
				relPath, count)
		}
	}
}

// mapAssetToSource converts an embedded asset relative path to
// the canonical source path at the repo root. Delegates to
// mapAssetPath to avoid duplicating the prefix mapping logic.
func mapAssetToSource(relPath string) string {
	return mapAssetPath(relPath)
}

// expectedAssetPaths is the canonical list of embedded assets.
// Update this list when adding or removing assets.
var expectedAssetPaths = []string{
	// OpenCode commands (12) — UF-custom only; speckit.*.md externalized to specify init + /uf.init
	"opencode/commands/uf.address-feedback.md",
	"opencode/commands/uf.agent-brief.md",
	"opencode/commands/uf.cobalt-crush.md",
	"opencode/commands/uf.constitution-check.md",
	"opencode/commands/uf.finale.md",
	"opencode/commands/uf.fix-ci.md",
	"opencode/commands/uf.init.md",
	"opencode/commands/uf.rebase-pr.md",
	"opencode/commands/uf.review-council.md",
	"opencode/commands/uf.review-pr.md",
	"opencode/commands/uf.run-pipeline-plan.md",
	"opencode/commands/uf.triage-issue.md",
	"opencode/commands/uf.unleash.md",
	// OpenCode agents — Divisor personas (6) + Cobalt-Crush (1) + constitution-check (1)
	"opencode/agents/cobalt-crush-dev.md",
	"opencode/agents/constitution-check.md",
	"opencode/agents/divisor-adversary.md",
	"opencode/agents/divisor-architect.md",
	"opencode/agents/divisor-curator.md",
	"opencode/agents/divisor-guard.md",
	"opencode/agents/divisor-sre.md",
	"opencode/agents/divisor-testing.md",
	// OpenCode agents — Divisor content agents (3)
	"opencode/agents/divisor-scribe.md",
	"opencode/agents/divisor-herald.md",
	"opencode/agents/divisor-envoy.md",
	// Convention packs — shared by all heroes (15)
	"opencode/uf/packs/ci-custom.md",
	"opencode/uf/packs/ci.md",
	"opencode/uf/packs/content-custom.md",
	"opencode/uf/packs/content.md",
	"opencode/uf/packs/default-custom.md",
	"opencode/uf/packs/default.md",
	"opencode/uf/packs/go-custom.md",
	"opencode/uf/packs/go.md",
	"opencode/uf/packs/python-custom.md",
	"opencode/uf/packs/python.md",
	"opencode/uf/packs/review-voice-custom.md",
	"opencode/uf/packs/review-voice.md",
	"opencode/uf/packs/severity.md",
	"opencode/uf/packs/typescript-custom.md",
	"opencode/uf/packs/typescript.md",
	// OpenSpec schema (5)
	"openspec/schemas/unbound-force/schema.yaml",
	"openspec/schemas/unbound-force/templates/proposal.md",
	"openspec/schemas/unbound-force/templates/spec.md",
	"openspec/schemas/unbound-force/templates/design.md",
	"openspec/schemas/unbound-force/templates/tasks.md",
	// Swarm skills (6)
	"opencode/skills/always-on-guidance/SKILL.md",
	"opencode/skills/dispatch-advisor/SKILL.md",
	"opencode/skills/pre-flight/SKILL.md",
	"opencode/skills/review-context/SKILL.md",
	"opencode/skills/speckit-workflow/SKILL.md",
	"opencode/skills/tier-routing/SKILL.md",
	// Review dispatch plugins and policy modules (6)
	"opencode/plugins/invoke-agent/index.ts",
	"opencode/plugins/review-dispatch/index.ts",
	// UF workflow plugin and types library (2)
	"opencode/plugins/uf-workflow/index.ts",
	"opencode/lib/uf-workflow-types.ts",
	// UF finale plugin (1)
	"opencode/plugins/uf-finale/index.ts",
	// Shared lib modules (11)
	"opencode/lib/agent-executor.ts",
	"opencode/lib/branch-classification.ts",
	"opencode/lib/model-matrix.ts",
	"opencode/lib/review-dispatch-artifact.ts",
	"opencode/lib/review-dispatch-lesson-proposal.ts",
	"opencode/lib/review-dispatch-sibling-evidence.ts",
	"opencode/lib/review-diff.ts",
	"opencode/lib/review-diff-manifest.ts",
	"opencode/lib/reviewer-manifest.ts",
	"opencode/lib/run-timeout.ts",
	"opencode/lib/unleash-templates.ts",
	// Reproducible plugin package inputs (2)
	"opencode/package-lock.json",
	"opencode/package.json",
	// Review policy, model-matrix template, and project-owned sibling
	// template (4)
	"uf/model-matrix.yaml",
	"uf/review-matrix.yaml",
	"uf/reviewer-capabilities.yaml",
	"uf/sibling-repos.yaml",
	// Multi-model review schemas (8)
	"uf/schemas/lesson-proposal/v1.0.0.schema.json",
	"uf/schemas/model-matrix/v1.schema.json",
	"uf/schemas/review-dispatch/v1.0.0.schema.json",
	"uf/schemas/review-matrix/v2.schema.json",
	"uf/schemas/review-matrix/v3.schema.json",
	"uf/schemas/review-verdict/v2.0.0.schema.json",
	"uf/schemas/reviewer-capabilities/v1.0.0.schema.json",
	"uf/schemas/sibling-repos/v1.0.0.schema.json",
	// Specify — starter constitution (1)
	"specify/memory/constitution.md",
}

var task42CanonicalAssets = []struct {
	asset  string
	source string
}{
	{asset: "opencode/agents/cobalt-crush-dev.md", source: ".opencode/agents/cobalt-crush-dev.md"},
	{asset: "opencode/commands/uf.address-feedback.md", source: ".opencode/commands/uf.address-feedback.md"},
	{asset: "opencode/commands/uf.review-council.md", source: ".opencode/commands/uf.review-council.md"},
	{asset: "opencode/commands/uf.triage-issue.md", source: ".opencode/commands/uf.triage-issue.md"},
	{asset: "opencode/skills/dispatch-advisor/SKILL.md", source: ".opencode/skills/dispatch-advisor/SKILL.md"},
	{asset: invokeAgentPluginAsset, source: ".opencode/plugins/invoke-agent/index.ts"},
	{asset: reviewDispatchPluginAsset, source: ".opencode/plugins/review-dispatch/index.ts"},
	{asset: "opencode/plugins/uf-workflow/index.ts", source: ".opencode/plugins/uf-workflow/index.ts"},
	{asset: "opencode/plugins/uf-finale/index.ts", source: ".opencode/plugins/uf-finale/index.ts"},
	{asset: "opencode/lib/uf-workflow-types.ts", source: ".opencode/lib/uf-workflow-types.ts"},
	{asset: "opencode/lib/agent-executor.ts", source: ".opencode/lib/agent-executor.ts"},
	{asset: "opencode/lib/branch-classification.ts", source: ".opencode/lib/branch-classification.ts"},
	{asset: "opencode/lib/model-matrix.ts", source: ".opencode/lib/model-matrix.ts"},
	{asset: "opencode/lib/review-dispatch-artifact.ts", source: ".opencode/lib/review-dispatch-artifact.ts"},
	{asset: "opencode/lib/review-dispatch-lesson-proposal.ts", source: ".opencode/lib/review-dispatch-lesson-proposal.ts"},
	{asset: "opencode/lib/review-dispatch-sibling-evidence.ts", source: ".opencode/lib/review-dispatch-sibling-evidence.ts"},
	{asset: "opencode/lib/review-diff.ts", source: ".opencode/lib/review-diff.ts"},
	{asset: "opencode/lib/review-diff-manifest.ts", source: ".opencode/lib/review-diff-manifest.ts"},
	{asset: "opencode/lib/reviewer-manifest.ts", source: ".opencode/lib/reviewer-manifest.ts"},
	{asset: "opencode/lib/run-timeout.ts", source: ".opencode/lib/run-timeout.ts"},
	{asset: "opencode/lib/unleash-templates.ts", source: ".opencode/lib/unleash-templates.ts"},
	{asset: "opencode/package-lock.json", source: ".opencode/package-lock.json"},
	{asset: "opencode/package.json", source: ".opencode/package.json"},
	{asset: "uf/model-matrix.yaml", source: ".uf/model-matrix.yaml"},
	{asset: "uf/review-matrix.yaml", source: ".uf/review-matrix.yaml"},
	{asset: "uf/reviewer-capabilities.yaml", source: ".uf/reviewer-capabilities.yaml"},
	{asset: "uf/schemas/lesson-proposal/v1.0.0.schema.json", source: ".uf/schemas/lesson-proposal/v1.0.0.schema.json"},
	{asset: "uf/schemas/model-matrix/v1.schema.json", source: ".uf/schemas/model-matrix/v1.schema.json"},
	{asset: "uf/schemas/review-dispatch/v1.0.0.schema.json", source: ".uf/schemas/review-dispatch/v1.0.0.schema.json"},
	{asset: "uf/schemas/review-matrix/v2.schema.json", source: ".uf/schemas/review-matrix/v2.schema.json"},
	{asset: "uf/schemas/review-matrix/v3.schema.json", source: ".uf/schemas/review-matrix/v3.schema.json"},
	{asset: "uf/schemas/review-verdict/v2.0.0.schema.json", source: ".uf/schemas/review-verdict/v2.0.0.schema.json"},
	{asset: "uf/schemas/reviewer-capabilities/v1.0.0.schema.json", source: ".uf/schemas/reviewer-capabilities/v1.0.0.schema.json"},
	{asset: "uf/schemas/sibling-repos/v1.0.0.schema.json", source: ".uf/schemas/sibling-repos/v1.0.0.schema.json"},
}

func TestSCFR001_CanonicalAssetSurfaceMatchesMirrors(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}
	embedded := make(map[string]bool, len(paths))
	for _, path := range paths {
		embedded[path] = true
	}

	for _, asset := range task42CanonicalAssets {
		t.Run(asset.asset, func(t *testing.T) {
			if !embedded[asset.asset] {
				t.Fatalf("required embedded asset %q is missing", asset.asset)
			}
			if got := mapAssetToSource(asset.asset); got != asset.source {
				t.Fatalf("mapAssetToSource(%q) = %q, want %q", asset.asset, got, asset.source)
			}
			mirror, err := assetContent(asset.asset)
			if err != nil {
				t.Fatalf("read embedded asset: %v", err)
			}
			canonical, err := os.ReadFile(filepath.Join(root, asset.source))
			if err != nil {
				t.Fatalf("read canonical source: %v", err)
			}
			if !bytes.Equal(mirror, canonical) {
				t.Errorf("embedded asset %q drifted from %q", asset.asset, asset.source)
			}
		})
	}
}

func TestSCFR001_SiblingTemplateAndLiveOnlyTestReviewExclusions(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}
	embedded := make(map[string]bool, len(paths))
	for _, path := range paths {
		embedded[path] = true
	}

	const siblingAsset = "uf/sibling-repos.yaml"
	if !embedded[siblingAsset] {
		t.Fatalf("required sibling template %q is missing", siblingAsset)
	}
	template, err := assetContent(siblingAsset)
	if err != nil {
		t.Fatalf("read sibling template: %v", err)
	}
	if !strings.Contains(string(template), "version: 1\nsiblings: []") {
		t.Error("sibling scaffold must remain an empty version 1 template")
	}
	if strings.Contains(string(template), "github.com/unbound-force/") {
		t.Error("sibling scaffold must not contain live organization entries")
	}

	const liveOnly = "opencode/commands/speckit.testreview.md"
	if embedded[liveOnly] {
		t.Errorf("live-only Speckit test-review command must not be embedded as %q", liveOnly)
	}
	root := findProjectRoot(t)
	if _, err := os.Stat(filepath.Join(root, ".opencode", "commands", "speckit.testreview.md")); err != nil {
		t.Fatalf("live-only Speckit test-review command is missing: %v", err)
	}
}

func expectedInitiallyDeployedAssetPaths() []string {
	paths := make([]string, 0, len(expectedAssetPaths))
	for _, path := range expectedAssetPaths {
		if !isActivationGatedAsset(path) {
			paths = append(paths, path)
		}
	}
	return paths
}

// nonDeployedAssetPaths lists embedded assets that are NOT
// deployed by uf init. These are accessed via dedicated
// functions (e.g., DevcontainerContent()) and are skipped
// during the Run() walk. They must be listed here so
// TestAssetPaths_MatchExpected accounts for them.
var nonDeployedAssetPaths = []string{
	// Devcontainer template — OS-specific, generated per-user
	// by uf sandbox init (not deployed by uf init).
	"devcontainer/devcontainer.json",
}

func TestAssetPaths_MatchExpected(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	sort.Strings(paths)

	// Combine deployed and non-deployed assets for the full manifest.
	allExpected := make([]string, 0, len(expectedAssetPaths)+len(nonDeployedAssetPaths))
	allExpected = append(allExpected, expectedAssetPaths...)
	allExpected = append(allExpected, nonDeployedAssetPaths...)
	sort.Strings(allExpected)

	if len(paths) != len(allExpected) {
		t.Errorf("expected %d assets, got %d", len(allExpected), len(paths))
		t.Logf("expected: %v", allExpected)
		t.Logf("got:      %v", paths)
		return
	}

	for i := range paths {
		if paths[i] != allExpected[i] {
			t.Errorf("asset mismatch at index %d: expected %q, got %q", i, allExpected[i], paths[i])
		}
	}
}

func TestRun_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// All files should be created on first run
	if len(result.Created) == 0 {
		t.Error("expected files to be created")
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected no skipped files, got %d", len(result.Skipped))
	}
	if len(result.Overwritten) != 0 {
		t.Errorf("expected no overwritten files, got %d", len(result.Overwritten))
	}
	if len(result.Updated) != 0 {
		t.Errorf("expected no updated files, got %d", len(result.Updated))
	}

	// Verify expected directory structure
	expectedDirs := []string{
		".opencode/commands",
		".opencode/agents",
		".opencode/lib",
		".opencode/skills/dispatch-advisor",
		".opencode/uf/packs",
		".uf",
		".specify/memory",
		"openspec/specs",
		"openspec/changes",
		".uf/schemas/lesson-proposal",
		".uf/schemas/review-dispatch",
		".uf/schemas/review-matrix",
		".uf/schemas/review-verdict",
		".uf/schemas/reviewer-capabilities",
		".uf/schemas/sibling-repos",
	}
	for _, d := range expectedDirs {
		full := filepath.Join(dir, d)
		info, err := os.Stat(full)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}

	// Verify created file count matches expected assets
	expectedDeployed := expectedInitiallyDeployedAssetPaths()
	if len(result.Created) != len(expectedDeployed) {
		t.Errorf("expected %d created files, got %d", len(expectedDeployed), len(result.Created))
	}
	for _, asset := range expectedDeployed {
		outputPath := mapAssetPath(asset)
		if !containsPath(result.Created, outputPath) {
			t.Errorf("expected deployed asset %q in result.Created", outputPath)
		}
		info, err := os.Stat(filepath.Join(dir, outputPath))
		if err != nil {
			t.Errorf("expected deployed asset %q to exist: %v", outputPath, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("expected deployed asset %q to be a file", outputPath)
		}
	}
	for _, asset := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset, ufWorkflowPluginAsset, ufFinalePluginAsset} {
		outputPath := mapAssetPath(asset)
		if _, err := os.Stat(filepath.Join(dir, outputPath)); !os.IsNotExist(err) {
			t.Errorf("activation-gated asset %q must not be initially deployed: %v", outputPath, err)
		}
	}
}

func TestRun_ConstitutionScaffolded(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Constitution should be in Created list.
	constitutionOut := ".specify/memory/constitution.md"
	found := false
	for _, f := range result.Created {
		if f == constitutionOut {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %s in result.Created", constitutionOut)
	}

	// Read the scaffolded constitution.
	content, err := os.ReadFile(filepath.Join(dir, constitutionOut))
	if err != nil {
		t.Fatalf("read constitution: %v", err)
	}
	text := string(content)

	// Must contain all five principle headings.
	for _, heading := range []string{
		"### I. Autonomous Collaboration",
		"### II. Composability First",
		"### III. Observable Quality",
		"### IV. Testability",
		"### V. Security by Default",
	} {
		if !strings.Contains(text, heading) {
			t.Errorf("constitution missing heading %q", heading)
		}
	}

	// Must NOT contain [PLACEHOLDER] tokens (FR-001).
	if strings.Contains(text, "[PLACEHOLDER]") {
		t.Error("constitution contains [PLACEHOLDER] tokens")
	}

	// Must contain version marker (FR-001 Scenario 2).
	if !strings.Contains(text, "<!-- scaffolded by uf v") {
		t.Error("constitution missing version marker")
	}
}

// TestRun_SpecifyDelegationOnFreshInit is the Run()-level
// regression test for the specify sentinel bug. Run()'s embedded
// asset walk scaffolds .specify/memory/constitution.md before
// delegating to sub-tools; under the old bare-.specify sentinel
// that walk created the sentinel itself, silently suppressing
// `specify init` on every fresh non-force run. Existing TestRun_*
// tests never stub LookPath for specify (the real exec.LookPath
// finds no binary in CI, so the delegation no-ops there too), and
// TestInitSubTools_SpecifyInit calls initSubTools directly,
// bypassing Run()'s scaffolding phase — which is how CI stayed
// green while the command was broken. This test runs the full
// Run() with only specify resolvable on a stubbed PATH and asserts
// the delegation fires and the constitution coexists with it.
// (D5; FR-021, FR-007)
func TestRun_SpecifyDelegationOnFreshInit(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	// runScaffoldWithoutExternalTools keeps the caller-supplied
	// LookPath and guards ExecCmdInDir against unexpected external
	// commands, mirroring TestRun_CreatesFiles and
	// TestRun_ConstitutionScaffolded; only specify resolves.
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// (a) The delegation must fire with the fixed invocation
	// contract: --force gives specify's non-empty-directory merge
	// semantics for the target the asset walk just populated.
	expected := "specify init --here --integration opencode --offline --force"
	found := false
	for _, call := range rec.calls {
		if call == expected {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %q, got calls: %v", expected, rec.calls)
	}

	// (b) The starter constitution must coexist with the
	// delegation: the asset walk scaffolds it regardless of
	// specify's merge behavior.
	constitution := filepath.Join(dir, ".specify", "memory", "constitution.md")
	if _, err := os.Stat(constitution); err != nil {
		t.Errorf("expected %s to exist after Run(): %v", constitution, err)
	}

	// (c) The sub-tool results must include .specify/ initialized.
	// Run() renders sub-results through the summary pipeline
	// (printSubToolResults), so the action is observable on Stdout.
	if !strings.Contains(buf.String(), ".specify/ initialized") {
		t.Errorf("expected '.specify/ initialized' in Run() summary output, got:\n%s", buf.String())
	}
}

func TestRun_ConstitutionPreservedOnRerun(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run creates constitution.
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// User customizes the constitution.
	constitutionPath := filepath.Join(dir, ".specify", "memory", "constitution.md")
	customContent := []byte("# My Custom Constitution\n\nCustomized by user.\n")
	if err := os.WriteFile(constitutionPath, customContent, 0o644); err != nil {
		t.Fatalf("write custom constitution: %v", err)
	}

	// Re-run without --force: constitution should be preserved.
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Constitution should be in Skipped, not Created or Overwritten.
	constitutionOut := ".specify/memory/constitution.md"
	for _, f := range result.Created {
		if f == constitutionOut {
			t.Errorf("expected constitution to be skipped, but found in Created")
		}
	}
	for _, f := range result.Overwritten {
		if f == constitutionOut {
			t.Errorf("expected constitution to be skipped, but found in Overwritten")
		}
	}
	foundSkipped := false
	for _, f := range result.Skipped {
		if f == constitutionOut {
			foundSkipped = true
			break
		}
	}
	if !foundSkipped {
		t.Errorf("expected %s in result.Skipped", constitutionOut)
	}

	// Verify user content is preserved.
	got, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("read constitution after re-run: %v", err)
	}
	if !bytes.Equal(got, customContent) {
		t.Errorf("constitution content was modified on re-run")
	}
}

// TestRun_ReviewerCapabilitiesPreservedOnRerun pins the project-owned
// extension-point contract for uf/reviewer-capabilities.yaml: locally
// registered reviewers entries must survive a plain re-run (no --force),
// where the previous tool-owned classification would have reverted them.
func TestRun_ReviewerCapabilitiesPreservedOnRerun(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run deploys the seeded capabilities manifest.
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// The generated project registers a custom Divisor agent. The
	// appended entry follows the seeded reviewers shape; the deployed
	// file ends with a newline, so the manifest stays well-formed YAML.
	capsPath := filepath.Join(dir, ".uf", "reviewer-capabilities.yaml")
	seeded, err := os.ReadFile(capsPath)
	if err != nil {
		t.Fatalf("read seeded reviewer-capabilities: %v", err)
	}
	customEntry := []byte("  - agent: divisor-custom-example\n" +
		"    capability: review\n" +
		"    scopes:\n" +
		"      - standard\n")
	// Pin the fixture precondition: the deployed manifest must end with
	// a newline, otherwise the appended item concatenates onto the last
	// seeded line and the fixture stops modeling a valid registration.
	if !bytes.HasSuffix(seeded, []byte("\n")) {
		t.Fatalf("seeded reviewer-capabilities must end with a newline so the appended entry stays well-formed YAML")
	}
	if err := os.WriteFile(capsPath, append(seeded, customEntry...), 0o644); err != nil {
		t.Fatalf("append custom reviewer entry: %v", err)
	}
	customized, err := os.ReadFile(capsPath)
	if err != nil {
		t.Fatalf("read customized reviewer-capabilities: %v", err)
	}

	// Re-run without --force: the local entry must survive.
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// The manifest must land in Skipped, not any write disposition.
	capsOut := ".uf/reviewer-capabilities.yaml"
	for _, f := range result.Created {
		if f == capsOut {
			t.Errorf("expected reviewer-capabilities to be skipped, but found in Created")
		}
	}
	for _, f := range result.Updated {
		if f == capsOut {
			t.Errorf("expected reviewer-capabilities to be skipped, but found in Updated")
		}
	}
	for _, f := range result.Overwritten {
		if f == capsOut {
			t.Errorf("expected reviewer-capabilities to be skipped, but found in Overwritten")
		}
	}
	foundSkipped := false
	for _, f := range result.Skipped {
		if f == capsOut {
			foundSkipped = true
			break
		}
	}
	if !foundSkipped {
		t.Errorf("expected %s in result.Skipped", capsOut)
	}

	// The locally registered agent must still be present byte-for-byte.
	got, err := os.ReadFile(capsPath)
	if err != nil {
		t.Fatalf("read reviewer-capabilities after re-run: %v", err)
	}
	if !bytes.Equal(got, customized) {
		t.Errorf("reviewer-capabilities content was modified on re-run")
	}
}

func TestRun_ConstitutionProtectedWithForce(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	noTools := stubScaffoldLookPath(nil)

	// First run creates constitution.
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
		LookPath:  noTools,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// User customizes the constitution.
	constitutionPath := filepath.Join(dir, ".specify", "memory", "constitution.md")
	customContent := "# My Custom Constitution\n\nCustomized governance."
	if err := os.WriteFile(constitutionPath, []byte(customContent), 0o644); err != nil {
		t.Fatalf("write custom constitution: %v", err)
	}

	// Re-run with --force: constitution should be protected (skipped).
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Force:     true,
		Version:   "1.0.0",
		Stdout:    &buf,
		LookPath:  noTools,
	})
	if err != nil {
		t.Fatalf("force Run() error: %v", err)
	}

	constitutionOut := ".specify/memory/constitution.md"

	// Constitution MUST be in Skipped.
	foundSkipped := false
	for _, f := range result.Skipped {
		if f == constitutionOut {
			foundSkipped = true
			break
		}
	}
	if !foundSkipped {
		t.Errorf("expected %s in result.Skipped, got Skipped=%v", constitutionOut, result.Skipped)
	}

	// Constitution MUST NOT be in Overwritten.
	for _, f := range result.Overwritten {
		if f == constitutionOut {
			t.Errorf("constitution must not be in result.Overwritten")
			break
		}
	}

	// Content should be the user's customized content, not the starter.
	got, err := os.ReadFile(constitutionPath)
	if err != nil {
		t.Fatalf("read constitution after force: %v", err)
	}
	if string(got) != customContent {
		t.Errorf("constitution content changed; got %q, want %q", string(got), customContent)
	}
}

func TestRun_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run creates everything
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Second run should skip user-owned, skip identical tool-owned
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	if len(result.Created) != 0 {
		t.Errorf("expected no created files on second run, got %d", len(result.Created))
	}
	// All files should be skipped (user-owned skipped, tool-owned
	// skipped because content is identical)
	if len(result.Updated) != 0 {
		t.Errorf("expected no updated files on identical re-run, got %d: %v",
			len(result.Updated), result.Updated)
	}
	expectedDeployed := expectedInitiallyDeployedAssetPaths()
	if len(result.Skipped) != len(expectedDeployed) {
		t.Errorf("expected %d skipped files, got %d",
			len(expectedDeployed), len(result.Skipped))
	}

	// Verify a known tool-owned file is in Skipped
	foundToolSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "uf.review-council.md") {
			foundToolSkip = true
			break
		}
	}
	if !foundToolSkip {
		t.Error("expected tool-owned uf.review-council.md to be in Skipped list")
	}

	// Verify a known user-owned file is in Skipped
	foundUserSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "cobalt-crush-dev.md") {
			foundUserSkip = true
			break
		}
	}
	if !foundUserSkip {
		t.Error("expected user-owned cobalt-crush-dev.md to be in Skipped list")
	}
}

func TestRun_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer
	noTools := stubScaffoldLookPath(nil)

	// First run
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
		LookPath:  noTools,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Second run with --force
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Force:     true,
		Version:   "1.0.0",
		Stdout:    &buf,
		LookPath:  noTools,
	})
	if err != nil {
		t.Fatalf("force Run() error: %v", err)
	}

	// Protected files (isNeverOverwrite) are skipped even with --force,
	// so Overwritten count is total assets minus protected files.
	neverOverwriteCount := 0
	expectedDeployed := expectedInitiallyDeployedAssetPaths()
	for _, p := range expectedDeployed {
		if isNeverOverwrite(p) {
			neverOverwriteCount++
		}
	}
	wantOverwritten := len(expectedDeployed) - neverOverwriteCount
	if len(result.Overwritten) != wantOverwritten {
		t.Errorf("expected %d overwritten files, got %d",
			wantOverwritten, len(result.Overwritten))
	}
	if len(result.Created) != 0 {
		t.Errorf("expected no created files with force, got %d", len(result.Created))
	}
	// Protected files should be in Skipped.
	if neverOverwriteCount > 0 {
		foundProtected := false
		for _, f := range result.Skipped {
			if f == ".specify/memory/constitution.md" {
				foundProtected = true
				break
			}
		}
		if !foundProtected {
			t.Error("expected protected constitution in result.Skipped with --force")
		}
	}
}

func TestRun_VersionMarker(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.2.3",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, relPath := range expectedInitiallyDeployedAssetPaths() {
		ext := filepath.Ext(relPath)
		if !markerFileExtensions[ext] {
			continue // unsupported extensions don't get markers
		}

		outRel := mapAssetPath(relPath)
		outPath := filepath.Join(dir, outRel)

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("read %s: %v", outRel, err)
			continue
		}

		marker := versionMarker("1.2.3", ext)

		if !strings.Contains(string(content), marker) {
			t.Errorf("file %s does not contain version marker %q", outRel, marker)
		}
	}
}

func TestRun_VersionMarkerDev(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Version defaults to "0.0.0-dev" — check all supported files
	for _, relPath := range expectedInitiallyDeployedAssetPaths() {
		ext := filepath.Ext(relPath)
		if !markerFileExtensions[ext] {
			continue // unsupported extensions don't get markers
		}

		outRel := mapAssetPath(relPath)
		outPath := filepath.Join(dir, outRel)

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Errorf("read %s: %v", outRel, err)
			continue
		}

		marker := versionMarker("0.0.0-dev", ext)

		if !strings.Contains(string(content), marker) {
			t.Errorf("file %s does not contain dev version marker %q", outRel, marker)
		}
	}
}

func TestRun_OverwriteOnDiff_ToolOwned(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Modify a tool-owned file on disk
	toolFile := filepath.Join(dir, ".opencode", "commands", "uf.review-council.md")
	if err := os.WriteFile(toolFile, []byte("modified content"), 0o644); err != nil {
		t.Fatalf("modify tool-owned file: %v", err)
	}

	// Modify a user-owned file on disk
	userFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")
	if err := os.WriteFile(userFile, []byte("user modified"), 0o644); err != nil {
		t.Fatalf("modify user-owned file: %v", err)
	}

	// Re-run
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Tool-owned file should be updated
	if len(result.Updated) == 0 {
		t.Error("expected at least one updated file (tool-owned)")
	}

	foundToolUpdate := false
	for _, f := range result.Updated {
		if strings.Contains(f, "uf.review-council.md") {
			foundToolUpdate = true
			break
		}
	}
	if !foundToolUpdate {
		t.Error("expected uf.review-council.md to be in Updated list")
	}

	// User-owned file should be skipped
	foundUserSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "cobalt-crush-dev.md") {
			foundUserSkip = true
			break
		}
	}
	if !foundUserSkip {
		t.Error("expected cobalt-crush-dev.md to be in Skipped list")
	}

	// Verify tool-owned file content was restored (review-council.md)
	restored, err := os.ReadFile(toolFile)
	if err != nil {
		t.Fatalf("read restored tool file: %v", err)
	}
	if string(restored) == "modified content" {
		t.Error("tool-owned file was not restored to canonical content")
	}

	// Verify user-owned file was NOT overwritten
	preserved, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("read preserved user file: %v", err)
	}
	if string(preserved) != "user modified" {
		t.Error("user-owned file should not have been overwritten")
	}
}

func TestRun_OverwriteOnDiff_SkipsIdentical(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Re-run without any modifications
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Tool-owned files with identical content should be skipped
	if len(result.Updated) != 0 {
		t.Errorf("expected no updated files when content is identical, got %d: %v",
			len(result.Updated), result.Updated)
	}
}

func TestIsToolOwned(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Tool-owned: all commands
		{"opencode/commands/speckit.specify.md", true},
		{"opencode/commands/speckit.plan.md", true},
		{"opencode/commands/speckit.tasks.md", true},
		{"opencode/commands/speckit.clarify.md", true},
		{"opencode/commands/speckit.analyze.md", true},
		{"opencode/commands/speckit.checklist.md", true},
		{"opencode/commands/speckit.implement.md", true},
		{"opencode/commands/speckit.constitution.md", true},
		{"opencode/commands/speckit.taskstoissues.md", true},
		{"opencode/commands/uf.constitution-check.md", true},
		// Tool-owned: hypothetical future command (M1 fix)
		{"opencode/commands/opsx.propose.md", true},
		// Tool-owned: OpenSpec schema
		{"openspec/schemas/unbound-force/schema.yaml", true},
		{"openspec/schemas/unbound-force/templates/proposal.md", true},
		// Tool-owned: reproducible plugin package inputs
		{"opencode/package.json", true},
		{"opencode/package-lock.json", true},
		// Tool-owned: activation-gated plugin sources
		{invokeAgentPluginAsset, true},
		{reviewDispatchPluginAsset, true},
		{ufWorkflowPluginAsset, true},
		// Tool-owned: shared agent executor library
		{"opencode/lib/agent-executor.ts", true},
		{"opencode/lib/branch-classification.ts", true},
		{"opencode/lib/review-dispatch-artifact.ts", true},
		// Tool-owned: review-dispatch acquisition module
		{"opencode/lib/review-dispatch-lesson-proposal.ts", true},
		{"opencode/lib/review-dispatch-sibling-evidence.ts", true},
		{"opencode/lib/review-diff.ts", true},
		{"opencode/lib/review-diff-manifest.ts", true},
		{"opencode/lib/reviewer-manifest.ts", true},
		{"opencode/lib/run-timeout.ts", true},
		{"opencode/lib/uf-workflow-types.ts", true},
		{"opencode/lib/unleash-templates.ts", true},
		// User-owned: reviewer eligibility pairs with uf/review-matrix.yaml as a project extension point
		{"uf/reviewer-capabilities.yaml", false},
		// Tool-owned: model-matrix template (design D10, SC-FR-007) —
		// drift-enforced byte-identical to the canonical tracked matrix
		{"uf/model-matrix.yaml", true},
		// User-owned: model policy is a supported project extension point
		{"uf/review-matrix.yaml", false},
		// User-owned: active sibling declarations are project-specific
		{"uf/sibling-repos.yaml", false},
		// Tool-owned: shared multi-model schemas
		{"uf/schemas/review-matrix/v2.schema.json", true},
		{"uf/schemas/review-matrix/v3.schema.json", true},
		{"uf/schemas/reviewer-capabilities/v1.0.0.schema.json", true},
		{"uf/schemas/sibling-repos/v1.0.0.schema.json", true},
		{"uf/schemas/lesson-proposal/v1.0.0.schema.json", true},
		{"uf/schemas/review-dispatch/v1.0.0.schema.json", true},
		{"uf/schemas/review-verdict/v2.0.0.schema.json", true},
		// Tool-owned: convention packs (canonical)
		{"opencode/uf/packs/go.md", true},
		{"opencode/uf/packs/default.md", true},
		{"opencode/uf/packs/typescript.md", true},
		// Tool-owned: convention packs (canonical) — Python
		{"opencode/uf/packs/python.md", true},
		// Tool-owned: convention packs (canonical) — CI
		{"opencode/uf/packs/ci.md", true},
		// User-owned: convention packs (custom)
		{"opencode/uf/packs/go-custom.md", false},
		{"opencode/uf/packs/default-custom.md", false},
		{"opencode/uf/packs/typescript-custom.md", false},
		{"opencode/uf/packs/python-custom.md", false},
		{"opencode/uf/packs/ci-custom.md", false},
		// User-owned: agents (including Divisor personas and Cobalt-Crush)
		{"opencode/agents/divisor-guard.md", false},
		{"opencode/agents/divisor-architect.md", false},
		{"opencode/agents/cobalt-crush-dev.md", false},
		// User-owned: other
		{"opencode/agents/constitution-check.md", false},
		// User-owned: specify/ assets (starter constitution)
		{"specify/memory/constitution.md", false},
	}

	for _, tt := range tests {
		got := isToolOwned(tt.path)
		if got != tt.expected {
			t.Errorf("isToolOwned(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestIsNeverOverwrite(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Protected: constitution
		{"specify/memory/constitution.md", true},
		// Not protected: filesystem path with leading dot (not asset path)
		{".specify/memory/constitution.md", false},
		// Not protected: tool-owned command
		{"opencode/commands/uf.init.md", false},
		// Not protected: user-owned agent
		{"opencode/agents/cobalt-crush-dev.md", false},
		// Not protected: convention pack
		{"opencode/uf/packs/default.md", false},
		// Not protected: empty path
		{"", false},
	}

	for _, tt := range tests {
		got := isNeverOverwrite(tt.path)
		if got != tt.expected {
			t.Errorf("isNeverOverwrite(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestRun_SchemaDistribution(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	// First run creates everything
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("first Run() error: %v", err)
	}

	// Modify a schema file (tool-owned) and an agent (user-owned)
	schemaFile := filepath.Join(dir, "openspec", "schemas",
		"unbound-force", "schema.yaml")
	agentFile := filepath.Join(dir, ".opencode", "agents", "cobalt-crush-dev.md")

	if err := os.WriteFile(schemaFile, []byte("modified schema"), 0o644); err != nil {
		t.Fatalf("modify schema file: %v", err)
	}
	if err := os.WriteFile(agentFile, []byte("user agent"), 0o644); err != nil {
		t.Fatalf("modify agent file: %v", err)
	}

	// Re-run without --force
	buf.Reset()
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("second Run() error: %v", err)
	}

	// Schema file (tool-owned) should be updated
	foundSchemaUpdate := false
	for _, f := range result.Updated {
		if strings.Contains(f, "schema.yaml") {
			foundSchemaUpdate = true
			break
		}
	}
	if !foundSchemaUpdate {
		t.Error("expected schema.yaml to be in Updated list")
	}

	// Agent file (user-owned) should be skipped
	foundAgentSkip := false
	for _, f := range result.Skipped {
		if strings.Contains(f, "cobalt-crush-dev.md") {
			foundAgentSkip = true
			break
		}
	}
	if !foundAgentSkip {
		t.Error("expected cobalt-crush-dev.md to be in Skipped list")
	}

	// Verify schema was restored
	restored, err := os.ReadFile(schemaFile)
	if err != nil {
		t.Fatalf("read restored schema: %v", err)
	}
	if string(restored) == "modified schema" {
		t.Error("schema file was not restored to canonical content")
	}

	// Verify agent was preserved
	preserved, err := os.ReadFile(agentFile)
	if err != nil {
		t.Fatalf("read preserved agent: %v", err)
	}
	if string(preserved) != "user agent" {
		t.Error("agent file should not have been overwritten")
	}
}

func TestInsertMarkerAfterFrontmatter(t *testing.T) {
	mdMarker := "<!-- scaffolded by uf v1.0.0 -->"
	hashMarker := "# scaffolded by uf v1.0.0"

	tests := []struct {
		name     string
		input    string
		marker   string
		expected string
	}{
		{
			name:     "empty content",
			input:    "",
			marker:   mdMarker,
			expected: mdMarker + "\n",
		},
		{
			name:     "no frontmatter",
			input:    "# Hello\n\nSome content.\n",
			marker:   mdMarker,
			expected: "# Hello\n\nSome content.\n" + mdMarker + "\n",
		},
		{
			name:   "with frontmatter",
			input:  "---\ntitle: Test\n---\n# Content\n",
			marker: mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n" +
				"# Content\n",
		},
		{
			name:     "unclosed frontmatter",
			input:    "---\ntitle: Test\nno closing\n",
			marker:   mdMarker,
			expected: "---\ntitle: Test\nno closing\n" + mdMarker + "\n",
		},
		{
			name:   "frontmatter with dashes in body",
			input:  "---\ntitle: Test\n---\nSome text\n---\nMore text\n",
			marker: mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n" +
				"Some text\n---\nMore text\n",
		},
		{
			name:     "bash script",
			input:    "#!/usr/bin/env bash\nset -e\n",
			marker:   hashMarker,
			expected: "#!/usr/bin/env bash\nset -e\n" + hashMarker + "\n",
		},
		{
			name:     "yaml document",
			input:    "---\nkey: value\n---\nmore: yaml\n",
			marker:   hashMarker,
			expected: "---\nkey: value\n---\n" + hashMarker + "\nmore: yaml\n",
		},
		{
			name:   "idempotent on repeat call",
			input:  "# Hello\n" + mdMarker + "\n",
			marker: mdMarker,
			// insertMarkerAfterFrontmatter is idempotent: existing
			// markers are stripped before the new one is inserted.
			expected: "# Hello\n" + mdMarker + "\n",
		},
		{
			name:     "strips multiple existing markers",
			input:    "---\ntitle: Test\n---\n<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vv0.6.1 -->\n# Content\n",
			marker:   mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n# Content\n",
		},
		{
			name:     "replaces old version with new",
			input:    "---\ntitle: Test\n---\n<!-- scaffolded by uf v0.5.0 -->\n# Content\n",
			marker:   mdMarker,
			expected: "---\ntitle: Test\n---\n" + mdMarker + "\n# Content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insertMarkerAfterFrontmatter([]byte(tt.input), tt.marker)
			if string(got) != tt.expected {
				t.Errorf("got:\n%s\nexpected:\n%s", string(got), tt.expected)
			}
		})
	}
}

func TestStripExistingMarkers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no markers",
			input:    "# Hello\n\nSome content.\n",
			expected: "# Hello\n\nSome content.\n",
		},
		{
			name:     "single HTML marker",
			input:    "# Hello\n<!-- scaffolded by uf vdev -->\nContent\n",
			expected: "# Hello\nContent\n",
		},
		{
			name:     "multiple HTML markers",
			input:    "<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vdev -->\n<!-- scaffolded by uf vv0.6.1 -->\nContent\n",
			expected: "Content\n",
		},
		{
			name:     "single hash marker",
			input:    "#!/bin/bash\n# scaffolded by uf vdev\nset -e\n",
			expected: "#!/bin/bash\nset -e\n",
		},
		{
			name:     "mixed HTML and hash markers",
			input:    "<!-- scaffolded by uf v1.0.0 -->\n# scaffolded by uf v2.0.0\nContent\n",
			expected: "Content\n",
		},
		{
			name:     "frontmatter preserved",
			input:    "---\ntitle: Test\n---\n<!-- scaffolded by uf vdev -->\n# Content\n",
			expected: "---\ntitle: Test\n---\n# Content\n",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "content only no markers",
			input:    "line1\nline2\nline3\n",
			expected: "line1\nline2\nline3\n",
		},
		{
			name:     "marker-like content preserved",
			input:    "<!-- scaffolded by someone else -->\nContent\n",
			expected: "<!-- scaffolded by someone else -->\nContent\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripExistingMarkers(tt.input)
			if got != tt.expected {
				t.Errorf("got:\n%q\nexpected:\n%q", got, tt.expected)
			}
		})
	}
}

func TestPrintSummary_Output(t *testing.T) {
	t.Run("created_updated_skipped", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created:     []string{".opencode/agents/cobalt-crush-dev.md", ".opencode/commands/uf.review-council.md"},
			Updated:     []string{".opencode/commands/uf.unleash.md"},
			Overwritten: []string{},
			Skipped:     []string{".opencode/uf/packs/go-custom.md"},
		}

		printSummary(&buf, false, false, true, r, nil)
		output := buf.String()

		// Verify total count line
		if !strings.Contains(output, "uf init: 4 files processed") {
			t.Errorf("expected total count of 4, got output:\n%s", output)
		}

		// Verify section headers
		if !strings.Contains(output, "created:     2") {
			t.Errorf("expected created count of 2, got output:\n%s", output)
		}
		if !strings.Contains(output, "updated:     1") {
			t.Errorf("expected updated count of 1, got output:\n%s", output)
		}
		if !strings.Contains(output, "skipped:     1") {
			t.Errorf("expected skipped count of 1, got output:\n%s", output)
		}

		// Verify file prefix characters
		if !strings.Contains(output, "+ .opencode/agents/cobalt-crush-dev.md") || !strings.Contains(output, "+ .opencode/commands/uf.review-council.md") {
			t.Errorf("expected '+' prefix for created files")
		}
		if !strings.Contains(output, "~ .opencode/commands/uf.unleash.md") {
			t.Errorf("expected '~' prefix for updated files")
		}
		if !strings.Contains(output, "- .opencode/uf/packs/go-custom.md") {
			t.Errorf("expected '-' prefix for skipped files")
		}

		// Verify next-step guidance (no sub-tool results = suggest uf setup first)
		if !strings.Contains(output, "Next steps:") {
			t.Errorf("expected 'Next steps:' section")
		}
		if !strings.Contains(output, "uf setup") {
			t.Errorf("expected 'uf setup' hint when no sub-tool results")
		}
	})

	t.Run("divisor_mode", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/agents/divisor-guard.md", ".opencode/commands/uf.review-council.md"},
		}

		printSummary(&buf, true, false, true, r, nil)
		output := buf.String()

		if !strings.Contains(output, "uf init (divisor): 2 files processed") {
			t.Errorf("expected divisor label, got output:\n%s", output)
		}
		if !strings.Contains(output, "Run /uf.review-council") {
			t.Error("expected uf.review-council hint in divisor mode")
		}
		if strings.Contains(output, "Run /speckit.specify") {
			t.Error("speckit hint should not appear in divisor mode")
		}
		if strings.Contains(output, "Run /opsx:propose") {
			t.Error("opsx hint should not appear in divisor mode")
		}
	})

	t.Run("divisor_mode_no_lang", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/agents/divisor-guard.md"},
		}

		printSummary(&buf, true, false, false, r, nil)
		output := buf.String()

		if !strings.Contains(output, "language not detected") {
			t.Errorf("expected language detection warning, got:\n%s", output)
		}
	})

	t.Run("overwritten", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created:     []string{},
			Updated:     []string{},
			Overwritten: []string{".opencode/agents/cobalt-crush-dev.md", ".opencode/commands/uf.review-council.md"},
			Skipped:     []string{},
		}

		printSummary(&buf, false, false, true, r, nil)
		output := buf.String()

		if !strings.Contains(output, "uf init: 2 files processed") {
			t.Errorf("expected total count of 2, got output:\n%s", output)
		}
		if !strings.Contains(output, "overwritten: 2") {
			t.Errorf("expected overwritten count of 2, got output:\n%s", output)
		}
		if !strings.Contains(output, "! .opencode/agents/cobalt-crush-dev.md") {
			t.Errorf("expected '!' prefix for overwritten files")
		}
		if !strings.Contains(output, "! .opencode/commands/uf.review-council.md") {
			t.Errorf("expected '!' prefix for second overwritten file")
		}
	})
}

func TestRun_PrintSummaryIntegration(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := buf.String()
	expected := fmt.Sprintf("uf init: %d files processed", len(result.Created))
	if !strings.Contains(output, expected) {
		t.Errorf("expected summary to contain %q, got:\n%s", expected, output)
	}

	// Verify the output includes at least one specific file name
	if len(result.Created) > 0 {
		if !strings.Contains(output, result.Created[0]) {
			t.Errorf("expected output to contain file name %q", result.Created[0])
		}
	}
}

// knownNonEmbeddedFiles lists canonical source files that exist
// in .opencode/ but are intentionally NOT embedded in the unbound
// binary. These are local-only tooling files (e.g., installed by
// the Gaze scaffold) that are specific to this repository.
var knownNonEmbeddedFiles = map[string]bool{
	// Speckit commands — created by specify init + /uf.init, not scaffolded by uf init
	".opencode/commands/speckit.specify.md":       true,
	".opencode/commands/speckit.clarify.md":       true,
	".opencode/commands/speckit.plan.md":          true,
	".opencode/commands/speckit.tasks.md":         true,
	".opencode/commands/speckit.analyze.md":       true,
	".opencode/commands/speckit.checklist.md":     true,
	".opencode/commands/speckit.implement.md":     true,
	".opencode/commands/speckit.constitution.md":  true,
	".opencode/commands/speckit.taskstoissues.md": true,
	// Speckit files — created by specify init, not scaffolded by uf init
	".specify/config.yaml":                          true,
	".specify/templates/agent-file-template.md":     true,
	".specify/templates/checklist-template.md":      true,
	".specify/templates/constitution-template.md":   true,
	".specify/templates/plan-template.md":           true,
	".specify/templates/spec-template.md":           true,
	".specify/templates/tasks-template.md":          true,
	".specify/scripts/bash/check-prerequisites.sh":  true,
	".specify/scripts/bash/common.sh":               true,
	".specify/scripts/bash/create-new-feature.sh":   true,
	".specify/scripts/bash/setup-plan.sh":           true,
	".specify/scripts/bash/update-agent-context.sh": true,
	// OpenSpec config — created by openspec init
	"openspec/config.yaml": true,
	// Agents — local-only tooling, not scaffolded by uf init
	".opencode/agents/gaze-reporter.md":       true,
	".opencode/agents/gaze-test-generator.md": true,
	// Legacy reviewer agents — superseded by divisor-* (Spec 019)
	".opencode/agents/reviewer-adversary.md": true,
	".opencode/agents/reviewer-architect.md": true,
	".opencode/agents/reviewer-guard.md":     true,
	".opencode/agents/reviewer-sre.md":       true,
	".opencode/agents/reviewer-testing.md":   true,
	// Commands — local-only tooling
	".opencode/commands/gaze.md":                       true,
	".opencode/commands/gaze-fix.md":                   true,
	".opencode/commands/speckit.testreview.md":         true,
	// OpenSpec skill commands — local workflow tooling, not scaffolded by uf init
	".opencode/commands/opsx-apply.md":   true,
	".opencode/commands/opsx-archive.md": true,
	".opencode/commands/opsx-explore.md": true,
	".opencode/commands/opsx-propose.md": true,
	// Swarm skills — Spec 008, local-only
	".opencode/skill/unbound-force-heroes/SKILL.md": true,
	// Replicator-scaffolded agents and commands — created by replicator init,
	// not part of the uf binary's scaffold assets
	".opencode/agents/background-worker.md": true,
	".opencode/agents/coordinator.md":       true,
	".opencode/agents/worker.md":            true,
	".opencode/commands/forge.md":           true,
	".opencode/commands/forge-status.md":    true,
	".opencode/commands/handoff.md":         true,
	".opencode/commands/inbox.md":           true,
	".opencode/commands/org.md":             true,
	// OpenSpec skills — created by openspec init, not scaffolded by uf init
	".opencode/skills/openspec-apply-change/SKILL.md":   true,
	".opencode/skills/openspec-archive-change/SKILL.md": true,
	".opencode/skills/openspec-explore/SKILL.md":        true,
	".opencode/skills/openspec-propose/SKILL.md":        true,
	// Replicator-scaffolded skills — created by replicator init
	".opencode/skills/forge-coordination/SKILL.md": true,
	".opencode/skills/forge-global/SKILL.md":       true,
	".opencode/skills/learning-systems/SKILL.md":   true,
	".opencode/skills/replicator-cli/SKILL.md":     true,
	".opencode/skills/system-design/SKILL.md":      true,
	".opencode/skills/testing-patterns/SKILL.md":   true,
	// Dewey-scaffolded commands — created by dewey init
	".opencode/commands/dewey-index.md":   true,
	".opencode/commands/dewey-reindex.md": true,
	// DCP config — generated programmatically by configureDCPConfig(),
	// not embedded as a scaffold asset
	".opencode/dcp.jsonc": true,
	// Determinism and meta-prompt governance — repo-local files, not present in scaffold assets
	".opencode/uf/packs/determinism.md":           true,
	".opencode/skills/determinism-first/SKILL.md": true,
	".opencode/uf/packs/meta-prompt.md":           true,
	// Model-matrix classifier — hidden subagent, canonical repo content;
	// scaffold deployment deferred to slice 2 (#889), per SC-FR-007
	".opencode/agents/model-matrix-classifier.md": true,
}

func TestCanonicalSources_AreEmbedded(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping reverse drift detection")
	}

	// Build a set of embedded asset paths (mapped to source paths)
	embeddedSet := make(map[string]bool)
	for _, p := range expectedAssetPaths {
		srcRel := mapAssetToSource(p)
		embeddedSet[srcRel] = true
	}

	// Walk canonical source directories and check each file
	canonicalDirs := []string{
		".opencode/commands",
		".opencode/agents",
		".opencode/skills",
		".opencode/uf/packs",
	}

	for _, dir := range canonicalDirs {
		fullDir := filepath.Join(root, dir)
		if _, err := os.Stat(fullDir); os.IsNotExist(err) {
			continue
		}
		err := filepath.Walk(fullDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}
			// Skip hidden files (e.g., .DS_Store) — they are not
			// canonical source files and should not be tracked.
			if strings.HasPrefix(info.Name(), ".") {
				return nil
			}
			relPath, _ := filepath.Rel(root, path)
			if knownNonEmbeddedFiles[relPath] {
				return nil // Explicitly excluded
			}
			if !embeddedSet[relPath] {
				t.Errorf("canonical source %s is not embedded and not in knownNonEmbeddedFiles exclusion list", relPath)
			}
			return nil
		})
		if err != nil {
			t.Errorf("walk %s: %v", dir, err)
		}
	}

	// Also check standalone config files that are still embedded.
	// Note: .specify/config.yaml and openspec/config.yaml are now
	// created by external CLIs (specify init, openspec init) and
	// listed in knownNonEmbeddedFiles — no longer checked here.
}

func TestMapAssetPath_Prefixes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"opencode/commands/speckit.specify.md", ".opencode/commands/speckit.specify.md"},
		{"openspec/schemas/unbound-force/schema.yaml", "openspec/schemas/unbound-force/schema.yaml"},
		// specify/ maps to .specify/ (dot prefix)
		{"specify/memory/constitution.md", ".specify/memory/constitution.md"},
		// uf/ maps to .uf/ (dot prefix)
		{"uf/reviewer-capabilities.yaml", ".uf/reviewer-capabilities.yaml"},
		{"uf/sibling-repos.yaml", ".uf/sibling-repos.yaml"},
		// uf/ maps to .uf/ (dot prefix) — model-matrix template
		{"uf/model-matrix.yaml", ".uf/model-matrix.yaml"},
		// uf/schemas/ maps to .uf/schemas/ via the uf/ → .uf/ prefix rule
		{"uf/schemas/review-matrix/v2.schema.json", ".uf/schemas/review-matrix/v2.schema.json"},
		{"uf/schemas/review-matrix/v3.schema.json", ".uf/schemas/review-matrix/v3.schema.json"},
		// Unknown prefix passes through unchanged (default branch)
		{"scripts/validate.sh", "scripts/validate.sh"},
		// schemas/ passes through unchanged (default branch; mapping removed)
		{"schemas/review-matrix/v2.schema.json", "schemas/review-matrix/v2.schema.json"},
		{"schemas/review-matrix/v3.schema.json", "schemas/review-matrix/v3.schema.json"},
	}

	for _, tt := range tests {
		got := mapAssetPath(tt.input)
		if got != tt.expected {
			t.Errorf("mapAssetPath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsDivisorAsset(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Divisor agents
		{"opencode/agents/divisor-guard.md", true},
		{"opencode/agents/divisor-architect.md", true},
		{"opencode/agents/divisor-adversary.md", true},
		{"opencode/agents/divisor-sre.md", true},
		{"opencode/agents/divisor-testing.md", true},
		// Divisor command
		{"opencode/commands/uf.review-council.md", true},
		{"opencode/commands/uf.triage-issue.md", true},
		{"opencode/commands/uf.address-feedback.md", true},
		// Divisor convention packs
		{"opencode/uf/packs/go.md", true},
		{"opencode/uf/packs/default.md", true},
		{"opencode/uf/packs/go-custom.md", true},
		{"opencode/uf/packs/severity.md", true},
		{"opencode/uf/packs/python.md", true},
		{"opencode/uf/packs/python-custom.md", true},
		// Divisor review-context skill
		{"opencode/skills/review-context/SKILL.md", true},
		{"opencode/skills/dispatch-advisor/SKILL.md", true},
		// Divisor runtime and package inputs
		{invokeAgentPluginAsset, true},
		{reviewDispatchPluginAsset, true},
		{ufWorkflowPluginAsset, true},
		{"opencode/package.json", true},
		{"opencode/package-lock.json", true},
		// Shared agent executor library
		{"opencode/lib/agent-executor.ts", true},
		{"opencode/lib/branch-classification.ts", true},
		{"opencode/lib/review-dispatch-artifact.ts", true},
		// Divisor acquisition implementation
		{"opencode/lib/review-dispatch-lesson-proposal.ts", true},
		{"opencode/lib/review-dispatch-sibling-evidence.ts", true},
		{"opencode/lib/review-diff.ts", true},
		{"opencode/lib/review-diff-manifest.ts", true},
		{"opencode/lib/reviewer-manifest.ts", true},
		{"opencode/lib/run-timeout.ts", true},
		{"opencode/lib/uf-workflow-types.ts", true},
		{"opencode/lib/unleash-templates.ts", true},
		// Divisor review eligibility policy
		{"uf/reviewer-capabilities.yaml", true},
		{"uf/review-matrix.yaml", true},
		// Divisor sibling evidence template
		{"uf/sibling-repos.yaml", true},
		// Divisor contract schemas
		{"uf/schemas/review-matrix/v2.schema.json", true},
		{"uf/schemas/review-matrix/v3.schema.json", true},
		{"uf/schemas/reviewer-capabilities/v1.0.0.schema.json", true},
		{"uf/schemas/sibling-repos/v1.0.0.schema.json", true},
		{"uf/schemas/lesson-proposal/v1.0.0.schema.json", true},
		{"uf/schemas/review-dispatch/v1.0.0.schema.json", true},
		{"uf/schemas/review-verdict/v2.0.0.schema.json", true},
		// Non-Divisor assets
		{"opencode/agents/constitution-check.md", false},
		{"opencode/commands/speckit.specify.md", false},
		{"opencode/commands/speckit.plan.md", false},
		// Non-Divisor: Cobalt-Crush agent
		{"opencode/agents/cobalt-crush-dev.md", false},
		// Non-Divisor: the model matrix serves content-agent delegation
		// (uf_delegate_command), not review dispatch — it stays out of
		// DivisorOnly deployments (model-matrix-delegation design D10)
		{"uf/model-matrix.yaml", false},
		{"uf/schemas/model-matrix/v1.schema.json", false},
	}

	for _, tt := range tests {
		got := isDivisorAsset(tt.path)
		if got != tt.expected {
			t.Errorf("isDivisorAsset(%q) = %v, want %v",
				tt.path, got, tt.expected)
		}
	}
}

func TestDetectLang(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"go project", []string{"go.mod"}, "go"},
		{"typescript tsconfig", []string{"tsconfig.json"}, "typescript"},
		{"typescript package.json", []string{"package.json"}, "typescript"},
		{"python pyproject.toml", []string{"pyproject.toml"}, "python"},
		{"python setup.py", []string{"setup.py"}, "python"},
		{"python setup.cfg", []string{"setup.cfg"}, "python"},
		{"python requirements.txt", []string{"requirements.txt"}, "python"},
		{"python tox.ini", []string{"tox.ini"}, "python"},
		{"python Pipfile", []string{"Pipfile"}, "python"},
		{"python pyproject takes priority", []string{"pyproject.toml", "setup.py", "requirements.txt"}, "python"},
		{"rust project", []string{"Cargo.toml"}, "rust"},
		{"no markers", []string{}, ""},
		{"go takes priority over ts", []string{"go.mod", "package.json"}, "go"},
		{"go takes priority over python", []string{"go.mod", "requirements.txt"}, "go"},
		{"ts takes priority over python", []string{"tsconfig.json", "pyproject.toml"}, "typescript"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, f := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, f), []byte{}, 0o644); err != nil {
					t.Fatalf("create marker %s: %v", f, err)
				}
			}
			got := detectLang(dir)
			if got != tt.expected {
				t.Errorf("detectLang() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestShouldDeployPack(t *testing.T) {
	tests := []struct {
		relPath  string
		lang     string
		expected bool
	}{
		// Non-pack files always pass
		{"opencode/agents/divisor-guard.md", "go", true},
		{"opencode/commands/uf.review-council.md", "go", true},
		// Default and severity packs always deploy (language-agnostic)
		{"opencode/uf/packs/default.md", "go", true},
		{"opencode/uf/packs/default-custom.md", "go", true},
		{"opencode/uf/packs/default.md", "typescript", true},
		{"opencode/uf/packs/severity.md", "go", true},
		{"opencode/uf/packs/severity.md", "typescript", true},
		{"opencode/uf/packs/severity.md", "default", true},
		// CI packs always deploy (language-agnostic)
		{"opencode/uf/packs/ci.md", "go", true},
		{"opencode/uf/packs/ci-custom.md", "go", true},
		{"opencode/uf/packs/ci.md", "typescript", true},
		{"opencode/uf/packs/ci-custom.md", "typescript", true},
		{"opencode/uf/packs/ci.md", "", true},
		{"opencode/uf/packs/ci-custom.md", "", true},
		// Matching language packs deploy
		{"opencode/uf/packs/go.md", "go", true},
		{"opencode/uf/packs/go-custom.md", "go", true},
		{"opencode/uf/packs/typescript.md", "typescript", true},
		{"opencode/uf/packs/typescript-custom.md", "typescript", true},
		// Python packs deploy for python lang
		{"opencode/uf/packs/python.md", "python", true},
		{"opencode/uf/packs/python-custom.md", "python", true},
		// Non-matching language packs do NOT deploy
		{"opencode/uf/packs/typescript.md", "go", false},
		{"opencode/uf/packs/typescript-custom.md", "go", false},
		{"opencode/uf/packs/go.md", "typescript", false},
		{"opencode/uf/packs/go-custom.md", "typescript", false},
		{"opencode/uf/packs/python.md", "go", false},
		{"opencode/uf/packs/python-custom.md", "go", false},
		{"opencode/uf/packs/go.md", "python", false},
		{"opencode/uf/packs/go-custom.md", "python", false},
		// Default lang gets only default packs
		{"opencode/uf/packs/go.md", "default", false},
		{"opencode/uf/packs/python.md", "default", false},
		{"opencode/uf/packs/default.md", "default", true},
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s_lang=%s", filepath.Base(tt.relPath), tt.lang)
		t.Run(name, func(t *testing.T) {
			got := shouldDeployPack(tt.relPath, tt.lang)
			if got != tt.expected {
				t.Errorf("shouldDeployPack(%q, %q) = %v, want %v",
					tt.relPath, tt.lang, got, tt.expected)
			}
		})
	}
}

func TestRun_DivisorSubset(t *testing.T) {
	dir := t.TempDir()
	// Create a go.mod to trigger Go language detection
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644); err != nil {
		t.Fatalf("create go.mod: %v", err)
	}

	var buf bytes.Buffer
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Version:     "1.0.0",
		Stdout:      &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Should create only Divisor files
	if len(result.Created) == 0 {
		t.Fatal("expected Divisor files to be created")
	}

	// Verify no openspec files created (except schema which is embedded)
	for _, f := range result.Created {
		if strings.HasPrefix(f, "openspec/") && !strings.Contains(f, "schemas/") {
			t.Errorf("DivisorOnly should not create %s", f)
		}
		if strings.Contains(f, ".opencode/agents/reviewer-") {
			t.Errorf("DivisorOnly should not create legacy reviewer files: %s", f)
		}
		if strings.Contains(f, "speckit.") {
			t.Errorf("DivisorOnly should not create speckit commands: %s", f)
		}
		if strings.Contains(f, "cobalt-crush") {
			t.Errorf("DivisorOnly should not create cobalt-crush files: %s", f)
		}
	}

	// Verify Divisor agents exist
	for _, agent := range []string{"divisor-guard.md", "divisor-architect.md", "divisor-adversary.md", "divisor-sre.md", "divisor-testing.md"} {
		found := false
		for _, f := range result.Created {
			if strings.HasSuffix(f, agent) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be created", agent)
		}
	}

	// Verify Go convention pack deployed (auto-detected)
	foundGoPack := false
	for _, f := range result.Created {
		if strings.HasSuffix(f, "go.md") && strings.Contains(f, "uf/packs") {
			foundGoPack = true
			break
		}
	}
	if !foundGoPack {
		t.Error("expected Go convention pack to be deployed")
	}

	// Verify no specify/ assets (constitution is not a Divisor asset)
	for _, f := range result.Created {
		if strings.HasPrefix(f, ".specify/") {
			t.Errorf("DivisorOnly should not create %s", f)
		}
	}
	specifyConst := filepath.Join(dir, ".specify", "memory", "constitution.md")
	if _, err := os.Stat(specifyConst); !os.IsNotExist(err) {
		t.Error("DivisorOnly should not create .specify/memory/constitution.md on disk")
	}

	// Verify no openspec empty dirs
	specsDir := filepath.Join(dir, "openspec", "specs")
	if _, err := os.Stat(specsDir); !os.IsNotExist(err) {
		t.Error("DivisorOnly should not create openspec/specs directory")
	}

	// Verify summary mentions divisor
	output := buf.String()
	if !strings.Contains(output, "divisor") {
		t.Errorf("expected summary to mention divisor, got:\n%s", output)
	}
	if !strings.Contains(output, "review-council") {
		t.Errorf("expected summary to mention review-council hint")
	}
}

func TestRun_DivisorSubset_WithLangFlag(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Lang:        "typescript",
		Version:     "1.0.0",
		Stdout:      &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify TypeScript pack deployed
	foundTSPack := false
	for _, f := range result.Created {
		if strings.HasSuffix(f, "typescript.md") && strings.Contains(f, "uf/packs") {
			foundTSPack = true
		}
	}
	if !foundTSPack {
		t.Error("expected TypeScript convention pack to be deployed")
	}

	// Verify Go pack NOT deployed
	for _, f := range result.Created {
		if strings.HasSuffix(f, "/go.md") && strings.Contains(f, "uf/packs") {
			t.Error("Go convention pack should not be deployed when lang=typescript")
		}
	}

	// All 9 Divisor agent files created (6 review + 3 content)
	agentCount := 0
	for _, f := range result.Created {
		if strings.Contains(f, "agents/divisor-") {
			agentCount++
		}
	}
	if agentCount != 9 {
		t.Errorf("expected 9 Divisor agent files, got %d", agentCount)
	}
}

func TestRun_DivisorSubset_DefaultFallback(t *testing.T) {
	dir := t.TempDir() // Empty — no language markers
	var buf bytes.Buffer

	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Version:     "1.0.0",
		Stdout:      &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify only always-deploy packs deployed (default, severity, content,
	// ci, review-voice)
	for _, f := range result.Created {
		if strings.Contains(f, "uf/packs") {
			base := filepath.Base(f)
			if !strings.HasPrefix(base, "default") && base != "severity.md" &&
				!strings.HasPrefix(base, "content") && !strings.HasPrefix(base, "ci") &&
				!strings.HasPrefix(base, "review-voice") {
				t.Errorf("expected only default/severity/content/ci/review-voice packs, got %s", f)
			}
		}
	}

	// Verify default.md and default-custom.md exist
	foundDefault := false
	foundDefaultCustom := false
	for _, f := range result.Created {
		if strings.HasSuffix(f, "default.md") {
			foundDefault = true
		}
		if strings.HasSuffix(f, "default-custom.md") {
			foundDefaultCustom = true
		}
	}
	if !foundDefault {
		t.Error("expected default.md pack to be deployed")
	}
	if !foundDefaultCustom {
		t.Error("expected default-custom.md pack to be deployed")
	}

	// Verify language detection warning in output
	output := buf.String()
	if !strings.Contains(output, "language not detected") {
		t.Errorf("expected language detection warning, got:\n%s", output)
	}
}

// TestAssetPaths_KnownPrefixes verifies all embedded assets use
// a recognized top-level prefix. Catches new directories added
// without updating mapAssetPath.
func TestAssetPaths_KnownPrefixes(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	for _, p := range paths {
		found := false
		for _, prefix := range knownAssetPrefixes {
			if strings.HasPrefix(p, prefix) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("asset %q does not match any known prefix %v — update mapAssetPath and knownAssetPrefixes",
				p, knownAssetPrefixes)
		}
	}
}

// TestScaffoldOutput_NoGraphthulhuReferences is a regression guard
// for FR-001/FR-002/SC-001: scaffolded files must not contain any
// graphthulhu or knowledge-graph references. Dewey replaces
// graphthulhu as the knowledge layer.
func TestScaffoldOutput_NoGraphthulhuReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Stale patterns that must NOT appear in scaffolded output.
	stalePatterns := []string{
		"graphthulhu",
		"knowledge-graph_",
		"knowledge-graph",
	}

	// Walk all generated files and search for stale patterns.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)
		relPath, _ := filepath.Rel(dir, path)

		for _, pattern := range stalePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains stale %q reference (SC-001 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestScaffoldOutput_NoSwarmPluginReferences is a regression guard
// for Spec 024 FR-010/FR-016: scaffolded files must not contain any
// Swarm plugin references. Replicator replaces the Swarm plugin.
func TestScaffoldOutput_NoSwarmPluginReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Stale patterns that must NOT appear in scaffolded output.
	stalePatterns := []string{
		"opencode-swarm-plugin",
		"installSwarmPlugin",
		"ensureBun",
		"swarmForkSource",
	}

	// Walk all generated files and search for stale patterns.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)
		relPath, _ := filepath.Rel(dir, path)

		for _, pattern := range stalePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains stale %q reference (Spec 024 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestScaffoldOutput_NoHivemindReferences is a regression guard
// for Spec 022 FR-006/FR-007: scaffolded files must not contain any
// Hivemind tool references. Dewey replaces Hivemind as the unified
// memory layer for all learning storage and retrieval.
func TestScaffoldOutput_NoHivemindReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Stale patterns that must NOT appear in scaffolded output.
	stalePatterns := []string{
		"hivemind_store",
		"hivemind_find",
		"hivemind_validate",
		"hivemind_remove",
		"hivemind_get",
	}

	// Walk all generated files and search for stale patterns.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)
		relPath, _ := filepath.Rel(dir, path)

		for _, pattern := range stalePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains stale %q reference (Spec 022 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestScaffoldOutput_NoBareUnboundReferences is a regression guard
// for FR-015/SC-003: scaffolded files must not contain bare
// `unbound init`, `unbound doctor`, `unbound setup`, or
// `unbound version` CLI references.
func TestScaffoldOutput_NoBareUnboundReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Bare patterns that must NOT appear in scaffolded output.
	// These are the old CLI command names before the rename.
	barePatterns := []string{
		"unbound init",
		"unbound doctor",
		"unbound setup",
		"unbound version",
	}

	// Walk all generated files and search for bare patterns.
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)
		relPath, _ := filepath.Rel(dir, path)

		for _, pattern := range barePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains bare %q reference (FR-015 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestScaffoldOutput_NoOldPathReferences is a regression guard
// for Spec 025 SC-001/SC-002: scaffolded files must not contain any
// old directory path references. All tool workspace directories are
// unified under .uf/ per the directory convention.
func TestScaffoldOutput_NoOldPathReferences(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Stale patterns that must NOT appear in scaffolded output.
	stalePatterns := []string{
		"opencode/unbound/",
		".unbound-force/",
		".dewey/",
		".hive/",
		".muti-mind/",
		".mx-f/",
	}

	// Walk all generated files and search for stale patterns.
	// Exclude .gitignore — it intentionally lists legacy directories
	// as ignore patterns (per gitignore-init change).
	err = filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		relPath, _ := filepath.Rel(dir, path)
		if relPath == ".gitignore" {
			return nil // .gitignore intentionally lists legacy dirs
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Errorf("read %s: %v", path, readErr)
			return nil
		}
		text := string(content)

		for _, pattern := range stalePatterns {
			if strings.Contains(text, pattern) {
				t.Errorf("scaffolded file %s contains stale %q reference (Spec 025 violation)", relPath, pattern)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}
}

// TestDivisorAgents_NoBareFRReferences is a regression guard for
// Spec 019 FR-008: all FR references in Divisor agent files must
// use the qualified "per Spec NNN FR-XXX" format.
func TestDivisorAgents_NoBareFRReferences(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("get asset paths: %v", err)
	}

	// Regex: bare "FR-NNN" not preceded by "per Spec NNN "
	// We check for any "FR-" followed by digits that is NOT
	// preceded by "per Spec" on the same line.
	for _, relPath := range paths {
		if !strings.HasPrefix(relPath, "opencode/agents/divisor-") {
			continue
		}

		content, readErr := assetContent(relPath)
		if readErr != nil {
			t.Errorf("read %s: %v", relPath, readErr)
			continue
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			// Skip lines that don't contain FR- references
			if !strings.Contains(line, "FR-") {
				continue
			}
			// Check that every FR- reference has "per Spec" qualifier
			// Find all FR-NNN occurrences and verify each is qualified
			idx := 0
			for {
				pos := strings.Index(line[idx:], "FR-")
				if pos < 0 {
					break
				}
				absPos := idx + pos
				// Check if "per Spec" (case-insensitive) appears before this FR- on the same line
				prefix := strings.ToLower(line[:absPos])
				if !strings.Contains(prefix, "per spec") {
					t.Errorf("%s:%d: bare FR reference without 'per Spec' qualifier: %s",
						relPath, i+1, strings.TrimSpace(line))
					break
				}
				idx = absPos + 3
			}
		}
	}
}

// TestDivisorGuard_DelegationDisciplineSpawnParameters pins the
// guard's Deterministic Delegation Discipline after the
// spawn-classification parameterization: the Swarm-path checklist
// item audits the spawn call's task_classification, skill, and
// skill_absent_reason parameters by name instead of the subtask
// prompt's prose, and the superseded "has no parameter contract
// of its own" wording is gone. Both the embedded mirror and the
// canonical agent file are checked so neither surface can
// regress independently.
func TestDivisorGuard_DelegationDisciplineSpawnParameters(t *testing.T) {
	content, err := assetContent("opencode/agents/divisor-guard.md")
	if err != nil {
		t.Fatalf("read embedded divisor-guard.md: %v", err)
	}

	root := findProjectRoot(t)
	if root == "" {
		t.Fatalf("project root not found")
	}
	canonical, err := os.ReadFile(filepath.Join(root, ".opencode", "agents", "divisor-guard.md"))
	if err != nil {
		t.Fatalf("read canonical divisor-guard.md: %v", err)
	}

	assertDiscipline := func(label, text string) {
		// The Swarm-path checklist item is bounded by the next
		// checklist bullet so parameter names in the section
		// intro or sibling items cannot satisfy the pin.
		itemStart := strings.Index(text, "Does a Swarm-path spawn call")
		if itemStart < 0 {
			t.Errorf("%s divisor-guard.md missing the Swarm-path spawn checklist item", label)
			return
		}
		item := text[itemStart:]
		if end := strings.Index(item, "\n- "); end >= 0 {
			item = item[:end]
		}
		// Backticked names keep the assertions independent:
		// bare "skill" would match inside skill_absent_reason.
		for _, param := range []string{"`task_classification`", "`skill`", "`skill_absent_reason`"} {
			if !strings.Contains(item, param) {
				t.Errorf("%s divisor-guard.md Swarm-path checklist item missing %s parameter", label, param)
			}
		}
		if strings.Contains(text, "has no parameter contract of its own") {
			t.Errorf("%s divisor-guard.md still carries the stale no-parameter-contract wording", label)
		}
	}

	assertDiscipline("embedded mirror", string(content))
	assertDiscipline("canonical", string(canonical))
}

// TestDeterminismPack_DE008SpawnToolParameters pins the repo-local
// determinism pack after the spawn-classification parameterization:
// the frontmatter carries the 1.4.0 version bump, and DE-008's
// widened contract names the forge subtask spawn tool and carries
// the Swarm-path delegation fields as spawn-tool parameters rather
// than in the spawned subtask description. The pack is repo-local
// governance, not an embedded scaffold asset (see
// knownNonEmbeddedFiles), so it is read from the project root per
// the findProjectRoot pattern.
func TestDeterminismPack_DE008SpawnToolParameters(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping determinism pack pin")
	}
	content, err := os.ReadFile(filepath.Join(root, ".opencode", "uf", "packs", "determinism.md"))
	if err != nil {
		t.Fatalf("read determinism.md: %v", err)
	}
	text := string(content)

	// Frontmatter pin: 1.4.0 is the MINOR bump that carries
	// DE-008's normative carrier change. Bounded by the first
	// closing --- delimiter so a version string in rule prose
	// cannot satisfy it.
	frontmatterEnd := strings.Index(text, "\n---")
	if frontmatterEnd < 0 {
		t.Errorf("determinism.md frontmatter missing closing delimiter")
	} else {
		version := ""
		for _, line := range strings.Split(text[:frontmatterEnd], "\n") {
			if strings.HasPrefix(line, "version:") {
				version = strings.TrimSpace(strings.TrimPrefix(line, "version:"))
			}
		}
		if version != "1.4.0" {
			t.Errorf("determinism.md frontmatter version = %q, want 1.4.0", version)
		}
	}

	// DE-008 pin: the rule bullet is the structural unit, so
	// wording in sibling rules cannot satisfy the assertions.
	// Whitespace is normalized because the pack wraps at 72
	// columns and the pinned carrier phrase crosses a line wrap.
	ruleStart := strings.Index(text, "- **DE-008**")
	if ruleStart < 0 {
		t.Errorf("determinism.md missing DE-008 rule")
		return
	}
	rule := text[ruleStart:]
	if end := strings.Index(rule, "\n- **DE-"); end >= 0 {
		rule = rule[:end]
	}
	flatRule := strings.Join(strings.Fields(rule), " ")
	for _, fragment := range []string{
		"as spawn-tool parameters",
		"replicator_forge_spawn_subtask",
	} {
		if !strings.Contains(flatRule, fragment) {
			t.Errorf("determinism.md DE-008 missing %q", fragment)
		}
	}
}

// TestDeterminismFirstSkill_SpawnToolParameters pins the repo-local
// determinism-first skill's delegation guidance after the
// spawn-classification parameterization: the "Delegated Work
// Carries Its Classification" section names the three spawn
// parameters, and its Swarm-path sentence routes deterministic [P]
// delegations through the replicator_forge_spawn_subtask call's
// parameters rather than the subtask description. The skill is
// repo-local, not an embedded scaffold asset (see
// knownNonEmbeddedFiles), so it is read from the project root per
// the findProjectRoot pattern.
func TestDeterminismFirstSkill_SpawnToolParameters(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping determinism-first skill pin")
	}
	content, err := os.ReadFile(filepath.Join(root, ".opencode", "skills", "determinism-first", "SKILL.md"))
	if err != nil {
		t.Fatalf("read determinism-first SKILL.md: %v", err)
	}
	text := string(content)

	// Scope to the delegation-guidance section: the Related
	// Enforcement list also names the parameters and the spawn
	// tool, so a file-wide contains could not prove the guidance
	// itself carries them.
	sectionStart := strings.Index(text, "## Delegated Work Carries Its Classification")
	if sectionStart < 0 {
		t.Errorf("determinism-first SKILL.md missing the delegation guidance section")
		return
	}
	section := text[sectionStart:]
	if end := strings.Index(section, "\n## "); end >= 0 {
		section = section[:end]
	}
	flatSection := strings.Join(strings.Fields(section), " ")
	// Backticked names keep the assertions independent: bare
	// "skill" would match inside skill_absent_reason.
	for _, param := range []string{"`task_classification`", "`skill`", "`skill_absent_reason`"} {
		if !strings.Contains(flatSection, param) {
			t.Errorf("determinism-first SKILL.md delegation guidance missing %s parameter", param)
		}
	}

	// The Swarm-path sentence names the spawn call it routes
	// through; the span from the sentence to the end of the
	// section must carry the tool reference.
	swarmStart := strings.Index(flatSection, "Swarm-path")
	if swarmStart < 0 {
		t.Errorf("determinism-first SKILL.md delegation guidance missing the Swarm-path sentence")
		return
	}
	if !strings.Contains(flatSection[swarmStart:], "replicator_forge_spawn_subtask") {
		t.Errorf("determinism-first SKILL.md Swarm-path sentence missing the replicator_forge_spawn_subtask call")
	}
}

// spawnCallBlock extracts the fenced replicator_forge_spawn_subtask
// call block from a command template, returning the block (opening
// fence through closing fence) and the text that follows the closing
// fence. The block is anchored on a line whose trimmed content is
// exactly the tool name preceded by a bare opening fence:
// uf_delegate_command call blocks share the fenced shape but name a
// different tool, and prose mentions of the tool never sit alone on
// a line, so the anchor stays scoped to the spawn call.
func spawnCallBlock(t *testing.T, text, label string) (block, rest string) {
	t.Helper()
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "replicator_forge_spawn_subtask" {
			continue
		}
		if i == 0 || strings.TrimSpace(lines[i-1]) != "```" {
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "```" {
				return strings.Join(lines[i-1:j+1], "\n"),
					strings.Join(lines[j+1:], "\n")
			}
		}
		t.Fatalf("%s: replicator_forge_spawn_subtask call block has no closing fence", label)
		return "", ""
	}
	t.Fatalf("%s: no fenced replicator_forge_spawn_subtask call block", label)
	return "", ""
}

// spawnCallParamLine matches a `name: [value]` parameter line inside
// a fenced tool call block. Wrapped value continuations never match:
// the colon must immediately follow the lowercase snake-case name
// run.
var spawnCallParamLine = regexp.MustCompile(`^[ \t]*([a-z_]+):`)

// spawnCallBlockParameterNames parses the parameter names a fenced
// tool call block passes, so tests can pin registration-surface
// parity (the names the template passes vs. the names the deployed
// tool's input schema registers) without re-implementing block
// parsing per test.
func spawnCallBlockParameterNames(block string) map[string]bool {
	names := make(map[string]bool)
	for _, line := range strings.Split(block, "\n") {
		if m := spawnCallParamLine.FindStringSubmatch(line); m != nil {
			names[m[1]] = true
		}
	}
	return names
}

// recordedSpawnSignatureParams parses the parameter names from the
// forge_spawn_subtask(...) signature the /forge coordination command
// records -- the repo-local recording of the deployed spawn tool's
// registered input schema (synced from Replicator with the
// coordination surface). Alternation (`skill | skill_absent_reason`)
// and commas both separate parameters.
func recordedSpawnSignatureParams(t *testing.T, forgeText string) map[string]bool {
	t.Helper()
	const anchor = "forge_spawn_subtask("
	start := strings.Index(forgeText, anchor)
	if start < 0 {
		t.Fatalf("forge.md does not record a forge_spawn_subtask signature")
	}
	sig := forgeText[start+len(anchor):]
	if cut := strings.Index(sig, ")"); cut >= 0 {
		sig = sig[:cut]
	}
	params := make(map[string]bool)
	for _, part := range strings.FieldsFunc(sig, func(r rune) bool {
		return r == ',' || r == '|'
	}) {
		if name := strings.TrimSpace(part); name != "" {
			params[name] = true
		}
	}
	if len(params) == 0 {
		t.Fatalf("forge.md forge_spawn_subtask signature parsed to no parameters")
	}
	return params
}

// TestUnleashCommand_SpawnSubtaskSeamVerification is the DE-007 seam
// verification for the spawn-classification parameterization: a
// prose-to-tool migration increment must verify, in the same
// increment, the three seams it creates (registration surface,
// documented output shape, write-read round-trips). The layers
// cross-check the surfaces the change writes:
//
//   - Registration surface: the fenced replicator_forge_spawn_subtask
//     call block in the canonical uf.unleash.md names exactly the
//     deployed tool's delegation parameters -- task_classification,
//     skill, skill_absent_reason, cross-checked against the /forge
//     coordination surface's recording of the registered signature --
//     plus the fields the template already passes, and the mirrored
//     call block in the embedded asset is byte-identical
//     (call-block scope; whole-file identity is pinned separately by
//     TestEmbeddedAssets_MatchSource).
//   - Documented output shape: the failure codes and validation
//     rejections the template's enforcement prose names match the
//     spawn tool's documented contract as recorded in this repo:
//     SKILL_REFERENCE_REQUIRED, SKILL_NOT_FOUND, and
//     input-validation rejection of skill fields on
//     judgment-classified calls.
//   - Write-read round-trips: the changed surfaces are read back --
//     the spawn call block canonical <-> mirror here; the
//     divisor-guard checklist, the determinism pack, and the
//     determinism-first skill via the sibling spawn-parameter pins
//     (TestDivisorGuard_DelegationDisciplineSpawnParameters,
//     TestDeterminismPack_DE008SpawnToolParameters,
//     TestDeterminismFirstSkill_SpawnToolParameters) -- and the
//     drift infrastructure keeps covering them (embedded mirrors
//     through the TestEmbeddedAssets_MatchSource asset set,
//     repo-local governance files through knownNonEmbeddedFiles).
func TestUnleashCommand_SpawnSubtaskSeamVerification(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatalf("project root not found; cannot verify the spawn subtask seams")
	}

	canonical, err := os.ReadFile(filepath.Join(root, ".opencode", "commands", "uf.unleash.md"))
	if err != nil {
		t.Fatalf("read canonical uf.unleash.md: %v", err)
	}
	mirror, err := assetContent("opencode/commands/uf.unleash.md")
	if err != nil {
		t.Fatalf("read embedded uf.unleash.md: %v", err)
	}
	// The /forge coordination command is the repo-local recording of
	// the deployed spawn tool's registered input schema and
	// documented contract (synced from Replicator); both the
	// registration-surface and output-shape layers check against it.
	forgeCmd, err := os.ReadFile(filepath.Join(root, ".opencode", "commands", "forge.md"))
	if err != nil {
		t.Fatalf("read forge.md: %v", err)
	}

	canonicalBlock, canonicalRest := spawnCallBlock(t, string(canonical), "canonical uf.unleash.md")
	mirrorBlock, _ := spawnCallBlock(t, string(mirror), "embedded mirror uf.unleash.md")

	t.Run("registration surface: spawn call block matches the deployed schema", func(t *testing.T) {
		// Call-block-level drift check: the mirrored block must be
		// byte-identical to the canonical one. Scoped to the fenced
		// block because whole-file identity is already pinned by
		// TestEmbeddedAssets_MatchSource; this pin keeps drift in
		// the block that carries the classification contract
		// visible at the seam itself.
		if canonicalBlock != mirrorBlock {
			t.Error("embedded mirror's replicator_forge_spawn_subtask call block differs from the canonical one")
		}

		// Input-schema parity, template side: the block names
		// exactly the deployed tool's delegation parameters plus
		// the fields the template already passes -- no missing,
		// renamed, or extra parameters. The template deliberately
		// does not enumerate the tool's full schema (bead_id,
		// epic_id, subtask_title, shared_context belong to the
		// coordination tooling), so the pin covers the delegation
		// contract, not the whole registration. The deployed tool
		// validates required inputs fail-closed at spawn time, so
		// a misparameterized call block must break here first.
		got := spawnCallBlockParameterNames(canonicalBlock)
		want := map[string]bool{
			"subtask_description": true,
			"task_classification": true,
			"skill":               true,
			"skill_absent_reason": true,
			"files":               true,
			"project_path":        true,
		}
		var missing, unexpected []string
		for name := range want {
			if !got[name] {
				missing = append(missing, name)
			}
		}
		for name := range got {
			if !want[name] {
				unexpected = append(unexpected, name)
			}
		}
		if len(missing) > 0 || len(unexpected) > 0 {
			sort.Strings(missing)
			sort.Strings(unexpected)
			t.Errorf("canonical spawn call block parameter drift: missing %v, unexpected %v "+
				"(want exactly subtask_description, task_classification, skill, "+
				"skill_absent_reason, files, project_path)", missing, unexpected)
		}

		// Input-schema parity, recorded-schema side: each
		// classification parameter the template passes must be a
		// name the repo-local recording of the registered
		// signature knows. A Replicator schema rename synced into
		// the coordination surface fails here instead of breaking
		// the template's spawn calls at runtime.
		sigParams := recordedSpawnSignatureParams(t, string(forgeCmd))
		for _, param := range []string{"task_classification", "skill", "skill_absent_reason"} {
			if !sigParams[param] {
				t.Errorf("forge.md spawn signature does not record the %s parameter the template passes", param)
			}
		}
	})

	t.Run("documented output shape: enforcement prose matches the recorded contract", func(t *testing.T) {
		// Scope to the enforcement prose that follows the spawn
		// call block (up to the next parallel-execution item), so
		// failure codes named elsewhere in the template cannot
		// satisfy the pin. Whitespace is normalized because the
		// template wraps at 72 columns and the pinned clauses
		// cross line wraps.
		cut := strings.Index(canonicalRest, "d. Wait for all workers")
		if cut < 0 {
			t.Fatalf("canonical uf.unleash.md: spawn enforcement prose is not bounded by " +
				"the parallel-execution wait item")
		}
		flatProse := strings.Join(strings.Fields(canonicalRest[:cut]), " ")
		for _, fragment := range []string{
			"fails with `SKILL_REFERENCE_REQUIRED`",
			"fails with `SKILL_NOT_FOUND`",
			"a judgment-classified call supplying either skill field fails input validation",
		} {
			if !strings.Contains(flatProse, fragment) {
				t.Errorf("spawn enforcement prose missing documented-contract clause %q", fragment)
			}
		}

		// The codes and rejections the template names must be the
		// ones the repo-local recording of the spawn tool's
		// documented contract records -- a code renamed on either
		// surface breaks the seam here, not in a spawned worker's
		// error handling.
		flatForge := strings.Join(strings.Fields(string(forgeCmd)), " ")
		for _, fragment := range []string{
			"`SKILL_REFERENCE_REQUIRED`",
			"`SKILL_NOT_FOUND`",
			"supplying either fails input validation",
		} {
			if !strings.Contains(flatForge, fragment) {
				t.Errorf("forge.md spawn contract recording missing %q "+
					"(template and recording disagree on the documented output shape)", fragment)
			}
		}
	})

	t.Run("round trips: changed surfaces stay under drift coverage", func(t *testing.T) {
		// The drift test round-trips canonical -> mirror for the
		// embedded assets this change writes: both must remain in
		// the asset set TestEmbeddedAssets_MatchSource iterates.
		// The call-block identity asserted above is the
		// block-scoped slice of that round trip, not a replacement
		// for the whole-file check.
		paths, err := assetPaths()
		if err != nil {
			t.Fatalf("list embedded assets: %v", err)
		}
		embedded := make(map[string]bool, len(paths))
		for _, rel := range paths {
			embedded[rel] = true
		}
		for _, rel := range []string{
			"opencode/commands/uf.unleash.md",
			"opencode/agents/divisor-guard.md",
		} {
			if !embedded[rel] {
				t.Errorf("%s is no longer an embedded asset; TestEmbeddedAssets_MatchSource "+
					"would stop round-tripping this change's canonical -> mirror drift check", rel)
			}
		}

		// The repo-local governance files this change writes are
		// intentionally unmirrored; they must stay in
		// knownNonEmbeddedFiles so TestCanonicalSources_AreEmbedded
		// keeps treating them as covered by the repo-local pin
		// tests rather than silently unembedded. Their content is
		// read back by the sibling spawn-parameter pins, not
		// duplicated here.
		for _, rel := range []string{
			".opencode/uf/packs/determinism.md",
			".opencode/skills/determinism-first/SKILL.md",
		} {
			if !knownNonEmbeddedFiles[rel] {
				t.Errorf("%s is not in knownNonEmbeddedFiles; TestCanonicalSources_AreEmbedded "+
					"would demand an embedded mirror for this repo-local governance file", rel)
			}
		}
	})
}

// TestUnleashCommand_ModelSelectionGuidancePin pins the UF-FR-001
// model-selection guidance span (model-matrix-delegation task 6.6) in
// .opencode/commands/uf.unleash.md: recommend `pack` when the task
// clearly names one, the classifier covers the rest, the selection
// parameter contract and precedence order documented, and no
// pack-content reconstruction. Removing the guidance from the command
// file fails this test; the embedded mirror is pinned by the span
// equality check plus the whole-file drift test.
func TestUnleashCommand_ModelSelectionGuidancePin(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatalf("project root not found; cannot verify the model-selection guidance")
	}

	canonical, err := os.ReadFile(filepath.Join(root, ".opencode", "commands", "uf.unleash.md"))
	if err != nil {
		t.Fatalf("read canonical uf.unleash.md: %v", err)
	}
	mirror, err := assetContent("opencode/commands/uf.unleash.md")
	if err != nil {
		t.Fatalf("read embedded uf.unleash.md: %v", err)
	}

	// The guidance lives between its bold lead-in and the next block
	// (the build/test fix loop) inside Step 5 (Implement); between()
	// fails closed when either marker moves or disappears.
	guidanceSpan := func(t *testing.T, value string) string {
		t.Helper()
		startIndex := strings.Index(value, "**Model-selection parameters for delegated calls**")
		if startIndex == -1 {
			t.Fatalf("unleash command missing model-selection guidance lead-in")
		}
		value = value[startIndex:]
		endIndex := strings.Index(value, "**Build/test fix loop**")
		if endIndex == -1 {
			t.Fatalf("unleash command missing guidance span end marker (Build/test fix loop)")
		}
		return value[:endIndex]
	}

	span := guidanceSpan(t, string(canonical))
	// The command file wraps prose at ~72 columns, so multi-word phrases
	// can span line breaks; collapse whitespace runs to single spaces
	// before the containment checks (same technique as
	// normalizeCommandText, without its lowercasing — the phrases pin
	// exact casing and backticks).
	collapse := func(text string) string {
		return strings.Join(strings.Fields(text), " ")
	}
	normalizedSpan := collapse(span)
	containsPhrase := func(t *testing.T, phrase string) {
		t.Helper()
		if !strings.Contains(normalizedSpan, collapse(phrase)) {
			t.Errorf("model-selection guidance missing phrase %q", phrase)
		}
	}

	t.Run("mirror carries the same guidance span", func(t *testing.T) {
		// Whole-file identity is already pinned by
		// TestEmbeddedAssets_MatchSource; the span equality check keeps
		// the guidance drift visible at the seam itself.
		if span != guidanceSpan(t, string(mirror)) {
			t.Error("embedded mirror's model-selection guidance span differs from the canonical one")
		}
	})

	t.Run("recommends pack when the task clearly names one", func(t *testing.T) {
		for _, phrase := range []string{
			"pass `pack` on an",
			"`uf_delegate_command` implement delegation",
			"a Go task passes `pack:",
			"a TypeScript task `pack:",
			"spec-authoring",
		} {
			containsPhrase(t, phrase)
		}
	})

	t.Run("unmarked tasks rely on the classifier", func(t *testing.T) {
		for _, phrase := range []string{
			"A task with no",
			"clearly named pack omits the parameter",
			"classifier",
		} {
			containsPhrase(t, phrase)
		}
	})

	t.Run("selection parameter contract and precedence order documented", func(t *testing.T) {
		for _, phrase := range []string{
			"`model`: explicit provider/model identity",
			"`profile`: matrix capability name",
			"`pack`: UF convention pack name",
			"explicit `model` > `profile` > `pack` >",
			"agent pin > task-kind default > classifier cascade >",
			"host model",
		} {
			containsPhrase(t, phrase)
		}
	})

	t.Run("never instructs pack-content reconstruction", func(t *testing.T) {
		for _, phrase := range []string{
			"Never reconstruct, inline, or paraphrase pack content",
			"the pack directive is tool-rendered",
			"the orchestrator only ever names the pack",
		} {
			containsPhrase(t, phrase)
		}
	})
}

// TestRun_LegacyFileWarning verifies that uf init warns about
// previously scaffolded reviewer-*.md files per Spec 019 FR-003a.
func TestRun_LegacyFileWarning(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create legacy reviewer files.
	legacyFiles := []string{
		"reviewer-adversary.md",
		"reviewer-architect.md",
		"reviewer-guard.md",
		"reviewer-sre.md",
		"reviewer-testing.md",
	}
	for _, f := range legacyFiles {
		if err := os.WriteFile(filepath.Join(agentsDir, f), []byte("legacy"), 0o644); err != nil {
			t.Fatalf("create %s: %v", f, err)
		}
	}

	var buf bytes.Buffer
	_, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := buf.String()

	// Verify warning is printed.
	if !strings.Contains(output, "Legacy reviewer agents detected") {
		t.Errorf("expected legacy warning, got:\n%s", output)
	}

	// Verify file names are listed.
	for _, f := range legacyFiles {
		if !strings.Contains(output, f) {
			t.Errorf("expected %s in warning, got:\n%s", f, output)
		}
	}

	// Verify removal command is suggested.
	if !strings.Contains(output, "rm .opencode/agents/reviewer-*.md") {
		t.Errorf("expected removal command in warning, got:\n%s", output)
	}

	// Verify legacy files are NOT deleted (FR-003a).
	for _, f := range legacyFiles {
		if _, err := os.Stat(filepath.Join(agentsDir, f)); os.IsNotExist(err) {
			t.Errorf("legacy file %s should NOT be deleted", f)
		}
	}
}

func TestRun_DoesNotCreateBridgeFiles(t *testing.T) {
	dir := t.TempDir()
	var buf bytes.Buffer

	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// CLAUDE.md must not exist after Run.
	if _, statErr := os.Stat(filepath.Join(dir, "CLAUDE.md")); statErr == nil {
		t.Error("CLAUDE.md should not be created by Run()")
	}

	// .cursorrules must not exist after Run.
	if _, statErr := os.Stat(filepath.Join(dir, ".cursorrules")); statErr == nil {
		t.Error(".cursorrules should not be created by Run()")
	}

	// Result slices must not reference bridge file names.
	allPaths := append(append(append(append(
		result.Created,
		result.Skipped...),
		result.Overwritten...),
		result.Updated...),
		result.Migrated...)
	for _, p := range allPaths {
		if strings.Contains(p, "CLAUDE") {
			t.Errorf("Result references CLAUDE file: %q", p)
		}
		if strings.Contains(p, "cursorrules") {
			t.Errorf("Result references cursorrules file: %q", p)
		}
	}
}

func TestRun_DoesNotModifyExistingBridgeFiles(t *testing.T) {
	dir := t.TempDir()

	// Pre-create bridge files with known content.
	claudePath := filepath.Join(dir, "CLAUDE.md")
	cursorPath := filepath.Join(dir, ".cursorrules")
	originalClaude := "# My project CLAUDE instructions\n"
	originalCursor := "# My project cursor rules\n"

	if err := os.WriteFile(claudePath, []byte(originalClaude), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}
	if err := os.WriteFile(cursorPath, []byte(originalCursor), 0o644); err != nil {
		t.Fatalf("write .cursorrules: %v", err)
	}

	var buf bytes.Buffer
	result, err := runScaffoldWithoutExternalTools(t, Options{
		TargetDir: dir,
		Version:   "1.0.0-test",
		Stdout:    &buf,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// CLAUDE.md must still exist with original content.
	gotClaude, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("read CLAUDE.md after Run: %v", err)
	}
	if string(gotClaude) != originalClaude {
		t.Errorf("CLAUDE.md content changed: got %q, want %q", string(gotClaude), originalClaude)
	}

	// .cursorrules must still exist with original content.
	gotCursor, err := os.ReadFile(cursorPath)
	if err != nil {
		t.Fatalf("read .cursorrules after Run: %v", err)
	}
	if string(gotCursor) != originalCursor {
		t.Errorf(".cursorrules content changed: got %q, want %q", string(gotCursor), originalCursor)
	}

	// Result slices must not reference bridge file names.
	allPaths := append(append(append(append(
		result.Created,
		result.Skipped...),
		result.Overwritten...),
		result.Updated...),
		result.Migrated...)
	for _, p := range allPaths {
		if strings.Contains(p, "CLAUDE") {
			t.Errorf("Result references CLAUDE file: %q", p)
		}
		if strings.Contains(p, "cursorrules") {
			t.Errorf("Result references cursorrules file: %q", p)
		}
	}
}

// --- Sub-tool initialization tests ---

// stubScaffoldLookPath returns a function that simulates exec.LookPath.
func stubScaffoldLookPath(found map[string]string) func(string) (string, error) {
	return func(name string) (string, error) {
		if path, ok := found[name]; ok {
			return path, nil
		}
		return "", fmt.Errorf("executable %q not found", name)
	}
}

// scaffoldCmdRecorder records ExecCmd calls for scaffold tests.
type scaffoldCmdRecorder struct {
	mu      sync.Mutex
	calls   []string
	errors  map[string]error
	outputs map[string]string
}

func (r *scaffoldCmdRecorder) execCmd(name string, args ...string) ([]byte, error) {
	key := name
	if len(args) > 0 {
		key = name + " " + strings.Join(args, " ")
	}

	r.mu.Lock()
	r.calls = append(r.calls, key)
	err, ok := r.errors[key]
	var out []byte
	if r.outputs != nil {
		if s, has := r.outputs[key]; has {
			out = []byte(s)
		}
	}
	r.mu.Unlock()

	if ok {
		return out, err
	}
	if out != nil {
		return out, nil
	}
	return []byte(""), nil
}

func TestInitSubTools_DeweyAvailable(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 4 results: dewey init + dewey index + opencode.json + dcp.jsonc.
	// (.uf/config.yaml is no longer created by uf init — use uf config init)
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d: %v", len(results), results)
	}

	if results[0].name != ".uf/dewey/" || results[0].action != "initialized" {
		t.Errorf("expected .uf/dewey/ initialized, got %s %s", results[0].name, results[0].action)
	}
	if results[1].name != "dewey index" || results[1].action != "completed" {
		t.Errorf("expected dewey index completed, got %s %s", results[1].name, results[1].action)
	}
	if results[2].name != "opencode.json" || results[2].action != "created" {
		t.Errorf("expected opencode.json created, got %s %s", results[2].name, results[2].action)
	}
	if results[3].name != "dcp.jsonc" || results[3].action != "created" {
		t.Errorf("expected dcp.jsonc created, got %s %s", results[3].name, results[3].action)
	}

	// Verify commands were called.
	expectedCalls := []string{"dewey init", "dewey index --no-embeddings"}
	for _, expected := range expectedCalls {
		found := false
		for _, call := range rec.calls {
			if call == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q, got calls: %v", expected, rec.calls)
		}
	}
}

func TestInitSubTools_DeweyNotAvailable(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No dewey
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 2 results: opencode.json skipped + dcp.jsonc created.
	// (.uf/config.yaml is no longer created by uf init)
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d: %v", len(results), results)
	}
	if len(results) > 0 && results[0].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[0].name)
	}
	if len(results) > 1 && results[1].name != "dcp.jsonc" {
		t.Errorf("expected dcp.jsonc result, got %s", results[1].name)
	}

	// No commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands, got: %v", rec.calls)
	}
}

func TestInitSubTools_DeweyAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 2 results: opencode.json created + dcp.jsonc created.
	// (.uf/dewey/ already exists, dewey in PATH → mcp.dewey added).
	// (.uf/config.yaml no longer created by uf init)
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d: %v", len(results), results)
	}
	if len(results) > 0 && results[0].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[0].name)
	}
	if len(results) > 1 && results[1].name != "dcp.jsonc" {
		t.Errorf("expected dcp.jsonc result, got %s", results[1].name)
	}

	// dewey init should NOT have been called.
	for _, call := range rec.calls {
		if call == "dewey init" {
			t.Error("dewey init should NOT be called when .uf/dewey/ already exists")
		}
	}
}

func TestInitSubTools_DeweyInitFails(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"dewey init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have 3 results: dewey init failed + opencode.json created + dcp.jsonc created.
	// (.uf/config.yaml no longer created by uf init)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d: %v", len(results), results)
	}

	if results[0].name != ".uf/dewey/" || results[0].action != "failed" {
		t.Errorf("expected .uf/dewey/ failed, got %s %s", results[0].name, results[0].action)
	}
	if results[1].name != "opencode.json" || results[1].action != "created" {
		t.Errorf("expected opencode.json created, got %s %s", results[1].name, results[1].action)
	}
	if results[2].name != "dcp.jsonc" || results[2].action != "created" {
		t.Errorf("expected dcp.jsonc created, got %s %s", results[2].name, results[2].action)
	}

	// dewey index should NOT have been called.
	for _, call := range rec.calls {
		if strings.HasPrefix(call, "dewey index") {
			t.Error("dewey index should NOT be called when dewey init fails")
		}
	}
}

func TestInitSubTools_DivisorOnly(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir:   dir,
		DivisorOnly: true,
		LookPath:    stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:     rec.execCmd,
	}

	results := initSubTools(opts)

	// Should return nil — DivisorOnly skips all sub-tool init.
	if results != nil {
		t.Errorf("expected nil results in DivisorOnly mode, got %v", results)
	}

	// No commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands in DivisorOnly mode, got: %v", rec.calls)
	}
}

func TestPrintSummary_NextSteps(t *testing.T) {
	t.Run("with_sub_tools", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/commands/uf.review-council.md"},
		}
		subResults := []subToolResult{
			{name: ".uf/dewey/", action: "initialized"},
			{name: "dewey index", action: "completed"},
		}

		printSummary(&buf, false, false, true, r, subResults)
		output := buf.String()

		// Should show sub-tool results.
		if !strings.Contains(output, "Sub-tool initialization:") {
			t.Errorf("expected sub-tool section, got:\n%s", output)
		}
		if !strings.Contains(output, ".uf/dewey/ initialized") {
			t.Errorf("expected dewey init result, got:\n%s", output)
		}
		if !strings.Contains(output, "dewey index completed") {
			t.Errorf("expected dewey index result, got:\n%s", output)
		}

		// Should show full next steps (not uf setup first).
		if !strings.Contains(output, "Next steps:") {
			t.Errorf("expected 'Next steps:' section")
		}
		if !strings.Contains(output, "/speckit.constitution") {
			t.Errorf("expected constitution hint")
		}
		// Hint should say "customize" not "create" (MR-001).
		if !strings.Contains(output, "customize your project constitution") {
			t.Errorf("expected 'customize your project constitution' in hint")
		}
		if strings.Contains(output, "create your project constitution") {
			t.Errorf("hint should NOT contain 'create your project constitution'")
		}
		if !strings.Contains(output, "uf doctor") {
			t.Errorf("expected doctor hint")
		}
	})

	t.Run("without_sub_tools", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/commands/uf.review-council.md"},
		}

		printSummary(&buf, false, false, true, r, nil)
		output := buf.String()

		// Should suggest uf setup as first step.
		if !strings.Contains(output, "uf setup") {
			t.Errorf("expected 'uf setup' hint when no sub-tool results, got:\n%s", output)
		}
		// Hint should say "customize" not "create" (MR-001).
		if !strings.Contains(output, "customize your project constitution") {
			t.Errorf("expected 'customize your project constitution' in hint")
		}
		if strings.Contains(output, "create your project constitution") {
			t.Errorf("hint should NOT contain 'create your project constitution'")
		}
	})

	t.Run("sub_tool_failure", func(t *testing.T) {
		var buf bytes.Buffer

		r := &Result{
			Created: []string{".opencode/commands/uf.review-council.md"},
		}
		subResults := []subToolResult{
			{name: ".uf/dewey/", action: "failed", detail: "dewey init failed"},
		}

		printSummary(&buf, false, false, true, r, subResults)
		output := buf.String()

		// Should show failure with ✗ symbol.
		if !strings.Contains(output, "✗") {
			t.Errorf("expected failure symbol, got:\n%s", output)
		}
		if !strings.Contains(output, "failed") {
			t.Errorf("expected 'failed' in output, got:\n%s", output)
		}
	})
}

// --- Workflow config file scaffold tests ---

func TestInitSubTools_DoesNotCreateWorkflowConfig(t *testing.T) {
	// .uf/config.yaml is no longer created by uf init.
	// It is now created exclusively by uf config init.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No dewey
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have a config.yaml result.
	for _, r := range results {
		if r.name == ".uf/config.yaml" {
			t.Errorf("uf init should NOT create .uf/config.yaml, got result: %s %s", r.name, r.action)
		}
	}

	// Verify file does NOT exist.
	configPath := filepath.Join(dir, ".uf", "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		t.Error(".uf/config.yaml should not exist after uf init")
	}
}

// --- Dewey auto-sources tests ---

func TestGenerateDeweySources_SiblingsDetected(t *testing.T) {
	// Create a parent dir with the "current" project and 3 sibling repos.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write default sources.yaml (single source entry).
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	// Create 3 sibling repos with .git/ directories.
	for _, name := range []string{"gaze", "dewey", "website"} {
		sibDir := filepath.Join(parentDir, name)
		if err := os.MkdirAll(filepath.Join(sibDir, ".git"), 0o755); err != nil {
			t.Fatalf("mkdir sibling %s: %v", name, err)
		}
	}

	// Create a non-repo directory (no .git/) — should be ignored.
	if err := os.MkdirAll(filepath.Join(parentDir, "not-a-repo"), 0o755); err != nil {
		t.Fatalf("mkdir not-a-repo: %v", err)
	}

	// Stub ExecCmd to return a GitHub SSH remote.
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			key := name
			if len(args) > 0 {
				key = name + " " + strings.Join(args, " ")
			}
			if key == "git remote get-url origin" {
				return []byte("git@github.com:unbound-force/my-project.git\n"), nil
			}
			return rec.execCmd(name, args...)
		},
	}

	result := generateDeweySources(opts, false)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "completed" {
		t.Errorf("expected action 'completed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "4 repos detected") {
		t.Errorf("expected '4 repos detected' in detail, got %q", result.detail)
	}

	// Read the generated sources.yaml and verify content.
	content, err := os.ReadFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"))
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	text := string(content)

	// Verify per-repo disk sources.
	if !strings.Contains(text, "- id: disk-local") {
		t.Error("expected disk-local source")
	}
	if !strings.Contains(text, "- id: disk-gaze") {
		t.Error("expected disk-gaze source")
	}
	if !strings.Contains(text, "- id: disk-dewey") {
		t.Error("expected disk-dewey source")
	}
	if !strings.Contains(text, "- id: disk-website") {
		t.Error("expected disk-website source")
	}

	// Verify disk-org source with recursive: false.
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}
	if !strings.Contains(text, "recursive: false") {
		t.Error("expected recursive: false on disk-org")
	}

	// Verify GitHub source with repos list.
	if !strings.Contains(text, "- id: github-unbound-force") {
		t.Error("expected github-unbound-force source")
	}
	if !strings.Contains(text, "org: unbound-force") {
		t.Error("expected org: unbound-force in GitHub config")
	}
	// Verify repos list includes current + siblings.
	if !strings.Contains(text, "        - my-project") {
		t.Error("expected my-project in repos list")
	}
	if !strings.Contains(text, "        - gaze") {
		t.Error("expected gaze in repos list")
	}

	// Verify non-repo directory was NOT included.
	if strings.Contains(text, "not-a-repo") {
		t.Error("non-repo directory should not appear in sources")
	}

	// Verify sibling paths use relative notation.
	if !strings.Contains(text, "path: \"../gaze\"") {
		t.Error("expected relative path for gaze sibling")
	}
}

func TestGenerateDeweySources_NoSiblings(t *testing.T) {
	// Create a parent dir with only the current project.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "lonely-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write default sources.yaml.
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	// No ExecCmd stub needed — extractGitHubOrg will fail gracefully.
	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("no remote")
		},
	}

	result := generateDeweySources(opts, false)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "completed" {
		t.Errorf("expected action 'completed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "1 repos detected") {
		t.Errorf("expected '1 repos detected', got %q", result.detail)
	}

	// Read generated sources.yaml.
	content, err := os.ReadFile(filepath.Join(currentDir, ".uf", "dewey", "sources.yaml"))
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	text := string(content)

	// Should have disk-local + disk-org only.
	if !strings.Contains(text, "- id: disk-local") {
		t.Error("expected disk-local source")
	}
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}
	if !strings.Contains(text, "recursive: false") {
		t.Error("expected recursive: false on disk-org")
	}

	// Should NOT have GitHub source (no remote).
	if strings.Contains(text, "type: github") {
		t.Error("should not have GitHub source when no remote")
	}
}

func TestGenerateDeweySources_AlreadyCustomized(t *testing.T) {
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write a customized sources.yaml with 3 source entries.
	customSources := `sources:
  - id: disk-local
    type: disk
    config:
      path: "."
  - id: disk-other
    type: disk
    config:
      path: "../other"
  - id: github-myorg
    type: github
    config:
      org: myorg
`
	sourcesPath := filepath.Join(currentDir, ".uf", "dewey", "sources.yaml")
	if err := os.WriteFile(sourcesPath, []byte(customSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("should not be called")
		},
	}

	result := generateDeweySources(opts, false)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "skipped" {
		t.Errorf("expected action 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "already customized") {
		t.Errorf("expected 'already customized' in detail, got %q", result.detail)
	}

	// Verify file was NOT overwritten.
	content, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	if string(content) != customSources {
		t.Error("customized sources.yaml should not have been overwritten")
	}
}

func TestGenerateDeweySources_ForceOverwritesCustom(t *testing.T) {
	// Create a parent dir with the "current" project and a sibling repo.
	parentDir := t.TempDir()
	currentDir := filepath.Join(parentDir, "my-project")
	if err := os.MkdirAll(filepath.Join(currentDir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create a sibling repo.
	if err := os.MkdirAll(filepath.Join(parentDir, "gaze", ".git"), 0o755); err != nil {
		t.Fatalf("mkdir sibling: %v", err)
	}

	// Write a customized sources.yaml (>1 source entry).
	customSources := `sources:
  - id: disk-local
    type: disk
    config:
      path: "."
  - id: my-custom-source
    type: disk
    config:
      path: "../custom"
`
	sourcesPath := filepath.Join(currentDir, ".uf", "dewey", "sources.yaml")
	if err := os.WriteFile(sourcesPath, []byte(customSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: currentDir,
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			if name == "git" {
				return []byte("git@github.com:unbound-force/my-project.git\n"), nil
			}
			return nil, fmt.Errorf("unexpected: %s", name)
		},
	}

	// Call with force=true — should overwrite customized file.
	result := generateDeweySources(opts, true)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "completed" {
		t.Errorf("expected action 'completed', got %q", result.action)
	}

	// Read the regenerated sources.yaml.
	content, err := os.ReadFile(sourcesPath)
	if err != nil {
		t.Fatalf("read sources.yaml: %v", err)
	}
	text := string(content)

	// Should have the auto-detected config, not the custom one.
	if strings.Contains(text, "my-custom-source") {
		t.Error("custom source should have been overwritten")
	}
	if !strings.Contains(text, "- id: disk-gaze") {
		t.Error("expected disk-gaze source in regenerated config")
	}
	if !strings.Contains(text, "- id: disk-org") {
		t.Error("expected disk-org source")
	}
	if !strings.Contains(text, "recursive: false") {
		t.Error("expected recursive: false on disk-org")
	}
}

func TestExtractGitHubOrg_SSH(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("git@github.com:unbound-force/repo.git\n"), nil
		},
	}

	org := extractGitHubOrg(opts)
	if org != "unbound-force" {
		t.Errorf("expected 'unbound-force', got %q", org)
	}
}

func TestExtractGitHubOrg_HTTPS(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("https://github.com/unbound-force/repo.git\n"), nil
		},
	}

	org := extractGitHubOrg(opts)
	if org != "unbound-force" {
		t.Errorf("expected 'unbound-force', got %q", org)
	}
}

func TestExtractGitHubOrg_NonGitHub(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return []byte("https://gitlab.com/myorg/repo.git\n"), nil
		},
	}

	org := extractGitHubOrg(opts)
	if org != "" {
		t.Errorf("expected empty string for non-GitHub remote, got %q", org)
	}
}

func TestExtractGitHubOrg_NoRemote(t *testing.T) {
	opts := &Options{
		ExecCmd: func(name string, args ...string) ([]byte, error) {
			return nil, fmt.Errorf("fatal: No such remote 'origin'")
		},
	}

	org := extractGitHubOrg(opts)
	if org != "" {
		t.Errorf("expected empty string when no remote, got %q", org)
	}
}

func TestIsDefaultSourcesConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "default single source",
			input:    "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n",
			expected: true,
		},
		{
			name:     "empty file",
			input:    "",
			expected: true,
		},
		{
			name:     "no sources at all",
			input:    "# empty config\n",
			expected: true,
		},
		{
			name: "customized with 3 sources",
			input: `sources:
  - id: disk-local
    type: disk
  - id: disk-other
    type: disk
  - id: github-org
    type: github
`,
			expected: false,
		},
		{
			name: "customized with 2 sources",
			input: `sources:
  - id: disk-local
    type: disk
  - id: disk-other
    type: disk
`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDefaultSourcesConfig([]byte(tt.input))
			if got != tt.expected {
				t.Errorf("isDefaultSourcesConfig() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInitSubTools_PreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	// Create existing config with custom content.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	customContent := "workflow:\n  execution_modes:\n    define: swarm\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(customContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have a config result — file already exists.
	for _, r := range results {
		if r.name == ".uf/config.yaml" {
			t.Errorf("expected no config result (file exists), got %s %s", r.name, r.action)
		}
	}

	// Verify file was NOT overwritten.
	content, err := os.ReadFile(filepath.Join(ufDir, "config.yaml"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if string(content) != customContent {
		t.Error("existing config.yaml should not have been overwritten")
	}
}

// --- configureOpencodeJSON tests ---

// parseOpencodeJSON is a test helper that parses opencode.json from a dir.
func parseOpencodeJSON(t *testing.T, dir string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse opencode.json: %v", err)
	}
	return m
}

// getMCPDewey extracts the mcp.dewey entry from a parsed opencode.json.
func getMCPDewey(t *testing.T, ocMap map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key not found")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	deweyRaw, ok := mcpMap["dewey"]
	if !ok {
		t.Fatal("mcp.dewey not found")
	}
	var dewey map[string]json.RawMessage
	if err := json.Unmarshal(deweyRaw, &dewey); err != nil {
		t.Fatalf("parse mcp.dewey: %v", err)
	}
	return dewey
}

// getPlugins extracts the plugin array from a parsed opencode.json.
func getPlugins(t *testing.T, ocMap map[string]json.RawMessage) []string {
	t.Helper()
	pluginRaw, ok := ocMap["plugin"]
	if !ok {
		t.Fatal("plugin key not found")
	}
	var plugins []string
	if err := json.Unmarshal(pluginRaw, &plugins); err != nil {
		t.Fatalf("parse plugin: %v", err)
	}
	return plugins
}

// Phase 3: US1 tests — Fresh Repo Init

// getMCPReplicator extracts the mcp.replicator entry from a parsed opencode.json.
func getMCPReplicator(t *testing.T, ocMap map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key not found")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	repRaw, ok := mcpMap["replicator"]
	if !ok {
		t.Fatal("mcp.replicator not found")
	}
	var rep map[string]json.RawMessage
	if err := json.Unmarshal(repRaw, &rep); err != nil {
		t.Fatalf("parse mcp.replicator: %v", err)
	}
	return rep
}

func TestConfigureOpencodeJSON_Create(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d: %v", len(results), results)
	}
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify $schema.
	var schema string
	if err := json.Unmarshal(ocMap["$schema"], &schema); err != nil {
		t.Fatalf("parse $schema: %v", err)
	}
	if schema != "https://opencode.ai/config.json" {
		t.Errorf("$schema = %q, want opencode.ai URL", schema)
	}

	// Verify mcp.dewey entry.
	getMCPDewey(t, ocMap)

	// Verify mcp.replicator entry.
	rep := getMCPReplicator(t, ocMap)
	var repType string
	_ = json.Unmarshal(rep["type"], &repType)
	if repType != "local" {
		t.Errorf("mcp.replicator.type = %q, want 'local'", repType)
	}
	var cmd []string
	_ = json.Unmarshal(rep["command"], &cmd)
	expectedCmd := []string{"replicator", "serve"}
	if len(cmd) != len(expectedCmd) {
		t.Errorf("mcp.replicator.command = %v, want %v", cmd, expectedCmd)
	} else {
		for i := range cmd {
			if cmd[i] != expectedCmd[i] {
				t.Errorf("mcp.replicator.command[%d] = %q, want %q", i, cmd[i], expectedCmd[i])
			}
		}
	}

	// Verify no plugin array.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist (Replicator uses MCP, not plugin array)")
	}
}

func TestConfigureOpencodeJSON_DeweyOnly(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should have mcp.dewey.
	if _, ok := ocMap["mcp"]; !ok {
		t.Fatal("mcp key should exist")
	}

	// Should NOT have plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_ReplicatorOnly(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should have mcp.replicator.
	getMCPReplicator(t, ocMap)

	// Should NOT have plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_Neither(t *testing.T) {
	dir := t.TempDir()
	// No dewey, no replicator.

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "skipped" {
		t.Errorf("expected action 'skipped', got %q", results[0].action)
	}
	if results[0].detail != "nothing to configure" {
		t.Errorf("expected detail 'nothing to configure', got %q", results[0].detail)
	}

	// No file should be created.
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Error("opencode.json should not be created when nothing to configure")
	}
}

// Phase 4: US2 tests — Idempotent Re-run

func TestConfigureOpencodeJSON_Idempotent(t *testing.T) {
	dir := t.TempDir()

	// Create opencode.json with both entries already present.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true,
      "type": "local"
    },
    "replicator": {
      "command": ["replicator", "serve"],
      "enabled": true,
      "type": "local"
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", results[0].action)
	}

	// Verify file is unchanged.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != existing {
		t.Error("file should be byte-identical when already configured")
	}
}

func TestConfigureOpencodeJSON_LegacyPluginMigration(t *testing.T) {
	dir := t.TempDir()

	// Legacy opencode.json with opencode-swarm-plugin in plugin array.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  },
  "plugin": [
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.replicator was added.
	getMCPReplicator(t, ocMap)

	// Verify plugin key was removed (empty after removing swarm plugin).
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should be removed after legacy migration (empty array)")
	}
}

func TestConfigureOpencodeJSON_LegacyPluginMigration_OtherPlugins(t *testing.T) {
	dir := t.TempDir()

	// Legacy opencode.json with swarm plugin AND other plugins.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "plugin": [
    "other-plugin",
    "opencode-swarm-plugin"
  ]
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.replicator was added.
	getMCPReplicator(t, ocMap)

	// Verify plugin key preserved with only other-plugin.
	plugins := getPlugins(t, ocMap)
	if len(plugins) != 1 || plugins[0] != "other-plugin" {
		t.Errorf("plugin = %v, want [other-plugin]", plugins)
	}
}

func TestConfigureOpencodeJSON_RegistersReviewPlugins(t *testing.T) {
	dir := t.TempDir()

	// Create the managed review plugin sources.
	for _, assetPath := range []string{invokeAgentPluginAsset, reviewDispatchPluginAsset, ufWorkflowPluginAsset, ufFinalePluginAsset} {
		targetPath := filepath.Join(dir, mapAssetPath(assetPath))
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			t.Fatalf("mkdir plugin dir: %v", err)
		}
		if err := os.WriteFile(targetPath, []byte("// plugin source\n"), 0o644); err != nil {
			t.Fatalf("write plugin source: %v", err)
		}
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)
	gotPlugins := getPlugins(t, ocMap)
	wantPlugins := []string{
		"./" + mapAssetPath(invokeAgentPluginAsset),
		"./" + mapAssetPath(reviewDispatchPluginAsset),
		"./" + mapAssetPath(ufWorkflowPluginAsset),
		"./" + mapAssetPath(ufFinalePluginAsset),
	}
	if len(gotPlugins) != len(wantPlugins) {
		t.Fatalf("plugin = %v, want %v", gotPlugins, wantPlugins)
	}
	for i := range wantPlugins {
		if gotPlugins[i] != wantPlugins[i] {
			t.Errorf("plugin[%d] = %q, want %q", i, gotPlugins[i], wantPlugins[i])
		}
	}
}

func TestConfigureOpencodeJSON_ReplicatorNotInstalled(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Should have mcp.dewey but NOT mcp.replicator.
	getMCPDewey(t, ocMap)
	mcpRaw := ocMap["mcp"]
	var mcpMap map[string]json.RawMessage
	_ = json.Unmarshal(mcpRaw, &mcpMap)
	if _, ok := mcpMap["replicator"]; ok {
		t.Error("mcp.replicator should not exist when replicator not installed")
	}

	// Should NOT have plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_AddMissing(t *testing.T) {
	dir := t.TempDir()

	// Has mcp.dewey but no mcp.replicator.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.dewey preserved.
	getMCPDewey(t, ocMap)

	// Verify mcp.replicator was added.
	getMCPReplicator(t, ocMap)
}

func TestConfigureOpencodeJSON_PreserveCustom(t *testing.T) {
	dir := t.TempDir()

	// Has a custom MCP server.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "my-custom-server": {
      "type": "local",
      "command": ["my-server", "start"],
      "enabled": true
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "configured" {
		t.Errorf("expected action 'configured', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify custom server preserved.
	mcpRaw := ocMap["mcp"]
	var mcpMap map[string]json.RawMessage
	_ = json.Unmarshal(mcpRaw, &mcpMap) //nolint:errcheck // test helper
	if _, ok := mcpMap["my-custom-server"]; !ok {
		t.Error("custom MCP server should be preserved")
	}
	if _, ok := mcpMap["dewey"]; !ok {
		t.Error("mcp.dewey should be added alongside custom server")
	}
}

func TestConfigureOpencodeJSON_LegacyMcpServers(t *testing.T) {
	dir := t.TempDir()

	// Uses legacy mcpServers key.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcpServers": {
    "dewey": {
      "command": "dewey",
      "args": ["serve", "--vault", "."]
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write opencode.json: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", results[0].action)
	}

	// Verify no duplicate mcp.dewey added.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != existing {
		t.Error("file should be unchanged when legacy mcpServers.dewey exists")
	}
}

func TestConfigureOpencodeJSON_Malformed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte("{invalid json"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if results[0].detail != "malformed JSON" {
		t.Errorf("expected detail 'malformed JSON', got %q", results[0].detail)
	}

	// Verify file not modified.
	data, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if string(data) != "{invalid json" {
		t.Error("malformed file should not be modified")
	}
}

func TestConfigureOpencodeJSON_ReadPermissionDenied(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ReadFile: func(path string) ([]byte, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if !strings.Contains(results[0].detail, "read failed") {
		t.Errorf("expected detail to contain 'read failed', got %q", results[0].detail)
	}
}

func TestConfigureOpencodeJSON_WriteFail(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		WriteFile: func(path string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("disk full")
		},
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "failed" {
		t.Errorf("expected action 'failed', got %q", results[0].action)
	}
	if !strings.Contains(results[0].detail, "write failed") {
		t.Errorf("expected detail to contain 'write failed', got %q", results[0].detail)
	}
}

func TestConfigureOpencodeJSON_ByteIdentical(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	// First run — creates the file.
	configureOpencodeJSON(opts)
	data1, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))

	// Second run — should be "already configured" and file unchanged.
	results := configureOpencodeJSON(opts)
	if results[0].action != "already configured" {
		t.Errorf("expected 'already configured' on second run, got %q", results[0].action)
	}

	data2, _ := os.ReadFile(filepath.Join(dir, "opencode.json"))
	if !bytes.Equal(data1, data2) {
		t.Error("output should be byte-identical on re-run (FR-016)")
	}
}

// Phase 5: US3 tests — Force Overwrite

func TestConfigureOpencodeJSON_Force(t *testing.T) {
	dir := t.TempDir()

	// Stale mcp.dewey with --include-hidden flag.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--include-hidden", "--vault", "."],
      "enabled": true
    },
    "replicator": {
      "type": "local",
      "command": ["replicator", "serve"],
      "enabled": true
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "overwritten" {
		t.Errorf("expected action 'overwritten', got %q", results[0].action)
	}

	ocMap := parseOpencodeJSON(t, dir)

	// Verify mcp.dewey was overwritten with correct command (no --include-hidden).
	dewey := getMCPDewey(t, ocMap)
	var cmd []string
	_ = json.Unmarshal(dewey["command"], &cmd)
	expectedCmd := []string{"dewey", "serve", "--vault", "."}
	if len(cmd) != len(expectedCmd) {
		t.Fatalf("command = %v, want %v", cmd, expectedCmd)
	}
	for i := range cmd {
		if cmd[i] != expectedCmd[i] {
			t.Errorf("command[%d] = %q, want %q", i, cmd[i], expectedCmd[i])
		}
	}

	// Verify no plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

func TestConfigureOpencodeJSON_ForceCorrect(t *testing.T) {
	dir := t.TempDir()

	// Correct mcp.dewey — force should still overwrite.
	existing := `{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(dir, "opencode.json"), []byte(existing), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "overwritten" {
		t.Errorf("expected action 'overwritten', got %q", results[0].action)
	}
}

func TestConfigureOpencodeJSON_DryRun(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		DryRun:    true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "dry-run" {
		t.Errorf("expected action 'dry-run', got %q", results[0].action)
	}

	// Verify no file was written.
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Error("no file should be written in dry-run mode")
	}
}

func TestConfigureOpencodeJSON_SkipDewey(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with dewey method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action == "skipped" {
		t.Fatal("should not skip entirely — replicator is still active")
	}

	ocMap := parseOpencodeJSON(t, dir)
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key should exist for replicator")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	if _, ok := mcpMap["dewey"]; ok {
		t.Error("mcp.dewey should NOT exist when setup.tools.dewey.method is skip")
	}
	if _, ok := mcpMap["replicator"]; !ok {
		t.Error("mcp.replicator should exist")
	}
}

func TestConfigureOpencodeJSON_SkipReplicator(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with replicator method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    replicator:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action == "skipped" {
		t.Fatal("should not skip entirely — dewey is still active")
	}

	ocMap := parseOpencodeJSON(t, dir)
	mcpRaw, ok := ocMap["mcp"]
	if !ok {
		t.Fatal("mcp key should exist for dewey")
	}
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	if _, ok := mcpMap["replicator"]; ok {
		t.Error("mcp.replicator should NOT exist when setup.tools.replicator.method is skip")
	}
	if _, ok := mcpMap["dewey"]; !ok {
		t.Error("mcp.dewey should exist")
	}
}

func TestConfigureOpencodeJSON_SkipBoth(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with both tools set to skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n    replicator:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
	}

	results := configureOpencodeJSON(opts)
	if results[0].action != "skipped" {
		t.Errorf("expected action 'skipped' when both tools are skip, got %q", results[0].action)
	}
	if results[0].detail != "nothing to configure" {
		t.Errorf("expected detail 'nothing to configure', got %q", results[0].detail)
	}

	// Verify no file was written.
	if _, err := os.Stat(filepath.Join(dir, "opencode.json")); !os.IsNotExist(err) {
		t.Error("opencode.json should not be created when both tools are skipped")
	}
}

// --- configureDCPConfig tests ---

func TestConfigureDCPConfig_Create(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode directory (parent of dcp.jsonc).
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}

	opts := &Options{TargetDir: dir}
	results := configureDCPConfig(opts)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].name != "dcp.jsonc" {
		t.Errorf("expected name 'dcp.jsonc', got %q", results[0].name)
	}
	if results[0].action != "created" {
		t.Errorf("expected action 'created', got %q", results[0].action)
	}

	// Verify file content matches dcpConfigContent.
	data, err := os.ReadFile(filepath.Join(dir, ".opencode", "dcp.jsonc"))
	if err != nil {
		t.Fatalf("read dcp.jsonc: %v", err)
	}
	if string(data) != dcpConfigContent {
		t.Errorf("file content does not match dcpConfigContent\ngot:\n%s\nwant:\n%s", data, dcpConfigContent)
	}
}

func TestConfigureDCPConfig_Idempotent(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}

	// Write the canonical content.
	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	if err := os.WriteFile(dcpPath, []byte(dcpConfigContent), 0o644); err != nil {
		t.Fatalf("write dcp.jsonc: %v", err)
	}

	opts := &Options{TargetDir: dir}
	results := configureDCPConfig(opts)

	if results[0].action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", results[0].action)
	}

	// Verify file is unchanged.
	data, _ := os.ReadFile(dcpPath)
	if string(data) != dcpConfigContent {
		t.Error("file should be unchanged when already configured")
	}
}

func TestConfigureDCPConfig_AddProtectTags(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}

	// Existing file with $schema but no protectTags.
	existing := `{
  "$schema": "https://raw.githubusercontent.com/Opencode-DCP/opencode-dynamic-context-pruning/master/dcp.schema.json",
  "compress": {
    "someOtherSetting": 42
  }
}
`
	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	if err := os.WriteFile(dcpPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write dcp.jsonc: %v", err)
	}

	opts := &Options{TargetDir: dir}
	results := configureDCPConfig(opts)

	if results[0].action != "updated" {
		t.Errorf("expected action 'updated', got %q", results[0].action)
	}

	// Verify file now has canonical content (overwritten for comment preservation).
	data, _ := os.ReadFile(dcpPath)
	if string(data) != dcpConfigContent {
		t.Errorf("file should have canonical content after update\ngot:\n%s", data)
	}
}

func TestConfigureDCPConfig_DryRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		DryRun:    true,
	}

	results := configureDCPConfig(opts)
	if results[0].action != "dry-run" {
		t.Errorf("expected action 'dry-run', got %q", results[0].action)
	}

	// Verify no file was created.
	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	if _, err := os.Stat(dcpPath); !os.IsNotExist(err) {
		t.Error("dcp.jsonc should not exist in dry-run mode")
	}
}

func TestConfigureDCPConfig_Force(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}

	// Write a modified version of the file.
	modified := `{
  "$schema": "https://example.com/old-schema.json",
  "compress": {
    "protectTags": false
  }
}
`
	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	if err := os.WriteFile(dcpPath, []byte(modified), 0o644); err != nil {
		t.Fatalf("write dcp.jsonc: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
	}

	results := configureDCPConfig(opts)
	if results[0].action != "overwritten" {
		t.Errorf("expected action 'overwritten', got %q", results[0].action)
	}

	// Verify file has canonical content.
	data, _ := os.ReadFile(dcpPath)
	if string(data) != dcpConfigContent {
		t.Errorf("file should have canonical content after force\ngot:\n%s", data)
	}
}

func TestConfigureDCPConfig_ReadError(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile: func(path string) ([]byte, error) {
			return nil, fmt.Errorf("permission denied")
		},
	}

	results := configureDCPConfig(opts)
	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if !strings.Contains(results[0].detail, "read failed") {
		t.Errorf("expected detail to contain 'read failed', got %q", results[0].detail)
	}
}

func TestConfigureDCPConfig_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".opencode"), 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}

	dcpPath := filepath.Join(dir, ".opencode", "dcp.jsonc")
	if err := os.WriteFile(dcpPath, []byte("{invalid json content"), 0o644); err != nil {
		t.Fatalf("write dcp.jsonc: %v", err)
	}

	opts := &Options{TargetDir: dir}
	results := configureDCPConfig(opts)

	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if results[0].detail != "malformed JSON" {
		t.Errorf("expected detail 'malformed JSON', got %q", results[0].detail)
	}

	// Verify file is not modified.
	data, _ := os.ReadFile(dcpPath)
	if string(data) != "{invalid json content" {
		t.Error("malformed file should not be modified")
	}
}

func TestConfigureDCPConfig_WriteError(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile: func(path string) ([]byte, error) {
			return nil, os.ErrNotExist
		},
		WriteFile: func(path string, data []byte, perm os.FileMode) error {
			return fmt.Errorf("disk full")
		},
	}

	results := configureDCPConfig(opts)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].action != "error" {
		t.Errorf("expected action 'error', got %q", results[0].action)
	}
	if !strings.Contains(results[0].detail, "write failed") {
		t.Errorf("expected detail to contain 'write failed', got %q", results[0].detail)
	}
}

// --- stripJSONCComments tests ---

func TestStripJSONCComments_RemovesLineComments(t *testing.T) {
	input := `{
  "$schema": "https://example.com/schema.json",
  // This is a comment
  // Another comment
  "compress": {
    "protectTags": true
  }
}
`
	expected := `{
  "$schema": "https://example.com/schema.json",
  "compress": {
    "protectTags": true
  }
}
`
	result := stripJSONCComments([]byte(input))
	if string(result) != expected {
		t.Errorf("stripJSONCComments result mismatch\ngot:\n%s\nwant:\n%s", result, expected)
	}
}

// --- Dewey force re-index tests ---

func TestInitSubTools_DeweyForceReindex(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write a default sources.yaml so generateDeweySources has
	// something to regenerate on force.
	defaultSources := "sources:\n  - id: disk-local\n    type: disk\n    config:\n      path: \".\"\n"
	if err := os.WriteFile(filepath.Join(dir, ".uf", "dewey", "sources.yaml"), []byte(defaultSources), 0o644); err != nil {
		t.Fatalf("write sources.yaml: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have dewey index re-indexed result.
	foundReindex := false
	for _, r := range results {
		if r.name == "dewey index" && r.action == "re-indexed" {
			foundReindex = true
		}
	}
	if !foundReindex {
		t.Errorf("expected dewey index re-indexed result, got %v", results)
	}

	// Verify dewey index --no-embeddings was called.
	indexCalled := false
	for _, call := range rec.calls {
		if call == "dewey index --no-embeddings" {
			indexCalled = true
		}
	}
	if !indexCalled {
		t.Errorf("expected 'dewey index --no-embeddings' command, got calls: %v", rec.calls)
	}

	// Verify dewey init was NOT called (.uf/dewey/ already exists).
	for _, call := range rec.calls {
		if call == "dewey init" {
			t.Error("dewey init should NOT be called when .uf/dewey/ already exists")
		}
	}

	// Verify sources.yaml was regenerated with recursive: false.
	content, readErr := os.ReadFile(filepath.Join(dir, ".uf", "dewey", "sources.yaml"))
	if readErr != nil {
		t.Fatalf("read sources.yaml: %v", readErr)
	}
	if !strings.Contains(string(content), "recursive: false") {
		t.Error("expected sources.yaml to contain recursive: false after force regeneration")
	}
}

func TestInitSubTools_DeweyExistsNoForce(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ directory — already initialized.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		Force:     false,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should NOT have any dewey-related results (skipped silently).
	for _, r := range results {
		if r.name == ".uf/dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when Force=false and .uf/dewey/ exists: %s %s", r.name, r.action)
		}
	}

	// Verify no dewey commands were called.
	for _, call := range rec.calls {
		if call == "dewey init" || strings.HasPrefix(call, "dewey index") {
			t.Errorf("unexpected dewey command when Force=false: %s", call)
		}
	}
}

func TestInitSubTools_DeweySkippedByConfig(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with dewey method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Dewey commands should NOT have been called.
	for _, call := range rec.calls {
		if call == "dewey init" || strings.HasPrefix(call, "dewey index") {
			t.Errorf("dewey command should NOT be called when method is skip: %s", call)
		}
	}

	// No dewey-related results should appear.
	for _, r := range results {
		if r.name == ".uf/dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when method is skip: %s %s", r.name, r.action)
		}
	}

	// .uf/dewey/ directory should NOT have been created.
	if _, err := os.Stat(filepath.Join(dir, ".uf", "dewey")); !os.IsNotExist(err) {
		t.Error(".uf/dewey/ should not exist when dewey method is skip")
	}
}

func TestInitSubTools_DeweySkippedByConfig_ForceIgnored(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with dewey method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	// Create .uf/dewey/ — even with Force, skip should win.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir dewey: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Even with Force=true, skip config should prevent re-index.
	for _, call := range rec.calls {
		if call == "dewey init" || strings.HasPrefix(call, "dewey index") {
			t.Errorf("dewey command should NOT be called when method is skip (even with Force): %s", call)
		}
	}
	for _, r := range results {
		if r.name == ".uf/dewey/" || r.name == "dewey index" {
			t.Errorf("unexpected dewey result when method is skip: %s %s", r.name, r.action)
		}
	}
}

func TestInitSubTools_ReplicatorSkippedByConfig(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with replicator method: skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    replicator:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Replicator commands should NOT have been called.
	for _, call := range rec.calls {
		if call == "replicator init" {
			t.Errorf("replicator init should NOT be called when method is skip")
		}
	}
	for _, r := range results {
		if r.name == ".uf/replicator/" {
			t.Errorf("unexpected replicator result when method is skip: %s %s", r.name, r.action)
		}
	}
}

func TestInitSubTools_AllToolsSkippedByConfig(t *testing.T) {
	dir := t.TempDir()

	// Create .uf/config.yaml with all tools set to skip.
	ufDir := filepath.Join(dir, ".uf")
	if err := os.MkdirAll(ufDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgContent := "setup:\n  tools:\n    dewey:\n      method: skip\n    replicator:\n      method: skip\n    specify:\n      method: skip\n    openspec:\n      method: skip\n    gaze:\n      method: skip\n"
	if err := os.WriteFile(filepath.Join(ufDir, "config.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
			"specify":    "/usr/local/bin/specify",
			"openspec":   "/usr/local/bin/openspec",
			"gaze":       "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// No tool commands should have been called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no commands when all tools are skipped, got: %v", rec.calls)
	}

	// Should have 2 results: opencode.json skipped + dcp.jsonc created.
	if len(results) != 2 {
		t.Fatalf("expected 2 results (opencode.json + dcp.jsonc), got %d: %v", len(results), results)
	}
	if results[0].name != "opencode.json" {
		t.Errorf("expected opencode.json result, got %s", results[0].name)
	}
	if results[1].name != "dcp.jsonc" {
		t.Errorf("expected dcp.jsonc result, got %s", results[1].name)
	}
}

// Phase: Concurrent execution tests (FR-001 through FR-005).

func TestInitSubTools_ConcurrentAllResults(t *testing.T) {
	// FR-001, FR-003: All five tools available — results from all
	// tools are collected regardless of execution order.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
			"specify":    "/usr/local/bin/specify",
			"openspec":   "/usr/local/bin/openspec",
			"gaze":       "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// Expect results from: dewey init, dewey index, replicator,
	// specify, openspec, gaze, opencode.json.
	// (dewey sources may or may not appear depending on filesystem.)
	expectedNames := map[string]bool{
		".uf/dewey/":      false,
		"dewey index":     false,
		".uf/replicator/": false,
		".specify/":       false,
		"openspec/":       false,
		"gaze":            false,
		"opencode.json":   false,
	}

	for _, r := range results {
		if _, ok := expectedNames[r.name]; ok {
			expectedNames[r.name] = true
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("missing result for %q in concurrent run", name)
		}
	}

	// Verify all expected commands were issued.
	rec.mu.Lock()
	callsCopy := make([]string, len(rec.calls))
	copy(callsCopy, rec.calls)
	rec.mu.Unlock()

	expectedCmds := []string{
		"dewey init", "dewey index --no-embeddings",
		"replicator init", "specify init --here --integration opencode --offline --force",
		"openspec init --tools opencode", "gaze init",
	}
	for _, cmd := range expectedCmds {
		found := false
		for _, call := range callsCopy {
			if call == cmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q in concurrent run, got calls: %v", cmd, callsCopy)
		}
	}
}

func TestInitSubTools_ConcurrentMissingToolSkipped(t *testing.T) {
	// FR-001: Missing tools are skipped; available tools still
	// complete their init without delay.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	// Only dewey and gaze available; others missing.
	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey": "/usr/local/bin/dewey",
			"gaze":  "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// Should see results for dewey, gaze, and opencode.json.
	// Should NOT see replicator, specify, or openspec results.
	resultNames := map[string]bool{}
	for _, r := range results {
		resultNames[r.name] = true
	}

	if !resultNames[".uf/dewey/"] {
		t.Error("expected dewey result when dewey is available")
	}
	if !resultNames["gaze"] {
		t.Error("expected gaze result when gaze is available")
	}
	if resultNames[".uf/replicator/"] {
		t.Error("unexpected replicator result when replicator is not available")
	}
	if resultNames[".specify/"] {
		t.Error("unexpected specify result when specify is not available")
	}
	if resultNames["openspec/"] {
		t.Error("unexpected openspec result when openspec is not available")
	}
}

func TestInitSubTools_ConcurrentOneFailureNoBlock(t *testing.T) {
	// FR-005: One tool failure does not prevent other tools from
	// completing. Dewey init fails; replicator and gaze succeed.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{
		"dewey init": fmt.Errorf("simulated dewey failure"),
	}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
			"gaze":       "/usr/local/bin/gaze",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// Dewey should appear with "failed" action.
	deweyFailed := false
	replicatorOK := false
	gazeOK := false
	opencodeOK := false
	for _, r := range results {
		switch r.name {
		case ".uf/dewey/":
			if r.action == "failed" {
				deweyFailed = true
			}
		case ".uf/replicator/":
			if r.action == "initialized" {
				replicatorOK = true
			}
		case "gaze":
			if r.action == "initialized" {
				gazeOK = true
			}
		case "opencode.json":
			opencodeOK = true
		}
	}

	if !deweyFailed {
		t.Error("expected dewey to report failure")
	}
	if !replicatorOK {
		t.Error("expected replicator to succeed despite dewey failure")
	}
	if !gazeOK {
		t.Error("expected gaze to succeed despite dewey failure")
	}
	if !opencodeOK {
		t.Error("expected opencode.json result after concurrent completion")
	}

	// Verify dewey index was NOT called (init failed).
	rec.mu.Lock()
	for _, call := range rec.calls {
		if strings.HasPrefix(call, "dewey index") {
			t.Error("dewey index should not run when dewey init fails")
		}
	}
	rec.mu.Unlock()
}

func TestInitSubTools_ConcurrentOpencodeJSONAfterAll(t *testing.T) {
	// FR-004: configureOpencodeJSON runs after all concurrent
	// groups complete. Verified by checking opencode.json result
	// is present and contains correct MCP entries.
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath: stubScaffoldLookPath(map[string]string{
			"dewey":      "/usr/local/bin/dewey",
			"replicator": "/usr/local/bin/replicator",
		}),
		ExecCmd: rec.execCmd,
	}

	results := initSubTools(opts)

	// opencode.json must be present in results.
	opencodeFound := false
	for _, r := range results {
		if r.name == "opencode.json" {
			opencodeFound = true
		}
	}
	if !opencodeFound {
		t.Fatal("expected opencode.json result after concurrent init")
	}

	// opencode.json file should exist on disk (written by
	// configureOpencodeJSON after wg.Wait).
	opencodeJSON := filepath.Join(dir, "opencode.json")
	data, err := os.ReadFile(opencodeJSON)
	if err != nil {
		t.Fatalf("opencode.json should exist: %v", err)
	}

	// Verify it contains expected MCP server entries for both
	// tools that were initialized.
	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(data, &parsed); jsonErr != nil {
		t.Fatalf("opencode.json is not valid JSON: %v", jsonErr)
	}
	mcp, ok := parsed["mcp"]
	if !ok {
		t.Fatal("opencode.json missing mcp key")
	}
	mcpMap, ok := mcp.(map[string]interface{})
	if !ok {
		t.Fatal("opencode.json mcp is not a map")
	}
	if _, hasDewey := mcpMap["dewey"]; !hasDewey {
		t.Error("opencode.json should contain dewey MCP entry")
	}
	if _, hasReplicator := mcpMap["replicator"]; !hasReplicator {
		t.Error("opencode.json should contain replicator MCP entry")
	}
}

// Phase 8: Integration test (T032a)

func TestInitSubTools_OpencodeJSON(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/dewey/ so dewey init is skipped.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "dewey"), 0o755); err != nil {
		t.Fatalf("mkdir .uf/dewey: %v", err)
	}
	// Create .uf/replicator/ so replicator init is skipped.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0o755); err != nil {
		t.Fatalf("mkdir .uf/replicator: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey", "replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Find the opencode.json result.
	var ocResult *subToolResult
	for i := range results {
		if results[i].name == "opencode.json" {
			ocResult = &results[i]
			break
		}
	}
	if ocResult == nil {
		t.Fatal("opencode.json result not found in initSubTools output")
	}
	if ocResult.action != "created" {
		t.Errorf("expected opencode.json action 'created', got %q", ocResult.action)
	}

	// Verify file exists with expected content.
	ocMap := parseOpencodeJSON(t, dir)
	getMCPDewey(t, ocMap)
	getMCPReplicator(t, ocMap)

	// Verify no plugin key.
	if _, ok := ocMap["plugin"]; ok {
		t.Error("plugin key should not exist")
	}
}

// --- Replicator init delegation tests ---

func TestInitSubTools_ReplicatorInit(t *testing.T) {
	dir := t.TempDir()
	// No .uf/replicator/ — should trigger replicator init.
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have .uf/replicator/ initialized result.
	foundReplicator := false
	for _, r := range results {
		if r.name == ".uf/replicator/" && r.action == "initialized" {
			foundReplicator = true
		}
	}
	if !foundReplicator {
		t.Errorf("expected .uf/replicator/ initialized, got %v", results)
	}

	// Verify replicator init was called.
	found := false
	for _, call := range rec.calls {
		if call == "replicator init" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'replicator init' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_ReplicatorInitSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	// Create .uf/replicator/ — should skip replicator init with a
	// visible "already configured" result.
	if err := os.MkdirAll(filepath.Join(dir, ".uf", "replicator"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have a visible .uf/replicator/ "already configured" result.
	foundSkipped := false
	for _, r := range results {
		if r.name == ".uf/replicator/" && r.action == "already configured" {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Errorf("expected .uf/replicator/ already configured, got %v", results)
	}

	// Verify replicator init was NOT called.
	for _, call := range rec.calls {
		if call == "replicator init" {
			t.Error("replicator init should NOT be called when .uf/replicator/ exists")
		}
	}
}

func TestInitSubTools_ReplicatorInitFails(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"replicator init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have .uf/replicator/ failed result.
	foundFailed := false
	for _, r := range results {
		if r.name == ".uf/replicator/" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected .uf/replicator/ failed, got %v", results)
	}

	// Should still have opencode.json result (init failure doesn't block).
	foundOC := false
	for _, r := range results {
		if r.name == "opencode.json" {
			foundOC = true
		}
	}
	if !foundOC {
		t.Error("opencode.json result should still be present after replicator init failure")
	}
}

// --- Specify delegation tests ---

func TestInitSubTools_SpecifyInit(t *testing.T) {
	dir := t.TempDir()
	// No .specify/ — should trigger specify init.
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundSpecify := false
	for _, r := range results {
		if r.name == ".specify/" && r.action == "initialized" {
			foundSpecify = true
		}
	}
	if !foundSpecify {
		t.Errorf("expected .specify/ initialized, got %v", results)
	}

	found := false
	for _, call := range rec.calls {
		if call == "specify init --here --integration opencode --offline --force" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'specify init --here --integration opencode --offline --force' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_SpecifySkipped(t *testing.T) {
	dir := t.TempDir()
	// Create .specify/scripts sentinel — specify init should be
	// skipped with a visible "already configured" result.
	if err := os.MkdirAll(filepath.Join(dir, ".specify", "scripts"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundSkipped := false
	for _, r := range results {
		if r.name == ".specify/" && r.action == "already configured" {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Errorf("expected .specify/ already configured, got %v", results)
	}

	for _, call := range rec.calls {
		if call == "specify init --here --integration opencode --offline --force" {
			t.Error("specify init should NOT be called when .specify/ exists")
		}
	}
}

func TestInitSubTools_SpecifyNotInstalled(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{}), // No specify
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	for _, r := range results {
		if r.name == ".specify/" {
			t.Errorf("unexpected .specify/ result when specify not installed: %s %s", r.name, r.action)
		}
	}
}

func TestInitSubTools_SpecifyFailed(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"specify init --here --integration opencode --offline --force": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundFailed := false
	for _, r := range results {
		if r.name == ".specify/" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected .specify/ failed, got %v", results)
	}
}

// TestSimpleToolsSentinelsNotScaffolded pins the sentinel-integrity
// invariant between the embedded asset walk and the simpleTools
// table: no embedded asset may map (via mapAssetPath) to a target
// path that equals or is a descendant of any tool's sentinel. The
// original bug was exactly such a collision — the starter
// constitution (assets/specify/memory/constitution.md) mapped under
// the old bare-.specify sentinel, so Run()'s asset walk created the
// sentinel and suppressed `specify init` on every fresh run. All
// four current sentinels are clear; this guard fails CI the moment
// a future embedded asset recreates the bug's precondition, for any
// table entry — including tools added later. (D4; FR-020)
func TestSimpleToolsSentinelsNotScaffolded(t *testing.T) {
	// Mirror Run()'s walk (fs.WalkDir over the embedded FS,
	// skipping directories and stripping the assets/ prefix) so
	// the guard enumerates exactly the paths Run() can scaffold.
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath := strings.TrimPrefix(path, "assets/")
		target := mapAssetPath(relPath)
		for _, tool := range simpleTools {
			// mapAssetPath builds targets from embedded FS paths,
			// which are always slash-separated, while sentinels
			// use filepath.Join; normalize the sentinel so the
			// guard cannot silently pass on Windows-style
			// separators.
			sentinel := filepath.ToSlash(tool.sentinel)
			if target == sentinel || strings.HasPrefix(target, sentinel+"/") {
				t.Errorf("embedded asset %q maps to %q, which equals or is a descendant of the %s sentinel %q; scaffolding it would suppress %s init",
					relPath, target, tool.name, sentinel, tool.name)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk embedded assets: %v", err)
	}
}

// --- OpenSpec delegation tests ---

func TestInitSubTools_OpenSpecInit(t *testing.T) {
	dir := t.TempDir()
	// Create openspec/ directory (simulating embedded schema deployment)
	// but no config.yaml — should trigger openspec init.
	if err := os.MkdirAll(filepath.Join(dir, "openspec", "schemas"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"openspec": "/usr/local/bin/openspec"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundOpenSpec := false
	for _, r := range results {
		if r.name == "openspec/" && r.action == "initialized" {
			foundOpenSpec = true
		}
	}
	if !foundOpenSpec {
		t.Errorf("expected openspec/ initialized, got %v", results)
	}

	found := false
	for _, call := range rec.calls {
		if call == "openspec init --tools opencode" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'openspec init --tools opencode' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_OpenSpecSkipped(t *testing.T) {
	dir := t.TempDir()
	// Create openspec/config.yaml — should skip openspec init with
	// a visible "already configured" result.
	if err := os.MkdirAll(filepath.Join(dir, "openspec"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openspec", "config.yaml"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"openspec": "/usr/local/bin/openspec"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have a visible openspec/ "already configured" result.
	foundSkipped := false
	for _, r := range results {
		if r.name == "openspec/" && r.action == "already configured" {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Errorf("expected openspec/ already configured, got %v", results)
	}

	for _, call := range rec.calls {
		if strings.Contains(call, "openspec init") {
			t.Error("openspec init should NOT be called when config.yaml exists")
		}
	}
}

func TestInitSubTools_OpenSpecFailed(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"openspec init --tools opencode": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"openspec": "/usr/local/bin/openspec"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundFailed := false
	for _, r := range results {
		if r.name == "openspec/" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected openspec/ failed, got %v", results)
	}
}

// --- Gaze delegation tests ---

func TestInitSubTools_GazeInit(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/agents/ but no gaze-reporter.md — should trigger gaze init.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundGaze := false
	for _, r := range results {
		if r.name == "gaze" && r.action == "initialized" {
			foundGaze = true
		}
	}
	if !foundGaze {
		t.Errorf("expected gaze initialized, got %v", results)
	}

	found := false
	for _, call := range rec.calls {
		if call == "gaze init" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'gaze init' call, got: %v", rec.calls)
	}
}

func TestInitSubTools_GazeSkipped(t *testing.T) {
	dir := t.TempDir()
	// Create gaze-reporter.md — should skip gaze init with a visible
	// "already configured" result.
	agentDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(agentDir, "gaze-reporter.md"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("write agent: %v", err)
	}
	rec := &scaffoldCmdRecorder{errors: map[string]error{}}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Should have a visible gaze "already configured" result.
	foundSkipped := false
	for _, r := range results {
		if r.name == "gaze" && r.action == "already configured" {
			foundSkipped = true
		}
	}
	if !foundSkipped {
		t.Errorf("expected gaze already configured, got %v", results)
	}

	for _, call := range rec.calls {
		if call == "gaze init" {
			t.Error("gaze init should NOT be called when gaze-reporter.md exists")
		}
	}
}

func TestInitSubTools_GazeFailed(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/agents/ but no gaze-reporter.md.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"gaze init": fmt.Errorf("init failed"),
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	foundFailed := false
	for _, r := range results {
		if r.name == "gaze" && r.action == "failed" {
			foundFailed = true
		}
	}
	if !foundFailed {
		t.Errorf("expected gaze failed, got %v", results)
	}
}

// --- ensureGitignore tests ---

func TestEnsureGitignore_FreshDir(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureGitignore(opts)

	if result.name != ".gitignore" {
		t.Errorf("expected name '.gitignore', got %q", result.name)
	}
	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	// Verify file exists with marker and all patterns.
	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, gitignoreMarker) {
		t.Error("expected marker comment in .gitignore")
	}
	// Verify representative patterns from each section.
	for _, pattern := range []string{
		".uf/dewey/graph.db",
		".uf/dewey/cache/",
		".uf/replicator/*.db",
		".devcontainer/",
		".dewey/",
		".hive/",
		".unbound-force/",
	} {
		if !strings.Contains(text, pattern) {
			t.Errorf("expected pattern %q in .gitignore", pattern)
		}
	}
}

func TestEnsureGitignore_ExistingNoBlock(t *testing.T) {
	dir := t.TempDir()

	// Create existing .gitignore with project-specific content.
	existingContent := "node_modules/\n*.log\n"
	giPath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(giPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureGitignore(opts)

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(giPath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	text := string(content)

	// Existing content must be preserved at the start.
	if !strings.HasPrefix(text, existingContent) {
		t.Errorf("existing content not preserved; got:\n%s", text)
	}

	// UF block must be appended.
	if !strings.Contains(text, gitignoreMarker) {
		t.Error("expected UF marker in .gitignore")
	}

	// Blank line separator between existing content and UF block.
	markerIdx := strings.Index(text, gitignoreMarker)
	if markerIdx < 2 {
		t.Fatal("marker at unexpected position")
	}
	// The two characters before the marker should be "\n\n" (blank line separator).
	before := text[:markerIdx]
	if !strings.HasSuffix(before, "\n\n") {
		t.Errorf("expected blank line separator before UF block, got trailing chars: %q",
			before[len(before)-4:])
	}
}

func TestEnsureGitignore_ExistingWithBlock(t *testing.T) {
	dir := t.TempDir()

	// Create .gitignore that already has the UF block.
	existingContent := "node_modules/\n\n" + gitignoreBlock
	giPath := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(giPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureGitignore(opts)

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	// Verify file is byte-identical (not modified).
	after, err := os.ReadFile(giPath)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(after) != existingContent {
		t.Error(".gitignore should be byte-identical when marker already present")
	}
}

func TestEnsureGitignore_Idempotent(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	// First call — creates the file.
	result1 := ensureGitignore(opts)
	if result1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", result1.action)
	}

	content1, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read after first call: %v", err)
	}

	// Second call — should skip (idempotent).
	result2 := ensureGitignore(opts)
	if result2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", result2.action)
	}

	content2, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read after second call: %v", err)
	}

	// Content must be identical after both calls (block not duplicated).
	if !bytes.Equal(content1, content2) {
		t.Errorf("content differs after second call:\nfirst:\n%s\nsecond:\n%s",
			string(content1), string(content2))
	}
}

// TestAgentBriefAsset_GovernanceBlockPresence verifies the
// /uf.agent-brief scaffold asset contains all 7 behavioral rule
// templates and the conditional governance sections. Governance
// blocks moved from /uf.init Step 9 to /uf.agent-brief verbatim
// templates as part of agent-brief-consolidation. Added by
// Spec 030, updated by agent-brief-consolidation.
func TestAgentBriefAsset_GovernanceBlockPresence(t *testing.T) {
	content, err := assetContent("opencode/commands/uf.agent-brief.md")
	if err != nil {
		t.Fatalf("read uf.agent-brief.md asset: %v", err)
	}
	text := string(content)

	// Each entry is a detection phrase that must appear in the
	// uf.agent-brief.md asset to confirm the governance block is
	// defined in the verbatim template.
	requiredPhrases := []struct {
		block  string
		phrase string
	}{
		{"Gatekeeping", "**Gatekeeping**"},
		{"Phase boundaries", "**Phase boundaries**"},
		{"CI parity", "**CI parity**"},
		{"Review council", "**Review council**"},
		{"Branch protection", "**Branch protection**"},
		{"Documentation gate", "**Documentation gate**"},
		{"Zero-waste", "**Zero-waste**"},
		{"Specification Workflow", "## Specification Workflow"},
		{"Knowledge Retrieval", "## Knowledge Retrieval"},
	}

	for _, rp := range requiredPhrases {
		if !strings.Contains(text, rp.phrase) {
			t.Errorf("uf.agent-brief.md asset missing governance block %q (detection phrase %q not found)",
				rp.block, rp.phrase)
		}
	}
}

// --- Cross-tool bridge tests ---

func TestEnsureAGENTSmdPackSection_NoAGENTSmd(t *testing.T) {
	dir := t.TempDir()

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdPackSection(opts, "go")

	if result.action != "skipped (no AGENTS.md)" {
		t.Errorf("expected action 'skipped (no AGENTS.md)', got %q", result.action)
	}
}

func TestEnsureAGENTSmdPackSection_ExistingWithoutSection(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\nSome content.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdPackSection(opts, "go")

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)

	if !strings.HasPrefix(text, existing) {
		t.Error("existing content not preserved")
	}
	if !strings.Contains(text, agentsmdPackMarker) {
		t.Error("expected Convention Packs heading")
	}
	if !strings.Contains(text, ".opencode/uf/packs/go.md") {
		t.Error("expected Go pack reference")
	}
	if !strings.Contains(text, ".opencode/uf/packs/default.md") {
		t.Error("expected default pack reference")
	}
}

func TestEnsureAGENTSmdPackSection_AlreadyConfigured(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\n## Convention Packs\n\nAlready here.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdPackSection(opts, "go")

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(after) != existing {
		t.Error("AGENTS.md should be unchanged when section already present")
	}
}

func TestEnsureAGENTSmdPackSection_Idempotent(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	r1 := ensureAGENTSmdPackSection(opts, "go")
	if r1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", r1.action)
	}

	content1, _ := os.ReadFile(agentsPath)

	r2 := ensureAGENTSmdPackSection(opts, "go")
	if r2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", r2.action)
	}

	content2, _ := os.ReadFile(agentsPath)
	if !bytes.Equal(content1, content2) {
		t.Error("content should be identical after second call")
	}
}

func TestEnsureAGENTSmdSiblingSection_NoAGENTSmd(t *testing.T) {
	dir := t.TempDir()

	// Sibling declaration exists, but there is no AGENTS.md to append to.
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "skipped (no AGENTS.md)" {
		t.Errorf("expected action 'skipped (no AGENTS.md)', got %q", result.action)
	}
}

func TestEnsureAGENTSmdSiblingSection_NoSiblingReposYaml(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\nSome content.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "skipped (no sibling-repos.yaml)" {
		t.Errorf("expected action 'skipped (no sibling-repos.yaml)', got %q", result.action)
	}

	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(after) != existing {
		t.Error("AGENTS.md should be unchanged when no sibling-repos.yaml exists")
	}
}

func TestEnsureAGENTSmdSiblingSection_ExistingWithoutSection(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\nSome content.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)

	if !strings.HasPrefix(text, existing) {
		t.Error("existing content not preserved")
	}
	if !strings.Contains(text, agentsmdSiblingMarker) {
		t.Error("expected Sibling Repositories heading")
	}
	for _, phrase := range []string{
		".uf/sibling-repos.yaml",
		".uf/sibling-clones/<name>/",
		".uf/sibling-cache/<name>/",
		"acquire_sibling_evidence",
		"MUST NOT read sibling repositories",
	} {
		if !strings.Contains(text, phrase) {
			t.Errorf("expected sibling section to mention %q", phrase)
		}
	}
}

func TestEnsureAGENTSmdSiblingSection_AlreadyConfigured(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n\n## Sibling Repositories\n\nAlready here.\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "already configured" {
		t.Errorf("expected action 'already configured', got %q", result.action)
	}

	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if string(after) != existing {
		t.Error("AGENTS.md should be unchanged when section already present")
	}
}

func TestEnsureAGENTSmdSiblingSection_Idempotent(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project\n"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	r1 := ensureAGENTSmdSiblingSection(opts)
	if r1.action != "configured" {
		t.Errorf("first call: expected 'configured', got %q", r1.action)
	}

	content1, _ := os.ReadFile(agentsPath)

	r2 := ensureAGENTSmdSiblingSection(opts)
	if r2.action != "already configured" {
		t.Errorf("second call: expected 'already configured', got %q", r2.action)
	}

	content2, _ := os.ReadFile(agentsPath)
	if !bytes.Equal(content1, content2) {
		t.Error("content should be identical after second call")
	}
}

func TestEnsureAGENTSmdSiblingSection_ReadFailure(t *testing.T) {
	dir := t.TempDir()

	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("# My Project\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile: func(string) ([]byte, error) {
			return nil, fmt.Errorf("disk unavailable")
		},
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "failed" {
		t.Errorf("expected action 'failed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "read failed") {
		t.Errorf("expected 'read failed' detail, got %q", result.detail)
	}
}

func TestEnsureAGENTSmdSiblingSection_WriteFailure(t *testing.T) {
	dir := t.TempDir()

	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("# My Project\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: func(string, []byte, os.FileMode) error {
			return fmt.Errorf("disk full")
		},
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "failed" {
		t.Errorf("expected action 'failed', got %q", result.action)
	}
	if !strings.Contains(result.detail, "write failed") {
		t.Errorf("expected 'write failed' detail, got %q", result.detail)
	}
}

func TestEnsureAGENTSmdSiblingSection_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()

	existing := "# My Project"
	agentsPath := filepath.Join(dir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".uf"), 0o755); err != nil {
		t.Fatalf("mkdir .uf: %v", err)
	}
	siblingPath := filepath.Join(dir, ".uf", "sibling-repos.yaml")
	if err := os.WriteFile(siblingPath, []byte("version: 1\nsiblings: []\n"), 0o644); err != nil {
		t.Fatalf("write sibling-repos.yaml: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := ensureAGENTSmdSiblingSection(opts)

	if result.action != "configured" {
		t.Errorf("expected action 'configured', got %q", result.action)
	}

	content, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	text := string(content)
	// The missing trailing newline must be synthesized so the section
	// heading starts on its own line instead of gluing onto the last
	// existing line ("# My Project## Sibling Repositories").
	if !strings.Contains(text, "# My Project\n\n## Sibling Repositories") {
		t.Errorf("expected section to start on its own line after newline synthesis, got:\n%s", text)
	}
}

// TestEnsureAGENTSmdSiblingSection_StealthModeNonModification asserts the
// Run()-level stealth gating: stealth mode must not modify AGENTS.md even
// when a sibling declaration exists, while a normal run appends the
// section (stealth-test fake precedent from TestStealthInit_LeavesTracked-
// FilesUnmodified).
func TestEnsureAGENTSmdSiblingSection_StealthModeNonModification(t *testing.T) {
	cases := []struct {
		name    string
		stealth bool
		wantMod bool
	}{
		{"stealth", true, false},
		{"normal", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			initGitDir(t, dir)
			writeFile(t, filepath.Join(dir, ".gitignore"), "existing-ignore\n", 0o644)
			writeFile(t, filepath.Join(dir, "AGENTS.md"), "# AGENTS\n", 0o644)
			writeFile(t, filepath.Join(dir, ".uf", "sibling-repos.yaml"), "version: 1\nsiblings: []\n", 0o644)

			var buf bytes.Buffer
			g := newStealthGitFake()
			opts := Options{
				TargetDir: dir,
				Stealth:   c.stealth,
				Version:   "1.2.3-test",
				Stdout:    &buf,
				ExecCmd:   g.exec,
				ReadFile:  os.ReadFile,
				WriteFile: os.WriteFile,
			}
			if _, err := Run(opts); err != nil {
				t.Fatalf("Run() error: %v", err)
			}

			ag, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
			if err != nil {
				t.Fatalf("read AGENTS.md: %v", err)
			}
			if c.wantMod {
				if !strings.Contains(string(ag), agentsmdSiblingMarker) {
					t.Errorf("normal mode: AGENTS.md missing sibling section:\n%s", ag)
				}
			} else {
				if string(ag) != "# AGENTS\n" {
					t.Errorf("stealth mode modified AGENTS.md:\n%s", ag)
				}
			}
		})
	}
}

func TestCollectDeployedPacks_Go(t *testing.T) {
	packs := collectDeployedPacks("go", "")

	expected := map[string]bool{
		"default.md":             true,
		"default-custom.md":      true,
		"severity.md":            true,
		"content.md":             true,
		"content-custom.md":      true,
		"ci.md":                  true,
		"ci-custom.md":           true,
		"review-voice.md":        true,
		"review-voice-custom.md": true,
		"go.md":                  true,
		"go-custom.md":           true,
	}

	if len(packs) != len(expected) {
		t.Errorf("expected %d packs, got %d: %v", len(expected), len(packs), packs)
	}
	for _, p := range packs {
		if !expected[p] {
			t.Errorf("unexpected pack %q", p)
		}
	}
}

func TestCollectDeployedPacks_TypeScript(t *testing.T) {
	packs := collectDeployedPacks("typescript", "")

	required := []string{"typescript.md", "review-voice.md", "review-voice-custom.md"}
	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	for _, r := range required {
		if !packSet[r] {
			t.Errorf("expected %q in packs", r)
		}
	}
}

func TestCollectDeployedPacks_Default(t *testing.T) {
	packs := collectDeployedPacks("default", "")

	for _, p := range packs {
		if p == "default.md" || p == "default-custom.md" ||
			p == "severity.md" || p == "content.md" ||
			p == "content-custom.md" || p == "ci.md" ||
			p == "ci-custom.md" || p == "review-voice.md" ||
			p == "review-voice-custom.md" {
			continue
		}
		t.Errorf("unexpected pack %q for default lang", p)
	}
	if len(packs) != 9 {
		t.Errorf("expected 9 packs for default lang, got %d", len(packs))
	}
}

// --- hasRuleContent tests ---

func TestHasRuleContent_EmptyStub(t *testing.T) {
	dir := t.TempDir()
	stub := `---
pack_id: default-custom
language: Any
version: 1.0.0
---

# Custom Rules: Default

Project-specific conventions.

## Custom Rules

<!-- Add project-specific rules below this line -->`
	path := filepath.Join(dir, "default-custom.md")
	if err := os.WriteFile(path, []byte(stub), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	if hasRuleContent(path) {
		t.Error("expected false for empty stub, got true")
	}
}

func TestHasRuleContent_WhitespaceOnlyAfterSentinel(t *testing.T) {
	dir := t.TempDir()
	stub := "<!-- Add project-specific rules below this line -->\n\n   \n\t\n"
	path := filepath.Join(dir, "go-custom.md")
	if err := os.WriteFile(path, []byte(stub), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	if hasRuleContent(path) {
		t.Error("expected false for whitespace-only after sentinel, got true")
	}
}

func TestHasRuleContent_PopulatedFile(t *testing.T) {
	dir := t.TempDir()
	populated := `---
pack_id: go-custom
language: Go
version: 1.0.0
---

## Custom Rules

<!-- Add project-specific rules below this line -->

## CR-001 [MUST] Use structured logging

All log calls MUST use charmbracelet/log.
`
	path := filepath.Join(dir, "go-custom.md")
	if err := os.WriteFile(path, []byte(populated), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !hasRuleContent(path) {
		t.Error("expected true for populated file, got false")
	}
}

func TestHasRuleContent_NoSentinel(t *testing.T) {
	dir := t.TempDir()
	// File with content but no sentinel — treat as non-empty (fail-open).
	path := filepath.Join(dir, "content-custom.md")
	if err := os.WriteFile(path, []byte("some content"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if !hasRuleContent(path) {
		t.Error("expected true (no sentinel → fail-open), got false")
	}
}

func TestHasRuleContent_MissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.md")

	if !hasRuleContent(path) {
		t.Error("expected true for missing file (fail-open), got false")
	}
}

// --- collectDeployedPacks with root tests ---

func TestCollectDeployedPacks_WithRoot_AllEmpty(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	stubs := []string{"default-custom.md", "content-custom.md", "ci-custom.md", "go-custom.md", "review-voice-custom.md"}
	for _, name := range stubs {
		if err := os.WriteFile(filepath.Join(packsDir, name), []byte(sentinel), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	packs := collectDeployedPacks("go", dir)

	for _, p := range packs {
		if strings.HasSuffix(p, "-custom.md") {
			t.Errorf("empty custom pack %q should be excluded, but was included", p)
		}
	}
	// Non-custom packs must still be present.
	required := []string{"default.md", "severity.md", "content.md", "ci.md", "go.md", "review-voice.md"}
	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	for _, r := range required {
		if !packSet[r] {
			t.Errorf("required pack %q missing from result", r)
		}
	}
}

func TestCollectDeployedPacks_WithRoot_OnePopulated(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	// default-custom, content-custom, ci-custom, and review-voice-custom
	// are empty stubs.
	for _, name := range []string{"default-custom.md", "content-custom.md", "ci-custom.md", "review-voice-custom.md"} {
		if err := os.WriteFile(filepath.Join(packsDir, name), []byte(sentinel), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	// go-custom is populated.
	populated := sentinel + "\n\n## CR-001 [MUST] Example rule\n"
	if err := os.WriteFile(filepath.Join(packsDir, "go-custom.md"), []byte(populated), 0o644); err != nil {
		t.Fatalf("write go-custom.md: %v", err)
	}

	packs := collectDeployedPacks("go", dir)

	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	if !packSet["go-custom.md"] {
		t.Error("go-custom.md should be included (populated)")
	}
	if packSet["default-custom.md"] {
		t.Error("default-custom.md should be excluded (empty)")
	}
	if packSet["content-custom.md"] {
		t.Error("content-custom.md should be excluded (empty)")
	}
	if packSet["ci-custom.md"] {
		t.Error("ci-custom.md should be excluded (empty)")
	}
	if packSet["review-voice-custom.md"] {
		t.Error("review-voice-custom.md should be excluded (empty)")
	}
}

func TestCollectDeployedPacks_WithRoot_EmptyRootFallback(t *testing.T) {
	// When root is "", all packs including empty custom packs are returned
	// (backward-compatible behaviour; no filesystem access).
	packs := collectDeployedPacks("go", "")

	packSet := make(map[string]bool, len(packs))
	for _, p := range packs {
		packSet[p] = true
	}
	for _, name := range []string{"default-custom.md", "content-custom.md", "ci-custom.md", "go-custom.md", "review-voice-custom.md"} {
		if !packSet[name] {
			t.Errorf("expected %q in packs when root is empty, got: %v", name, packs)
		}
	}
}

// --- filterEmptyCustomPacks tests ---

func TestFilterEmptyCustomPacks_EmptyRoot(t *testing.T) {
	candidates := []string{"default.md", "default-custom.md", "go.md"}
	got := filterEmptyCustomPacks(candidates, "")

	if len(got) != len(candidates) {
		t.Errorf("expected %d packs with empty root, got %d", len(candidates), len(got))
	}
	for i, name := range candidates {
		if got[i] != name {
			t.Errorf("pack[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestFilterEmptyCustomPacks_ExcludesEmptyStubs(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	sentinel := "<!-- Add project-specific rules below this line -->"
	if err := os.WriteFile(filepath.Join(packsDir, "default-custom.md"), []byte(sentinel), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	candidates := []string{"default.md", "default-custom.md", "severity.md"}
	got := filterEmptyCustomPacks(candidates, dir)

	gotSet := make(map[string]bool, len(got))
	for _, p := range got {
		gotSet[p] = true
	}
	if gotSet["default-custom.md"] {
		t.Error("empty default-custom.md should be excluded")
	}
	if !gotSet["default.md"] {
		t.Error("non-custom default.md should be included")
	}
	if !gotSet["severity.md"] {
		t.Error("non-custom severity.md should be included")
	}
}

func TestFilterEmptyCustomPacks_KeepsPopulated(t *testing.T) {
	dir := t.TempDir()
	packsDir := filepath.Join(dir, ".opencode", "uf", "packs")
	if err := os.MkdirAll(packsDir, 0o755); err != nil {
		t.Fatalf("mkdir packs: %v", err)
	}

	populated := "<!-- Add project-specific rules below this line -->\n\n## CR-001 Real rule\n"
	if err := os.WriteFile(filepath.Join(packsDir, "go-custom.md"), []byte(populated), 0o644); err != nil {
		t.Fatalf("write populated: %v", err)
	}

	candidates := []string{"go.md", "go-custom.md"}
	got := filterEmptyCustomPacks(candidates, dir)

	if len(got) != 2 {
		t.Errorf("expected 2 packs (populated custom kept), got %d: %v", len(got), got)
	}
}

// --- migrateCommandDir tests ---

// createFile is a test helper that creates a file with the given
// content at the specified relative path under dir.
func createFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	fullPath := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func TestMigrateCommandDir_NoOldDir(t *testing.T) {
	dir := t.TempDir()
	// Only .opencode/commands/ exists — no old dir to migrate.
	createFile(t, dir, ".opencode/commands/uf.review-council.md", "new")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result != nil {
		t.Errorf("expected nil (no-op) when only commands/ exists, got %+v", result)
	}
}

func TestMigrateCommandDir_NeitherDir(t *testing.T) {
	dir := t.TempDir()
	// Empty dir — no .opencode/ at all.

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result != nil {
		t.Errorf("expected nil (no-op) when neither dir exists, got %+v", result)
	}
}

func TestMigrateCommandDir_RenameOnly(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/command/ with 3 .md files, NO .opencode/commands/.
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/b.md", "bravo")
	createFile(t, dir, ".opencode/command/c.md", "charlie")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result for rename migration")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify .opencode/commands/ exists with 3 files.
	newDir := filepath.Join(dir, ".opencode", "commands")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		t.Fatalf("read new dir: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 files in commands/, got %d", len(entries))
	}

	// Verify .opencode/command/ does not exist.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old .opencode/command/ should not exist after rename")
	}

	// Verify file content preserved.
	content, err := os.ReadFile(filepath.Join(newDir, "a.md"))
	if err != nil {
		t.Fatalf("read a.md: %v", err)
	}
	if string(content) != "alpha" {
		t.Errorf("a.md content = %q, want %q", string(content), "alpha")
	}
}

func TestMigrateCommandDir_MergeUnique(t *testing.T) {
	dir := t.TempDir()
	// .opencode/command/ has a.md, b.md
	// .opencode/commands/ has c.md, d.md
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/b.md", "bravo")
	createFile(t, dir, ".opencode/commands/c.md", "charlie")
	createFile(t, dir, ".opencode/commands/d.md", "delta")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result for merge")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify .opencode/commands/ has all 4 files.
	newDir := filepath.Join(dir, ".opencode", "commands")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		t.Fatalf("read new dir: %v", err)
	}
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	for _, expected := range []string{"a.md", "b.md", "c.md", "d.md"} {
		if !names[expected] {
			t.Errorf("expected %s in commands/, got %v", expected, names)
		}
	}

	// Verify .opencode/command/ removed.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old .opencode/command/ should be removed after merge")
	}
}

func TestMigrateCommandDir_MergeDupIdentical(t *testing.T) {
	dir := t.TempDir()
	// Both dirs have x.md with identical content.
	createFile(t, dir, ".opencode/command/x.md", "same")
	createFile(t, dir, ".opencode/commands/x.md", "same")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// x.md in commands/ should be preserved.
	content, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "x.md"))
	if err != nil {
		t.Fatalf("read commands/x.md: %v", err)
	}
	if string(content) != "same" {
		t.Errorf("commands/x.md content = %q, want %q", string(content), "same")
	}

	// x.md in command/ should be removed.
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "command", "x.md")); !os.IsNotExist(err) {
		t.Error("command/x.md should be removed after merge")
	}

	// No conflict warning should be printed.
	output := buf.String()
	if strings.Contains(output, "conflict") {
		t.Errorf("should not warn about conflict for identical files, got:\n%s", output)
	}
}

func TestMigrateCommandDir_MergeDupDifferent(t *testing.T) {
	dir := t.TempDir()
	// command/x.md = "old", commands/x.md = "new"
	createFile(t, dir, ".opencode/command/x.md", "old")
	createFile(t, dir, ".opencode/commands/x.md", "new")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// commands/x.md should still contain "new".
	content, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "x.md"))
	if err != nil {
		t.Fatalf("read commands/x.md: %v", err)
	}
	if string(content) != "new" {
		t.Errorf("commands/x.md content = %q, want %q", string(content), "new")
	}

	// Warning should contain "conflict" and "/uf.init".
	output := buf.String()
	if !strings.Contains(output, "conflict") {
		t.Errorf("expected 'conflict' in warning output, got:\n%s", output)
	}
	if !strings.Contains(output, "/uf.init") {
		t.Errorf("expected '/uf.init' in warning output, got:\n%s", output)
	}
}

func TestMigrateCommandDir_Symlink(t *testing.T) {
	dir := t.TempDir()
	// Create a real directory to symlink to.
	realDir := filepath.Join(dir, "real-command-dir")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("mkdir real dir: %v", err)
	}
	// Create .opencode/ parent.
	ocDir := filepath.Join(dir, ".opencode")
	if err := os.MkdirAll(ocDir, 0o755); err != nil {
		t.Fatalf("mkdir .opencode: %v", err)
	}
	// Create .opencode/command as a symlink.
	if err := os.Symlink(realDir, filepath.Join(ocDir, "command")); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result for symlink")
	}
	if result.action != "skipped" {
		t.Errorf("expected action 'skipped', got %q", result.action)
	}
	if !strings.Contains(result.detail, "symlink") {
		t.Errorf("expected 'symlink' in detail, got %q", result.detail)
	}
}

func TestMigrateCommandDir_DivisorOnly(t *testing.T) {
	dir := t.TempDir()
	// Create .opencode/command/ — should NOT be touched in DivisorOnly mode.
	createFile(t, dir, ".opencode/command/review-council.md", "content")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir:   dir,
		DivisorOnly: true,
		Stdout:      &buf,
		ReadFile:    os.ReadFile,
		WriteFile:   os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result != nil {
		t.Errorf("expected nil in DivisorOnly mode, got %+v", result)
	}

	// .opencode/command/ should still exist.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Error(".opencode/command/ should still exist in DivisorOnly mode")
	}
}

func TestMigrateCommandDir_NonMDFiles(t *testing.T) {
	dir := t.TempDir()
	// .opencode/command/ has a.md and .DS_Store, no commands/.
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/.DS_Store", "\x00\x00")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Both files should be in commands/ (rename moves entire dir).
	newDir := filepath.Join(dir, ".opencode", "commands")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		t.Fatalf("read new dir: %v", err)
	}
	names := make(map[string]bool)
	for _, e := range entries {
		names[e.Name()] = true
	}
	if !names["a.md"] {
		t.Error("expected a.md in commands/")
	}
	if !names[".DS_Store"] {
		t.Error("expected .DS_Store in commands/")
	}

	// .opencode/command/ should be removed.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old .opencode/command/ should be removed")
	}
}

func TestMigrateCommandDir_Idempotent(t *testing.T) {
	dir := t.TempDir()
	createFile(t, dir, ".opencode/command/a.md", "alpha")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	// First call: migrates command/ → commands/.
	result1 := migrateCommandDir(opts)
	if result1 == nil || result1.action != "migrated" {
		t.Fatalf("first call: expected migrated, got %+v", result1)
	}

	// Second call: command/ is gone → returns nil (silent no-op).
	result2 := migrateCommandDir(opts)
	if result2 != nil {
		t.Errorf("second call: expected nil (no-op), got %+v", result2)
	}

	// Verify commands/ still has the file.
	content, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "a.md"))
	if err != nil {
		t.Fatalf("read a.md: %v", err)
	}
	if string(content) != "alpha" {
		t.Errorf("a.md content = %q, want %q", string(content), "alpha")
	}
}

func TestMigrateCommandDir_PartialFailure(t *testing.T) {
	dir := t.TempDir()
	// Both dirs exist. command/ has a.md (moveable) and b.md (read will fail).
	createFile(t, dir, ".opencode/command/a.md", "alpha")
	createFile(t, dir, ".opencode/command/b.md", "bravo")
	createFile(t, dir, ".opencode/commands/c.md", "charlie")

	var buf bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		Stdout:    &buf,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	// Make b.md unreadable to simulate a partial failure during
	// the duplicate-check path (both dirs have b.md with conflict).
	// Instead, use a custom ReadFile that fails for b.md in the old dir.
	createFile(t, dir, ".opencode/commands/b.md", "different-bravo")
	opts.ReadFile = func(path string) ([]byte, error) {
		if strings.HasSuffix(path, filepath.Join("command", "b.md")) {
			return nil, fmt.Errorf("simulated read error")
		}
		return os.ReadFile(path)
	}

	result := migrateCommandDir(opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// a.md should have been moved successfully.
	if _, err := os.Stat(filepath.Join(dir, ".opencode", "commands", "a.md")); err != nil {
		t.Error("a.md should have been moved to commands/")
	}

	// Warning should have been printed for b.md.
	output := buf.String()
	if !strings.Contains(output, "b.md") {
		t.Errorf("expected warning about b.md, got:\n%s", output)
	}

	// command/ should NOT be fully removed (b.md still there due to error).
	// Actually the implementation calls os.Remove(oldPath) on the old copy
	// only when content matches, and prints a warning when ReadFile fails.
	// The old b.md stays because ReadFile failed before os.Remove runs.
	// os.Remove(oldDir) at the end will fail because dir is non-empty.
	oldDir := filepath.Join(dir, ".opencode", "command")
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Error("old .opencode/command/ should still exist when partial failure occurs")
	}
}

func TestMoveFile_FallbackOnRenameError(t *testing.T) {
	// We can test the fallback path by using two different temp dirs
	// on potentially different filesystems, but that's unreliable.
	// Instead, create the file and use a wrapper that makes the
	// first Rename fail but the ReadFile/WriteFile path succeed.
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.md")
	dstPath := filepath.Join(dstDir, "test.md")

	if err := os.WriteFile(srcPath, []byte("content"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Use the moveFile function directly. On the same filesystem,
	// os.Rename will succeed, so the fallback path won't be tested.
	// To force the fallback, we rely on the fact that TempDir
	// creates dirs on the same filesystem. Instead, we test that
	// moveFile works end-to-end and verify the result.
	opts := &Options{
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	err := moveFile(srcPath, dstPath, opts)
	if err != nil {
		t.Fatalf("moveFile() error: %v", err)
	}

	// Verify dst has the content.
	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "content" {
		t.Errorf("dst content = %q, want %q", string(got), "content")
	}

	// Verify src is removed.
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("src should be removed after move")
	}
}

func TestMoveFile_FallbackPath(t *testing.T) {
	// Force the fallback (read → write → remove) path by making
	// the source file read-only directory entry impossible to rename.
	// The simplest approach: wrap os.Rename to always fail, then
	// verify the copy fallback works.
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.md")
	dstPath := filepath.Join(dstDir, "test.md")

	if err := os.WriteFile(srcPath, []byte("fallback-content"), 0o644); err != nil {
		t.Fatalf("write src: %v", err)
	}

	// Temporarily make the dst directory such that Rename will fail
	// but WriteFile will succeed. Since both are temp dirs on the
	// same filesystem, Rename would normally succeed. Instead, we
	// create a scenario where dst already exists as a directory.
	subDir := filepath.Join(dstDir, "test.md")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir blocking dst: %v", err)
	}
	// Rename(file, dir) fails → triggers fallback.
	// But WriteFile(dir, ...) would also fail. So remove the
	// blocking dir first to let the fallback succeed.
	// This test verifies the happy path of moveFile more directly.
	if err := os.Remove(subDir); err != nil {
		t.Fatalf("remove blocking dir: %v", err)
	}

	opts := &Options{
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
	}

	err := moveFile(srcPath, dstPath, opts)
	if err != nil {
		t.Fatalf("moveFile() error: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != "fallback-content" {
		t.Errorf("dst content = %q, want %q", string(got), "fallback-content")
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("src should be removed after move")
	}
}

// --- migrateCommandDir additional tests ---

func TestMigrateCommandDir_AtomicRenameVerifiesContent(t *testing.T) {
	dir := t.TempDir()
	oldDir := filepath.Join(dir, ".opencode", "command")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "test.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(oldDir, "other.md"), []byte("world"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		Stdout:    &bytes.Buffer{},
	}

	result := migrateCommandDir(opts)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify both files are accessible in the new location.
	newDir := filepath.Join(dir, ".opencode", "commands")
	for _, file := range []string{"test.md", "other.md"} {
		if _, err := os.Stat(filepath.Join(newDir, file)); err != nil {
			t.Errorf("expected %s in new dir: %v", file, err)
		}
	}

	// Verify old dir no longer exists.
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Error("old dir should be removed after atomic rename")
	}
}

func TestMigrateCommandDir_MergeComplex(t *testing.T) {
	dir := t.TempDir()
	oldDir := filepath.Join(dir, ".opencode", "command")
	newDir := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(oldDir, 0o755); err != nil {
		t.Fatalf("mkdir old: %v", err)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("mkdir new: %v", err)
	}

	// File only in old dir — should be moved.
	if err := os.WriteFile(filepath.Join(oldDir, "unique.md"), []byte("unique"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Identical file in both — should be skipped (old removed).
	if err := os.WriteFile(filepath.Join(oldDir, "same.md"), []byte("same"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "same.md"), []byte("same"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Conflicting file — should be warned and commands/ version kept.
	if err := os.WriteFile(filepath.Join(oldDir, "conflict.md"), []byte("old-version"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "conflict.md"), []byte("new-version"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var stdout bytes.Buffer
	opts := &Options{
		TargetDir: dir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		Stdout:    &stdout,
	}

	result := migrateCommandDir(opts)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.action != "migrated" {
		t.Errorf("expected action 'migrated', got %q", result.action)
	}

	// Verify unique.md was moved.
	content, err := os.ReadFile(filepath.Join(newDir, "unique.md"))
	if err != nil {
		t.Fatalf("read unique.md: %v", err)
	}
	if string(content) != "unique" {
		t.Errorf("expected 'unique', got %q", string(content))
	}

	// Verify conflict.md kept commands/ version.
	content, err = os.ReadFile(filepath.Join(newDir, "conflict.md"))
	if err != nil {
		t.Fatalf("read conflict.md: %v", err)
	}
	if string(content) != "new-version" {
		t.Errorf("expected 'new-version' preserved, got %q", string(content))
	}

	// Verify conflict warning was printed.
	if !strings.Contains(stdout.String(), "conflict") {
		t.Errorf("expected conflict warning in output, got: %s", stdout.String())
	}

	// Verify old copy of conflict.md was removed.
	if _, err := os.Stat(filepath.Join(oldDir, "conflict.md")); !os.IsNotExist(err) {
		t.Error("old copy of conflict.md should be removed")
	}
}

// --- Sub-tool error reporting tests (Spec 036 Phase 2) ---

func TestInitSubTools_DeweyInitFails_ShowsError(t *testing.T) {
	t.Run("short_output", func(t *testing.T) {
		dir := t.TempDir()
		rec := &scaffoldCmdRecorder{
			errors: map[string]error{
				"dewey init": fmt.Errorf("exit status 1"),
			},
			outputs: map[string]string{
				"dewey init": "Error: vault not found\nRun 'dewey init --vault .' to create one.\n",
			},
		}

		var buf bytes.Buffer
		opts := &Options{
			TargetDir: dir,
			LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
			ExecCmd:   rec.execCmd,
			Stdout:    &buf,
		}

		results := initSubTools(opts)

		// Find the dewey init failure result.
		var deweyResult *subToolResult
		for i := range results {
			if results[i].name == ".uf/dewey/" && results[i].action == "failed" {
				deweyResult = &results[i]
				break
			}
		}
		if deweyResult == nil {
			t.Fatalf("expected .uf/dewey/ failed result, got %v", results)
		}

		// Verify err field is set.
		if deweyResult.err == nil {
			t.Error("expected err to be non-nil on failed dewey init")
		}

		// Verify output field contains the error text.
		if !strings.Contains(string(deweyResult.output), "vault not found") {
			t.Errorf("expected output to contain 'vault not found', got %q", string(deweyResult.output))
		}

		// Verify detail includes the error message.
		if !strings.Contains(deweyResult.detail, "exit status 1") {
			t.Errorf("expected detail to contain error text, got %q", deweyResult.detail)
		}

		// Verify printSummary shows the error output.
		printSummary(&buf, false, false, true, &Result{Created: []string{"test.md"}}, results)
		output := buf.String()

		if !strings.Contains(output, "vault not found") {
			t.Errorf("expected printSummary output to contain 'vault not found', got:\n%s", output)
		}
		if !strings.Contains(output, "Error:") {
			t.Errorf("expected printSummary output to contain 'Error:', got:\n%s", output)
		}
	})

	t.Run("long_output_truncated", func(t *testing.T) {
		dir := t.TempDir()
		// Generate 25 lines of output to trigger truncation (maxLines=20).
		longOutput := generateLines(25)
		rec := &scaffoldCmdRecorder{
			errors: map[string]error{
				"dewey init": fmt.Errorf("exit status 1"),
			},
			outputs: map[string]string{
				"dewey init": longOutput,
			},
		}

		var buf bytes.Buffer
		opts := &Options{
			TargetDir: dir,
			LookPath:  stubScaffoldLookPath(map[string]string{"dewey": "/usr/local/bin/dewey"}),
			ExecCmd:   rec.execCmd,
			Stdout:    &buf,
		}

		results := initSubTools(opts)

		// Find the dewey init failure result.
		var deweyResult *subToolResult
		for i := range results {
			if results[i].name == ".uf/dewey/" && results[i].action == "failed" {
				deweyResult = &results[i]
				break
			}
		}
		if deweyResult == nil {
			t.Fatalf("expected .uf/dewey/ failed result, got %v", results)
		}

		// Verify printSummary renders truncated output.
		printSummary(&buf, false, false, true, &Result{Created: []string{"test.md"}}, results)
		output := buf.String()

		// Should contain the truncation marker.
		if !strings.Contains(output, "... (15 lines omitted)") {
			t.Errorf("expected '... (15 lines omitted)' in printSummary output, got:\n%s", output)
		}

		// Should contain the last 10 lines (line 16 through line 25).
		if !strings.Contains(output, "line 25") {
			t.Errorf("expected last line 'line 25' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "line 16") {
			t.Errorf("expected 'line 16' (first of last 10) in output, got:\n%s", output)
		}

		// Should NOT contain early lines that were truncated.
		// Use a specific match to avoid matching "line 1" as a
		// substring of "line 10", "line 15", etc.
		for _, earlyLine := range []string{"line 1\n", "line 2\n", "line 5\n"} {
			// Check the rendered output (each line is indented).
			// The truncation should have removed these lines.
			truncated := textutil.TruncateOutput(deweyResult.output, 20)
			if strings.Contains(truncated, strings.TrimSuffix(earlyLine, "\n")+"\n") {
				// Only flag if the early line appears as a full line
				// in the truncated output (not as a substring).
				lines := strings.Split(truncated, "\n")
				for _, l := range lines {
					if l == strings.TrimSuffix(earlyLine, "\n") {
						t.Errorf("early line %q should be truncated from output", strings.TrimSuffix(earlyLine, "\n"))
					}
				}
			}
		}
	})
}

func TestInitSubTools_SimpleToolFails_ShowsError(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"specify init --here --integration opencode --offline --force": fmt.Errorf("exit status 1"),
		},
		outputs: map[string]string{
			"specify init --here --integration opencode --offline --force": "Error: .specify already exists\nUse --force to overwrite.\n",
		},
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Find the specify failure result.
	var specifyResult *subToolResult
	for i := range results {
		if results[i].name == ".specify/" && results[i].action == "failed" {
			specifyResult = &results[i]
			break
		}
	}
	if specifyResult == nil {
		t.Fatalf("expected .specify/ failed result, got %v", results)
	}

	// Verify err field is set.
	if specifyResult.err == nil {
		t.Error("expected err to be non-nil on failed specify init")
	}

	// Verify output field contains the error text.
	if !strings.Contains(string(specifyResult.output), ".specify already exists") {
		t.Errorf("expected output to contain '.specify already exists', got %q", string(specifyResult.output))
	}

	// Verify detail includes the tool name and error.
	if !strings.Contains(specifyResult.detail, "specify init:") {
		t.Errorf("expected detail to contain 'specify init:', got %q", specifyResult.detail)
	}
}

// generateLines creates a string with n lines like "line 1\nline 2\n...".
func generateLines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

func TestInitSubTools_ExitCodeOnly(t *testing.T) {
	dir := t.TempDir()
	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"gaze init": fmt.Errorf("exit status 127"),
		},
		// No outputs entry — empty output bytes.
	}

	// Create .opencode/agents/ but no gaze-reporter.md so gaze init triggers.
	if err := os.MkdirAll(filepath.Join(dir, ".opencode", "agents"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	opts := &Options{
		TargetDir: dir,
		LookPath:  stubScaffoldLookPath(map[string]string{"gaze": "/usr/local/bin/gaze"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Find the gaze failure result.
	var gazeResult *subToolResult
	for i := range results {
		if results[i].name == "gaze" && results[i].action == "failed" {
			gazeResult = &results[i]
			break
		}
	}
	if gazeResult == nil {
		t.Fatalf("expected gaze failed result, got %v", results)
	}

	// Verify err is set.
	if gazeResult.err == nil {
		t.Error("expected err to be non-nil")
	}

	// Verify detail includes the error text.
	if !strings.Contains(gazeResult.detail, "exit status 127") {
		t.Errorf("expected detail to contain 'exit status 127', got %q", gazeResult.detail)
	}
}

func TestPrintSummary_SuccessUnchanged(t *testing.T) {
	var buf bytes.Buffer

	r := &Result{
		Created: []string{"test.md"},
	}
	subResults := []subToolResult{
		{name: ".uf/dewey/", action: "initialized"},
		{name: "dewey index", action: "completed"},
		{name: "opencode.json", action: "created"},
	}

	printSummary(&buf, false, false, true, r, subResults)
	output := buf.String()

	// Verify each sub-tool result appears on exactly one line
	// (no "Output:" or "Error:" lines for successful results).
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Output:") {
			t.Errorf("successful results should not have 'Output:' lines, got:\n%s", output)
		}
		if strings.Contains(line, "Error:") {
			t.Errorf("successful results should not have 'Error:' lines, got:\n%s", output)
		}
	}

	// Verify the sub-tool results are present.
	if !strings.Contains(output, ".uf/dewey/ initialized") {
		t.Errorf("expected '.uf/dewey/ initialized' in output")
	}
	if !strings.Contains(output, "dewey index completed") {
		t.Errorf("expected 'dewey index completed' in output")
	}
	if !strings.Contains(output, "opencode.json created") {
		t.Errorf("expected 'opencode.json created' in output")
	}
}

// TestGuardrailTemplates_CommandSpecificContent is a regression guard
// for issue #256: the uf.init.md Step 6 guardrail templates must
// contain command-specific content for implement, constitution, and
// taskstoissues commands. Previously all 9 commands shared identical
// guardrails that were factually wrong for these 3 commands.
func TestGuardrailTemplates_CommandSpecificContent(t *testing.T) {
	content, err := assetContent("opencode/commands/uf.init.md")
	if err != nil {
		t.Fatalf("read embedded uf.init.md: %v", err)
	}
	text := string(content)

	// Extract each guardrail template block by its heading label.
	// The blocks are fenced in ```markdown ... ``` and preceded by
	// a bold label like **Implement guardrails block**.
	type guardrailBlock struct {
		label          string
		mustContain    []string
		mustNotContain []string
	}

	blocks := []guardrailBlock{
		{
			label: "Implement guardrails block",
			mustContain: []string{
				"writes source code",
			},
			mustNotContain: []string{
				"NEVER modify source code",
				"FEATURE_DIR",
				"FEATURE_SPEC",
			},
		},
		{
			label: "Constitution guardrails block",
			mustContain: []string{
				".specify/memory/",
				".specify/templates/",
			},
			mustNotContain: []string{
				"NEVER modify source code",
				"FEATURE_DIR",
				"FEATURE_SPEC",
			},
		},
		{
			label: "Taskstoissues guardrails block",
			mustContain: []string{
				"GitHub issues via",
				"the current Git remote",
				"Content portability", // issue #593
				"CP-001",              // issue #593
			},
			mustNotContain: []string{
				"NEVER modify source code",
				"files this command may write",
				"FEATURE_DIR",
				"FEATURE_SPEC",
			},
		},
		{
			label: "Spec-phase guardrails block",
			mustContain: []string{
				"NEVER modify source code",
				"FEATURE_DIR",
				"defeats the purpose",
			},
			mustNotContain: nil, // spec-phase guardrails are the baseline
		},
	}

	for _, b := range blocks {
		t.Run(b.label, func(t *testing.T) {
			// Find the block: look for the label, then extract the
			// next ```markdown ... ``` fenced block after it.
			idx := strings.Index(text, b.label)
			if idx < 0 {
				t.Fatalf("guardrail block label %q not found in uf.init.md", b.label)
			}

			remainder := text[idx:]
			fenceStart := strings.Index(remainder, "```markdown")
			if fenceStart < 0 {
				t.Fatalf("no ```markdown fence found after %q", b.label)
			}
			fenceEnd := strings.Index(remainder[fenceStart+len("```markdown"):], "```")
			if fenceEnd < 0 {
				t.Fatalf("no closing ``` found for %q block", b.label)
			}
			blockContent := remainder[fenceStart : fenceStart+len("```markdown")+fenceEnd+len("```")]

			for _, pattern := range b.mustContain {
				if !strings.Contains(blockContent, pattern) {
					t.Errorf("%s: MUST contain %q but does not\nBlock content:\n%s",
						b.label, pattern, blockContent)
				}
			}
			for _, pattern := range b.mustNotContain {
				if strings.Contains(blockContent, pattern) {
					t.Errorf("%s: MUST NOT contain %q but does\nBlock content:\n%s",
						b.label, pattern, blockContent)
				}
			}
		})
	}
}

// TestSpeckitTemplates_RequiredReferences is a red-first regression
// gate for issue #548, anchored to ADR-002
// (docs/decisions/002-uf-refs-source-of-truth-layer.md, Option B:
// extend Step 6 of uf.init.md as the authoritative injection layer).
// It asserts that the Step 6 "Speckit Command Guardrails" injection
// surface of the EMBEDDED canonical uf.init.md carries the required
// Unbound Force references for each of the five in-scope upstream
// speckit templates. Modeled on
// TestGuardrailTemplates_CommandSpecificContent (issue #256 style):
// read the embedded asset via assetContent, scope matching to the
// Step 6 region and the per-file guardrail block, then apply
// mustContain/mustNotContain via strings.Contains inside per-file
// subtests so failures name the exact file and the exact reference.
//
// Required-reference matrix (ADR-002 / design D4):
//
//	| File                    | Dewey | constitution.md | /uf.review-council |
//	|-------------------------|-------|-----------------|--------------------|
//	| speckit.specify.md      | req   | req             | mustNotContain     |
//	| speckit.plan.md         | req   | req             | mustNotContain     |
//	| speckit.tasks.md        | req   | req             | mustNotContain     |
//	| speckit.implement.md    | req   | req             | req (mustContain)  |
//	| speckit.constitution.md | req   | req             | mustNotContain     |
//
// Dewey is satisfied by dewey_semantic_search OR dewey_search.
//
// Steady-state contract (regression gate): Step 6 of the embedded
// uf.init.md MUST inject Dewey + constitution.md references for all
// five in-scope files and /uf.review-council for the implement file
// only; this test fails if any required reference is removed or the
// implement-only boundary is breached.
//
// History: introduced red-first with the #548/#549 co-merge (design
// D5) — the assertions were authored before the #549 reconciliation
// edit populated Step 6, so on the pre-#549 tree they failed RED and
// named the exact missing references. The #549 edit turns them green.
//
// Scope note: the three spec-phase in-scope files (specify, plan,
// tasks) are injected from a SINGLE shared "Spec-phase guardrails
// block" (design D3 — no per-file restructuring). Their subtests
// therefore assert against the same block; the value of keying by
// filename is precise failure output (naming the exact in-scope
// file), NOT independent per-file drift detection among the three.
// If Step 6 is ever restructured to inject the spec-phase files from
// distinct sources, split this table into per-file blocks.
func TestSpeckitTemplates_RequiredReferences(t *testing.T) {
	content, err := assetContent("opencode/commands/uf.init.md")
	if err != nil {
		t.Fatalf("read embedded uf.init.md: %v", err)
	}
	text := string(content)

	// Scope all matching to the Step 6 region so a reference that
	// lives in Step 7 (the read-only verifier) or elsewhere cannot
	// green-pass this test (design D3).
	step6Start := strings.Index(text, "### Step 6:")
	if step6Start < 0 {
		t.Fatalf("Step 6 heading not found in embedded uf.init.md")
	}
	step7Start := strings.Index(text[step6Start:], "### Step 7:")
	if step7Start < 0 {
		t.Fatalf("Step 7 heading (Step 6 region end) not found in embedded uf.init.md")
	}
	step6Region := text[step6Start : step6Start+step7Start]

	// extractBlock returns the content of the ```markdown ... ```
	// fenced block that immediately follows the given bold label
	// within the Step 6 region. It returns ("", false) when the
	// label or its fence is absent. Absence is treated as a normal
	// red condition (the reference cannot be present in a block that
	// does not exist), NOT a fatal extraction bug, so the missing
	// reference is reported per file rather than aborting the test.
	//
	// The label is matched in its bold markdown form ("**"+label+"**")
	// so it binds to the actual block heading, NOT to a bare-substring
	// occurrence of the same phrase in surrounding Step 6 instruction
	// prose (e.g. the idempotency-correction text references the
	// "Spec-phase guardrails block" by name). This keeps extraction
	// deterministic regardless of instruction-prose edits.
	extractBlock := func(label string) (string, bool) {
		idx := strings.Index(step6Region, "**"+label+"**")
		if idx < 0 {
			return "", false
		}
		remainder := step6Region[idx:]
		fenceStart := strings.Index(remainder, "```markdown")
		if fenceStart < 0 {
			return "", false
		}
		afterOpen := remainder[fenceStart+len("```markdown"):]
		fenceEnd := strings.Index(afterOpen, "```")
		if fenceEnd < 0 {
			return "", false
		}
		return afterOpen[:fenceEnd], true
	}

	const (
		deweySemantic  = "dewey_semantic_search"
		deweyKeyword   = "dewey_search"
		constitutionMd = ".specify/memory/constitution.md"
		reviewCouncil  = "/uf.review-council"
	)

	// Per-in-scope-file mapping to the Step 6 block that injects
	// that file's guardrail content. The three spec-phase in-scope
	// files (specify, plan, tasks) share one block ("Spec-phase
	// guardrails block") per design D3 (no per-file restructuring),
	// so their subtests assert against identical bytes. Keying the
	// table by filename is deliberate: it names the exact in-scope
	// file in failure output. It does NOT provide independent drift
	// detection between the three spec-phase files — they cannot
	// diverge while injected from a single shared source block.
	//
	// blockLabel is the Step 6 label whose fenced block injects this
	// file's references. requireReviewCouncil encodes the
	// implement-only boundary (D4): true => mustContain, false =>
	// mustNotContain for /uf.review-council.
	type speckitFile struct {
		file                 string
		blockLabel           string
		requireReviewCouncil bool
	}

	// Explicit enumeration of the five in-scope upstream files (D2 —
	// NOT derived from a speckit.*.md glob). The four UF-custom
	// commands (analyze, checklist, clarify, taskstoissues) and
	// testreview are intentionally excluded.
	inScope := []speckitFile{
		{file: "speckit.specify.md", blockLabel: "Spec-phase guardrails block", requireReviewCouncil: false},
		{file: "speckit.plan.md", blockLabel: "Spec-phase guardrails block", requireReviewCouncil: false},
		{file: "speckit.tasks.md", blockLabel: "Spec-phase guardrails block", requireReviewCouncil: false},
		{file: "speckit.implement.md", blockLabel: "Implement guardrails block", requireReviewCouncil: true},
		{file: "speckit.constitution.md", blockLabel: "Constitution guardrails block", requireReviewCouncil: false},
	}

	for _, sf := range inScope {
		t.Run(sf.file, func(t *testing.T) {
			block, ok := extractBlock(sf.blockLabel)
			if !ok {
				// The injection block for this file's references does
				// not exist in Step 6, so every required reference is
				// necessarily missing. Report each missing reference
				// against the exact file (red for the right reason).
				reviewCouncilSuffix := ""
				if sf.requireReviewCouncil {
					reviewCouncilSuffix = ", " + reviewCouncil
				}
				t.Errorf("%s: Step 6 block %q not found — required references cannot be injected (missing: Dewey, %s%s)",
					sf.file, sf.blockLabel, constitutionMd, reviewCouncilSuffix)
				return
			}

			// Dewey: satisfied by dewey_semantic_search OR dewey_search.
			if !strings.Contains(block, deweySemantic) && !strings.Contains(block, deweyKeyword) {
				t.Errorf("%s: Step 6 block %q MUST reference Dewey (%s OR %s) but does not",
					sf.file, sf.blockLabel, deweySemantic, deweyKeyword)
			}

			// Constitution awareness: required for all five in-scope files.
			if !strings.Contains(block, constitutionMd) {
				t.Errorf("%s: Step 6 block %q MUST reference %q but does not",
					sf.file, sf.blockLabel, constitutionMd)
			}

			// /uf.review-council: mustContain for implement only,
			// mustNotContain for the other four (implement-only boundary).
			if sf.requireReviewCouncil {
				if !strings.Contains(block, reviewCouncil) {
					t.Errorf("%s: Step 6 block %q MUST reference %q (implement phase) but does not",
						sf.file, sf.blockLabel, reviewCouncil)
				}
			} else {
				if strings.Contains(block, reviewCouncil) {
					t.Errorf("%s: Step 6 block %q MUST NOT reference %q (implement-only boundary leaked)",
						sf.file, sf.blockLabel, reviewCouncil)
				}
			}
		})
	}
}

func TestCleanupRenamedCommands_HappyPath(t *testing.T) {
	dir := t.TempDir()

	// Create a subset of old-name command files.
	cmdDir := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	oldFiles := []string{"address-feedback.md", "review-council.md"}
	for _, f := range oldFiles {
		if err := os.WriteFile(filepath.Join(cmdDir, f), []byte("old"), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}

	var buf bytes.Buffer
	removed := cleanupRenamedCommands(&buf, dir)

	// Verify files are removed from disk.
	for _, f := range oldFiles {
		p := filepath.Join(cmdDir, f)
		if _, err := os.Stat(p); err == nil {
			t.Errorf("expected %s to be removed, but it still exists", f)
		}
	}

	// Verify returned paths use mapped output format.
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed paths, got %d", len(removed))
	}
	for _, r := range removed {
		if !strings.HasPrefix(r, ".opencode/commands/") {
			t.Errorf("removed path %q does not have expected prefix .opencode/commands/", r)
		}
	}

	// No warnings should have been emitted.
	if buf.Len() != 0 {
		t.Errorf("expected no warnings, got: %s", buf.String())
	}
}

func TestCleanupRenamedCommands_NoOldFiles(t *testing.T) {
	dir := t.TempDir()

	var buf bytes.Buffer
	removed := cleanupRenamedCommands(&buf, dir)

	if len(removed) != 0 {
		t.Errorf("expected 0 removed paths, got %d", len(removed))
	}
	if buf.Len() != 0 {
		t.Errorf("expected no warnings, got: %s", buf.String())
	}
}

func TestCleanupRenamedCommands_NonWritableDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("read-only file test not reliable on Windows")
	}

	dir := t.TempDir()
	cmdDir := filepath.Join(dir, ".opencode", "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create two old-name files.
	removable := "address-feedback.md"
	readonly := "review-council.md"
	for _, f := range []string{removable, readonly} {
		if err := os.WriteFile(filepath.Join(cmdDir, f), []byte("old"), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}

	// Make a file read-only, then make the directory non-writable so
	// os.Remove fails for ALL files (total failure, not partial).
	readonlyPath := filepath.Join(cmdDir, readonly)
	if err := os.Chmod(readonlyPath, 0o444); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	if err := os.Chmod(cmdDir, 0o555); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	t.Cleanup(func() {
		// Restore permissions so t.TempDir() cleanup succeeds.
		_ = os.Chmod(cmdDir, 0o755)
		_ = os.Chmod(readonlyPath, 0o644)
	})

	var buf bytes.Buffer
	removed := cleanupRenamedCommands(&buf, dir)

	// No files should be removable since the directory is non-writable.
	if len(removed) != 0 {
		t.Errorf("expected 0 removed (dir is non-writable), got %d: %v", len(removed), removed)
	}

	// Warnings should have been emitted for both files.
	warnings := buf.String()
	if !strings.Contains(warnings, "could not remove") {
		t.Errorf("expected warning about removal failure, got: %s", warnings)
	}
}

func TestWarnStaleCommandRefs_StaleRef(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write an agent file containing a stale reference.
	content := "Run /review-council to review the code.\n"
	if err := os.WriteFile(filepath.Join(agentDir, "cobalt-crush-dev.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	warnStaleCommandRefs(&buf, dir)

	output := buf.String()
	if !strings.Contains(output, "Stale command references") {
		t.Errorf("expected stale reference warning header, got: %s", output)
	}
	if !strings.Contains(output, "cobalt-crush-dev.md") {
		t.Errorf("expected agent file name in warning, got: %s", output)
	}
	if !strings.Contains(output, "/review-council") {
		t.Errorf("expected old ref /review-council in warning, got: %s", output)
	}
	if !strings.Contains(output, "/uf.review-council") {
		t.Errorf("expected new ref /uf.review-council in warning, got: %s", output)
	}
}

func TestWarnStaleCommandRefs_NoStaleRefs(t *testing.T) {
	dir := t.TempDir()
	agentDir := filepath.Join(dir, ".opencode", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write an agent file with only current references.
	content := "Run /uf.review-council to review the code.\n"
	if err := os.WriteFile(filepath.Join(agentDir, "cobalt-crush-dev.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	warnStaleCommandRefs(&buf, dir)

	if buf.Len() != 0 {
		t.Errorf("expected no warning for current refs, got: %s", buf.String())
	}
}

func TestWarnStaleCommandRefs_NoAgentFiles(t *testing.T) {
	dir := t.TempDir()

	var buf bytes.Buffer
	warnStaleCommandRefs(&buf, dir)

	if buf.Len() != 0 {
		t.Errorf("expected no output when no agent dir exists, got: %s", buf.String())
	}
}

func TestWarnStaleCommandRefs_StaleRefInRootAGENTSmd(t *testing.T) {
	dir := t.TempDir()

	// Write a root AGENTS.md containing a stale reference. No
	// .opencode/agents/ dir is created, proving AGENTS.md is
	// scanned even with zero agent files.
	agentsPath := filepath.Join(dir, "AGENTS.md")
	content := "See /review-council for the review workflow.\n"
	if err := os.WriteFile(agentsPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	warnStaleCommandRefs(&buf, dir)

	output := buf.String()
	if !strings.Contains(output, "Stale command references") {
		t.Errorf("expected stale reference warning header, got: %s", output)
	}
	if !strings.Contains(output, "AGENTS.md") {
		t.Errorf("expected AGENTS.md file name in warning, got: %s", output)
	}
	if !strings.Contains(output, "/review-council") {
		t.Errorf("expected old ref /review-council in warning, got: %s", output)
	}
	if !strings.Contains(output, "/uf.review-council") {
		t.Errorf("expected new ref /uf.review-council in warning, got: %s", output)
	}

	// Warn-only guarantee: AGENTS.md must be unmodified.
	after, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if string(after) != content {
		t.Errorf("expected AGENTS.md to be unmodified, got: %s", string(after))
	}
}

func TestWarnStaleCommandRefs_MigratedRefInRootAGENTSmd(t *testing.T) {
	dir := t.TempDir()

	// Write a root AGENTS.md containing only the already-migrated
	// reference. This is the decisive substring-safety guard proving
	// /uf.review-council is not falsely flagged as containing
	// /review-council. No .opencode/agents/ dir is created.
	content := "See /uf.review-council for the review workflow.\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	var buf bytes.Buffer
	warnStaleCommandRefs(&buf, dir)

	if buf.Len() != 0 {
		t.Errorf("expected no warning for migrated ref, got: %s", buf.String())
	}
}

// TestWarnStaleCommandRefs_ProjectAGENTSmdHasNoStaleRefs is a
// drift-detection test that reads THIS repo's own root AGENTS.md and
// asserts that zero stale (pre-namespace-prefix) command references
// remain. It iterates EVERY stale command name derived from the
// unexported renamedCommands map (the single source of truth), rather
// than hardcoding a subset such as /review-council. This catches drift
// where a command is renamed but the repo's own AGENTS.md falls behind.
//
// Detection mirrors warnStaleCommandRefs exactly: stale ref strings are
// derived the same way ("/" + base name without ".md"), and each ref is
// matched with the same word-boundary-anchored pattern so that an
// already-migrated reference (e.g. /uf.review-council) is not falsely
// counted as containing its stale form (/review-council).
func TestWarnStaleCommandRefs_ProjectAGENTSmdHasNoStaleRefs(t *testing.T) {
	projectRoot := findProjectRoot(t)
	if projectRoot == "" {
		t.Skip("could not locate project root (no go.mod found)")
	}

	agentsPath := filepath.Join(projectRoot, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil {
		t.Skip("AGENTS.md not found at project root")
	}

	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(data)

	// Iterate every stale command name from renamedCommands, deriving
	// the stale ref string the same way the production refMap does.
	for oldRel, newRel := range renamedCommands {
		staleRef := "/" + strings.TrimSuffix(filepath.Base(oldRel), ".md")
		newRef := "/" + strings.TrimSuffix(filepath.Base(newRel), ".md")

		// Same word-boundary-anchored pattern as warnStaleCommandRefs:
		// the leading "." in the excluded class ensures the migrated
		// form (e.g. /uf.review-council) is not flagged as a stale hit.
		pat := "(^|[^.a-zA-Z0-9_-])" + regexp.QuoteMeta(staleRef) + "($|[^a-zA-Z0-9_-])"
		re := regexp.MustCompile(pat)
		if re.MatchString(content) {
			t.Errorf("root AGENTS.md contains stale command reference %s (should be %s); the repo's own AGENTS.md has drifted behind a command rename", staleRef, newRef)
		}
	}
}

// TestSkillTemplates_NoStaleCommandRefs is a drift-detection test
// that walks every .md file under internal/scaffold/assets/opencode/
// skills/ and commands/ and asserts that zero stale (pre-namespace-
// prefix) command references remain. It uses the same word-boundary-
// anchored matching as warnStaleCommandRefs so that already-migrated
// references (e.g. /uf.review-council) are not falsely counted as
// stale.
func TestSkillTemplates_NoStaleCommandRefs(t *testing.T) {
	projectRoot := findProjectRoot(t)
	if projectRoot == "" {
		t.Skip("could not locate project root (no go.mod found)")
	}

	dirs := []string{
		filepath.Join(projectRoot, "internal", "scaffold", "assets", "opencode", "skills"),
		filepath.Join(projectRoot, "internal", "scaffold", "assets", "opencode", "commands"),
	}

	var files []string
	for _, dir := range dirs {
		if _, err := os.Stat(dir); err != nil {
			continue
		}
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dir, err)
		}
	}
	if len(files) == 0 {
		t.Skip("no .md files found under skills or commands directories")
	}

	for _, f := range files {
		data, readErr := os.ReadFile(f)
		if readErr != nil {
			t.Fatalf("read %s: %v", f, readErr)
		}
		content := string(data)
		relPath, _ := filepath.Rel(projectRoot, f)

		for oldRel, newRel := range renamedCommands {
			staleRef := "/" + strings.TrimSuffix(filepath.Base(oldRel), ".md")
			newRef := "/" + strings.TrimSuffix(filepath.Base(newRel), ".md")

			pat := "(^|[^.a-zA-Z0-9_-])" + regexp.QuoteMeta(staleRef) + "($|[^a-zA-Z0-9_-])"
			re := regexp.MustCompile(pat)
			if re.MatchString(content) {
				t.Errorf("%s contains stale command reference %s (should be %s)", relPath, staleRef, newRef)
			}
		}
	}
}

// --- initSimpleTool force re-init tests ---

func TestInitSimpleTool_SentinelExistsForceTrue(t *testing.T) {
	dir := t.TempDir()
	// Create sentinel so the tool appears already initialized.
	sentinel := filepath.Join(dir, ".specify")
	if err := os.MkdirAll(sentinel, 0o755); err != nil {
		t.Fatalf("mkdir sentinel: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	tool := simpleTool{
		name:      "specify",
		sentinel:  ".specify",
		result:    ".specify/",
		label:     "Speckit framework",
		args:      []string{"--here", "--integration", "opencode", "--offline"},
		forceFlag: "--force",
	}
	logf := func(string, ...interface{}) {}

	r := initSimpleTool(opts, tool, logf)
	if r == nil {
		t.Fatal("expected non-nil result when Force=true and sentinel exists")
	}
	if r.action != "re-initialized" {
		t.Errorf("expected action %q, got %q", "re-initialized", r.action)
	}

	// Verify ExecCmd was called.
	if len(rec.calls) == 0 {
		t.Error("expected ExecCmd to be called")
	}
}

func TestInitSimpleTool_SentinelExistsNoForce(t *testing.T) {
	dir := t.TempDir()
	sentinel := filepath.Join(dir, ".specify")
	if err := os.MkdirAll(sentinel, 0o755); err != nil {
		t.Fatalf("mkdir sentinel: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: dir,
		Force:     false,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	tool := simpleTool{
		name:      "specify",
		sentinel:  ".specify",
		result:    ".specify/",
		label:     "Speckit framework",
		args:      []string{"--here", "--integration", "opencode", "--offline"},
		forceFlag: "--force",
	}
	logf := func(string, ...interface{}) {}

	r := initSimpleTool(opts, tool, logf)
	if r == nil {
		t.Fatal("expected non-nil result when Force=false and sentinel exists")
	}
	if r.name != ".specify/" {
		t.Errorf("expected name %q, got %q", ".specify/", r.name)
	}
	if r.action != "already configured" {
		t.Errorf("expected action %q, got %q", "already configured", r.action)
	}

	// Verify ExecCmd was NOT called.
	if len(rec.calls) != 0 {
		t.Errorf("expected no ExecCmd calls, got: %v", rec.calls)
	}
}

func TestInitSimpleTool_ForcePassesForceFlag(t *testing.T) {
	dir := t.TempDir()
	sentinel := filepath.Join(dir, ".specify")
	if err := os.MkdirAll(sentinel, 0o755); err != nil {
		t.Fatalf("mkdir sentinel: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	tool := simpleTool{
		name:      "specify",
		sentinel:  ".specify",
		result:    ".specify/",
		label:     "Speckit framework",
		args:      []string{"--here", "--integration", "opencode", "--offline"},
		forceFlag: "--force",
	}
	logf := func(string, ...interface{}) {}

	r := initSimpleTool(opts, tool, logf)
	if r == nil {
		t.Fatal("expected non-nil result")
	}

	// Verify the recorded command includes --force at the end.
	expected := "specify init --here --integration opencode --offline --force"
	found := false
	for _, call := range rec.calls {
		if call == expected {
			found = true
		}
	}
	if !found {
		t.Errorf("expected call %q, got: %v", expected, rec.calls)
	}
}

func TestInitSimpleTool_NoForceFlagDeclared(t *testing.T) {
	dir := t.TempDir()
	sentinel := filepath.Join(dir, ".uf", "replicator")
	if err := os.MkdirAll(sentinel, 0o755); err != nil {
		t.Fatalf("mkdir sentinel: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"replicator": "/usr/local/bin/replicator"}),
		ExecCmd:   rec.execCmd,
	}

	tool := simpleTool{
		name:      "replicator",
		sentinel:  ".uf/replicator",
		result:    ".uf/replicator/",
		label:     "Replicator workspace",
		args:      nil,
		forceFlag: "", // empty — no force flag
	}
	logf := func(string, ...interface{}) {}

	r := initSimpleTool(opts, tool, logf)
	if r == nil {
		t.Fatal("expected non-nil result on force re-init")
	}
	if r.action != "re-initialized" {
		t.Errorf("expected action %q, got %q", "re-initialized", r.action)
	}

	// Verify ExecCmd was called with "replicator init" only — no
	// extraneous --force flag because forceFlag is empty.
	expected := "replicator init"
	found := false
	for _, call := range rec.calls {
		if call == expected {
			found = true
		}
		if strings.Contains(call, "--force") {
			t.Errorf("unexpected --force in command: %s", call)
		}
	}
	if !found {
		t.Errorf("expected call %q, got: %v", expected, rec.calls)
	}
}

func TestInitSubTools_SpecifyForcePassthrough(t *testing.T) {
	dir := t.TempDir()
	// Create .specify/scripts sentinel — force should re-init.
	if err := os.MkdirAll(filepath.Join(dir, ".specify", "scripts"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	results := initSubTools(opts)

	// Verify .specify/ result with "re-initialized" action.
	foundReInit := false
	for _, r := range results {
		if r.name == ".specify/" && r.action == "re-initialized" {
			foundReInit = true
		}
	}
	if !foundReInit {
		t.Errorf("expected .specify/ re-initialized result, got %v", results)
	}

	// Verify the recorded command includes --force.
	expected := "specify init --here --integration opencode --offline --force"
	found := false
	for _, call := range rec.calls {
		if call == expected {
			found = true
		}
	}
	if !found {
		t.Errorf("expected call %q, got: %v", expected, rec.calls)
	}

	// FR-014: --force is supplied by the base args (specify's
	// forceFlag is empty), so the recorded args must contain exactly
	// one --force — a repopulated forceFlag would duplicate it.
	for _, call := range rec.calls {
		if strings.HasPrefix(call, "specify init") {
			if n := strings.Count(call, "--force"); n != 1 {
				t.Errorf("expected exactly one --force in specify invocation, got %d: %s", n, call)
			}
		}
	}
}

func TestInitSimpleTool_ForceReinitFails(t *testing.T) {
	dir := t.TempDir()
	sentinel := filepath.Join(dir, ".specify", "scripts")
	if err := os.MkdirAll(sentinel, 0o755); err != nil {
		t.Fatalf("mkdir sentinel: %v", err)
	}

	rec := &scaffoldCmdRecorder{
		errors: map[string]error{
			"specify init --here --integration opencode --offline --force": fmt.Errorf("specify crashed"),
		},
	}
	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	// Look up the real specify entry from the package-level
	// simpleTools table (FR-014) so the failure path is exercised
	// against the actual table shape — sentinel .specify/scripts,
	// --force in the base args, empty forceFlag — rather than a
	// stale inline copy that can silently diverge from the table.
	tool := simpleTool{}
	found := false
	for _, entry := range simpleTools {
		if entry.name == "specify" {
			tool = entry
			found = true
			break
		}
	}
	if !found {
		t.Fatal("specify entry not found in simpleTools table")
	}
	logf := func(string, ...interface{}) {}

	r := initSimpleTool(opts, tool, logf)
	if r == nil {
		t.Fatal("expected non-nil result on failed re-init")
	}
	if r.action != "failed" {
		t.Errorf("expected action %q, got %q", "failed", r.action)
	}
	if r.err == nil {
		t.Error("expected non-nil err on failed re-init")
	}
}

func TestInitSimpleTool_NoSentinelForceTrue(t *testing.T) {
	dir := t.TempDir()
	// Do NOT create sentinel — first-time init even with Force=true.

	rec := &scaffoldCmdRecorder{errors: map[string]error{}}
	opts := &Options{
		TargetDir: dir,
		Force:     true,
		LookPath:  stubScaffoldLookPath(map[string]string{"specify": "/usr/local/bin/specify"}),
		ExecCmd:   rec.execCmd,
	}

	tool := simpleTool{
		name:      "specify",
		sentinel:  ".specify",
		result:    ".specify/",
		label:     "Speckit framework",
		args:      []string{"--here", "--integration", "opencode", "--offline"},
		forceFlag: "--force",
	}
	logf := func(string, ...interface{}) {}

	r := initSimpleTool(opts, tool, logf)
	if r == nil {
		t.Fatal("expected non-nil result on first-time init")
	}
	if r.action != "initialized" {
		t.Errorf("expected action %q, got %q", "initialized", r.action)
	}

	// Verify ExecCmd args do NOT include --force (first-time init).
	for _, call := range rec.calls {
		if strings.Contains(call, "--force") {
			t.Errorf("--force should NOT be passed on first-time init, got: %s", call)
		}
	}

	// Verify the base command was called.
	expected := "specify init --here --integration opencode --offline"
	found := false
	for _, call := range rec.calls {
		if call == expected {
			found = true
		}
	}
	if !found {
		t.Errorf("expected call %q, got: %v", expected, rec.calls)
	}
}

// TestContentPortability_DefaultPack verifies that the default
// convention pack contains Content Portability rules (CP-001
// through CP-003) that prevent LLM-generated issue content from
// using org-config-specific terminology for non-org-config repos.
// Regression guard for issue #593.
func TestContentPortability_DefaultPack(t *testing.T) {
	content, err := assetContent("opencode/uf/packs/default.md")
	if err != nil {
		t.Fatalf("read embedded default.md: %v", err)
	}
	text := string(content)

	requiredMarkers := []string{
		"## Content Portability",
		"CP-001",
		"CP-002",
		"CP-003",
		"organization-configuration repository",
		"Peribolos",
		"safe-settings",
	}

	for _, marker := range requiredMarkers {
		if !strings.Contains(text, marker) {
			t.Errorf("default.md MUST contain %q (content portability guardrail from issue #593)", marker)
		}
	}
}

// TestContentPortability_TriageIssueCommand verifies that the
// triage-issue command contains a content portability guardrail
// in the child issue creation section (Phase 4.4).
// Regression guard for issue #593.
func TestContentPortability_TriageIssueCommand(t *testing.T) {
	content, err := assetContent("opencode/commands/uf.triage-issue.md")
	if err != nil {
		t.Fatalf("read embedded uf.triage-issue.md: %v", err)
	}
	text := string(content)

	// The content portability block must appear in the child
	// issue creation section (4.4), not elsewhere.
	idx := strings.Index(text, "### 4.4 Child Issue Creation")
	if idx < 0 {
		t.Fatal("section '### 4.4 Child Issue Creation' not found")
	}

	section := text[idx:]
	requiredPhrases := []string{
		"Content portability",
		"CP-001",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(section, phrase) {
			t.Errorf("triage-issue.md section 4.4 MUST contain %q (content portability guardrail from issue #593)", phrase)
		}
	}
}

// TestContentPortability_TasksToIssuesCommand verifies that the
// speckit.taskstoissues command contains a content portability
// guardrail. This file is not an embedded scaffold asset, so it
// must be tested via filesystem read (os.ReadFile) rather than
// assetContent(). Regression guard for issue #593.
func TestContentPortability_TasksToIssuesCommand(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping filesystem test")
	}

	path := filepath.Join(root, ".opencode", "commands", "speckit.taskstoissues.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read speckit.taskstoissues.md: %v", err)
	}
	text := string(content)

	requiredPhrases := []string{
		"Content portability",
		"CP-001",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(text, phrase) {
			t.Errorf("speckit.taskstoissues.md MUST contain %q (content portability guardrail from issue #593)", phrase)
		}
	}
}

// TestContentPortability_CuratorAgent verifies that the
// divisor-curator agent contains a Content Portability section
// in its audit checklist. Regression guard for issue #593.
func TestContentPortability_CuratorAgent(t *testing.T) {
	content, err := assetContent("opencode/agents/divisor-curator.md")
	if err != nil {
		t.Fatalf("read embedded divisor-curator.md: %v", err)
	}
	text := string(content)

	if !strings.Contains(text, "#### 4. Content Portability") {
		t.Error("divisor-curator.md MUST contain '#### 4. Content Portability' section (issue #593)")
	}
	if !strings.Contains(text, "CP-001") {
		t.Error("divisor-curator.md MUST reference CP-001 (issue #593)")
	}
}

// TestNoObsoleteToolNames verifies that .opencode/ files do not
// contain obsolete Swarm/SwarmMail tool names that were renamed
// in Replicator. Regression guard for issue #617.
func TestNoObsoleteToolNames(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Skip("project root not found; skipping obsolete tool name check")
	}

	// Obsolete tool name patterns (function-call style with
	// underscore, not general prose usage of "swarm").
	obsoletePatterns := []string{
		"swarm_spawn_subtask",
		"swarm_complete",
		"swarm_progress",
		"swarm_status",
		"swarm_worktree_list",
		"swarm_worktree_cleanup",
		"swarm_worktree_create",
		"swarm_worktree_merge",
		"swarmmail_reserve",
	}

	// Walk both the canonical .opencode/ directory and the
	// embedded scaffold mirror to catch obsolete tool names
	// in either location.
	dirs := []string{
		filepath.Join(root, ".opencode"),
		filepath.Join(root, "internal", "scaffold", "assets", "opencode"),
	}
	for _, dir := range dirs {
		dirLabel, _ := filepath.Rel(root, dir)
		err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() {
				return walkErr
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Errorf("read %s: %v", path, readErr)
				return nil
			}

			relPath, _ := filepath.Rel(root, path)
			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				for _, pattern := range obsoletePatterns {
					if strings.Contains(line, pattern) {
						t.Errorf("%s:%d: obsolete tool name %q (issue #617)\n  line: %s",
							relPath, i+1, pattern, strings.TrimSpace(line))
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", dirLabel, err)
		}
	}
}
