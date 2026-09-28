---
tier: validated
compiled_at: 2026-08-14T19:54:55Z
compiled_by: claude-opus-4-6
promoted_at: 2026-08-14T20:37:01Z
promoted_by: jay-flowers
sources:
  - scaffold-1
  - scaffold-20260802T181536-jay-flowers
  - scaffold-20260802T181543-jay-flowers
  - scaffold-testing-20260812T175847-jay-flowers
  - scaffold-testing-20260812T175854-jay-flowers
topic: scaffold
related: blast-radius-scope-audit, release-pipeline
---

# Scaffold Dual-Copy Synchronization

## Current State (August 2026)

The scaffold system (`internal/scaffold/`) uses `embed.FS` to bundle canonical asset files. These assets have active copies under `.opencode/` directories. 5 learnings capture the synchronization discipline required.

## The Dual-Copy Problem

Every scaffolded file exists in two places:
1. **Source of truth**: `internal/scaffold/assets/` (embedded via `embed.FS`)
2. **Active copy**: `.opencode/commands/`, `.opencode/agents/`, `.opencode/skills/`, `.specify/scripts/bash/`

These MUST be byte-identical. The scaffold engine (`uf init`) auto-overwrites active copies from embedded assets when `isToolOwned` is true.

## Enforcement

### Test Guard
`TestEmbeddedAssetsMatchSource` enforces byte-identity at build time. Any drift between scaffold source and active copy causes test failure.

### Workflow
When editing a scaffolded file:
1. Edit the scaffold source (`internal/scaffold/assets/...`) FIRST
2. Copy to active location (`cp` to `.opencode/...`)
3. Run `go test -run TestEmbeddedAssets_MatchSource ./internal/scaffold/`
4. If adding a new asset, bump the hardcoded count in `cmd/unbound-force/main_test.go` (e.g., '39 files processed')

### Non-Embedded Files
Not all `.opencode/` files are scaffolded. The test framework maintains a `knownNonEmbeddedFiles` list for:
- OpenSpec skills (`openspec-*`)
- Replicator-scaffolded skills
- Hidden files (`.DS_Store`)

When adding `.opencode/skills` to `canonicalDirs`, all non-scaffolded skills must be listed in `knownNonEmbeddedFiles`.

## Cross-Repo Gotcha
Gaze discovered that fixing a scaffolded agent (e.g., `gaze-test-generator.md`) in the WRONG repo masks the untouched source of truth. The canonical source for gaze agents is `gaze/internal/scaffold/assets/agents/`, not `unbound-force/.opencode/agents/`. Cross-repo `Fixes:` references can close issues in the wrong repository.

## Rename Propagation
`embed.FS` renames auto-propagate to the embedded copy, BUT hardcoded strings in Cobra `Long` descriptions are NOT caught by `embed.FS`. Always grep for hardcoded path references after renames.

## Stale Comments
Test comments reflecting file counts (e.g., '6 files') become stale as specs add assets. These are caught by review council but should be updated proactively.

## Related Learnings
- Hivemind reference migration (Spec 022): `hivemind_*` → `replicator_org_*/replicator_comms_*/replicator_forge_*/dewey_*` caught by `TestScaffoldOutput_NoHivemindReferences`

## History
- scaffold-1: Core dual-copy pattern
- scaffold-20260802: Non-embedded files and canonicalDirs
- scaffold-testing-20260812: Skills directory integration
- gaze cross-repo-workflow: Wrong-repo fix anti-pattern
