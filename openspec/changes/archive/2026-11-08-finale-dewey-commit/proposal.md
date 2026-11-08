## Why

Dewey learnings (`.uf/dewey/learnings/*.md`) and compiled
articles (`.uf/dewey/compiled/*.md`) produced during a
change's workflow are directly related to the active
change and MUST be included in the feature PR. However,
`/uf.finale` has no explicit instruction to stage these
files. Agents may omit them, causing knowledge artifacts
to be lost or left uncommitted.

## What Changes

Add an explicit step to `/uf.finale` (Step 2: Check for
Changes to Commit) that stages Dewey learnings and
compiled articles before proceeding to commit. The step
is idempotent -- if no Dewey files exist, it is a no-op.

## Capabilities

### New Capabilities
- `dewey-staging-step`: Explicit `git add` of
  `.uf/dewey/learnings/` and `.uf/dewey/compiled/`
  directories during the finale staging phase.

### Modified Capabilities
- `uf.finale step 2`: Adds a sub-step after "Stage all
  changes" to ensure Dewey artifacts are included.

### Removed Capabilities
- None.

## Impact

- **File**: `internal/scaffold/assets/opencode/commands/uf.finale.md`
- **Behavior**: Agents running `/uf.finale` will always
  stage Dewey knowledge artifacts alongside the rest of
  the working tree. No change to user-facing gates or
  approval flows.
- **Risk**: Minimal. The `git add` is idempotent and
  targets only `.uf/dewey/` subdirectories.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change ensures knowledge artifacts produced during
a workflow are preserved in the commit. It strengthens
artifact-based communication by preventing knowledge
loss between sessions.

### II. Composability First

**Assessment**: N/A

No new dependencies or cross-repo coupling introduced.
The change is scoped to a single scaffold template.

### III. Observable Quality

**Assessment**: PASS

Dewey learnings and compiled articles are
machine-parseable Markdown with provenance metadata.
Ensuring they are committed improves the observability
of the knowledge captured during each change.

### IV. Testability

**Assessment**: N/A

The change is to a command template (Markdown), not to
executable code. Existing scaffold drift-detection tests
will catch file-count changes. No new testable component
is introduced.

## Superseded

Superseded by `feedback-loop-scope` (2026-11-08): Dewey learnings no longer
ride the feature PR. Moved to the archive without merging its specs,
so this change's requirements never reach `openspec/specs`.
