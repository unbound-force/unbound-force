## Why

A Divisor reviewer that exhausts its output budget while reasoning can return no text (`tokens.output: 0`, `finish: "length"`, `responseText` empty). It is currently recorded as `status: "success"` with `workflow_verdict: APPROVE` derived from zero findings. The council can then approve a change that one reviewer never reviewed. Issue #1052 reproduced this in 4 of 8 eval runs (GLM 5.3 via OpenAI-compatible provider, opencode 1.19.0) — including a clean-change run where `divisor-adversary` was the only adversary reviewer and its empty vote produced `APPROVE WITH ADVISORIES`.

Traced at upstream `210f54a6`:

1. `.opencode/lib/agent-executor.ts:606` returns `success` when the child has no error and the model matches, without checking whether `responseText` (574) is empty or the run finished via length limit.
2. Findings are harvested submissions, so a length-terminated reviewer has 0 findings.
3. `runVerdict` (`.opencode/plugins/review-dispatch/index.ts:3994`) maps 0 findings to `APPROVE`, assigned at `5085` as `workflow_verdict`.
4. The only empty signal is `empty_coverage_attestation` at `4912`, which applies only to manifest-mode runs and is advisory. The affected run was inline mode (637-line diff).

Unattended review PRs (gh-hud, and Dewey Phase 5 `dewey#153` rolling review PRs) depend on that vote being real. This is the highest-priority Tier 1 item in the v0.22.0 release plan (#1060). The harness also needs a check for empty/length-terminated sessions.

Fixes #1052.

## What Changes

1. **Executor: fail empty output, surface finish reason** — In `.opencode/lib/agent-executor.ts`, after resolving `text` and `usage`, classify final text whose trimmed length is zero as failure before returning `success`:
   - If `text.trim().length === 0`, return `failed` with `code: "child_empty_response"`, retryable, carrying the (empty) text and usage for diagnostics.
   - Surface the finish reason from `step-finish` parts as a structured field on `success` results (`null` when unreported), so dispatch can apply length policy to non-empty (truncated) runs.

2. **Dispatch: non-approving incomplete vote** — In `.opencode/plugins/review-dispatch/index.ts`, re-derive the rule at finalization (defense-in-depth for executor-bypassing callers and truncated runs):
   - Incomplete response (empty text or finish reason `length`) with zero accepted submissions → reclassify `failed` (`child_empty_response`, retryable, `workflow_verdict: null`), excluded from `successfulPersonaVotes`, counted in `run_counts.failed`.
   - Incomplete response with accepted submissions → RD-FR-026 precedence: the run is classified `success`, its submissions contribute, and its `workflow_verdict` is forced to the kind-native non-approving value (`REQUEST CHANGES` for council/test-review, `NEEDS-CLARIFICATION` for triage, `AUTHOR-DECIDES` for feedback).
   - A `failed` `child_empty_response` run with accepted submissions is restored to `success` under the same precedence; post-submission session aborts keep RD-FR-026's rule unchanged.
   - Surface a `child_empty_response` warning (exact kind) in the dispatch artifact and council report so operators see which reviewer was incomplete and why.

3. **Eval harness guard** — Add assertions that any reviewer session with `finish: "length"` or empty output never yields an approving vote: zero submissions → non-success (`failed`, `workflow_verdict: null`); accepted submissions → `success` with a non-approving verdict.

4. **Mirrors** — Apply the same edits to `internal/scaffold/assets/` mirrors and keep `TestEmbeddedAssets_MatchSource` byte-identical.

No change to finding deduplication, verdict ladders, or artifact finalization beyond the incomplete-response reclassification and verdict forcing. No new dependency.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `shared-agent-executor` (`agent-executor.ts`): `invoke_agent` detects empty final text and returns `status: failed` (`child_empty_response`, retryable) instead of `success`; success results carry structured finish-reason provenance.
- `review-dispatch` (`review-dispatch/index.ts`): incomplete reviewers do not produce `APPROVE` votes; with no submissions they appear as `failed` in `run_counts`, with submissions they carry a forced non-approving verdict; either way a warning/provenance entry surfaces, and the council cannot approve on an incomplete vote alone.

### Removed Capabilities
- None.

## Impact

- **Live plugin/library**: `.opencode/lib/agent-executor.ts` (empty-response guard, finish-reason provenance, error code/message, retryable flag, usage preservation) and `.opencode/plugins/review-dispatch/index.ts` (finalization re-derivation, verdict forcing, warning kind, report rendering, run_counts handling). Scaffold mirrors under `internal/scaffold/assets/` — byte-identical copies enforced by `TestEmbeddedAssets_MatchSource`.
- **Tests**: `.opencode/test/agent-executor.test.ts` (or nearest executor suite), `.opencode/test/review-dispatch` / `consolidate-dispatch.test.ts` / `golden-output-shape.test.ts` — empty/length cases, with/without-submissions branches, verdict and run_counts assertions, warning golden files; must not weaken coverage floor (94% statements) or modify gate values.
- **Eval harness**: fixture that drives an end-to-end council run where one reviewer returns empty text — must yield non-approving council result.
- **Related changes**: `fix-review-dispatch-recovery` (RD-FR-025..027) is active in parallel and edits the same finalization path (`successfulPersonaVotes` ~2210, warnings ~4482/4919, run-record construction ~5085). Precedence is specified here: incomplete-final-response failure applies to runs with no accepted submissions; runs with accepted submissions keep RD-FR-026's success-and-contribute outcome plus the forced non-approving verdict, and post-submission aborts keep RD-FR-026's rule unchanged. Triggers are disjoint at the executor (aborts surface as `child_message_error`/`invocation_failed`; length termination surfaces on the success path).
- **Documentation**: CHANGELOG.md entry required for the behavioral change (empty/length-terminated reviewer runs no longer approve). AGENTS.md and README.md assessed as no-change (no new commands, gates, or user-facing workflow); the warning renders through the existing council-report warnings section.
- **Risk**: Low. Change is fail-closed (incomplete → failed or forced non-approving). Existing complete runs unaffected. The only behavioral shift is that a council with one incomplete reviewer no longer silently approves — which is the intended fix.
- **Out of scope**: raising the 32k token cap/provide 65k limit, reasoning-model tuning, and broadening `empty_coverage_attestation` to inline mode beyond the new warning — those are follow-ups if needed.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

Change preserves artifact-based collaboration: incomplete reviewers are reclassified or verdict-forced at their boundaries into machine-parseable artifacts (`child_empty_response` provenance, forced non-approving verdict). Downstream heroes (dispatch, council report, gh-hud retries) consume the artifact without synchronous negotiation. No runtime coupling added.

### II. Composability First

**Assessment**: PASS

`shared-agent-executor` remains independently usable; its new guard only changes the status of an empty run from `success` to `failed` with the same `InvocationResult` shape, plus an additive finish-reason field on success results. `review-dispatch` already handles failed runs; no mandatory dependency introduced.

### III. Observable Quality

**Assessment**: PASS

Every incomplete reviewer now produces machine-parseable evidence: `status: failed` or a forced non-approving `workflow_verdict`, `error.code: child_empty_response`, `run_counts.failed`, and a warning in the dispatch artifact/council report. The previous silent `APPROVE` is replaced by observable failure/retry handling. Provenance includes model, finish reason, and token usage.

### IV. Testability

**Assessment**: PASS

Guard is isolated and testable: executor unit tests for empty text and `finish: length`; dispatch golden tests for both submission branches (failed reclassification and forced non-approving verdict), no-APPROVE outcomes, failed count, warning. No external service required; existing coverage ratchet enforced.

### V. Security by Default

**Assessment**: PASS

Input validation: treats untrusted model output (empty text, length termination) as a failure case rather than trusted approval. No new permission, no secrets, no supply-chain change.

### VI. Deterministic Execution

**Assessment**: PASS

Incomplete-response detection is a deterministic procedure (trimmed text length == 0, finish reason == `length`, submissions-presence branch) executed in a tested tool (`agent-executor.ts`, dispatch finalization), not re-reasoned from prose. No new prescribed output templates; existing council report rendering preserved.
