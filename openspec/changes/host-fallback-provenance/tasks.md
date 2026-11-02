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
      mirroring the advisor host-model-fallback block (L801-922,
      including the cancelled-case assertions at L918-920):
      (a) success — dispatch `{ agent, prompt, source: "host" }`
      with the default `matrixFixture` (whose `standard` profile
      carries `provider/standard`, proving profile bypass per
      RD-FR-037 scenario "Host run ignores a configured profile
      model") and assert `result.provenance` matches
      `{ requested_model: null, resolved_parent_model: "p/m",
      resolved_parent_variant: "active", requested_variant:
      null, model_mismatch: false }`
      plus that the child session is invoked with the resolved
      host model `p/m` and variant `active`, not the profile
      model `provider/standard` (the scenario's first THEN
      clause, verified at the execution seam);
      (b) `host_model_unavailable` with `failingClient` — drive
      the fixture error text with redactable material (a Bearer
      token, a URL query string, a home path) and assert
      `error.code` is `host_model_unavailable`, `retryable` is
      true, and the returned `result.error.message` is the
      `sanitizeInvocationError()` form (no Bearer value, no
      query string, no home path — exactly the guarantees the
      spec's sanitization clause scopes itself to; nothing is
      persisted on this pre-execution early exit); also assert
      the session's run-file listing is empty for the
      correlation;
      (c) `cancelled` with `hangingClient` + AbortController,
      asserting the run-file listing is likewise empty;
      (d) direct-bypass precedence — dispatch
      `{ agent, prompt, source: "host", model: "provider/bypass" }`
      and assert the child executes with the bypass model, the
      provenance records the requested model (non-host shape),
      and host resolution is never consulted (the client's
      `session.message` host-lookup spy is not called) — RD-FR-037
      scenario "Direct bypass overrides host resolution"; the
      scenario's finalization clause is pinned compositionally
      by finalize-error-boundaries.test.ts:133 ("rejects a host
      run that carries requested model provenance") — cite that
      pin in the test comment;
      (e) variant-less host result — fake host resolution
      returning a model with no active variant; assert
      `result.provenance.resolved_parent_variant` is null —
      RD-FR-037 scenario "Host run without an active variant".
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
      succeed with the native verdict and zero errors; run the
      pass for both an active-variant host run and a variant-less
      host run (`resolved_parent_variant: null`), both
      finalizing cleanly (RD-FR-037 scenarios "Host-fallback
      panel finalizes cleanly" and "Host run without an active
      variant" — the #901 reproduction at the unit seam).
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

- [ ] 3.1 Sync the edited live plugin byte-identical into the
      scaffold asset copy at
      `internal/scaffold/assets/opencode/plugins/review-dispatch/`
      (`index.ts`; design D4) and run the drift gate:
      `go test -race -count=1 ./internal/scaffold/...` —
      `TestEmbeddedAssets_MatchSource` MUST pass. Re-sync and
      re-run if any phase-2 test forced a live-plugin change.
      Note: the drift gate is intentionally red from task 1.1
      until this sync (the mirror is batched here to avoid
      re-sync churn); the phase-1/2 checkpoints gate the vitest
      suite only, per MP-018's per-suite reading.
- [ ] 3.2 Add the CHANGELOG.md entry for the behavior change
      (host-fallback runs execute on the parent session model
      and finalize cleanly; host resolution failure is now
      retryable `host_model_unavailable`; issue #901).
- [ ] 3.3 Documentation-gate assessment across CHANGELOG.md,
      AGENTS.md, README.md, and docs/: the changed panel outcome
      (host-fallback triage/review panels finalizing with their
      native verdicts; host runs executing on the parent session
      model with profile/override bypass; advisor lists as the
      sanctioned pinning mechanism) is user-facing configuration
      behavior. The org constitution's Cross-Repo Documentation
      rule therefore applies (its only exemptions are purely
      internal changes — refactoring, test-only, CI-only): the
      docs/configuration.md addition is tasked in 3.4, and the
      in-repo and unbound-force/website documentation issues are
      filed in 3.4/3.5 per the documentation-gate precedent
      (#778 ↔ website#316, #891 ↔ website#360,
      model-matrix-scaffold-deployment's website#362). Record
      the assessment in the PR body.
- [ ] 3.4 Add the host-fallback semantics note to the Review
      Council Configuration section of docs/configuration.md
      (parent-session model execution for host-source runs,
      profile/override bypass, advisor lists as the sanctioned
      pinning mechanism, and the failure semantics:
      `host_model_unavailable` on a pre-execution early exit is
      retryable in the caller-re-invocation sense only — no run
      file exists for DR-FR-002 redispatch to target) and file
      the in-repo documentation issue; link it in the PR body.
- [ ] 3.5 File the unbound-force/website documentation issue for
      the user-facing panel-outcome change per the Cross-Repo
      Documentation rule (precedent: website#316, website#360,
      website#362); link it in the PR body.
- [ ] 3.6 File the follow-up issue for splitting
      host-resolution failure classes (structural vs transient)
      across the advisor and host branches, scoped to also state
      that `retryable` on a pre-execution early exit means
      caller re-invocation only — no DR-FR-002 redispatch
      target (design D2 note); link it in the PR body.
- [ ] 3.7 Constitution-alignment verification (proposal
      identifies all six principles): confirm the tests run
      without external services (IV — vitest doubles only), the
      run-file provenance assertions verify observable side
      effects (III), the fix reuses the shared
      `resolveHostModel()` with no new coupling (II), and
      host-fallback panels finalize their artifact (I); confirm
      the failure mode is the sanitized retryable
      `host_model_unavailable` error (V) and the change is
      runtime dispatch code with no instruction-prose additions
      (VI). Record the verification summary in the PR body.
- [ ] 3.8 Final CI parity (MP-023): run `make plugin-test` and
      `go test -race -count=1 ./internal/scaffold/...` — both
      MUST pass before the change is marked complete.
