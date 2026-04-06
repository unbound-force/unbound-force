# Contract: Path Mapping

**Spec**: 025-uf-directory-convention
**Type**: Internal contract (no external API surface)

## Purpose

Defines the complete mapping from old directory paths to
new `.uf/`-based paths. This is the authoritative reference
for all find-and-replace operations.

## Path Mapping Rules

### Rule 1: Per-Repo Root Directory

All per-repo tool directories consolidate under `.uf/`.

```
.dewey/           → .uf/dewey/
.hive/            → .uf/replicator/
.unbound-force/   → .uf/
.muti-mind/       → .uf/muti-mind/
.mx-f/            → .uf/mx-f/
```

### Rule 2: Convention Pack Directory

The convention pack directory moves from `unbound/` to
`uf/` within `.opencode/`.

```
.opencode/unbound/packs/  → .opencode/uf/packs/
```

This applies to both:
- Deployed files (`.opencode/uf/packs/`)
- Scaffold assets (`internal/scaffold/assets/opencode/uf/packs/`)

### Rule 3: Subdirectory Preservation

All subdirectories within the old paths are preserved
under the new parent:

```
.unbound-force/workflows/   → .uf/workflows/
.unbound-force/artifacts/   → .uf/artifacts/
.unbound-force/config.yaml  → .uf/config.yaml
.unbound-force/hero.json    → .uf/hero.json
.muti-mind/backlog/         → .uf/muti-mind/backlog/
.muti-mind/artifacts/       → .uf/muti-mind/artifacts/
.muti-mind/config.yaml      → .uf/muti-mind/config.yaml
.mx-f/data/                 → .uf/mx-f/data/
.mx-f/impediments/          → .uf/mx-f/impediments/
.mx-f/retros/               → .uf/mx-f/retros/
```

### Rule 4: No Backward Compatibility

- Old paths are NOT detected
- Old paths are NOT migrated
- Old paths are NOT warned about
- Old paths are NOT supported as fallbacks

## Exclusions

The following are NOT changed:

- **Historical spec documents** (`specs/001-*/` through
  `specs/024-*/`): These are archival records.
- **opencode.json Dewey command**: The `--vault .` argument
  is correct for both old and new Dewey versions.
- **`.opencode/agents/`**: Agent directory path unchanged.
- **`.opencode/command/`**: Command directory path unchanged.
- **`.opencode/skill/`**: Skill directory path unchanged.
- **`.specify/`**: Specify directory path unchanged.
- **`openspec/`**: OpenSpec directory path unchanged.

## Verification

After implementation, the following grep patterns MUST
return zero matches in production code (excluding `specs/`):

```bash
# Old per-repo directories
grep -r '\.dewey/' --include='*.go' --exclude-dir=specs
grep -r '\.hive/' --include='*.go' --exclude-dir=specs
grep -r '\.unbound-force/' --include='*.go' --exclude-dir=specs
grep -r '\.muti-mind/' --include='*.go' --exclude-dir=specs
grep -r '\.mx-f/' --include='*.go' --exclude-dir=specs

# Old convention pack path
grep -r 'opencode/unbound/' --include='*.go' --exclude-dir=specs
grep -r 'opencode/unbound/' --include='*.md' \
    internal/scaffold/assets/
```
