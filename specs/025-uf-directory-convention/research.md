# Research: Unified .uf/ Directory Convention

**Branch**: `025-uf-directory-convention` | **Date**: 2026-04-06
**Spec**: `specs/025-uf-directory-convention/spec.md`

## R1: Current Directory Layout Inventory

### Per-Repo Tool Directories (5 directories at repo root)

| Current Path | Tool | Created By | Checked By | Referenced In |
|-------------|------|-----------|-----------|---------------|
| `.dewey/` | Dewey | `dewey init` (delegated from `uf init`) | `uf doctor` checkDewey | scaffold.go, checks.go, setup.go |
| `.hive/` | Replicator | `replicator init` (delegated from `uf init`) | `uf doctor` checkReplicator | scaffold.go, checks.go |
| `.unbound-force/` | Orchestration | `uf init` initSubTools | `uf doctor` (implicit) | scaffold.go, config.go, engine.go, models.go, store.go |
| `.muti-mind/` | Muti-Mind | `mutimind init` | N/A | cmd/mutimind/main.go |
| `.mx-f/` | Mx F | `mxf collect` | N/A | internal/metrics/store.go |

### Convention Pack Directory

| Current Path | New Path | Files |
|-------------|----------|-------|
| `.opencode/unbound/packs/` | `.opencode/uf/packs/` | 9 files: go.md, go-custom.md, typescript.md, typescript-custom.md, default.md, default-custom.md, severity.md, content.md, content-custom.md |

### Scaffold Asset Directory

| Current Path | New Path |
|-------------|----------|
| `internal/scaffold/assets/opencode/unbound/packs/` | `internal/scaffold/assets/opencode/uf/packs/` |

## R2: Reference Count by File

Comprehensive inventory of all path references that must change.
Counted by grepping for old path patterns in production code.

### Go Source Files

| File | Old Pattern | Count | Notes |
|------|------------|-------|-------|
| `internal/scaffold/scaffold.go` | `.dewey/`, `.hive/`, `.unbound-force/`, `opencode/unbound/packs/` | ~15 | `initSubTools()`, `isConventionPack()`, `isDivisorAsset()`, `workflowConfigContent`, `configureOpencodeJSON()` |
| `internal/scaffold/scaffold_test.go` | `.dewey/`, `.hive/`, `.unbound-force/`, `opencode/unbound/packs/` | ~70 | `expectedAssetPaths`, test assertions, drift detection |
| `internal/doctor/checks.go` | `.dewey/`, `.hive/`, `.unbound-force/`, `opencode/unbound/packs/` | ~12 | `checkDewey()`, `checkReplicator()`, `checkScaffoldedFiles()` |
| `internal/doctor/doctor_test.go` | `.dewey/`, `.hive/` | ~14 | Test setup and assertions |
| `internal/orchestration/config.go` | `.unbound-force/` | ~3 | `LoadWorkflowConfig()`, comments |
| `internal/orchestration/engine.go` | `.unbound-force/` | ~3 | `Orchestrator` struct field comments |
| `internal/orchestration/models.go` | `.unbound-force/` | ~2 | `WorkflowInstance` comment |
| `internal/orchestration/store.go` | `.unbound-force/` | ~1 | Comment |
| `internal/orchestration/*_test.go` | `.unbound-force/` | ~5 | Test paths |
| `internal/setup/setup.go` | `.dewey/`, `.hive/` | ~3 | Comments only |
| `internal/setup/setup_test.go` | `.dewey/`, `.hive/` | ~2 | Test assertions |
| `cmd/mutimind/main.go` | `.muti-mind/` | ~3 | Default flag values |
| `internal/metrics/store.go` | `.mx-f/` | ~1 | Comment |

### Markdown Agent/Command Files (scaffold assets)

| File | Old Pattern | Count |
|------|------------|-------|
| `internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` | `.opencode/unbound/packs/` | ~2 |
| `internal/scaffold/assets/opencode/agents/divisor-*.md` (5 files) | `.opencode/unbound/packs/` | ~2 each |
| `internal/scaffold/assets/opencode/agents/muti-mind-po.md` | `.muti-mind/` | ~2 |
| `internal/scaffold/assets/opencode/agents/mx-f-coach.md` | `.mx-f/` | ~2 |
| `internal/scaffold/assets/opencode/command/review-council.md` | `.opencode/unbound/packs/` | ~1 |
| `internal/scaffold/assets/opencode/command/unleash.md` | `.unbound-force/`, `.opencode/unbound/packs/` | ~3 |
| `internal/scaffold/assets/opencode/command/workflow-*.md` | `.unbound-force/` | ~2 each |
| `internal/scaffold/assets/opencode/skill/unbound-force-heroes/SKILL.md` | `.unbound-force/` | ~2 |

### Live Agent/Command Files (deployed copies)

| File | Old Pattern |
|------|------------|
| `.opencode/agents/cobalt-crush-dev.md` | `.opencode/unbound/packs/` |
| `.opencode/agents/divisor-*.md` (8 files) | `.opencode/unbound/packs/` |
| `.opencode/agents/muti-mind-po.md` | `.muti-mind/` |
| `.opencode/agents/mx-f-coach.md` | `.mx-f/` |
| `.opencode/command/review-council.md` | `.opencode/unbound/packs/` |
| `.opencode/command/unleash.md` | `.unbound-force/` |
| `.opencode/command/workflow-*.md` | `.unbound-force/` |
| `.opencode/skill/unbound-force-heroes/SKILL.md` | `.unbound-force/` |

### Config and Documentation Files

| File | Old Pattern | Count |
|------|------------|-------|
| `opencode.json` | `--vault .` (Dewey serve) | 1 |
| `.gitignore` | `.unbound-force/`, `.dewey/` | 2 |
| `AGENTS.md` | `.unbound-force/`, `.dewey/`, `.hive/`, `.muti-mind/`, `.mx-f/`, `.opencode/unbound/packs/` | ~10 |
| `schemas/hero-manifest/v1.0.0.schema.json` | `.unbound-force/hero.json` | 1 |
| `schemas/acceptance-decision/samples/*.json` | `.unbound-force/artifacts/` | 1 |
| `scripts/validate-hero-contract.sh` | `.unbound-force/hero.json` | ~10 |

## R3: Dewey Workspace Path Change

The Dewey `serve` command currently uses `--vault .` which
tells Dewey to use the repo root as its vault. The workspace
directory (`.dewey/`) is created by `dewey init` at the
repo root.

**Current**: `dewey serve --vault .` → creates/uses `.dewey/`
**New**: `dewey serve --vault .` → creates/uses `.uf/dewey/`

This requires Dewey to change its default workspace directory
from `.dewey/` to `.uf/dewey/`. This is tracked in
`unbound-force/dewey#33`.

The `opencode.json` Dewey MCP entry does NOT need to change
its `--vault .` argument — the vault path stays the same.
What changes is Dewey's internal default for where it stores
its workspace data within the vault.

## R4: Replicator Directory Change

The Replicator currently uses `.hive/` for per-repo data.
`replicator init` creates `.hive/` at the repo root.

**Current**: `replicator init` → creates `.hive/`
**New**: `replicator init` → creates `.uf/replicator/`

This requires Replicator to change its default per-repo
directory. This is tracked in `unbound-force/replicator#9`.

## R5: Scaffold Asset Directory Rename

The convention pack assets must be moved from:
```
internal/scaffold/assets/opencode/unbound/packs/
```
to:
```
internal/scaffold/assets/opencode/uf/packs/
```

This is a `git mv` operation. The embedded filesystem
(`go:embed assets`) will automatically pick up the new
paths. The `isConventionPack()` function must be updated
to check `opencode/uf/packs/` instead of
`opencode/unbound/packs/`.

## R6: No Backward Compatibility

Per the spec, there is no migration path. Old directories
are ignored. This simplifies implementation significantly:

- No detection of old directories
- No migration logic
- No dual-path support
- No deprecation warnings

Users upgrading simply `rm -rf` old directories and re-run
`uf init`.

## R7: Cross-Repo Dependencies

Two external repos must be updated before this spec can
be fully implemented:

1. **Dewey** (`unbound-force/dewey#33`): Must change
   default workspace from `.dewey/` to `.uf/dewey/`.
   Hard blocker for `uf init` and `uf doctor`.

2. **Replicator** (`unbound-force/replicator#9`): Must
   change default per-repo dir from `.hive/` to
   `.uf/replicator/`. Hard blocker for `uf init` and
   `uf doctor`.

**Implementation strategy**: The unbound-force repo changes
can be implemented and tested independently by mocking the
external tool behavior. The actual integration testing
requires the external repos to be updated first.

## R8: Test Impact Analysis

### Tests That Must Be Updated

| Test File | Impact | Reason |
|-----------|--------|--------|
| `internal/scaffold/scaffold_test.go` | HIGH | `expectedAssetPaths` list, drift detection, all path assertions |
| `internal/doctor/doctor_test.go` | HIGH | `.dewey/`, `.hive/` directory setup, path assertions |
| `internal/orchestration/config_test.go` | MEDIUM | `.unbound-force/config.yaml` path |
| `internal/orchestration/engine_test.go` | MEDIUM | Workflow directory paths |
| `internal/orchestration/store_test.go` | LOW | Directory path in test setup |
| `internal/setup/setup_test.go` | LOW | Comment references, minor assertions |
| `cmd/mutimind/main_test.go` | LOW | Default flag value assertions |

### Regression Tests to Add

1. **No old path references**: A test that greps all Go
   source and scaffold assets for old path patterns
   (`.dewey/`, `.hive/`, `.unbound-force/`, `.muti-mind/`,
   `.mx-f/`, `opencode/unbound/`) and fails if any are
   found (excluding `specs/` historical documents).

2. **Scaffold asset path consistency**: Verify all
   scaffold assets under `opencode/uf/packs/` deploy
   correctly.

## R9: Workflow Config Content Change

The `workflowConfigContent` constant in `scaffold.go`
currently generates a file with the header comment
`# .unbound-force/config.yaml`. This must change to
`# .uf/config.yaml`.

The `initSubTools()` function creates this file at
`.unbound-force/config.yaml`. This path must change to
`.uf/config.yaml`.

## R10: opencode.json Dewey Serve Command

The current `opencode.json` has:
```json
"command": ["dewey", "serve", "--vault", "."]
```

The `--vault .` argument tells Dewey to use the current
directory as the vault root. Dewey then creates its
workspace at `{vault}/.dewey/`. After the Dewey change
(#33), it will create at `{vault}/.uf/dewey/`.

**Decision**: The `opencode.json` command does NOT need
to change. The `--vault .` argument is correct for both
old and new Dewey versions. The workspace subdirectory
is Dewey's internal concern.

However, the `configureOpencodeJSON()` function in
`scaffold.go` hardcodes the Dewey entry. This entry
remains unchanged — no code change needed here.
