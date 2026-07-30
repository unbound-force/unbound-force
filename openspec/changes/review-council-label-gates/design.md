## Context

The `/triage-issue` command in `.opencode/commands/triage-issue.md`
currently applies GitHub labels automatically without user
confirmation, with a single exception: the `duplicate` label
has an AskUserQuestion gate (lines 273-276) because it carries
implicit "close" semantics.

Per the #346 audit, all label mutations are irreversible
external actions (T1 weakness type) and require AskUserQuestion
confirmation. The existing `duplicate` gate pattern is the
correct model — it needs to be generalized to all labels.

The proposal (proposal.md) confirms this change aligns with
Constitution Principle V (Security by Default) by enforcing
a least-privilege confirmation boundary before GitHub state
modifications.

## Goals / Non-Goals

### Goals
- Add AskUserQuestion gate before `gh label create` for all
  labels (not just duplicate)
- Add AskUserQuestion gate before `gh issue edit --add-label`
  for all labels (not just duplicate)
- Preserve the existing duplicate-specific gate messaging
  (lines 273-276) since it provides additional context about
  close semantics
- Maintain the existing flow: re-run check -> label existence
  check -> create if needed -> apply label

### Non-Goals
- Changing the triage classification logic or panel verdicts
- Adding batch confirmation (e.g., "apply all N labels at
  once") — each label gets its own gate for maximum control
- Modifying the JSON artifact schema (`actions_taken`)
- Changing the `/review-council` command (despite the issue
  title referencing it, the actual label mutation code is in
  `/triage-issue`)

## Decisions

### D1: Gate placement — before each gh command

Place AskUserQuestion gates immediately before each `gh`
command that mutates GitHub state, not at a higher level.

**Rationale**: This matches the existing duplicate-label
pattern and ensures that even if the flow is restructured
in the future, each irreversible action retains its guard.
It also gives the user precise context about what will happen
(create vs. apply) at each step.

### D2: Separate gates for create and apply

Use separate AskUserQuestion prompts for label creation
(`gh label create`) and label application
(`gh issue edit --add-label`) rather than a single combined
prompt.

**Rationale**: These are distinct operations with different
failure modes and implications. Label creation affects the
entire repository; label application affects a single issue.
A user might want to create the label but not apply it (or
vice versa).

### D3: Preserve duplicate-specific messaging as addendum

Keep the existing duplicate-label gate (lines 273-276) as
additional messaging layered on top of the general gate. The
general apply gate fires for all labels including duplicate.
The duplicate-specific text adds context about close semantics.

**Rationale**: The duplicate label has unique implications
(signals the issue should be closed). Removing its specific
messaging would reduce user awareness. The general gate
handles confirmation; the duplicate text handles education.

### D4: Skip semantics on user decline

When a user selects "No -- skip" at any gate, the command
MUST skip that specific operation, record the skip in
`actions_taken`, and continue with remaining triage actions
(comment posting, child issue creation).

**Rationale**: Declining a label should not abort the entire
triage. The user may want the triage comment posted even if
they disagree with the label classification.

## Risks / Trade-offs

### R1: Increased prompt count per triage

**Risk**: Each triage now requires 1-2 additional user
confirmations (label create + label apply). For common
single-label triage, this adds one or two prompts.

**Mitigation**: Acceptable per org security policy. The T1
weakness (irreversible action without confirmation) is a
higher-priority concern than UX friction. Future work could
add a batch confirmation mode if friction becomes a problem.

### R2: Skipped labels not reflected in artifact

**Risk**: If a user declines label application, the JSON
artifact's `labels_applied` array will be empty or partial,
which could cause confusion on re-triage.

**Mitigation**: D4 requires recording skips in
`actions_taken`. The existing re-run check (line 257) already
detects previously applied labels, so partial application
is handled.
