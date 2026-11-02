<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, and (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file — parallel
  workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.

  Coverage strategy (Constitution IV): every new code path —
  the tolerant gh runner (including the per-invocation
  timeout), JSON-array-on-any-exit parsing, the no-checks and
  not-found stderr markers, all five bucket-classification
  outcomes (pending, fail, cancel, skipping, unknown), and the
  bucket-derived `ci_pending` — gets at least one positive and
  one negative test in .opencode/test/ (vitest 5, mocked gh
  process output, no network or gh binary). The suite runs
  with coverage via `make plugin-test` (94 percent
  statements floor, governance-controlled, not modified);
  `make check` enforces repo gates. TypeScript pack MUSTs
  apply to all edits: no `any` (CS-003), explicit return
  type annotations on exported symbols (CS-005), JSDoc with
  @param/@returns on every exported or documented helper
  (DR-001), and all Promises handled (AP-005).
-->

## 1. Pre-Implementation Verification (read-only)

- [x] 1.1 Verify in `.opencode/plugins/uf-finale/index.ts` that
      `createUfFinaleCiPollTool`'s polling loop is the only
      `gh pr checks` call site in the uf-finale plugin, and grep
      the repo for remaining `--json` field lists containing
      `conclusion` or `detailsUrl` against `gh pr checks` —
      expected: exactly two call sites in scope (uf-finale
      polling loop; review-dispatch section 1.2c) and none
      beyond them.
- [x] 1.2 Confirm the test harness entry points:
      `.opencode/test/helpers/gh-spawn-mock.ts` (Bun.spawn mock)
      supports or can be adapted for direct
      `createUfFinaleCiPollTool` factory invocation as used by
      `.opencode/test/uf-finale.test.ts`, and locate the
      `uf_finale_ci_poll` command-string pin in
      `.opencode/test/golden-output-shape.test.ts`.

## 2. uf_finale_ci_poll Fix (FCP-FR-001..004, design D1-D5)

All tasks in this group modify
`.opencode/plugins/uf-finale/index.ts` — sequential.

- [x] 2.1 Add a module-local exit-code-tolerant gh runner
      (argument-array `Bun.spawn`, `AbortSignal` per-invocation
      timeout, resolves `{stdout, stderr, exitCode}` without
      throwing) with JSDoc citing design D2 and a pointer
      comment to review-dispatch's `runGhCommandTolerant` as the
      prior-art twin. The timeout is normative per FCP-FR-001
      and its hung-invocation scenario.
- [x] 2.2 Rewire the polling loop: request
      `--json name,state,bucket,link`; parse stdout as a JSON
      array on ANY exit code, except the FCP-FR-002 marker
      shapes, and treat other non-array stdout as `GH_CLI_ERROR`
      carrying gh's stderr; detect the checkless shape
      (/no checks reported/i stderr, empty stdout) and the
      PR-not-found stderr pattern, returning code `PR_NOT_FOUND`
      with `retryable: false`; classify from `bucket` (pending
      keeps polling, fail/cancel populate failedChecks,
      pass/skipping pass, unknown fail-closed as failed) with
      `failedChecks` entries `{name, conclusion <- state,
      url <- link}`. Preserve the `gh pr view` conflict
      pre-check, `CI_POLL_TIMEOUT`, and `INVALID_PR_NUMBER`
      unchanged. Mark the stderr markers with gh-version
      verification comments.
- [x] 2.3 Sync the scaffold twin
      `internal/scaffold/assets/opencode/plugins/uf-finale/index.ts`
      byte-identical and verify with `diff`.

## 3. uf_finale_ci_poll Tests (FCP-FR-005, design D7)

- [x] 3.1 Migrate and extend the `uf_finale_ci_poll` describe
      block in `.opencode/test/uf-finale.test.ts`: move
      checks-command mocking to the spawn-mock harness (adapting
      `.opencode/test/helpers/gh-spawn-mock.ts` if needed);
      replace `conclusion`/`detailsUrl` fixtures with verified
      `name,state,bucket,link` shapes; add scenarios for exit 8
      with a pending array (poll continues, later pass), exit 1
      with a failing array, exit 1 with empty stdout plus the
      no-checks marker (both workflows branches), PR-not-found
      stderr (asserting `PR_NOT_FOUND`), cancel-as-failed,
      skipping-as-pass, a simulated hung invocation bounded by
      the per-invocation timeout, and unknown-bucket fail-closed
      (`state: "NEUTRAL"`, `bucket: "blocked"`). Keep the
      terminal-silence tests green.
- [x] 3.2 [P] Update the `uf_finale_ci_poll` command pin in
      `.opencode/test/golden-output-shape.test.ts` to
      `gh pr checks 43 --json name,state,bucket,link` (different
      file from 3.1; the pin is the primary regression pin and
      must fail against the pre-fix implementation).

## 4. review-dispatch ci_checks Sub-Call (FPR-FR-017, design D6)

- [x] 4.1 Switch the section 1.2c sub-call in
      `.opencode/plugins/review-dispatch/index.ts` to the
      exit-code-tolerant `runGhCommandTolerant` path with
      `--json name,state,bucket`, mapping `status <- state` and
      `conclusion <- bucket` into the existing
      `{name, status, conclusion}` shape; treat a JSON array as
      data on any exit code so pending (exit 8) and failing
      (exit 1) PRs populate `ci_checks`; keep the non-fatal
      degradation path for genuine failures; derive `ci_pending`
      from the bucket vocabulary (true when any check buckets as
      `pending`) instead of comparing `status` to `"completed"`;
      add the reverse pointer comment on `runGhCommandTolerant`
      naming the new uf-finale twin (design D2/R4); sync the
      review-dispatch scaffold twin under
      `internal/scaffold/assets/opencode/plugins/` (the
      `review-dispatch/index.ts` mirror) byte-identical and
      verify with `diff`.
- [x] 4.2 Update the ci_checks fixtures in
      `.opencode/test/fetch-pr-review-context.test.ts` to the
      verified `name,state,bucket` shapes and add pending
      (exit 8) and failing (exit 1) sub-call scenarios; assert
      the green path returns populated `ci_checks` with
      `ci_pending: false` and no `ci_checks degraded` warning,
      the pending path returns `ci_pending: true`, and keep a
      genuine sub-call-failure degradation test.

## 5. Verification and Records

- [x] 5.1 Run `make plugin-test` (94 percent statements
      coverage floor enforced; do not modify the floor) and
      `make coverage-gate`; fix any regressions.
- [x] 5.2 Run `make check` for CI parity (lint, vet, build,
      coverage gates, plugin tests, scaffold drift tests).
- [x] 5.3 Documentation gate (restorative user-visible
      behavior — do NOT claim the no-change exemption):
      (a) file the in-repo documentation issue with the `docs`
      label; (b) file the unbound-force/website documentation
      issue with the `docs` label (constitution Cross-Repo
      Documentation rule); (c) add a
      `fix-uf-finale-ci-poll-gh-schema` entry under
      `## Unreleased` -> `### Fixed` in `CHANGELOG.md` with the
      trailing reference `(Spec:
      openspec/changes/fix-uf-finale-ci-poll-gh-schema/,
      Fixes: #895, Docs: #NNN, unbound-force/website#NNN)`,
      threading both issue numbers at filing time; (d) add a
      Docs bullet to proposal.md's Impact section recording
      both issue numbers and the content-pipeline disposition
      (blog/tutorial skipped: restorative change measured
      against the pre-break baseline). Precedent: PR #744.
- [x] 5.4 Verify constitution alignment against the implemented
      result: Principle IV (the FCP-FR-001..005 scenarios and
      the pass/pending/fail FPR-FR-017 scenarios are passing
      mocked tests; the two FPR-FR-017 divergence scenarios —
      "PR-not-found stderr degrades the advisory sub-call"
      and "Unknown JSON fields are never requested" — are
      implementation-verified but were not test-pinned in the
      implement-phase suite; their regression pins were added
      in the code-review phase per the round-6 adjudication;
      coverage floor holds), Principle V
      (checks invocation is argument-array spawn — no shell
      layer), Principle VI (field list, exit codes, and stderr
      markers pinned in tests), and Principles I-III (output
      contract and envelope unchanged). Report any deviation.
