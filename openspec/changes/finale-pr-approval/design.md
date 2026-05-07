## Context

`/finale` Step 3 already implements a user confirmation
pattern for commit messages: generate content, display it,
ask "Approve, edit, or provide your own?", then proceed with
the approved version. Step 5 generates PR title and body but
creates the PR immediately without confirmation.

The existing Step 3 pattern is the design template for this
change. The PR approval step mirrors it exactly, maintaining
consistency in the command's interaction model.

## Goals / Non-Goals

### Goals
- Add user confirmation of PR title and body before creation
- Mirror the existing Step 3 commit message approval pattern
- Add a guardrail preventing PR creation without approval
- Keep the interaction simple: show, ask, act

### Non-Goals
- Adding PR template support or auto-population from issue
  templates
- Changing the PR body generation logic (what is generated)
- Adding draft PR support or PR label selection
- Modifying the existing PR case (when PR already exists)

## Decisions

### Mirror Step 3's confirmation pattern exactly

The confirmation prompt MUST use the same structure as
Step 3c:

```
**Proposed PR:**

  Title: <title>

  Body:
  <body>

Approve, edit, or provide your own?
```

Rationale: Consistency. The user already knows this
interaction model from commit messages. Using the same
"Approve, edit, or provide your own?" prompt reduces
cognitive load.

### Insert between generation and creation

The confirmation step goes between body generation
(current step c) and `gh pr create` (current step d).
This follows the principle of Autonomous Collaboration:
user-approved artifacts are higher quality communication
than auto-generated ones.

### Add guardrail to the Guardrails section

Add "NEVER create a PR without user approval of the title
and body" to match the existing "NEVER commit without user
approval of the message" guardrail. These are symmetric
operations and should have symmetric protections.

### No change to PR-exists skip path

When a PR already exists (Step 5's first branch), the
command skips creation entirely. This path is unchanged —
there is nothing to approve when no PR is being created.

## Risks / Trade-offs

### Additional user interaction

Adding a confirmation step means one more prompt before
the PR is created. This is intentional friction — the
same trade-off already accepted for commit messages in
Step 3. The user retains control over what is published
in their name.

### Fork detection interaction ordering

PR #180 adds fork detection as Step 5a. The confirmation
step occurs after fork detection and content generation,
before `gh pr create`. The ordering is: detect fork →
generate title → generate body → **confirm with user** →
create PR. No conflict.
