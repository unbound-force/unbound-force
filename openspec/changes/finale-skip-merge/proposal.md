## Why

The `/finale` command currently merges the PR
automatically after CI passes. This prevents human
reviewers from examining the PR before it lands on
main. In a team workflow, PRs should remain open for
review after CI passes — the author returns to main
and waits for approval before merging.

## What Changes

Remove step 7 (Merge PR) from the `/finale` command.
The command still commits, pushes, creates the PR,
watches CI checks, and returns to main — but leaves
the PR open for reviewers.

### Before

```
commit → push → create PR → watch CI → merge → main
```

### After

```
commit → push → create PR → watch CI → main
```

The PR remains open. The user merges via GitHub UI
or `gh pr merge` after reviewers approve.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `/finale`: No longer merges the PR. Stops after CI
  checks pass and returns to main. Summary reports PR
  as "ready for review" instead of "merged via rebase".

### Removed Capabilities
- Automatic PR merge from `/finale`. Users merge
  manually after review.

## Impact

- `.opencode/command/finale.md`: Remove step 7 (Merge
  PR), update step 9 summary template, update
  guardrails.
- `internal/scaffold/assets/opencode/command/finale.md`:
  Same changes (scaffold copy).
- No Go code changes. Markdown-only.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

No inter-hero communication affected. The change
improves collaboration by enabling human review
before merge.

### II. Composability First

**Assessment**: N/A

No dependency changes.

### III. Observable Quality

**Assessment**: PASS

PR review is a quality gate. Keeping PRs open for
review strengthens the observable quality principle.

### IV. Testability

**Assessment**: N/A

No testable code changes — Markdown command definition
only.
