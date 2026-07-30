<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Add confirmation gate to acceptance decision

All edits target `.opencode/agents/muti-mind-po.md`.
No [P] markers -- single file, sequential edits.

- [ ] 1.1 Insert an AskUserQuestion confirmation rule in
  the Acceptance Authority section, between the decision
  evaluation description (line 68) and the CLI command
  block (lines 69-72). The new text MUST instruct the
  agent to:
  (a) Present the target backlog item (`BI-NNN`), the
      proposed decision (`accept`/`reject`/`conditional`),
      and a rationale summary to the user.
  (b) Use the **AskUserQuestion tool** with options
      `["Confirm decision", "Abort"]`.
  (c) Only invoke `go run cmd/mutimind/main.go decide`
      if the user selects "Confirm decision".
  (d) If the user selects "Abort", inform them the
      decision was not recorded and do NOT invoke the
      CLI command.

- [ ] 1.2 Update the CLI command block preamble
  (line 69: "To generate these artifacts, you MUST use
  the Go CLI backend to ensure proper schema compliance:")
  to clarify that the CLI invocation is contingent on
  user confirmation. Suggested wording: "After the user
  confirms the decision, use the Go CLI backend to
  ensure proper schema compliance:"

**Checkpoint**: Read the modified file end-to-end.
Verify that the Acceptance Authority section contains
a mandatory AskUserQuestion gate before the `decide`
command block. Verify the gate presents item ID,
decision type, and rationale summary. Verify abort
handling is specified.

## 2. Verification

- [ ] 2.1 Verify no other unguarded CLI invocations exist
  in `muti-mind-po.md`. Search for `go run`, `bash`, and
  `mutimind` references. The `add` command at line 60
  should already have an Interactive Approval rule. The
  `generate-artifact` command at line 76 is a read-only
  operation (generates JSON output, does not mutate
  state) and does not require a confirmation gate.

- [ ] 2.2 Verify constitution alignment: confirm no
  artifact formats, hero interfaces, or machine-parseable
  outputs were modified. Only instruction text in the
  agent markdown file was changed.

- [ ] 2.3 Run `go test -race -count=1 ./...` to verify
  no test regressions. While no tests directly cover
  agent markdown files, this confirms no unintended
  side effects from the change.
<!-- scaffolded by uf vdev -->
