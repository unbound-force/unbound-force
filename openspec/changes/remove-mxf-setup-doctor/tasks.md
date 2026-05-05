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

## 1. Remove mxf from setup and doctor source

- [x] 1.1 [P] In `internal/setup/setup.go`: remove the `installMxF` function (lines 453-466). Remove the step 3 block that calls it (lines 269-275: the `[3/14] Mx F` progress line, `shouldSkipTool("mxf")` check, and `installMxF` call). Renumber all subsequent step progress strings from `[4/14]` through `[14/14]` to `[3/13]` through `[13/13]`. Update any step count comments.
- [x] 1.2 [P] In `internal/doctor/checks.go`: remove the `mxf` entry from the `coreTools` slice (lines 116-119: `{ name: "mxf", recommended: true }`).

## 2. Update tests

- [x] 2.1 [P] In `internal/setup/setup_test.go`: remove `TestSetupRun_MxFMissing_BundledHint` (line 1862) and `TestSetupRun_MxFPresent` (line 1912). Remove `"mxf"` entries from LookPath stub maps at lines 189, 1934, 1978, 2075, 2127. Update any step count or step order assertions that reference 14 steps to 13. Update the install order comment at line 119 to remove `mxf (brew)`.
- [x] 2.2 [P] In `internal/doctor/doctor_test.go`: remove the `mxf` severity assertion block (lines 470-476). Remove `"mxf": true` from the LookPath stub at line 673. Remove `"mxf": "/usr/local/bin/mxf"` from stub at line 949. Remove the `mxf` install hint test case at line 1698. Remove the `{Name: "mxf", Severity: Warn}` expected result at line 2907.

## 3. Verification

- [x] 3.1 Run `make check` (build, test, vet, lint). Fix any failures.
- [x] 3.2 Verify no remaining `mxf` references in setup or doctor code: grep `internal/setup/setup.go` and `internal/doctor/checks.go` for `mxf` — the only remaining reference should be in `checkHeroAvailability` display names (`"mx-f"` at checks.go:625), which is correct and must stay.
- [x] 3.3 Update `CHANGELOG.md` with entry for this change.
<!-- spec-review: passed -->
<!-- code-review: passed -->
