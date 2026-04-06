# Contract: Scaffold Engine Changes

**Spec**: 025-uf-directory-convention
**Type**: Internal contract

## Purpose

Defines the exact changes required in the scaffold engine
(`internal/scaffold/scaffold.go`) and its test file.

## Function Changes

### isConventionPack(relPath string) bool

**Before**: `strings.HasPrefix(relPath, "opencode/unbound/packs/")`
**After**: `strings.HasPrefix(relPath, "opencode/uf/packs/")`

### isDivisorAsset(relPath string) bool

No direct path change — delegates to `isConventionPack()`.
Verify the comment referencing `opencode/unbound/packs/`
is updated to `opencode/uf/packs/`.

### workflowConfigContent constant

**Before**: `# .unbound-force/config.yaml`
**After**: `# .uf/config.yaml`

### initSubTools(opts *Options) []subToolResult

Path changes:
1. `.unbound-force` → `.uf` (workflow config directory)
2. `.unbound-force/config.yaml` → `.uf/config.yaml` (config path)
3. `.dewey` → `.uf/dewey` (Dewey workspace check — NOTE:
   this depends on Dewey #33 changing where `dewey init`
   creates its workspace)
4. `.hive` → `.uf/replicator` (Replicator workspace check —
   NOTE: this depends on Replicator #9 changing where
   `replicator init` creates its workspace)

Display name changes in subToolResult:
- `.unbound-force/config.yaml` → `.uf/config.yaml`
- `.dewey/` → `.uf/dewey/`
- `.hive/` → `.uf/replicator/`

### generateDeweySources(opts *Options, force bool)

**Before**: `filepath.Join(opts.TargetDir, ".dewey", "sources.yaml")`
**After**: `filepath.Join(opts.TargetDir, ".uf", "dewey", "sources.yaml")`

### configureOpencodeJSON(opts *Options)

No path changes needed. The Dewey MCP entry uses
`["dewey", "serve", "--vault", "."]` which is correct
for both old and new Dewey versions.

## Asset Directory Rename

```bash
git mv internal/scaffold/assets/opencode/unbound \
       internal/scaffold/assets/opencode/uf
```

This moves the entire `unbound/` directory (containing
`packs/`) to `uf/`. All 9 pack files move with it.

## Test Changes

### expectedAssetPaths

All entries matching `opencode/unbound/packs/*` must be
updated to `opencode/uf/packs/*`.

### TestScaffoldOutput_NoGraphthulhuReferences

No change (unrelated).

### TestScaffoldOutput_NoBareUnboundReferences

This test greps for bare "unbound" references. It must
be updated to account for the new `opencode/uf/packs/`
path (which no longer contains "unbound").

### TestScaffoldOutput_NoHivemindReferences

No change (unrelated).

### New Test: TestScaffoldOutput_NoOldPathReferences

Add a regression test that greps all scaffold asset
content for old path patterns:
- `.dewey/`
- `.hive/`
- `.unbound-force/`
- `.muti-mind/`
- `.mx-f/`
- `opencode/unbound/`

Excludes: version markers, historical references in
spec documents.

## Coverage Strategy

- **Unit tests**: All path-dependent functions tested
  with the new paths via existing test patterns
  (t.TempDir, injected LookPath/ExecCmd).
- **Drift detection**: Existing drift detection tests
  continue to verify scaffold assets match canonical
  sources.
- **Regression test**: New grep-based test prevents
  old path references from being reintroduced.
