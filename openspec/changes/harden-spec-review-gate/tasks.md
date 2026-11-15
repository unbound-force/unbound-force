# Tasks: harden-spec-review-gate

<!--
  [P] marks tasks eligible for parallel execution.
  Tasks without [P] run sequentially first, then [P] tasks run
  in parallel.
-->

## 1. Research and scaffolding

- [ ] 1.1 Trace current dispatch/checkpoint flow — locate
  `dispatch_review_runs` and `consolidate_dispatch` in
  `.opencode/plugins/review-dispatch/index.ts`; `checkpoint`,
  `get_progress`, and the `uf_unleash_detect_resume` spec-review
  resume path in `.opencode/plugins/uf-workflow/index.ts`; the
  shared progress-state shape in
  `.opencode/lib/uf-workflow-types.ts`; progress state schema,
  `.uf/artifacts/dispatch/` writes, and
  `/uf.run-pipeline-plan` post-delegation handling in
  `.opencode/commands/uf.run-pipeline-plan.md` and
  `.opencode/commands/uf.unleash.md`
- [ ] 1.2 Map existing iteration bookkeeping and stale-state
  path — identify where `current_step: spec-review` is read from
  progress vs. dispatch artifact and how iteration numbers are
  assigned/capped
- [ ] 1.3 Define progress-state schema delta for HARDEN-001/002
  — fields: `verdict`, `correlation_id`, `iteration`,
  `finding_ids[]`, `decisions[]`, `overrides[]` and failure modes
  for `missing_per_finding_decisions`,
  `pending_approving_iteration`, and `iteration_already_recorded`;
  the spec's HARDEN-001 requirement text is the single normative
  shape

## 2. Consolidation persistence before model resumption (HARDEN-001)

- [ ] 2.1 Make `consolidate_dispatch` persist spec-review verdict
  to progress state before returning — write `spec-review
  {iteration, verdict, correlation_id, finding_count,
  finding_ids}` atomically with the artifact (Decision 1a) or via
  the documented fallback; ensure `get_progress` reflects the
  latest iteration immediately
- [ ] 2.2 Add stale-state guard to `get_progress` — reconcile
  dispatch artifact not yet in progress state (if two-store),
  never return iteration N-1 when N is persisted; add test for
  `5355acba-62f4-4859-8775-7294b3e8702c` replay
- [ ] 2.3 Add unit/integration tests for HARDEN-001 scenarios —
  REQUEST CHANGES recorded (state visible immediately after
  `consolidate_dispatch` returns, before any edit) vs. stale
  check, APPROVE still advances without decisions

## 3. Per-finding gate before advancing spec-review (HARDEN-002)

- [ ] 3.1 Extend `checkpoint` to enforce per-finding decisions
  for non-APPROVE spec-review — require `apply` (with fix
  reference) | `override` (with reason) | `pause` for every
  `finding_id`; refuse with `missing_per_finding_decisions:
  [ids]`; refuse advancement after `apply` decisions until a
  subsequent iteration consolidates to APPROVE
  (`pending_approving_iteration`); all-override decision sets
  advance with `overrides[]` recorded
- [ ] 3.2 Validate `override` auditing — non-empty reason
  required, record `overrides: [{finding_id, reason}]` (empty
  when none), wire through `checkpoint` data
- [ ] 3.3 Add tests for refusal, apply-then-refuse,
  apply-then-APPROVE-then-advance, audited override, and
  all-override success paths

## 4. Iteration re-dispatch guard (HARDEN-003)

- [ ] 4.1 Add dispatch-layer guard for already-recorded iteration
  — on `dispatch_review_runs` for spec-review, check
  `progress_state.spec-review.iterations[N]`; if a verdict
  exists, return `iteration_already_recorded` naming the recorded
  iteration and verdict; enforce iteration cap before dispatch
- [ ] 4.2 Add tests for same-iteration refusal (no overwrite, no
  auto-bump) and normal next-iteration success

## 5. Orchestrator relay and resume adjudication (HARDEN-004/006)

- [ ] 5.1 Implement `relay_spec_review_gate` in
  `.opencode/plugins/uf-workflow/index.ts` — reads progress state
  + dispatch artifact
  (`.uf/artifacts/dispatch/<correlation>.json`), renders the
  gate-pause relay (finding severities, descriptions, file:line,
  correlation id) deterministically without an LLM; undecided
  gate → `{gate_paused: true, relay: …}`; decided gate, approved
  gate, or no spec-review record → `{gate_paused: false}`
- [ ] 5.2 Update `/uf.run-pipeline-plan` so the post-delegation
  step is one unconditional intent-level call — after a
  `/uf.unleash` child returns, the orchestrator calls
  `relay_spec_review_gate` (the tool reads progress state and
  the dispatch artifact and computes the pause state); present
  the rendered relay to the human when the tool returns
  `gate_paused: true` and proceed only when it returns
  `gate_paused: false`; no prose-evaluated condition gates the
  call, the child's text is never decision input, and the
  recorded iteration is never re-dispatched
- [ ] 5.3 Update `/uf.unleash` spec-review exit/relay to cite
  artifact source and prevent `overrides: []` auto-write when
  findings were never decided; align command prose with the
  tool-enforced gate
- [ ] 5.4 Extend `uf_unleash_detect_resume` to honor recorded
  adjudication (HARDEN-006) — all-override decision set →
  `spec_review` done; `apply` decisions → done only with the
  subsequent APPROVE iteration recorded (dirty spec artifacts do
  not re-fail the probe); undecided → incomplete, naming the
  iteration
- [ ] 5.5 Sync scaffold mirrors if command files changed —
  `internal/scaffold/assets/opencode/commands/uf.unleash.md` ↔
  `.opencode/commands/uf.unleash.md` and
  `internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md`
  ↔ `.opencode/commands/uf.run-pipeline-plan.md`, verify
  `TestEmbeddedAssets_MatchSource`
- [ ] 5.6 Add DE-007 seam verification naming the three layers
  for the seams this change creates — (1) registration surface:
  `relay_spec_review_gate`'s registered args match the
  documented contract, updating registration assertions in
  `.opencode/test/uf-workflow.test.ts` if they pin the argument
  set; (2) documented output shape: golden assertions on the
  `{gate_paused, relay}` envelope (including the
  no-spec-review-record case, reachable because the relay is
  called unconditionally) and on the
  `missing_per_finding_decisions`,
  `pending_approving_iteration`, and
  `iteration_already_recorded` error shapes; (3) write→read
  round-trips: the record `consolidate_dispatch` writes is
  readable through `get_progress`, `checkpoint`, and
  `relay_spec_review_gate`

## 6. Regression coverage for bypass path (HARDEN-005/006) [P]

- [ ] 6.1 Add regression fixture replaying
  `ses_e254aedb8ffeqlLkdOgSpKNj60` — mock dispatch store
  correlation `5355acba-…` 1 HIGH + 7 MEDIUM/LOW, 3 spec edits,
  compaction-summary return; assert (a) consolidation persisted
  before edits, (b) `checkpoint` refused, (c) `get_progress` not
  stale, (d) `relay_spec_review_gate` renders 8 findings from the
  artifact, no `overrides: []`
- [ ] 6.2 Add edge-case coverage — child returns empty text,
  child returns correct REQUEST CHANGES text (relay still
  artifact-sourced), re-dispatch after decisions succeeds as the
  next iteration
- [ ] 6.3 Add resume-probe coverage (HARDEN-006) — override-only
  adjudication survives resume (done, no re-dispatch);
  apply-decided + subsequent APPROVE survives resume with dirty
  spec artifacts; undecided iteration reports incomplete and
  re-enters the gate

## 7. Verification and constitution alignment

- [ ] 7.1 Run `openspec validate harden-spec-review-gate` and
  fix deltas/scenarios
- [ ] 7.2 Verify constitution alignment — confirm I Autonomous
  Collaboration (artifact+state), II Composability First (no new
  mandatory dep), III Observable Quality (HIC provenance), IV
  Testability (isolated fixtures), V Security by Default (HIGH
  bypass closed, audited overrides), VI Deterministic Execution
  (tool-enforced, not prose)
- [ ] 7.3 Run `go test -race -count=1 ./...`, `make plugin-test`,
  and `make check`; ensure no scaffold drift
  (`TestEmbeddedAssets_MatchSource`)
- [ ] 7.4 Documentation updates before PR merge per the
  proposal's Documentation Impact assessment — add the
  CHANGELOG.md entry for the spec-review gate hardening and
  confirm the no-update rationale for AGENTS.md, README.md,
  and docs/ still holds at implement time
- [ ] 7.5 Record instruction-lint delta totals for the
  `uf.unleash.md` and `uf.run-pipeline-plan.md` edits — per-file
  rows and totals from `make lint-instructions` — in the change
  record (DE-004 efficiency evidence)
