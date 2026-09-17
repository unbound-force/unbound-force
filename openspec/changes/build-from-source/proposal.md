## Why

The README and QUICKSTART.md install sections list Homebrew and dnf
but omit "Build from Source" instructions. Other UF components
([gaze](https://github.com/unbound-force/gaze#build-from-source),
[dewey](https://github.com/unbound-force/dewey#build-from-source),
[replicator](https://github.com/unbound-force/replicator/pull/104))
document this path, creating an inconsistency across the
organization. Contributors, users on unsupported platforms, and
users who prefer building from source have no documented path
despite `make install` already handling this (including the `uf`
symlink).

Tracked by [#536](https://github.com/unbound-force/unbound-force/issues/536).

## What Changes

Add a "Build from Source" subsection to both `README.md` (in the
`## Getting Started` section after Fedora/RHEL) and `QUICKSTART.md`
(in the `## Install` section after the dnf subsection). The
subsection documents prerequisites, build command, and verification.

## Capabilities

### New Capabilities
- `Build from Source docs`: README and QUICKSTART.md subsections
  with prerequisites (Go 1.25+, make), clone + install command
  (`make install`), and verification (`uf version`). Documents
  that `make install` creates both the `unbound-force` binary and
  the `uf` symlink in `$GOPATH/bin`.

### Modified Capabilities
- None

### Removed Capabilities
- None

## Impact

- **Files**: `README.md` (~10 lines), `QUICKSTART.md` (~12 lines)
- **Behavior**: No code changes. Documentation only.
- **Risk**: None. Additive content with no effect on build, tests,
  or CI.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

Documentation-only change. No artifact interfaces, MCP tools, or
inter-agent communication affected.

### II. Composability First

**Assessment**: PASS

Documenting how to build from source directly supports standalone
installability -- users can build unbound-force independently
without relying on Homebrew, dnf, or pre-built binaries.

### III. Observable Quality

**Assessment**: N/A

No output formats, provenance metadata, or machine-parseable
interfaces affected.

### IV. Testability

**Assessment**: N/A

No code changes. No testable components introduced or modified.
<!-- scaffolded by uf vdev -->
