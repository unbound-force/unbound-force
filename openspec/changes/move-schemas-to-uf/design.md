## Context

The `uf init` scaffold deploys embedded JSON schema assets (review
dispatch, review verdict, envelope, hero manifest, etc.) to a
top-level `schemas/` directory in consumer projects. This collides
with projects that have their own `schemas/` directory, like
`fullsend-ai/agents`.

The scaffold already maps `uf/` assets to `.uf/` in the target
project via `mapAssetPath`. Moving schema assets from the embedded
`assets/schemas/` root to `assets/uf/schemas/` leverages this
existing path mapping — `uf/` → `.uf/` — so schemas deploy to
`.uf/schemas/` without any new mapping code.

## Goals / Non-Goals

### Goals

- Eliminate namespace collision between scaffolded schemas and
  host-project `schemas/` directories
- Leverage existing `uf/` → `.uf/` path mapping rather than adding
  new deployment logic
- Maintain all existing scaffold behavior (overwrite-on-diff,
  stealth exclusion, Divisor-only filtering)

### Non-Goals

- Changing schema `$id` values (schemas remain self-identifying
  regardless of filesystem path)
- Modifying `internal/schemas/` Go package or repo root `schemas/`
  directory
- Auto-migrating existing `schemas/` deployments on consumer upgrade
- Adding any new scaffold features or changing scaffolding semantics

## Decisions

### D1: Move embedded assets, don't add a new mapping prefix

**Decision**: Move `internal/scaffold/assets/schemas/` →
`internal/scaffold/assets/uf/schemas/` and delete the `schemas/`
case from `mapAssetPath` entirely.

**Rationale**: The `uf/` → `.uf/` mapping already exists in
`mapAssetPath`. Adding schemas under the `uf/` asset path means
they inherit this mapping without any new code. This is simpler
than maintaining a separate `schemas/` prefix that does a
pass-through mapping.

**Alternatives considered**:
- Keep the `schemas/` prefix but change `mapAssetPath` to emit
  `.uf/schemas/`: More code, confusing embedded dir structure
- Add a new top-level `.schemas/` prefix: No benefit over `.uf/`

### D2: Remove schemas/ from stealthExcludePaths

**Decision**: Delete `"schemas/"` from `stealthExcludePaths` in
`scaffold_stealth.go`.

**Rationale**: After the move, all scaffolded schemas live under
`.uf/schemas/`, which is already covered by the `.uf/` entry in
`stealthExcludePaths`. The explicit `"schemas/"` entry becomes
redundant and would exclude a top-level `schemas/` that no longer
belongs to the scaffold.

### D3: No migration of existing deployments

**Decision**: Consumer repos that already have `schemas/` deployed
from a prior `uf init` are not automatically cleaned up.

**Rationale**: The scaffold does not track which files it wrote
beyond the `isToolOwned` predicate. Removing a `schemas/` directory
could destroy user data if they placed their own schemas in the
same directory. Users can manually remove scaffolded schemas if
they were the only contents.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Assets that previously loaded schemas from `schemas/` may need path updates | Schema `$id` values are absolute URIs, not filesystem paths. Consumer code references schemas by `$id`, not filesystem path. |
| Stealth mode no longer excludes a top-level `schemas/` | If a consumer project has its own `schemas/` that was previously hidden by stealth, it will become visible after upgrade. This is intentional: the scaffold should not hide user-owned directories. |
| Test path string updates are mechanical but numerous (~50) | Each test assertion is a literal string; there is no shared constant. The update is straightforward search-and-replace. |
<!-- scaffolded by uf vdev -->