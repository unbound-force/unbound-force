<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Per-task derivation library

- [x] 1.1 Add `derivePerTaskTimeoutMs(signals)` pure helper in `.opencode/lib/run-timeout.ts` alongside `derivePerRunTimeoutMs` (D1): inputs `phaseTaskCount`, `phaseFileCount`, `promptBytes`, `tier` (`lightweight`/`standard`/`heavy`/tier-less 0), `maxParallelRuns`; monotonic in phase size + files, contention multiplier 0.15 capped at 1.5, clamp to `min(ceiling, 1_800_000)` with floor-at-base precedence, explicit `timeout` passthrough wins (PTB-FR-001).
- [x] 1.2 Add `.opencode/test/run-timeout.test.ts` per-task unit tests: formula arithmetic vs flat default, monotonicity in task/file count and tier, contention only above `max_parallel_runs` 1 and capped at 1.5, floor never below base, ceiling clamp including `1_800_000` bound and ceiling-below-base → base floor, tier-less zero term, determinism, explicit-source passthrough (PTB-FR-001 scenarios).

## 2. Executor and budget registry wiring

- [x] 2.1 In `.opencode/lib/delegation-budget.ts`, expose per-task applicability for `probeSubtreeActivity` / inherited `idle_timeout` without adding a new global default: per-task children of a budgeted command-path parent resolve via existing registry + `resolveInheritedDeadline` (≤8 hops) and `clampToBudget` (60 s reserve / 30 s minimum) (D2, PTB-FR-002).
- [x] 2.2 In `.opencode/lib/agent-executor.ts`, wire per-task delegation budgeting: resolve inherited deadline before child creation, clamp requested timeout, fail fast `budget_exhausted` (non-retryable) when budget exhausted, register/unregister deadline around `executeAgentSession`, and attach optional per-task timeout provenance + `resume` assembly (files_modified from worktree diff / file tracking + tasks_completed from `parse_tasks` / `tasks.md` checkbox diff, `resume: null` when unreadable) to the `timeout` result envelope (D2/D3, PTB-FR-001..003, AE-FR-010).
- [x] 2.3 Extend `.opencode/test/delegation-budget.test.ts` and `internal/scaffold` executor fakes for per-task registry / clamp / idle inheritance: registry hit vs ancestry walk, clamp boundaries, `budget_exhausted` fail-fast, `probeSubtreeActivity` progress vs stall, `resume` hint assembly, `resume: null` fail-closed, `child_active` handling (PTB-FR-002/003 scenarios).

## 3. uf_delegate_command and unleash orchestration

- [x] 3.1 Update `.opencode/plugins/uf-workflow/index.ts` `uf_delegate_command` failure envelope to carry the optional per-task `resume` hint alongside `timeout` when `checkpoint_step: "implement"` + `phase` are present; keep edits inside the `uf_delegate_command` section as in #930, preserve `TimeoutSchema` `1_800_000` ceiling (PTB-FR-003, gatekeeping).
- [x] 3.2 Update `.opencode/commands/uf.unleash.md` implementation-phase delegation call site: derive timeout per PTB-FR-001 (phase task count + files touched + tier via `derivePerTaskTimeoutMs`), wire idle semantics under command-path parent per PTB-FR-002, thread `checkpoint_step`/`phase` and relay `resume.files_modified`/`tasks_completed` into retries (must not retry blank when `child_active: true` or resume present) per PTB-FR-003, and amend the phase-checkpoint rendering to include remaining parent budget (`remaining_ms` + human line) per PTB-FR-004/UO-FR-010 (D1–D4). Keep 72-column wrap.
- [ ] 3.3 [P] Add Go content gate in `internal/scaffold/` asserting the `uf.unleash.md` per-task retrieval / remaining-budget prose (mirrors D4 pattern from #930).
- [ ] 3.4 [P] Extend `.opencode/test/unleash-delegate.test.ts` for per-task schema bounds (retain `1_800_000` ceiling), derived timeout, idle abort vs `budget_exhausted`, resume envelope, and `checkpoint` phase payload `remaining_ms`.

## 4. Mirrors, docs, and scaffolding

- [ ] 4.1 Sync every edited `.opencode/` file to `internal/scaffold/assets/opencode/` byte-identically and every edited `internal/` asset to its scaffold mirror; register `.opencode/lib/run-timeout.ts` per-task helper on the scaffold asset surface (`Shared lib modules` enumeration + drift table) as in #849.
- [ ] 4.2 Add `CHANGELOG.md` Unreleased entry for derived per-task timeout, idle inheritance, lightweight `resume` hint, outer-budget checkpoint visibility, and provenance.
- [ ] 4.3 File the user-facing documentation issue in `unbound-force/website` covering derived per-task timeouts, idle vs wall-clock under a budgeted parent, the `resume` advisory hint (`files_modified`/`tasks_completed`), and the checkpoint remaining-budget line per the Cross-Repo Documentation rule; or record the internal-tooling exemption rationale in the CHANGELOG entry from 4.2 if classified non-user-facing. The classification MUST weigh the tightened per-task `timeout` contract, new `timeout.resume` / `remaining_ms` fields, and command prose guidance.

## 5. Verification

- [ ] 5.1 Run `make check` (lint, Go tests with `-race -count=1`, build, coverage-gate, `plugin-test`) and confirm the `.opencode/` plugin suite stays at or above the 94 percent statements coverage floor — CI parity with `ci_local.yml`.
- [ ] 5.2 Verify constitution alignment: derivation is a pure standalone function (II, IV, VI), budget clamping reuses the existing registry with no new mandatory dependency (II), provenance + `remaining_ms` + budget line are machine-parseable artifacts with no runtime coupling (I, III), input validation is preserved not weakened and no secrets/deps are introduced (V), and all new behaviors are testable via injected fakes without external services (IV). Record the verification outcome in the CHANGELOG entry from 4.2.
- [ ] 5.3 Verify gatekeeping: `git diff` shows the `TimeoutSchema` `1_800_000` bound unchanged, no coverage/CRAP/severity gate files are modified, and the canonical `.uf/review-matrix.yaml` and its scaffold mirror are value-identical to the base branch; stop and report per the gatekeeping rule if any check fails.
