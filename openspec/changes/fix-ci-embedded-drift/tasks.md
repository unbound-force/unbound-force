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

## 1. Scaffold Asset Synchronization

- [x] 1.1 [P] Copy `.opencode/commands/uf.review-pr.md` to `internal/scaffold/assets/opencode/commands/uf.review-pr.md`
- [x] 1.2 [P] Copy `.uf/review-matrix.yaml` to `internal/scaffold/assets/uf/review-matrix.yaml`

## 2. Review Matrix CI Test Update

- [x] 2.1 Update `TestReviewMatrixSchema_CanonicalPolicyValidates` in `internal/schemas/ci_test.go` to accept nil or empty model values on profiles — the v3 schema makes model optional, and the canonical review-matrix.yaml uses `model: {}` for host-resolved dispatch

## 3. Verification

- [x] 3.1 Run `go test -race -count=1 -run 'TestEmbeddedAssets_MatchSource|TestSCFR001_CanonicalAssetSurfaceMatchesMirrors' ./internal/scaffold/` to confirm scaffold drift tests pass
- [x] 3.2 Run `go test -race -count=1 -run TestReviewMatrixSchema_CanonicalPolicyValidates ./internal/schemas/` to confirm schema test passes
- [x] 3.3 Run `make check` to confirm all checks pass
- [x] 3.4 Verify constitution alignment — all five principles remain satisfied (no changes to artifact formats, composability, quality metrics, test isolation, or security)
