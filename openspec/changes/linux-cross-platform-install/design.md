## Context

GoReleaser cross-compiles `unbound-force` for darwin,
linux, and windows (amd64 + arm64). Linux binaries are
uploaded to GitHub Releases as tar.gz but have no native
package manager distribution. The only Homebrew channel is
a Cask, which Homebrew rejects on Linux.

`uf setup` also has a hard-coded macOS-only Ollama
installation path (`brew install --cask ollama-app`).

Both issues block Linux adoption.

## Goals / Non-Goals

### Goals
- Linux users can install `unbound-force` via
  `brew install unbound-force/tap/unbound-force` (Formula)
- Fedora/RHEL users can install via
  `dnf install <github-release-url>.rpm`
- `uf setup` detects `dnf` and uses RPM URLs when Homebrew
  is absent
- `uf setup` installs Ollama correctly on Linux
- `uf doctor` reports `dnf` as a detected package manager
- All new code is testable via injectable dependencies

### Non-Goals
- DEB packages for Debian/Ubuntu (future increment)
- COPR repository for `dnf install unbound-force` without
  URL (future increment)
- RPM distribution for sibling repos (gaze, dewey,
  replicator) -- those repos adopt independently later
- Windows package manager support (winget, choco)
- Replacing the macOS Cask -- it remains for signed binary
  distribution

## Decisions

### D1: Homebrew Formula alongside Cask (not replacing)

Add a `brews:` section to `.goreleaser.yaml` that produces
a Formula in `Formula/unbound-force.rb`. The existing
`homebrew_casks:` section remains unchanged.

**Rationale**: The Cask serves a specific purpose on macOS
-- signed binary distribution with quarantine removal.
Removing it would regress the macOS experience. A Formula
coexists cleanly: `brew install name` resolves Formula
first, `brew install --cask name` uses the Cask.

**Constitution**: Composability First -- additive channel,
no removal of existing functionality.

### D2: Formula uses `skip_upload: true` with CI patching

The Formula uses `skip_upload: true` in GoReleaser,
matching the existing Cask pattern. The release job uploads
the generated Formula as a release asset. The sign-macos
job downloads it, patches darwin checksums with signed
binary values, and pushes the final Formula to the
homebrew-tap alongside the Cask.

**Rationale**: After macOS code signing, the darwin tar.gz
files are replaced with signed versions (different
checksums). The Formula must reference the signed checksums
for macOS to avoid install failures. Patching in the
sign-macos job is the same proven pattern used for the
Cask.

**Alternative considered**: Linux-only Formula (no darwin
entries). Rejected because it would split the UX --
`brew install` would work on Linux but not macOS, forcing
macOS users to know about `--cask`.

### D3: RPM via GoReleaser `nfpms:` (built-in)

Use GoReleaser's built-in `nfpms:` to generate RPM
packages. No external tooling or additional CI steps
required. RPMs are uploaded to GitHub Releases as standard
release assets.

**Rationale**: GoReleaser already has the binary artifacts
and metadata. `nfpms:` is a zero-dependency addition that
produces well-formed RPMs with proper metadata, license,
and symlink support.

**Alternative considered**: Separate `rpmbuild` step or
Fedora COPR. Both require infrastructure beyond the
existing GoReleaser pipeline. COPR can be added later as
an incremental improvement without changing the RPM
generation.

### D4: `uf setup` RPM install via URL (not repo)

When `dnf` is detected and Homebrew is absent, `uf setup`
installs `unbound-force` RPM from a GitHub Release URL:

```
dnf install https://github.com/unbound-force/
  unbound-force/releases/latest/download/
  unbound-force_<version>_linux_<arch>.rpm
```

`dnf` supports direct URL installation natively. No
repository configuration needed.

**Rationale**: Simplest path with zero infrastructure
requirements. Users on Fedora/RHEL already trust
`dnf install <url>` for one-off tools.

**Version detection**: The RPM URL includes a version
string. `uf setup` will resolve the latest version via
`gh api` or a hardcoded GitHub Releases URL pattern.
For v1, using `/latest/download/` redirect is sufficient.

**Note**: `/latest/download/` only works if the RPM
filename does not contain the version number, OR if we
use the GitHub API to resolve the latest tag first. Since
GoReleaser includes the version in the filename by
default, we will use `gh api` to resolve the latest
release tag, then construct the URL.

### D5: `ManagerDnf` as a new package manager kind

Add `ManagerDnf ManagerKind = "dnf"` to `models.go`.
Detected via `LookPath("dnf")` in `DetectEnvironment()`,
following the exact pattern used for Homebrew at
`environ.go:94`.

**Rationale**: The doctor/setup pattern is
manager-detection-first. Each `installXxx()` function
checks available managers and selects the appropriate
install method. Adding `dnf` to the detection allows all
install functions to query it uniformly.

### D6: OS-aware Ollama installation

`installOllama()` currently calls
`brew install --cask ollama-app` unconditionally. Change
to:
- macOS: `brew install --cask ollama-app` (unchanged)
- Linux: `brew install ollama` (Ollama publishes a
  Homebrew formula that works on Linuxbrew)
- Linux without Homebrew: skip with install hint pointing
  to `https://ollama.com/download`

**Rationale**: Ollama maintains both a Cask (`ollama-app`)
and a Formula (`ollama`) in their official Homebrew tap.
The Formula works on Linux. No `dnf install` path for
Ollama -- it is not in Fedora repos and does not produce
RPMs. The official install method on Linux is their
install script, but per user requirement we avoid
curl|bash scripts.

### D7: Formula directory in homebrew-tap

The Formula will be placed in `Formula/unbound-force.rb`
in the `unbound-force/homebrew-tap` repository. This
directory does not currently exist -- the CI job will
create it when pushing the first Formula.

**Rationale**: Standard Homebrew tap convention: `Casks/`
for casks, `Formula/` for formulas. Both can coexist.

## Risks / Trade-offs

### R1: sign-macos job complexity increases

The sign-macos CI job must now patch checksums in two
files (Cask + Formula) instead of one. Mitigation: the
patching logic is identical -- extract darwin sha256 from
signed archives, `awk`-substitute into the Ruby file.
Duplication is ~10 lines.

### R2: RPM filename includes version

GoReleaser's `nfpms:` generates filenames like
`unbound-force_0.12.0_linux_amd64.rpm`. This means
`dnf install` URLs include the version, and there is no
stable `/latest/download/unbound-force.rpm` redirect.
`uf setup` must resolve the current version somehow.

Mitigation options (in order of preference):
1. Use `uf` binary's own version to construct the URL
   (the user already has `uf` if running `uf setup`)
2. Use the GitHub API (`gh release view --json`)
3. Use a version-free `nfpms.file_name_template`

Option 1 is simplest for the `uf setup` case. For
first-time install (no `uf` yet), the user follows
documentation with explicit version in the URL.

### R3: Only `unbound-force` gets RPMs initially

Sibling tools (gaze, dewey, replicator, mxf) remain
Cask-only. `uf setup` on Fedora without Homebrew will
skip these tools with "download from GitHub releases"
hints. This is acceptable for initial Linux support --
`unbound-force` is the entry point, and `uf setup` can
install the tools that have alternative install methods
(Node.js tools via npm, Go tools via `go install`).

### R4: No `dnf` path for Ollama or Dewey

Ollama and Dewey do not publish RPMs. On Fedora without
Homebrew, these will be skipped with install hints.
This is acceptable because both are optional
(Constitution Principle II -- Composability First).
