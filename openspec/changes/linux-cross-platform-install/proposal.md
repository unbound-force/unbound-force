## Why

The `unbound-force` CLI cannot be installed via Homebrew on
Linux. GoReleaser builds Linux binaries and uploads them to
GitHub Releases, but the Homebrew distribution is configured
as a Cask (`homebrew_casks:`), which is macOS-only by
Homebrew convention. Linux users attempting
`brew install unbound-force/tap/unbound-force` receive a
platform rejection error.

Additionally, `uf setup` installs Ollama via
`brew install --cask ollama-app`, which also fails on Linux.

This blocks adoption on Fedora, RHEL, and any Linux
distribution. The binaries exist -- they just lack a
delivery channel.

## What Changes

1. Add GoReleaser `nfpms:` configuration to produce RPM
   packages for `unbound-force`, uploaded to GitHub Releases
   alongside existing tar.gz archives.

2. Add GoReleaser `brews:` configuration to produce a
   Homebrew Formula (cross-platform) alongside the existing
   Cask (macOS-only). Same `brew install` command works on
   both macOS and Linux.

3. Update the CI release workflow to push the Formula to
   the homebrew-tap repository, with darwin checksum
   patching by the sign-macos job (same pattern as the
   existing Cask).

4. Add `dnf` detection to `uf doctor` environment scanning
   so the tool chain is aware of the native Fedora/RHEL
   package manager.

5. Update `uf setup` to install tools via `dnf install`
   from GitHub Release RPM URLs when Homebrew is not
   available and `dnf` is detected.

6. Fix `installOllama()` in `uf setup` to be OS-aware:
   use `brew install --cask ollama-app` on macOS and
   `brew install ollama` (formula) on Linux.

7. Update install hint strings in `uf doctor` to be
   OS-aware for Ollama and to suggest `dnf install` when
   dnf is the detected package manager.

## Capabilities

### New Capabilities
- `rpm-distribution`: GoReleaser produces `.rpm` packages
  for `unbound-force`, enabling `dnf install <url>.rpm`
  on Fedora/RHEL.
- `homebrew-formula`: Cross-platform Homebrew Formula
  enables `brew install unbound-force/tap/unbound-force`
  on Linux (Linuxbrew).
- `dnf-detection`: `uf doctor` detects `dnf` as a package
  manager and reports it in the environment scan.
- `dnf-install-path`: `uf setup` can install
  `unbound-force` via `dnf install` from GitHub Release
  RPM URLs when dnf is available.

### Modified Capabilities
- `installOllama`: OS-aware -- uses cask on macOS,
  formula on Linux.
- `install-hints`: OS-aware hints for Ollama; dnf-aware
  hints when dnf is detected.
- `release-workflow`: Generates and pushes both a Cask
  and a Formula to the homebrew-tap.

### Removed Capabilities
- None.

## Impact

- `.goreleaser.yaml`: New `nfpms:` and `brews:` sections.
- `.github/workflows/release.yml`: Upload and patch
  Formula alongside Cask in sign-macos job.
- `internal/doctor/models.go`: New `ManagerDnf` constant.
- `internal/doctor/environ.go`: dnf detection, OS-aware
  install hints for Ollama.
- `internal/setup/setup.go`: dnf install path, OS-aware
  Ollama install.
- `internal/doctor/*_test.go`: Tests for dnf detection.
- `internal/setup/*_test.go`: Tests for dnf install path
  and OS-aware Ollama.
- `unbound-force/homebrew-tap`: Will receive a new
  `Formula/` directory with `unbound-force.rb` on next
  release (no code change in this repo).

Scope is limited to `unbound-force` only. Sibling repos
(gaze, dewey, replicator) can adopt the same pattern
later.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change does not affect inter-hero artifact formats,
communication protocols, or metadata. It modifies only the
distribution and installation mechanism.

### II. Composability First

**Assessment**: PASS

This change directly supports Composability First by
removing a platform barrier to standalone installation.
Linux users can install `unbound-force` independently
without macOS. The RPM and Formula are additive channels
-- the existing Cask continues to work unchanged on macOS.
No mandatory dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

`uf doctor` gains `dnf` detection, improving the
machine-parseable environment report with the new
manager kind. Install hints become OS-aware, improving
diagnostic accuracy. No existing output formats change.

### IV. Testability

**Assessment**: PASS

All new code follows the existing injectable dependency
pattern (`LookPath`, `ExecCmd`, `Getenv` on Options
structs). The `dnf` detection and install paths are
testable in isolation via the same function injection
used for Homebrew detection. No external services or
network access required for tests.
