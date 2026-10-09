## Why

`uf init` deploys embedded scaffold JSON schemas to a top-level
`schemas/` directory in the target project. This collides with
projects that already have their own `schemas/` directory (e.g.,
`fullsend-ai/agents`), creating merge conflicts and namespace
pollution.

The `uf init` scaffold already supports deploying `uf/` assets to
`.uf/` in the target project — a hidden, non-colliding namespace.
Moving the embedded scaffold schemas from
`internal/scaffold/assets/schemas/` to
`internal/scaffold/assets/uf/schemas/` leverages this existing
mechanism, deploying schemas to `.uf/schemas/` instead of
`schemas/`.

The canonical schema source in the repo root `schemas/` and the
`internal/schemas/` Go package are NOT affected. The embedded
scaffold assets are a deployed subset for consumer repos, not the
source of truth.

## What Changes

- Move embedded scaffold assets from `internal/scaffold/assets/schemas/`
  to `internal/scaffold/assets/uf/schemas/`
- Update scaffold path mapping logic (`scaffold.go`) to remove the
  `schemas/` prefix from asset prefix lists and path mapping
- Update stealth exclusion logic (`scaffold_stealth.go`) to remove
  `schemas/` from exclusion paths (already covered by `.uf/`)
- Update scaffold tests (`scaffold_test.go`) with new path strings

## Capabilities

### Modified Capabilities
- `uf init`: Schema deployment target changes from `schemas/` to
  `.uf/schemas/` in consumer projects

## Impact

- **scaffold.go**: Remove `schemas/` from `knownAssetPrefixes`,
  delete `schemas/` case from `mapAssetPath`, update `isToolOwned`
  and `isDivisorAsset` prefix checks
- **scaffold_stealth.go**: Remove `schemas/` from `stealthExcludePaths`
- **scaffold_test.go**: Update ~50 path string references
- **Consumer repos**: After upgrading, `uf init` deploys schemas to
  `.uf/schemas/` instead of `schemas/`. Existing `schemas/`
  deployments are not automatically cleaned up.
- **No impact**: `internal/schemas/` Go package, repo root
  `schemas/` directory, schema `$id` values, CI/CD workflows

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is internal scaffolding infrastructure. It does not
affect hero-to-hero communication, artifact exchange formats, or
inter-agent coupling.

### II. Composability First

**Assessment**: PASS

By deploying scaffold schemas to `.uf/schemas/` instead of a
top-level `schemas/`, we avoid namespace collisions with host
projects' own schemas directories. This improves standalone
deployability: `uf init` no longer risks clobbering or merging
with an unrelated project directory.

### III. Observable Quality

**Assessment**: PASS

Schemas remain machine-parseable JSON. Provenance metadata is
unchanged. The only difference is the deployment path in the
consumer repo.

### IV. Testability

**Assessment**: PASS

Existing `scaffold_test.go` coverage is maintained with updated
path strings. The change is a path relocation, not new logic, so
test scope is proportional. No new external dependencies are
introduced.

### V. Security by Default

**Assessment**: PASS

The `.uf/` directory is already hidden by convention (dot-prefixed),
which is a slight security improvement over the visible `schemas/`
directory. No new dependencies or privilege changes.
<!-- scaffolded by uf vdev -->