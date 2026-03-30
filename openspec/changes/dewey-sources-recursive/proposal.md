## Why

The `generateDeweySources()` function creates a
`disk-org` entry in `.dewey/sources.yaml` that points
to the parent directory (`../`) with no `recursive`
setting. This causes Dewey to recursively index the
entire parent directory, which:

1. **Duplicates indexing**: sibling repos are already
   indexed individually via `disk-<repo>` entries.
   The `disk-org` recursive crawl re-indexes all
   their contents.
2. **Triggers embedding failures**: large spec files
   in sibling repos exceed the Granite embedding
   model's 512-token context window, producing
   `"input length exceeds context length"` errors.
3. **Causes UUID collisions**: identical scaffolded
   files across repos produce the same deterministic
   UUID (Dewey issue #17).

The `disk-org` entry should only index top-level files
in the parent directory (design papers, orchestration
plans, gap reports) -- not recurse into sibling repos.

Additionally, `uf init --force` currently re-indexes
but does not regenerate `sources.yaml`. If a user
upgrades to a version with this fix, they need
`--force` to get the updated `sources.yaml` with
`recursive: false`.

## What Changes

1. Add `recursive: false` to the `disk-org` entry in
   `generateDeweySources()`
2. Make `uf init --force` regenerate `sources.yaml`
   before re-indexing (overwrite even if customized)

## Capabilities

### New Capabilities
- `force-regenerate-sources`: `uf init --force`
  regenerates `.dewey/sources.yaml` before re-indexing

### Modified Capabilities
- `disk-org-indexing`: The `disk-org` source entry
  now includes `recursive: false`, limiting indexing
  to top-level files in the parent directory

### Removed Capabilities
- None

## Impact

- **Files modified**: `internal/scaffold/scaffold.go`,
  `internal/scaffold/scaffold_test.go`
- **No new files**: Pure fix + small feature
- **Existing `.dewey/sources.yaml` files**: Updated
  on next `uf init --force` (not on default re-run,
  preserving idempotency)

## Constitution Alignment

### I. Autonomous Collaboration

**Assessment**: N/A

No changes to inter-hero artifact communication.

### II. Composability First

**Assessment**: PASS

Dewey remains optional. The `recursive: false` setting
doesn't affect repos without sibling directories.
Force regeneration respects the same tool availability
checks.

### III. Observable Quality

**Assessment**: PASS

Reduces false-positive embedding errors and duplicate
index entries, improving Dewey's search quality and
reliability.

### IV. Testability

**Assessment**: PASS

All changes are testable via `t.TempDir()` with
injected dependencies. Existing test patterns are
reused.
