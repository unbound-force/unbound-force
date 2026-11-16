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
  that resolves to a persisted human-decision record, and reject
  free-text justifications alone.
- Make the synthetic compaction prompt mechanically non-consent
  through command/agent prose pinned by a contract test on both the
  live and scaffold copies.
- Reject checkpoint step names outside the command's declared
  set unless the human-decision reference is present.
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

(a) Require a `human_decision_ref` carried in the `checkpoint`
    payload, validated against the persisted human-decision
    record (`.uf/progress/<key>/human-decisions.json`). Entries
    reach the record through exactly two surfaces: the
    orchestrator relay — `uf_delegate_command` persists entries
    from its typed `human_decisions` envelope parameter at spawn
    time, a machinery write the delegated child never authors —
    and the in-session question path — the
    `record_human_decision` tool records a `question`-tool
    answer the session received, with the SYNTH-006 controls
    (machine-stamped `decision_id`/`recorded_at`, append-only,
    synthetic-prompt deny-list). `checkpoint` accepts references
    only; no inline self-authorizing payload.
(b) Infer authorization from free-text heuristics (e.g. scan
    `checklist_override` for "User explicitly directed").

Decision: (a). Free-text heuristics are exactly the bypass vector.
A structured reference ties authorization to a verifiable
artifact (a file-backed human-decision record entry), and the
checkpoint tool can validate the reference mechanically
without LLM judgment because it reads files, not session
context — plugin tool factories receive only `(exec, cwd)` with
no session access (DE-001: a tool-enforced gate is only as
mechanical as its readable input, so the spec names the record
the tool reads). Heuristic scanning would be re-bypassable
by rewording the justification; an unverifiable reference
string would be the same bypass in structured clothing, which
is why the reference MUST resolve to a recorded entry and why
the orchestrator-relay surface is a machinery write rather
than a child-callable free-form append.

The shape of `human_decision_ref` and the human-decision record
entry (`decision_id`, source, question, answer, `recorded_at`)
lives in `.opencode/lib/uf-workflow-types.ts` alongside the
existing spec-review decision types so all producers/consumers
share one normative definition.

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

Decision: implement strict as the MUST, keyed on the call's
`command` argument resolving to `uf.unleash` — the same command
keying the HARDEN-002 spec-review gate uses — because
`checkpoint` is a shared tool: `uf.fix-ci`, `uf.review-council`,
`uf.address-feedback`, and `uf.finale` each prescribe their own
step vocabularies under their own command identifiers, and a
globally enforced allowlist would reject every one of them,
including the review council /uf.unleash itself dispatches. The
spec-review finding on the first draft (conflated scopes in one
MUST) required a single deterministic predicate, and strict is
the one the originating issue's acceptance criterion names
("checkpoint rejects step names that aren't declared"). Strict
closes the incident path (`checklist-override-phase8-9` is
undeclared) and every future undeclared step name, gate-flavored
or not; the `*override*` pattern is a documented consequence
(such names are never in the declared set), not an alternative
predicate. The allowlist is the union of both writer surfaces —
the pipeline's user-facing steps (clarify, plan, tasks,
spec-review, implement, implement-phase-N, code-review,
retrospective, demo), the machinery steps the uf-workflow
plugin auto-checkpoints under command `uf.unleash`
(branch-safety-gate, tier-routing, resumability-detection), and
the gate and budget steps the `uf.unleash` command prose records
(tier-routing re-key, quick-pre-flight, quick-fix-budget) — and
a lockstep consistency test enumerates both surfaces scoped to
steps recorded under `uf.unleash` (the plugin's
`reply-comments` auto-checkpoint belongs to
`uf.address-feedback` and is correctly out of scope) so a newly
recorded step name missing from the allowlist fails the suite
instead of wedging the pipeline. The maintenance cost — new
legitimate steps must be added to the allowlist — is accepted
and recorded below.

### Decision 4: Prose placement and contract test

Command prose (`uf.unleash.md` and `uf.run-pipeline-plan.md`)
and the `cobalt-crush-dev` agent file
(`.opencode/agents/cobalt-crush-dev.md` — the incident child's
persona; no "autonomy agent" file exists in this repository)
are the three prose surfaces. The new paragraph is placed near
each file's existing session-resume, human-input, or
gatekeeping prose so it is visible at the decision point and
survives future edits to other sections. The contract test
(Go or TS, matching the style of
`TestEmbeddedAssets_MatchSource`) asserts the paragraph exists
verbatim in both the live copy and the scaffold mirror for all
three files, failing if either copy drifts.

### Decision 5: ESCALATED relay reuse

Issue #1118 asks that APPROVE WITH ADVISORIES marked ESCALATED
be relayed. The review-dispatch finalizer's canonical decision
maps every APPROVE WITH ADVISORIES consolidation to ESCALATED,
so the marker is not a distinguishing subset — the relay
predicate is "latest consolidated spec-review result is
APPROVE WITH ADVISORIES", full stop. #1094 already added
`relay_spec_review_gate` for REQUEST CHANGES. This change
reuses that tool: the spec-review step's post-dispatch gate
checks both REQUEST CHANGES and APPROVE WITH ADVISORIES and
calls the same relay path; a plain APPROVE (canonical decision
APPROVED) still advances without relay. No new tool is added;
the existing tool's gate predicate is widened and tested.

### Threat model

The gate closes the incident class: a confused or
compaction-storming child quoting its context (the synthetic
continue prompt, free-text justifications, prior-session
citations) as authorization. It is not a cryptographic
boundary against a deliberately adversarial session: plugin
tool factories receive only `(exec, cwd)` with no caller
identity, and any session with exec access can corrupt local
state directly — the same residual risk every tool-enforced
gate in this suite accepts (HARDEN-002's `spec_review_decisions`
reason strings have the identical property). Within that
ceiling the design applies the strongest available controls:
the orchestrator-relay surface is a machinery write the child
never authors; the question-path tool machine-stamps entries
and refuses synthetic-prompt answers; authorization requires a
pre-existing recorded entry (two deliberate tool calls, not
one); and the append-only record with bounded identifiers is
an auditable trail.

The enforcement layers are deliberately defense-in-depth with
distinct authority: the command-keyed step allowlist
(SYNTH-003) is the authoritative gate — an undeclared step is
rejected regardless of its data — while the data-key scan
(SYNTH-001, any key at any nesting depth containing `override`)
is a tripwire for override data riding declared step names. No
mechanical predicate can detect the semantic intent of
free-form data keys: a renamed key (`checklist_bypass`) is
inert to the tooling — downstream consumers read specific
documented keys (fix-loop budgets, phase verdicts), never
arbitrary ones — and an agent honoring an unrecognized
free-form data key as authorization is exactly the prose
failure SYNTH-002 pins and its contract test regression-locks.

## Schema delta (SYNTH-001/003/006)

The SYNTH-001, SYNTH-003, and SYNTH-006 spec text is the single
normative shape. This section consolidates it for implementation;
any divergence resolves in the spec text's favor. Task 2.1 encodes
these shapes once in `.opencode/lib/uf-workflow-types.ts` (per
Decision 1) so every producer and consumer shares one definition.

### `human_decision_ref` — `checkpoint` argument

- Type: string; at most 256 characters; MUST match the UUID v4
  pattern of the `decision_id` it names (SYNTH-006).
- Semantics: names a `decision_id` present in the persisted
  human-decision record for the state key (SYNTH-001). Reference
  only — `checkpoint` MUST NOT accept an inline human-decision
  payload in place of a reference.
- Required when any `data` key at any nesting depth contains the
  substring `override` (case-insensitive), covering checklist,
  spec-review advisory escalation, and quality-gate overrides
  (SYNTH-001).
- Bounds and pattern violations are rejected with a
  machine-parseable error BEFORE any security-sensitive operation —
  file path derivation, error construction, state persistence
  (SYNTH-006 ordering).

### Human-decision record entry (SYNTH-006)

Store: `.uf/progress/<key>/human-decisions.json`, append-only.

- `decision_id`: UUID v4, machine-generated (callers never supply
  it).
- `source`: `question-tool` or `orchestrator-relay`.
- `question`, `answer`: at most 16 KiB each.
- `recorded_at`: machine-generated.
- Callers supply only source, question, and answer.

Write surfaces (exactly two, SYNTH-001):

1. Orchestrator relay — `uf_delegate_command` persists entries
   from its typed `human_decisions` envelope parameter at spawn
   time, before the delegated child session runs; a machinery
   write the child never authors.
2. In-session question path — the `record_human_decision` tool
   records a `question`-tool answer the session received,
   refusing entries whose answer, after normalization (trimming,
   whitespace collapsing, case folding), contains the normalized
   synthetic compaction prompt as a substring.

Store discipline (SYNTH-006): the record path derives through the
same sanitized key resolution as the progress state store — slug
sanitization, path-traversal rejection, symlink refusal — with
directory mode 0o700 and file mode 0o600, atomic
temporary-file-plus-rename writes, serialized appends (advisory
lock or equivalent exclusive-write protocol), and VCS exclusion
under the gitignored `.uf/progress/` tree. An absent record file is
the empty-record case; any other read failure fails closed. Errors
and logs name bounded identifiers (`decision_id`, step, data key)
only — never raw question or answer text, and question or answer
content is never logged at INFO or above.

### Error codes

All four codes ride the existing `failure(message,
{retryable, code})` envelope (uf-workflow-types.ts) with
machine-parseable, bounded-identifier payloads:

- `missing_human_decision` — a gate override or undeclared-step
  authorization carries no valid reference: absent, fabricated
  (names no recorded entry), or otherwise unresolvable. Names the
  offending step and override-carrying data key. An absent record
  file reports this code (the empty-record case), never
  `human_decision_record_unreadable`.
- `undeclared_step` — a step name outside the declared set for
  `uf.unleash` without a valid `human_decision_ref`. Names the
  step; step names are bounded at 128 characters.
- `escalated_advisory_unrelayed` — the latest consolidated
  spec-review result under command `uf.unleash` is APPROVE WITH
  ADVISORIES (persisted canonical decision `ESCALATED`) and no
  human-decision reference covers it. Rejects the advancement
  steps `implement` and `implement-phase-N`; names the correlation
  and finding count; mirrors the HARDEN-002 hold shape (a hold
  over advancement steps, not an allowlist of permitted ones).
- `human_decision_record_unreadable` — a record read failure other
  than ENOENT (permission denied, corrupted JSON, any other I/O
  error). Fails closed rather than treating the failure as a
  missing reference or crashing.

### Declared step allowlist for `uf.unleash` (SYNTH-003)

The declared set is exactly:

`branch-safety-gate`, `tier-routing`,
`resumability-detection`, `clarify`, `plan`, `tasks`,
`spec-review`, `implement`, `implement-phase-N` (N a positive
integer), `code-review`, `quick-pre-flight`, `quick-fix-budget`,
`retrospective`, `demo`.

- The predicate keys on the call's `command` argument resolving to
  `uf.unleash` — the same command keying the HARDEN-002
  spec-review gate uses. Every other command identifier carries
  its own step vocabulary and is unaffected.
- The set is the union of both writer surfaces (the pipeline's
  user-facing steps, the plugin's auto-checkpointed machinery
  steps, and the command prose's gate/budget steps); the task 4.2
  lockstep consistency test enumerates both surfaces scoped to
  steps recorded under `uf.unleash` so a newly recorded step name
  missing from the allowlist fails the suite.
- A step name matching `*override*` is never in the declared set,
  so it is rejected whenever the human-decision reference is
  absent or invalid.
- Step names are bounded at 128 characters (shared with the
  `undeclared_step` contract above).

## Risks / Trade-offs

- **Rejected legitimate override with no human in the loop.**
  Mitigation: unattended runs (dewey#153) already use the
  orchestrator relay path — the orchestrator records the relayed
  human decision via `record_human_decision` before delegating,
  so the child's `checkpoint` call resolves the reference from
  the human-decision record and unattended advancement is
  possible when the orchestrator has relayed the decision.
  Pure free-text overrides with no human input remain rejected,
  which is the intended security property.

- **Strict step list maintenance cost.** The strict predicate is
  the MUST, so new legitimate steps must be added to the
  declared allowlist or `checkpoint` rejects them. The cost is
  accepted: the allowlist is small, changes with the command
  prose that declares the steps, and a rejected legitimate step
  fails loudly with `undeclared_step` rather than silently
  advancing.

- **Scaffold mirror drift.** Any prose change must be mirrored
  byte-identically or `TestEmbeddedAssets_MatchSource` fails.
  The contract test for the new prose catches drift in both
  directions.
