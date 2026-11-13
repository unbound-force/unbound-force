## Why

Artifact provenance records only the ref an artifact was built
from (`provenance.Stamp` returns `Record{Ref}`). A ref name is
mutable: `main` points at a different commit after every merge,
so a record saying `ref: main` cannot tell a reader which source
produced the artifact. Provenance that does not identify an
immutable revision does not meet Observable Quality.

## What Changes

- Add `refs.Lookup(dir, ref)`, which resolves a ref name to the
  commit object name it currently points at in the repository at
  `dir`, peeling annotated tags, and supporting both SHA-1 and
  SHA-256 repositories.
- Change `provenance.Stamp(ref) Record` to
  `provenance.Stamp(dir, ref) (Record, error)`, adding a `Commit`
  field resolved through `refs.Lookup`. Stamp fails instead of
  returning a record without a commit.

`Lookup` receives ref names that may come from untrusted input
(PR branch names, tags). It MUST pass the name to git as a single
argument, never through a shell, and never in a position git can
read as an option.

## Capabilities

### New Capabilities
- `ref-resolution`: resolve a ref name to a commit object name.
- `artifact-provenance`: provenance records carry the resolved
  commit alongside the ref name.

### Modified Capabilities
- None. `artifact-provenance` has no canonical spec yet, so its
  commit-pinning requirement is added as a new capability (see
  below).

### Removed Capabilities
- None.

## Impact

- `internal/refs`: new `lookup.go` and tests.
- `internal/provenance`: `Stamp` signature and `Record` shape
  change. `Stamp` has no callers outside its package yet, so no
  call sites change.
- Runtime requirement: git 2.24 or later (`--end-of-options`).
- No new module dependencies. No coverage threshold changes;
  both packages already sit in the helper coverage scope.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0,
.specify/memory/constitution.md), all six principles.

### I. Autonomous Collaboration

**Assessment**: PASS

Provenance records are the artifact-level contract other heroes
read. Adding an immutable `commit` field makes records
self-describing without any runtime coupling between producers
and consumers.

### II. Composability First

**Assessment**: PASS

No new dependencies. `refs.Lookup` uses the local `git` binary
the project already requires and is usable on its own.

### III. Observable Quality

**Assessment**: PASS

This is the motivating principle: records gain a commit object
name, so every artifact's provenance identifies the exact source
revision. The JSON shape is `{"ref": ..., "commit": ...}`.

### IV. Testability

**Assessment**: PASS

Both packages are tested in isolation against repositories
created in `t.TempDir()`; no network or external service is
used. Tests assert the resolved value against git's own report,
not only its format. See design.md, Coverage Strategy.

### V. Security by Default

**Assessment**: PASS

Ref names are external input reaching process execution. They are
passed in argv form after `--end-of-options`, never through a
shell, and rev-parse output is validated against an exact
object-name pattern before use (design.md D2, D4). No dependency
is added; `git` is invoked by name from PATH, as the toolchain
already requires.

### VI. Deterministic Execution

**Assessment**: N/A

No command, agent, or skill prose changes. The change is Go code
with tests; there is no mechanical procedure described in prose.
