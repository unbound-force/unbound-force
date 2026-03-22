# Research: Binary Rename

**Date**: 2026-03-22
**Branch**: `013-binary-rename`

## R1: Symlink vs Hardlink vs Wrapper Script for Alias

**Decision**: Symlink (`uf` → `unbound-force`).

**Rationale**: Symlinks are the standard Homebrew
mechanism for binary aliases (`bin.install_symlink`).
They are transparent to the user, produce identical
behavior when invoked, and require no maintenance. The
Cobra root command detects the invoked name via
`os.Args[0]` but does not need to -- `uf` and
`unbound-force` produce identical output regardless.

**Alternatives considered**:
- Hardlink: Rejected because Homebrew does not support
  hardlinks in formulas, and some filesystems (e.g.,
  APFS across volumes) restrict them.
- Wrapper script: Rejected because it adds a process
  fork, complicates debugging, and is unnecessary
  complexity for a simple alias.

## R2: Cobra Root Command Configuration

**Decision**: Set `Use: "unbound-force"` with the long
description mentioning `(alias: uf)`.

**Rationale**: The `Use` field determines how the binary
appears in help output and error messages. Using the
primary name ensures that help text, man pages, and
completion scripts all reference the canonical name.
The alias is mentioned in the description so users
discover it naturally.

**Alternatives considered**:
- `Use: "uf"`: Rejected because the primary name should
  be the canonical one. Users who installed via Homebrew
  see `unbound-force` in their formula list.
- Detect `os.Args[0]` and set `Use` dynamically:
  Rejected because it adds complexity for no functional
  benefit. The output is the same regardless.

## R3: GoReleaser Binary Name Configuration

**Decision**: Set `binary: unbound-force` in the builds
section. The `uf` symlink is handled by the Homebrew
cask configuration (the project uses `homebrew_casks`,
not `brews`/formulas).

**Rationale**: GoReleaser's `binary` field directly
controls the output filename. The Homebrew cask's
install configuration handles the `uf` alias via a
`binary` stanza with `target:` parameter or a
post-install hook creating the symlink.

**Alternatives considered**:
- Build two separate binaries: Rejected because it
  doubles the build time and artifact size for zero
  functional benefit.
- Post-build rename script: Rejected because GoReleaser
  handles this natively.

## R4: Makefile Install Target Design

**Decision**: Add an `install` target that builds to
`$GOPATH/bin/unbound-force` and creates a symlink
`$GOPATH/bin/uf` → `$GOPATH/bin/unbound-force`.

**Rationale**: Developers using `go install` get the
binary named correctly by the directory convention
(`cmd/unbound-force/`). The Makefile `install` target
adds the `uf` symlink which `go install` cannot create.
This is a convenience for local development.

**Alternatives considered**:
- Only rely on `go install`: Rejected because `go
  install` cannot create symlinks. Developers would
  need to manually create the `uf` alias.

## R5: Scope of String Replacements

**Decision**: Replace `unbound init`, `unbound doctor`,
`unbound setup`, and `unbound version` references in
living documents only. Use `uf` as the replacement in
user-facing hints and instructions (shorter, more
convenient). Use `unbound-force` in formal contexts
(Homebrew formulas, `go install` instructions,
AGENTS.md build commands).

**Rationale**: The short alias `uf` is what developers
will type daily. Using it in hints and instructions
ensures the instructions match the developer's muscle
memory. The full name `unbound-force` appears where
formality matters (package managers, build systems,
documentation headers).

**Alternatives considered**:
- Always use `unbound-force` in all references: Rejected
  because it is 13 characters vs 2. Daily-use commands
  should be short.
- Always use `uf` everywhere: Rejected because formal
  contexts (Homebrew, `go install`, AGENTS.md build
  commands) should use the canonical name for clarity.

## R6: Cross-Repo Update Strategy

**Decision**: Meta repo rename is implemented first.
Cross-repo updates (gaze, website, homebrew-tap)
proceed as independent follow-up tasks that can run
in parallel after the meta repo merge.

**Rationale**: The meta repo contains the binary source,
scaffold engine, and embedded assets. It must be renamed
first. Cross-repo changes are string replacements in
documentation and agent files -- they have no code
dependencies on the meta repo rename being complete,
only documentation consistency dependencies.

**Alternatives considered**:
- Atomic cross-repo rename (all repos at once): Rejected
  because it requires coordinating merges across 4
  repos simultaneously, which is fragile and
  unnecessary.
