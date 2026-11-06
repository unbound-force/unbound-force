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

## 1. Gate Configuration

- [x] 1.1 Raise the statements threshold from 90 to 94 in
  `.opencode/vitest.config.ts` and rewrite the provenance comment
  per design D6: preserve the fanout D14 attribution for the
  branches floor (85, unchanged); record the unleash-tier-routing
  baselines (94.61% round-1, 94.58% round-2 final recorded), the
  2026-10-22 measurement (95.18% statements, 3521/3699), issue
  #780, and the governance constraint that the statements floor
  changes only through a spec-workflow governance change. No other
  configuration changes: branches stays 85, include/exclude globs
  and reporters untouched.

## 2. Verification

- [x] 2.1 Passing mode: run `cd .opencode && npx vitest run
  --coverage` — full suite green, coverage summary reports
  statements at or above 94 (reference: 95.18%, 3521/3699,
  measured at the 2026-10-22 branch point), thresholds met.
- [x] 2.2 Failing mode (scratch verification, no shipped
  changes): run `cd .opencode && npx vitest run --coverage
  --coverage.thresholds.statements=96` — the run exits non-zero
  with an error naming statements, the measured value, and the
  threshold (verified during design: exit 1, `ERROR: Coverage for
  statements (95.18%) does not meet global threshold (96%)`). The
  CLI override leaves the config file untouched; confirm `git
  diff -- .opencode/vitest.config.ts` shows only task 1.1's edit.
- [x] 2.3 Full target: run `make plugin-test` — green with the
  raised floor (lockless pinned-npm install followed by
  `npx vitest run --coverage`).
- [x] 2.4 CI parity: run `make check` — lint, Go tests (including
  the scaffold drift contracts, which cover the live-only status
  of `.opencode/vitest.config.ts`), build, coverage-gate, and
  plugin-test all green.

## 3. Documentation

- [x] 3.1 [P] Add a Changed entry under Unreleased in
  CHANGELOG.md for ts-plugin-coverage-gate: statements coverage
  floor for the `.opencode/` TypeScript plugin suite raised from
  90 to 94 percent, enforced by the existing `plugin-test` /
  `make check` / Local CI `plugin-tests` chain; closes the
  coverage-gate deferral tracked by issue #780.
- [x] 3.2 [P] Add one sentence to the Testing Conventions section
  of AGENTS.md recording the 94 percent statements floor on the
  `.opencode/` plugin suite and that the floor is
  governance-controlled (changes only through a spec-workflow
  governance change per the gatekeeping rule).
- [x] 3.3 Assess the documentation gate: confirm no user-facing
  documentation (README.md, docs/) requires updates — the floor is
  an internal quality gate with no user-facing surface; claim the
  CI-only exemption from the documentation-issue rule and note it
  in the pull request description.

## 4. Constitution Alignment and Validation

- [x] 4.1 Verify constitution alignment against the proposal's
  PASS assessments: Observable Quality (threshold failure names
  metric, measured value, and threshold — proven by 2.2);
  Deterministic Execution (tool-owned Vitest thresholds,
  checked-in value, identical local and CI command); Testability
  (both gate modes proven by executable commands in 2.1 and 2.2);
  Composability First (suite independently gateable via its
  documented command); Autonomous Collaboration (self-describing
  provenance comment from 1.1).
- [x] 4.2 Run `openspec validate --change ts-plugin-coverage-gate`
  — valid with 0 errors (house-norm long-text warnings are
  acceptable per repo precedent).
