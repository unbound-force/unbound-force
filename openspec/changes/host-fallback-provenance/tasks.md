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

## 1. Host-source branch in the live dispatch core

- [ ] 1.1 Add the `source === "host"` model-resolution branch to
      `executeDispatchRun` in
      `.opencode/plugins/review-dispatch/index.ts` (design D1):
      after the `options.model` direct-bypass branch and before
      the tier-based `else` branch, call the shared
      `resolveHostModel()` with
      `{ client, sessionID: context.sessionID, messageID:
      context.messageID, directory: deps.directory }` and
      `context.abort`; set `selectedModel`/`selectedVariant` from
      the host result, `requestedModelLabel = null`, and populate
      the existing `hostResolvedParent` record. Failure paths
      (design D2): abort signal → early exit `cancelled`;
      resolution error → early exit `host_model_unavailable`
      (retryable, sanitized). Strict `=== "host"` comparison, no
      normalization. Update the function's doc comment
      ("Responsibilities, in order") and the
      `ExecuteDispatchRunOptions.model`/`tier` doc comments to
      state the host-source resolution and that the
      `tier ?? "standard"` default is unreachable for host
      entries.
- [ ] 1.2 Checkpoint: run `cd .opencode && npx vitest run` —
      the existing suite MUST stay green (no existing test pins
      host-source execution; the planner and finalizer tests are
      unaffected; the `tier_model_unavailable` test at
      dispatch-agent-run.test.ts L487 uses no `source` and stays
      valid). Fix any regression before proceeding.

## 2. Host-source test coverage (RD-FR-037 scenarios)

- [ ] 2.1 [P] In `.opencode/test/dispatch-agent-run.test.ts`,
      add a `describe("host-source model resolution")` block
      mirroring the advisor host-model-fallback block (L801-880):
      (a) success — dispatch `{ agent, prompt, source: "host" }`
      with the default `matrixFixture` (whose `standard` profile
      carries `provider/standard`, proving profile bypass per
      RD-FR-037 scenario "Host run ignores a configured profile
      model") and assert `result.provenance` matches
      `{ requested_model: null, resolved_parent_model: "p/m",
      resolved_parent_variant: "active", requested_variant: null }`;
      (b) `host_model_unavailable` with `failingClient`
      (retryable true, sanitized); (c) `cancelled` with
      `hangingClient` + AbortController.
- [ ] 2.2 [P] In `.opencode/test/dispatch-review-runs.test.ts`,
      add a test that dispatches a panel whose `plan_entries`
      include a `source: "host"`, `model: null`, `tier: null`
      include-entry (host-fallback shape from `baseRuns`) and
      asserts the dispatch succeeds and the persisted run file
      (via the session's run-file listing) carries host-run
      provenance: `requested_model`, `provider`, `model_id`,
      `variant` null; `resolved_parent_model` non-null
      (RD-FR-037 scenario "Host run inherits the parent model").
- [ ] 2.3 [P] In `.opencode/test/consolidate-dispatch.test.ts`,
      add a provenance-carrying consolidation test over a
      scratch dispatch session that includes a successful
      host-source run file with the dispatcher-recorded host
      provenance shape; assert consolidation + finalization
      succeed with the native verdict and zero errors
      (RD-FR-037 scenario "Host-fallback panel finalizes
      cleanly" — the #901 reproduction at the unit seam).
- [ ] 2.4 [P] In `.opencode/test/dispatch-recovery.test.ts`,
      add a test that re-dispatches a failed host-source run
      (persisted prompt + plan entry with `source: "host"`,
      null model/tier) through `redispatch_failed_runs` and
      asserts the replacement run resolves the host model and
      carries host-run provenance (RD-FR-037 scenario
      "Re-dispatched host run reuses host resolution").
- [ ] 2.5 Checkpoint: run `make plugin-test` — the full vitest
      suite with the 94 percent statements coverage floor MUST
      pass. Gatekeeping: if the floor fails, report and stop;
      never adjust `.opencode/vitest.config.ts`.

## 3. Scaffold mirror, CI parity, documentation

- [ ] 3.1 Sync the edited live plugin byte-identical into
      `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
      (design D4) and run the drift gate:
      `go test -race -count=1 ./internal/scaffold/...` —
      `TestEmbeddedAssets_MatchSource` MUST pass. Re-sync and
      re-run if any phase-2 test forced a live-plugin change.
- [ ] 3.2 [P] Add the CHANGELOG.md entry for the behavior change
      (host-fallback runs execute on the parent session model
      and finalize cleanly; host resolution failure is now
      retryable `host_model_unavailable`; issue #901).
- [ ] 3.3 [P] Documentation-gate assessment: the change is an
      internal bug fix restoring RD-FR-008-documented behavior
      with no API, schema, AGENTS.md, or README surface change —
      record the exemption rationale in the PR body (or file the
      documentation issue if review assesses the changed triage
      outcomes as user-facing documentation).
- [ ] 3.4 Constitution-alignment verification (proposal
      identifies all four principles): confirm the tests run
      without external services (IV — vitest doubles only), the
      run-file provenance assertions verify observable side
      effects (III), the fix reuses the shared
      `resolveHostModel()` with no new coupling (II), and
      host-fallback panels finalize their artifact (I). Record
      the verification summary in the PR body.
- [ ] 3.5 Final CI parity (MP-023): run `make plugin-test` and
      `go test -race -count=1 ./internal/scaffold/...` — both
      MUST pass before the change is marked complete.
