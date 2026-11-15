# Tasks: synthetic-compaction-continue-prompt-rec

<!--
  [P] marks tasks eligible for parallel execution.
  Tasks without [P] run sequentially first, then [P] tasks run
  in parallel.
-->

## 1. Research and scaffolding

- [ ] 1.1 Trace checkpoint, progress-state, and relay flow — locate
  `checkpoint` / `get_progress` / `relay_spec_review_gate` in
  `.opencode/plugins/uf-workflow/index.ts`; shared types in
  `.opencode/lib/uf-workflow-types.ts`; `consolidate_dispatch` gate
  record in `.opencode/plugins/review-dispatch/index.ts`; and prose
  in `.opencode/commands/uf.unleash.md`,
  `.opencode/commands/uf.run-pipeline-plan.md`, and
  `.opencode/agents/` autonomy agent
- [ ] 1.2 Define checkpoint schema delta for SYNTH-001/003 — fields
  for `human_decision_ref` (question-tool or relay id), error codes
  `missing_human_decision` / `undeclared_step` / `escalated_advisory
  _unrelayed`, and the declared step allowlist for `uf.unleash`;
  the SYNTH-001 spec text is the single normative shape
- [ ] 1.3 Identify scaffold mirror pairs for prose changes —
  enumerate `internal/scaffold/assets/opencode/commands/uf.unleash.md`
  ↔ `.opencode/commands/uf.unleash.md`,
  `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md`
  ↔ `.opencode/commands/uf.run-pipeline-plan.md`, and
  `internal/scaffold/assets/opencode/agents/…` ↔ `.opencode/agents/…`

## 2. Gate-override human-decision reference (SYNTH-001)

- [ ] 2.1 Extend shared types — add `human_decision_ref` shape to
  `.opencode/lib/uf-workflow-types.ts` (question-tool answer
  correlation or orchestrator relay id) alongside the existing
  spec-review decision types so all consumers share one definition
- [ ] 2.2 Enforce human-decision reference in `checkpoint` — when
  `data` carries a gate override (checklist, spec-review advisory
  escalation, quality gate), require a matching recorded human
  decision (`question`-tool answer in session context or relayed
  human decision from delegation input); reject free-text
  justifications quoting the synthetic prompt with
  `missing_human_decision` naming the offending step and field
- [ ] 2.3 Add unit tests for SYNTH-001 — free-text synthetic prompt
  quote without ref is rejected; prior-session free-text citation
  without own ref is rejected; valid `human_decision_ref` (both
  question-tool and relay variants) is accepted and persisted with
  provenance

## 3. Synthetic-prompt-not-consent prose and contract test (SYNTH-002)

- [ ] 3.1 Add synthetic-prompt-not-consent prose to commands and
  agent — in `.opencode/commands/uf.unleash.md` (near session-
  resume/human-input handling), `.opencode/commands/uf.run-pipeline-
  plan.md`, and the autonomy agent, state explicitly that
  compaction/synthetic continue prompts ("Continue if you have next
  steps, or stop and ask for clarification if you are unsure how to
  proceed") are not human input and MUST NOT be treated as
  authorization for any gate override
- [ ] 3.2 Mirror prose to scaffold copies — copy each edited command/
  agent file byte-identically to its `internal/scaffold/assets/…`
  mirror so `TestEmbeddedAssets_MatchSource` passes
- [ ] 3.3 Add contract test pinning the prose — assert the exact
  synthetic-prompt-not-consent paragraph exists verbatim in both
  the live copy and the scaffold mirror for each touched file;
  drift in either direction is a test failure

## 4. Declared checkpoint step enforcement (SYNTH-003)

- [ ] 4.1 Add step-name predicate to `checkpoint` — maintain the
  declared step allowlist for `uf.unleash` (`clarify`, `plan`,
  `tasks`, `spec-review`, `implement`, `implement-phase-N`,
  `code-review`, `retrospective`, `demo`); reject steps not in the
  allowlist, or at minimum those matching `*override*`, unless a
  valid `human_decision_ref` per SYNTH-001 is present, with an
  `undeclared_step` error naming the step
- [ ] 4.2 Add tests for SYNTH-003 — `checklist-override-phase8-9`
  without ref is rejected; override step with valid ref is accepted;
  normal declared step `implement` / `implement-phase-2` is
  unaffected; case covering widening to full allowlist is documented

## 5. Escalated advisory relay (SYNTH-004)

- [ ] 5.1 Widen `relay_spec_review_gate` predicate — treat
  `APPROVE WITH ADVISORIES` with `ESCALATED` the same as
  `REQUEST CHANGES` for relay purposes; the relay path and output
  shape remain the existing tool's contract
- [ ] 5.2 Gate advancement on escalated advisories — `checkpoint`
  MUST refuse to advance from `spec-review` to `implement` while
  the latest consolidated spec-review result is ESCALATED and no
  human-decision reference covers the escalation; error
  `escalated_advisory_unrelayed` names the correlation and finding
  count
- [ ] 5.3 Add tests for SYNTH-004 — ESCALATED 16-finding case from
  `8ef4f276-…` is relayed and blocks advancement without a decision;
  non-escalated APPROVE WITH ADVISORIES advances without relay

## 6. Regression coverage for incident JSON (SYNTH-005)

- [ ] 6.1 Add deterministic regression fixture replaying
  `ses_e200afc75ffeSuftvVXtQnBJhR` — supply the exact incident
  `checkpoint` payload (command `uf.unleash`, step
  `checklist-override-phase8-9`, data `checklist_override` quoting
  'Continue if you have next steps' and prior-session override,
  no `human_decision_ref`); assert rejection per SYNTH-001 and
  SYNTH-003 and no persisted progress for that step
- [ ] 6.2 Add synthetic-text-alone assertion — the literal prompt
  text "Continue if you have next steps, or stop and ask for
  clarification if you are unsure how to proceed" supplied as a
  justification field alone MUST NOT satisfy SYNTH-001

## 7. Verification and constitution alignment

- [ ] 7.1 Run `openspec validate synthetic-compaction-continue-prompt-rec`
  and fix deltas/scenarios until clean
- [ ] 7.2 Verify constitution alignment — confirm I Autonomous
  Collaboration (artifact-mediated authorization), II Composability
  First (no new mandatory dep), III Observable Quality (machine-
  parseable errors + provenance), IV Testability (isolated fixtures,
  no external services), V Security by Default (synthetic-prompt
  privilege escalation closed), VI Deterministic Execution (tool-
  enforced gate, not prose)
- [ ] 7.3 Run `go test -race -count=1 ./...`, `make plugin-test`, and
  `make check`; ensure no scaffold drift
  (`TestEmbeddedAssets_MatchSource`) and contract test passes
- [ ] 7.4 Documentation updates before PR merge per the proposal's
  Documentation Impact assessment — add CHANGELOG.md entry for the
  checkpoint gate hardening and confirm no-update rationale for
  AGENTS.md, README.md, and docs/ still holds at implement time
- [ ] 7.5 Record instruction-lint delta totals for the command/agent
  prose edits — per-file rows and totals from `make lint-instructions`
  — in the change record (DE-004 efficiency evidence)
