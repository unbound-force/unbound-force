# Research: Knowledge Graph Integration

**Feature**: 010-knowledge-graph-integration
**Date**: 2026-03-08
**Purpose**: Resolve unknowns from Technical Context and
evaluate integration patterns.

## R-001: OpenCode MCP Server Configuration

**Decision**: Configure graphthulhu as a local MCP server in
`opencode.json` at the project root using stdio transport.

**Rationale**: OpenCode's configuration format is well-defined.
Local MCP servers are declared under the `"mcp"` key with
`"type": "local"` and a `"command"` array. The server is
launched as a subprocess by OpenCode at session start and
communicates via stdio. This matches the per-session lifecycle
model from the spec clarifications.

**Configuration format**:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "knowledge-graph": {
      "type": "local",
      "command": [
        "graphthulhu",
        "serve",
        "--backend", "obsidian",
        "--vault", ".",
        "--include-hidden",
        "--read-only"
      ],
      "enabled": true
    }
  }
}
```

**Key findings**:

- Config file goes in project root as `opencode.json` (or
  `opencode.jsonc`). Supports env var substitution via
  `{env:VAR_NAME}` syntax.
- Local servers use `"type": "local"` with `"command"` as an
  array of strings. The command is the executable plus args.
- MCP server tools are registered with the server name as a
  prefix (e.g., `knowledge-graph/search`). Tools can be
  enabled/disabled per-agent via the `"tools"` config key.
- No existing MCP configuration exists in this repo. An
  `opencode.json` file needs to be created.
- The existing `speckit.taskstoissues` command references a
  GitHub MCP server that is also not yet configured. The
  `opencode.json` file should include both servers when
  created.
- The `--vault "."` argument uses a relative path, which
  graphthulhu resolves relative to the working directory
  (the project root when launched by OpenCode).

**Alternatives considered**:

- Remote MCP server (`"type": "remote"` with URL): Rejected
  because it requires graphthulhu to run as an HTTP server
  independently of OpenCode, adding operational complexity.
  The stdio/local pattern is simpler and matches the
  per-session lifecycle.
- Per-agent MCP tool filtering: Available but not needed
  initially. All agents in the project should have access to
  the knowledge graph. Can be restricted later if needed.

## R-002: Hidden Directory Support (Upstream PR)

**Decision**: Contribute a `--include-hidden` flag to
graphthulhu via an upstream pull request. If the PR is not
accepted in a reasonable timeframe, maintain a fork.

**Rationale**: graphthulhu skips all directories starting with
`.` in three code locations within `vault/vault.go`. The change
is small (2-3 files, ~50-80 lines including tests), follows
existing code patterns (the `Option` constructor pattern used
by `WithDailyFolder`), and is backward-compatible (default
behavior unchanged). No existing issues or PRs address this --
the field is clear.

**Code locations requiring change**:

1. `vault/vault.go` `Load()` method: `filepath.Walk` callback
   checks `strings.HasPrefix(info.Name(), ".")` and returns
   `filepath.SkipDir`. Add guard: `&& !c.includeHidden`.
2. `vault/vault.go` `addWatcherDirs()` method: Same pattern
   as `Load()`. Add same guard.
3. `vault/vault.go` `handleEvent()` method: Checks
   `strings.Contains(event.Name, "/.")`. Add guard:
   `&& !c.includeHidden`.

**Implementation approach**:

- Add `includeHidden bool` field to `vault.Client` struct.
- Add `WithIncludeHidden(bool) Option` constructor.
- Guard three skip locations with `!c.includeHidden`.
- Add `--include-hidden` CLI flag in `main.go` `runServe()`.
- Always skip `.git` regardless of flag (hardcoded exclusion)
  to prevent indexing git internals.
- Add test: temp vault with hidden dir containing `.md` files,
  verify skipped by default, included with flag.

**Complexity**: Small. Low risk -- additive change, default
behavior preserved. Follows existing codebase patterns.

**Fallback**: If upstream PR is not accepted, fork graphthulhu
under the `unbound-force` GitHub org. MIT license permits this
without restriction.

**Alternatives considered**:

- Symlinks (create visible symlinks to hidden dirs): Rejected
  because it pollutes the repo with workaround files and
  creates confusion about canonical paths.
- Copy-on-sync (script copies hidden dir content to visible
  locations): Rejected because it duplicates files and
  introduces sync consistency issues.
- `--include-dirs` flag (whitelist specific dirs): Rejected
  in favor of `--include-hidden` because a simple boolean is
  easier to use and covers the common case. Users who need
  granular control can use the existing directory structure
  (hidden dirs are either all included or all excluded).

## R-003: graphthulhu Installation and Distribution

**Decision**: Install graphthulhu via `go install` for
development. Document binary download from GitHub Releases
as the primary installation method for users.

**Rationale**: graphthulhu provides pre-built binaries via
GitHub Releases (goreleaser-based). The latest release is
v0.4.0. For developers with Go installed, `go install
github.com/skridlevsky/graphthulhu@latest` is the simplest
path. For users without Go, downloading the binary from
releases is straightforward.

**Installation commands**:

```bash
# Option A: go install (requires Go toolchain)
go install github.com/skridlevsky/graphthulhu@latest

# Option B: Download binary from GitHub Releases
# https://github.com/skridlevsky/graphthulhu/releases
# Download the appropriate binary for your platform,
# extract, and add to PATH.
```

**Current status**: graphthulhu is not currently installed on
this development machine. It must be installed before testing.

**Alternatives considered**:

- Homebrew formula: Not available. Could be added to the
  `unbound-force/homebrew-tap` repo as a future convenience,
  but not required for initial integration.
- Docker: graphthulhu does not provide a Docker image. Could
  be created but adds complexity for a single-binary Go tool.
  Rejected for initial integration.
- npm package: Not applicable (Go binary, not Node.js).

## R-004: Vault Path Configuration Strategy

**Decision**: Use relative path `"."` (current directory) as
the vault path in the OpenCode MCP configuration.

**Rationale**: When OpenCode launches an MCP server subprocess,
it sets the working directory to the project root. Using `"."`
as the vault path means graphthulhu indexes from the project
root, which is the correct behavior for any Unbound Force repo
(whether the meta repo, Gaze, or a future hero repo). This
makes the configuration portable across repos without
hardcoding absolute paths.

**Validation needed**: Verify that OpenCode sets `cwd` to the
project root when spawning MCP subprocesses. If not, an
absolute path or `{env:PWD}` substitution may be needed.

**Alternatives considered**:

- Absolute path: Rejected because it ties the config to a
  specific machine/user directory structure.
- Environment variable (`{env:GRAPHTHULHU_VAULT_PATH}`):
  Available as fallback but adds configuration overhead.
  Relative path is simpler.
