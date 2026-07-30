## Why

The `/triage-issue` command applies GitHub labels automatically
without user confirmation, except for the `duplicate` label
which has an AskUserQuestion gate. Per org policy established
by the issue #346 audit, ALL label mutations (irreversible
external actions) require AskUserQuestion confirmation.

This inconsistency creates a T1 weakness: irreversible GitHub
label mutations execute without mandatory human confirmation.
Labels affect issue triage workflows, automation triggers,
and project board state. Applying a wrong label silently is
harder to detect and revert than preventing it upfront.

Fixes: https://github.com/unbound-force/unbound-force/issues/352

## What Changes

Generalize the existing `duplicate` label AskUserQuestion gate
pattern in `.opencode/commands/triage-issue.md` to cover ALL
label mutations:

1. Before `gh label create`: add AskUserQuestion with options
   `["Yes -- create label", "No -- skip"]`
2. Before `gh issue edit --add-label`: add AskUserQuestion with
   options `["Yes -- apply label", "No -- skip"]`

The existing `duplicate` label gate (lines 273-276) remains
as-is since it already follows the correct pattern. The change
extends this same pattern to every label operation.

## Capabilities

### New Capabilities
- `label-creation-gate`: AskUserQuestion confirmation before
  creating new labels in the repository via `gh label create`
- `label-application-gate`: AskUserQuestion confirmation before
  applying any label to an issue via `gh issue edit --add-label`

### Modified Capabilities
- `triage-issue-label-application`: Changes from automatic
  application (with duplicate-only gate) to confirmation-gated
  application for all labels

### Removed Capabilities
- None

## Impact

- **File**: `.opencode/commands/triage-issue.md` (section 4.2
  Label Application, lines ~243-276)
- **Behavior**: Users will now be prompted before any label is
  created or applied, not just for `duplicate`
- **UX trade-off**: Adds one confirmation step per label
  operation. For single-label triage this is one extra prompt.
  For multi-label scenarios the prompts are sequential. This is
  acceptable because label mutations are irreversible external
  actions and the security benefit outweighs the friction.
- **No code changes**: This is a command specification change
  only (Markdown agent instructions)

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.2.0).

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent instruction text in a single command
file. It does not affect artifact-based communication between
heroes or introduce runtime coupling.

### II. Composability First

**Assessment**: N/A

No new dependencies are introduced. The change is confined to
the `/triage-issue` command specification within this meta
repository.

### III. Observable Quality

**Assessment**: N/A

No machine-parseable output formats are affected. The JSON
triage artifact structure remains unchanged.

### IV. Testability

**Assessment**: N/A

The change is to agent instruction text, not executable code.
The triage-issue command's testability characteristics are
unchanged.

### V. Security by Default

**Assessment**: PASS

This change directly advances Security by Default. It closes
a T1 weakness (irreversible external action without mandatory
confirmation) by extending the existing AskUserQuestion gate
pattern to all label mutations, enforcing a least-privilege
confirmation boundary before any GitHub state modification.
