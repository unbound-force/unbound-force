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

  NOTE: All tasks in this change modify the same file
  (.opencode/commands/finale.md), so no tasks are
  parallel-eligible.
-->

## 1. Add mergeability gate to step 6

- [x] 1.1 Add "no checks reported" detection sub-step
  to step 6 (Watch CI Checks) in
  `.opencode/commands/finale.md`. When `gh pr checks`
  returns "no checks reported," insert a mergeability
  query: `gh pr view <number> --json
  mergeable,mergeStateStatus`. Document the three
  `mergeable` values (`CONFLICTING`, `UNKNOWN`,
  `MERGEABLE`) and the agent's required response to
  each per FR-001.

- [x] 1.2 Add conflict recovery options sub-step per
  FR-002. When `mergeable` is `CONFLICTING`, the agent
  MUST present three options: (a) rebase onto target
  branch and re-push, (b) stop for manual resolution,
  (c) continue with warning. Include the specific
  commands (`git fetch`, `git rebase`, `git push
  --force-with-lease`) and the abort path
  (`git rebase --abort`) for failed rebases.

- [x] 1.3 Add workflow file cross-reference sub-step
  per FR-003. When `mergeable` is `MERGEABLE` and no
  checks are reported, check for workflow files in
  `.github/workflows/`. If workflow files exist, warn
  the user. If none exist, accept and proceed.

## 2. Update summary step

- [x] 2.1 Update step 8 (Summary) in
  `.opencode/commands/finale.md` to include a warning
  line when checks were skipped due to merge conflict
  (user chose "continue anyway" in FR-002). The summary
  MUST display: "CI checks did not run due to merge
  conflict."

## 3. Verification

- [x] 3.1 Read the final `.opencode/commands/finale.md`
  and verify: (a) step 6 now handles "no checks
  reported" with the mergeability gate, (b) conflict
  recovery options match FR-002 exactly, (c) workflow
  cross-reference matches FR-003, (d) step 8 includes
  the conflict warning, (e) existing step 6 behavior
  for "checks pass" and "checks fail" is unchanged.

- [x] 3.2 Verify constitution alignment: confirm the
  change does not introduce new hero dependencies
  (Composability First), does not alter artifact formats
  (Autonomous Collaboration), and improves observability
  by surfacing blockers (Observable Quality). N/A for
  Testability (Markdown-only change) and Security by
  Default (no new inputs or dependencies).
<!-- scaffolded by uf vdev -->
<!-- spec-review: passed -->
<!-- code-review: passed -->
