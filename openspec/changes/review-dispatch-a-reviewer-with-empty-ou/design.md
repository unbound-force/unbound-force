## Context

Current flow (at `210f54a6`) for a Divisor reviewer child:

1. `agent-executor.ts:574` extracts `text = responseText(parts)` and `usage`.
2. `agent-executor.ts:606` returns `status: "success"` if `info.error` is absent and model matches — even when `text` is `""` (empty) or the run ended via `finish: "length"` with `tokens.output: 0` and `tokens.reasoning: 32000`.
3. Dispatch harvests submissions → reviewer with empty output has 0 findings.
4. `runVerdict([]) → "APPROVE"` (3994) is assigned at 5085 as `workflow_verdict`. The same zero-findings `APPROVE` applies to a length-terminated run whose truncated, non-empty text yielded no submissions.
5. Only signal is `empty_coverage_attestation` (4912), scoped to manifest-mode, advisory.

Issue #1052 shows the impact: the dispatched run appears as `status success | workflow_verdict APPROVE`, `run_counts` shows `success 5, failed 0`, and the council report says "adversary APPROVE" with no empty indication. For inline mode (637-line diff) no warning fires at all.

Existing error handling in the executor already distinguishes retryable failures (`timeout`, `child_prompt_failed`, `child_message_error` with `isRetryable`). The dispatch already handles `status != "success"` by setting `workflow_verdict: null` (5085) and excluding the run from `successfulPersonaVotes` (2210), plus existing retry/no-success paths (RD-FR-001..003, RD-FR-025..027).

The fix belongs at the executor boundary: empty output should be a failed artifact there, so all downstream consumers (review-dispatch, eval harness, gh-hud retry loops) see a failed run without per-consumer patching. Length policy for truncated-but-non-empty text, and the precedence against accepted submissions, need the dispatch layer's submission context, so the rule is re-derived there (D2).

## Goals / Non-Goals

### Goals

- Classify a reviewer whose final response is incomplete (empty final text, or finish reason `length`) so it never casts an approving vote: with no accepted submissions → `failed` (`child_empty_response`, retryable); with accepted submissions → `success` with a forced kind-native non-approving verdict.
- Surface the finish reason as structured provenance on executor success results so dispatch can apply length policy to non-empty (truncated) runs.
- Ensure an incomplete reviewer never contributes an `APPROVE` vote to `successfulPersonaVotes` / council verdict; it appears as `run_counts.failed` or a forced non-approving vote, plus a visible warning.
- Keep change minimal and deterministic: no new dependency, no gate weakening, scaffold mirrors byte-identical.
- Make failure observable end-to-end (executor → dispatch artifact → council report → eval harness).

### Non-Goals

- Changing the output token budget (32k vs 65k) or reasoning-model behavior — that only changes frequency, not the fail-open.
- Reworking `empty_coverage_attestation` beyond the new `child_empty_response` warning (follow-up if broader inline coverage needed).
- Changing finding deduplication, severity ladders, or `finalize_review_dispatch` schema — only the status, verdict-forcing, and warning path for incomplete final responses.
- Adding automatic retry loops inside dispatch — re-execution stays operator-invoked (RD-FR-062).

## Decisions

### D1: Fail-closed at the executor (primary guard)

Insert check after `text`/`usage` extraction in `agent-executor.ts:574-606`, before model-mismatch and success return:

- `if (text.trim().length === 0)` → return `failed` via `failedResult(provenance, "child_empty_response", message, true)` with `text`/`usage` attached. The 256-char message should name the condition (empty final text, finish length, provider/model if available) so the dispatch warning can surface it without re-parsing.
- Length-terminated runs: extract the finish reason from `step-finish` parts. With empty text the empty-text check already fails it; preserve `finishReason` in the error message when present. With non-empty (truncated) text the executor returns `success` — it cannot judge truncation severity and cannot see harvested submissions — but MUST attach the finish reason as a structured field on the success result (`null` when unreported), enabling the dispatch-layer policy (D2, RD-FR-060).
- `failedResult` truncation (MAX_ERROR_LENGTH) applies; caller dispatch already handles truncated messages. Keep `retryable: true` so operator-invoked re-runs (DR-FR-002, gh-hud) can re-execute the reviewer — matching existing transient failure handling.
- No whitespace-only success: trimming avoids counting whitespace as content.

Alternative rejected: returning `success` with a warning flag. That would require every consumer to remember to exclude the run from `successfulPersonaVotes`. Fail-closed at the boundary guarantees no `APPROVE` vote without per-consumer changes.

### D2: Re-derivation and precedence at dispatch finalization (secondary guard)

Even though the executor fix handles the normal path, `review-dispatch` re-derives the incomplete-response rule when building `DispatchRun` entries (around 5085) — covering executor-bypassing callers (direct `dispatch_agent_run`) and length-terminated non-empty runs the executor passes as success. A run is incomplete when its final text is empty or its reported finish reason is `length`:

- `success` + incomplete + zero accepted submissions → reclassify `failed`, `child_empty_response`, `retryable: true`, `workflow_verdict: null`. `run_counts` follow from `status`.
- `success` + incomplete + accepted submissions → keep `success`; force `workflow_verdict` to the kind-native non-approving value (`REQUEST CHANGES` for council/test-review, `NEEDS-CLARIFICATION` for triage, `AUTHOR-DECIDES` for feedback).
- `failed` with `child_empty_response` + accepted submissions → RD-FR-026 precedence: restore `success`, submissions contribute, forced non-approving verdict. The executor fails empty text unconditionally (it cannot see submissions); the dispatch layer restores when the run's work product exists.

Coordination with `fix-review-dispatch-recovery` (RD-FR-025..027): the triggers are disjoint — session aborts surface as `child_message_error`/`invocation_failed`, never `child_empty_response`, and clean length termination surfaces on the success path. The restore branch runs before RD-FR-025 success-scoping, so restored runs never produce orphaned-submissions warnings. Post-submission aborts keep RD-FR-026's rule entirely unchanged: success, contribute, verdict derived from findings.

Mechanism: the forced verdict is applied at the same place `runVerdict` maps findings to verdicts today (dispatchRuns construction, ~5085), so `successfulPersonaVotes` (~2210), `calculatedWorkflowValue`, and the finalizer's recomputation agree by construction. Successful runs keep a non-null `workflow_verdict`, so the payload still passes the finalizer's semantic validation ("Successful runs MUST have a non-null workflow_verdict"); no change to `successfulPersonaVotes`' predicate is needed — a forced `REQUEST CHANGES` vote counts as a non-approving vote, which is the intent.

### D3: Observable warning and report

- Warning kind is exactly the string `child_empty_response` (machine-checkable, RD-FR-061), alongside existing kinds like `empty_coverage_attestation`, `orphaned_submissions`. Warning `detail` carries `agent=<agent> run=<id>` and, when available, `finish=length tokens.output=0 tokens.reasoning=32000`.
- Dispatch artifact (`payload.warnings`, `build_warnings` around 4482) includes the new warning when any run was reclassified or verdict-forced. Council report renders it in the warnings section — operators see which reviewer was incomplete and its finish reason/tokens.
- No duplication: when a run's final text is empty, emit `child_empty_response` and suppress `empty_coverage_attestation` for that run (prefer the more specific kind), regardless of delivery mode; never emit two `child_empty_response` warnings for the same run. `empty_coverage_attestation` still fires independently for non-empty manifest-mode text lacking a coverage attestation.
- Schema impact: `error.code` and warning kinds are existing string enums extended additively; `finalize_review_dispatch` schema shape does not change. Downstream consumers (eval harness, gh-hud) handle unknown codes/kinds additively — failed-run and warning handling already treat them as data.
- `run_counts` naturally reflects the no-submissions branch (failed +1) because the run's `status` is `failed`. Eval harness can assert on `run_counts.failed > 0` and on absence of `APPROVE` from an incomplete reviewer.

### D4: Eval harness check

Deterministic two-branch check: a reviewer session with `finish: "length"` or empty output MUST NOT yield an approving vote — with zero accepted submissions it satisfies `status != "success" && workflow_verdict == null`; with accepted submissions it is `success` with a non-approving verdict. The harness currently passes even when adversary was empty (all 8 runs passed); this closes the harness gap noted in #1052 Notes.

### D5: Mirrors and tests

- Mirror edits byte-identically to `internal/scaffold/assets/opencode/lib/agent-executor.ts` and `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts` — drift gate `TestEmbeddedAssets_MatchSource` enforces.
- Executor tests: empty text → `child_empty_response` failed; whitespace-only → failed; non-empty → success; length finish + empty → failed; length finish + non-empty → success carrying the structured finish-reason field (`null` when unreported); `retryable: true`; usage preserved.
- Dispatch tests: incomplete + zero submissions (empty or truncated text) → `failed`, `workflow_verdict: null`, excluded from `successfulPersonaVotes`, `run_counts.failed` incremented; incomplete + accepted submissions → `success` with forced kind-native non-approving verdict and submissions contributing; executor-failed `child_empty_response` + submissions → restored `success` with forced verdict; RD-FR-026 abort-with-submissions case unchanged (verdict from findings); council with one incomplete + one real reviewer → non-`APPROVE`; warning exact kind with agent/run/finish/tokens; no `empty_coverage_attestation` duplication when text is empty; no automatic re-execution within one dispatch call; all-failed exhaustion → non-approving per the RD-FR-007 zero-success path.
- No change to `.opencode/vitest.config.ts` coverage floor or CI determinism gates.

## Risks / Trade-offs

- **Risk**: A brief but legitimate empty response (e.g., reviewer intentionally silent on clean code) would be reclassified as failure rather than `APPROVE`. Accepted: reviewers are expected to emit at least a summary/coverage attestation; empty is indistinguishable from truncation and should not be treated as a reviewed approval. The tradeoff is fail-closed over availability — appropriate for v0.22.0 Tier 1 where unattended review PRs must not approve on missing evidence.
- **Risk**: Retryable `child_empty_response` could loop if the model deterministically hits the length limit. Bounded by design (RD-FR-062): a dispatch executes each planned run at most once (no automatic re-execution), re-runs require operator-invoked DR-FR-002 partial re-dispatch, and exhaustion lands on the RD-FR-007 zero-success non-approving path. Follow-up can add token-budget tuning without changing this fix.
- **Risk**: A length-terminated run with accepted submissions now votes its forced non-approving value instead of a findings-derived verdict, so a council where it was the only reviewer can no longer return `APPROVE WITH ADVISORIES`. Accepted: an incomplete final response means the reviewer's coverage attestation and full reasoning are missing; its findings still surface in the artifact, but its vote cannot count as a completed review.
- **Risk**: Slight increase in `run_counts.failed` for existing flaky models. Mitigated by observability: warning + error code makes root cause obvious for gh-hud operators; alternative (silent APPROVE) is strictly worse.
- **Trade-off**: Not widening the budget to 65k now avoids touching provider config; keeps change small and parallelizable with other Tier 1 issues (#1078, #1055, #1051). Budget tuning can land separately.
