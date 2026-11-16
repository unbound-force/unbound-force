# Tasks: synthetic-compaction-continue-prompt-rec

<!--
  [P] marks tasks eligible for parallel execution.
  Tasks without [P] run sequentially first, then [P] tasks run
  in parallel.
-->

## 1. Research and scaffolding

- [x] 1.1 Trace checkpoint, progress-state, and relay flow — locate
  `checkpoint` / `get_progress` / `relay_spec_review_gate` /
  `uf_delegate_command` in
  `.opencode/plugins/uf-workflow/index.ts`; shared types in
  `.opencode/lib/uf-workflow-types.ts`; `consolidate_dispatch` gate
  record in `.opencode/plugins/review-dispatch/index.ts` (note the
  `canonicalDecision` mapping of every APPROVE WITH ADVISORIES to
  ESCALATED); and prose
  in `.opencode/commands/uf.unleash.md`,
  `.opencode/commands/uf.run-pipeline-plan.md`, and
  `.opencode/agents/cobalt-crush-dev.md`
- [x] 1.2 Define checkpoint schema delta for SYNTH-001/003/006 —
  fields for `human_decision_ref` (naming a recorded
  `decision_id`), error codes `missing_human_decision` /
  `undeclared_step` / `escalated_advisory_unrelayed` /
  `human_decision_record_unreadable`, and the declared step
  allowlist for `uf.unleash`; the SYNTH-001 and SYNTH-006 spec
  text is the single normative shape
- [x] 1.3 Identify scaffold mirror pairs for prose changes —
  enumerate `internal/scaffold/assets/opencode/commands/uf.unleash.md`
  ↔ `.opencode/commands/uf.unleash.md`,
  `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md`
  ↔ `.opencode/commands/uf.run-pipeline-plan.md`, and
  `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md`
  ↔ `.opencode/agents/cobalt-crush-dev.md`

## 2. Gate-override human-decision reference (SYNTH-001, SYNTH-006)

- [x] 2.1 Extend shared types — add `human_decision_ref` and the
  human-decision record entry shape (`decision_id` UUID v4,
  source `question-tool` or `orchestrator-relay`, question,
  answer, `recorded_at`) to `.opencode/lib/uf-workflow-types.ts`
  alongside the existing spec-review decision types so all
  consumers share one definition
- [x] 2.2 Persist orchestrator relays in the delegation envelope —
  add the typed `human_decisions` parameter to
  `uf_delegate_command`; the plugin machinery appends the
  relayed decisions to the human-decision record
  (`.uf/progress/<key>/human-decisions.json`) at spawn time,
  before the child session runs, under the SYNTH-006 store
  contract (sanitized key resolution and traversal/symlink
  guards as the progress store, 0o700/0o600, atomic
  temp-file-plus-rename writes, concurrency-safe appends)
- [x] 2.3 Add `record_human_decision` tool — record in-session
  `question`-tool answers as append-only entries with
  machine-generated `decision_id` and `recorded_at` (callers
  supply only source, question, answer); refuse entries whose
  answer, after normalization (trimming, whitespace collapsing,
  case folding), contains the normalized synthetic compaction
  prompt as a substring; enforce the SYNTH-006 input bounds
  (question/answer at most 16 KiB each) and
  bounded-identifier-only errors
- [x] 2.4 Enforce human-decision reference in `checkpoint` — when
  any `data` key at any nesting depth contains `override`
  (case-insensitive) marking a gate override, require a
  `human_decision_ref` naming a `decision_id` present in the
  human-decision record (ref only — no inline payload); reject
  free-text justifications quoting the synthetic prompt,
  fabricated references naming no recorded entry, and
  non-ENOENT read failures (fail closed with
  `human_decision_record_unreadable`), with
  `missing_human_decision` naming the offending step and data
  key; treat an absent record file as the empty-record case;
  require `human_decision_ref` to match the UUID v4 pattern and
  at most 256 characters
- [x] 2.5 Add unit tests for SYNTH-001/006 — free-text synthetic
  prompt quote without ref is rejected; fabricated reference
  naming no recorded entry is rejected identically;
  prior-session free-text citation without own ref is rejected;
  declared step `implement` carrying `data.checklist_override`
  without ref is rejected; override data at nested depth is
  detected; valid `human_decision_ref` (question-tool and
  orchestrator-relay recorded entries) is
  accepted and persisted with provenance; traversal-shaped key
  is rejected; first-use missing record rejects with
  `missing_human_decision` (not unreadable); corrupted record
  fails closed; synthetic-prompt answer is refused at recording
  time including case/whitespace/quoting variation and
  prompt-embedded-in-larger-text; non-UUID or oversize
  reference is rejected before lookup; oversize or malformed
  inputs are rejected before any security-sensitive operation;
  concurrent appends are serialized (no lost entries)

## 3. Synthetic-prompt-not-consent prose and contract test (SYNTH-002)

- [x] 3.1 Add synthetic-prompt-not-consent and recovery prose to
  commands and agent — in `.opencode/commands/uf.unleash.md`
  (near session-resume/human-input handling),
  `.opencode/commands/uf.run-pipeline-plan.md`, and
  `.opencode/agents/cobalt-crush-dev.md` (near its gatekeeping
  prose), state explicitly that
  compaction/synthetic continue prompts ("Continue if you have next
  steps, or stop and ask for clarification if you are unsure how to
  proceed") are not human input and MUST NOT be treated as
  authorization for any gate override, and state the rejection
  recovery path at these decision points: on
  `missing_human_decision`, `undeclared_step`, or
  `escalated_advisory_unrelayed`, stop and obtain a real human
  decision via the `question` tool or the orchestrator relay —
  never compose a `record_human_decision` entry for a decision
  not actually received
- [x] 3.2 Mirror prose to scaffold copies — copy each edited command/
  agent file byte-identically to its `internal/scaffold/assets/…`
  mirror so `TestEmbeddedAssets_MatchSource` passes
- [x] 3.3 Add contract test pinning the prose — assert the exact
  synthetic-prompt-not-consent paragraph exists verbatim in both
  the live copy and the scaffold mirror for all three files
  (uf.unleash.md, uf.run-pipeline-plan.md, cobalt-crush-dev.md);
  drift in either direction is a test failure

## 4. Declared checkpoint step enforcement (SYNTH-003)

- [x] 4.1 Add step-name predicate to `checkpoint` — key the
  predicate on the call's `command` argument resolving to
  `uf.unleash` (the HARDEN-002 command keying; other commands'
  checkpoints are unaffected) and maintain the
  declared step allowlist for `uf.unleash` as the union of both
  writer surfaces: user-facing steps (`clarify`, `plan`,
  `tasks`, `spec-review`, `implement`, `implement-phase-N`,
  `code-review`, `retrospective`, `demo`), plugin
  auto-checkpointed machinery steps (`branch-safety-gate`,
  `tier-routing`, `resumability-detection`), and command-prose
  gate/budget steps (`quick-pre-flight`, `quick-fix-budget`);
  reject steps not in the allowlist unless a valid
  `human_decision_ref` per SYNTH-001 is present, with an
  `undeclared_step` error naming the step (step names bounded
  at 128 characters)
- [x] 4.2 Add lockstep consistency test — enumerate the step names
  prescribed by `.opencode/commands/uf.unleash.md` checkpoint
  calls and the plugin's auto-checkpoint call sites that record
  under command `uf.unleash` (the `reply-comments` step belongs
  to `uf.address-feedback` and is out of scope), and fail
  when a step name either surface records under `uf.unleash`
  is absent from the allowlist
- [x] 4.3 Add tests for SYNTH-003 — `checklist-override-phase8-9`
  without ref is rejected; override step with valid ref is
  accepted; an undeclared non-override step name is rejected
  (strict predicate); normal declared steps (`implement`,
  `implement-phase-2`, `branch-safety-gate`, `tier-routing`,
  `quick-pre-flight`) are unaffected; another command's
  checkpoint (`uf.review-council` `phase-1a-preflight`) is
  unaffected by the uf.unleash predicate

## 5. Escalated advisory relay (SYNTH-004)

- [x] 5.1 Widen `relay_spec_review_gate` predicate — treat
  `APPROVE WITH ADVISORIES` (persisted canonical decision
  `ESCALATED` — the finalizer maps every AWA consolidation to
  it, so the marker is not a distinguishing subset) the same as
  `REQUEST CHANGES` for relay purposes; a plain `APPROVE`
  (canonical decision `APPROVED`) still advances without relay;
  the relay path and output shape remain the existing tool's
  contract
- [x] 5.2 Gate advancement on advisory results — while the
  latest consolidated spec-review result under command
  `uf.unleash` is APPROVE WITH ADVISORIES and no human-decision
  reference covers it, `checkpoint` MUST reject the advancement
  steps `implement` AND `implement-phase-N` (any positive
  integer N — the `phase` argument records `implement-phase-N`
  directly with no prerequisite `implement` record) with
  `escalated_advisory_unrelayed` naming the correlation and
  finding count, mirroring the HARDEN-002 hold shape
- [x] 5.3 Add tests for SYNTH-004 — ESCALATED 16-finding case from
  `8ef4f276-…` is relayed and blocks advancement without a decision;
  plain APPROVE (canonical decision APPROVED) advances without
  relay; every APPROVE WITH ADVISORIES consolidation is treated
  as ESCALATED per the canonicalDecision mapping

## 6. Regression coverage for incident JSON (SYNTH-005)

- [x] 6.1 Add deterministic regression fixture replaying
  `ses_e200afc75ffeSuftvVXtQnBJhR` — supply the exact incident
  `checkpoint` payload (command `uf.unleash`, step
  `checklist-override-phase8-9`, data `checklist_override` quoting
  'Continue if you have next steps' and prior-session override,
  no `human_decision_ref`); assert rejection per SYNTH-001 and
  SYNTH-003 and no persisted progress for that step
- [x] 6.2 Add synthetic-text-alone assertion — the literal prompt
  text "Continue if you have next steps, or stop and ask for
  clarification if you are unsure how to proceed" supplied as a
  justification field alone MUST NOT satisfy SYNTH-001

## 7. Seam verification (DE-007)

- [x] 7.1 Seam-verify the widened plugin tool contracts — DE-007's
  three layers for the `checkpoint`, `record_human_decision`,
  `uf_delegate_command` (human_decisions envelope), and
  `relay_spec_review_gate` seams this change creates or
  widens: registration surface (tool registration plus
  input-schema parity — the zod args schema accepts the new
  `human_decision_ref` field, the `human_decisions` envelope
  parameter, and the new tool's args, matching the documented
  input; the allowlist lockstep consistency test from task 4.2
  is part of this layer), documented output shape (golden tests
  for the new error envelopes `missing_human_decision`,
  `undeclared_step`, `escalated_advisory_unrelayed`,
  `human_decision_record_unreadable` and the success envelope
  carrying provenance), and write→read
  round-trips as explicit write→read pairs — human-decision
  entries (envelope- and record_human_decision-written) read
  back through `checkpoint`'s reference validation;
  checkpoint-written progress state read back through
  `get_progress` (unchanged — context listing only, not a
  widened tool); the widened relay predicate verified against
  dispatch artifacts and the spec-review gate record through
  `relay_spec_review_gate`

## 8. Verification and constitution alignment

- [x] 8.1 Run `openspec validate synthetic-compaction-continue-prompt-rec`
  and fix deltas/scenarios until clean
- [x] 8.2 Verify constitution alignment — confirm I Autonomous
  Collaboration (artifact-mediated authorization), II Composability
  First (no new mandatory dep), III Observable Quality (machine-
  parseable errors + provenance), IV Testability (isolated fixtures,
  no external services), V Security by Default (synthetic-prompt
  privilege escalation closed), VI Deterministic Execution (tool-
  enforced gate, not prose)
- [x] 8.3 Run `go test -race -count=1 ./...`, `make plugin-test`, and
  `make check`; ensure no scaffold drift
  (`TestEmbeddedAssets_MatchSource`) and contract test passes
- [x] 8.4 Documentation updates before PR merge per the proposal's
  Documentation Impact assessment — add CHANGELOG.md entry for the
  checkpoint gate hardening and confirm no-update rationale for
  AGENTS.md, README.md, and docs/ still holds at implement time
- [x] 8.5 Record instruction-lint delta totals for the command/agent
  prose edits — per-file rows and totals from `make lint-instructions`
  — in the change record (DE-004 efficiency evidence)
