## Context

`.opencode/commands/uf.triage-issue.md` is the canonical command spec
for issue triage (895 lines, `<protect>`-wrapped, scaffolded by
`uf vdev`). A byte-identical mirror lives at
`internal/scaffold/assets/opencode/commands/uf.triage-issue.md`;
`internal/scaffold/scaffold_test.go` drift tests enforce parity under
`make check`.

The command's Phase 2 and Phase 4 document dispatch data flows that do
not match the review-dispatch plugin
(`.opencode/plugins/review-dispatch/index.ts`, 4730 lines). Issue #735
records four gaps, observed live during triage of
fullsend-ai/fullsend#8465 (7 runs, 3 content-persona skips):

1. §2.3 claims `dispatch_review_runs` returns `status`, run counters,
   consolidated `findings`, `proposals`, and `warnings`. Actual
   caller-visible output is a one-line counts summary (the plugin's
   `createDispatchReviewRunsTool` at line 4202 returns
   `{ output: <summary line>, metadata: <full result> }`; only
   `output` reaches the calling model).
2. §2.4 consumes the `proposals` array that therefore never reaches
   the caller — and reads a `reference` field that does not exist
   (proposals are stored as `{information, tag, category}` at
   `index.ts:2973` and mapped to `{agent, information, tag,
   category}` in the consolidate payload at `:3239`).
3. Phase 3 (§3.1) requires per-persona consolidation of verdicts plus
   category/objectivity/reasoning/split but names no data source. The
   run-recorded `workflow_verdict` (kind-native, computed at
   consolidation as `run.status === "success" ?
   workflowVerdictValue(kind, runVerdict(run.findings)) : null`)
   appears only in the `consolidate_dispatch` payload's `runs`
   entries; the assessment contract fields exist only in child
   session transcripts (`PersistedRunData.text` holds child final
   text on disk but is dropped from the payload).
4. §4.6 permits `consolidate_dispatch` only as a finalization call
   ("exactly once per finalization attempt, with provenance as its
   only argument"), leaving the tool's retrieval mode (no provenance
   → raw payload) undocumented in the command.

Consequences observed: improvised retrieval via `list_sessions` +
`db_read_session` across all 7 child sessions; two disagreeing verdict
sources (7/7 child texts VALID vs run-recorded 5
NEEDS-CLARIFICATION / 2 VALID); precedence inferred from §4.6's
finalizer-authoritative clause rather than stated.

The identical false claims (items 1 and 2) also live in the sibling
commands `uf.review-council.md` (§4), `uf.address-feedback.md`
(§2.3.7), and `uf.review-pr.md` (§5a). Spec review of this change
(additional detail below) confirmed the sibling drift and the
field-name error, and flagged an unexplained `@opencode-ai/plugin`
1.4.10 → 1.19.0 bump riding in the working tree; the maintainer
resolved both by expanding this change's scope.

This change is docs-first with one extraction: correct the command
texts so the documented data flows match the tools, extract the
assessment-field reads and validation into the tested
`read_assessment_fields` plugin tool (D4, determinism pack DE-001),
and adopt the maintainer-approved dependency pin lockstep (D7). No
existing plugin tool code, schemas, or outputs change. The proposal's
constitution alignment (all six principles assessed) carries over:
defined interfaces over improvised coupling, additive tool extension,
traceable verdict provenance, existing drift tests plus the new tool's
tests plus `make plugin-test` as verification.

## Goals / Non-Goals

**Goals**:

- §2.3 documents the counts-only `dispatch_review_runs` output and
  adds a sanctioned retrieval step: one `consolidate_dispatch` call
  WITHOUT provenance immediately after dispatch completes, fetching
  the raw payload (runs with `workflow_verdict`, findings,
  advisories, run counts, top-level `proposals`), with documented
  failure behavior (transient failure retried exactly once then fail
  closed; permanent failure fail closed without retry; all-skip
  dispatch recorded as a valid outcome per the no-vote semantics).
- §2.4 sources proposals from that payload, using the proposal's
  `category` field (not a nonexistent `reference`) as the learning
  category.
- §3.1 names authoritative sources: run-recorded `workflow_verdict`
  governs verdict resolution; a single `read_assessment_fields` call
  supplies category/objectivity/reasoning/split as validated fields
  (D4); run-recorded wins on disagreement; missing or unparseable
  child text skips that persona's assessment fields without blocking
  verdict resolution.
- The `read_assessment_fields` tool ships in the review-dispatch
  plugin with unit and integration tests (tasks 2.1-2.4), owning the
  persisted-text reads and untrusted-input validation that DE-001
  forbids as command prose.
- §4.6 documents `consolidate_dispatch` dual mode and scopes the
  exactly-once bound to the provenance-bearing finalization call.
- The sibling Store Lesson Proposals sections receive the
  proposal-sourcing correction shaped per command (FR-005):
  review-council §4 and address-feedback §2.3.7 source from the
  retrieval payload with a dual-mode note added to their
  finalization sections (§5, §2.3.8); review-pr §5a adds the
  retrieval step after its `dispatch_review_runs` call while §5b's
  direct `finalize_review_dispatch` path is unchanged.
- The dependency pin update ships with its full lockstep set and
  validation evidence (D7), as a dedicated commit (task 4.5).
- Mirrors stay byte-identical; `make check` drift gates stay green.

**Non-Goals**:

- Changing the existing `dispatch_review_runs`, `dispatch_status`,
  or `consolidate_dispatch` tool outputs or schemas (shared with
  review-council, address-feedback, speckit-testreview). The new
  `read_assessment_fields` tool (D4) is additive.
- Adding assessment contract fields to the consolidate payload
  (payload schema change; follow-up — task 5.4 files the tracking
  issue).
- Auditing `speckit-testreview` for other dispatch-data-flow claims
  (verified: it does not consume the proposals array; follow-up —
  task 5.4 files the tracking issue).
- Any CI workflow, coverage, or severity-threshold change (the
  pin-enforcement test expectation update is the tracked exception
  required by D7).

## Decisions

### D1: Docs-only fix (Option B), not tool-output change (Option A)

Making `dispatch_review_runs` return the arrays would change a shared
tool contract used by four commands and force plugin test churn and
review across unaffected consumers. Pointing the command at the
existing retrieval mode of `consolidate_dispatch` achieves the same
data flow with zero runtime change. The trade-off is one extra tool
call per invocation (the retrieval call), which is deterministic and
side-effect-free.

### D2: One retrieval call, placed in §2.3, serving §2.4 and Phase 3

The no-provenance `consolidate_dispatch` call happens immediately
after `dispatch_review_runs` completes, once per invocation. Its
payload serves three consumers: §2.4 proposals (top-level `proposals`
array of `{agent, information, tag, category}`), §3.1 verdict
resolution (per-run `workflow_verdict`), and Phase 4 finalization
context. A single call avoids repeated retrieval and keeps the
data-flow linear: dispatch → retrieve → store proposals → classify →
finalize. If the retrieval call fails transiently, the command
retries the side-effect-free call exactly once, then fails closed
(no fabricated data); permanent failures fail closed without retry,
and an all-skip dispatch (no session) is a valid outcome per the
no-vote semantics — documented as delta scenarios (FR-001).

### D3: Verdict precedence — run-recorded `workflow_verdict` governs

The kind-native `workflow_verdict` is tool-computed from
`runVerdict(run.findings)` via `workflowVerdictValue(kind, ...)`,
i.e., derived from the dispatch store's structured run records, not
from child prose. It matches what the finalizer later persists
(§4.6's authoritative-finalizer clause). The fullsend#8465 case
showed child prose can disagree (all VALID vs 5 NEEDS-CLARIFICATION).
The command will state explicitly: when child transcript text and
run-recorded `workflow_verdict` disagree, the run-recorded verdict
governs. This makes the majority-rule inputs deterministic.

### D4: Assessment fields via the `read_assessment_fields` tool
(DE-001 extraction)

The consolidate payload lacks category/objectivity/reasoning/split,
and the payload schema extension is a Non-Goal (follow-up, task
5.4). The first remediation of this change sanctioned bounded
transcript reads (`list_sessions` + `db_read_session`) with
per-value validation written as command prose; spec review flagged
that prose as a DE-001 violation — mechanical procedure prescribed
in a command where an executable tool is feasible, the same class
org precedent (persist-unleash-summary D1,
address-feedback-model-normalization) extracts to plugin tools. The
resolution is full extraction: a new review-dispatch plugin tool,
`read_assessment_fields`, owns the entire procedure as executable,
tested code. Given the dispatch's run identifiers, it reads each
run's persisted child final text from the dispatch store
(`PersistedRunData.text` — already persisted on disk, dropped from
the payload), parses the structured response contract fields, and
enforces the untrusted-input validation (Constitution V):
`category` a non-empty string of at most 128 characters;
`objectivity` at most 64; `reasoning` at most 4096; split
recommendations an array of at most 16 strings of at most 512
characters each; every value type-checked against the contract
shape; invalid values skipped as unavailable for that run.
Consequences: the command's Phase 3 calls one tool and prescribes no
mechanical procedure; no session-database reads remain in the
documented flow (the earlier transcript-read concession is
eliminated, not merely bounded); and when the payload schema
extension lands (follow-up, task 5.4), the tool's backend can
switch from persisted-text parsing to payload fields without any
command change. Tasks 2.1-2.4 implement the tool with unit and
integration tests; FR-003 references the tool and its bounds.

### D5: Dual-mode documentation in §4.6 with scoped exactly-once

§4.6 keeps its determinism guarantees intact: the finalization call
remains exactly-once per finalization attempt, provenance-only
arguments, never call `finalize_review_dispatch` directly, response
authoritative. The new text documents that the no-provenance
retrieval call from §2.3 is a distinct mode with no finalization side
effects and does not count against the exactly-once bound. This
removes the apparent prohibition that forced improvisation. The same
dual-mode note is carried into the sibling finalization sections it
affects (review-council §5, address-feedback §2.3.8) so the
retrieval call cannot be read as violating their exactly-once
clauses; review-pr §5b finalizes via direct `finalize_review_dispatch`
and is left unchanged, with the retrieval step added to §5a only.

### D6: Capability and format

Spec delta targets the `triage-issue` capability in nested modern
format (`specs/triage-issue/spec.md`), ADDED requirements for the
dispatch data-flow contract, following the dispatch-finalize-
determinism exemplar: FR-NNN prefixes (`FR-001`-`FR-007`), MUST
language, `#### Scenario:` blocks with `**GIVEN**/**WHEN**/**THEN**`
steps, and a Scaffold Asset Parity requirement modeled on its
FR-005. Failure-mode scenarios (transient and permanent retrieval
failure, all-skip dispatch, empty proposals array, missing and
aggregate-unavailable child text) are included per spec-review
feedback, not just happy paths.

### D7: Dependency pin lockstep (maintainer-approved)

The working tree's `@opencode-ai/plugin` 1.4.10 → 1.19.0 bump is
adopted rather than reverted (maintainer decision, 2026-10-19).
Rationale: 1.4.10's tool-argument handling silently strips runtime
arguments (worked around locally by `fix-tool-args-schema-mismatch`;
live `CHECKPOINT_ERROR` failures documented in its retrospective),
and 1.19.0 tracks the current SDK line. The pin is governance-gated,
so the change ships the complete lockstep set in the same PR:
scaffold asset mirrors (`internal/scaffold/assets/opencode/
package.json`), the
`TestPluginPackageInputs_ExactVersions`
expectedRuntime map, the pin-governance testdata
(`plugin-package-provenance.json` updated to 1.19.0 with fresh
`verified_at` and recorded upstream-verification evidence), and the AGENTS.md Technology
Stack and README prose — with upstream verification (npm registry
integrity and each package's actual license — MIT for
`@opencode-ai/plugin` and `@opencode-ai/sdk`; Apache-2.0 for
`@ai-sdk/provider`; AFL-2.1 OR BSD-3-Clause for `json-schema` —
plus a msgpackr 1 → 2 changelog review, following org precedent
`review-council-multi-model-fanout` task 1.5), a locally executed
Node engine-range check (the manifests declare ini 7.0.0's exact
range "^22.22.2 || ^24.15.0 || >=26.0.0", recorded as reproducible
evidence rather than deferred to CI: the declaration is verified
against ini 7's own `engines` field; the resolved local Node is
v24.11.1, which predates the >=24.15.0 bound on the 24.x line, so
npm reports an advisory EBADENGINE warning rather than a failure —
npm enforces engines advisively, not strictly) declared in the
`engines` fields of the manifests and their scaffold mirrors
(task 4.4), and `make
plugin-test` (33 files, 903 tests, coverage 94.55% statements /
86.98% branches) plus the pin-enforcement test
(`TestPluginPackageInputs_ExactVersions`) as
validation evidence, both green against 1.19.0 on 2026-10-19,
and a single-commit rollback path. The lockstep file group ships as
a dedicated commit staged by explicit path, separate from the
command-doc and plugin-tool commits (task 4.5), per the AGENTS.md
commit-scope rule. SC-FR-003 (unarchived
`review-council-multi-model-fanout` delta) recorded the 1.4.10 pin
for that change's verification; this change supersedes the
repo-wide pin, the live enforcement is updated in lockstep
rather than left to drift, and task 5.5 adds the superseded-by
annotation to the affected delta.

## Risks / Trade-offs

- **Retrieval call adds a step**: deterministic, side-effect-free,
  one call. Acceptable vs changing shared tool output.
- **New plugin tool**: `read_assessment_fields` is additive (no
  existing tool contract changes) and covered by unit and
  integration tests (tasks 2.2-2.3) under `make plugin-test`. Its
  persisted-text parsing depends on the child structured response
  contract; unparseable text degrades to unavailable fields rather
  than failure (FR-003 scenarios).
- **Docs drift re-introduction**: mitigated by the existing
  byte-identical mirror drift tests under `make check`; the task list
  includes the copy + `cmp` verification steps.
- **Dependency pin update**: governance-gated change adopted by
  maintainer decision; mitigations are the lockstep set, validation
  evidence (tasks 3.3, 4.1, 4.3), rollback path, and transitive
  risk notes in the proposal. The phase-boundary concern raised in
  review is resolved by shipping the lockstep file group as a
  dedicated commit (task 4.5).
- **`openspec validate` strict failures**: the exemplar change also
  fails strict validation (validator quirk with ADDED requirements);
  the repo's real gate is the FR-NNN/MUST/Scenario convention, which
  this delta follows.

## Migration Plan

None for the command-doc edits. The dependency pin update requires
`node_modules` refresh (`npm ci`) in environments that run the
plugins locally; CI installs from the lockfile.

## Open Questions

None blocking. Follow-up candidates (out of scope, noted in Impact;
task 5.4 files the tracking issues and records their numbers in
this change): payload schema extension for assessment fields —
tracked as #755 (when it lands, `read_assessment_fields` can read
the payload instead of persisted text without command changes); a
dispatch-data-flow audit of `speckit-testreview` — tracked as #756
(it does not consume the proposals array); the recovery-path fix
for failed-run submissions already tracked in
unbound-force/unbound-force#751.
