## Context

The `mxf` binary is built alongside `unbound-force` in
the same GoReleaser config. Both binaries are included in
the same tar.gz archive. However, `mxf` is not included
in the RPM (`nfpms:`) or Homebrew Formula (`brews:`), and
`uf setup` tries to install it from a non-existent
Homebrew package (`unbound-force/tap/mxf`).

## Goals / Non-Goals

### Goals
- `mxf` is included in the RPM and Formula alongside
  `unbound-force`
- `uf setup` correctly detects and reports `mxf` status
- Install hints point users to the `unbound-force`
  package (not a non-existent `mxf` package)

### Non-Goals
- Creating a separate Homebrew package for `mxf`
- Splitting `mxf` into its own repository
- Changing `mxf` functionality or commands

## Decisions

### D1: Bundle mxf in existing packages

Add `mxf` to the existing `unbound-force` RPM and
Formula rather than creating separate packages.

**Rationale**: `mxf` is in the same repo, same release
cycle, and already in the same tar.gz archive. A
separate package would require a separate tap entry and
a separate signing pipeline -- unnecessary overhead for
a companion binary.

**Constitution**: Composability First is maintained --
`mxf` remains independently executable with no runtime
dependency on `unbound-force`. Bundling is a distribution
convenience.

### D2: Simplify installMxF() to PATH verification

Replace the `brew install unbound-force/tap/mxf` call
with a simple `LookPath("mxf")` check. If `mxf` is not
found, report it as a note pointing to the
`unbound-force` package rather than attempting a broken
install.

**Rationale**: Since `mxf` ships inside the
`unbound-force` package, installing `unbound-force`
automatically provides `mxf`. There is no separate
install action needed. The step becomes a verification
rather than an installation.

**Alternative considered**: Remove the `mxf` step
entirely from `uf setup`. Rejected because the
verification step provides useful diagnostic output
when `mxf` is unexpectedly missing (e.g., if the user
installed from a raw tar.gz and only extracted one
binary).

### D3: Update install hints to reference parent package

Change `homebrewInstallCmd("mxf")` and
`genericInstallCmd("mxf")` in `environ.go` to reference
the `unbound-force` package:
- Homebrew: `brew install unbound-force/tap/unbound-force`
- Generic: `Install unbound-force (mxf is bundled)`

**Rationale**: The current hint
`brew install unbound-force/tap/mxf` references a
package that does not exist, producing a confusing
error.

## Risks / Trade-offs

### R1: Users cannot install mxf without unbound-force

Bundling means users cannot install `mxf` alone via
Homebrew or RPM. This is acceptable because `mxf` is
designed as a companion tool within the Unbound Force
ecosystem and lives in the same repository.

### R2: RPM package size increases slightly

Adding the `mxf` binary to the RPM increases its size.
The binary is ~10MB compressed -- negligible for a
package manager install.
