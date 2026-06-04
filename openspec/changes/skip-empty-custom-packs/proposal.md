## Why

Three convention pack stubs (`default-custom.md`,
`go-custom.md`, `content-custom.md`) are unconditionally
imported via `@` directives in `CLAUDE.md` by every `uf init`
run. Each file contains only a YAML frontmatter block, a
section heading, usage instructions, and a placeholder HTML
comment — no actual rules. Across all five repos in the
Unbound Force org, none of these files contain real content.

Loading empty files into every agent session wastes context
tokens with no operational benefit. The cost is proportional
to session frequency: high-throughput swarm workflows amplify
the waste. The issue is systemic — `buildCLAUDEmdBlock` in
the scaffold engine adds these imports unconditionally, so
`uf init --reinit` will re-add them even if a user manually
removes them.

## What Changes

The scaffold engine's `CLAUDE.md` block builder is changed to
detect whether a `*-custom.md` pack contains actual rule
content before emitting its `@` import line. A custom pack is
considered "empty" if it has no non-whitespace content below
the `<!-- Add project-specific rules below this line -->`
sentinel comment. Empty custom packs are silently omitted from
the generated `CLAUDE.md` block.

The scaffold continues to write the stub files on first `uf
init` (they remain available for users to fill in). The change
is purely in which files get imported into `CLAUDE.md`, not in
which files get written to disk.

The embedded scaffold asset templates (`internal/scaffold/
assets/opencode/uf/packs/*-custom.md`) are not modified —
they remain the canonical starter stubs.

## Capabilities

### New Capabilities
- `empty-pack-skip`: When generating the `CLAUDE.md` managed
  block, custom packs that contain no rules below the
  placeholder sentinel are omitted from `@` import lines.

### Modified Capabilities
- `collectDeployedPacks`: Extended with an optional filesystem
  check parameter; when a project root is provided, custom
  pack entries are filtered through `hasRuleContent()`.
- `buildCLAUDEmdBlock`: Passes the project root to
  `collectDeployedPacks` so the live files on disk are
  evaluated.
- `ensureCLAUDEmd`: Already calls `buildCLAUDEmdBlock` with
  the target directory — no additional change needed here.

### Removed Capabilities
- None.

## Impact

- `internal/scaffold/scaffold.go`: Two functions modified
  (`collectDeployedPacks`, `buildCLAUDEmdBlock`), one new
  helper added (`hasRuleContent`).
- `internal/scaffold/scaffold_test.go`: New test cases for the
  empty-pack-skip behaviour.
- `CLAUDE.md` in this repo: Three `@` import lines removed on
  the next `uf init --reinit` (or by direct edit as part of
  this change).
- No CLI interface changes. No schema changes. No changes to
  the embedded asset templates.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change does not alter any artifact format or inter-hero
communication protocol. `CLAUDE.md` is a convention pack
loader, not an artifact boundary. Heroes continue to
communicate through the same well-defined artifacts. The only
effect is that agents receive a smaller (but equally correct)
context window.

### II. Composability First

**Assessment**: PASS

The scaffold engine remains fully standalone. The new
`hasRuleContent` helper reads only local filesystem paths
already resolved within `Run()`'s working directory. No new
external dependencies are introduced. Repos that have
populated custom packs continue to receive their `@` imports
unchanged.

### III. Observable Quality

**Assessment**: PASS

The change produces no machine-parseable output of its own,
but it improves the quality of every agent session by
eliminating noise from the context. The `uf init` run log
could optionally emit a notice when a pack is skipped, but
this is not required for correctness.

### IV. Testability

**Assessment**: PASS

The new `hasRuleContent` helper is a pure function of a file
path; it can be tested in isolation using `t.TempDir()` with
crafted stub files. The existing scaffold test suite already
uses this pattern. No external services are required.
