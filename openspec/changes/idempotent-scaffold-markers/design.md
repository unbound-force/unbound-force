## Context

`insertMarkerAfterFrontmatter` (scaffold.go:393) inserts a
provenance marker after YAML frontmatter or at the end of a
file. It is a pure function but not idempotent -- repeated
calls accumulate markers. The `Run()` function relies on
`bytes.Equal` to skip unchanged tool-owned files, but this
defense fails when markers are already baked into the embedded
assets (which happens through the asset-sync feedback loop in
this repo).

Separately, `versionMarker()` formats `"v%s"` and GoReleaser
injects `{{.Tag}}` (already `v`-prefixed), producing `vv0.6.1`.

## Goals / Non-Goals

### Goals
- Make `insertMarkerAfterFrontmatter` idempotent: any input
  produces output with exactly one marker, regardless of how
  many markers the input already contains.
- Fix the double-v version prefix in release builds.
- Clean all existing files to exactly one marker each.
- Add regression tests preventing future accumulation.

### Non-Goals
- Changing the marker format or moving to a different
  provenance mechanism.
- Making markers user-configurable.
- Changing the drift detection test architecture (approach A
  keeps byte-identical comparison).

## Decisions

### D1: Strip-then-insert (Approach A)

Add `stripExistingMarkers(s string) string` that removes all
lines matching the marker patterns:
- `<!-- scaffolded by uf ... -->`
- `# scaffolded by uf ...`

Call it at the top of `insertMarkerAfterFrontmatter` before
the existing insertion logic runs. This is the minimal change
that breaks the accumulation cycle.

**Rationale**: Approach A was chosen over Approach B
(marker-free assets with marker-aware drift test) because it
requires no changes to the drift detection architecture. The
function becomes self-correcting -- even if markers somehow
leak into assets, the output is always clean.

### D2: Fix version at GoReleaser level

Change `.goreleaser.yaml` ldflags from `{{.Tag}}` to
`{{.Version}}`. GoReleaser's `.Version` is the semver without
the `v` prefix (e.g., `0.6.1`). The `v` prefix in
`versionMarker`'s format string (`"v%s"`) then produces the
correct `v0.6.1`.

**Rationale**: Fixing at the injection point rather than in
`versionMarker()` avoids needing to handle the `dev` case
(which has no `v` prefix). The `version` variable throughout
the codebase remains a clean semver without prefix, and
display functions add `v` where appropriate.

### D3: Assets keep exactly one marker

Both embedded assets and their canonical live counterparts
retain exactly one `<!-- scaffolded by uf vdev -->` marker.
The drift test (`TestEmbeddedAssets_MatchSource`) continues
to pass with no changes because both sides are byte-identical.

When `uf init` runs:
1. Read asset (1 marker: `vdev`)
2. Strip existing markers, insert new marker (1 marker: `vX`)
3. If version matches existing file, `bytes.Equal` succeeds,
   file skipped. If version changed, content differs, file
   updated.

### D4: Marker pattern matching uses string prefix

`stripExistingMarkers` uses `strings.HasPrefix` on trimmed
lines rather than regex. The marker patterns are fixed and
well-known -- regex would add complexity without benefit.

Two patterns matched:
- `<!-- scaffolded by uf ` (HTML comment, Markdown files)
- `# scaffolded by uf ` (hash comment, YAML/shell files)

**Rationale**: Aligns with Observable Quality (constitution
III) -- the function's behavior is predictable and
inspectable without needing to reason about regex edge cases.

## Risks / Trade-offs

### R1: Blank line accumulation after stripping

When markers are stripped from between frontmatter and
content, the resulting blank lines could accumulate. The
existing insertion logic places the marker immediately after
the frontmatter closing `---\n`, so no extra blank lines are
introduced. Stripped marker lines leave no trace (the
line-by-line filter drops them entirely).

**Mitigation**: The strip function removes whole lines
including their trailing newline. No blank-line ghosts remain.

### R2: Non-embedded speckit commands

Nine speckit command files (`.opencode/command/speckit.*.md`)
carry stale markers from when they were embedded assets. They
are listed in `knownNonEmbeddedFiles` and `uf init` no longer
manages them. They need a one-time manual cleanup but will
not accumulate new markers going forward.

**Mitigation**: Include these 9 files in the cleanup task.
No code change needed since `uf init` already ignores them.
