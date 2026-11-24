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

## 1. Declared-checks execution root

- [ ] 1.1 In `.opencode/lib/preflight-declared.ts`, change
  `runDeclaredChecks` to take `{configRoot, execRoot}` per design
  D4.
  - Config, `run` paths, `requires` probes, and `PROJECT_ROOT_ENV`
    come from `configRoot`.
  - Execution `cwd`, `computeDiffScope`, and the `classifyFailures`
    merge-base come from `execRoot`.
  - When `execRoot === configRoot`, output must be identical to the
    current output.
- [ ] 1.2 Sync `internal/scaffold/assets/opencode/lib/preflight-declared.ts`
  byte-identical (`cmp`).
- [ ] 1.3 Extend `.opencode/test/preflight-declared.test.ts`:
  - Omitted root is unchanged (pin the existing fixtures).
  - With `execRoot`, the check runs with that `cwd` and scope uses
    its `HEAD`.
  - Scripts and config still resolve from `configRoot`.
- [ ] 1.4 Add a real-git case to
  `.opencode/test/preflight-declared-roundtrip.test.ts`:
  - The primary checkout is on a commit where a required check
    passes.
  - A worktree at a PR-head commit makes it fail.
  - The soft-gate result is `branch-caused` relative to the
    head's merge-base, not the primary's.

## 2. `review_head_root` tool

- [ ] 2.1 Add a module-level live-root registry in
  `.opencode/plugins/uf-workflow/index.ts`, injectable into
  factories for tests.
- [ ] 2.2 Implement `createReviewHeadRootTool(exec, gh, cwd, fs,
  registry)` per design D1-D3.
  - `prepare`: stale pruning, the checkout reuse path, fetch of
    `refs/pull/<N>/head` from the remote matching the base repo,
    SHA verification, and detached worktree add.
  - `release`: removes registered roots only.
  - Export `REVIEW_HEAD_ROOT_ERROR_CODES` (`HEAD_UNAVAILABLE`,
    `HEAD_MOVED`, `WORKTREE_FAILED`, `INVALID_ROOT`) and
    `ReviewHeadRootData`.
  - Register `review_head_root` in `UfWorkflowPlugin.server()`.
- [ ] 2.3 Add the optional `root` argument to
  `createPreflightDeclaredChecksTool`. Validate it against the
  plugin directory and the registry (else `INVALID_ROOT`) and pass
  `{configRoot: cwd, execRoot: root ?? cwd}`.
- [ ] 2.4 Sync `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
  byte-identical (`cmp`).

## 3. Tool tests

- [ ] 3.1 Create `.opencode/test/review-head-root.test.ts` with
  injected git/gh/fs fakes, covering:
  - The checkout reuse path, the dirty-at-head path creating a
    worktree, and a present commit creating a worktree.
  - Missing commit: fetch then create.
  - Remote matching across SSH/HTTPS/`.git` URL forms.
  - Failures: `HEAD_MOVED`; `HEAD_UNAVAILABLE` for no remote and
    for a retryable fetch error; `WORKTREE_FAILED` cleans its temp
    dir.
  - Release: of a created root, of the reused checkout (no-op),
    and of an unknown path (`INVALID_ROOT`, nothing deleted).
  - Stale pruning, argument-schema rejections, and that every call
    uses argument arrays.
- [ ] 3.2 [P] Add `review_head_root` to the uf-workflow tool lists
  in `.opencode/test/uf-workflow.test.ts` and
  `.opencode/test/scratch-smoke.test.ts`.
- [ ] 3.3 [P] Add golden output-shape blocks for `review_head_root`
  and for `uf_preflight_declared_checks` with `root` to
  `.opencode/test/golden-output-shape.test.ts`.
- [ ] 3.4 Add a real-git round trip: `prepare` creates a worktree at
  a second commit, then `uf_preflight_declared_checks` with that
  `root` runs there, then `release` removes it. The user's
  `HEAD`, branch, and index are unchanged throughout.
- [ ] 3.5 DE-007 seam verification naming the three layers:
  registration surface, documented output shape, and write→read
  round trip (3.4).

## 4. Skill and commands

- [ ] 4.1 `.opencode/skills/pre-flight/SKILL.md`: add the optional
  consumer-supplied `root` per design D5.
  - Phase 4 runs detected tools with that working directory.
  - Phase 2b passes `root` to the declared-checks tool.
  - Phase 5 adds `**Gates ran at**: <head_sha> (<method>)`.
- [ ] 4.2 `.opencode/commands/uf.review-pr.md`:
  - Add the `prepare` call after Step 3, with one JSON output
    example (DE-002).
  - Add per-code STOP handling with remediation (design D6).
  - Pass `root` to Step 3.7.
  - Add the single "release before finalizing whenever prepare
    succeeded" rule.
- [ ] 4.3 `.opencode/commands/uf.review-council.md`, for PR input
  only:
  - Add `prepare` at the start of Code Review Mode step 1.
  - Pass `root` to Phase 1a.
  - Gate Phase 1b on `method: "checkout"`, with the
    worktree-skip note.
  - On `prepare` failure: coverage `FAIL` plus `INCONCLUSIVE`
    through the shared protocol.
  - Add the release rule.
  - Leave local/uncommitted input unchanged.
- [ ] 4.4 [P] Sync the scaffold mirrors of 4.1-4.3 byte-identical
  (`cmp`).
- [ ] 4.5 Add contract clauses in
  `internal/scaffold/command_contract_test.go` for both commands:
  - `review_head_root` prepare and release are present.
  - Every `REVIEW_HEAD_ROOT_ERROR_CODES` code is handled.
  - `root` is passed to pre-flight.
  - The council's Phase 1b checkout gate is present.
  - A forbidden fragment keeps `gh pr checkout` out of the gate
    path.
- [ ] 4.6 Add a TS parity test pinning each command's code list to
  `REVIEW_HEAD_ROOT_ERROR_CODES` in both copies.

## 5. Documentation and verification

- [ ] 5.1 Verify constitution alignment against the proposal:
  - I: pre-flight output carries `Gates ran at`.
  - II: an omitted `root` is unchanged, and gaze-owned files are
    untouched.
  - III: structured envelopes with stable codes.
  - IV: injectable executors plus a real-git round trip.
- [ ] 5.2 Add a `CHANGELOG.md` entry, and file the documentation
  issue for the user-facing change (PR-mode reviews now gate the
  PR head) per the Documentation gate.
- [ ] 5.3 File a follow-up issue in unbound-force/gaze for an
  analysis-root argument to `gaze-reporter`. Reference it from the
  Phase 1b skip note's CHANGELOG line.
- [ ] 5.4 Add `Closes #1203` to the PR body by hand
  (`originating_issue` carries #1204 only).
- [ ] 5.5 Run `make check` and `make lint-instructions`. Confirm
  green with coverage floors unchanged (plugin floor 94%).
