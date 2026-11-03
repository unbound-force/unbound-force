<!--
  [P] marks tasks eligible for parallel execution.
  Every .opencode/ file edited here has a byte-identical
  mirror under internal/scaffold/assets/opencode/.
-->

## 1. Budget library

- [ ] 1.1 Create `.opencode/lib/delegation-budget.ts`: the
  `globalThis` symbol-keyed deadline registry (register,
  unregister, lookup), `resolveInheritedDeadline` (registry hit,
  else bounded `parentID` walk, errors -> null),
  `clampToBudget` (60 s reserve, 30 s minimum), and
  `probeSubtreeActivity` (BFS over `session.children` capped at
  32 sessions, latest message per session, fingerprint plus
  `pending_question`) (D1, D2).
- [ ] 1.2 Add `.opencode/test/delegation-budget.test.ts` covering
  registry, ancestry walk depth and errors, clamp boundaries,
  fingerprint changes, probe failures, and the session cap.

## 2. Executor

- [ ] 2.1 In `executeAgentSession`, resolve the inherited deadline
  before creating the child, fail fast with `budget_exhausted`,
  apply the clamp, register the child's deadline after creation,
  and unregister in `finally` (D1).
- [ ] 2.2 Add the optional idle watchdog
  (`idleTimeoutMilliseconds`, injectable poll interval and
  clock); abort with `idle_timeout` only after a confirming
  successful probe (D2).
- [ ] 2.3 Attach the `timeout` details object to limit-ended
  results (D4) and extend executor tests.

## 3. uf_delegate_command

- [ ] 3.1 Widen `timeout` to 14,400,000 ms for command-path
  delegations (prompt-only stays at 1,800,000 via
  `superRefine`) and add `idle_timeout` (default 600,000) (D2,
  D3).
- [ ] 3.2 Thread `timeout` and, for command-path delegations,
  `resume` (read through `readState`) into the
  `DELEGATE_FAILED` envelope (D4).
- [ ] 3.3 Extend `.opencode/test/unleash-delegate.test.ts` for the
  schema bounds, idle abort, and resume envelope.

## 4. Command prose

- [ ] 4.1 Update `.opencode/commands/uf.run-pipeline-plan.md`:
  delegation 1 `timeout: 14400000`, and a timeout-recovery
  branch on `timeout.kind` / `pending_question` /
  `resume.last_checkpoint_step` (D5). Keep 72-column wrap.
- [ ] 4.2 Add a Go content gate in `internal/scaffold/` asserting
  the recovery prose.

## 5. Mirrors, docs, verification

- [ ] 5.1 Sync every edited `.opencode/` file to
  `internal/scaffold/assets/opencode/` byte-identically.
- [ ] 5.2 Add the CHANGELOG entry under Unreleased.
- [ ] 5.3 Run `make plugin-test` and
  `go test -race -count=1 ./...`.
