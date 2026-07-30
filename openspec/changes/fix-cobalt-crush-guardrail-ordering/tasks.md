<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.

  NOTE: All tasks in Group 1 modify the same file
  (cobalt-crush.md), so none are parallel-eligible.
-->

## 1. Relocate Branch Safety Guardrails

Target: `internal/scaffold/assets/opencode/commands/cobalt-crush.md`

- [ ] 1.1 Add a `### Pre-conditions` subsection immediately
  after the `## Instructions` heading (line 40) and before
  `### When arguments are provided` (line 42). Insert the
  full content of the current `## Branch Safety Guardrails`
  section (lines 99-111) into this new subsection, preserving
  all four guardrail rules verbatim.

- [ ] 1.2 Remove the original `## Branch Safety Guardrails`
  section (lines 97-111) and its preceding blank line from
  the end of the file.

- [ ] 1.3 Verify the final file structure follows this order:
  `## Instructions` > `### Pre-conditions` (guardrails) >
  `### When arguments are provided` >
  `### When no arguments are provided`. Confirm no duplicate
  guardrail sections exist.

## 2. Verification

- [ ] 2.1 Run `make build` to verify the scaffold asset
  compiles correctly (embedded via embed.FS).

- [ ] 2.2 Run `make test` to verify no drift-detection tests
  fail due to the content reordering.

- [ ] 2.3 Verify constitution alignment: confirm the change
  is purely structural (content reordering, no behavioral
  modifications) per the Observable Quality PASS assessment
  in the proposal.
