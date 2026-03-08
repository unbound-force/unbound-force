# Contract: Unbound CLI

**Spec**: 003-specification-framework
**Date**: 2026-03-08

## Overview

The `unbound` CLI is a Go binary that scaffolds the
specification framework into target repositories. It follows
the same architecture as Gaze's `gaze init` command: files
are embedded in the binary via `go:embed` and extracted at
runtime.

## Installation

```bash
# Homebrew (recommended)
brew install unbound-force/tap/unbound

# Go install
go install github.com/unbound-force/unbound-force/\
cmd/unbound@latest

# Build from source
git clone https://github.com/unbound-force/unbound-force
cd unbound-force
go build -o unbound ./cmd/unbound
```

## Commands

### `unbound init`

Scaffolds the specification framework into the current
directory.

```bash
unbound init [--force]
```

**Flags**:
- `--force`: Overwrite all existing files (both user-owned
  and tool-owned). Without this flag, user-owned files that
  already exist are skipped.

**Behavior**:
1. Resolve current working directory as target
2. Check for `go.mod` or `.git` -- warn if absent (still
   proceed)
3. Walk embedded `assets/` filesystem
4. For each embedded file, compute output path under target
   directory
5. Insert version marker after YAML frontmatter:
   `<!-- scaffolded by unbound v1.0.0 -->`
6. Apply file ownership rules (see below)
7. Write file or skip based on ownership and existence
8. Print summary

**Output directories created**:

```text
<target>/
+-- .specify/
|   +-- templates/            # 6 Speckit templates
|   +-- scripts/bash/         # 5 Speckit scripts
+-- .opencode/
|   +-- command/              # 10 OpenCode commands
|   +-- agents/               # 1 agent file
+-- openspec/
    +-- schemas/
    |   +-- unbound-force/    # Custom OpenSpec schema
    +-- config.yaml           # Default OpenSpec config
    +-- specs/                # Empty (user populates)
    +-- changes/              # Empty (user populates)
```

**Exit codes**:
- 0: Success

**Result tracking** (matches Gaze `scaffold.Result`):
- `Created []string`: Files written for the first time
- `Skipped []string`: Files that existed, not modified
- `Overwritten []string`: Files replaced via `--force`
- `Updated []string`: Tool-owned files replaced via
  overwrite-on-diff

### `unbound version`

Prints the installed unbound version.

```bash
unbound version
```

Output: `unbound v1.0.0 (commit abc1234, built 2026-03-08)`

## File Ownership Model

Following Gaze's pattern, each file has an ownership
classification that determines behavior on re-run:

### User-Owned Files

Files the user is expected to customize. Skipped if they
already exist (unless `--force`).

| File | Rationale |
|------|-----------|
| `.specify/templates/*.md` | Users may customize templates |
| `.specify/scripts/bash/*.sh` | Users may add project-specific logic |
| `.specify/config.yaml` | Users customize project-specific settings |
| `.opencode/agents/*.md` | Users customize agent behavior |
| `openspec/config.yaml` | Users customize project context |

### Tool-Owned Files

Files maintained by the unbound tool. Overwritten if content
differs from the embedded version (even without `--force`).

| File | Rationale |
|------|-----------|
| `.opencode/command/speckit.*.md` | Pipeline commands must stay canonical |
| `.opencode/command/constitution-check.md` | Governance command must stay canonical |
| `openspec/schemas/unbound-force/*` | Schema must match canonical version |

### Ownership Decision Function

```go
func isToolOwned(relPath string) bool {
    if strings.HasPrefix(relPath, "openspec/schemas/") {
        return true
    }
    switch {
    case strings.HasPrefix(relPath,
        "opencode/command/speckit."):
        return true
    case relPath == "opencode/command/constitution-check.md":
        return true
    }
    return false
}
```

## Version Marker

Every scaffolded file gets a version marker:

```html
<!-- scaffolded by unbound v1.0.0 -->
```

Inserted after YAML frontmatter (if present) or appended
at the end. This provides provenance -- you can inspect any
scaffolded file and know which version created it.

## Version Injection

Via `ldflags` at build time (set by GoReleaser):

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

```yaml
# .goreleaser.yaml
ldflags:
  - -s -w
  - -X main.version={{.Tag}}
  - -X main.commit={{.Commit}}
  - -X main.date={{.CommitDate}}
```

## Release Pipeline

GoReleaser v2 configuration:

- **Builds**: darwin/amd64, darwin/arm64, linux/amd64,
  linux/arm64 with CGO_ENABLED=0
- **Archives**: `unbound_<version>_<os>_<arch>.tar.gz`
- **Checksums**: `checksums.txt` (SHA-256)
- **Homebrew cask**: Auto-published to
  `unbound-force/homebrew-tap` as `Casks/unbound.rb`
- **macOS signing**: Optional job for code signing and
  notarization (if `MACOS_SIGN_P12` secret is configured)

## Drift Detection Test

`TestEmbeddedAssetsMatchSource` ensures embedded copies
under `internal/scaffold/assets/` are byte-identical to
canonical source files. On mismatch:

```
drift detected: internal/scaffold/assets/X differs
from .opencode/X
Run: cp .opencode/X internal/scaffold/assets/X
```

This test runs in CI and prevents the embedded copies from
diverging from the canonical files used by the project's
own developers.
