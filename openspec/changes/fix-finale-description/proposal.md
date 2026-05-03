## Why

The `/unleash` command's Step 8 (Demo) describes
`/finale` as "commit, push, create PR, merge, and
return to main." This contradicts `/finale`'s own
guardrails which state: "NEVER merge the PR — /finale
creates PRs for review, not for immediate merge."

The inaccurate description sets a false expectation
that `/finale` will merge the PR, when it actually
stops after creating the PR and watching CI checks.

## What Changes

Fix two lines in `/unleash` (live + scaffold copy):
1. Line 605: remove "merge" from the `/finale`
   description
2. Line 632: change "merge and release" to "create PR
   and watch CI"

## Capabilities

### Modified Capabilities
- `/unleash` Step 8 Demo output: corrected `/finale`
  description to match its actual behavior

### New Capabilities
- None

### Removed Capabilities
- None

## Impact

### Files Affected

| Area | Changes |
|------|---------|
| `.opencode/command/unleash.md` | Fix 2 lines describing `/finale` |
| `internal/scaffold/assets/opencode/command/unleash.md` | Sync scaffold copy |

### External Dependencies
- None

## Constitution Alignment

### III. Observable Quality

**Assessment**: PASS — fixing inaccurate documentation
improves the accuracy of the system's self-description.
<!-- scaffolded by uf vdev -->
