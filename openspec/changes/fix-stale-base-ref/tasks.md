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

## 1. Implement resolve_base_ref tool

- [x] 1.1 Add `resolve_base_ref` tool to
  `.opencode/plugins/uf-workflow/index.ts`. Use the
  `createResolveBaseRefTool(exec)` factory pattern with
  an injected executor for `git rev-parse --verify`.
  Implement the three-level fallback chain
  (`upstream/main` -> `origin/main` -> `main`). Return
  `success({ ref, sha, source })` on first match or
  `failure(...)` with `retryable: false` when all three
  fail.

- [x] 1.2 Add unit tests to
  `.opencode/test/uf-workflow.test.ts`. Cover all four
  scenarios: upstream/main resolves, origin/main
  resolves, local main resolves, all fail. Mock the
  executor function -- no real git calls.

## 2. Update command files

- [x] 2.1 [P] Update
  `internal/scaffold/assets/opencode/commands/uf.review-council.md`.
  In the "Resolve Immutable Input Context" section for
  the local (no-PR) case, replace the hardcoded `main`
  base ref with a call to `resolve_base_ref`. Add
  instruction to announce the selected ref and SHA.
  Add instruction to abort if the tool returns failure.

- [x] 2.2 [P] Update
  `internal/scaffold/assets/opencode/commands/uf.unleash.md`.
  Remove the `git diff --name-only main...HEAD`
  instruction and the `## Key Files Changed` section
  from the Step 10 demo output template.

## 3. Verify

- [x] 3.1 Run `make plugin-test` to verify unit tests
  pass.

- [x] 3.2 Verify constitution alignment: the tool
  returns structured output with provenance (Principle
  III), uses factory injection for test isolation
  (Principle IV), introduces no mandatory dependencies
  (Principle II), and uses only hardcoded string literal
  refs with no user-supplied input (Principle V).

- [x] 3.3 File documentation issue for user-facing
  changes (documentation gate). Update CHANGELOG.md
  with change entry.

<!-- spec-review: passed -->
<!-- code-review: passed -->
