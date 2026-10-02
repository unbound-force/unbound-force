<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency on
  prior tasks in the group, (c) can safely execute without
  ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Confirmation-Gate Visibility Directive — /uf.finale

- [x] 1.1 Read `.opencode/commands/uf.finale.md` and locate
      Step 3 (Generate and Confirm Commit Message), the
      mandatory gate block around lines 211-258.
- [x] 1.2 Add a visibility directive immediately before the
      `question` tool invocation in Step 3: instruct the agent
      to print the full proposed commit message (summary, body,
      and attribution footer) as plain assistant output BEFORE
      invoking the `question` tool.
- [x] 1.3 Locate Step 5f (Create or Find PR — human
      confirmation gate, around lines 453-469). Add the same
      visibility directive: instruct the agent to print the
      full PR title and body as plain assistant output BEFORE
      invoking the `question` tool.

## 2. Confirmation-Gate Visibility Directive — /uf.review-council

- [x] 2.1 Read `.opencode/commands/uf.review-council.md` and
      locate Step 7f (Verdict Mapping and Human Confirmation,
      around lines 681-742).
- [x] 2.2 Add a visibility directive immediately before the
      `question` tool invocation in Step 7f: instruct the agent
      to print the full verdict context (verdict type, review
      body, and all inline comments) as plain assistant output
      BEFORE invoking the `question` tool.

## 3. Post-CI Momentum Checkpoint — /uf.finale

- [x] 3.1 Read `.opencode/commands/uf.finale.md` and locate the
      checkpoint instruction at the end of Step 6 (around
      lines 946-947) and the beginning of Step 7 (line 949).
- [x] 3.2 Add an explicit momentum checkpoint directive after
      the Step 6 checkpoint: instruct the agent to (a) mark
      Step 6 complete in the execution checklist, (b) proceed
      immediately to Step 7 WITHOUT producing any user-facing
      output about CI results, and (c) defer all CI result
      reporting until Step 8 (Summary) is reached.

## 4. Verification

- [x] 4.1 Verify that the confirmation-gate visibility directive
      appears in all three locations: `/uf.finale` Step 3,
      `/uf.finale` Step 5f, and `/uf.review-council` Step 7f.
- [x] 4.2 Verify that the post-CI momentum checkpoint appears
      after Step 6 in `/uf.finale` and before Step 7.
- [x] 4.3 Verify constitution alignment: confirm that all
      changes are directive additions to existing command
      templates (no new dependencies, no structural changes)
      and that the four constitution principles (Autonomous
      Collaboration, Composability First, Observable Quality,
      Testability) remain satisfied.

<!-- spec-review: passed -->
<!-- code-review: passed -->
