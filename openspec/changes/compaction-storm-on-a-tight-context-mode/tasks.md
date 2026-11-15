<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Foundations (selection + activity probe)

- [ ] 1.1 Add headroom config and footprint estimator
  (DELEG-FR-001) — classification: judgment (authorship).
  `.uf/model-matrix.yaml` `selection.footprint_margin_tokens`
  default 24k with range validation; footprint estimator in
  the `uf-workflow` plugin reading on-disk command bodies
  via repo-root path joins (traversal/symlink guarded,
  I/O-failure conservative default) and supplied to
  selection via `SelectionInputs`. Files:
  `.opencode/lib/model-matrix.ts`,
  `.opencode/plugins/uf-workflow/index.ts`,
  `.uf/model-matrix.yaml`, `docs/configuration.md`
  tool-owned mirror.

- [ ] 1.2 Extend `ProbeState`/`messageFingerprint` to
  exclude compaction/synthetic markers (DT-FR-002
  modification, DELEG-FR-002 accounting) — classification:
  judgment (authorship). Classify `system-reminder`
  compaction summaries and `Continue if you have next
  steps...` as non-progress; expose
  `compactionsSinceProgress`/`compactionCount` for the
  watchdog. Files: `.opencode/lib/delegation-budget.ts`.

## 2. Core behaviors

- [ ] 2.1 Implement headroom gate in `resolveSelection`
  (DELEG-FR-001) — classification: judgment (authorship).
  Compute `usable = limit.context - limit.output`, compare
  to `footprint + margin`, reject/penalize and append to
  `selection_fallback_reason`; all-rejected terminal
  fallback to host model with the full rejection chain;
  inconsistent limits rejected with reason; no-declared
  limits passes; preserve explicit-model bypass and
  validation order. Files: `.opencode/lib/model-matrix.ts`,
  `.opencode/plugins/uf-workflow/index.ts`.

- [ ] 2.2 Implement compaction-storm detector
  (DELEG-FR-002) — classification: judgment (authorship).
  Count N=3 consecutive compaction events without
  non-read progress, abort to `code: "context_thrash"`
  retryable with `resume` envelope (reuse
  `DelegateResumeState`), configurable via
  `delegation.storm_threshold`. Files:
  `.opencode/lib/agent-executor.ts`,
  `.opencode/lib/delegation-budget.ts`.

- [ ] 2.3 Implement final-text success guard
  (DELEG-FR-003) — classification: judgment (authorship).
  Post-execution structural classification (anchored,
  case-insensitive, length-bounded matchers on wrapper
  structure, never bare substrings) of compaction summary
  / synthetic handoff prompt; reclassify to
  `context_thrash` or `synthetic_final_text`, preserving
  sanitized original text in the error. Files:
  `.opencode/lib/agent-executor.ts`.

## 3. Tests (parallel-eligible)

- [ ] 3.1 [P] Unit tests for headroom gate and footprint
  estimator — classification: judgment (test
  authorship). Below/at/above threshold, missing limits
  bypass, inconsistent limits rejected, all-rejected host
  fallback, fallback_reason content, explicit-model
  uncaptured pack still emitted; plus the estimator's
  guard paths — traversal/symlink path-escape rejection
  and the I/O-failure conservative default (DE-001: the
  fail-closed property that a read failure cannot
  understate the footprint is test-verified). Files:
  `.opencode/test/model-matrix.test.ts`,
  `.opencode/test/model-matrix-integration.test.ts`,
  `.opencode/test/uf-workflow.test.ts`.

- [ ] 3.2 [P] Unit tests for storm detection —
  classification: judgment (test authorship). N-1 no
  trigger, N triggers, successful non-read progress
  resets the counter, reads interleaved between
  compactions do not reset it (the storm fires through
  read-only activity), failed non-read calls do not
  reset it, compaction excluded from probe progress.
  Files:
  `.opencode/test/delegation-budget.test.ts` (new or
  existing), `.opencode/test/agent-executor.test.ts`.

- [ ] 3.3 Unit tests for final-text guard —
  classification: judgment (test authorship). Compaction
  summary -> not success, synthetic prompt ->
  `synthetic_final_text`, legitimate final text (including
  prose that merely mentions compaction) -> success
  preserved, error preserves sanitized original text.
  Runs sequentially after 3.2: both modify
  `.opencode/test/agent-executor.test.ts` (MP-017 —
  same-file tasks MUST NOT both be [P]). Files:
  `.opencode/test/agent-executor.test.ts`.

- [ ] 3.4 [P] Integration, contract, and seam tests —
  classification: judgment (test authorship).
  `uf_delegate_command` surfaces
  `selection_fallback_reason` headroom message and
  `context_thrash`+`resume`/`synthetic_final_text`;
  `make check` still green for #930/#950 idle paths.
  Seam verification across the three prose-to-tool
  migration layers (DE-007): (1) registration surface —
  the new config keys are recognized and range-validated
  where knobs are registered; (2) documented output shape
  — `selection_fallback_reason` entries,
  `InvokeAgentResult` codes, `timeout.details`, and the
  `resume` envelope match their documented shape; (3)
  write→read round-trip — checkpoint data written by
  `uf_delegate_command` (fallback reason, storm window)
  reads back unchanged through the progress-state
  consumer (`get_progress`). Files:
  `.opencode/test/unleash-delegate.test.ts`,
  `docs/configuration.md` coverage.

## 4. Wiring and verification

- [ ] 4.1 Wire checkpoint/provenance — classification:
  judgment (authorship). `uf_delegate_command` checkpoint
  `selection_fallback_reason`; `InvokeAgentResult`
  `timeout.details`/`resume` fields documented;
  `docs/configuration.md` regenerated via scaffold; assert
  the write→read round-trip — the checkpoint data written
  here is readable by the orchestrator resume consumer
  via `get_progress` (seam 3 of task 3.4). Files:
  `.opencode/plugins/uf-workflow/index.ts`,
  `docs/configuration.md`.

- [ ] 4.2 Verification: constitution alignment and repo
  checks — classification: deterministic (exact-command
  verification; skill/script reference: repo-root
  `Makefile`, targets `check` and `lint-instructions`).
  Confirm I Autonomous Collaboration (artifact resume) /
  II Composability (no new deps) / III Observable Quality
  (fallback_reason, codes) / IV Testability (injected
  clock/mocks) / V Security by Default (fail-closed knob
  validation, sanitized errors) / VI Deterministic
  Execution; run the checks via the `Makefile` targets.
