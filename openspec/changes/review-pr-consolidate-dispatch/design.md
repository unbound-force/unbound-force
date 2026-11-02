## Context

`/uf.review-pr` is the post-PR review command: it resolves PR
metadata, runs pre-flight checks, dispatches the Divisor council
through `dispatch_review_runs`, and finalizes the dispatch. Its
finalization step (Sections 5, 5a, 5b of
`.opencode/commands/uf.review-pr.md`, mirrored at
`internal/scaffold/assets/opencode/commands/uf.review-pr.md`)
still follows the pre-migration shape:

- Section 5 instructs the parent to deduplicate successful run
  findings by normalized file plus root cause and apply
  cross-persona grouping and compound severity -- conflicting
  with the finalizer's exact run-finding to consolidated-finding
  correspondence check.
- Section 5a retrieves the dispatch payload through a separate
  no-provenance `consolidate_dispatch` call, solely to source
  lesson proposals and run records.
- Section 5b mandates calling `finalize_review_dispatch` with a
  hand-assembled closed payload: every plan entry, every run
  with complete provenance, deduplicated findings, advisories,
  run counts, and verdict.
- No bounded retry policy distinguishes mechanical rejections
  from persistence failures.

The `dispatch-finalize-determinism` change (0593caeb) migrated
`/uf.review-council`, `/uf.triage-issue`, and `/uf.address-feedback`
to a single provenance-bearing `consolidate_dispatch` call, but
left `/uf.review-pr` untouched. Issue #906 documents the
consequence in a production run on dewey PR #138 (session
`ses_e66b2decfffe6iTzZlb9ktkM5E`, dispatch
`0274aa47-53a2-4715-9dd0-4951382a6f9a`): 5 runs succeeded, 4
skipped, 15 findings; the parent rebuilt the ~50 KB payload
three times -- rejected once for `Unrecognized key:
"read_only"`, rejected once for `consolidated finding has
incorrect contributing run ids` after Section 5's grouping
merged the architect (vault.go:627) and guard (vault.go:625)
findings, and accepted only after the plugin's 15 consolidated
findings passed through unchanged.

A second, smaller defect: `uf.review-council.md` Section 3
(~L360) still carries the pre-migration "First deduplicate ...
cross-persona grouping" prose. Its Section 5 finalization is
already consolidate-with-provenance, so the prose does not
break finalization today, but an agent following Section 3
literally could reintroduce the same rejection class.

Constitution alignment is recorded in the proposal (all six
principles PASS); the decisions below note where each touches
Autonomous Collaboration, Composability First, Observable
Quality, or Deterministic Execution.

## Goals / Non-Goals

### Goals
- `/uf.review-pr` finalizes with exactly one
  `consolidate_dispatch` call per finalization attempt,
  carrying provenance `{branch, commit, workflow_id:
  "uf.review-pr"}` resolved from the frozen PR input context.
- The command never assembles, restates, or recomputes the
  finalization payload and never calls
  `finalize_review_dispatch` directly.
- Mechanical rejections get a bounded retry (at most two,
  identical call, no assessment-data changes); everything else
  fails closed to `INCONCLUSIVE`.
- Cross-persona grouping and compound severity survive for
  review-comment presentation only, never touching
  finalization data.
- Lesson proposals and warnings come from the same consolidate
  result.
- Scaffold mirrors stay byte-identical; content-gate tests pin
  the new contract and the absence of manual-finalize prose.
- `uf.review-council.md` Section 3 prose is explicitly
  presentation-only.

### Non-Goals
- Fixing `speckit.testreview.md` Sections 3.5/3.7 (same
  manual-finalize pattern; issue #756 tracks it as a separate
  change).
- Fixing host-source model provenance resolution in plugin
  code (issue #901; a separate cause not addressed by command
  edits).
- Changing plugin behavior: `consolidate_dispatch`,
  `finalize_review_dispatch`, verdict derivation, and artifact
  schemas are untouched.
- Changing verdict semantics, review-comment grouping rules,
  human-confirmation gates, or GitHub mutation flows.

## Decisions

### D1. Provenance from the frozen PR input context

`branch` and `commit` resolve only from the Step 2 frozen input
context (`head_ref`, `head_sha`), following the
`/uf.address-feedback` Section 2.3.8 pattern -- not the
`git branch --show-current` / `git rev-parse HEAD` pattern the
checkout-reviewing commands use. Rationale: the finalizer
validates the provenance commit against
`input_context.head_sha`, and `/uf.review-pr` reviews a remote
PR head that need not match the local checkout (the command
itself instructs using only the frozen SHAs for every
downstream step). A missing or invalid frozen field fails
closed with an explicit error naming the field. This is
Autonomous Collaboration: the self-describing artifact carries
provenance resolved from validated input, not retyped from a
mutable environment.

### D2. One call replaces Sections 5, 5a, and 5b

The single provenance-bearing `consolidate_dispatch` call
replaces all three sections. Section 5a's no-provenance
retrieval existed solely to source proposals and run records
before finalization; the finalization result carries the same
data (proposals, warnings, artifact path, canonical
review-verdict v2 projection). Dropping the retrieval halves
the tool calls and deletes the retrieval-failure handling
branch (transient/permanent/not-found classification) that
exists only for that call. The all-skip case (no dispatch
session exists because every plan entry was skipped) follows
the already-migrated review-council Section 5 wording: the call
reports it, the failure is non-mechanical, and the fail-closed
path applies. Observable Quality: one deterministic call whose
structured response is authoritative for every downstream
read.

Trade-off accepted: on the fail-closed path the parent no
longer holds a pre-finalization proposals snapshot. Mitigation:
proposals persist in the dispatch store, the failure report
names them as stranded, and fail-closed semantics already
forbid fabricating lessons from improvised data.

The same rewrite updates the Section 4c dispatch-capture prose
(~L446-460), which cross-references the deleted Section 5a/5b
paths and the no-provenance retrieval: the correlation-id
forwarding mandate and the proposals-arrival sentence are
rewritten to point at the single finalization call, removing
the stale-cross-reference defect class this change exists to
fix.

### D3. Grouping and compound severity become presentation-only

The incident's second rejection came from applying Section 5's
deduplication and cross-persona grouping to finalization data.
The rewrite keeps `severity.md` grouping and compound severity
for composing the GitHub review comment and the terminal
report, and states explicitly that they never merge, filter,
re-order, or otherwise alter findings, contributing run ids,
verdicts, run counts, or advisories in finalization data. The
verdict-ladder prose (blocking run yields REQUEST CHANGES, and
so on) is likewise removed from the finalization path: the
verdict arrives from the tool's finalization result
(vote-derived per the capability's FR-009), including the
zero-success results. Same treatment for `uf.review-council.md`
Section 3, whose stale prose is reworded to the same
presentation-only scope.

### D4. Bounded retry restated for the one-call path

The retry policy is the existing FR-004 contract applied
unchanged: classify from the structured response alone,
default-deny; mechanical means a validation-stage payload
rejection with enumerated defects, no persisted artifact, and a
null calculated assessment; retry the identical call (same
correlation id, same provenance) at most twice; never alter
assessment data; exhausted retries or non-mechanical failures
fail closed to `INCONCLUSIVE` with the calculated assessment
retained as non-authoritative human context. Because the
command no longer assembles a payload, "correct the shape and
retry" degenerates to "retry the identical call" -- the wording
already used by the migrated review-council Section 5.
Deterministic Execution: the procedure is identical across
attempts.

### D5. Content gates live in scaffold_test.go

The change is command-prose only; plugin behavior is already
tested. The new assertions follow the existing content-gate
precedent in `internal/scaffold/scaffold_test.go`
(`TestQuestionToolPresentationContractGates`,
`TestSpecReviewPerFindingGateContracts`), which reads the
canonical `.opencode/commands/` sources and their embedded
mirrors through one read path. New assertions cover: presence
of the one-call contract, frozen-context provenance,
never-assemble/never-call-finalizer prohibitions, bounded
retry, presentation-only grouping; and absence of
manual-finalize prose (no `finalize_review_dispatch` call
instruction, no finalization-path deduplication mandate).
The pinned-clause contract test
`TestCommandContracts_CorrelationIdForwarding`
(`internal/scaffold/command_contract_test.go`) asserts
Section 5a/5b strings over the live `uf.review-pr.md` and its
mirror; its `uf.review-pr.md` clause set is re-pinned to the
new one-call contract in the same change.
Byte-identical mirror parity remains enforced by the existing
`TestEmbeddedAssets_MatchSource` drift tests. Testability:
plain string assertions over repository files, no external
services, run under `make check`.

### D6. Stale site labels in golden-output-shape.test.ts are
refreshed

`.opencode/test/golden-output-shape.test.ts` documents the
plugin tool-call sites by command file and line (for example,
`finalize_review_dispatch x2 (uf.review-pr.md:543,
speckit.testreview.md:278)`). These are labels and coverage
comments, not file reads, so the suite still passes after the
rewrite -- but leaving them stale would claim a
`finalize_review_dispatch` documentation site in
`uf.review-pr.md` that no longer exists. The implementation
refreshes the inventory comments and the `it.each` site labels
-- removing the stale `finalize_review_dispatch` attribution,
documenting the new `uf.review-pr.md` site of the finalization
`consolidate_dispatch` call, and updating the per-plugin site
counts (keeping the underlying plugin-tool coverage via the
remaining documented site), preserving honest provenance
documentation.

### D7. Change metadata threads the originating issue

`.openspec.yaml` records `originating_issue: 906` so
`/uf.finale` emits `Closes #906` in the PR body, following the
`address-feedback-model-normalization` precedent. The proposal
and tasks reference #756 and #901 as explicitly out of scope so
reviewers do not expect them in this change.

## Risks / Trade-offs

- **Presentation grouping drift.** Future command edits could
  reintroduce finalization-path grouping. Mitigated by FR-015's
  explicit presentation-only statement plus the content-gate
  absence assertions (D5).
- **Lessons stored later.** Proposal storage moves from
  post-dispatch (old Section 5a) to post-finalization. On the
  fail-closed path lessons are not stored. Accepted per D2:
  stranded proposals remain recoverable from the persisted
  dispatch store, and fabricating lessons is already forbidden.
- **Line-pinned test labels.** The rewrite shifts line numbers
  cited in test comments and site labels; these are refreshed
  (D6) but remain line-sensitive documentation. The drift and
  content-gate tests themselves are string-based and stable.
- **Four-command consistency.** `/uf.review-pr` now follows
  `/uf.address-feedback`'s frozen-context provenance while
  `speckit.testreview.md` still carries the old pattern. The
  divergence is temporary and tracked by #756; this change
  does not widen it.
