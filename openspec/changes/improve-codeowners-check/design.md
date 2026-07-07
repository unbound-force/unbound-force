## Context

The CODEOWNERS existence check in `/review-pr` (Step 11a,
lines 802-820) and `/review-council` (Step 7c, lines 474-484)
uses `2>/dev/null` to suppress stderr from `gh api` calls.
This masks all errors — not just 404 — making non-404 failures
(network, 500, 429 rate limits) invisible. Additionally, only
two of the three valid GitHub CODEOWNERS paths are checked.

Both commands use identical logic. The scaffold source copies
under `internal/scaffold/assets/opencode/commands/` mirror the
active copies exactly. All four files must be updated together.

The proposal's constitution alignment confirmed this change
improves Observable Quality (Principle III) by surfacing
previously-hidden errors.

## Goals / Non-Goals

### Goals
- Distinguish 404 (no file) from non-404 errors in `gh api`
  responses
- Check all three valid GitHub CODEOWNERS locations:
  `CODEOWNERS`, `.github/CODEOWNERS`, `docs/CODEOWNERS`
- Display a warning when a non-404 error occurs so the user
  knows the check was inconclusive
- Keep `/review-pr` and `/review-council` CODEOWNERS logic
  identical
- Sync scaffold source copies with active command copies

### Non-Goals
- Changing the CODEOWNER warning message text itself
- Making the CODEOWNERS check a blocking gate (it remains
  advisory)
- Parsing CODEOWNERS file contents to validate entries
- Adding retry logic for transient API failures

## Decisions

### D1: Error handling approach

The agent commands are markdown instructions interpreted by
an LLM, not executable shell scripts. The `gh api` bash
snippets serve as illustrative pseudocode for the agent.

**Decision**: Replace the `2>/dev/null` pattern with
explicit error-handling instructions. The updated text will:

1. Instruct the agent to call `gh api` for each of the three
   paths.
2. Treat 404 responses as "file not found" (silent, expected).
3. Treat any non-404 error as an inconclusive check and
   display a warning to the user.
4. Consider CODEOWNERS found if any of the three paths
   returns successfully.

The bash snippet will use `--silent` (suppress progress) with
response code inspection rather than blanket stderr
suppression.

### D2: Path ordering

**Decision**: Check paths in order of convention prevalence:
`.github/CODEOWNERS` (most common), `CODEOWNERS` (root),
`docs/CODEOWNERS` (least common). Short-circuit on first
success — if found at the first path, skip remaining checks.

### D3: Warning message for inconclusive checks

**Decision**: Add a new warning message for non-404 errors:

```
Note: CODEOWNERS check was inconclusive (API error).
Could not determine if this repo uses CODEOWNERS.
```

This is distinct from the existing CODEOWNER review warning
and uses "Note:" severity (informational, not blocking).

### D4: Four-file update strategy

**Decision**: Update all four files in a single change:
- `.opencode/commands/review-pr.md` (active copy)
- `.opencode/commands/review-council.md` (active copy)
- `internal/scaffold/assets/opencode/commands/review-pr.md`
  (scaffold source)
- `internal/scaffold/assets/opencode/commands/review-council.md`
  (scaffold source)

The scaffold drift detection tests will verify consistency
after the change.

## Risks / Trade-offs

### R1: LLM interpretation variance

The CODEOWNERS check is agent instructions, not compiled code.
Different LLMs may interpret error-handling instructions with
slight variation. Mitigation: keep the instructions concrete
with explicit bash pseudocode and clear conditional logic.

### R2: API call count increase

Adding `docs/CODEOWNERS` increases the maximum `gh api` calls
from 2 to 3 per review. Mitigation: short-circuit on first
success. Most repos use `.github/CODEOWNERS`, so the typical
case remains 1 call.

### R3: Low severity change

This is an advisory check improvement. The risk of regression
is minimal since the CODEOWNERS check does not affect the
review verdict.
