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

## 1. Implement incomplete-response guard (executor + dispatch core)

- [x] 1.1 Guard `agent-executor.ts` (574→606) — return `failed` `child_empty_response` retryable when `text.trim().length === 0` (including length-terminated empty), preserving `usage`/message via `failedResult`/`MAX_ERROR_LENGTH`; surface the finish reason from `step-finish` parts as a structured field on `success` results (`null` when unreported).
- [x] 1.2 Harden `review-dispatch/index.ts` finalization (2210/5085) — re-derive the incomplete rule: `success` + incomplete (empty text or finish reason `length`) + zero accepted submissions → reclassify `failed` `child_empty_response` retryable with `workflow_verdict: null`, excluded from `successfulPersonaVotes`, counted in `run_counts.failed`; incomplete + accepted submissions → keep/restore `success` (RD-FR-026 precedence) with `workflow_verdict` forced to the kind-native non-approving value (`REQUEST CHANGES` / `NEEDS-CLARIFICATION` / `AUTHOR-DECIDES`); post-submission aborts keep RD-FR-026's rule unchanged.
- [x] 1.3 Emit `child_empty_response` warning (4482/4919) — exact kind string, agent/run id, finish reason and token counts when available; suppress `empty_coverage_attestation` for the same run when its final text is empty; no duplicate `child_empty_response` warnings per run.
- [x] 1.4 Mirror both edits byte-identically under `internal/scaffold/assets/opencode/lib/agent-executor.ts` and `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`.

## 2. Tests, harness, and documentation

- [x] 2.1 [P] Executor unit tests — empty string → `child_empty_response` failed; whitespace-only → failed; non-empty → success; length finish + empty → failed; length finish + non-empty → success carrying the structured finish-reason field (`null` when unreported); `retryable: true`, usage preserved.
- [x] 2.2 [P] Dispatch tests — incomplete + zero submissions (empty or truncated text) → `failed`, `workflow_verdict: null`, excluded from `successfulPersonaVotes`, `run_counts.failed`; incomplete + accepted submissions → success + forced kind-native non-approving verdict with submissions contributing (including executor-failed `child_empty_response` restored to success); RD-FR-026 abort-with-submissions case unchanged (verdict from findings); council with one incomplete + one real reviewer → non-`APPROVE`; warning exact kind, no `empty_coverage_attestation` duplication when text empty; no automatic re-execution within one dispatch call; all-failed exhaustion → non-approving per RD-FR-007 zero-success path.
- [x] 2.3 Eval harness non-approval check — session with `finish:length` or empty output MUST NOT yield an approving vote: zero submissions → `status != success && workflow_verdict == null`; accepted submissions → success with non-approving verdict; reproduction of #1052's 4/8 empty case fails before fix, passes after.
- [x] 2.4 [P] Documentation — CHANGELOG.md entry for the behavioral change (empty/length-terminated reviewer runs no longer approve); confirm AGENTS.md and README.md need no change (no new commands, gates, or user-facing workflow).

## 3. Verification and governance

- [x] 3.1 Verify constitution alignment end-to-end — autonomous artifacts, composability (shared `agent-executor.ts` independently testable), observable quality (machine-parseable `child_empty_response` + warning + forced non-approving verdict), testability (isolated unit tests), security/determinism (input validation, deterministic procedure); confirm `TestEmbeddedAssets_MatchSource` byte-identity, 94% statements floor held, no gate weakened.
