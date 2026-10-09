## Why

`/uf.finale` has two reliability gaps that undermine its role as the
trusted end-of-branch workflow:

1. **Invisible confirmation gates** (GitHub issue #542): The mandatory
   human-confirmation gates for commit messages (Step 3) and PR
   title/body (Step 5f) use the `question` tool to present approval
   options, but the `question` tool renders only the short question
   string and short option labels. The multi-line proposed text
   (commit message, PR body) is shown as a blockquote in the
   instruction text, but after context compression or across
   sub-task boundaries, the agent may invoke the `question` tool
   without the user having seen the actual content being approved.
   The user clicks "Approve" on text they cannot see.

2. **Skipped post-CI steps** (GitHub issue #572): After Step 6
   (Watch CI Checks) succeeds, Steps 7 (Return to Main) and 8
   (Summary) are silently skipped when the session has undergone
   context compression or crossed a sub-task boundary. The CI
   watch step has a natural "report to user" moment that breaks
   momentum through the remaining checklist. The agent reports
   CI results and stops, leaving the user on the feature branch
   with no summary.

Both issues erode trust in `/uf.finale` as a reliable, autonomous
workflow — the user cannot be sure what they approved or that all
steps completed.

## What Changes

- **Confirmation-gate visibility directive**: Add an explicit
  directive to `/uf.finale` Steps 3 and 5f requiring the agent
  to print the full proposed text (commit message, PR title/body)
  as plain assistant output BEFORE invoking the `question` tool.
  This ensures the text appears in the transcript regardless of
  context compression. Audit `/uf.review-council` Step 7f for
  the same pattern and apply the same directive there.

- **Post-CI checkpoint directive**: Add an explicit checkpoint
  instruction after the Step 6 CI pass/fail block in `/uf.finale`
  directing the agent to: (a) mark Step 6 complete in the
  execution checklist, (b) immediately proceed to Step 7 WITHOUT
  surfacing CI results to the user until all remaining steps
  (7 and 8) are complete. This prevents the natural "report CI
  results" moment from breaking the workflow chain.

## Capabilities

### New Capabilities

- `confirmation-gate-visibility`: Directive requiring agents to
  print full proposed text as plain output before invoking the
  `question` tool for approval gates, ensuring the user can see
  what they are approving regardless of context state.

- `post-ci-momentum-checkpoint`: Directive requiring agents to
  continue through Steps 7-8 without interruption after Step 6
  CI completion, deferring all user-facing output until the
  summary is ready.

### Modified Capabilities

- `/uf.finale`: Steps 3, 5f gain the confirmation-gate visibility
  directive. Step 6 gains the post-CI momentum checkpoint.
- `/uf.review-council`: Step 7f gains the confirmation-gate
  visibility directive (audit finding).

### Removed Capabilities

None.

## Impact

- **Files modified**:
  - `.opencode/commands/uf.finale.md` — Steps 3, 5f, 6
  - `.opencode/commands/uf.review-council.md` — Step 7f
- **Behavior**: Agents will produce more verbose output (full
  text before question tool) but the user experience improves
  because they can actually see what they are approving. The
  post-CI checkpoint ensures all 8 steps complete reliably.
- **Risk**: Low. These are directive additions to existing
  command templates, not structural changes. No new tools,
  no new data flows.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change strengthens artifact-based communication by ensuring
that confirmation gates produce visible, transcript-resident
output. The execution checklist already serves as the
self-describing state artifact; this change makes the gate
content equally observable.

### II. Composability First

**Assessment**: PASS

No new dependencies are introduced. The changes are purely
directive additions to existing command templates. Each command
remains independently installable and usable.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observability. Confirmation gate
content becomes visible in the transcript (machine-parseable
output with provenance). The post-CI checkpoint ensures the
execution checklist accurately reflects completed steps,
improving the machine-parseable state of the workflow.

### IV. Testability

**Assessment**: PASS

Command templates are testable via scripted agent sessions.
The directives can be verified by checking that (a) proposed
text appears as plain output before `question` tool invocation,
and (b) Steps 7-8 execute after Step 6 without user-facing
interruption. No external services required.
