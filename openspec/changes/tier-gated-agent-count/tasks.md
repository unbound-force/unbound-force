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

## 1. Schema & Configuration

- [x] 1.1 Extend `LimitsSchema` with optional `tier_caps`
  field (object with `lightweight`, `standard`, `heavy`
  keys, each `z.number().int().positive().nullable()`.
  optional). Extend the `Limits` TypeScript interface to
  match. File: `.opencode/plugins/review-dispatch/index.ts`
- [x] 1.2 Add `tier_caps` to `DEFAULT_LIMITS` constant
  with values `{ lightweight: 2, standard: null,
  heavy: null }`.
  File: `.opencode/plugins/review-dispatch/index.ts`
- [x] 1.3 Update `limitsFor()` to merge `tier_caps`
  defaults with matrix overrides. Nested merge: each
  tier key merges individually so a matrix can override
  just `lightweight` without clobbering other tier
  defaults.
  File: `.opencode/plugins/review-dispatch/index.ts`

## 2. Core Tier Cap Logic

- [x] 2.1 Implement `applyTierCap()` function that
  accepts the list of included plan entries, the tier
  string, and the resolved limits. Returns the filtered
  list plus any tier-cap skip entries. Priority order:
  always-required > scope-match (alphabetical) >
  curator-relevant. Agents beyond the cap get
  `decision: "skip"`, `reason_code: "tier-cap"`.
  Floor enforcement: clamp effective cap to
  `max(cap, always_required_count)`.
  File: `.opencode/plugins/review-dispatch/index.ts`
- [x] 2.2 Integrate `applyTierCap()` into `buildPlan()`
  after the relevance loop and before run resolution.
  Guard conditions: skip when `input.full` is true;
  skip when command mode is `triage`, `feedback`, or
  `test`; skip when the tier's cap is null.
  File: `.opencode/plugins/review-dispatch/index.ts`
- [x] 2.3 Add heavy tier cap advisory warning: when
  `tier_caps.heavy` is configured (non-null) and the
  diff is security-sensitive, emit a plan-level
  advisory with severity `MEDIUM`.
  File: `.opencode/plugins/review-dispatch/index.ts`

## 3. Tests

- [x] 3.1 [P] Add unit tests for `applyTierCap()`:
  lightweight caps at 2, standard/heavy uncapped,
  always-required agents survive, alphabetical
  determinism, full flag bypass, mode exclusion.
  Cover all 12 Given/When/Then scenarios from the spec.
  File: `.opencode/test/review-dispatch.test.ts`
- [x] 3.2 [P] Add integration test for custom tier_caps
  override via review matrix limits: verify a matrix
  with `tier_caps.lightweight: 3` produces a plan with
  at most 3 agents for a lightweight diff.
  File: `.opencode/test/review-dispatch.test.ts`
- [x] 3.3 [P] Add tests for floor enforcement and heavy
  cap advisory: verify cap clamped when below
  always-required count, verify advisory emitted for
  security-sensitive heavy diff with custom cap.
  File: `.opencode/test/review-dispatch.test.ts`

## 4. Verification

- [x] 4.1 Run `make plugin-test` to verify all existing
  and new tests pass
- [x] 4.2 Verify constitution alignment: Observable
  Quality (tier-cap reason code in plan output),
  Composability First (relevance logic unchanged),
  Security by Default (floor enforcement + advisory),
  Testability (new function is unit-testable in
  isolation)
- [x] 4.3 Run `make check` for full CI parity
- [x] 4.4 Update CHANGELOG.md with tier-cap feature entry

<!-- spec-review: passed -->
<!-- code-review: passed -->
