# Implementation Plan: Unified .uf/ Directory Convention

**Branch**: `025-uf-directory-convention` | **Date**: 2026-04-06 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/025-uf-directory-convention/spec.md`

## Summary

Consolidate all per-repo tool directories (`.dewey/`,
`.hive/`, `.unbound-force/`, `.muti-mind/`, `.mx-f/`)
under a single `.uf/` directory and rename the convention
pack path from `.opencode/unbound/packs/` to
`.opencode/uf/packs/`. This is a mechanical path renaming
operation affecting ~260 references across ~30 files with
no logic changes, no new features, and no backward
compatibility. Two external dependencies (Dewey #33,
Replicator #9) must support the new paths.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra` (CLI), `embed.FS` (scaffold), `github.com/charmbracelet/log` (logging), `gopkg.in/yaml.v3` (config parsing)
**Storage**: Filesystem only (JSON workflow state, YAML config, Markdown agent files)
**Testing**: Standard library `testing` package, `t.TempDir()` for isolation, `-race -count=1`
**Target Platform**: macOS, Linux (darwin/arm64, linux/amd64)
**Project Type**: CLI tool (meta-repository)
**Performance Goals**: N/A (path renaming, no runtime performance impact)
**Constraints**: No backward compatibility — clean cut, no migration
**Scale/Scope**: ~260 path references across ~30 files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check

| Principle | Status | Rationale |
|-----------|--------|-----------|
| I. Autonomous Collaboration | **PASS** | Path renaming does not affect inter-hero artifact communication. The artifact envelope format is unchanged. Heroes continue to produce and consume artifacts at the new paths using the same envelope schema. |
| II. Composability First | **PASS** | Each hero's data directory moves under `.uf/<hero-name>/` but remains independently accessible. Heroes that are not installed simply don't create their subdirectory. No mandatory dependencies are introduced. |
| III. Observable Quality | **PASS** | All machine-parseable output formats (JSON, YAML) are unchanged. Only filesystem paths change. Provenance metadata in artifacts is unaffected. |
| IV. Testability | **PASS** | All path-dependent functions remain testable via `t.TempDir()` and injected dependencies (`LookPath`, `ExecCmd`). No external services required. Coverage strategy defined below. |

### Post-Design Re-Check

| Principle | Status | Rationale |
|-----------|--------|-----------|
| I. Autonomous Collaboration | **PASS** | No changes to artifact format or inter-hero communication. Path changes are internal to each tool's configuration. |
| II. Composability First | **PASS** | The `.uf/` directory structure preserves per-tool isolation. Each subdirectory is created independently by its owning tool. Missing tools result in missing subdirectories — no errors. |
| III. Observable Quality | **PASS** | Doctor output uses new paths. JSON output format unchanged. Schema descriptions updated to reference new paths. |
| IV. Testability | **PASS** | All changes are testable with existing patterns. New regression test prevents old path references from being reintroduced. Coverage strategy: unit tests for all path-dependent functions, regression test for old path grep. |

## Coverage Strategy

Per Constitution Principle IV, this section defines the
testing approach for all changes:

### Unit Tests (existing, updated)

- `scaffold_test.go`: Update `expectedAssetPaths` for
  `opencode/uf/packs/*`, update all path assertions in
  `initSubTools` tests, update drift detection baselines.
- `doctor_test.go`: Update directory creation and path
  assertions for `.uf/dewey/`, `.uf/replicator/`,
  `.opencode/uf/packs/`.
- `config_test.go`: Update `.unbound-force/` → `.uf/`
  in test paths.
- `engine_test.go`: Update workflow directory paths.
- `store_test.go`: Update store directory paths.

### Regression Tests (new)

- `TestScaffoldOutput_NoOldPathReferences`: Grep all
  scaffold asset content for old path patterns. Fail if
  any found. Patterns: `.dewey/`, `.hive/`,
  `.unbound-force/`, `.muti-mind/`, `.mx-f/`,
  `opencode/unbound/`.

### Integration Verification

- `make test` must pass with zero failures.
- `make check` (build + test + vet + lint) must pass.

## Project Structure

### Documentation (this feature)

```text
specs/025-uf-directory-convention/
├── plan.md              # This file
├── research.md          # Path inventory and analysis
├── data-model.md        # Path mapping table
├── quickstart.md        # Implementation approach summary
├── contracts/
│   ├── path-mapping.md  # Authoritative path mapping
│   ├── scaffold-changes.md  # Scaffold engine changes
│   └── doctor-changes.md    # Doctor package changes
└── tasks.md             # Task breakdown (created by /speckit.tasks)
```

### Source Code (repository root)

```text
internal/
├── scaffold/
│   ├── scaffold.go          # Path constants, isConventionPack, initSubTools
│   ├── scaffold_test.go     # expectedAssetPaths, path assertions
│   └── assets/
│       └── opencode/
│           └── uf/          # Renamed from unbound/
│               └── packs/   # 9 convention pack files
├── doctor/
│   ├── checks.go            # checkDewey, checkReplicator, checkScaffoldedFiles
│   └── doctor_test.go       # Path assertions
├── orchestration/
│   ├── config.go            # LoadWorkflowConfig path
│   ├── engine.go            # Orchestrator field comments
│   ├── models.go            # WorkflowInstance comment
│   └── store.go             # Store comment
├── setup/
│   └── setup.go             # Comment references
└── metrics/
    └── store.go             # Comment reference

cmd/
└── mutimind/
    └── main.go              # Default flag values

scripts/
└── validate-hero-contract.sh  # hero.json path

schemas/
├── hero-manifest/
│   └── v1.0.0.schema.json   # Description text
└── acceptance-decision/
    └── samples/              # Sample path reference
```

**Structure Decision**: No new packages or directories.
All changes are in-place modifications to existing files
plus a directory rename (`git mv`) for scaffold assets.

## Implementation Phases

### Phase 1: Scaffold Engine + Asset Rename (Core)

The scaffold engine is the foundation — it determines
where files are deployed. Must be done first.

1. `git mv` asset directory
2. Update `isConventionPack()`, `isDivisorAsset()` comment
3. Update `workflowConfigContent`
4. Update `initSubTools()` paths
5. Update `generateDeweySources()` path
6. Update `scaffold_test.go`

### Phase 2: Doctor Checks

Doctor validates the environment. Must use new paths
to produce correct diagnostics.

1. Update `checkDewey()` workspace path
2. Update `checkReplicator()` path and display name
3. Update `checkScaffoldedFiles()` packs path
4. Update `doctor_test.go`

### Phase 3: Orchestration + Hero CLIs

Internal packages that reference the old paths.

1. Update orchestration config, engine, models, store
2. Update orchestration tests
3. Update `cmd/mutimind/main.go` defaults
4. Update `internal/metrics/store.go` comment

### Phase 4: Agent/Command Markdown Files

Scaffold asset copies and live deployed files.

1. Update all agent `.md` files (pack path references)
2. Update all command `.md` files (workflow path references)
3. Update skill SKILL.md
4. Sync scaffold copies with canonical sources

### Phase 5: Config, Docs, Scripts, Schemas

Everything else.

1. Update `.gitignore`
2. Update `AGENTS.md`
3. Update `scripts/validate-hero-contract.sh`
4. Update schema descriptions
5. Add regression test
6. Final verification: `make check`

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Missing a path reference | Medium | Low | Regression test greps for old patterns |
| Dewey #33 not ready | High | Medium | Mock Dewey behavior in tests; code changes are independent |
| Replicator #9 not ready | High | Medium | Mock Replicator behavior in tests; code changes are independent |
| Scaffold drift after rename | Low | Low | Existing drift detection tests catch this |
| Breaking live agent files | Low | Medium | Sync scaffold copies after updating canonical sources |

## Complexity Tracking

No constitution violations. No complexity justifications needed.
All changes are mechanical path renames within existing patterns.
<!-- scaffolded by uf vdev -->
