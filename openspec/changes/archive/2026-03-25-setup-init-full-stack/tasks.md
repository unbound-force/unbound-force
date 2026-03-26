## 1. Setup: New Install Functions

- [x] 1.1 Add `installMxF(opts *Options, env doctor.DetectedEnvironment) stepResult` to `internal/setup/setup.go` following the `installGaze()` pattern exactly (including DryRun branches): check `LookPath("mxf")`, DryRun + Homebrew/no-Homebrew, try `brew install unbound-force/tap/mxf`, skip with GitHub releases link if no Homebrew
- [x] 1.2 Add `installGH(opts *Options, env doctor.DetectedEnvironment) stepResult` to `internal/setup/setup.go` following the `installGaze()` pattern exactly (including DryRun branches): check `LookPath("gh")`, try `brew install gh`, skip with `https://cli.github.com` link if no Homebrew
- [x] 1.3 Add `installOpenSpec(opts *Options, env doctor.DetectedEnvironment) stepResult` to `internal/setup/setup.go` following the `installSwarmPlugin()` pattern (including DryRun and bun preference): check `LookPath("openspec")`, prefer `bun add -g @fission-ai/openspec@latest`, fall back to `npm install -g @fission-ai/openspec@latest`. Include actionable error message for EACCES failures.

## 2. Setup: Dewey Init Steps

- [x] 2.1 Add `initDewey(opts *Options) stepResult` to `internal/setup/setup.go`: check if `.dewey/` exists (skip if so), check `LookPath("dewey")` (skip if not found), check `DryRun`, run `dewey init`. Non-interactive, no TTY guard needed.
- [x] 2.2 Add `indexDewey(opts *Options) stepResult` to `internal/setup/setup.go`: check if `.dewey/` exists (skip if not), check `LookPath("dewey")` (skip if not found), check `DryRun`, run `dewey index`. Include Ollama server hint in error message.

## 3. Setup: Updated Step Ordering

- [x] 3.1 Insert `installMxF()` call in `Run()` after Gaze (Step 2), renumber as Step 3
- [x] 3.2 Insert `installGH()` call in `Run()` after Mx F, renumber as Step 4
- [x] 3.3 Insert `installOpenSpec()` call in `Run()` after Bun and before Swarm plugin (inside the `nodeAvailable` block), renumber as Step 7
- [x] 3.4 Insert `initDewey()` call in `Run()` after Dewey install + model pull, renumber as Step 14
- [x] 3.5 Insert `indexDewey()` call in `Run()` after `initDewey()`, renumber as Step 15
- [x] 3.6 Update all step number comments in `Run()` to reflect the new 16-step ordering (steps 5-11 are inside nodeAvailable conditional)

## 4. Scaffold: Options Expansion

- [x] 4.1 Add `LookPath func(string) (string, error)` and `ExecCmd func(string, ...string) ([]byte, error)` fields to `scaffold.Options` in `internal/scaffold/scaffold.go`
- [x] 4.2 In `Run()`, default `LookPath` to `exec.LookPath` and `ExecCmd` to a `exec.Command(...).CombinedOutput()` wrapper if nil. This MUST happen at the top of `Run()` before any code path that calls `initSubTools()`.
- [x] 4.3 Verify all existing scaffold tests pass after Options expansion: `go test -race -count=1 ./internal/scaffold/...`

## 5. Scaffold: Sub-Tool Initialization

- [x] 5.1 Add `initSubTools(opts *Options) []subToolResult` function to `internal/scaffold/scaffold.go`: skip if `DivisorOnly`. If `LookPath("dewey")` succeeds and `.dewey/` doesn't exist, run `dewey init` (check error before proceeding to index) + `dewey index`. Return list of results for printSummary.
- [x] 5.2 Call `initSubTools()` in `Run()` after file scaffolding and before `printSummary()`

## 6. Scaffold: Updated printSummary

- [x] 6.1 Update `printSummary()` signature in `internal/scaffold/scaffold.go` to accept sub-tool results (e.g., `subToolResults []subToolResult`). Update ALL existing call sites (2 production, 4+ test).
- [x] 6.2 Add next-step guidance to `printSummary()`: pre-compute tool availability in `Run()` using `opts.LookPath` for `dewey`, `openspec`, `opencode`. Pass as boolean flags or a struct. Show constitution/doctor/speckit/opsx steps when tools available; show `uf setup` as first step when tools missing.

## 7. Setup: Forward Injected Functions

- [x] 7.1 Update `runUnboundInit()` in `internal/setup/setup.go` to forward `opts.LookPath` and `opts.ExecCmd` to the `scaffold.Options` struct, maintaining the testability injection chain

## 8. Setup Tests

- [x] 8.1 Update `TestSetupRun_AllMissing` in `internal/setup/setup_test.go`: add `"brew install unbound-force/tap/mxf"`, `"brew install gh"`, `"bun add -g @fission-ai/openspec@latest"` or `"npm install -g @fission-ai/openspec@latest"`, `"dewey init"`, `"dewey index"` to expected commands
- [x] 8.2 Update `TestSetupRun_AllPresent` in `internal/setup/setup_test.go`: add `"mxf"`, `"gh"`, `"openspec"` to LookPath stub, add `.dewey/` directory to temp dir
- [x] 8.3 Add `TestSetupRun_MxFMissing_BrewInstall`: verify `brew install unbound-force/tap/mxf` called when mxf is missing
- [x] 8.4 Add `TestSetupRun_MxFNoHomebrew`: verify mxf install is skipped with GitHub releases link when Homebrew is not available
- [x] 8.5 Add `TestSetupRun_GHMissing_BrewInstall`: verify `brew install gh` called when gh is missing
- [x] 8.6 Add `TestSetupRun_GHNoHomebrew`: verify gh install is skipped with `https://cli.github.com` link when Homebrew is not available
- [x] 8.7 Add `TestSetupRun_OpenSpecMissing_Install`: verify openspec installed via bun or npm when missing and Node.js available
- [x] 8.8 Add `TestSetupRun_OpenSpecNpmFails`: verify graceful handling when npm install fails (actionable error message)
- [x] 8.9 Add `TestSetupRun_DeweyInit`: verify `dewey init` called when `.dewey/` doesn't exist
- [x] 8.10 Add `TestSetupRun_DeweyInitFails`: verify graceful handling when `dewey init` returns error, and `dewey index` is skipped
- [x] 8.11 Add `TestSetupRun_DeweyIndex`: verify `dewey index` called after `.dewey/` exists
- [x] 8.12 Add `TestSetupRun_DeweyIndexFails`: verify graceful handling when `dewey index` returns error
- [x] 8.13 Update `TestSetupRun_DryRun`: verify dry-run output includes "Would install" for mxf, gh, openspec, and "Would run" for dewey init/index

## 9. Scaffold Tests

- [x] 9.1 Add `TestInitSubTools_DeweyAvailable` in `internal/scaffold/scaffold_test.go`: stub LookPath with dewey, verify `dewey init` + `dewey index` called, check error handling
- [x] 9.2 Add `TestInitSubTools_DeweyNotAvailable` in `internal/scaffold/scaffold_test.go`: stub LookPath without dewey, verify no subprocess calls
- [x] 9.3 Add `TestInitSubTools_DeweyAlreadyInitialized` in `internal/scaffold/scaffold_test.go`: create `.dewey/` dir, verify `dewey init` is NOT called
- [x] 9.4 Add `TestInitSubTools_DeweyInitFails` in `internal/scaffold/scaffold_test.go`: stub ExecCmd to fail on `dewey init`, verify `dewey index` is NOT called, verify warning in results
- [x] 9.5 Add `TestInitSubTools_DivisorOnly` in `internal/scaffold/scaffold_test.go`: set `DivisorOnly: true`, verify `initSubTools` returns nil (skips all sub-tool init)
- [x] 9.6 Update all existing `printSummary` test calls to match the new function signature. Verify existing assertions still pass. Verify next-step guidance includes constitution and doctor steps.

## 10. Documentation

- [x] 10.1 Update `AGENTS.md` Recent Changes section: add entry for this change describing the setup/init improvements (new tool installations, dewey init/index, printSummary guidance)
- [x] 10.2 Update `AGENTS.md` description of `uf setup` behavior if mentioned in Build & Test Commands or other sections

## 11. Verify

- [x] 11.1 Run `go build ./...` to verify compilation
- [x] 11.2 Run `go test -race -count=1 ./internal/setup/...` to verify setup tests pass
- [x] 11.3 Run `go test -race -count=1 ./internal/scaffold/...` to verify scaffold tests pass
- [x] 11.4 Run `go test -race -count=1 ./...` to verify full test suite passes
