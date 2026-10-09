## Why

When a developer starts an OpenSpec change (or Speckit spec)
from a GitHub issue, there is no mechanism to carry the
originating issue number through the change artifacts into
the PR body. The resulting PR does not contain a `Closes #N`
or `Fixes #N` reference, so merging the PR does not
automatically close the originating issue. Developers must
manually remember to add the reference, and often forget.

This change threads the originating issue number through the
change's machine-readable metadata and updates `/uf.finale`
(PR body generation) to emit the `Closes #N` trailer
automatically.

Addresses #554. Precedes and accompanies #563 but is a
SEPARATE, focused PR.

## What Changes

1. Add an `originating_issue` field to the OpenSpec change
   frontmatter (`.openspec.yaml`) and the Speckit spec
   metadata. The field is optional and holds a GitHub issue
   number (integer).

2. Update `openspec new change` (and the Speckit equivalent)
   to accept an `--issue <N>` flag that populates the field.
   When omitted, the field is absent (backward compatible).

3. Update `/uf.finale` Step 5d (PR body generation) to read
   the `originating_issue` from the change's metadata and,
   when present, append a `Closes #<N>` line to the PR body
   after the Summary section.

4. Update the OpenSpec schema and Speckit templates to
   document the new field.

## Capabilities

### New Capabilities
- `originating-issue-metadata`: Optional machine-readable
  field in change artifacts that records the GitHub issue
  number that initiated the change.

### Modified Capabilities
- `openspec-new-change`: Accepts `--issue <N>` flag to
  populate `originating_issue` in `.openspec.yaml`.
- `uf-finale-pr-body`: Reads `originating_issue` from
  change metadata and emits `Closes #<N>` in the PR body
  when present.

### Removed Capabilities
- None.

## Impact

- `openspec/` schema and CLI (new flag, new field).
- `.specify/templates/` (Speckit spec template gains
  optional `originating_issue` frontmatter).
- `.opencode/commands/uf.finale.md` (Step 5d gains
  originating-issue lookup and `Closes #N` emission).
- No impact on existing changes that omit the field
  (fully backward compatible).

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The originating issue number is stored as a
machine-readable field in the change artifact
(`.openspec.yaml` or spec frontmatter). Any downstream
consumer (e.g. `/uf.finale`) reads it from the artifact
without requiring synchronous interaction with the
developer or another hero. The artifact remains
self-describing.

### II. Composability First

**Assessment**: PASS

The field is optional. Changes that omit it behave
exactly as before. No new mandatory dependencies are
introduced. Each hero (OpenSpec, Speckit, finale) can
function independently when the field is absent.

### III. Observable Quality

**Assessment**: PASS

The `originating_issue` field is machine-parseable
(integer in YAML frontmatter). Its presence or absence
is deterministic and verifiable. The `Closes #N` trailer
in the PR body is a standard GitHub convention that
produces observable, auditable traceability.

### IV. Testability

**Assessment**: PASS

The change is testable in isolation:
- `openspec new change --issue 554` can be verified to
  produce `.openspec.yaml` with `originating_issue: 554`.
- `/uf.finale` PR body generation can be tested with a
  mock change directory containing the field, asserting
  that `Closes #554` appears in the output.
- Absence of the field produces no `Closes` line
  (backward-compatible test case).
