## Why

The scaffold engine's `insertMarkerAfterFrontmatter` function
appends a provenance marker (`<!-- scaffolded by uf vX.Y.Z -->`)
every time it processes a file, without checking whether one
already exists. Combined with the asset-to-canonical sync cycle
in this repo (drift test enforces byte-identical copies), markers
accumulate with each `uf init` run. Currently 21 live files and
12 embedded assets carry 3-6 duplicate markers each.

A secondary bug compounds the issue: GoReleaser injects
`{{.Tag}}` (which includes a `v` prefix, e.g., `v0.6.1`) into
the `version` variable, and `versionMarker()` adds another `v`
prefix via its `"v%s"` format string, producing `vv0.6.1`.

## What Changes

Make `insertMarkerAfterFrontmatter` idempotent by stripping
existing markers before inserting the new one. Fix the double-v
version prefix. Clean up all existing files.

## Capabilities

### New Capabilities
- `stripExistingMarkers`: Helper function that removes all
  scaffold provenance marker lines from content, regardless
  of version or comment format (HTML or hash).

### Modified Capabilities
- `insertMarkerAfterFrontmatter`: Now calls
  `stripExistingMarkers` before inserting, ensuring output
  always contains exactly one marker.
- `versionMarker`: No code change to this function; the
  double-v is fixed at the GoReleaser ldflags level by
  switching from `{{.Tag}}` to `{{.Version}}`.

### Removed Capabilities
- None.

## Impact

- **internal/scaffold/scaffold.go**: New
  `stripExistingMarkers` function; modified
  `insertMarkerAfterFrontmatter` call site.
- **internal/scaffold/scaffold_test.go**: Updated "double
  insert" test expectation (now idempotent). New tests for
  `stripExistingMarkers` and embedded asset marker count
  regression.
- **.goreleaser.yaml**: Two ldflags lines changed
  (`{{.Tag}}` to `{{.Version}}`).
- **30 Markdown files**: Mechanical cleanup -- strip
  duplicate markers to exactly one per file. Affects
  `.opencode/command/`, `.opencode/uf/packs/`,
  `.opencode/skill/`, and their embedded asset copies
  under `internal/scaffold/assets/`.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is internal to the scaffold engine. It does not
affect inter-hero artifact communication or self-describing
outputs. Provenance markers are deployment metadata, not
inter-hero artifacts.

### II. Composability First

**Assessment**: PASS

The fix is entirely within the scaffold engine. No new
dependencies are introduced. Standalone functionality is
preserved -- `uf init` continues to work identically from
the user's perspective, just without accumulating garbage
markers.

### III. Observable Quality

**Assessment**: PASS

Provenance markers are a form of observable quality metadata.
This change ensures they are correct (single marker, accurate
version) rather than noisy (multiple markers, malformed
version). Machine parseability improves because consumers of
these markers can rely on exactly one per file.

### IV. Testability

**Assessment**: PASS

`stripExistingMarkers` is a pure function (string in, string
out) -- trivially testable in isolation. The modified
`insertMarkerAfterFrontmatter` remains a pure function. New
tests cover idempotency, marker stripping, and embedded asset
regression. No external services or shared state required.
