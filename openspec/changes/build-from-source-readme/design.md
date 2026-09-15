## Context

The unbound-force README `## Getting Started` section lists Homebrew
and dnf installation methods. QUICKSTART.md has a more detailed
`## Install` section with the same two methods. Neither includes
"Build from Source" instructions. The project has a `make install`
target that builds the binary to `$GOPATH/bin/unbound-force` and
creates a `uf` symlink.

## Goals / Non-Goals

### Goals
- Add a "Build from Source" subsection to both README.md and
  QUICKSTART.md
- Match the structure used by gaze, dewey, and replicator
- Document the correct build command for unbound-force specifically

### Non-Goals
- Modifying the Makefile or build system
- Documenting cross-compilation or GoReleaser usage
- Adding version injection via ldflags to the Makefile (separate
  concern)

## Decisions

### D1: Use `make install` instead of `make build` or bare `go build`

The unbound-force `make build` target runs `go build ./...` which
is a compile check that produces no named binary. `make install`
is the correct target: it runs
`go build -o $GOPATH/bin/unbound-force ./cmd/unbound-force/` and
then creates a `uf` symlink via `ln -sf`. This gives users a
working binary on their `$PATH` with the canonical `uf` alias.

This differs from replicator (which uses `make build` with ldflags
to `bin/replicator`) and from gaze/dewey (which use bare
`go build` to the project root).

### D2: Place after Fedora/RHEL in README

Following the same ordering as other UF repos: package managers
first (easiest), then build from source (most effort). In README,
the subsection goes after the Fedora/RHEL block and before the
QUICKSTART.md reference paragraph.

### D3: Update QUICKSTART.md as well

Issue #536 explicitly requires both files. In QUICKSTART.md, the
subsection goes after the "Fedora / RHEL (dnf -- minimal)" section
and before "## For Project Maintainers".

### D4: Verification uses `uf version`

Since `make install` places the binary on `$PATH` (via
`$GOPATH/bin`), verification is simply `uf version` -- no relative
path needed. This differs from replicator which uses
`bin/replicator version` (local binary). The `uf` alias is the
primary user-facing name and is created automatically by
`make install`.

### D5: Mention `$GOPATH/bin` must be on `$PATH`

Users who have Go installed but haven't added `$GOPATH/bin` to
their `$PATH` will not find the `uf` command after `make install`.
A brief note about this avoids a common gotcha.

## Risks / Trade-offs

- **Low risk**: Documentation-only change. No code, tests, or CI
  affected.
- **Make dependency**: Documenting `make install` introduces a soft
  dependency on `make` being installed. Standard on macOS and Linux.
- **$GOPATH/bin on PATH**: Users must have `$GOPATH/bin` on their
  PATH. This is standard Go setup but worth mentioning.
- **Composability First alignment**: Documenting build-from-source
  strengthens this constitutional principle by ensuring users can
  build unbound-force without any external distribution channel.
<!-- scaffolded by uf vdev -->
