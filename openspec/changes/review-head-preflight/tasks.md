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

## 1. Declared-checks execution root and base

- [x] 1.1 In `.opencode/lib/preflight-declared.ts`, add optional
  `execRoot` (default `cwd`) and `baseSha` parameters to
  `runDeclaredChecks` per design D6:
  - Config, `run` paths, `requires` probes, and `PROJECT_ROOT_ENV`
    come from `cwd`.
  - Execution, diff scope, and the baseline worktree come from
    `execRoot`.
  - `baseSha` replaces default-branch resolution with
    `merge-base <baseSha> HEAD`.
  - Omitted parameters keep today's output.
- [x] 1.2 Sync `internal/scaffold/assets/opencode/lib/preflight-declared.ts`
  byte-identical (`cmp`).
- [x] 1.3 Extend `.opencode/test/preflight-declared.test.ts`:
  - Omitted parameters are unchanged.
  - With `execRoot`, the run `cwd` and the scope `HEAD` come from
    it.
  - Scripts and config still resolve from `cwd`.
  - With `baseSha`, scope and the baseline merge-base use it.

## 2. `review_head_root` tool

- [x] 2.1 Implement `createReviewHeadRootTool(exec, gh, cwd, fs)` in
  `.opencode/plugins/uf-workflow/index.ts` per design D1-D5.
  - Self-ignoring `.uf/review-head/`.
  - Exact root validation (D3).
  - Pruning.
  - The reuse check with `--untracked-files=normal`.
  - The trust gate: same-repo, or the author's base-repo
    permission is write, maintain, or admin (`gh api`).
  - Remote matching and fetch with `cat-file` re-verification.
  - `worktree add --detach --`.
  - `changed_files` and `config_sha`.
  - `release` with the directory-removal fallback.
  - Export `REVIEW_HEAD_ROOT_ERROR_CODES` (`UNTRUSTED_HEAD`,
    `HEAD_UNAVAILABLE`, `HEAD_MOVED`, `WORKTREE_FAILED`,
    `INVALID_ROOT`), plus `ReviewHeadRootData` and an exported
    `validateReviewHeadRoot` helper.
  - Register the tool in `UfWorkflowPlugin.server()`.
- [x] 2.2 Add optional `root` and `base_sha` to
  `createPreflightDeclaredChecksTool`:
  - Validate `root` with `validateReviewHeadRoot`, else
    `INVALID_ROOT`.
  - Pass both to `runDeclaredChecks`.
  - Name `INVALID_ROOT` in the tool description.
- [x] 2.3 Sync `internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts`
  byte-identical (`cmp`).

## 3. Tool tests

- [x] 3.1 Create `.opencode/test/review-head-root.test.ts` with
  injected git/gh/fs fakes:
  - Reuse: clean at the head; an untracked file forces a worktree;
    a modified tracked file forces a worktree.
  - Trust: same-repo; write-permission fork; read-permission
    fork; deleted fork (`head.repo` null); permission lookup error.
    The last three give `UNTRUSTED_HEAD` with no fetch or worktree
    call made.
  - Fetch: missing head fetched then present; base missing; present
    after fetch with no `FETCH_HEAD` read; still missing gives
    `HEAD_MOVED`.
  - Remote matching across scp, `ssh://`, `https://` with
    userinfo, `.git`, and case variants.
  - `HEAD_UNAVAILABLE` for no remote, a fetch error (retryable),
    and a gh error.
  - `WORKTREE_FAILED` cleans up its directory.
  - Pruning: an old review-head worktree is removed; a young one, a
    non-tool worktree, and `.uf/review-head-x-evil` are kept.
  - Release: of a created root; double release; of the plugin
    directory (no-op); path variants refused; `worktree remove`
    failure still deletes the directory.
  - Argument-schema rejections; `--` before every path argument;
    argv-only execution.
- [x] 3.2 [P] Add `review_head_root` to the uf-workflow tool lists
  in `.opencode/test/uf-workflow.test.ts` and
  `.opencode/test/scratch-smoke.test.ts`.
- [x] 3.3 [P] Add golden output-shape blocks for `review_head_root`
  and for `uf_preflight_declared_checks` with `root` to
  `.opencode/test/golden-output-shape.test.ts`.
- [x] 3.4 Add a real-git round trip to
  `.opencode/test/preflight-declared-roundtrip.test.ts`:
  - The primary checkout is on a commit whose required check
    passes, with an untracked user file.
  - `prepare` at a second commit where the check fails.
  - Declared checks with `root` and `base_sha` report
    `branch-caused`.
  - `release` removes the worktree.
  - The user's `HEAD`, branch, index, untracked file, `git status`
    output, and `.git/config` are unchanged throughout.
- [x] 3.5 DE-007 seam verification naming the three layers:
  registration surface, documented output shape, and write→read
  round trip (3.4).
  - Layer 1: `review_head_root` is pinned in the uf-workflow and
    scratch-smoke tool lists, and `uf_preflight_declared_checks`
    args are pinned to `mode`, `root`, `base_sha`. The dynamic
    `registration-surface.test.ts` loop covers both zod raw
    shapes.
  - Layer 2: golden blocks dispatch `prepare` (reuse path) and both
    `INVALID_ROOT` paths through the registry.
  - Layer 3: the real-git round trip in
    `preflight-declared-roundtrip.test.ts` (prepare → declared
    checks in the root → release, with the user's tree unchanged).

## 4. Skill, commands, and policy file

- [x] 4.1 `.opencode/skills/pre-flight/SKILL.md`: add the optional
  consumer-supplied `root` and `changed_files` per design D7.
  - Phases 1, 2, 4, and the 4a worktree fallback run from `root`.
  - Phase 2a uses `changed_files` verbatim.
  - Phase 2b passes `root` and `base_sha` and handles
    `INVALID_ROOT`.
  - Phase 4a keys its baseline on `base_sha` when supplied.
  - Phase 5 adds the provenance line.
- [x] 4.2 `.opencode/commands/uf.review-pr.md`:
  - Add the `prepare` call after Step 3, with one JSON output
    example (DE-002).
  - Add the per-code "not run" handling.
  - Pass `root`, `changed_files`, and `base_sha` to Step 3.7.
  - Add one "release once right after the last local gate, or before
    any earlier STOP; a release failure is a warning" rule.
- [x] 4.3 `.opencode/commands/uf.review-council.md`, for PR input
  only:
  - Add `prepare` at the start of Code Review Mode step 1.
  - Every code, `UNTRUSTED_HEAD` included, goes through the
    Phase 1a FAIL path. The mandatory wording stays unchanged.
  - Thread the root into Phase 1a.
  - Gate Phase 1b on `method: "checkout"`.
  - Add the same release rule.
  - Leave local/uncommitted input unchanged.
- [x] 4.4 Sync the scaffold mirrors of 4.1-4.3 byte-identical
  (`cmp`).
- [x] 4.5 Update the runtime-contract comment in `.uf/preflight.yaml`
  and `internal/scaffold/assets/uf/preflight.yaml`: in PR mode the
  working directory is the review-head root. Keep the
  `projectOwnedPolicyMirrors` shape pins green.
- [x] 4.6 Add contract clauses in
  `internal/scaffold/command_contract_test.go`:
  - Both commands: prepare and release are present; every code's
    handling; the release-warning rule; root, `changed_files`, and
    `base_sha` threading.
  - Council: the Phase 1b checkout gate.
  - Skill: the Phase 1, 2, 2a, and 4 root clauses.
  - A forbidden fragment keeps `gh pr checkout` out of the gate
    path.
- [x] 4.7 Add a TS parity test pinning each command's code list to
  `REVIEW_HEAD_ROOT_ERROR_CODES` in both copies.

## 5. Documentation and verification

- [x] 5.1 Verify constitution alignment against the proposal:
  - I: provenance line.
  - II: omitted arguments are unchanged, and gaze-owned files are
    untouched.
  - III: structured envelopes with stable codes.
  - IV: injectable executors plus the real-git round trip.
- [ ] 5.2 Record DE-004 line deltas for each edited instruction file
  in the CHANGELOG entry. File the documentation issue for the
  user-facing change per the Documentation gate.
- [ ] 5.3 File a follow-up issue in unbound-force/gaze for an
  analysis-root argument to `gaze-reporter`.
- [ ] 5.4 Add `Closes #1203` to the PR body by hand
  (`originating_issue` carries #1204 only).
- [x] 5.5 Run `make check` and `make lint-instructions`. Confirm
  green with coverage floors unchanged (plugin floor 94%).
  Local results on the rebased branch:
  - Plugin suite: 2710/2710 with `--testTimeout 60000`, 96.55%
    statements. Two real-git push tests time out at 5s under load,
    as noted in checkout-pr-head 4.2.
  - Go coverage gate passes.
  - `lint-instructions`: 0 errors.
  - `golangci-lint` on internal/scaffold: 0 issues.
  - Pre-existing on main, unrelated to this change:
    `golangci-lint` errcheck at `internal/gateway/gateway_test.go:4383`,
    and `TestStealthInit_Integration_GitStatusClean` (the external
    `specify` sub-tool exits 2 locally).
