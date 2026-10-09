## Context

The Unbound Force project migrated slash commands to a `/uf.*`
namespace. Active source code and current OpenSpec specs still
contain bare pre-namespace references (e.g. `/review-council`).
This change performs targeted textual updates to align those
references with the canonical `/uf.*` names.

Constitution alignment: PASS across all four principles (see
proposal.md). This change is purely textual — no behavioral,
API, or dependency changes.

## Goals / Non-Goals

### Goals

- Update all stale `/review-council` references in
  `cmd/unbound-force/main.go` help text to `/uf.review-council`.
- Update all stale `/review-council` references in
  `openspec/specs/review-council/spec.md` to `/uf.review-council`.
- Preserve all out-of-scope files unchanged (archived specs,
  changelog, migration logic, stale-ref detection tests, Speckit
  internal commands).

### Non-Goals

- Migrating Speckit internal `/speckit.*` commands (these are
  Speckit's own namespace, not Unbound Force's).
- Updating archived specs, namespace-migration specs, or
  changelog entries.
- Modifying compatibility/migration logic that maps old names
  to new names.
- Changing tests that intentionally detect stale refs.
- Adding new commands or changing command behavior.

## Decisions

1. **Textual substitution only**: Each stale reference is
   replaced with its canonical `/uf.*` equivalent. No structural
   changes to help text formatting or spec layout.

2. **Single spec file in scope**: Only
   `openspec/specs/review-council/spec.md` is modified. Other
   specs in `openspec/specs/` that reference `/review-council`
   in historical or migration context are preserved.

3. **No test changes**: Tests that detect stale refs are
   explicitly preserved. This change fixes the source of the
   staleness, not the detectors.

4. **Delta spec approach**: A delta spec documents the
   MODIFIED requirements in `openspec/specs/review-council/`.
   The delta uses ADDED/MODIFIED/REMOVED sections with RFC 2119
   language and Given/When/Then scenarios.

## Risks / Trade-offs

- **Low risk**: All changes are textual substitutions in help
  strings and spec prose. No logic, no API, no behavior changes.
- **Verification**: A simple grep for `/review-council` in
  `cmd/` and `openspec/specs/` after the change should return
  zero hits (excluding out-of-scope files).
