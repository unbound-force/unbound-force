<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
-->

## 1. Cask Transformation

- [x] 1.1 Add an internal Go transformer and a narrow command entry
  point that replace exactly one expected legacy Cask hook with the
  literal declarative `postflight_steps` replacement specified in the
  delta spec.
- [x] 1.2 Add isolated standard-library regression tests for successful
  transformation, surrounding-content preservation, missing hooks,
  duplicate hooks, and mismatched hooks; cover all five defined
  behavioral contracts using the delta spec's literal input and output
  fixtures, with exact success output and unchanged output after each
  rejection.
- [x] 1.3 Ensure transformer failures are contextual and leave no
  partially written Cask output.

## 2. Release Integration

- [x] 2.1 Update `.github/workflows/release.yml` to run the transformer
  after signed-checksum patching and before copying the Cask to the tap.
- [x] 2.2 Add fail-closed content and Homebrew static validation in the
  existing macOS release job for one supported stanza, no deprecated
  stanza, and both required install commands.
- [x] 2.3 Add an isolated staged-tap Cask install/uninstall smoke test
  on the macOS release runner; assert the installed executable and
  `uf` alias exist, clean up afterward, and prevent tap copy or push on
  failure.
- [x] 2.4 Add automated workflow regression coverage proving the order
  checksum patching, transformation, semantic validation, staged smoke
  test, tap copy, and `git push`, and proving transformer or validation
  failure makes publication unreachable.
- [x] 2.5 Verify the workflow preserves existing action pins,
  permissions, signing, checksum validation, and release quality gates.

## 3. Verification

- [x] 3.1 Run `go test -race -count=1 ./...` and `go test -cover` for
  the transformer package; verify fixture tests cover all five defined
  behavioral contracts without changing existing CI coverage gates.
- [x] 3.2 Run `goreleaser check` and inspect a snapshot-generated Cask,
  when GoReleaser is available, to confirm the release integration
  accepts the current configuration.
- [x] 3.3 Run applicable local release-workflow checks and verify the
  workflow regression test covers failed transformation and semantic
  validation paths.
- [x] 3.4 Verify the mandatory Homebrew static validation and staged-tap
  install/uninstall smoke test on the macOS release runner before a tap
  publication is accepted.

## 4. Governance Review

- [x] 4.1 Reassess the completed change against all five constitution
  principles and confirm its test strategy, generated-artifact
  validation, and release safeguards remain aligned.
- [x] 4.2 Assess documentation impact before completion; record that
  this CI-only packaging change requires no user-facing documentation
  update unless implementation changes that scope.

<!-- spec-review: passed -->
<!-- code-review: passed -->
