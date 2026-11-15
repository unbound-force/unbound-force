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

## 1. Planner changes in the live plugin (RD-FR-063, RD-FR-064)

- [x] 1.1 Thread internal profile-name provenance: add
      `profileName: string | null` to the `ResolvedRun`
      interface (`.opencode/plugins/review-dispatch/index.ts`,
      design D1) and populate it at the four resolution
      sites — `explicitRuns()` (named profile for the
      `profile` form, null for the `model` form),
      `baseRuns()`'s advisor branch (the selected profile;
      host branch null), `fullPanelRuns()` (`"standard"`),
      and `addAugmentation()` (the adjacent tier). The
      `PlanEntry` mapping in `planReviewDispatch` Pass 2
      MUST NOT emit the field; the strict plan entry schema
      and `plan_version` stay unchanged.
- [x] 1.2 Add the D2 validation to `explicitRuns()`'s
      profile branch: after `profilePair()` succeeds, a null
      model records a validation error naming the profile,
      stating that explicit-source runs cannot inherit the
      host model, and giving the remediation (configure
      `profiles.<name>.model` in `.uf/review-matrix.yaml` or
      the gitignored override). The existing machinery must
      flip the entry to `skip` and force plan
      `inconclusive`/`INCONCLUSIVE`.
- [x] 1.3 Add the D1 advisory emission to
      `planReviewDispatch`: during Pass 2, accumulate each
      `decision: "include"` run with `source: "advisor"` and
      `model: null` into a profile-keyed map (keyed by
      `profileName`); after Pass 2, emit one `LOW`-severity
      advisory per distinct profile — naming the profile,
      the affected agents in plan order, the host-model
      inheritance fallback (#693), and the remediation —
      appended to the existing advisory block and sorted by
      profile name. `status` MUST stay `ready` when no plan
      errors exist. Update the `planReviewDispatch` and
      `baseRuns`/`explicitRuns` doc comments to state both
      behaviors.
- [x] 1.4 Checkpoint: run `cd .opencode && npx vitest run` —
      the existing suite MUST stay green. In particular the
      `optional profile models` tests
      (`test/review-dispatch.test.ts`, model-null advisor
      entry and explicit-model advisor entry) and the
      `run resolution and fail-closed limits` block MUST
      pass unmodified. Fix any regression before proceeding.

## 2. Aggregate summary changes in the live plugin (RD-FR-065)

- [x] 2.1 Capture failure causes in the `dispatchReviewRuns`
      batch-result loop (design D3): the
      fulfilled-but-failed branch (`runResult.error !==
      null`) pushes `run failed: <agent>/<sequence>:
      <code>: <message>` into `warnings`, zipping the
      settled results with the batch array (allSettled
      preserves order) for the agent and sequence; the
      rejected-promise branch keeps the existing
      `run rejected:` warning. Maintain a first-cause
      variable set once from the first failure in batch
      order.
- [x] 2.2 Fold the first cause into the all-failed result
      (design D4): when `runs_failed > 0` and
      `runs_completed === 0`, the aggregate return sets
      `message` to the first run-failure cause, falling
      back to `warnings[0]` when no run cause exists;
      `dispatchSummaryLine` stays unchanged. Add the D5
      clause to the `dispatch_review_runs` tool
      description's error-summary sentence (all-failed
      dispatches report the first run-failure cause as
      their message). Do not set an aggregate `retryable`.
- [x] 2.3 Checkpoint: run `cd .opencode && npx vitest run` —
      the existing suite MUST stay green except assertions
      that pinned empty `warnings` on failed-run scenarios
      or an absent all-failed `message`; update those to
      the new shapes in the same commit. The
      `no failure message reported` pin (retryable-true
      case, empty warnings) MUST stay valid.

## 3. Test coverage (RD-FR-063 through RD-FR-065 scenarios)

- [x] 3.1 In `.opencode/test/review-dispatch.test.ts`, add a
      `describe("planner tier resolvability (RD-FR-063/064)")`
      block reusing the fixture-matrix pattern: (a) advisor
      run on a model-less profile — delete
      `profiles.standard.model`, plan a code review for an
      advisor-listed agent, assert the include-entry has
      `model: null`, `status` is `ready`, and exactly one
      advisory names `standard`, the agent, the host-model
      inheritance fallback, and the remediation; (b)
      full-panel variant — one advisory for `standard`
      naming the review-capable agents, not per-entry; (c)
      augmentation variant — opt-in augmentation with a
      model-less adjacent profile; (d) negatives — a
      model-configured matrix emits no advisory, and a
      host-fallback entry emits none; (e) determinism —
      two identical plans produce identical advisory
      arrays; (f) D2 fail-closed —
      `runs.code.divisor-architect: [{ profile: heavy }]`
      with `profiles.heavy` model-less asserts
      `inconclusive`/`INCONCLUSIVE`, entry `decision:
      "skip"` with non-empty `validation_errors`, and plan
      errors naming the agent and `heavy`; (g) D2
      negatives — direct-model explicit run unaffected,
      advisor entry on the same model-less profile stays
      includable per RD-FR-063.
- [x] 3.2 [P] In `.opencode/test/dispatch-review-runs.test.ts`,
      add a `describe("all-failed summary carries causes
      (RD-FR-065)")` block driving the real tool: (a) a
      plan whose every run fails pre-execution (advisor
      entry with host resolution failing through the fake
      client) asserting `status: "error"`,
      `runs_completed: 0`, `message` carrying the error
      code, and the summary line ending with the cause
      rather than `no failure message reported`; (b) the
      `run failed:` warning naming agent, sequence, code,
      and message, plus a `run rejected:` entry for a
      thrown batch promise; (c) partial failure keeps
      `status: "ok"` with no `message` while the cause
      stays in `warnings`; (d) the 512-character bound and
      line collapsing via `dispatchSummaryLine` on a
      hand-built long-message result; (e) first-cause
      determinism across two identical invocations.
- [x] 3.3 Checkpoint: run `make plugin-test` — the full
      vitest suite with the 94 percent statements coverage
      floor MUST pass. Gatekeeping: if the floor fails,
      report and stop; never adjust
      `.opencode/vitest.config.ts`.

## 4. Scaffold mirror, CI parity, documentation

- [x] 4.1 Sync the edited live plugin byte-identical into the
      scaffold asset mirror
      `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
      and run the drift gate:
      `go test -race -count=1 ./internal/scaffold/...` —
      `TestEmbeddedAssets_MatchSource` MUST pass. Re-sync
      and re-run if any phase-3 test forced a live-plugin
      change. No `.opencode/commands/`, `.opencode/skills/`,
      or `openspec/` files change in this change, so no
      other mirrors are affected. Note: the drift gate is
      intentionally red from task 1.1 until this sync; the
      phase-1/2/3 checkpoints gate the vitest suite only.
- [x] 4.2 [P] Add the CHANGELOG.md entry under Unreleased
      with the `(Spec:
      openspec/changes/dispatch-tier-resolvability/,
      Fixes: #1009)` tail: the planner now signals advisor
      host-model inheritance with a per-profile advisory,
      fails closed on explicit profile runs against
      model-less profiles, and the all-failed
      `dispatch_review_runs` summary carries the first
      run-failure cause instead of `no failure message
      reported`.
- [x] 4.3 [P] Documentation-gate assessment across
      `CHANGELOG.md`, `AGENTS.md`, `README.md`, and
      `docs/`: the advisory and the all-failed summary
      cause are user-facing (calling agents and operators
      consume them), so the Cross-Repo Documentation rule
      applies. File the in-repo documentation issue and
      the unbound-force/website documentation issue
      (precedent: #891 ↔ website#360, #913 ↔ website#368),
      and link both in the PR body. Record the assessment
      in the PR body.
- [x] 4.4 Constitution-alignment verification (proposal
      identifies all six principles): confirm the tests run
      without external services (IV — vitest doubles only),
      the assertions verify observable side effects — plan
      advisories, warnings, message, summary line (III),
      the change reuses the existing advisory, validation,
      warnings, and message channels with no new dependency
      (II), the signals are self-describing artifacts the
      caller consumes without hero coupling (I), the
      captured causes are already-sanitized error text and
      repository configuration names (V), and the change
      is runtime dispatch code with no instruction-prose
      additions (VI). Record the verification summary in
      the PR body.
- [x] 4.5 Final CI parity (MP-023): run `make plugin-test`
      and `go test -race -count=1 ./internal/scaffold/...`
      — both MUST pass before the change is marked
      complete.
