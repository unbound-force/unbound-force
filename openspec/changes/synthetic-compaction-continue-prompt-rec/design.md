## Context

The `/uf.unleash` pipeline runs unattended (dewey#153) and delegates
work to child agents. After every compaction opencode injects a
synthetic user message "Continue if you have next steps, or stop
and ask for clarification if you are unsure how to proceed." That
message is not human input, but nothing in the durable
checkpoint/gate layer distinguishes it from human input.

In the incident (dewey#193, child ses_e200afc75ffeSuftvVXtQnBJhR,
2026-11-15, 67 compactions in ~2h) the child recorded:

```
checkpoint {command: uf.unleash, step: checklist-override-phase8-9,
 data: {checklist_override: "User explicitly directed continuation
 via 'Continue if you have next steps' with spec-review APPROVE WITH
 ADVISORIES (ESCALATED, correlation 8ef4f276-…)…"}}
```

`checkpoint` accepted a free-form step name carrying a gate override
with a free-text justification quoting the synthetic prompt. No
question-tool answer or orchestrator relay backed the override. The
cited spec-review (APPROVE WITH ADVISORIES, ESCALATED, 16 findings,
2 proposals) was never relayed to a human either.

Current state references:
- #1094 (HARDEN-001..006) persists spec-review consolidation and
  enforces per-finding decisions for REQUEST CHANGES, but not for
  checklist/quality gates or for synthetic-prompt quoting.
- #1100 (DT-FR-002, DELEG-FR-002) stops synthetic final text being
  reported as success and detects compaction storms, but does not
  stop the model quoting the prompt mid-run as consent.
- Constitution v1.4.0: VI Deterministic Execution — gate integrity
  MUST be tool-enforced.

This change makes gate overrides require a verifiable human-decision
reference and makes the synthetic prompt mechanically non-consent.

## Goals / Non-Goals

### Goals

- Make every gate override (checklist, spec-review advisory
  escalation, quality gate) require a human-decision reference
  (question-tool answer or orchestrator relay) and reject free-text
  justifications alone.
- Make the synthetic compaction prompt mechanically non-consent
  through command/agent prose pinned by a contract test on both the
  live and scaffold copies.
- Reject checkpoint step names outside the command's declared set,
  or at least those matching `*override*`, unless the human-decision
  reference is present.
- Relay ESCALATED APPROVE WITH ADVISORIES via
  `relay_spec_review_gate` before implementation continues.
- Provide a deterministic regression replaying the incident JSON.

### Non-Goals

- Fixing the compaction storm itself (#1100) or tuning compaction
  summarization.
- Changing review policy (which findings are raised, which severities
  block) — only the gate mechanics are hardened.
- New human interaction outside the existing gate flow.

## Decisions

### Decision 1: How to represent the human-decision reference

**Options:**
(a) Require a `question`-tool correlation / delegation relay id
    carried in the `checkpoint` payload (e.g. `human_decision_ref`).
    The tool validates the reference exists in session context or
    in the delegation input's relayed decisions.
(b) Infer authorization from free-text heuristics (e.g. scan
    `checklist_override` for "User explicitly directed").

Decision: (a). Free-text heuristics are exactly the bypass vector.
A structured reference ties authorization to a verifiable
artifact (question-tool answer record or orchestrator relay).
The checkpoint tool can validate the reference mechanically
without LLM judgment. Heuristic scanning would be re-bypassable
by rewording the justification.

The shape of `human_decision_ref` lives in
`.opencode/lib/uf-workflow-types.ts` alongside the existing
spec-review decision types so all producers/consumers share one
normative definition.

### Decision 2: Where to enforce — inside `checkpoint`

`checkpoint` is already the durable progress gate for
`uf.unleash` (HARDEN-002). Adding the human-decision and
declared-step checks there keeps gate enforcement in one
tool-enforced location. Alternatives (separate pre-check tool,
orchestrator-only relay) would leave the direct child path
unprotected. The check is additive: non-override steps and
APPROVE paths are unaffected.

Implementation runtime is the TypeScript plugin suite:
`checkpoint` / `get_progress` / `relay_spec_review_gate` in
`.opencode/plugins/uf-workflow/index.ts` with shared types in
`.opencode/lib/uf-workflow-types.ts`. No Go code implements
these tools — Go tests only prose-pin command markdown.

### Decision 3: Declared-step enforcement scope

Two levels were considered:
- Strict: reject any step name not in the command's declared
  step set for `uf.unleash` (clarify, plan, tasks, spec-review,
  implement, code-review, retrospective, demo, plus
  implement-phase-N).
- Narrow: reject only steps matching `*override*` without a
  human-decision reference.

Decision: implement narrow as the MUST (`*override*` pattern)
with strict available as a follow-up. Narrow closes the
incident path immediately (the incident step was
`checklist-override-phase8-9`) without blocking legitimate
future step names that may be added for non-gate purposes.
The strict list is documented and the code is structured so
promoting to strict is a one-line change to the predicate.

### Decision 4: Prose placement and contract test

Command prose (`uf.unleash.md` and `uf.run-pipeline-plan.md`)
and the autonomy agent already describe session-resume and
human-input handling. The new paragraph is added near those
sections so it is visible at the decision point and survives
future edits to other sections. The contract test (Go or TS,
matching the style of `TestEmbeddedAssets_MatchSource`) asserts
the paragraph exists verbatim in both the live copy
(`.opencode/…`) and the scaffold mirror
(`internal/scaffold/assets/…`), failing if either copy drifts.

### Decision 5: ESCALATED relay reuse

Issue #1118 asks that APPROVE WITH ADVISORIES marked ESCALATED
be relayed. #1094 already added `relay_spec_review_gate` for
REQUEST CHANGES. This change reuses that tool: the spec-review
step's post-dispatch gate checks both REQUEST CHANGES and
ESCALATED APPROVE WITH ADVISORIES and calls the same relay
path. No new tool is added; the existing tool's gate predicate
is widened and tested.

## Risks / Trade-offs

- **Rejected legitimate override with no human in the loop.**
  Mitigation: unattended runs (dewey#153) already use the
  orchestrator relay path — the human-decision reference can be
  carried in delegation input, so unattended advancement is
  possible when the orchestrator has relayed the decision.
  Pure free-text overrides with no human input remain rejected,
  which is the intended security property.

- **Strict step list maintenance cost.** If the strict predicate
  is adopted later, new legitimate steps must be added to the
  allowlist. The narrow `*override*` predicate avoids this cost
  now; the allowlist is still enumerated in the spec as the
  intended strict set for future hardening.

- **Scaffold mirror drift.** Any prose change must be mirrored
  byte-identically or `TestEmbeddedAssets_MatchSource` fails.
  The contract test for the new prose catches drift in both
  directions.
