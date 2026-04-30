## Context

`uf init` embeds ~30 Markdown assets (agents, commands,
packs) inside the Go binary via `go:embed`. Updating a
single agent prompt requires a full binary release. The
OpenPackage CLI (`opkg`) is an emerging package manager
for AI coding tool assets. Delegating asset installation
to `opkg` decouples content updates from binary releases,
letting users receive agent/command/pack updates without
waiting for a new `uf` version.

The scaffold engine already supports sub-tool delegation
(Dewey, Replicator, Specify, OpenSpec, Gaze) via
`initSubTools()`. This change adds `opkg` as a post-walk
delegation that runs **after** the embedded asset walk,
overlaying harness-specific content on top of the baseline
embedded files.

## Goals / Non-Goals

### Goals
- Decouple agent/command/pack content from binary release
  cycle by delegating to `opkg install` when available
- Maintain full backward compatibility: `uf init` works
  identically when `opkg` is absent (embedded fallback)
- Create two OpenPackage source trees (`review-council`,
  `workflows`) with manifests and content ready for
  `opkg publish`
- Install `opkg` CLI via `uf setup` (step 15, graceful
  skip when formula unavailable)

### Non-Goals
- Removing embedded assets from the binary (they remain
  as the fallback path)
- Implementing `opkg` itself (it is an external tool)
- Migrating non-command/non-agent assets (e.g., OpenSpec
  schemas, skills) to OpenPackage
- Option B from the proposal (standalone `opkg install`
  without `uf`) -- deferred for future evaluation

## Decisions

### D1. Post-walk delegation: embed baseline, opkg overlay

`openPackageInstall()` runs **after** `fs.WalkDir` completes.
All embedded assets are always written (baseline); `opkg`
then routes harness-specific content to the correct
directories (`.opencode/`, `.cursor/`, `.claude/`). This
model simplifies failure handling (embedded baseline is
always present, opkg adds on top) and avoids write-then-
check churn. The embedded `.openpackage/packages/` trees
deployed during the walk provide local source paths for
`opkg install`, avoiding remote registry lookups.

Alternative considered: pre-walk execution with
`assetDelegatedToOpenPackage()` skip guard. Rejected
because it creates partial-state risk (if opkg fails
mid-walk, nothing is written) and complicates tests.
The post-walk overlay model is safer and simpler.

Alternative considered: post-walk cleanup (delete files
written by the walk that opkg also installed). Rejected
because it creates unnecessary I/O and complicates
rollback on partial failure.

### D2. Two packages, not one

The content is split into `@unbound-force/review-council`
(Divisor agents, review commands, convention packs) and
`@unbound-force/workflows` (Speckit commands, OpenSpec
commands, constitution-check). This mirrors the logical
split in the scaffold engine (`isDivisorAsset()` vs.
general assets) and allows users to install the review
council alone (`uf init --divisor` installs only
`review-council`).

`workflows` declares a `^0.1.0` dependency on
`review-council` because some workflow commands reference
review-council agents (e.g., `/unleash` runs
`/review-council`).

### D3. ExecCmdInDir injectable

`opkg install` needs a working directory set to the
target directory (packages install relative to cwd). A
new `ExecCmdInDir` field on `Options` follows the
existing `ExecCmd`/`LookPath` injection pattern. Default
implementation wraps `exec.Command` with `cmd.Dir` set.
Tests inject a fake that records calls without spawning
processes.

### D4. SkipOpenPackage test isolation

All existing scaffold tests set `SkipOpenPackage: true`
to keep embedded-asset assertions deterministic. Tests
that exercise the opkg delegation path use separate test
functions with controlled `LookPath`/`ExecCmdInDir`
fakes. This prevents test flakiness from opkg
availability on the CI machine.

### D5. installOpkg graceful degradation

`uf setup` step 15 attempts `brew install openpackage`.
On failure (formula not yet published, Homebrew absent),
it returns `"skipped"` with a manual-install hint rather
than a hard error. This follows the Composability First
principle: the toolchain works without opkg.

### D6. OpenPackage directory conventions

Each package follows opkg's directory conventions:
```
packages/<name>/
  openpackage.yml       # Manifest (name, version, deps)
  README.md             # Package documentation
  agents/<name>/*.md    # Agent files
  commands/<name>/*.md  # Command files
  rules/<name>/*.md     # Convention packs
  mcp.jsonc             # Optional MCP config template
```

Agent/command content in `packages/` is regenerated from
the embedded scaffold assets to stay in sync. The
scaffold assets remain the source of truth; the package
trees are distribution-format copies.

## Risks / Trade-offs

### Content drift between embedded assets and packages

The package trees (`packages/`) are copies of the
embedded assets (`internal/scaffold/assets/`). If one is
updated without the other, users see different content
depending on whether `opkg` is available. Mitigation:
a future drift-detection test (similar to
`TestEmbeddedAssets_MatchSource`) should verify the
package trees match the embedded assets.

### opkg not yet widely available

The `brew install openpackage` formula may not exist at
the time of release. All delegation paths gracefully
degrade: `openPackageInstall()` returns false when opkg
is missing, and `installOpkg()` returns "skipped" when
the formula is unavailable. No user action is blocked.

### DivisorOnly mode interaction

When `DivisorOnly=true`, only `@unbound-force/review-council`
is installed (not `workflows`). The
`assetDelegatedToOpenPackage()` function covers all
Divisor assets, so the skip logic works correctly in both
modes.
