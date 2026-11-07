---
tier: validated
compiled_at: 2026-10-25T16:37:07Z
compiled_by: rits/zai-org/glm-5-3
sources:
  - ci-1
  - pre-flight-skill-20260623T123057-yvonne-devlin
  - pre-flight-skill-20260623T123105-yvonne-devlin
  - pre-flight-skill-20260623T130130-yvonne-devlin
  - ci-causality-soft-gate-20260701T100619-yvonne-devlin
  - ci-causality-soft-gate-20260701T100625-yvonne-devlin
  - ci-parity-20261014T123137-jay-flowers
  - gh-cli-20261015T234427-jay-flowers
  - ci-parity-20261019T022441-jay-flowers
  - ci-parity-20261019T111110-jay-flowers
topic: ci-coverage
---

# CI Coverage

## Current State (October 2026)

### Scope-based Go coverage gate

- `make check` = `lint` + `test` + `build` + `coverage-gate` + `plugin-test`; both gates mandatory — Go 
  `coverage-gate` AND TypeScript `plugin-test` (94 percent statements floor).
- `make coverage-gate` runs `cmd/coverage-gate` (logic in `internal/coveragegate`) with the `coverage-gate.json` 
  manifest against `coverage.out`, produced by `go test -race -count=1 -coverprofile` — tests always run `-race 
  -count=1`; CI enforces this.
- It checks ONLY packages changed in the diff, each mapped to exactly one scope with its own threshold: `helper` (80) 
  — I/O coordination, health checks, scaffolding (both `cmd/*` packages and 
  `internal/{scaffold,doctor,artifacts,schemas,config,gateway,homebrew,sandbox,setup}`);
  pure scopes (90) — `artifact`, `schema`, `config` (`internal/coveragegate`), `mapping` (`internal/textutil`), `path`.
- The gate is additive — it complements, never weakens, the global 80 percent and 90 percent-internal ratchets in 
  `ci_local.yml`. It fails a changed package that is below its scope threshold, missing from the manifest, multiply 
  matched across scopes, or lacking measured coverage. New `internal/` packages MUST be placed in exactly one scope.

### CRAP baseline model & cross-platform causality

- CI CRAP regressions are caused by cross-platform coverage divergence, NOT floating-point precision: IEEE 754 makes 
  the CRAP formula bit-identical on ARM64 and x86_64, but the coverage percentages fed into it differ because 
  platform-conditional code paths (macOS vs Linux branches, `LookPath` tool availability, SELinux detection) exercise 
  different lines per platform.
- Current model (main): a manually committed `.gaze/baseline.json`, generated locally by `make crapload-baseline` 
  (`gaze crap --format=json` plus jq path normalization), compared in CI by `ci_crapload.yml` via the pinned 
  complytime/org-infra `reusable_crapload_analysis.yml` workflow; `make crapload-check` does the local comparison. 
  Since the baseline is dev-machine-generated, platform divergence versus the CI runner is the first suspect for 
  apparent regressions.
- The epsilon tolerance (0.5) and CI-runner baseline regeneration exist ONLY on the unmerged 
  `origin/036-ci-crap-baseline` branch — main has neither.
- `gaze crap --test-short` now defaults to FALSE (opt-in); only gaze's own self-check defaults to `-short`. 
  Short-guarded tests contribute zero coverage when skipped.

### Pre-flight baselining

- The pre-flight skill (`.opencode/skills/pre-flight`) detects CI tools and runs them locally in three modes, one 
  unambiguous behavior each: `hard-gate` — run all, stop on first failure (`/uf.unleash` phase checkpoints only); 
  `ci-aware` — build a CI coverage matrix against PR check results, skip CI-verified tools, run the rest 
  (`/uf.review-pr`); `soft-gate` — run all, classify failures branch-caused vs pre-existing, gate only on branch-caused 
  (`/uf.review-council`; `/uf.unleash` inherits soft-gate for free via delegation to review-council).
- Two-tier baseline (soft-gate, only when needed; method recorded): Tier 1 CI API — `gh api 
  .../commits/${DEFAULT_BRANCH}/check-runs` (fast, reuses CI-computed data); Tier 2 fallback — `git worktree add 
  /tmp/preflight-baseline-${SHORT_SHA} main --detach` for environments without `gh` or CI.
- Only failing tools need baseline comparison — passing tools are not branch-caused by definition. Unknown 
  classifications conservatively count as branch-caused (no false negatives, occasional false positives; matches 
  `/review-pr`).

### CI parity discipline

- Rule: replicate CI checks locally before marking tasks complete; derive commands from `.github/workflows/` — make 
  targets alone under-represent what CI checks.
- `make lint` is `go vet` + `golangci-lint` — Go-only, no Markdown. MegaLinter (`ci_checks.yml`) covers Go, Actions, 
  Bash, Gitleaks; markdownlint is DISABLED in `.mega-linter.yml` — CI does not lint Markdown here, so a local 
  markdownlint pass is deliberately beyond-CI style consistency, not CI parity.
- Derive parity checklists from the changed-file list mapped to workflows, not the change's primary language: 
  manifest touches (`package.json`, `go.mod`) must cover `ci_security.yml` (OSV-Scanner via
  `reusable_vuln_scan` plus the plugin transitive throwaway-lock job) and `ci_dependencies.yml` (dependency review); only `.go` changes need build/coverage-gate; PR 
  titles face commitlint. Every CI surface appears in the parity note as covered by a local task or explicitly excluded 
  with rationale, verified against `.mega-linter.yml` and the workflow files first.
- Operational: file documentation issues via `gh issue create --body-file <tmpfile>` — `--body "$(cat <<'EOF' ...)"` 
  fails "bad substitution" on backticks.

## History

- **2026-05 (ci-1)**: Diagnosed CI CRAP regressions as cross-platform coverage divergence, ruling out float precision 
  (IEEE 754 bit-identical for the CRAP formula). Proposed fixes: generate the baseline ON the CI runner via a 
  post-merge workflow, plus epsilon = 0.5 regression tolerance (a local evaluate job in `ci_crapload.yml`, since 
  complytime/org-infra has no regression-epsilon input). Neither merged to main — they live only on 
  `origin/036-ci-crap-baseline`.
- **2026-06 (pre-flight skill extraction)**: Consolidated duplicated CI-detection logic from review-council 
  (workflow-driven) and review-pr (config-file-driven) into the shared skill — intentional behavioral expansion, not 
  strict parity (each consumer gains the other's detection: review-council finds tools like yamllint; review-pr finds 
  CI-only commands like coverage ratchets), stated explicitly so newly discovered tools are not treated as regressions. 
  Scaffold lessons: skills distributed via `uf init` need an asset copy at 
  `internal/scaffold/assets/opencode/skills/<name>/SKILL.md` plus an `expectedAssetPaths` entry 
  (`TestEmbeddedAssets_MatchSource` enforces drift); `TestRunInit_FreshDir` in `cmd/unbound-force` hardcodes the `uf 
  init` file count — adding scaffold assets bumps it, so run the full `make check`, not just scaffold package tests.
- **2026-07 (soft-gate addition)**: Added soft-gate as a third mode rather than extending hard-gate with optional 
  baseline parameters, preserving one-unambiguous-behavior-per-mode. Consumer inheritance came free via delegation 
  (/unleash to review-council); only /unleash phase checkpoints stayed hard-gate. Introduced the two-tier baseline (CI 
  API first, worktree fallback) with the only-failing-tools optimization.
- **gaze -short default flip**: `gaze crap --test-short` changed from default-on to opt-in (default FALSE); only 
  gaze's self-check still defaults `-short`. Motivated by short-guarded tests contributing zero coverage when skipped, 
  distorting CRAP and coverage measurement.
- **2026-10 (CI parity audits)**: Learnings tightened parity practice: `make lint` is Go-only and make targets 
  under-represent CI (10-14); checklists must map changed files to ALL workflows — ci_local, ci_checks, plus 
  security/dependency for manifest touches — every surface covered or explicitly excluded (10-19); every claimed CI 
  surface verified against `.mega-linter.yml` and workflow files before writing it into a parity note (10-19).
