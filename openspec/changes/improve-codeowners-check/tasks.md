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

## 1. Update CODEOWNERS check in active command copies

- [x] 1.1 [P] Update `/review-pr` CODEOWNERS check
  - File: `.opencode/commands/review-pr.md`
  - Lines 802-820: Replace the `2>/dev/null` bash snippet
    and "skip silently" instruction with:
    - Three-path check: `.github/CODEOWNERS`, `CODEOWNERS`,
      `docs/CODEOWNERS` (short-circuit on first success)
    - 404 handling: treat as "not found" (silent)
    - Non-404 error handling: display inconclusive warning
    - Remove line 820 ("If any API call fails: skip
      silently.")

- [x] 1.2 [P] Update `/review-council` CODEOWNERS check
  - File: `.opencode/commands/review-council.md`
  - Lines 474-484: Replace the `2>/dev/null` bash snippet
    with the same updated logic as review-pr:
    - Three-path check with short-circuit
    - 404 vs non-404 error distinction
    - Inconclusive warning for non-404 errors
  - Ensure the CODEOWNERS check text is identical to the
    `/review-pr` version (adjusted for indentation context)

## 2. Sync scaffold source copies

- [x] 2.1 [P] Sync scaffold review-pr.md
  - File: `internal/scaffold/assets/opencode/commands/
    review-pr.md`
  - Copy the updated CODEOWNERS check section from
    `.opencode/commands/review-pr.md` to match exactly

- [x] 2.2 [P] Sync scaffold review-council.md
  - File: `internal/scaffold/assets/opencode/commands/
    review-council.md`
  - Copy the updated CODEOWNERS check section from
    `.opencode/commands/review-council.md` to match exactly

## 3. Verification

- [x] 3.1 Verify scaffold drift detection
  - Run `make test` to confirm scaffold drift detection
    tests pass (active copies match scaffold sources)

- [x] 3.2 Verify constitution alignment
  - Confirm Observable Quality (Principle III) is satisfied:
    non-404 errors are now surfaced rather than suppressed
  - Confirm no other constitution principles are violated

- [x] 3.3 Verify acceptance criteria from issue #324
  - All three CODEOWNERS paths checked
  - Non-404 errors produce a visible warning
  - 404 errors handled silently
  - `/review-pr` and `/review-council` use identical logic
  - Scaffold copies synced
  - No change to the CODEOWNER warning message itself
<!-- spec-review: passed -->
<!-- code-review: passed -->
