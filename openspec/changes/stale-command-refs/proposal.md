## Why

Active source code and current OpenSpec specs still reference bare
pre-namespace slash-command names (e.g. `/review-council`) instead
of the canonical `/uf.*` names (e.g. `/uf.review-council`). This
creates user confusion: the documented commands do not match the
commands users actually invoke. GitHub issue #625 requests fixing
these stale references.

## What Changes

Update stale pre-namespace command references to canonical `/uf.*`
names in two locations:

1. **`cmd/unbound-force/main.go`** (line 99): Help text mentions
   `/review-council` instead of `/uf.review-council`.
2. **`openspec/specs/review-council/spec.md`**: Six occurrences of
   `/review-council` in requirement descriptions and Given/When/Then
   scenarios.

Scope is limited to active source/help text and current OpenSpec
specs that describe current behavior. The following are explicitly
out of scope and MUST NOT be modified:

- Namespace-migration specs or archived specs
- Changelog entries
- Migration maps or compatibility logic
- Tests that intentionally detect stale refs
- Speckit internal commands (`.specify/`, `.opencode/command/`)
  which use `/speckit.*` names by design

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `review-council-spec`: All `/review-council` references updated
  to `/uf.review-council` in requirement text and scenarios.
- `cli-help-text`: `--divisor` flag help text updated from
  `/review-council` to `/uf.review-council`.

### Removed Capabilities

None.

## Impact

- **`cmd/unbound-force/main.go`**: One-line change in the
  `--divisor` flag description string.
- **`openspec/specs/review-council/spec.md`**: Six textual
  substitutions in requirement descriptions and scenario steps.
- No behavioral changes. No test changes. No API changes.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change only updates documentation strings and spec text. It
does not alter artifact-based communication between heroes. The
canonical `/uf.*` names make command references more precise,
improving self-description of artifacts.

### II. Composability First

**Assessment**: PASS

No dependencies are introduced or removed. Each hero remains
independently installable. The change is purely textual.

### III. Observable Quality

**Assessment**: PASS

No change to machine-parseable output or provenance metadata.
Help text and spec text are human-facing documentation.

### IV. Testability

**Assessment**: PASS

No behavioral changes. Existing tests that detect stale refs
are explicitly preserved (out of scope). No new test surface
introduced.
