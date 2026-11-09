<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Plugin tool

- [x] 1.1 Implement `createCheckoutPrHeadTool(exec, gh, cwd)` in
  `.opencode/plugins/uf-workflow/index.ts` per design D2
  (AFC-FR-001..003): argument-array git/gh calls, the six failure
  codes plus `CHECKOUT_PR_HEAD_ERROR`, exported
  `CHECKOUT_PR_HEAD_ERROR_CODES` and `CheckoutPrHeadData`; register
  `checkout_pr_head` in `UfWorkflowPlugin.server()`.
- [x] 1.2 Copy the plugin byte-identical to
  `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
  (verify with `cmp`). No manifest change — the asset already ships.

## 2. Tests

- [x] 2.1 Create `.opencode/test/checkout-pr-head.test.ts` with
  injected git/gh fakes over a mutable repo model: equal and
  descendant identity (no checkout), fork default-branch checkout,
  `DIRTY_TREE`, untracked-collision and fetch-failure
  `CHECKOUT_FAILED` (retryable split), `UNRELATED_BRANCH` after
  checkout and on merge-ref mismatch, detached and unset
  `NO_PUSH_TARGET`, `READ_ONLY_MERGE_REF`, `UNSAFE_PUSH_TARGET`
  for branch, merge ref, and remote forms, argument-schema
  rejections, and command code-list parity (AFC-FR-004).
- [x] 2.2 [P] Add `checkout_pr_head` to the uf-workflow tool lists in
  `.opencode/test/uf-workflow.test.ts` and
  `.opencode/test/scratch-smoke.test.ts`.
- [x] 2.3 [P] Add a golden output-shape block for
  `checkout_pr_head` to `.opencode/test/golden-output-shape.test.ts`
  (registry dispatch with mocked `execFileSync`).
- [x] 2.4 DE-007 seam verification naming the three layers:
  registration surface (2.2 name-list pins; the dynamic
  `registration-surface.test.ts` loop covers args as a zod raw
  shape; 2.1 pins the argument set and schema); documented output
  shape (2.3 golden ok payload and failure code through the
  registry, matching the Section 1.2.1 JSON example); write→read
  round trips (2.1 and 2.3 drive `gh pr checkout` mutating the repo
  model and re-read HEAD, branch, and config through the tool — the
  tool persists no files, so this is the only round trip it has).

## 3. Command and contract tests

- [x] 3.1 Replace Section 1.2.1 of
  `.opencode/commands/uf.address-feedback.md` with the tool call,
  one JSON output example (DE-002), and per-code handling; leave
  Section 4.4 unchanged (it reads the values "recorded in Section
  1.2.1"); do not touch Sections 4.5/4.6.
- [x] 3.2 Sync the mirror
  `internal/scaffold/assets/opencode/commands/uf.address-feedback.md`
  byte-identical (`cmp`).
- [x] 3.3 Rewrite the Section 1.2.1 clauses of
  `TestCommandContracts_AddressFeedback` in
  `internal/scaffold/command_contract_test.go` against the tool
  contract and forbid the removed git templates.

## 4. Documentation and verification

- [x] 4.1 Add a `CHANGELOG.md` entry with the DE-004 delta
  (Section 1.2.1 108 → 50 lines per copy), spec footer, and
  `(Docs: #967)` — the existing docs issue covers the behavior.
- [x] 4.2 Run `make check` and `make lint-instructions`; confirm
  green with coverage floors unchanged.
