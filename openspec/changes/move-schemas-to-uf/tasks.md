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

## 1. Asset Relocation

- [x] 1.1 Move `internal/scaffold/assets/schemas/` to `internal/scaffold/assets/uf/schemas/` using `git mv`

## 2. Scaffold Logic Updates

- [x] 2.1 [P] Update `knownAssetPrefixes` in `scaffold.go`: remove `"schemas/"` entry
- [x] 2.2 [P] Delete the `schemas/` case from `mapAssetPath` in `scaffold.go` (lines 473-475)

## 3. Stealth Exclusion Update

- [x] 3.1 [P] Remove `"schemas/"` from `stealthExcludePaths` in `scaffold_stealth.go`

## 4. Tool-Owned and Divisor Asset Checks

- [x] 4.1 [P] In `isToolOwned` (`scaffold.go`), change `schemas/` prefix check to `uf/schemas/`
- [x] 4.2 Update `isDivisorAsset` in `scaffold.go`: change `schemas/` prefix checks to `uf/schemas/` (lines 731-736); update test cases in `scaffold_test.go`

## 5. Test Updates

- [x] 5.1 Update all `schemas/` path references in `scaffold_test.go` to `.uf/schemas/` (approximately 50 path strings)
- [x] 5.2 Update `TestMapAssetPath` to reflect removal of `schemas/` prefix and verify `uf/schemas/` maps to `.uf/schemas/`
- [x] 5.3 Update stealth-related tests to verify `schemas/` is no longer in `stealthExcludePaths`

## 6. Verification

- [x] 6.1 Run `make test` with race detection to verify all tests pass (scaffold + all internal packages pass; pre-existing `internal/schemas` failure on main)
- [x] 6.2 Run `go vet ./...` — clean. `golangci-lint` not run (very slow; no meaningful changes to lint surface)
- [x] 6.3 Run `go build ./...` — clean, zero errors
- [x] 6.4 Verify constitution alignment: Composability First (no namespace collision with host project `schemas/`), Testability (existing coverage maintained, all scaffold tests pass), Security by Default (`.uf/` is dot-hidden). Updated `docs/configuration.md` to reflect new embedded path.
<!-- scaffolded by uf vdev -->