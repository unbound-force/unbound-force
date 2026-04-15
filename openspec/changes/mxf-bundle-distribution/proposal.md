## Why

The `mxf` binary is built by GoReleaser in the same repo
and same release cycle as `unbound-force`, and is bundled
in the same tar.gz archive. However, `uf setup` tries to
install it via `brew install unbound-force/tap/mxf` -- a
formula/cask that does not exist in the homebrew-tap. This
means `mxf` installation silently fails on every platform.

Additionally, the `mxf` binary is absent from both the
RPM package (`nfpms:`) and the Homebrew Formula
(`brews:`), so even users who successfully install
`unbound-force` via these channels do not get `mxf`.

The fix is straightforward: since `mxf` ships in the same
archive as `unbound-force`, include it in the existing
RPM and Formula packages and simplify `uf setup` to just
verify `mxf` is present rather than attempting a separate
Homebrew install of a non-existent package.

## What Changes

1. Add `mxf` binary to the existing `nfpms:` RPM package
   contents in `.goreleaser.yaml`.

2. Add `bin.install "mxf"` to the existing `brews:`
   Formula install block in `.goreleaser.yaml`.

3. Replace `installMxF()` in `uf setup` -- instead of
   `brew install unbound-force/tap/mxf`, verify that
   `mxf` is already in PATH (it ships with
   `unbound-force`). If missing, hint that it comes
   bundled with `unbound-force`.

4. Update the `homebrewInstallCmd` and
   `genericInstallCmd` hints for `mxf` in `uf doctor`
   to reflect that it ships with `unbound-force`.

## Capabilities

### New Capabilities
- None. This is a bug fix for an existing capability.

### Modified Capabilities
- `mxf-distribution`: The `mxf` binary is included in
  the RPM package and Homebrew Formula alongside
  `unbound-force`, instead of requiring a non-existent
  separate Homebrew package.
- `installMxF`: Simplified from a Homebrew install
  attempt to a PATH verification with an install hint
  pointing to the `unbound-force` package.

### Removed Capabilities
- `brew install unbound-force/tap/mxf`: Removed because
  no such package exists. This was a broken install path.

## Impact

- `.goreleaser.yaml`: Add `mxf` to `nfpms:` contents
  and `brews:` install block.
- `internal/setup/setup.go`: Simplify `installMxF()`.
- `internal/doctor/environ.go`: Update install hints.
- `internal/setup/setup_test.go`: Update mxf tests.
- `internal/doctor/doctor_test.go`: Update hint tests
  if affected.

Small, focused change. No new dependencies, no new
files, no architecture changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change does not affect inter-hero artifact formats
or communication protocols. It modifies only the
distribution packaging of a companion binary.

### II. Composability First

**Assessment**: PASS

Bundling `mxf` with `unbound-force` is appropriate
because they share the same repo, release cycle, and
archive. `mxf` remains independently executable -- it
has no runtime dependency on `unbound-force`. The
bundling is a distribution convenience, not a coupling.
Users who only want `mxf` can still extract it from the
tar.gz archive.

### III. Observable Quality

**Assessment**: N/A

No change to output formats, provenance, or
machine-parseable data.

### IV. Testability

**Assessment**: PASS

The simplified `installMxF()` follows the existing
injectable dependency pattern (`LookPath`). Tests
verify PATH detection without network access.
