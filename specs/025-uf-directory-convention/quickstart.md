# Quickstart: Unified .uf/ Directory Convention

**Branch**: `025-uf-directory-convention` | **Date**: 2026-04-06

## What This Changes

All per-repo tool directories consolidate under a single
`.uf/` directory. Convention packs move from
`.opencode/unbound/packs/` to `.opencode/uf/packs/`.

## Implementation Approach

This is a mechanical find-and-replace operation across
~30 files with ~260 path references. No new features,
no new data structures, no new packages.

### Phase 1: Scaffold Engine + Asset Rename

1. `git mv internal/scaffold/assets/opencode/unbound internal/scaffold/assets/opencode/uf`
2. Update `isConventionPack()` path prefix
3. Update `isDivisorAsset()` comment
4. Update `workflowConfigContent` header comment
5. Update `initSubTools()` paths: `.unbound-force/` → `.uf/`, `.dewey/` → `.uf/dewey/`, `.hive/` → `.uf/replicator/`
6. Update `generateDeweySources()` path
7. Update `scaffold_test.go` expectedAssetPaths and assertions

### Phase 2: Doctor Checks

1. Update `checkDewey()` workspace path
2. Update `checkReplicator()` `.hive/` → `.uf/replicator/`
3. Update `checkScaffoldedFiles()` packs path
4. Update `doctor_test.go` assertions

### Phase 3: Orchestration + Hero CLIs

1. Update `Orchestrator` struct comments
2. Update `config.go` path and comments
3. Update `models.go` comments
4. Update orchestration tests
5. Update `cmd/mutimind/main.go` default flag values
6. Update `internal/metrics/store.go` comment

### Phase 4: Agent/Command Markdown Files

1. Update all scaffold asset `.md` files
2. Update all live `.opencode/` agent/command files
3. Sync scaffold copies with canonical sources

### Phase 5: Config, Docs, Scripts

1. Update `.gitignore`
2. Update `opencode.json` (if needed)
3. Update `AGENTS.md`
4. Update `scripts/validate-hero-contract.sh`
5. Update schema descriptions
6. Add regression test for old path references

## Verification

```bash
# All tests pass
make test

# No old path references in production code
grep -r '\.dewey/' --include='*.go' --exclude-dir=specs | wc -l  # 0
grep -r '\.hive/' --include='*.go' --exclude-dir=specs | wc -l   # 0
grep -r '\.unbound-force/' --include='*.go' --exclude-dir=specs | wc -l  # 0
grep -r 'opencode/unbound/' --include='*.go' --exclude-dir=specs | wc -l  # 0
```

## Dependencies

- **Dewey #33**: Must support `.uf/dewey/` workspace path
- **Replicator #9**: Must support `.uf/replicator/` data path

Both are hard blockers for integration testing but NOT
for code changes in this repo. The code changes can be
implemented and unit-tested independently.

## Risk Assessment

**Low risk**: All changes are mechanical path renames.
No logic changes, no new features, no behavioral changes.
The primary risk is missing a reference — mitigated by
the regression test that greps for old patterns.
