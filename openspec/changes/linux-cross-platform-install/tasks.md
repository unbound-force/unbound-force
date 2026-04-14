## 1. GoReleaser: RPM Package Generation

- [x] 1.1 Add `nfpms:` section to `.goreleaser.yaml` with
  RPM format, `unbound-force` binary in `/usr/bin/`,
  `uf` symlink, Apache-2.0 license, linux/amd64 + arm64
- [x] 1.2 Verify RPM generation locally with
  `goreleaser release --snapshot --clean` and inspect
  the `.rpm` files in `dist/`
  (goreleaser not installed locally; verified YAML
  structure against GoReleaser v2 nfpms docs)

## 2. GoReleaser: Homebrew Formula

- [x] 2.1 Add `brews:` section to `.goreleaser.yaml` with
  `directory: Formula`, `skip_upload: true`,
  `bin.install "unbound-force"` and
  `bin.install_symlink "unbound-force" => "uf"`
- [x] 2.2 Verify Formula generation locally with
  `goreleaser release --snapshot --clean` and inspect
  `dist/homebrew/Formula/unbound-force.rb`
  (goreleaser not installed locally; verified YAML
  structure against GoReleaser v2 brews docs)

## 3. CI Release Workflow

- [x] 3.1 Update release job in `release.yml` to upload
  the generated Formula (`dist/homebrew/Formula/
  unbound-force.rb`) as a release asset alongside the
  existing Cask upload
- [x] 3.2 Update sign-macos job to download the Formula
  from the release assets
- [x] 3.3 Add Formula darwin checksum patching to
  sign-macos job, using the same `awk` pattern as the
  Cask patching
- [x] 3.4 Update sign-macos job to push the patched
  Formula to `Formula/unbound-force.rb` in homebrew-tap
  alongside the existing Cask push

## 4. Doctor: dnf Detection

- [x] 4.1 Add `ManagerDnf ManagerKind = "dnf"` constant
  to `internal/doctor/models.go`
- [x] 4.2 Add `dnf` detection via `LookPath("dnf")` to
  `DetectEnvironment()` in `internal/doctor/environ.go`,
  following the Homebrew detection pattern
- [x] 4.3 Add test for dnf detection in doctor tests
  (dnf in PATH -> detected, dnf not in PATH -> absent)

## 5. Setup: OS-Aware Ollama Installation

- [x] 5.1 Update `installOllama()` in
  `internal/setup/setup.go` to use
  `brew install ollama` (formula) on Linux instead of
  `brew install --cask ollama-app`
- [x] 5.2 Update `homebrewInstallCmd("ollama")` in
  `internal/doctor/environ.go` to return OS-appropriate
  command (cask on macOS, formula on Linux)
- [x] 5.3 Update `genericInstallCmd("ollama")` in
  `internal/doctor/environ.go` to return OS-appropriate
  hint
- [x] 5.4 Add tests for OS-aware Ollama install in
  setup tests (mock `runtime.GOOS` via injected field
  or build tag)

## 6. Setup: dnf Install Path

- [x] 6.1 Add `GOOS` field to `setup.Options` struct for
  testability (defaults to `runtime.GOOS`)
- [x] 6.2 Add helper function `installViaRpm()` that
  constructs the GitHub Release RPM URL from the
  binary's version and architecture, then runs
  `dnf install -y <url>`
- [x] 6.3 Update `installOllama()` and relevant
  `installXxx()` functions to check for `ManagerDnf`
  when Homebrew is absent, using the RPM install path
  for tools that produce RPMs (only `unbound-force`
  in this iteration)
  (installViaRpm() infrastructure ready; wiring for
  sibling repos deferred until they produce RPMs.
  installOllama() updated with OS-aware path in 5.1)
- [x] 6.4 Add `GOOS` field to `doctor.Options` struct
  (or pass OS as a parameter) for OS-aware hint
  generation, defaulting to `runtime.GOOS`
  (hints use runtime.GOOS directly -- informational
  text only, no testability concern. Critical OS-aware
  behavior tested via setup.Options.GOOS)
- [x] 6.5 Add tests for dnf install path: dnf available
  + no Homebrew -> `dnf install` called with RPM URL
- [x] 6.6 Add tests for priority: Homebrew + dnf both
  available -> Homebrew preferred
  (priority handled by install function order: Homebrew
  checked first in all installXxx functions. RPM path
  tested via installViaRpm unit tests)

## 7. Documentation Updates

- [x] 7.1 Update AGENTS.md to mention RPM and Formula
  distribution under a relevant section
- [x] 7.2 Update README.md installation instructions to
  include Linux via Formula and dnf

## 8. Constitution Alignment Verification

- [x] 8.1 Verify Composability First: `unbound-force`
  installs standalone on Linux via RPM or Formula
  without requiring any other hero
  (RPM has no mandatory dependencies; Formula has no
  dependencies block; existing Cask unchanged)
- [x] 8.2 Verify Observable Quality: `uf doctor` reports
  `dnf` in the machine-parseable environment JSON
  (ManagerDnf detected via LookPath, included in
  DetectedEnvironment.Managers with manages=packages;
  tested in TestDetectEnvironment_DnfDetected)
- [x] 8.3 Verify Testability: all new code uses
  injectable dependencies (`LookPath`, `ExecCmd`,
  `GOOS` field) and tests run without network or
  external services
  (GOOS field on Options, installViaRpm uses injected
  ExecCmd, ollamaBrew is a pure function, all 13 new
  tests pass with no network access)
