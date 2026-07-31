## Why

The `/triage-issue` command (`.opencode/commands/triage-issue.md`)
applies label mutations to GitHub issues without user confirmation,
except for the `duplicate` label which is correctly gated behind
an `AskUserQuestion` prompt. This inconsistency was identified in
the issue #346 audit as a T1 weakness: irreversible external
actions (GitHub label mutations) without mandatory user
confirmation.

Line 243 of the command explicitly documents this asymmetry:
"Labels are applied automatically without user confirmation."
This violates the org policy established by the #346 audit that
ALL label mutations require `AskUserQuestion`.

Fixes: https://github.com/unbound-force/unbound-force/issues/354

## What Changes

1. **Add `AskUserQuestion` gate before `gh label create`**: When
   a label does not exist in the repository and needs to be
   created, prompt the user for confirmation before creating it.

2. **Add `AskUserQuestion` gate before `gh issue edit --add-label`**:
   When applying a label to the issue, prompt the user for
   confirmation before applying it.

3. **Update documentation**: Remove or update the "applied
   automatically without user confirmation" language at line 243
   to reflect the new gated behavior.

4. **Preserve existing duplicate gate**: The existing
   `AskUserQuestion` gate for the `duplicate` label (lines
   273-276) is already correct and serves as the pattern to
   generalize.

## Capabilities

### New Capabilities
- `label-creation-gate`: User confirmation required before
  creating new labels in the repository via `gh label create`
- `label-application-gate`: User confirmation required before
  applying any label to an issue via `gh issue edit --add-label`

### Modified Capabilities
- `triage-issue-label-flow`: Label application now requires user
  confirmation for all label categories, not just `duplicate`

### Removed Capabilities
- `auto-label-application`: Automatic label application without
  user confirmation is removed for all label categories

## Impact

- **Files**: `.opencode/commands/triage-issue.md` (Phase 4.2,
  lines 241-276) and its scaffold asset at
  `internal/scaffold/assets/opencode/commands/triage-issue.md`
- **User experience**: Two additional confirmation prompts per
  triage invocation (label creation if needed, label application).
  The `duplicate` label retains its existing specialized prompt
  with close-semantics warning.
- **Sibling issue**: Issue #352 tracks the same gap in
  `review-council.md` — this change addresses only `triage-issue`

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.2.0).

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies an interactive command's confirmation flow.
It does not affect artifact-based inter-hero communication or
self-describing outputs.

### II. Composability First

**Assessment**: N/A

This change is internal to the `/triage-issue` command. It does
not introduce dependencies between heroes or affect standalone
functionality.

### III. Observable Quality

**Assessment**: PASS

The triage artifact (`issue-triage/*.json`) already records
`actions_taken.labels_applied` and `label_creation_failed`.
The change does not alter artifact structure — user confirmation
or skip decisions are captured in the existing fields.

### IV. Testability

**Assessment**: PASS

No new Go code paths are introduced. However, the existing
scaffold drift detection test (`TestEmbeddedAssets_MatchSource`)
must pass after the scaffold asset is synchronized. The scaffold
asset at `internal/scaffold/assets/opencode/commands/triage-issue.md`
must be updated in lockstep with the command file.

### V. Security by Default

**Assessment**: PASS

This change strengthens security posture by adding confirmation
gates before irreversible external actions (GitHub label
mutations). It aligns with the input validation and least
privilege principles — ensuring the agent does not perform
external mutations without explicit human authorization.
