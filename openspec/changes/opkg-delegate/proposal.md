## Why

`uf init` embeds ~33 Markdown assets (agents, commands,
packs) inside the Go binary via `go:embed`. This creates
two problems:

1. **Release coupling** — updating a single agent prompt
   requires a full binary release (rebuild Go binary, cut
   GoReleaser release, push to Homebrew tap).
2. **No per-repo governance** — every repo that runs
   `uf init` gets whatever asset versions shipped with
   the installed `uf` binary. No repo-level control over
   content upgrades.

OpenPackage (`opkg`) is a package manager for AI coding
tool assets (agents, commands, rules). Delegating asset
installation to `opkg` decouples content updates from
binary releases and enables per-repo version pinning
via opkg lockfiles.

### Distribution Options

- **Option A** (this change): `uf init` detects `opkg`
  on PATH and delegates to `opkg install`; embedded
  assets remain as fallback when `opkg` is absent.
- **Option B** (future): Standalone `opkg install`
  without the `uf` binary — users install packages
  directly for AI harness-only setups.

### User-Facing Flow

The user journey is unchanged:

```
brew install unbound-force   # or dnf
uf setup                     # installs deps (including opkg)
uf init                      # scaffolds repo (delegates to opkg internally)
```

`opkg` runs inside `uf init` transparently. Users never
invoke it directly unless they choose Option B (future).

### Multi-Harness Support

opkg natively supports 40+ AI coding platforms. Files
and paths auto-convert to platform-specific locations
during install. `uf` has no harness-specific logic —
it calls `opkg install` and opkg handles routing.

Target harnesses for this change:

| Harness | Directory | Agents | Commands | Rules | MCP |
|---------|-----------|--------|----------|-------|-----|
| OpenCode | `.opencode/` | `agents/` | `commands/` | — | `opencode.json` |
| Cursor | `.cursor/` | `agents/` | `commands/` | `rules/` | `mcp.json` |
| Claude Code | `.claude/` | `agents/` | `commands/` | `rules/` | `.mcp.json` |

Platform detection is automatic — opkg detects existing
platform directories (`.opencode/`, `.cursor/`, `.claude/`)
at install time. For initial setup before any dirs exist,
`uf init` passes `--platforms opencode cursor claude-code`
to ensure all three are bootstrapped.

Per-file harness overrides use frontmatter:

```yaml
---
openpackage:
  opencode:
    mode: subagent
    temperature: 0.1
    tools:
      write: false
  cursor:
    mode: agent
  claude:
    tools: Read, Grep, Glob
---
```

### Version Governance

Each consuming repo controls its own upgrade cadence via
opkg lockfiles. `uf` does not manage versions — it calls
`opkg install` and opkg resolves from the repo's lockfile.
Breaking prompt changes become opt-in per repo, not
forced by binary version.

| Concern | Before | After |
|---------|--------|-------|
| Content release | Tied to binary | Independent via opkg publish |
| Per-repo pinning | Not possible | opkg lockfile |
| Harness coupling | OpenCode-only | OpenCode + Cursor + Claude Code |
| User-facing flow | `uf setup` → `uf init` | Same |
| opkg dependency | N/A | Optional (embedded fallback, OpenCode only) |

## What Changes

1. **Package source trees** — `.openpackage/packages/review-council/`
   and `.openpackage/packages/workflows/` contain OpenPackage manifests
   (`openpackage.yml`) and regenerated copies of agents,
   commands, rules, and MCP config. These are the source
   of truth for `opkg install`.

2. **`uf init` delegation** — `openPackageInstall()` in
   `internal/scaffold/scaffold.go` runs
   `opkg install @unbound-force/review-council` (and
   `@unbound-force/workflows` unless `--divisor`). On
   success, `assetDelegatedToOpenPackage()` skips writing
   the matching embedded assets. On failure or missing
   `opkg`, all embedded assets deploy as before.

3. **`ExecCmdInDir` injectable** — new `Options` field
   runs commands with a working directory set (needed for
   `opkg install` in the target dir). Follows the existing
   `ExecCmd`/`LookPath` injection pattern.

4. **`SkipOpenPackage` test flag** — `Options` field that
   disables `opkg` delegation so scaffold tests remain
   deterministic against embedded assets.

5. **`uf setup` step 15** — `installOpkg()` attempts
   `brew install openpackage`. Graceful skip when Homebrew
   is unavailable or the formula doesn't exist yet.

## Capabilities

### New Capabilities
- `opkg-delegation`: `uf init` delegates agent/command
  installation to OpenPackage when `opkg` is on PATH.
  Consuming repos govern content versions via opkg
  lockfiles — `uf` passes through to `opkg install`
  without managing versions itself.
- `multi-harness`: packages install to OpenCode, Cursor,
  and Claude Code simultaneously. opkg auto-detects
  present platforms; `uf init` bootstraps all three on
  initial setup via `--platforms opencode cursor claude-code`.
- `packages/`: Two OpenPackage source trees
  (`review-council`, `workflows`) with manifests and
  content ready for `opkg publish`.
- `installOpkg`: `uf setup` installs the `opkg` CLI
  via Homebrew as step 15.

### Open Questions

None. Package location resolved: `.openpackage/packages/`
is opkg's default lookup location. `opkg install
.openpackage/packages/review-council` works without
registry publishing or subpath syntax.

### Modified Capabilities
- `scaffold.Run()`: Checks for `opkg` before walking
  embedded assets; skips delegated paths on success.
- `scaffold.Options`: Added `ExecCmdInDir`,
  `SkipOpenPackage` fields.
- `uf setup`: Step count 14 → 15.

### Removed Capabilities
- None.

## Impact

- `internal/scaffold/scaffold.go`: `openPackageInstall()`
  adds `--platforms opencode cursor claude-code` for
  initial setup. `assetDelegatedToOpenPackage()`,
  `ExecCmdInDir` field, `SkipOpenPackage` field,
  skip logic in `Run()` — unchanged.
- `internal/scaffold/scaffold_test.go`: Updated tests
  with `SkipOpenPackage: true` to preserve deterministic
  embedded-asset assertions.
- `internal/setup/setup.go`: `installOpkg()`, step count
  14 → 15.
- `.openpackage/packages/review-council/`: 16 files (9 agents,
  2 commands, 3 rules, 1 MCP config, 1 manifest).
  All 9 agent files have `openpackage.cursor:` and
  `openpackage.claude:` frontmatter alongside
  `openpackage.opencode:`. Claude Code uses tool list
  syntax (`tools: Read, Grep, Glob`); Cursor uses
  `mode: agent`.
- `.openpackage/packages/workflows/`: 15 files (1 agent,
  13 commands, 1 manifest). Agent frontmatter same as above.

No new Go dependencies. All stdlib.

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: PASS

Package manifests and content are self-describing
artifacts (YAML metadata, Markdown content). `opkg`
operates independently — no runtime coupling between
the `uf` binary and `opkg`. Fallback to embedded assets
preserves autonomy when `opkg` is absent.

### II. Composability First

**Assessment**: PASS

Core design principle: `opkg` is optional. `uf init`
works identically without it (embedded fallback). Users
can also run `opkg install` standalone without `uf`
(Option B, future). The two paths are additive, not
mutually exclusive. Per-repo lockfiles give consuming
repos independent upgrade control without coupling to
the `uf` release cadence.

### III. Observable Quality

**Assessment**: PASS

`printSummary()` reports `opkg` delegation results
(installed/failed/skipped) alongside other sub-tool
outcomes. Users see which distribution path was used.

### IV. Testability

**Assessment**: PASS

`SkipOpenPackage` flag isolates embedded-asset tests
from `opkg` availability. `ExecCmdInDir` is injectable,
following the established `LookPath`/`ExecCmd` pattern.
No external services required for testing.
