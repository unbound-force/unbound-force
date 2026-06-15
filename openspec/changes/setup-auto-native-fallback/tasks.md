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

  All tasks modify internal/setup/setup.go and
  internal/setup/setup_test.go. No [P] markers —
  all tasks are sequential to avoid merge conflicts
  in the same two files.
-->

## 1. Add resolveMethod() and installViaGo() Helpers

- [x] 1.1 Add `resolveMethod(toolName string,
  env doctor.DetectedEnvironment) string` method on
  `*Options` in `internal/setup/setup.go`. Logic:
  (1) if `toolMethod(toolName)` returns non-`"auto"`,
  return that value (per-tool override wins);
  (2) if `opts.PackageManager` is `"homebrew"` or
  `"dnf"`, return that (global preference);
  (3) if `opts.PackageManager` is `"apt"`, map to
  `"auto"` and log info message ("apt support not yet
  implemented, using auto-mode fallback");
  (4) otherwise return `"auto"` for fallback chain.
  `resolveMethod()` is tool-agnostic — it resolves the
  user's preference, not the tool's capability.
- [x] 1.2 Add `installViaGo(opts *Options, toolName,
  goModule string) stepResult` function in
  `internal/setup/setup.go`. Behavior by outcome:
  - Go not in PATH: return `action: "skipped"`,
    `detail` with install hint, `err: nil`.
  - `go install` succeeds: return
    `action: "installed"`, `detail: "via go install"`.
  - `go install` fails: return `action: "failed"`,
    `detail` with module path and retry hint,
    `err: fmt.Errorf("go install %s: %w", ...)` per
    CS-006 error wrapping.
  - Dry-run: return `action: "dry-run"`,
    `detail: "Would install: go install <module>@latest"`.
  All errors in `stepResult.err` MUST be wrapped with
  operation context per CS-006.
- [x] 1.3 Add tests for `resolveMethod()` in
  `internal/setup/setup_test.go` (table-driven per
  TC-006): per-tool override takes precedence over
  global, global `PackageManager: "dnf"` returns `"dnf"`,
  `PackageManager: "homebrew"` returns `"homebrew"`,
  `PackageManager: "apt"` maps to `"auto"`,
  `"auto"` passes through, `"manual"` is handled by
  `shouldSkipTool` (not `resolveMethod`).
- [x] 1.4 Add tests for `installViaGo()` in
  `internal/setup/setup_test.go` (table-driven per
  TC-006): Go available and install succeeds (assert
  `action: "installed"`, `detail: "via go install"`);
  Go available but install fails (assert
  `action: "failed"`, `detail` contains module path
  and retry hint, `err` is non-nil and wrapped);
  Go not available (assert `action: "skipped"`,
  `detail` contains install hint, `err` is nil);
  dry-run mode (assert `detail` matches
  `"Would install: go install <module>@latest"`).

## 2. Update installGaze() Auto-Mode Fallback

- [x] 2.1 Update `installGaze()` (setup.go:519) to call
  `resolveMethod("gaze", env)` instead of
  `opts.toolMethod("gaze")`. Add dispatch cases for
  `"go"` (calls `installViaGo` with module
  `github.com/unbound-force/gaze/cmd/gaze`).
  Update auto-mode block: after Homebrew-absent check,
  try dnf via `installViaRpm()` when
  `HasManager(env, ManagerDnf)`, then try `installViaGo`
  as final fallback before skip. Update dry-run to
  reflect new fallback chain.
- [x] 2.2 Add tests for `installGaze()` fallback chain:
  dnf available + no Homebrew -> installs via RPM;
  no Homebrew + no dnf + Go available -> installs via
  `go install`; no Homebrew + no dnf + no Go -> skips;
  `PackageManager: "dnf"` -> skips Homebrew, uses dnf.

## 3. Update installReplicator() with Dispatch + Fallback

- [x] 3.1 Add `resolveMethod("replicator", env)` dispatch
  to `installReplicator()` (setup.go:721) with explicit
  `"rpm"`/`"dnf"` case (calls `installViaRpm` with repo
  `unbound-force/replicator`), `"homebrew"` case, and
  `"go"` case (calls `installViaGo` with module
  `github.com/unbound-force/replicator/cmd/replicator`).
  Follow the `installGaze()` dispatch pattern at
  setup.go:525-538.
- [x] 3.2 Add auto-mode fallback chain: Homebrew -> dnf
  (via `installViaRpm`) -> `go install` (via
  `installViaGo`) -> skip. Update dry-run to reflect
  new fallback chain.
- [x] 3.3 Add tests (table-driven per TC-006): same
  fallback chain as Gaze (Homebrew -> dnf -> go install
  -> skip). Scenarios: dnf available + no Homebrew ->
  installs via RPM; no Homebrew + no dnf + Go available
  -> installs via `go install`; no Homebrew + no dnf +
  no Go -> skips; `PackageManager: "dnf"` -> skips
  Homebrew, uses dnf.

## 4. Update installDewey() with Dispatch + Fallback

- [x] 4.1 Add `resolveMethod("dewey", env)` dispatch to
  `installDewey()` (setup.go:1128) with explicit
  `"homebrew"` and `"go"` cases. No `"rpm"`/`"dnf"`
  case — Dewey does not publish RPMs. If
  `resolveMethod` returns `"dnf"`, fall through to
  `"go"` since dnf is not available for Dewey.
- [x] 4.2 Add auto-mode fallback chain: Homebrew ->
  `go install` (via `installViaGo` with module
  `github.com/unbound-force/dewey/cmd/dewey`) -> skip.
  Preserve the `pullEmbeddingModel()` call after any
  successful install path.
- [x] 4.3 Add tests (table-driven per TC-006):
  no Homebrew + Go available -> installs via `go install`
  + pulls embedding model (assert `action: "installed"`,
  `detail: "via go install"`);
  no Homebrew + no Go -> skips (assert
  `action: "skipped"`);
  `PackageManager: "dnf"` -> `resolveMethod` returns
  `"dnf"`, `installDewey` has no `"dnf"` case, falls
  through to `go install` (assert `detail` contains
  `"via go install"`).

## 5. Update installGH() with Dispatch + Fallback

- [x] 5.1 Add `resolveMethod("gh", env)` dispatch to
  `installGH()` (setup.go:470) with explicit `"dnf"`
  case (calls `ExecCmd("dnf", "install", "-y", "gh")`
  directly — GH CLI has its own dnf repo, not
  GoReleaser RPMs) and `"homebrew"` case.
- [x] 5.2 Add auto-mode fallback chain: Homebrew -> dnf
  (`dnf install -y gh`) -> skip. On `dnf install gh`
  failure, return `action: "skipped"` (graceful
  degradation, not hard failure), `detail` with
  actionable message including GitHub CLI repo setup
  link, `err: nil`. Update dry-run to reflect new
  fallback chain.
- [x] 5.3 Add tests (table-driven per TC-006):
  dnf available + no Homebrew -> attempts
  `dnf install gh` (assert `action: "installed"`);
  dnf install fails -> skips with actionable download
  link (assert `action: "skipped"`, NOT `"failed"`,
  `detail` contains repo setup URL, `err` is nil);
  `PackageManager: "dnf"` -> skips Homebrew, uses dnf
  directly.

## 6. Dry-Run Path Updates

- [x] 6.1 Verify dry-run output in all four updated
  install functions reflects the new fallback chain.
  When Homebrew is absent, dry-run should show the
  next available method (dnf or `go install`) instead
  of "Would install: download from...".
- [x] 6.2 Add dry-run tests: verify each fallback tier
  produces the correct dry-run detail string for all
  four tools.

## 7. Verification

- [x] 7.1 Run `make check` — all tests pass, lint clean,
  build succeeds.
- [x] 7.2 Run `go test -race -count=1 ./internal/setup/`
  — verify new tests pass in isolation. Confirm existing
  `TestInstallViaRpm_*` tests (Success, NoVersion,
  DnfFails, DryRun) pass unchanged — no modifications
  to `installViaRpm()` signature or behavior.
- [x] 7.3 Manual smoke test: `uf setup --dry-run` on a
  system without Homebrew — verify tools show dnf or
  `go install` intent instead of "skipped".

## 8. Documentation

- [x] 8.1 Add CHANGELOG.md entry documenting the fix
  for issue #214: `uf setup` auto mode now falls back
  to dnf and `go install` when Homebrew is absent.

## 9. Constitution Alignment Verification

- [x] 9.1 Verify Composability First: each tool's
  fallback chain degrades gracefully — no tool requires
  a specific package manager. Standalone installation
  is preserved via `go install` as the universal
  fallback for Go-based tools.
- [x] 9.2 Verify Testability: all new code uses
  injectable dependencies (`LookPath`, `ExecCmd` on
  the Options struct). No tests require network access,
  external services, or shared mutable state. All tests
  verify observable side effects (returned `stepResult`
  values).
<!-- spec-review: passed -->
