## Why

The CODEOWNERS existence check in `/review-pr` (Step 11a) and
`/review-council` (Step 7c) has two issues identified in
[#324](https://github.com/unbound-force/unbound-force/issues/324):

1. **Stderr suppression**: `2>/dev/null` on `gh api` calls masks
   all errors, not just 404. Non-404 errors (network failures,
   500 server errors, 429 rate limits) are silently swallowed,
   causing the CODEOWNER warning to be skipped without the user
   knowing the check failed.
2. **Missing path**: Only `CODEOWNERS` and `.github/CODEOWNERS`
   are checked. GitHub also supports `docs/CODEOWNERS` as a
   valid location.

Both commands share this pattern — `/review-council` inherited
it from `/review-pr` during the Step 7 transplant in #314. The
fix should be applied to both commands together to maintain
consistency.

## What Changes

Update the CODEOWNERS check logic in both `/review-pr` and
`/review-council` agent command files, plus their scaffold
source copies:

1. Replace `2>/dev/null` with error-aware handling that
   distinguishes 404 (no file) from other errors (network,
   500, 429).
2. Add `docs/CODEOWNERS` as a third valid location.
3. Instruct agents to log a warning when `gh api` returns a
   non-404 error so the user knows the check was inconclusive.
4. Keep identical logic in both commands.

## Capabilities

### New Capabilities
- `codeowners-error-visibility`: Non-404 errors from `gh api`
  produce a visible warning instead of being silently
  suppressed.
- `docs-codeowners-path`: `docs/CODEOWNERS` is checked as a
  valid location alongside the existing two paths.

### Modified Capabilities
- `codeowners-check`: Updated error handling and path coverage
  in both `/review-pr` and `/review-council`.

### Removed Capabilities
- None.

## Impact

- **Files**: `.opencode/commands/review-pr.md`,
  `.opencode/commands/review-council.md`, and their scaffold
  source copies under `internal/scaffold/assets/opencode/
  commands/`.
- **Behavior**: Agents running `/review-pr` or `/review-council`
  will now see a warning when CODEOWNERS checks fail for
  non-404 reasons, rather than silently skipping. The existing
  CODEOWNER warning message itself is unchanged.
- **Risk**: Low. This is an advisory check, not a security gate.
  The change improves observability without altering the review
  verdict logic.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change modifies agent command instructions (markdown files).
It does not affect artifact-based communication or inter-hero
collaboration patterns.

### II. Composability First

**Assessment**: PASS

Both commands remain independently usable. The CODEOWNERS check
remains self-contained within each command file — no new
dependencies are introduced.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observability by surfacing errors
that were previously suppressed. Agents will now report
inconclusive checks rather than silently skipping them.

### IV. Testability

**Assessment**: N/A

The changed files are agent instruction markdown, not executable
code. The scaffold drift detection tests will verify that the
active command copies match their scaffold sources.
