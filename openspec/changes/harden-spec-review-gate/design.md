## Context

The spec-review gate is the only human-authorization point before
implementation. Issue #1078 proved it is bypassable when three
properties fail simultaneously:

1. `consolidate_dispatch` returns REQUEST CHANGES but nothing
   persists before the model resumes editing — so edits to
   `specs/git-trust-source/spec.md` can proceed without a gate.
2. `checkpoint` does not enforce per-finding decisions — so the
   child can mark progress as if gate passage occurred.
3. `get_progress` can return stale `current_step: spec-review`
   from iteration 2 even though iteration 3 already consolidated
   as REQUEST CHANGES — so the orchestrator re-dispatches
   iteration 3 and gets APPROVE under the same iteration number,
   dropping 7 findings.
4. The orchestrator relays from child text — so a compaction
   summary (the companion compaction-loop issue) loses all 8
   findings and the orchestrator records `overrides: []`.

Current state references:

- Constitution v1.4.0: VI Deterministic Execution — gate
  integrity MUST be tool-enforced, not model-prose-dependent.
- Proposal: `proposal.md` in this change — constitution
  alignment PASS for I (Autonomous Collaboration), II
  (Composability First), III (Observable Quality), IV
  (Testability), V (Security by Default), VI (Deterministic
  Execution).

The companion compaction-loop issue is the trigger, but out of
scope here — this change makes the gate hold regardless.

## Goals / Non-Goals

### Goals

- Make spec-review findings impossible to lose or silently
  self-fix via model prose.
- Enforce per-finding accountability: every finding gets `apply`
  (with a fix reference), `override` (with reason), or `pause`
  before `spec-review` advances.
- Define the post-decision control flow: `apply` decisions
  require a subsequent iteration that consolidates to APPROVE
  before advancement; an all-override decision set advances with
  the overrides recorded.
- Preserve correlation id + iteration lineage so re-dispatch
  cannot re-roll the same iteration to APPROVE.
- Make orchestrator relay deterministic and unconditional:
  `relay_spec_review_gate` reads progress state + dispatch
  artifact, computes the pause state, and renders without an
  LLM.
- Make the resume probe honor recorded adjudication so a
  human-closed gate is not re-dispatched on resume.
- Add regression fixtures that replay the exact bypass path
  (`ses_e254aedb8ffeqlLkdOgSpKNj60` → 3 spec edits → compaction
  return).

Policy note: this change does not alter which findings a review
raises or which severities block — that review policy is
unchanged. It does settle, as gate mechanics, what each recorded
decision produces: `apply` → a subsequent APPROVE iteration;
`override` → auditable advancement; `pause` → a held gate.

### Non-Goals

- Fixing the compaction-loop that turned the child's gate report
  into a summary — tracked separately. This change only makes the
  gate immune to that class of malformed return.
- Adding new human interaction outside the existing gate flow —
  the gate UX stays "human picks apply/override/pause per
  finding".

## Decisions

### Decision 1: Where to persist — inside `consolidate_dispatch`

**Rationale:** `consolidate_dispatch` already owns the dispatch
artifact (`.uf/artifacts/dispatch/<correlation>.json` with
verdict, findings, correlation id). The missing link is the
*progress-state* entry that `get_progress` reads. The
implementation runtime is the TypeScript plugin suite:
`dispatch_review_runs` / `consolidate_dispatch` live in
`.opencode/plugins/review-dispatch/index.ts`, `checkpoint` /
`get_progress` in `.opencode/plugins/uf-workflow/index.ts`, and
the progress-state record shape is shared through
`.opencode/lib/uf-workflow-types.ts`. Two options were considered:

- (a) Make `consolidate_dispatch` (review-dispatch plugin) write
  the progress-state record in the same finalization step that
  writes the dispatch artifact — one tool, two file writes before
  returning.
- (b) Keep the dispatch artifact write in `consolidate_dispatch`,
  add a mandatory `checkpoint` call immediately after that syncs
  to progress state; make the "record before resume" invariant
  hold via the command template plus `get_progress`
  reconciliation.

Decision: (a). Both stores are files under `.uf/` in the same
workspace and the shared types lib single-sources the record
shape, so the cross-plugin write is safe — and it closes the
window #1078 exploited rather than narrowing it. (b) remains the
documented fallback if the cross-plugin write proves unsafe; its
template ordering is already an enforcement mechanism for other
gates. Both satisfy I (Autonomous Collaboration — artifact +
state) and III (Observable Quality — correlation id in both
places).

**Trade-off:** (a) couples the two plugins through the shared
record shape in `uf-workflow-types.ts`; (b) leaves a between-tools
window. The coupling is acceptable: the lib exists for exactly
this shared-shape purpose.

### Decision 2: `checkpoint` as per-finding gate

**Rationale:** `checkpoint` is already the durable "advance
allowed?" decision point (Principle VI). Extending it to validate
`spec-review` advancement against the latest dispatch verdict is
minimal surface: if `verdict != APPROVE` and any finding lacks a
decision, refuse with `missing_per_finding_decisions:
[finding_ids]`; if decisions are complete but include `apply` and
no subsequent iteration consolidates to APPROVE, refuse with
`pending_approving_iteration`; an all-override decision set
advances with `overrides[]` recorded. This reuses the existing
`checkpoint` refusal pattern (already used for other phases) and
gives a single place to audit `overrides[]`.

**Alternatives rejected:**

- Client-side validation in `uf.unleash.md` prose — exactly what
  #1078 bypassed. Must be tool-enforced.
- New `gate_decision` tool — adds tool sprawl for a check
  naturally owned by the progress-state writer. A distinct tool
  would need its own persistence and would duplicate `checkpoint`
  state.

### Decision 3: Iteration re-dispatch guard

**Rationale:** Iterations are a monotonic log, not a retry slot.
Once iteration N has a recorded verdict, N is sealed. The
dispatch layer checks `progress_state.spec-review.iterations[N].verdict`
before allowing a new run for iteration N. If present, it MUST
return `iteration_already_recorded` naming the recorded iteration
and verdict. Auto-increment was rejected: a silent renumber masks
the mis-instruction instead of surfacing it. Iteration caps
(existing behavior) still apply; the cap check moves before
dispatch so re-rolling cannot bypass it.

**Edge:** Orchestrator prompts that say "re-dispatch iteration 3"
when iteration 3 is already recorded are treated as programmer
error and surfaced. The child should have been told "record
decisions for iteration 3" or "start iteration 4", not "re-run
3".

### Decision 4: Orchestrator relay via `relay_spec_review_gate`

**Rationale:** Child text is untrusted after compaction (or any
model deviation). The dispatch artifact is the source of truth:
it was written by `consolidate_dispatch` before the model could
intervene. On child return, `/uf.run-pipeline-plan` makes one
unconditional intent-level call: `relay_spec_review_gate` — a
uf-workflow plugin tool that reads progress state and
`.uf/artifacts/dispatch/<correlation>.json`, computes the pause
state (undecided non-APPROVE iteration → paused; decided,
approved, or no spec-review record → not paused), and renders
the gate relay (finding severities, descriptions, file:line,
correlation id) deterministically, without an LLM. The command
presents the rendered relay to the human when the tool returns
`gate_paused: true` and proceeds only when it returns
`gate_paused: false`. No prose-evaluated condition sits between
the child's return and the relay call — the pause decision is
a pure function of structured progress state, so it lives
inside the tool (DE-005); only the presentation to the human
remains judgment prose. Fixtures assert the tool's output
directly, so HARDEN-005's assertion (d) is observable without
an eval run.

Rendered-relay example (DE-002 — preserved near the
presentation point; the command presents this verbatim when
the tool returns `gate_paused: true`):

    ## Spec-review gate paused — iteration 3 REQUEST CHANGES
    Correlation 5355acba-62f4-4859-8775-7294b3e8702c, 8
    findings: 1 HIGH, 4 MEDIUM, 3 LOW. Per-finding decisions
    (apply | override | pause) required before spec-review
    may advance.
    1. HIGH deterministic-routing — tasks.md:88 …
    2. HIGH seam-verification — tasks.md (phase 5/6) …
    …

**Constitution refs:** I (artifacts are the collaboration
medium), III (artifact includes producer, correlation id,
timestamp), VI (deterministic render, not prose parse).

### Decision 5: Regression test shape

Fixtures replay the bypass path without needing real LLM calls:

- Mock dispatch store with correlation
  `5355acba-62f4-4859-8775-7294b3e8702c`, 8 findings, REQUEST
  CHANGES.
- Child simulation: 3 spec edits + compaction-summary return
  text.
- Assertions: (1) `get_progress` shows iteration 3 immediately
  after `consolidate_dispatch` returns — before any edit; (2)
  `checkpoint` advance refused without decisions, and refused
  after `apply` decisions until a subsequent APPROVE iteration;
  (3) `relay_spec_review_gate` renders 8 findings from the
  artifact; (4) no `overrides: []` auto-written; (5) re-dispatch
  of iteration 3 refused with `iteration_already_recorded`.

This satisfies IV (Testability — isolated, no external services)
and covers AC #5.

### Decision 6: Resume probe honors recorded adjudication

**Rationale:** `uf_unleash_detect_resume` currently derives
spec-review completion from approving dispatch artifacts,
reconciled with checkpoint state (artifacts authoritative). After
a human adjudicates an iteration — especially all-override, which
produces no approving artifact — an artifact-only probe reports
spec-review incomplete and the orchestrator re-dispatches a gate
the human already closed (prior learning
uf-unleash-resume-20261102T230750). The probe therefore honors
the recorded decision set for the latest iteration as that
iteration's adjudication: all-override → done; `apply` decisions
→ done only once the subsequent APPROVE iteration is recorded
(and the probe MUST NOT re-fail on the spec-artifact dirt the
applied fixes caused); undecided → incomplete, re-entering the
gate. This keeps artifacts authoritative for verdicts while
checkpoint data is authoritative for adjudication.

## Risks / Trade-offs

- **Double-write coupling (consolidate + progress state):** The
  dispatch artifact and progress state are separate files, so a
  brief window where the artifact exists but the state does not
  is possible. That window is bounded by enforced invariants
  rather than an edit gate: edits made before decisions are
  recorded cannot advance the gate (HARDEN-002), cannot re-roll
  the recorded iteration (HARDEN-003), and cannot hide findings
  (HARDEN-004 — the relay is artifact-sourced). Tests assert the
  state is visible immediately after `consolidate_dispatch`
  returns.
- **Per-finding UX friction:** Requiring a decision per finding
  (including 7 MEDIUM/LOW) adds human work. Mitigated: findings
  already require reading; bulk "apply all" with one commit ref
  is allowed as long as each finding is individually marked
  `apply` (no silent drop).
- **Append-only iteration log:** Iterations are sealed once
  recorded; growth is bounded by the existing iteration cap
  (~3–5). There is no auto-bump: a caller that needs a fresh
  review dispatches the next iteration explicitly.
- **Compaction-loop interaction:** This gate does not fix the
  underlying compaction behavior; it only prevents the gate from
  being bypassed when the child's text is malformed. A child that
  returns a correct gate report but with wrong finding text is
  still caught because relay is artifact-sourced.
- **Scaffold drift:** If `uf.unleash.md` or
  `uf.run-pipeline-plan.md` are changed, both canonical and
  scaffold copies (`internal/scaffold/assets/…` and
  `.opencode/…`) must be updated or
  `TestEmbeddedAssets_MatchSource` fails. Impact is CI-visible
  and intentional.
