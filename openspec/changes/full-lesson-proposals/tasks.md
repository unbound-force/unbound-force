<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, and (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Plugin contract (design D1-D2)

- [x] 1.1 In `.opencode/plugins/review-dispatch/index.ts`, make
      `projectSummaryProposal` copy every proposal field verbatim and
      drop `informationChars`, `tagChars`, and `agentChars` from
      `SUMMARY_LIMITS` (RD-FR-058).
- [x] 1.2 Add a post-assembly byte budget to
      `buildCompactFinalization` that drops trailing proposals whole
      and sets `proposals_truncated` (RD-FR-058, D2).
- [x] 1.3 Update the `consolidate_dispatch` tool description:
      proposal fields are never shortened; dropped entries are
      counted by `proposals_truncated`.

## 2. Tests

- [x] 2.1 `summary-retrieval.test.ts`: verbatim long proposal,
      whole-entry drop under budget for the summary and the compact
      result, no `truncated` key on proposals, updated limit
      literals.
- [x] 2.2 `consolidate-dispatch.test.ts`: end-to-end harvest
      regression (`submit_lesson_proposal` ->
      `dispatch_review_runs` -> summary, full, finalization,
      `read_dispatch_artifact`) asserting the complete text; update
      the RD-FR-054 preview assertion.

## 3. Commands (design D3)

- [x] 3.1 Correct the preview description in `uf.review-council.md`
      Section 4 and `uf.review-pr.md` Section 5.

## 4. Mirror, documentation, CI parity

- [x] 4.1 Copy every edited `.opencode/` file byte-identical to
      `internal/scaffold/assets/opencode/`.
- [x] 4.2 Add a CHANGELOG.md entry.
- [x] 4.3 Run `make check`.
