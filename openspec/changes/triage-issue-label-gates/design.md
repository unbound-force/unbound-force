## Context

The `/triage-issue` command in `.opencode/commands/triage-issue.md`
performs GitHub label mutations in Phase 4.2. Currently, only the
`duplicate` label is gated behind an `AskUserQuestion` prompt
(lines 273-276). All other labels are applied automatically
without user confirmation (line 243). This was identified as a
T1 weakness in the issue #346 audit (tracked as issue #354).

The existing duplicate-label gate pattern is the correct model
to generalize. The change is scoped to a single Markdown command
file with no Go source code modifications.

## Goals / Non-Goals

### Goals
- Gate all label creation (`gh label create`) behind
  `AskUserQuestion` confirmation
- Gate all label application (`gh issue edit --add-label`)
  behind `AskUserQuestion` confirmation
- Preserve the existing duplicate-label gate with its
  specialized close-semantics warning
- Update documentation to reflect the new gated behavior

### Non-Goals
- Modifying `review-council.md` (tracked by issue #352)
- Changing the triage artifact schema or structure
- Altering the label mapping table or category resolution
- Adding new label categories

## Decisions

### D1: Two-stage confirmation flow

Label application uses a two-stage flow when a label does not
exist:

1. **Label creation gate**: Before `gh label create`, prompt
   with `["Yes -- create label", "No -- skip"]`. If skipped,
   skip label application entirely (cannot apply a non-existent
   label).
2. **Label application gate**: Before `gh issue edit --add-label`,
   prompt with `["Yes -- apply label", "No -- skip"]`.

When the label already exists, only stage 2 applies.

**Rationale**: Each mutation is independently irreversible. Label
creation adds a permanent repository-level resource. Label
application modifies issue state. Both deserve separate
confirmation.

### D2: Unified gate for all non-duplicate labels

All non-duplicate labels (`bug`, `enhancement`, `question`,
`design-discussion`, `needs-info`) use the same generic
confirmation prompt. The `duplicate` label retains its existing
specialized prompt that warns about close semantics.

**Rationale**: Non-duplicate labels have similar risk profiles.
A single pattern reduces command complexity. The duplicate label
is special because applying it carries an implicit "this issue
should be closed" signal.

### D3: Skip-on-decline behavior

When the user declines label creation or application:
- Record the skip in `actions_taken` (existing field)
- Continue to Phase 4.3 (comment composition) — do not abort
  the entire triage

**Rationale**: Label application is one action among several.
Declining a label should not prevent the user from posting a
triage comment or creating child issues.

### D4: Re-run check preserves existing behavior

The existing re-run check at line 257 ("If the target label is
already applied, skip label application and note 'label already
present'") remains unchanged. The new gates only fire when the
label would actually be created or applied.

## Risks / Trade-offs

### Additional user friction

Adding two confirmation prompts per triage invocation increases
the number of interactions. This is an accepted trade-off: the
org policy from issue #346 mandates user confirmation for all
irreversible external actions. The prompts are quick to dismiss
and protect against unintended label mutations.

### Prompt fatigue

Users triaging many issues may experience prompt fatigue. This
is mitigated by clear, concise prompt text and the ability to
skip. A future enhancement could add a batch-mode flag, but that
is out of scope for this change.
