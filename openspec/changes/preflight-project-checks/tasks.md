## 1. Declared-check library and tool

- [x] 1.1 Create `.opencode/lib/preflight-declared.ts`:
  - the zod v1 schema, which is strict, has unique names, and
    validates `run` as a path (D3);
  - `parsePreflightConfig` with `PreflightConfigError`;
  - `isInScope`, `findOnPath`, `computeDiffScope` (Phase 2a
    fallbacks and the merge-base), `tailOutput` and
    `renderDeclaredRows`;
  - `runDeclaredChecks(cwd, deps, mode)` with the injected
    dependencies (D4) and the soft-gate worktree baseline (D5).
- [x] 1.2 In `.opencode/plugins/uf-workflow/index.ts`:
  - add `runProcessNoShell` (`execFile`, `shell: false`,
    timeout reported as exit 124, 64 MB buffer);
  - add `createPreflightDeclaredDependencies` with an eager
    YAML parser (D8);
  - add `createPreflightDeclaredChecksTool`, whose only
    argument is `mode` and whose error codes are
    `INVALID_PREFLIGHT_CONFIG` and `PREFLIGHT_DECLARED_FAILED`;
  - register it as `uf_preflight_declared_checks`.
- [x] 1.3 [P] Unit and mode tests in
  `.opencode/test/preflight-declared.test.ts`: validation cases,
  pure helpers, execution chain, determinism, env passthrough,
  and the hard-gate/ci-aware/soft-gate verdicts including
  worktree-failure cleanup.
- [x] 1.4 [P] Round-trip test in
  `.opencode/test/preflight-declared-roundtrip.test.ts` with a
  real filesystem and `execFile`: literal argv, a non-zero exit,
  a non-executable `run`, and an absent config.
- [x] 1.5 [P] Registration and golden seams: add the tool name to
  `uf-workflow.test.ts` and `scratch-smoke.test.ts`, pin its
  args to exactly `["mode"]`, and add the success and
  `INVALID_PREFLIGHT_CONFIG` cases to
  `golden-output-shape.test.ts`.

## 2. Skill and scaffold

- [x] 2.1 In `.opencode/skills/pre-flight/SKILL.md`:
  - add Phase 2b (a single tool call, one output example, and
    the handling rules);
  - add the Phase 2a and Phase 5 one-liners;
  - point the Phase 2 "no tools" exit at Phase 2b;
  - record the line delta (551 -> 589).
- [x] 2.2 Add `.uf/preflight.yaml` and
  `internal/scaffold/assets/uf/preflight.yaml`
  (byte-identical; a commented template with `checks: []`).
- [x] 2.3 In `internal/scaffold/scaffold.go`, make
  `uf/preflight.yaml` project-owned in `isToolOwned` with an
  extension-point comment, and add
  `opencode/lib/preflight-declared.ts` to `isToolOwned` and
  `isDivisorAsset`.
- [x] 2.4 Add the lib to `stageReviewPluginSources` in
  `internal/scaffold/plugin_activation.go`.
- [x] 2.5 Mirror the lib, plugin and skill into
  `internal/scaffold/assets/opencode/`.
- [x] 2.6 Update `internal/scaffold/scaffold_test.go`:
  - expected asset paths, ownership, mapping and Divisor-subset
    tables;
  - add `TestRun_PreflightChecksPreservedOnRerun` (plain re-run
    skips the file; `--force` resets it).
- [x] 2.7 Update the deployed-file count range in
  `cmd/unbound-force/main_test.go` to 82-86.

## 3. Documentation

- [x] 3.1 [P] Add a `CHANGELOG.md` Unreleased -> Added entry.
- [x] 3.2 [P] Add a `.uf/preflight.yaml` row to
  `docs/configuration.md`.
- [x] 3.3 [P] File the documentation issues required by the
  documentation gates: one for this repository, and one
  cross-repo issue against `unbound-force/website` for the new
  extension point. Record them in the PR description.
  - Filed: unbound-force/unbound-force#926 (docs) and
    unbound-force/website#373 (docs, tutorial). Both numbers to
    be recorded in the PR description.

## 4. Validation and verification

- [x] 4.1 Run `go test ./internal/scaffold/... ./cmd/...` and
  confirm it passes. This covers ownership, drift pairs, the
  re-run regression and the count range.
- [x] 4.2 Run the plugin test suite (`.opencode/test`) and
  confirm it passes.
- [x] 4.3 Run `make check` for CI parity.
  - Components run individually: `go build ./...` PASS;
    `go vet ./...` PASS; `go test -race -count=1` (all 13
    packages) PASS; plugin suite 1726 tests PASS at 95.88%
    statements (floor 94%); `coverage-gate` against
    `upstream/main` PASS (1 changed package within thresholds).
  - Documented discrepancy: bare `make check` aborts at lint
    because `golangci-lint` reports a pre-existing errcheck in
    `internal/gateway/gateway_test.go:4383` — reproduced
    byte-identically in a worktree at the merge base
    (e64a6564), so it is unrelated to this branch.
- [x] 4.4 End-to-end: in a repository whose `.uf/preflight.yaml`
  declares a required check that fails on the branch, run
  `/uf.review-council` and confirm that the declared rows appear
  in `pre_flight_results` exactly as rendered and that the
  soft-gate classification matches the merge-base result.
  - Verified via `uf_preflight_declared_checks` soft-gate runs:
    a check failing on branch and at the merge base classified
    `pre-existing` (verdict PASS); a check failing only on the
    branch (file absent at merge base) classified
    `branch-caused` (verdict FAIL). Baseline method `worktree`,
    merge base e64a6564. Declared rows rendered exactly as
    returned. Scratch config reverted afterward.
- [x] 4.5 Verify the proposal's constitution alignment:
  - II: an absent config is a no-op;
  - IV: no external services in tests;
  - V: no shell, path validation, invalid config fails closed,
    and no existing gate is weakened;
  - VI: the skill only calls the tool and inserts `rendered`.
  - Verified PASS on all four tenets: absent config returns
    `config_present: false` with verdict PASS; tests use scratch
    directories and local processes only; `runProcessNoShell`
    executes via `execFile` (shell disabled) with zod-validated
    repo-relative paths and fail-closed config errors; no gate
    values (coverage floor, severity, CI flags) modified; the
    skill's Phase 2b inserts the tool-rendered rows verbatim.

## 5. Review-council follow-ups

- [x] 5.1 Classify execFile output-limit kills by their real code
  (`ERR_CHILD_PROCESS_STDIO_MAXBUFFER` under Node 24 and Bun 1.4;
  `ENOBUFS` is never reported) in the exported
  `processResultFromExecFile`, annotate the output, and pin it
  with a table test of observed error shapes and a real flood
  round trip through `runProcessNoShell` with a small buffer.
- [x] 5.2 [P] Render mode-specific columns in
  `renderDeclaredRows` (causality only in `soft-gate`) and drop
  the skill's column post-processing instruction (DE-003).
- [x] 5.3 [P] Fence the command cell with a backtick run longer
  than any inside it, so backticks cannot end the code span.
- [x] 5.4 [P] Keep the real fallback reason when soft-gate
  recomputes the scope for a config without `paths`.
- [x] 5.5 Add real-git soft-gate round trips (pre-existing and
  branch-caused, with the script and config in a git-excluded
  `.uf/` tree, and worktree cleanup), and assert verdict,
  baseline method and causality on the non-repo failing case.
- [x] 5.6 [P] Parse real YAML: the shipped template through the
  subset helper (which gains `[]`), and commented block-list YAML
  through Bun.YAML when `bun` is on PATH (skipped otherwise; CI
  plugin tests run on Node only).
- [x] 5.7 [P] Assert `UF_PREFLIGHT_PROJECT_ROOT` on baseline
  re-runs, cover merge-base == HEAD in `computeDiffScope`, and
  `exec` the sleeper in the timeout round trip so no orphan
  survives.
