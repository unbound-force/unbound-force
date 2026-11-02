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
		{name: "detached-head fail-closed", all: []string{"`git branch --show-current` returns empty output (detached head", "**stop** with an explicit error naming the resolution", "never proceed with a fabricated, defaulted, or omitted branch"}},
		{name: "validated plan authority", all: []string{"load the `dispatch-advisor` skill", "display the returned json plan exactly", "validated plan is the sole invocation list", "bind this plan to the exact immutable input context"}},
		{name: "six review roles", all: []string{"`divisor-adversary`", "`divisor-architect`", "`divisor-curator`", "`divisor-guard`", "`divisor-sre`", "`divisor-testing`"}},
		{name: "three content roles", all: []string{"`divisor-envoy`", "`divisor-herald`", "`divisor-scribe`", "content-only agents are discovered and reported but never dispatched"}},

		{name: "finding deduplication", all: []string{"deduplicate successful run findings by normalized file plus root cause", "retain every contributing run id, agent, model, variant, source, and sequence", "compound severity rules"}},
		{name: "council verdict precedence", all: []string{"any blocking successful run yields `request changes`", "otherwise any advisory yields `approve with advisories`", "otherwise yield `approve`", "failed runs never vote"}},
		{name: "per-finding human decision gate", all: []string{"mandatory gate: human confirmation required", "one at a time, in severity order (critical, high, medium, low)", "apply suggested fix", "override & continue", "show more context", "pause & exit", "apply all remaining low/medium fixes", "never edit spec files or waive findings without explicit per-finding human confirmation", "human-directed consented edit, not an autonomous auto-fix", "must not proceed past a high/critical finding without"}},
		{name: "no-success cause precedence", all: []string{"availability-only provider, model, or runtime causes yield `unavailable`", "policy, plan, budget, limit, persistence, calculation, or mixed cause yields `inconclusive`", "both no-success results block automated progression"}},
		{name: "finalizer and dual artifacts", all: []string{"finalize every iteration through a single deterministic `consolidate_dispatch` call", "never call `finalize_review_dispatch` yourself", "the canonical review-verdict v2 projection (`review_verdict`)", "classify the failure from the structured response alone", "non-authoritative human context"}},
		{name: "iterative rerun semantics", all: []string{"re-resolve immutable input context", "recompute and validate the complete plan", "rerun all included plan runs, not only prior blockers", "three iterations are exceeded"}},
	}

	assertCommandCopies(t, "uf.review-council.md", true, clauses)
}

func TestCommandContracts_TriageIssue(t *testing.T) {
	clauses := []commandClause{
		{name: "deterministic issue input", all: []string{"preserve the issue `title` exactly", "preserve `body` as either its exact string or `null`", "preserve all comment bodies exactly", "exact `{title, body, comments}` object"}},
		{name: "portable comment ingestion", all: []string{"--jq '.[] | {id, created_at, body}'", "positive numeric `id`"}},
		{name: "deterministic duplicate-search fallback", all: []string{"run no third search and improvise no other keyword sets; two searches total is the fixed bound", `record "none found"`}},
		{name: "detached-head fail-closed", all: []string{"`git branch --show-current` returns empty output (detached head", "**stop** with an explicit error naming the resolution", "never proceed with a fabricated, defaulted, or omitted branch"}},
		{name: "planner owns normalization", all: []string{"planner owns nfc and newline normalization", "comment ordering", "versioned length-prefixed framing", "content hashing", "deterministic keyword categories", "exact `text_bytes`", "content tier"}},
		{name: "no synthetic diff", all: []string{"never synthesize a code diff from issue content", "never request, synthesize, or infer a code diff", "no `changed_files` field"}},
		{name: "fixed base vector", all: []string{"coverage regression", "text_bytes: 19", "60a7be4394be0186439a588e93c57dc9095c1b96f42f27464dbfeb2ce5cfdbf5"}},
		{name: "fixed boundary vectors", all: []string{"4096", "4f30e0423cec84abfc13ae44c9fe73dde044a0852c5492884e52ba020f7b8275", "4097", "19c10c78070cae1fb718628a7e276ed1b9d05d7d3f017bfb32d09bc309aa7207", "32768", "5e06dbe70b16a7d00fb864d511e8542bfdbc1473275d33076f33d5a917be0168", "32769", "15b20fab5e843a6526a81d7809f0c847f1aa237408cd6351b8c2c997d71ae3fe"}},
		{name: "six review universe", all: []string{"six known review-capable personas", "`divisor-adversary`", "`divisor-architect`", "`divisor-curator`", "`divisor-guard`", "`divisor-sre`", "`divisor-testing`"}},
		{name: "triage planner invocation", all: []string{"mode: \"triage\"", "full: false", "augment: false", "`dispatch_review_runs` once with all required inputs"}},
		{name: "one vote per persona", all: []string{"consolidate successful model runs into exactly one assessment per persona", "raw model runs never become independent panel votes", "persona with no successful run produces no persona vote"}},
		{name: "two-stage three-rule majority", all: []string{"three-rule majority in order", "first to the successful model-run verdicts for each persona", "then again to the resulting persona verdicts", "needs-clarification majority", "exclude needs-clarification", "tie-breaking"}},
		{name: "failed runs are non-voting", all: []string{"failed and non-success terminal runs do not vote", "failure is not dissent", "must not create assessments, findings, or advisories"}},
		{name: "no-success handling", all: []string{"at least one successful persona assessment is required", "return **unavailable** only when every blocking cause is provider, model, or runtime availability", "return **inconclusive** for any policy, plan, budget, limit, persistence, calculation, or mixed cause", "skip label, comment, and child-issue mutations"}},
		{name: "issue triage artifact", all: []string{"canonical issue-triage version 1 artifact", ".uf/artifacts/issue-triage/issue-<issue_number>.json", "exactly one consolidated assessment per successful persona"}},
		{name: "dispatch artifact", all: []string{"call `consolidate_dispatch` exactly once per finalization attempt", "\"workflow_id\": \"uf.triage-issue\"", "review-dispatch artifact path", "canonical review-verdict v2 projection"}},
	}

	assertCommandCopies(t, "uf.triage-issue.md", true, clauses)
}

func TestCommandContracts_FixCiCommandStructure(t *testing.T) {
	clauses := []commandClause{
		{name: "never-merge rule", all: []string{"MUST NEVER merge", "no option offers it", "rebase only"}},
		{name: "fr-008 terminal pr state", all: []string{"NEVER merges the PR, approves the PR, closes the PR", "posts any review on it", "remains open, unmerged, and unapproved", "unapproved by this command in every terminal state"}},
		{name: "mandatory gate markers", all: []string{"MANDATORY GATE: HUMAN CONFIRMATION REQUIRED", "END MANDATORY GATE"}},
		{name: "question-tool presentation contract", all: []string{"plain session text **above** the question tool", "per-option reasoning", "the recommended option FIRST and", `labeled "(Recommended)"`, "fail-closed, per FR-013"}},
		{name: "shared gate presentation contract", all: []string{"Gate Presentation Contract", "Do not restate the procedure at the gate sites", "through the **Gate Presentation Contract**", "supplying this gate's content", "a gated mutation never executes ungated"}},
		{name: "rebase gate wiring", all: []string{"Rebase onto <BASE_REMOTE>/<BASE_BRANCH>", "Continue without rebase -- keep the branch as-is", "before `git_fetch` or `git_rebase` runs"}},
		{name: "ambiguity gate wiring", all: []string{"Abort the rebase -- restore the pre-rebase state", "Leave the rebase in progress -- resolve manually", "`git_rebase_abort` without an explicit human choice"}},
		{name: "fix-plan gate wiring", all: []string{"Proceed with the fix plan", "Pause -- defer the decision", "the loop never starts ungated"}},
		{name: "rerun gate wiring", all: []string{"Skip the rerun -- keep the results as-is", "before `gh run rerun` runs"}},
		{name: "session-resume guard", all: []string{"SESSION-RESUME GUARD", "Call `get_progress` with `command: \"uf.fix-ci\"`", "no mutation executes on", "inferred consent"}},
		{name: "checkpoint steps", all: []string{"step: \"pr-resolved\"", "step: \"diagnosis\"", "step: \"rebase\"", "step: \"fix-loop\"", "step: \"final-verification\"", "Then call `clear_progress` with", "`command: \"uf.fix-ci\"` to clean up progress state"}},
		{name: "todowrite tracking", all: []string{"TodoWrite Progress Tracking", "initialize TodoWrite with all six phases", "both MUST be maintained", "re-initialize the TodoWrite list"}},
		{name: "ci polling through plugin tool", all: []string{"uf_finale_ci_poll", "never a bash"}},
		{name: "git plugin tools", all: []string{"git_fetch", "git_rebase", "git_rebase_abort", "git_push", "git_push_force_with_lease", "git_conflict_files", "scan_conflict_markers"}},
		{name: "derived eligibility routing", all: []string{"keyed on the derived `eligibility`", "any check buckets as `fail`", "every check `pass` or `skipping`", "unknown buckets classify fail-closed as `pending`"}},
		{name: "d1 choreography pins", all: []string{"https://github.com/owner/repo.git", "git@github.com:owner/repo.git", "prefer `upstream`, then `origin`", "the numeric path segment immediately following `/actions/runs/`", "^[0-9]+$", "git diff --cached --stat", "never `git add .`, never `git add -A`"}},
		{name: "fr-015 tool-routed fetch invocations", all: []string{"fetch_pr_review_context", "fetch_pr_checks", "fetch_failed_run_logs", "calling `fetch_pr_review_context` with no `pr_number`", "call `fetch_pr_checks` with `pr_number: <PR_NUMBER>`", "`fetch_failed_run_logs` once per unique run-id", "run_id: <RUN_ID>` and `repo: <BASE_REPOSITORY>`", "gh_not_found", "gh_auth_failed"}},
		{name: "review-council gate", all: []string{"/uf.review-council", "REQUEST CHANGES", "Exhaustion MUST NOT push"}},
		{name: "three-iteration cap", all: []string{"capped at 3", "iteration counter", "**STOP** without pushing"}},
		{name: "flaky-infra rerun path", all: []string{"gh run rerun <RUN_ID> --failed --repo <BASE_REPOSITORY>", "Rerun the failed jobs", "No code change is ever proposed for a flaky/infra-classified check"}},
	}

	assertCommandCopies(t, "uf.fix-ci.md", true, clauses)

	// Count pins: the shared presentation contract is defined exactly
	// once and wired into exactly four gates, and the never-merge and
	// terminal-state discipline repeats at a fixed number of sites.
	// Presence-only clauses cannot catch a gate silently losing its
	// wiring or the contract being restated inline at a gate site.
	countPins := []struct {
		fragment string
		want     int
	}{
		{"MANDATORY GATE: HUMAN CONFIRMATION REQUIRED", 4},
		{"END MANDATORY GATE", 4},
		{"remains open, unmerged, and unapproved", 3},
		{"through the **Gate Presentation Contract**", 4},
		{"supplying this gate's content", 4},
		{"plain session text **above** the question tool", 1},
		{"fail-closed, per FR-013", 1},
		{"the recommended option FIRST and", 1},
		{"MUST NEVER merge", 5},
	}
	for _, pin := range countPins {
		for _, copy := range loadCommandCopies(t, "uf.fix-ci.md", true) {
			if got := strings.Count(normalizeCommandText(copy.text), normalizeCommandText(pin.fragment)); got != pin.want {
				t.Errorf("%s: fragment %q occurs %d times, want %d", copy.name, pin.fragment, got, pin.want)
			}
		}
	}

	// The never-merge and FR-008 terminal-state rules are structural:
	// the command must not even name the merge tool or the PR-mutating
	// gh verbs, so no option or push path can reintroduce them.
	// Raw `git merge` (with a space) is a merge vector the plain
	// list cannot hold: a Contains check would also match
	// hyphenated non-merge subcommands such as `git merge-base`,
	// so the raw form is matched with a boundary that rejects a
	// following hyphen.
	rawGitMerge := regexp.MustCompile(`git merge($|[^-])`)
	for _, copy := range loadCommandCopies(t, "uf.fix-ci.md", true) {
		normalized := normalizeCommandText(copy.text)
		for _, forbidden := range []string{
			"git_merge",
			"gh pr merge",
			"gh pr approve",
			"gh pr close",
			"gh pr review",
			"gh pr update-branch",
		} {
			if strings.Contains(normalized, forbidden) {
				t.Errorf("%s: never-merge command must not reference %q", copy.name, forbidden)
			}
		}
		if rawGitMerge.MatchString(normalized) {
			t.Errorf("%s: never-merge command must not reference %q", copy.name, "git merge")
		}
	}

	// FR-015 (extend-fetch-pr-review-context): the three mechanical
	// fetch classes route through the plugin tools, so the raw gh
	// fetch invocations must not return. The pins match the exact
	// invocation forms rather than the bare command names because
	// the bare names legitimately survive in prose: `gh run view`
	// remains in Phase 2.0's run-id derivation (a design D6
	// non-goal), and `gh pr checks` appears in the never-a-bash
	// `--watch` guard. The `gh run rerun --failed` clause above
	// stays positive on purpose — rerun is a mutation, not a fetch,
	// and remains command prose.
	for _, copy := range loadCommandCopies(t, "uf.fix-ci.md", true) {
		normalized := normalizeCommandText(copy.text)
		for _, forbidden := range []string{
			"gh pr view <PR_NUMBER> --json number,state,headRefName,baseRefName,mergeable,mergeStateStatus,headRepository,isCrossRepository,url",
			"gh pr checks <PR_NUMBER> --json name,state,bucket,link",
			"gh run view <RUN_ID> --repo <BASE_REPOSITORY> --log-failed",
		} {
			if strings.Contains(normalized, normalizeCommandText(forbidden)) {
				t.Errorf("%s: tool-routed command must not contain the raw fetch invocation %q", copy.name, forbidden)
			}
		}
	}
}

func TestCommandContracts_AddressFeedback(t *testing.T) {
	clauses := []commandClause{
		{name: "tier one preserved", all: []string{"**tier 1 (direct)**", "single file affected", "clear match to a convention pack rule", "no security implications", "no architectural implications"}},
		{name: "deployment fallback preserved", all: []string{"fallback (unchanged)", "if no divisor agents are deployed", "fall back to the existing tier 1 assessment", "distinct from a planned tier 2 dispatch whose runs fail", "must not silently fall back to tier 1"}},
		{name: "tier two planner", all: []string{"advisor-backed tier 2 dispatch", "mode: \"feedback\"", "full: false", "augment: false", "plan_review_dispatch", "`dispatch_review_runs` once per tier 2 feedback item"}},
		{name: "immutable feedback signal", all: []string{"freeze this exact immutable input context", "deterministic feedback change signal", "exact same signal for planning, every child prompt, consolidation, and finalization", "do not reuse a plan when the item, context, affected files, categories, or triage data differ"}},
		{name: "one vote per persona", all: []string{"collapse all successful runs for the same persona to exactly one persona recommendation", "fan-out never creates extra persona votes"}},
		{name: "strictest recommendation", all: []string{"any successful run says `accept`, that persona says `accept`", "any `accept` yields native `accept`", "`accept` is stricter than `author-decides`"}},
		{name: "failures non-voting", all: []string{"failures are terminal, informational, and non-voting", "create no findings, advisories, or stronger recommendation"}},
		{name: "planned no-success", all: []string{"availability-only causes produce native, generic, and canonical `unavailable`", "policy, plan, budget, limit, persistence, calculation, or mixed cause", "native, generic, and canonical `inconclusive`", "retain the item for explicit human handling", "must not be converted to `author-decides`"}},
		{name: "tier two finalization", all: []string{"finalize every tier 2 item, including plan or no-success failures, with a single `consolidate_dispatch` call", "\"workflow_id\": \"uf.address-feedback\"", "never call `finalize_review_dispatch` yourself", "the canonical review-verdict v2 projection (`review_verdict`)"}},
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
		usesAggregate    bool
		acquisitionScope string
		immutableBinding string
	}{
		{name: "uf.review-council.md", scaffolded: true, usesAggregate: true, acquisitionScope: "acquired exactly once", immutableBinding: "exact immutable input context"},
		{name: "uf.triage-issue.md", scaffolded: true, usesAggregate: true, acquisitionScope: "acquired exactly once", immutableBinding: "bind it to the fetched issue object"},
		{name: "uf.address-feedback.md", scaffolded: true, usesAggregate: true, acquisitionScope: "acquired exactly once", immutableBinding: "exact same signal for planning"},
		{name: "speckit.testreview.md", scaffolded: false, usesAggregate: false, acquisitionScope: "exactly once before planning", immutableBinding: "bind it to the immutable local context"},
	}

	for _, command := range commands {
		command := command
		t.Run(command.name, func(t *testing.T) {
			for _, copy := range loadCommandCopies(t, command.name, command.scaffolded) {
				copy := copy
				t.Run(copy.name, func(t *testing.T) {
					normalized := normalizeCommandText(copy.text)

					// Shared tools present in ALL commands regardless of dispatch style.
					for _, toolName := range []string{
						"plan_review_dispatch",
						"finalize_review_dispatch",
					} {
						if !strings.Contains(normalized, toolName) {
							t.Errorf("%s: shared dispatch contract is missing policy tool %q", copy.name, toolName)
						}
					}

					// Aggregate commands use dispatch_review_runs which encapsulates
					// acquire_sibling_evidence, invoke_agent, and child submission tools.
					// Individual commands (testreview) still call those tools directly.
					if command.usesAggregate {
						if !strings.Contains(normalized, "dispatch_review_runs") {
							t.Errorf("%s: aggregate dispatch contract is missing tool %q", copy.name, "dispatch_review_runs")
						}
					} else {
						for _, toolName := range []string{
							"acquire_sibling_evidence",
							"invoke_agent",
							"submit_lesson_proposal",
						} {
							if !strings.Contains(normalized, toolName) {
								t.Errorf("%s: individual dispatch contract is missing tool %q", copy.name, toolName)
							}
						}
						if got := strings.Count(normalized, "acquire_sibling_evidence"); got != 1 {
							t.Errorf("%s: tool %q occurs %d times, want 1 to preserve call-once semantics", copy.name, "acquire_sibling_evidence", got)
						}
					}

					if got := strings.Count(normalized, "dewey_store_learning"); got != 1 {
						t.Errorf("%s: tool %q occurs %d times, want 1 to preserve call-once semantics", copy.name, "dewey_store_learning", got)
					}

					assertNormalizedContains(t, copy.name, normalized, "sibling acquisition", command.acquisitionScope)
					assertNormalizedContains(t, copy.name, normalized, "immutable input binding", command.immutableBinding)
					assertNormalizedContains(t, copy.name, normalized, "untrusted evidence", "bounded untrusted")
					if !containsAnyNormalized(normalized, []string{"cannot change tools", "cannot change policy, tools"}) {
						t.Errorf("%s: untrusted-evidence contract does not confine tool or policy changes", copy.name)
					}
					if command.usesAggregate {
						assertNormalizedContains(t, copy.name, normalized, "parent-only lessons", "only the parent command stores lesson proposals")
					} else {
						assertNormalizedContains(t, copy.name, normalized, "parent-only lessons", "only the parent command processes lesson proposals")
						assertNormalizedContains(t, copy.name, normalized, "lesson preparation", "prepare_lesson_learning")
					}
					assertNormalizedContains(t, copy.name, normalized, "store ready lessons once", "dewey_store_learning` exactly once")
					assertNormalizedContains(t, copy.name, normalized, "requested provenance", "requested")
					assertNormalizedContains(t, copy.name, normalized, "resolved provenance", "resolved")
					assertNormalizedContains(t, copy.name, normalized, "reported provenance", "reported")

					// submit_review_findings is expected only in commands whose children
					// produce code-review findings AND still document child protocol
					// explicitly. Not expected in triage-issue (verdict/category assessments)
					// or review-council (encapsulated in dispatch_review_runs).
					if strings.Contains(normalized, "submit_review_findings") {
						if got := strings.Count(normalized, "submit_review_findings"); got != 1 {
							t.Errorf("%s: submit_review_findings occurs %d times, want 1", copy.name, got)
						}
					} else if command.name != "uf.triage-issue.md" && command.name != "uf.review-council.md" {
						t.Errorf("%s: expected submit_review_findings for non-aggregate command", copy.name)
					}
				})
			}
		})
	}
}

func TestCommandContracts_CorrelationIdForwarding(t *testing.T) {
	// RD-CF-001 (forward-aggregate-correlation-id): every orchestration
	// command that chains dispatch_review_runs into consolidate_dispatch
	// must capture the correlation ID reported in the one-line summary
	// and pass it as `correlation_id` on every subsequent consolidation
	// call — retrieval, finalization, and identical-call retries — so the
	// ID survives a plugin-process or tool-session split. Presence-only
	// phrase locks per command, reusing the clause mechanism above; each
	// command's clauses pin its landed wording, including the per-item
	// scope in uf.address-feedback.md and the per-dispatch scope in
	// uf.unleash.md. uf.unleash.md has no retrieval-failure section, so
	// it carries no permanent-failure clause; its consolidation surface
	// is finalization plus re-dispatch retries. uf.review-pr.md has no
	// retrieval step either (review-pr-consolidate-dispatch replaced
	// its Sections 5a/5b with a single finalization call): its
	// consolidation surface is finalization plus identical-call retries
	// only, so it too carries no retrieval or permanent-failure clause.
	// uf.review-pr.md and uf.unleash.md join this file's covered command
	// set here (both are scaffolded).
	commands := []struct {
		name    string
		clauses []commandClause
	}{
		{
			name: "uf.review-pr.md",
			clauses: []commandClause{
				{name: "dispatch-step capture and forwarding", all: []string{"capture the `correlation_id` reported in that one-line summary", "pass it as the `correlation_id` argument of the single finalization call in Section 5 and of every identical-call retry", "forward it unconditionally, never conditioned on in-process session-map state"}},
				{name: "finalization forwards the captured id", all: []string{"with the captured `correlation_id` from Section 4c and provenance as its only arguments"}},
				{name: "finalization retry forwards the same captured id", all: []string{"same captured `correlation_id`, same provenance, no changes"}},
			},
		},
		{
			name: "uf.review-council.md",
			clauses: []commandClause{
				{name: "dispatch-step capture and forwarding", all: []string{"capture the `correlation_id` reported in that one-line summary", "pass it as `correlation_id` on every subsequent consolidation call in this workflow", "the Section 4 retrieval, the Section 5 finalization, and every identical-call retry", "forward it unconditionally, never conditioned on in-process session-map state"}},
				{name: "retrieval forwards the captured id", all: []string{"passing the captured `correlation_id` from Section 2"}},
				{name: "retry forwards the same captured id", all: []string{"passing the same captured `correlation_id`"}},
				{name: "finalization forwards the captured id", all: []string{"with the captured `correlation_id` from Section 2 and provenance as its only arguments"}},
				{name: "finalization retry forwards the same captured id", all: []string{"same captured `correlation_id`, same provenance, no changes"}},
				{name: "not-found with the forwarded id is permanent", all: []string{"consolidation carrying the forwarded `correlation_id` for a dispatch that executed runs", "with the ID forwarded, `not_found` means no persisted dispatch session exists", "a plugin-process or tool-session split can no longer produce it", "a `not_found` lacking the forwarded `correlation_id` indicates the capture step was skipped"}},
			},
		},
		{
			name: "uf.triage-issue.md",
			clauses: []commandClause{
				{name: "dispatch-step capture and forwarding", all: []string{"capture the `correlation_id` reported in that one-line summary", "pass it as `correlation_id` on every subsequent consolidation call in this workflow", "the Section 2.3 retrieval, the Section 4.6 finalization, and every identical-call retry", "forward it unconditionally, never conditioned on in-process session-map state"}},
				{name: "retrieval forwards the captured id", all: []string{"passing the captured `correlation_id` from the one-line summary"}},
				{name: "retry forwards the same captured id", all: []string{"passing the same captured `correlation_id`"}},
				{name: "finalization forwards the captured id", all: []string{"with the captured `correlation_id` from the Section 2.3 one-line summary and provenance as its only arguments"}},
				{name: "finalization retry forwards the same captured id", all: []string{"same provenance, same captured `correlation_id`, no changes"}},
				{name: "not-found with the forwarded id is permanent", all: []string{"consolidation carrying the forwarded `correlation_id` for a dispatch that executed runs", "with the ID forwarded, `not_found` means no persisted dispatch session exists", "a plugin-process or tool-session split can no longer produce it", "a `not_found` lacking the forwarded `correlation_id` indicates the capture step was skipped"}},
			},
		},
		{
			name: "uf.address-feedback.md",
			clauses: []commandClause{
				{name: "dispatch-step capture and per-item forwarding", all: []string{"capture the `correlation_id` reported in that one-line summary", "pass it as `correlation_id` on every subsequent consolidation call for that feedback item's dispatch", "the Section 2.3.7 retrieval, the Section 2.3.8 finalization, and every identical-call retry", "the capture and forwarding are scoped per item, not global to the workflow", "forward it unconditionally, never conditioned on in-process session-map state"}},
				{name: "retrieval forwards the item's captured id", all: []string{"passing the item's captured `correlation_id` from Section 2.3.4"}},
				{name: "retry forwards the same captured per-item id", all: []string{"passing the same captured per-item `correlation_id`"}},
				{name: "finalization forwards the item's captured id", all: []string{"the item's captured `correlation_id` from Section 2.3.4 passed as the `correlation_id` argument"}},
				{name: "finalization retry forwards the same captured per-item id", all: []string{"same provenance, same captured per-item `correlation_id`, no changes"}},
				{name: "not-found with the forwarded id is permanent", all: []string{"consolidation carrying the forwarded `correlation_id` for a dispatch that executed runs", "with the ID forwarded, `not_found` means no persisted dispatch session exists", "a plugin-process or tool-session split can no longer produce it", "a `not_found` lacking the forwarded `correlation_id` indicates the capture step was skipped"}},
			},
		},
		{
			name: "uf.unleash.md",
			clauses: []commandClause{
				{name: "dispatch-step capture and per-dispatch forwarding", all: []string{"capture the `correlation_id` reported in that one-line summary", "pass it as `correlation_id` on every subsequent consolidation call for this dispatch", "forward it unconditionally, never conditioned on in-process session-map state"}},
				{name: "per-dispatch-site scoping", all: []string{"Step 4 and Step 6 are separate dispatches with separate `correlation_id` values", "each dispatch's ID threads only to its own consolidation calls"}},
				{name: "finalization forwards the captured id", all: []string{"and the captured `correlation_id` from this dispatch"}},
				{name: "re-dispatch captures and forwards a fresh id", all: []string{"Each re-dispatch reports a fresh `correlation_id` in its one-line counts summary", "capture it and pass it on every `consolidate_dispatch` call for that re-dispatch"}},
			},
		},
	}

	for _, command := range commands {
		command := command
		t.Run(command.name, func(t *testing.T) {
			assertCommandCopies(t, command.name, true, command.clauses)
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

func TestCommandContracts_RejectStaleManualFinalizationLanguage(t *testing.T) {
	// The three rewritten commands must not reintroduce the pre-change
	// manual closed-payload assembly instructions alongside the one-call
	// consolidate_dispatch path. Presence-only clauses cannot detect the
	// contradiction, so the removed instruction phrases are asserted
	// absent. speckit.testreview.md legitimately retains direct
	// finalize_review_dispatch usage and is excluded.
	stalePhrases := []string{
		"Call `finalize_review_dispatch` for every iteration",
		"Call `finalize_review_dispatch` for every invocation",
		"Call `finalize_review_dispatch` once for every Tier 2 item",
		"version 1 payload and provenance",
	}

	commands := []struct {
		name       string
		scaffolded bool
	}{
		{name: "uf.review-council.md", scaffolded: true},
		{name: "uf.triage-issue.md", scaffolded: true},
		{name: "uf.address-feedback.md", scaffolded: true},
	}

	for _, command := range commands {
		command := command
		t.Run(command.name, func(t *testing.T) {
			for _, copy := range loadCommandCopies(t, command.name, command.scaffolded) {
				copy := copy
				normalized := normalizeCommandText(copy.text)
				for _, phrase := range stalePhrases {
					if strings.Contains(normalized, normalizeCommandText(phrase)) {
						t.Errorf("%s: stale manual finalization language found: %q", copy.name, phrase)
					}
				}
			}
		})
	}
}

func TestCommandContracts_RejectIncompatibleGhSlurpFlag(t *testing.T) {
	// FR-006: `gh api --slurp` is not supported by the installed gh CLI
	// (production incident ses_eca4843f9ffeKwVLQhSFCRh2Rl). The rewritten
	// commands must use the portable `--jq` projection form instead, and
	// a command edit or scaffold refresh must not silently reintroduce
	// the broken flag with no failing check.
	commands := []struct {
		name       string
		scaffolded bool
	}{
		{name: "uf.review-council.md", scaffolded: true},
		{name: "uf.triage-issue.md", scaffolded: true},
		{name: "uf.address-feedback.md", scaffolded: true},
	}

	for _, command := range commands {
		command := command
		t.Run(command.name, func(t *testing.T) {
			for _, copy := range loadCommandCopies(t, command.name, command.scaffolded) {
				copy := copy
				normalized := normalizeCommandText(copy.text)
				if strings.Contains(normalized, "--slurp") {
					t.Errorf("%s: incompatible gh flag %q found; the portable --jq projection form is required", copy.name, "--slurp")
				}
			}
		})
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
