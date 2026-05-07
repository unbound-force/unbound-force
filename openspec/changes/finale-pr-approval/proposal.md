## Why

`/finale` Step 3 shows the proposed commit message and asks
"Approve, edit, or provide your own?" before committing. Step 5
does not do the same for PR creation — it generates a PR title
and body from commit history and immediately runs `gh pr create`
without showing the user the proposed content.

The PR description is what reviewers see first. Silently
generating it removes the user's opportunity to add context,
link issues, explain trade-offs, or correct inaccuracies. This
is inconsistent with the commit message workflow and contrary
to the command's own guardrail philosophy (warn before
destructive actions, require approval for user-facing content).

Reported in issue #156.

## What Changes

Add a user confirmation step to `/finale` Step 5 between PR
content generation and `gh pr create`. The confirmation mirrors
Step 3's existing commit message approval pattern: show the
proposed title and body, then ask the user to approve, edit,
or provide their own.

## Capabilities

### New Capabilities
- `pr-content-approval`: User confirmation of PR title and
  body before creation, matching the existing commit message
  approval pattern in Step 3

### Modified Capabilities
- `/finale` Step 5: PR creation now pauses to show proposed
  title and body for user review before calling `gh pr create`

### Removed Capabilities
- None

## Impact

- `.opencode/commands/finale.md`: Add confirmation substep
  to Step 5 between body generation and `gh pr create`
- `internal/scaffold/assets/opencode/commands/finale.md`:
  Sync scaffold copy with live command
- `CHANGELOG.md`: Add change entry
- Guardrails section: Add "NEVER create a PR without user
  approval of the title and body"
- No Go code changes
- No behavioral change when a PR already exists (skip path)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The PR title and body are artifacts that communicate intent
to reviewers. Giving the user control over these artifacts
before they are published improves the quality of
artifact-based communication between contributors and
reviewers.

### II. Composability First

**Assessment**: N/A

This change modifies a single slash command's Markdown
instructions. No hero dependencies or standalone
functionality is affected.

### III. Observable Quality

**Assessment**: PASS

User-approved PR descriptions are higher quality than
auto-generated ones. The confirmation step ensures the
user has reviewed what will be published, improving the
signal quality of PRs as observable artifacts.

### IV. Testability

**Assessment**: N/A

This is a Markdown-only change to agent instructions.
No executable code is added or modified. The scaffold
drift test (`TestEmbeddedAssets_MatchSource`) validates
that the live and scaffold copies remain in sync.
