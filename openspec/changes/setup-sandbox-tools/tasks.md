<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Doctor: Install Hints and URLs

- [x] 1.1 [P] Add `podman` and `devpod` entries to
  `homebrewInstallCmd`, `genericInstallCmd`, and
  `installURL` in `internal/doctor/environ.go`.
  - `homebrewInstallCmd("podman")` -> `"brew install podman"`
  - `homebrewInstallCmd("devpod")` -> `"brew install devpod"`
  - `genericInstallCmd("podman")` -> download URL
  - `genericInstallCmd("devpod")` -> download URL
  - `installURL("podman")` -> `"https://podman.io"`
  - `installURL("devpod")` -> `"https://devpod.sh"`

## 2. Doctor: Podman Core Tool Check

- [x] 2.1 Add Podman to `coreToolSpecs` in
  `internal/doctor/checks.go` as a required tool with
  version check:
  - `name: "podman"`, `required: true`
  - `versionCmd: []string{"podman", "--version"}`
  - Add `parsePodmanVersion` function (parses
    "podman version X.Y.Z" output)
  - Add `checkPodmanVersion` function (>= 4.3)
  - `minVersion: "4.3"`
- [x] 2.2 Add `GOOS string` field to doctor `Options`
  struct in `internal/doctor/doctor.go`. When non-empty,
  it overrides `runtime.GOOS`. When empty, defaults to
  `runtime.GOOS`. Add a helper method
  `func (o *Options) goos() string` that returns the
  resolved value. Update `defaults()` if needed.
- [x] 2.3 Add Podman runtime health post-check in
  `checkCoreTools()` in `internal/doctor/checks.go`,
  following the Ollama post-check pattern:
  - After Podman passes presence + version, run
    `podman info` via `opts.ExecCmd`
  - On macOS (`opts.goos() == "darwin"`): first check
    `podman machine list --format '{{.Name}}'` for
    machine existence; if no machine, Fail with hint
    "podman machine init && podman machine start";
    if machine exists but `podman info` fails, Fail
    with hint "podman machine start"
  - On Linux: run `podman info` directly; if it fails,
    Fail with hint referencing
    "systemctl --user status podman.socket"
  - If `podman info` succeeds on any platform, Pass
    with message "running"
- [x] 2.4 [P] Add tests for `parsePodmanVersion` and
  `checkPodmanVersion` in
  `internal/doctor/doctor_test.go`:
  - Parse "podman version 5.3.1"
  - Parse "podman version 4.3.0"
  - Reject "podman version 4.2.9"
  - Reject unparseable output
- [x] 2.5 [P] Add test for Podman in `coreToolSpecs`:
  Podman missing -> Fail; Podman present, version
  sufficient -> Pass; Podman present, version too old
  -> Fail.
- [x] 2.6 [P] Add tests for GOOS injection and Podman
  runtime post-check in
  `internal/doctor/doctor_test.go`:
  - GOOS override works (set "darwin", verify macOS path)
  - GOOS default (empty -> runtime.GOOS)
  - macOS: machine exists + podman info succeeds -> Pass
  - macOS: no machine -> Fail with machine init hint
  - macOS: machine exists + podman info fails -> Fail
    with machine start hint
  - Linux: podman info succeeds -> Pass
  - Linux: podman info fails -> Fail with socket hint
  - Version too old -> runtime check skipped

## 3. Doctor: Enhanced DevPod Checks

- [x] 3.1 Add DevPod version check to `checkDevPod()`
  in `internal/doctor/checks.go`:
  - Run `devpod version`, parse output (format:
    `v0.X.Y`, strip leading `v`, split on dots,
    handle pre-release suffixes like `-beta`)
  - Add `parseDevPodVersion` function following doctor
    `versionParse` pattern (returns `(string, error)`)
  - Add `checkDevPodVersionMin` function
  - Minimum version 0.5.0; Warn if below
- [x] 3.2 Add DevPod provider check to `checkDevPod()`
  in `internal/doctor/checks.go`:
  - Run `devpod provider list`, use exact first-column
    name matching for "podman" (not substring)
  - If `devpod provider list` fails, skip with warning
  - Warn if missing with install hint:
    `devpod provider add docker --name podman -o DOCKER_PATH=podman`
- [x] 3.3 [P] Add tests for enhanced DevPod checks in
  `internal/doctor/doctor_test.go`:
  - `parseDevPodVersion`: parse "v0.6.15", "0.5.0",
    "v0.4.2-beta", reject unparseable
  - Version sufficient (0.6.15) -> Pass
  - Version too old (0.4.2) -> Warn
  - Version boundary (0.5.0) -> Pass
  - Provider present (exact match "podman") -> Pass
  - Provider name substring ("podman-custom") -> not
    matched (provider missing)
  - Provider missing -> Warn with hint
  - Provider list command fails -> skip with warning

## 4. Setup: Podman Installation

- [x] 4.1 Add `installPodman()` function to
  `internal/setup/setup.go`:
  - LookPath check -> "already installed"
  - Homebrew: `brew install podman`
  - No Homebrew: skip with download URL
  - macOS post-install: check for Podman machine via
    `podman machine list`, run `podman machine init`
    (with 180-second timeout) and `podman machine start`
    if no machine exists. Machine init/start failures
    are reported in detail but do not fail the step.
  - Post-install smoke test: run `podman info` to
    verify installation is functional. If it fails,
    detail includes "podman info failed" but the step
    reports "installed". If it passes, detail includes
    "verified".
  - DryRun support
  - shouldSkipTool("podman") support
- [x] 4.2 [P] Add tests for `installPodman()` in
  `internal/setup/setup_test.go`:
  - Already installed
  - Fresh install via Homebrew
  - No Homebrew -> skip
  - macOS machine init (happy path)
  - macOS machine already exists
  - macOS machine init fails -> "installed" +
    "machine init failed"
  - macOS machine start fails -> "installed" +
    "machine start failed"
  - Linux (no machine init)
  - Smoke test passes -> "installed" + "verified"
  - Smoke test fails -> "installed" + "podman info
    failed"
  - DryRun mode
  - Skip via config

## 5. Setup: DevPod Installation

- [x] 5.1 Add `installDevPod()` function to
  `internal/setup/setup.go`:
  - LookPath check -> "already installed"
  - Homebrew: `brew install devpod`
  - No Homebrew: skip with download URL
  - DryRun support
  - shouldSkipTool("devpod") support
- [x] 5.2 [P] Add tests for `installDevPod()` in
  `internal/setup/setup_test.go`:
  - Already installed
  - Fresh install via Homebrew
  - No Homebrew -> skip
  - DryRun mode
  - Skip via config

## 6. Setup: DevPod Provider Configuration

- [x] 6.1 Add `configureDevPodProvider()` function to
  `internal/setup/setup.go`:
  - Gate on both devpod and podman being available
  - Check `devpod provider list` for "podman" provider
    using exact first-column name matching
  - If `devpod provider list` fails, skip with warning
  - If provider missing: run
    `devpod provider add docker --name podman -o DOCKER_PATH=podman`
  - If provider add fails, report "failed" with the
    manual command as detail
  - DryRun support
- [x] 6.2 [P] Add tests for `configureDevPodProvider()`
  in `internal/setup/setup_test.go`:
  - Provider already registered -> "already installed"
  - Provider missing -> install
  - Provider add fails -> "failed" with manual hint
  - Provider list fails -> "skipped" with warning
  - DevPod not available -> skip
  - Podman not available -> skip
  - DryRun mode

## 7. Setup: Run() Integration

- [x] 7.1 Update `Run()` in `internal/setup/setup.go`:
  - Insert Podman step after Ollama (step 11/16)
  - Insert DevPod step (step 12/16)
  - Insert DevPod provider step (step 13/16),
    gated on devpod + podman availability
  - Renumber Dewey to 14/16, golangci-lint to 15/16,
    govulncheck to 16/16
  - Update all 13 existing `[N/13]` labels to `[N/16]`
  - Note: `coreToolSpecs` comment says "8 binaries"
    but slice has 7 entries; after adding Podman it
    will be 8, making the comment accurate
- [x] 7.2 [P] Update existing setup tests that assert
  step count or step labels:
  - Search for `/13]` in `setup_test.go` -> `/16]`
  - Search for step count assertions (e.g.,
    `len(results) == 13`) -> update to 16
  - Add `ExecCmd` mock fallbacks for `podman`,
    `devpod`, `devpod provider list` in the shared
    test helper to prevent cascading failures when
    `Run()` invokes new steps

## 8. Verification

- [x] 8.1 Run `go test -race -count=1 ./internal/doctor/`
  and `go test -race -count=1 ./internal/setup/`
- [x] 8.2 Run `go vet ./...` and `golangci-lint run`
- [x] 8.3 Verify constitution alignment: all new
  functions use Options DI pattern (Principle IV),
  each tool is independently skippable (Principle II),
  doctor output includes machine-parseable severity
  and hints (Principle III).
- [x] 8.4 Verify platform-aware runtime checks: confirm
  doctor tests cover both macOS and Linux code paths
  for Podman runtime health (machine check vs direct
  podman info). Verify GOOS override is injectable
  in doctor Options for test isolation.
- [x] 8.5 File `unbound-force/website` issue for
  documentation updates: setup step inventory (13->16),
  doctor check groups (Podman in Core Tools, enhanced
  DevPod group), and new install hints.
<!-- spec-review: passed -->
<!-- code-review: passed -->
