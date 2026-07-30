## Context

The `/finale` command executes `git push` in Step 4
(lines 156-165) without an AskUserQuestion confirmation
gate. This is the same T1 weakness pattern identified in
issue #346 and fixed for other commands by the
`askuser-tool-confirmations` change.

The proposal confirms constitution alignment: this change
modifies only agent prompt instructions (markdown), not
hero interfaces, artifact formats, or Go source logic.

## Goals / Non-Goals

### Goals
- Add an AskUserQuestion gate immediately before the
  `git push` execution in Step 4 of `/finale`
- Follow the established conventions from D1, D2, and D5
  of the `askuser-tool-confirmations` design (structured
  selection, action-descriptive labels, bold formatting)
- Keep the scaffold asset byte-identical to the command
  file

### Non-Goals
- Adding confirmation gates to other `/finale` steps
  (commit message approval in Step 3 and PR approval in
  Step 5 already have user confirmation)
- Modifying any Go source code beyond the scaffold asset
- Changing the push behavior itself (upstream detection,
  error handling)

## Decisions

### D1: Two-option confirmation before push

The AskUserQuestion gate presents two options:
- "Push to remote" -- proceed with `git push`
- "Abort -- do not push" -- stop and let the user handle
  it manually

**Rationale**: Push is a binary decision with no
meaningful middle ground. The user has already approved
the commit message in Step 3, so the only question is
whether to push it now. This matches the simplicity of
the action. More complex option sets (like edit, defer)
would add unnecessary friction for a straightforward
irreversible action.

### D2: Gate placement inside Step 4

The AskUserQuestion gate is placed after the upstream
detection check but before the actual `git push`
command. The confirmation message includes the target
(branch name and whether upstream is being set).

**Rationale**: The user needs to know where the push
will go before confirming. Placing the gate after
upstream detection provides that context.

### D3: Option labels include action context

Following the convention from the
`askuser-tool-confirmations` design (D2), option labels
describe consequences:
- "Push to remote" (not just "Yes")
- "Abort -- do not push" (not just "No")

**Rationale**: Action-descriptive labels reduce ambiguity
and help users understand the consequence of their
selection.

### D4: Scaffold asset updated in lockstep

The command file under `.opencode/commands/` has a
byte-identical copy under
`internal/scaffold/assets/opencode/commands/`. Both
must be updated together. The existing drift detection
test (`TestEmbeddedAssets_MatchSource`) enforces this.

**Rationale**: Required by the scaffold pattern.
Existing test infrastructure validates this
automatically.

## Risks / Trade-offs

### R1: Additional friction in the push step

The confirmation gate adds one interaction step to the
`/finale` workflow. Accepted trade-off: the safety
benefit of preventing accidental pushes outweighs the
minor friction increase, especially since `/finale`
already has multiple confirmation steps (commit message,
PR content).

### R2: No behavioral regression

The change is purely instructional (markdown edits). No
Go code logic changes, no API changes, no schema changes.
Risk of behavioral regression is minimal. The scaffold
drift test provides automated verification.
<!-- scaffolded by uf vdev -->
