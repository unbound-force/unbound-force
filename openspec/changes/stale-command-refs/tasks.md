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

## 1. Update Stale Command References

- [x] 1.1 [P] In `cmd/unbound-force/main.go` line 99, replace
      `/review-council` with `/uf.review-council` in the
      `--divisor` flag help text.
- [x] 1.2 [P] In `openspec/specs/review-council/spec.md`, replace
      all six occurrences of `/review-council` with
      `/uf.review-council` (lines 3, 13, 30, 39, 48, 92).

## 2. Verification

- [x] 2.1 Run `grep -rn '/review-council' cmd/ openspec/specs/`
      and confirm zero hits in active source and current specs
      (excluding archived specs, changelog, migration logic,
      and stale-ref detection tests).
- [x] 2.2 Run `make check` to verify build, lint, and tests pass.
- [x] 2.3 Verify constitution alignment: confirm no behavioral
      changes, no new dependencies, no API surface changes.

<!-- spec-review: passed -->
