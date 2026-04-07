## Why

`uf init` scaffolds files under `.opencode/`, `.specify/`,
and `openspec/`, creates `.uf/config.yaml`, and delegates
to sub-tools — but does not touch `.gitignore`. This
means every project that runs `uf init` needs the engineer
to manually add ignore patterns for `.uf/` runtime data
(databases, caches, locks) and legacy tool directories
(`.dewey/`, `.hive/`, `.unbound-force/`, etc.).

Without these patterns, runtime files get accidentally
committed (as happened with `.dewey/graph.db` and
`.dewey/cache/` in the unbound-force repo itself), and
legacy directories from pre-Spec-025 installations can
pollute the repo.

## What Changes

### New Capabilities

- `ensureGitignore()`: New function in the scaffold
  engine that appends a standard Unbound Force ignore
  block to `.gitignore` during `uf init`. The block
  covers `.uf/` runtime data and legacy tool directories.

### Modified Capabilities

- `Run()` in `scaffold.go`: Calls `ensureGitignore()`
  after file scaffolding but before sub-tool delegation.
  The result is included in the scaffold summary output.

### Removed Capabilities

None.

## Impact

- `internal/scaffold/scaffold.go` — new function +
  call site in `Run()`
- `internal/scaffold/scaffold_test.go` — 4 new tests
- `AGENTS.md` — Recent Changes entry
- No changes to scaffold assets, agent files, or
  convention packs

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The gitignore management is a self-contained scaffold
operation. It produces a self-describing output (the
ignore block includes a marker comment explaining its
purpose and origin).

### II. Composability First

**Assessment**: PASS

The function is append-only and idempotent. It does not
interfere with existing `.gitignore` content. Projects
that do not use `uf init` are unaffected.

### III. Observable Quality

**Assessment**: N/A

This change produces a text file modification, not a
machine-parseable artifact. The marker comment provides
provenance ("managed by uf init").

### IV. Testability

**Assessment**: PASS

All scenarios are testable in isolation using
`t.TempDir()`: fresh directory, existing `.gitignore`
without UF block, existing `.gitignore` with UF block
(idempotent). No external services required.
