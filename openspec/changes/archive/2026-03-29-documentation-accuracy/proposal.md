## Why

A documentation audit revealed 9 HIGH, 8 MEDIUM, and
5 LOW severity issues across the meta repo's living
documentation. The most impactful problems:

- `README.md` still references graphthulhu as the
  knowledge layer (replaced by Dewey in Specs 014/015)
- `AGENTS.md` references the `unbound` binary in the
  hero table (renamed to `unbound-force` in Spec 013)
- 5 completed specs (012-016) still have `status: draft`
  in their frontmatter despite all tasks being done
  and PRs merged
- `AGENTS.md` says "Gaze is the only hero with a
  functional implementation" when all 5 heroes are
  implemented
- `AGENTS.md` project structure tree is missing specs
  014 and 015
- Spec counts and scaffold file counts in README are
  stale

These inaccuracies actively mislead both human
contributors and AI agents that consume AGENTS.md
as context.

## What Changes

Documentation-only corrections across 3 file categories:

1. **README.md**: Replace graphthulhu references with
   Dewey, fix spec count (10 → 16), fix scaffold file
   count (47 → 50), update Knowledge Graph section
2. **AGENTS.md**: Fix `unbound` → `unbound-force` in
   hero table, update hero implementation status, add
   missing specs to file tree, update phase
   descriptions, add Dewey to sibling repos, fix stale
   counts
3. **Spec frontmatter**: Update `status: draft` →
   `status: complete` for specs 012-016

## Capabilities

### New Capabilities
- None (documentation corrections only)

### Modified Capabilities
- `readme-accuracy`: README.md reflects current
  tooling (Dewey), correct counts, and accurate
  project description
- `agents-accuracy`: AGENTS.md reflects current hero
  status, correct binary names, complete project
  structure, and updated descriptions
- `spec-status-accuracy`: All completed spec
  frontmatter reflects actual implementation status

### Removed Capabilities
- None

## Impact

- **Files modified**: `README.md`, `AGENTS.md`,
  `specs/012-*/spec.md`, `specs/013-*/spec.md`,
  `specs/014-*/spec.md`, `specs/015-*/spec.md`,
  `specs/016-*/spec.md`
- **No code changes**: Zero Go files modified
- **No test changes**: No behavioral changes to verify
- **AI agent context**: AGENTS.md is consumed by all
  OpenCode agents as project context. Fixing
  inaccuracies improves agent decision quality.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

No changes to artifact-based communication or
inter-hero interfaces. Documentation-only change.

### II. Composability First

**Assessment**: N/A

No changes to hero installability, extension points,
or dependencies. Documentation-only change.

### III. Observable Quality

**Assessment**: PASS

Correcting documentation improves the accuracy of
provenance metadata (spec status fields) and ensures
living documentation is a reliable source of truth
about the system's actual state.

### IV. Testability

**Assessment**: N/A

No testable code changes. Spec frontmatter status
fields are not validated by automated tests.
