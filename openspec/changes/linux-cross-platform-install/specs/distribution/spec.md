## ADDED Requirements

### Requirement: RPM Package Generation

GoReleaser MUST produce RPM packages for the
`unbound-force` binary via the `nfpms:` configuration.
The RPM MUST include:
- The `unbound-force` binary installed to `/usr/bin/`
- A `uf` symlink at `/usr/bin/uf` pointing to
  `/usr/bin/unbound-force`
- Apache-2.0 license metadata
- Package description matching the Cask description

RPM packages MUST be generated for `linux/amd64` and
`linux/arm64` architectures.

#### Scenario: RPM uploaded to GitHub Releases

- **GIVEN** a tagged release is pushed (`v*`)
- **WHEN** GoReleaser runs the release pipeline
- **THEN** `.rpm` files for linux/amd64 and linux/arm64
  are uploaded to the GitHub Release alongside the
  existing tar.gz archives

#### Scenario: RPM installs correctly via dnf

- **GIVEN** a Fedora/RHEL system with `dnf` available
- **WHEN** the user runs
  `dnf install <github-release-url>.rpm`
- **THEN** `unbound-force` is available at `/usr/bin/`
  and `uf` symlink is created at `/usr/bin/uf`

---

### Requirement: Homebrew Formula

GoReleaser MUST produce a Homebrew Formula via the
`brews:` configuration. The Formula MUST:
- Be placed in the `Formula/` directory of the
  `unbound-force/homebrew-tap` repository
- Include install instructions for the `unbound-force`
  binary and `uf` symlink
- Support both macOS and Linux platforms
- Use `skip_upload: true` to allow CI checksum patching

The existing Cask MUST NOT be modified or removed.

#### Scenario: Formula install on Linux

- **GIVEN** a Linux system with Homebrew (Linuxbrew)
- **WHEN** the user runs
  `brew install unbound-force/tap/unbound-force`
- **THEN** the Formula is used, and both `unbound-force`
  and `uf` are available in the PATH

#### Scenario: Formula install on macOS

- **GIVEN** a macOS system with Homebrew
- **WHEN** the user runs
  `brew install unbound-force/tap/unbound-force`
- **THEN** the Formula is used (Homebrew prefers Formula
  over Cask for `brew install` without `--cask`)

#### Scenario: Cask remains available

- **GIVEN** a macOS system with Homebrew
- **WHEN** the user runs
  `brew install --cask unbound-force/tap/unbound-force`
- **THEN** the Cask is used, providing signed binaries
  with quarantine removal

---

### Requirement: CI Formula Checksum Patching

The sign-macos CI job MUST patch the Formula's darwin
checksums with signed binary values, following the same
pattern used for the Cask.

The release job MUST upload the generated Formula as a
release asset (alongside the Cask) so the sign-macos job
can download and patch it.

The sign-macos job MUST push the patched Formula to
`Formula/unbound-force.rb` in the homebrew-tap repository.

#### Scenario: Formula has correct signed checksums

- **GIVEN** a release with macOS code signing enabled
- **WHEN** the sign-macos job completes
- **THEN** `Formula/unbound-force.rb` in homebrew-tap
  contains sha256 values matching the signed darwin
  archives

---

### Requirement: dnf Package Manager Detection

`uf doctor` MUST detect `dnf` as a package manager via
`LookPath("dnf")`. A new constant `ManagerDnf` of type
`ManagerKind` MUST be added with value `"dnf"`.

The detection MUST follow the existing pattern used for
Homebrew detection (PATH-based, no environment variable
fallback required).

The detected manager MUST appear in the doctor
environment report with `manages: ["packages"]`.

#### Scenario: dnf detected on Fedora

- **GIVEN** a Fedora system with `dnf` in PATH
- **WHEN** `uf doctor` runs environment detection
- **THEN** the report includes a manager entry with
  `kind: "dnf"` and `manages: ["packages"]`

#### Scenario: dnf not detected on macOS

- **GIVEN** a macOS system without `dnf` in PATH
- **WHEN** `uf doctor` runs environment detection
- **THEN** no `dnf` manager entry appears in the report

---

### Requirement: dnf Install Path in Setup

`uf setup` MUST support installing `unbound-force` via
`dnf install <rpm-url>` when:
- `dnf` is detected as a package manager, AND
- Homebrew is NOT detected

The RPM URL MUST be constructed from the GitHub Releases
URL pattern using the current binary version.

Tools that do not produce RPMs (gaze, dewey, replicator,
mxf, ollama) SHOULD be skipped with an appropriate
install hint when only `dnf` is available.

#### Scenario: Setup with dnf, no Homebrew

- **GIVEN** a Fedora system with `dnf` but without
  Homebrew
- **WHEN** the user runs `uf setup`
- **THEN** tools with RPM packages are installed via
  `dnf install <url>.rpm`
- **AND** tools without RPMs are skipped with download
  hints

#### Scenario: Setup with both Homebrew and dnf

- **GIVEN** a Linux system with both Homebrew and `dnf`
- **WHEN** the user runs `uf setup`
- **THEN** Homebrew is preferred (existing behavior)

---

### Requirement: OS-Aware Ollama Installation

`installOllama()` in `uf setup` MUST select the install
method based on the operating system:
- macOS: `brew install --cask ollama-app` (unchanged)
- Linux with Homebrew: `brew install ollama` (formula)
- Linux without Homebrew: skip with install hint

The Ollama Homebrew formula (`ollama`) MUST be used
instead of the Cask (`ollama-app`) on Linux.

#### Scenario: Ollama install on Linux with Homebrew

- **GIVEN** a Linux system with Homebrew
- **WHEN** `uf setup` runs the Ollama install step
- **THEN** `brew install ollama` is executed (not
  `--cask ollama-app`)

#### Scenario: Ollama install on macOS

- **GIVEN** a macOS system with Homebrew
- **WHEN** `uf setup` runs the Ollama install step
- **THEN** `brew install --cask ollama-app` is executed
  (unchanged behavior)

---

### Requirement: OS-Aware Install Hints

Install hint functions (`homebrewInstallCmd`,
`genericInstallCmd`) MUST return OS-appropriate hints
for Ollama:
- macOS: `brew install --cask ollama-app`
- Linux: `brew install ollama`

When `dnf` is the only detected package manager, install
hints for `unbound-force` SHOULD suggest `dnf install`
with the GitHub Release RPM URL.

#### Scenario: Ollama hint on Linux

- **GIVEN** a Linux system
- **WHEN** `uf doctor` generates an install hint for
  Ollama
- **THEN** the hint reads `brew install ollama` (not
  `--cask ollama-app`)

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
