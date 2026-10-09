package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type commandClause struct {
	name string
	all  []string
}

type commandCopy struct {
	name string
	text string
}

func TestCommandContracts_ReviewCouncil(t *testing.T) {
	clauses := []commandClause{
		{name: "ascii argument grammar", all: []string{"[code|specs] [pr_number] [--full]", "three token classes may appear in any order", "matching is exact, case-sensitive ascii", "accept at most one mode, one pr number, and one `--full` flag"}},
		{name: "mode pr full precedence", all: []string{"a pr implies `code`", "an explicit valid mode otherwise wins", "existing auto-detection rules", "parsed `full` value"}},
		{name: "argument rejection boundary", all: []string{"`specs` together with a pr number", "repeated modes, pr numbers, or `--full` flags", "conflicting `code` and `specs` modes", "zero, a sign, a decimal, unicode digits, or a value above 999999", "before any discovery or dispatch"}},
		{name: "immutable pr references", all: []string{"baserefname", "baserefoid", "headrefname", "headrefoid", "^[0-9a-f]{40}$", "use only `base_sha...head_sha`", "never substitute the current checkout or later ref values"}},
		{name: "local immutable references", all: []string{"call `resolve_base_ref`", "upstream/main", "origin/main", "resolve the head ref to an immutable sha", "use only the resolved `base_sha...head_sha` afterward"}},
		{name: "validated plan authority", all: []string{"load the `dispatch-advisor` skill", "display the returned json plan exactly", "validated plan is the sole invocation list", "bind this plan to the exact immutable input context"}},
		{name: "six review roles", all: []string{"`divisor-adversary`", "`divisor-architect`", "`divisor-curator`", "`divisor-guard`", "`divisor-sre`", "`divisor-testing`"}},
		{name: "three content roles", all: []string{"`divisor-envoy`", "`divisor-herald`", "`divisor-scribe`", "content-only agents are discovered and reported but never dispatched"}},
		{name: "host source behavior", all: []string{"for `host`, omit both `model` and `tier`", "defaults to the `standard` tier from the review matrix"}},
		{name: "model provenance", all: []string{"requested model/variant", "resolved parent model/variant", "reported child model", "child self-report", "conflict never overwrites authoritative invocation provenance"}},
		{name: "finding deduplication", all: []string{"deduplicate successful run findings by normalized file plus root cause", "retain every contributing run id, agent, model, variant, source, and sequence", "compound severity rules"}},
		{name: "council verdict precedence", all: []string{"any blocking successful run yields `request changes`", "otherwise any advisory yields `approve with advisories`", "otherwise yield `approve`", "failed runs never vote"}},
		{name: "advisory and human fix gate", all: []string{"mandatory gate: human confirmation required", "never apply auto-fixes to spec files without explicit human confirmation", "low/medium auto-fix", "high/critical report only"}},
		{name: "no-success cause precedence", all: []string{"availability-only provider, model, or runtime causes yield `unavailable`", "policy, plan, budget, limit, persistence, calculation, or mixed cause yields `inconclusive`", "both no-success results block automated progression"}},
		{name: "finalizer and dual artifacts", all: []string{"call `finalize_review_dispatch` for every iteration", "persists the `review-dispatch` artifact", "returns canonical `review-verdict` v2 data", "validation or persistence fails", "human-only"}},
		{name: "iterative rerun semantics", all: []string{"re-resolve immutable input context", "recompute and validate the complete plan", "rerun all included plan runs, not only prior blockers", "three iterations are exceeded"}},
	}

	assertCommandCopies(t, "uf.review-council.md", true, clauses)
}

func TestCommandContracts_TriageIssue(t *testing.T) {
	clauses := []commandClause{
		{name: "deterministic issue input", all: []string{"preserve the issue `title` exactly", "preserve `body` as either its exact string or `null`", "preserve all comment bodies exactly", "exact `{title, body, comments}` object"}},
		{name: "planner owns normalization", all: []string{"planner owns nfc and newline normalization", "comment ordering", "versioned length-prefixed framing", "content hashing", "deterministic keyword categories", "exact `text_bytes`", "content tier"}},
		{name: "no synthetic diff", all: []string{"never synthesize a code diff from issue content", "never request, synthesize, or infer a code diff", "no `changed_files` field"}},
		{name: "fixed base vector", all: []string{"coverage regression", "text_bytes: 19", "60a7be4394be0186439a588e93c57dc9095c1b96f42f27464dbfeb2ce5cfdbf5"}},
		{name: "fixed boundary vectors", all: []string{"4096", "4f30e0423cec84abfc13ae44c9fe73dde044a0852c5492884e52ba020f7b8275", "4097", "19c10c78070cae1fb718628a7e276ed1b9d05d7d3f017bfb32d09bc309aa7207", "32768", "5e06dbe70b16a7d00fb864d511e8542bfdbc1473275d33076f33d5a917be0168", "32769", "15b20fab5e843a6526a81d7809f0c847f1aa237408cd6351b8c2c997d71ae3fe"}},
		{name: "six review universe", all: []string{"six known review-capable personas", "`divisor-adversary`", "`divisor-architect`", "`divisor-curator`", "`divisor-guard`", "`divisor-sre`", "`divisor-testing`"}},
		{name: "triage planner invocation", all: []string{"mode: \"triage\"", "full: false", "augment: false", "`dispatch_agent_run` (not `invoke_agent`) for all dispatch-planned runs"}},
		{name: "one vote per persona", all: []string{"consolidate successful model runs into exactly one assessment per persona", "raw model runs never become independent panel votes", "persona with no successful run produces no persona vote"}},
		{name: "two-stage three-rule majority", all: []string{"three-rule majority in order", "first to the successful model-run verdicts for each persona", "then again to the resulting persona verdicts", "needs-clarification majority", "exclude needs-clarification", "tie-breaking"}},
		{name: "failed runs are non-voting", all: []string{"failed and non-success terminal runs do not vote", "failure is not dissent", "must not create assessments, findings, or advisories"}},
		{name: "no-success handling", all: []string{"at least one successful persona assessment is required", "return **unavailable** only when every blocking cause is provider, model, or runtime availability", "return **inconclusive** for any policy, plan, budget, limit, persistence, calculation, or mixed cause", "skip label, comment, and child-issue mutations"}},
		{name: "issue triage artifact", all: []string{"canonical issue-triage version 1 artifact", ".uf/artifacts/issue-triage/issue-<issue_number>.json", "exactly one consolidated assessment per successful persona"}},
		{name: "dispatch artifact", all: []string{"call `finalize_review_dispatch` for every invocation", "command: \"triage-issue\"", "review-dispatch artifact path", "canonical review-verdict v2 projection"}},
	}

	assertCommandCopies(t, "uf.triage-issue.md", true, clauses)
}

func TestCommandContracts_AddressFeedback(t *testing.T) {
	clauses := []commandClause{
		{name: "tier one preserved", all: []string{"**tier 1 (direct)**", "single file affected", "clear match to a convention pack rule", "no security implications", "no architectural implications"}},
		{name: "deployment fallback preserved", all: []string{"fallback (unchanged)", "if no divisor agents are deployed", "fall back to the existing tier 1 assessment", "distinct from a planned tier 2 dispatch whose runs fail", "must not silently fall back to tier 1"}},
		{name: "tier two planner", all: []string{"advisor-backed tier 2 dispatch", "mode: \"feedback\"", "full: false", "augment: false", "plan_review_dispatch", "`dispatch_agent_run` (not `invoke_agent`) for all dispatch-planned runs"}},
		{name: "immutable feedback signal", all: []string{"freeze this exact immutable input context", "deterministic feedback change signal", "exact same signal for planning, every child prompt, consolidation, and finalization", "do not reuse a plan when the item, context, affected files, categories, or triage data differ"}},
		{name: "one vote per persona", all: []string{"collapse all successful runs for the same persona to exactly one persona recommendation", "fan-out never creates extra persona votes"}},
		{name: "strictest recommendation", all: []string{"any successful run says `accept`, that persona says `accept`", "any `accept` yields native `accept`", "`accept` is stricter than `author-decides`"}},
		{name: "failures non-voting", all: []string{"failures are terminal, informational, and non-voting", "create no findings, advisories, or stronger recommendation"}},
		{name: "planned no-success", all: []string{"availability-only causes produce native, generic, and canonical `unavailable`", "policy, plan, budget, limit, persistence, calculation, or mixed cause", "native, generic, and canonical `inconclusive`", "retain the item for explicit human handling", "must not be converted to `author-decides`"}},
		{name: "tier two finalization", all: []string{"call `finalize_review_dispatch` once for every tier 2 item", "command `address-feedback`", "persists the additive `review-dispatch` artifact", "canonical `review-verdict` version 2"}},
		{name: "feedback triage artifact", all: []string{".uf/artifacts/feedback-triage/pr-<pr_number>-round-<m>.json", "existing feedback-triage artifact", "emitted in addition to each tier 2 `review-dispatch` artifact"}},
	}

	assertCommandCopies(t, "uf.address-feedback.md", true, clauses)
}

func TestCommandContracts_SpeckitTestReview(t *testing.T) {
	clauses := []commandClause{
		{name: "read-only test review", all: []string{"strictly read-only", "never modify files", "spec review mode", "testability", "constitution principle iv"}},
		{name: "testing-only discovery", all: []string{"discovered-agent input for this command must be exactly `[\"divisor-testing\"]`", "do not add any other persona", "every plan entry names only `divisor-testing`"}},
		{name: "testing-only plan", all: []string{"mode: \"test\"", "planner profiles through code-review policy", "matrix_mode` is `code`", "full: false", "augment: false", "no `issue` field"}},
		{name: "explicit advisor host resolution", all: []string{"explicit matrix runs for `divisor-testing` are authoritative", "advisor run only when the returned plan contains one", "exactly one host-source run", "planner alone owns source, model, variant, tier, ordering"}},
		{name: "planned invocation only", all: []string{"call `invoke_agent` for every executable entry", "for `explicit` and `advisor` entries", "for `host`, omit both fields", "never substitute a project, agent, provider, or other default"}},
		{name: "failure behavior", all: []string{"failed and non-success terminal runs do not vote or increase severity", "must not fabricate findings or advisories"}},
		{name: "no-success blocking", all: []string{"with zero successful assessments", "return `unavailable` only when every blocking cause is provider, model, or runtime availability", "policy, plan, budget, limit, persistence, calculation, or mixed cause returns `inconclusive`", "both results block automated progression"}},
		{name: "finalization", all: []string{"call `finalize_review_dispatch` exactly once", "command `speckit-testreview`", "persists the additive `review-dispatch` artifact", "canonical `review-verdict` version 2", "human-only context"}},
	}

	assertCommandCopies(t, "speckit.testreview.md", false, clauses)

	command := loadCommandCopies(t, "speckit.testreview.md", false)[0]
	if strings.Contains(normalizeCommandText(command.text), "task tool") {
		t.Error("speckit.testreview.md must dispatch testing runs through invoke_agent, not the Task tool")
	}
}

func TestCommandContracts_SharedContextAndToolBoundaries(t *testing.T) {
	commands := []struct {
		name             string
		scaffolded       bool
		acquisitionScope string
		immutableBinding string
	}{
		{name: "uf.review-council.md", scaffolded: true, acquisitionScope: "exactly once before the first plan", immutableBinding: "exact immutable input context"},
		{name: "uf.triage-issue.md", scaffolded: true, acquisitionScope: "exactly once before planning", immutableBinding: "bind it to the fetched issue object"},
		{name: "uf.address-feedback.md", scaffolded: true, acquisitionScope: "exactly once before the first plan", immutableBinding: "exact same signal for planning"},
		{name: "speckit.testreview.md", scaffolded: false, acquisitionScope: "exactly once before planning", immutableBinding: "bind it to the immutable local context"},
	}

	for _, command := range commands {
		command := command
		t.Run(command.name, func(t *testing.T) {
			for _, copy := range loadCommandCopies(t, command.name, command.scaffolded) {
				copy := copy
				t.Run(copy.name, func(t *testing.T) {
					normalized := normalizeCommandText(copy.text)
					toolNames := []string{
						"acquire_sibling_evidence",
						"plan_review_dispatch",
						"invoke_agent",
						"prepare_lesson_learning",
						"finalize_review_dispatch",
					}
					for _, toolName := range toolNames {
						if !strings.Contains(normalized, toolName) {
							t.Errorf("%s: shared dispatch contract is missing policy tool %q", copy.name, toolName)
						}
					}
					for toolName, want := range map[string]int{
						"acquire_sibling_evidence": 1,
						"dewey_store_learning":     1,
					} {
						if got := strings.Count(normalized, toolName); got != want {
							t.Errorf("%s: tool %q occurs %d times, want %d to preserve call-once semantics", copy.name, toolName, got, want)
						}
					}

					assertNormalizedContains(t, copy.name, normalized, "sibling acquisition", command.acquisitionScope)
					assertNormalizedContains(t, copy.name, normalized, "immutable input binding", command.immutableBinding)
					assertNormalizedContains(t, copy.name, normalized, "untrusted evidence", "bounded untrusted")
					if !containsAnyNormalized(normalized, []string{"cannot change tools", "cannot change policy, tools"}) {
						t.Errorf("%s: untrusted-evidence contract does not confine tool or policy changes", copy.name)
					}
					assertNormalizedContains(t, copy.name, normalized, "parent-only lessons", "only the parent command processes lesson proposals")
					assertNormalizedContains(t, copy.name, normalized, "lesson dedupe", "uf_lesson_provenance_v1")
					assertNormalizedContains(t, copy.name, normalized, "lesson preparation", "prepare_lesson_learning")
					assertNormalizedContains(t, copy.name, normalized, "store ready lessons once", "dewey_store_learning` exactly once")
					assertNormalizedContains(t, copy.name, normalized, "requested provenance", "requested")
					assertNormalizedContains(t, copy.name, normalized, "resolved provenance", "resolved")
					assertNormalizedContains(t, copy.name, normalized, "reported provenance", "reported")

					for delimiter, want := range map[string]int{
						"<!-- uf-lesson-proposal:v1 -->": 1,
						"<!-- /uf-lesson-proposal -->":   1,
					} {
						if got := strings.Count(copy.text, delimiter); got != want {
							t.Errorf("%s: lesson delimiter %q occurs %d times, want %d", copy.name, delimiter, got, want)
						}
					}
				})
			}
		})
	}
}

func TestCommandContracts_CanonicalScaffoldParity(t *testing.T) {
	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	for _, name := range []string{
		"uf.review-council.md",
		"uf.triage-issue.md",
		"uf.address-feedback.md",
	} {
		name := name
		t.Run(name, func(t *testing.T) {
			canonicalPath := filepath.Join(root, ".opencode", "commands", name)
			canonical, err := os.ReadFile(canonicalPath)
			if err != nil {
				t.Fatalf("read canonical command %s: %v", canonicalPath, err)
			}

			embeddedPath := filepath.ToSlash(filepath.Join("opencode", "commands", name))
			embedded, err := assetContent(embeddedPath)
			if err != nil {
				t.Fatalf("read embedded command %s: %v", embeddedPath, err)
			}
			if !bytes.Equal(canonical, embedded) {
				t.Errorf("canonical command %s differs byte-for-byte from embedded scaffold asset %s", canonicalPath, embeddedPath)
			}
		})

	}
}

func TestCommandContracts_SpeckitTestReviewRemainsLiveOnly(t *testing.T) {
	paths, err := assetPaths()
	if err != nil {
		t.Fatalf("list embedded scaffold assets: %v", err)
	}

	const forbiddenAsset = "opencode/commands/speckit.testreview.md"
	for _, path := range paths {
		if path == forbiddenAsset {
			t.Fatalf("live-only command %s unexpectedly appears in embedded scaffold assets", forbiddenAsset)
		}
	}

	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}
	livePath := filepath.Join(root, ".opencode", "commands", "speckit.testreview.md")
	if info, err := os.Stat(livePath); err != nil {
		t.Fatalf("stat live-only command %s: %v", livePath, err)
	} else if !info.Mode().IsRegular() {
		t.Fatalf("live-only command %s is not a regular file", livePath)
	}
}

func TestCommandContracts_RejectStaleDivisorDispatchLanguage(t *testing.T) {
	staleDirectTaskDispatch := regexp.MustCompile(`(?i)(delegate|dispatch|invoke)[^.]{0,160}(divisor-[a-z-]+|reviewer-[a-z-]+)[^.]{0,160}task tool|task tool[^.]{0,160}(divisor-[a-z-]+|reviewer-[a-z-]+)`)
	stalePersona := regexp.MustCompile(`(?i)\bdivisor-fullsend\b|\breviewer-(adversary|architect|curator|guard|sre|testing)\b`)

	commands := []struct {
		name       string
		scaffolded bool
	}{
		{name: "uf.review-council.md", scaffolded: true},
		{name: "uf.triage-issue.md", scaffolded: true},
		{name: "uf.address-feedback.md", scaffolded: true},
		{name: "speckit.testreview.md", scaffolded: false},
	}

	for _, command := range commands {
		command := command
		t.Run(command.name, func(t *testing.T) {
			for _, copy := range loadCommandCopies(t, command.name, command.scaffolded) {
				normalized := normalizeCommandText(copy.text)
				if match := staleDirectTaskDispatch.FindString(normalized); match != "" {
					t.Errorf("%s: stale direct Divisor Task dispatch wording found: %q", copy.name, match)
				}
				if match := stalePersona.FindString(normalized); match != "" {
					t.Errorf("%s: stale fullsend-only or reviewer-prefixed persona found: %q", copy.name, match)
				}
			}
		})
	}

	reviewCouncil := loadCommandCopies(t, "uf.review-council.md", false)[0]
	normalized := normalizeCommandText(reviewCouncil.text)
	if !strings.Contains(normalized, "`gaze-reporter` agent via the task tool") {
		t.Error("review-council canonical command must retain legitimate Gaze Task-tool delegation")
	}
}

func assertCommandCopies(t *testing.T, name string, scaffolded bool, clauses []commandClause) {
	t.Helper()

	for _, copy := range loadCommandCopies(t, name, scaffolded) {
		copy := copy
		t.Run(copy.name, func(t *testing.T) {
			normalized := normalizeCommandText(copy.text)
			for _, clause := range clauses {
				clause := clause
				t.Run(clause.name, func(t *testing.T) {
					for _, fragment := range clause.all {
						assertNormalizedContains(t, copy.name, normalized, clause.name, fragment)
					}
				})
			}
		})
	}
}

func loadCommandCopies(t *testing.T, name string, scaffolded bool) []commandCopy {
	t.Helper()

	root := findProjectRoot(t)
	if root == "" {
		t.Fatal("project root not found")
	}

	canonicalPath := filepath.Join(root, ".opencode", "commands", name)
	canonical, err := os.ReadFile(canonicalPath)
	if err != nil {
		t.Fatalf("read canonical command %s: %v", canonicalPath, err)
	}
	copies := []commandCopy{{name: "canonical", text: string(canonical)}}
	if !scaffolded {
		return copies
	}

	embeddedPath := filepath.ToSlash(filepath.Join("opencode", "commands", name))
	embedded, err := assetContent(embeddedPath)
	if err != nil {
		t.Fatalf("read embedded command %s: %v", embeddedPath, err)
	}
	return append(copies, commandCopy{name: "scaffold", text: string(embedded)})
}

func normalizeCommandText(text string) string {
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}

func assertNormalizedContains(t *testing.T, command, normalized, contract, fragment string) {
	t.Helper()

	normalizedFragment := normalizeCommandText(fragment)
	if !strings.Contains(normalized, normalizedFragment) {
		t.Errorf("%s: contract %q is missing semantic fragment %q", command, contract, fragment)
	}
}

func containsAnyNormalized(normalized string, fragments []string) bool {
	for _, fragment := range fragments {
		if strings.Contains(normalized, normalizeCommandText(fragment)) {
			return true
		}
	}
	return false
}
