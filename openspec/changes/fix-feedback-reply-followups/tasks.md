<!--
  [P] marks tasks eligible for parallel execution.
  Every .opencode/ file edited here has a byte-identical
  mirror under internal/scaffold/assets/opencode/.
-->

## 1. Shared classifier

- [ ] 1.1 Create `.opencode/lib/gh-error.ts`: `classifyGhFailure`
  implementing the D2 decision table (text keys first, exit
  code 4 as fallback authentication signal, first match wins),
  with the `GhFailureSignals` / `GhErrorClassification` types.
- [ ] 1.2 Add `.opencode/test/gh-error.test.ts`: decision-table
  rows, text-beats-exit-code precedence, and parity pins
  running identical inputs through `classifyGhError` and
  `ghFailureError`.

## 2. Consumer adapters

- [ ] 2.1 Rework `classifyGhError` in
  `.opencode/lib/feedback-replies.ts` into an adapter over
  `classifyGhFailure`, preserving its message templates.
- [ ] 2.2 Rework `ghFailureError` and `classifyError` in
  `.opencode/plugins/review-dispatch/index.ts` to classify
  through the shared module via a `{ghCode, retryable}`
  annotation; keep prologue direct checks untouched.
- [ ] 2.3 Add the review-dispatch-side pin in
  `.opencode/test/fetch-pr-checks.test.ts`: exit 4 with
  non-matching stderr classifies `gh_auth_failed`,
  retryable false.

## 3. createIssue timeout guard

- [ ] 3.1 In `createIssue`, skip the retry on a classified
  `timeout` and fail immediately with an error message that
  directs the operator to check the repository for the issue
  before re-running (D3).
- [ ] 3.2 Extend `.opencode/test/feedback-replies.test.ts`:
  timeout no-retry case (exactly one gh invocation), the
  text-timeout and exit-4-plus-rate-limit classification
  rows, and keep the existing transient-retry pin green.

## 4. Command wording

- [ ] 4.1 Replace the crash-recovery sentence in Section 1.7
  of `.opencode/commands/uf.address-feedback.md` so it says
  to pass the full confirmed item list again, unchanged,
  including items whose replies were already posted (FRR-FR-003).

## 5. Mirrors, registration, docs

- [ ] 5.1 Sync every edited `.opencode/` file to
  `internal/scaffold/assets/opencode/` byte-identically,
  including the new `gh-error.ts`.
- [ ] 5.2 Register `gh-error.ts` in `scaffold.go`
  (`isToolOwned`, `isDivisorAsset`),
  `plugin_activation.go` (`stageReviewPluginSources`), and the
  `scaffold_test.go` tables with their count comments.
- [ ] 5.3 Add the CHANGELOG entry under Unreleased.
- [ ] 5.4 File the user-facing documentation issue in
  unbound-force/unbound-force for the error-code alignment
  and the no-duplicate-filing behavior.

## 6. Verification

- [ ] 6.1 Run `make plugin-test` and
  `go test -race -count=1 ./internal/scaffold/...`.
- [ ] 6.2 Run `make lint-instructions` and
  `make coverage-gate`; confirm no new instruction-lint
  findings beyond the pre-existing IL-030/IL-031 warnings on
  uf.address-feedback.md.
- [ ] 6.3 Validate the change:
  `openspec validate fix-feedback-reply-followups --strict`.
