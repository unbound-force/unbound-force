## Why

`/review-council` produces a multi-agent review verdict locally
but has no mechanism to post findings to a GitHub PR. The
council's findings — aggregated from up to 5 Divisor personas —
exist only in the terminal session. PR authors and collaborators
cannot see the council's verdict without being present at the
terminal.

This is Phase 3 of #177 (Enrich /review-council with /review-pr
capabilities). Phase 1 (#296, review-context skill migration) is
complete. This change adds optional GitHub review posting so the
council's consolidated verdict can be shared on the PR.

## What Changes

Add an optional final step to `/review-council` Code Review Mode
that detects an open PR for the current branch, offers to post
the council's consolidated findings as a GitHub review, and
requires human confirmation before posting.

Key behaviors:
- PR detection via `gh pr view` or explicit PR number argument
- Council verdict mapped to GitHub API event types (APPROVE,
  REQUEST_CHANGES, COMMENT)
- Multi-persona findings aggregated into a single review body
  with per-persona sections
- Inline comments for file-specific findings (capped at 15)
- Pre-posting checks: duplicate detection, stale review
  warnings, CODEOWNER warnings
- Human confirmation required before posting
- Protocol 2 (Issue Linking) from review-context skill unlocked
  when a PR exists (previously skipped due to no PR body)
- No behavioral change when no PR exists (local-only mode
  preserved)

## Capabilities

### New Capabilities
- `GitHub review posting`: Post consolidated council verdict
  as a GitHub PR review with per-persona sections
- `PR detection`: Detect open PR for current branch or accept
  PR number as argument
- `Protocol 2 unlock`: Run review-context Issue Linking when
  PR context is available
- `Pre-posting checks`: Duplicate detection, stale review
  dismissal warnings, CODEOWNER requirement warnings

### Modified Capabilities
- `/review-council Code Review Mode`: Extended with optional
  Step 7 (GitHub posting) after the existing Step 6 (Final
  Report). Step 6 report enhanced with posting offer.
- `review-context skill consumption`: Protocol 2 now runs
  conditionally when a PR is detected

### Removed Capabilities
- None

## Impact

- **Files**: `.opencode/commands/review-council.md`,
  `internal/scaffold/assets/opencode/commands/review-council.md`
  (scaffold sync)
- **Skills**: `review-context` skill — Protocol 2 consumption
  changes from "always skip" to "skip when no PR, run when PR
  exists"
- **Dependencies**: Runtime dependency on `gh` CLI for posting
  (graceful degradation when unavailable)
- **User workflow**: `/review-council` gains optional post-PR
  capability while preserving its pre-PR local-first identity
- **No code changes**: This change modifies only command
  definition files (Markdown), not Go source code

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.2.0).

### I. Autonomous Collaboration

**Assessment**: PASS

The council's findings are posted as a GitHub review — a
self-describing artifact with clear provenance (per-persona
attribution, council verdict, AI-assisted label). The review
is consumable by any collaborator without requiring the
producing agent to be present. The posting step is fully
asynchronous — it does not require synchronous interaction
with any other hero.

### II. Composability First

**Assessment**: PASS

The posting step is opt-in: when no PR exists or `gh` is
unavailable, `/review-council` continues to function exactly
as before (local-only). The `gh` CLI is a soft dependency with
graceful degradation, not a hard prerequisite. The command
remains independently usable without GitHub integration.

### III. Observable Quality

**Assessment**: PASS

The posted review includes provenance metadata: which Divisor
personas contributed, the council verdict, and an AI-assisted
disclosure label. The review body is structured with
per-persona sections, making findings machine-parseable by
downstream tooling. Inline comments include severity levels
and concrete suggestions.

### IV. Testability

**Assessment**: PASS

The posting logic follows the same JSON-to-tempfile-to-`gh api`
pattern already proven in `/review-pr`. Each sub-step (PR
detection, review state fetching, finding aggregation, verdict
mapping, posting) is independently verifiable through its
observable outputs. The graceful degradation paths (`gh`
unavailable, API errors) are testable via their error messages.

### V. Security by Default

**Assessment**: PASS

The JSON payload is written to a temporary file rather than
interpolated into shell arguments, preventing shell injection.
Human confirmation is required before posting, preventing
unintended review submissions. The `gh` CLI handles
authentication and token scoping — no secrets are managed by
the command itself. The 15-comment cap and body size limits
prevent resource exhaustion.
