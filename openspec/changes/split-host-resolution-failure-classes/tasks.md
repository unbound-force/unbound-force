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

## 1. Shared classification helper in the live dispatch core

- [ ] 1.1 Add `StructuralHostModelError` to the
      `../../lib/agent-executor.js` import block in
      `.opencode/plugins/review-dispatch/index.ts`, and add
      one unexported module-level helper next to
      `executeDispatchRun` (design D1):
      `hostResolutionFailureResult(error, aborted,
      provenance): InvokeAgentResult`. It returns
      `failedResult(provenance, "cancelled", "dispatch
      cancelled", false)` when `aborted`; otherwise
      `error instanceof StructuralHostModelError` returns
      `failedResult(provenance, "host_model_structural",
      `${error.message}. Remediation: ${error.remediation};
      retrying will not fix this.`, false)`; every other
      error returns `failedResult(provenance,
      "host_model_unavailable", errorText(error), true)`.
      Rely on `failedResult()`'s single
      `sanitizeInvocationError()` pass; drop the now
      redundant explicit `sanitizeInvocationError(...)`
      wrapper at this seam. Replace BOTH catch blocks with
      `return earlyExit(hostResolutionFailureResult(error,
      context.abort.aborted, earlyProvenance(...)))`,
      preserving each branch's existing provenance argument
      (host: no tier; advisor: the tier label). Update the
      `executeDispatchRun` doc comment
      ("Responsibilities, in order") and the
      `ExecuteDispatchRunOptions.model`/`tier` doc comments
      to state the structural/transient split and the
      caller-re-invocation-only meaning of `retryable` on
      these early exits. No planner, finalizer, schema, or
      run-file shape change.
- [ ] 1.2 Checkpoint: run `cd .opencode && npx vitest run` —
      the existing suite MUST stay green. In particular, the
      transient assertions at
      `test/dispatch-agent-run.test.ts` L836-875 (advisor)
      and L1003-1035 (host) MUST still see
      `host_model_unavailable` / `retryable: true`, and the
      cancelled assertions at L877-921 and L1037-1062 MUST
      still see `cancelled`. Fix any regression before
      proceeding.

## 2. Test coverage (RD-FR-038 scenarios)

- [ ] 2.1 [P] In `.opencode/test/dispatch-agent-run.test.ts`,
      add a `describe("host-resolution failure
      classification (RD-FR-038)")` block reusing the
      `hostLookupClient({ message })` shaper (L932) and the
      empty `standard` profile matrix (L801-834):
      (a) host-branch structural — `{ agent, prompt,
      source: "host" }` with a parent message whose role is
      not `assistant`; assert `error.code` is
      `host_model_structural`, `retryable` false, the
      message contains `Remediation: ` and `retrying will
      not fix this.`, and `listRunFiles(correlation_id)` is
      empty;
      (b) advisor-branch structural — `{ agent, prompt,
      tier: "standard", source: "advisor" }` with the empty
      profile and a parent assistant message missing
      `providerID`/`modelID`; assert the same class,
      proving the split is symmetric by construction;
      (c) invalid variant — parent assistant message with a
      present-but-invalid `variant`; assert
      `host_model_structural` and the invalid-variant
      remediation text;
      (d) transient unchanged — a `message` override that
      throws a plain `Error` carrying a Bearer value, a URL
      query string, and a home path; assert
      `host_model_unavailable`, `retryable: true`, the
      sanitized form (no Bearer value, no query string, no
      home path), no remediation suffix, and an empty
      run-file listing;
      (e) abort precedence — a `message` override that
      aborts the controller, backed by an error whose shape
      would otherwise be structural; assert `cancelled` and
      `retryable: false`.
- [ ] 2.2 [P] In `.opencode/test/dispatch-recovery.test.ts`,
      add a test in the pattern of the early-exit test at
      L1207-1264: seed a failed host-source run
      (`source: "host"`, null model/tier, persisted prompt),
      make host resolution throw `StructuralHostModelError`
      on the re-run, and assert the result entry's
      `outcome` is `early_exit`, `error.code` is
      `host_model_structural`, `error.retryable` is false,
      the message carries the remediation string, the prior
      run file is restored byte-identically to its active
      path, and `superseded/` is empty.
- [ ] 2.3 [P] In `.opencode/test/consolidate-dispatch.test.ts`,
      add a test for a dispatch session whose included plan
      entry early-exited structurally (no run file): assert
      consolidation fails closed with an error containing
      `included plan entry <agent>/<sequence> has no run`
      and `operation_verdict` / final verdict `INCONCLUSIVE`
      (RD-FR-038 "Early exit is not a redispatch target").
- [ ] 2.4 Checkpoint: run `make plugin-test` — the full
      vitest suite with the 94 percent statements coverage
      floor MUST pass. Gatekeeping: if the floor fails,
      report and stop; never adjust
      `.opencode/vitest.config.ts`.

## 3. Scaffold mirror, CI parity, documentation

- [ ] 3.1 Sync the edited live plugin byte-identical into
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  (design D9) and run the drift gate:
      `go test -race -count=1 ./internal/scaffold/...` —
      `TestEmbeddedAssets_MatchSource` MUST pass. Re-sync
      and re-run if any phase-2 test forced a live-plugin
      change. Note: the drift gate is intentionally red from
      task 1.1 until this sync; the phase-1/2 checkpoints
      gate the vitest suite only.
- [ ] 3.2 [P] Add the CHANGELOG.md entry for the behavior
      change: structural host-resolution failures now
      surface non-retryable `host_model_structural` with the
      remediation string preserved, while transient failures
      keep retryable `host_model_unavailable`; `retryable`
      on a pre-execution early exit means caller
      re-invocation only (issue #914, lineage #901/#693).
- [ ] 3.3 [P] Update the "Host-fallback model resolution"
      section of `docs/configuration.md` (L296-312) to
      document both failure classes — `host_model_structural`
      (non-retryable, remediation-carrying) and
      `host_model_unavailable` (retryable) — and restate
      that `retryable` on a pre-execution early exit means
      caller re-invocation only, with no DR-FR-002 redispatch
      target and fail-closed consolidation ("included plan
      entry has no run" → `INCONCLUSIVE`).
- [ ] 3.4 Documentation-gate assessment across
      `CHANGELOG.md`, `AGENTS.md`, `README.md`, and `docs/`:
      the failure-class split is user-facing (operators and
      calling agents consume the error class), so the
      Cross-Repo Documentation rule applies. File the
      in-repo documentation issue and the
      unbound-force/website documentation issue (precedent:
      #778 ↔ website#316, #891 ↔ website#360), and link
      both in the PR body. Record the assessment in the PR
      body.
- [ ] 3.5 Constitution-alignment verification (proposal
      identifies all six principles): confirm the tests run
      without external services (IV — vitest doubles only),
      the tool-result assertions verify observable side
      effects (III), the change reuses the shared
      `StructuralHostModelError` with no new dependency
      (II), both dispatch branches classify identically
      through one helper (I), the remediation-carrying
      message passes through `sanitizeInvocationError()`
      (V), and the change is runtime dispatch code with no
      instruction-prose additions (VI). Record the
      verification summary in the PR body.
- [ ] 3.6 Final CI parity (MP-023): run `make plugin-test`
      and `go test -race -count=1
      ./internal/scaffold/...` — both MUST pass before the
      change is marked complete.
