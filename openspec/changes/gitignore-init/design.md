## Context

`uf init` manages several file types during scaffolding:
embedded assets (deployed via `embed.FS`), generated
config files (`opencode.json`, `.uf/config.yaml`), and
sub-tool delegation (`dewey init`, `replicator init`).
It does not currently manage `.gitignore`.

The `.uf/` directory contains a mix of tracked files
(config, backlog, manifests) and runtime data (databases,
caches, locks). Without proper `.gitignore` entries,
runtime files get committed. Legacy tool directories
(`.dewey/`, `.hive/`, etc.) from pre-Spec-025
installations also need to be ignored.

## Goals / Non-Goals

### Goals

- Ensure `.gitignore` has the correct UF ignore patterns
  after `uf init` runs
- Append-only — never overwrite or remove existing
  `.gitignore` content
- Idempotent — running `uf init` multiple times does not
  duplicate the ignore block
- Create `.gitignore` if it does not exist

### Non-Goals

- Do not make `.gitignore` a scaffold embedded asset
  (every project's `.gitignore` is different)
- Do not parse or validate existing `.gitignore` entries
- Do not remove legacy ignore patterns if they exist

## Decisions

### D1: Marker-based idempotency

Use a marker comment to detect whether the UF ignore
block is already present:

```gitignore
# Unbound Force — managed by uf init
```

If this marker exists anywhere in `.gitignore`, skip
the append. This is simpler and more reliable than
checking for individual patterns.

### D2: Append-only, never modify existing content

The function appends to the end of `.gitignore` with a
blank line separator. It never reads, parses, or modifies
existing content beyond checking for the marker string.

### D3: Generated in code, not an embedded asset

The ignore block is a Go string constant in
`scaffold.go`, not a file under
`internal/scaffold/assets/`. This avoids the complexity
of merging an asset file with an existing `.gitignore`.

### D4: Called from Run() in the config section

The function is called after file scaffolding (assets
deployed) but before sub-tool delegation (dewey init,
replicator init). This ensures `.gitignore` is ready
before sub-tools create runtime files.

## Risks / Trade-offs

### Risk: Stale patterns if .uf/ structure changes

If new runtime directories are added under `.uf/` in
the future, the ignore block needs updating. Mitigated
by the marker — the block can be extended in future
versions (though the idempotency check means existing
blocks won't be updated automatically).

### Trade-off: No update mechanism for existing blocks

If the ignore block content changes between versions,
existing `.gitignore` files keep the old block. This is
acceptable — the patterns are additive, and removing
the marker + re-running `uf init` would regenerate.
