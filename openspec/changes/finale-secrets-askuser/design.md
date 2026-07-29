## Context

The `/finale` command (`.opencode/commands/finale.md`) has a
secrets check gate in Step 2 (lines 63-80) that warns the user
about potential secret files before staging. The confirmation
is expressed as prose: "Ask for confirmation. If the user
declines, stop and let them handle it manually."

This T3 weakness (prose-only gate) means an agent can skip the
confirmation under context compression. The sibling change
`askuser-tool-confirmations` established the pattern for
converting free-text confirmations to AskUserQuestion tool
calls across `/review-pr`, `/address-feedback`, and
`/triage-issue`. This change applies the same pattern to
`/finale`.

## Goals / Non-Goals

### Goals
- Replace the prose-only secrets confirmation in `/finale`
  Step 2 with an explicit AskUserQuestion tool call
- Preserve the existing safety semantics (never stage secret
  files without user confirmation)
- Follow the conventions established by the
  `askuser-tool-confirmations` change (D1, D2, D5 decisions)
- Keep scaffold asset byte-identical to command file

### Non-Goals
- Adding new interaction points or changing `/finale` workflow
  logic
- Modifying any other confirmation gates in `/finale` (commit
  message approval, PR approval, CI failure handling)
- Modifying the AskUserQuestion tool itself
- Changing any Go source code beyond the scaffold asset copy

## Decisions

### D1: AskUserQuestion with action-descriptive options

Replace lines 79-80 with an explicit AskUserQuestion tool call
using two options:
- "Yes -- stage all files including flagged ones"
- "No -- stop and let me handle it manually"

If the user selects "No", the agent MUST stop immediately and
not run `git add .`.

**Rationale**: Follows the D2 convention from the sibling
change -- option labels describe what will happen, not just
bare yes/no. The structured selection is unambiguous and
cannot be skipped by an agent under context compression.

### D2: Preserve warning message format

The existing warning block (lines 70-77) that lists detected
secret files remains unchanged. Only the confirmation
mechanism (lines 79-80) is replaced.

**Rationale**: The warning text is informative and correctly
formatted. Changing it would expand scope beyond the T3 fix.

### D3: Scaffold asset updated in lockstep

The scaffold asset at
`internal/scaffold/assets/opencode/commands/finale.md` MUST
be updated with the identical change. The existing drift
detection test (`TestEmbeddedAssets_MatchSource`) enforces
byte-identical copies.

**Rationale**: Required by the scaffold pattern (D4 from the
sibling change). Existing test infrastructure validates this
automatically.

### D4: Bold formatting convention

The AskUserQuestion tool is referenced as
`**AskUserQuestion tool**` (bold, PascalCase) in the command
text, matching the convention in `opsx-propose.md`,
`opsx-apply.md`, and `opsx-archive.md` (D5 from the sibling
change).

**Rationale**: Consistency with established convention.

## Risks / Trade-offs

### R1: Minimal scope reduces risk

This change modifies exactly two lines in one command file
(plus its scaffold copy). The risk of behavioral regression
is minimal. The scaffold drift test provides automated
verification.

### R2: Single interaction point

Unlike the sibling change which converts 15 interaction
points across three commands, this change converts exactly
one interaction point. The implementation is straightforward
with no complex sequencing or multi-step interactions.
<!-- scaffolded by uf vdev -->
