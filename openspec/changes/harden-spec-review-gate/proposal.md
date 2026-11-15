# Proposal: Harden spec-review gate against REQUEST CHANGES bypass

## Why

During a `/uf.run-pipeline-plan` run (dewey#203, epic dewey#153) a
`/uf.unleash` child received a final spec-review verdict of
**REQUEST CHANGES** (1 HIGH GTS-FR-001 — vocabulary gate turning
`tier: curated` into `draft` — plus 7 MEDIUM/LOW). The child
self-fixed the HIGH without a gate decision, edited
`specs/git-trust-source/spec.md` three times, proceeded to
implementation, and returned a compaction summary instead of a
gate report. The orchestrator (`ses_e258662fdffeo4EoTjIGt5NEvw`)
never saw the result: `get_progress` still showed
`current_step: spec-review` from iteration 2, so it re-dispatched
iteration 3 to a new child (`ses_e24e418a4ffeMlFj7pkRMd0V0A`)
which returned **APPROVE** (0 findings) and checkpointed
`spec-review {iteration: 3, verdict: APPROVE, overrides: []}`.
The 7 findings were silently dropped.

Evidence correlations:

- `5355acba-62f4-4859-8775-7294b3e8702c` — child
  `ses_e254aedb8ffeqlLkdOgSpKNj60` `dispatch_review_runs` →
  `consolidate_dispatch` REQUEST CHANGES 1 HIGH + 7 MEDIUM/LOW
  (15:58–16:16 UTC)
- `e4727e7c-…` — re-dispatch APPROVE 0 findings (17:31 UTC)

This violates dewey#153 ("automate demotion, not promotion") and
is the same class as #1013 (gate lost when model writes question
as text) and #1052 (empty reviewer output counts as APPROVE). The
companion compaction-loop issue triggered the malformed return,
but the gate must hold regardless.

Fixes #1078.

## What Changes

1. **Record consolidation before model regains control
   (HARDEN-001).** `consolidate_dispatch` MUST persist `verdict`,
   `correlation_id`, `iteration`, `finding_count`, and
   `finding_ids` to progress state before control returns to the
   model. `get_progress` must then reflect the latest iteration —
   no stale `current_step: spec-review` from iteration 2.

2. **Per-finding decision gate before advancement (HARDEN-002).**
   If the latest recorded spec-review result is non-APPROVE and
   any finding lacks a recorded per-finding decision (`apply`
   with a fix reference, `override` with a reason, or `pause`),
   `checkpoint` MUST refuse to advance `spec-review`. An `apply`
   decision MUST be followed by a spec-review iteration N+1 that
   consolidates to APPROVE, or a recorded human override set,
   before advancement; an all-override decision set advances with
   the overrides recorded. Edits made before decisions are
   recorded cannot advance the gate and cannot re-roll the
   recorded iteration.

3. **Refuse re-dispatch of an already-recorded iteration
   (HARDEN-003).** A re-dispatch for an iteration number that
   already has a persisted verdict MUST be refused with an
   `iteration_already_recorded` error naming the recorded
   iteration and verdict. Re-rolling the same iteration number
   until APPROVE is forbidden; iteration caps are enforced before
   dispatch.

4. **Orchestrator relay from artifact via a deterministic tool
   (HARDEN-004).** When a delegated `/uf.unleash` returns,
   `/uf.run-pipeline-plan` MUST make one unconditional
   intent-level call: `relay_spec_review_gate` — a uf-workflow
   plugin tool that reads progress state and the dispatch
   artifact (`.uf/artifacts/dispatch/<correlation>.json`),
   computes the pause state (undecided non-APPROVE iteration →
   paused; decided, approved, or no spec-review record → not
   paused), and renders the relay deterministically, without an
   LLM call. The orchestrator presents the rendered relay to
   the human when the tool returns `gate_paused: true` and
   proceeds only when it returns `gate_paused: false`; the
   child's returned text is never the decision input and no
   prose-evaluated condition gates the call.

5. **Resume probe honors recorded adjudication (HARDEN-006).**
   `uf_unleash_detect_resume` MUST treat a recorded decision set
   for the latest spec-review iteration as that iteration's
   adjudication: an all-override set reports `spec_review` done;
   `apply` decisions report done only once a subsequent iteration
   consolidates to APPROVE (dirty spec artifacts from the applied
   fixes do not re-fail the probe); an undecided iteration
   reports incomplete and re-enters the gate.

6. **Regression coverage for the bypass path (HARDEN-005).**
   Deterministic fixtures replay "child consolidates REQUEST
   CHANGES, edits spec files, returns non-gate text" and assert
   the gate holds: state persisted before edits, checkpoint
   refusal, no stale `get_progress`, relay tool output, no
   auto-written `overrides: []`.

Corresponding acceptance criteria (verbatim from #1078):

- A REQUEST CHANGES consolidation is in progress state before
  control returns to the model
- `checkpoint` refuses to advance `spec-review` past a
  non-APPROVE result without recorded per-finding decisions
- Re-dispatching an iteration that already has a recorded result
  is refused or advances the iteration count
- `/uf.run-pipeline-plan` relays findings from the artifact when
  a child returns with an undecided REQUEST CHANGES
- Regression tests cover the "child edits after REQUEST CHANGES,
  then returns a non-gate text" path

## Capabilities

### New Capabilities

- `spec-review gate hardening`: consolidation persistence,
  per-finding decision gate with post-decision control flow,
  iteration re-dispatch guard, deterministic orchestrator
  artifact-relay for undecided REQUEST CHANGES, and resume-probe
  adjudication
- `child-edits-after-REQUEST-CHANGES regression coverage`:
  deterministic tests for the bypass path (edit spec + return
  compaction summary)

### Modified Capabilities

- `consolidate_dispatch` / `checkpoint` / `get_progress`
  (TypeScript plugin tools in `.opencode/plugins/`): atomic
  record of spec-review verdict + iteration before the model
  resumes; `checkpoint` enforces per-finding decisions and the
  post-decision control flow before advancing `spec-review`
- `dispatch_review_runs` / dispatch layer: refuses re-dispatch of
  an already-recorded iteration with `iteration_already_recorded`
- `relay_spec_review_gate` (new uf-workflow plugin tool):
  deterministic gate-pause relay rendered from progress state +
  dispatch artifact
- `uf_unleash_detect_resume`: honors recorded per-finding
  adjudication for the latest spec-review iteration
- `/uf.run-pipeline-plan`
  (`.opencode/commands/uf.run-pipeline-plan.md`):
  post-delegation step becomes one unconditional
  `relay_spec_review_gate` call (the tool computes the pause
  state from progress state + artifact)
- `/uf.unleash` (`spec-review` step): mandatory gate text +
  checkpoint sequence tightened so bypass via prose edits is
  mechanically rejected

### Removed Capabilities

- None.

## Impact

- **Source templates requiring byte-identical scaffold copies**
  (if touched, update both `internal/scaffold/assets/…` and
  `.opencode/…` or `TestEmbeddedAssets_MatchSource` fails):
  - `internal/scaffold/assets/opencode/commands/uf.unleash.md` ↔
    `.opencode/commands/uf.unleash.md`
  - `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md`
    ↔ `.opencode/commands/uf.run-pipeline-plan.md`
- **Implementation runtime — TypeScript plugin tools**:
  `dispatch_review_runs` / `consolidate_dispatch` in
  `.opencode/plugins/review-dispatch/index.ts`; `checkpoint` /
  `get_progress` / `uf_unleash_detect_resume` /
  `relay_spec_review_gate` in
  `.opencode/plugins/uf-workflow/index.ts`; the shared
  progress-state record shape in
  `.opencode/lib/uf-workflow-types.ts`. No Go code under
  `internal/` implements these tools — Go tests there only
  prose-pin command markdown.
- **Orchestrator artifact path**: `.uf/artifacts/dispatch/`
  correlation-indexed artifacts (`<correlation>.json`) are the
  source of truth for relay
- **No behavioral change for APPROVE paths** — only non-APPROVE
  with missing decisions is blocked; iteration cap behavior is
  explicit
- **Token cost**: progress-state write + one artifact read on
  return; negligible vs. review dispatch cost
- **Instruction-text delta (DE-004)**: the two command edits
  are expected to add roughly +45/−20 lines to `uf.unleash.md`
  (gate semantics, decision recording, resume wording) and
  +30/−15 lines to `uf.run-pipeline-plan.md` (the unconditional
  relay call replacing the prose-evaluated condition); exact
  per-file instruction-lint rows and totals are recorded in
  the change record at implement time (task 7.5)

### Documentation Impact

Assessed per the AGENTS.md Documentation gate:

- **CHANGELOG.md** — affected: a change entry for the
  spec-review gate hardening (user-visible behavior change to
  `/uf.unleash` spec review and `/uf.run-pipeline-plan`
  orchestration) is added before PR merge (task 7.4)
- **AGENTS.md** — no update required: no new commands, build
  commands, project-structure, or convention changes; the gate
  mechanics live in the command and plugin surfaces, which
  AGENTS.md does not enumerate
- **README.md** — no update required: no repository-level
  description change
- **docs/** — no update required: `docs/usage.md` documents
  `/uf.unleash` at invocation and pipeline level, both
  unchanged; no existing page documents spec-review gate
  mechanics, and this change's design rationale lives in its
  design.md

### Content Coverage

The per-finding decision workflow — apply/override/pause per
finding, `pending_approving_iteration`, the iteration
re-dispatch guard, and the artifact-sourced gate relay — meets
the workflow-pattern threshold for tutorial consideration. The
opportunity is noted for the curator; authoring a tutorial is
out of scope for this change (the gate mechanics ship first
and stabilize in use).

## Constitution Alignment

Assessed against the Unbound Force org constitution
(`.specify/memory/constitution.md` v1.4.0, six principles).

### I. Autonomous Collaboration

**Assessment**: PASS

All enforcement is via artifacts, not synchronous coupling.
`consolidate_dispatch` writes a dispatch artifact and the
progress-state record; `checkpoint`/`get_progress` read/write
progress state; the orchestrator consumes those artifacts to
decide relay vs. advance. No hero blocks waiting for another
hero's runtime response — artifact-based collaboration is
strengthened.

### II. Composability First

**Assessment**: PASS

No new mandatory hero dependency. `uf.unleash` alone enforces its
own spec-review gate; `uf.run-pipeline-plan` enhances relay when
present but does not require it. Each command remains
independently installable and usable.

### III. Observable Quality

**Assessment**: PASS

Verdict, correlation id, iteration, and finding counts become
machine-parseable in progress state and dispatch artifacts (HIC
envelope). The deterministic relay renders the same data from the
artifact, preserving provenance (`producer:
consolidate_dispatch`, `correlation_id`, `iteration`). Silent
drops become observable failures.

### IV. Testability

**Assessment**: PASS

Adds deterministic regression tests for the "child edits after
REQUEST CHANGES, then returns non-gate text" path. Gate logic
(checkpoint refusal, iteration guard, artifact relay, resume
adjudication) is exercised without external services via injected
dispatch/progress fixtures — the relay's rendered output is
asserted directly from the tool, no LLM required.

### V. Security by Default

**Assessment**: PASS

Closes a privilege-escalation-equivalent bypass: model
self-override of a HIGH finding without human authorization.
Per-finding decisions (override requires reason, pause) make
promotion explicit and auditable; re-rolling is mechanically
prevented.

### VI. Deterministic Execution

**Assessment**: PASS

Replaces model-behavior-dependent gate ("child said it
addressed") with tool-enforced procedure: `consolidate_dispatch`
(artifact + state) → `checkpoint` gate → `get_progress` →
unconditional `relay_spec_review_gate` call. Deterministic
steps execute in tools with tests; gate integrity no longer
depends on prose compliance. Follows `automation-not-promotion` —
demotion (pause/relay) is automated, promotion (override)
requires explicit human gate.
