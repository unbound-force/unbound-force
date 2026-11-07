## Context

The unleash-tier-routing change recorded the TypeScript plugin suite's
statements coverage baselines (94.61% at round-1 verification, tasks.md
9.6; 94.58% at the round-2 fresh measurement, tasks.md 10.7, which
stood as the final recorded value through the round-3 verification at
11.8) but deferred configuring a coverage gate — the AGENTS.md
gatekeeping rule classifies threshold changes as governance decisions.
The deferral is tracked by unbound-force/unbound-force#780.

Since then, the review-council-multi-model-fanout change (complete,
unarchived) landed the gate machinery itself: commit 7cb5d8e created
`.opencode/vitest.config.ts` with `coverage.thresholds` set to 90
statements / 85 branches (fanout design D14). The enforcement chain
already exists end to end:

- The thresholds live in `.opencode/vitest.config.ts` (v8 provider,
  suite-global).
- `make plugin-test` runs `cd .opencode && npx -y npm@11.6.2
  install --ignore-scripts --no-package-lock --no-audit
  --no-fund && npx vitest run --coverage` (the lockless install
  reconciled by `stop-shipping-plugin-lock`); a threshold
  breach exits non-zero.
- `make check` and the Local CI `plugin-tests` job (ci_local.yml)
  both invoke that target, so local and CI enforcement are the same
  command.

A fresh measurement on this change's branch point (2026-10-22,
`cd .opencode && npx vitest run --coverage`): 95.18% statements
(3521/3699), 87.89% branches (2041/2322), 98.78% functions (567/574),
95.39% lines (3375/3538). Both gate modes were verified on this
branch: the passing mode (suite green, thresholds met) and the
failing mode (CLI override `--coverage.thresholds.statements=96`
exits 1 with `ERROR: Coverage for statements (95.18%) does not meet
global threshold (96%)` — the failure names the metric, the measured
value, and the threshold).

The gap this change closes: the D14 floor (90) sits 5.18 points below
the measurement (95.18) — roughly 190 of 3699 statements could lose
coverage without any gate failing.

## Goals / Non-Goals

### Goals

- Raise the statements coverage floor in `.opencode/vitest.config.ts`
  from 90 to 94 — just below the lowest recorded baseline (94.58%) —
  cutting the silent-regression window from 5.18 to 1.18 points
  (roughly 190 to roughly 44 statements).
- Decide the threshold policy issue #780 deferred: static
  governance-raised floor versus auto-tightening ratchet (D1).
- Preserve gate provenance: the fanout D14 attribution for the
  branches floor, the baseline trail, and the #780 linkage (D6).
- Document the gate where agents read conventions (AGENTS.md) and
  where changes are recorded (CHANGELOG.md).

### Non-Goals

- Raising the branches threshold: it stays at 85 (fanout D14). The
  measured 87.89% is recorded here as reference for any follow-up.
- Per-file or per-glob thresholds (see D3).
- Ratchet or auto-tightening machinery, baseline write-back paths, or
  new CI jobs (see D1, D5).
- Changes to the Go-side coverage gate, CRAP-load monitoring, or any
  CI workflow file.
- New plugin code or test files: the gate is configuration; its
  verification is executable commands (see tasks.md).

## Decisions

### D1: Threshold policy — static governance-raised floor

The floor is a checked-in integer that changes only through a
spec-workflow governance change with documented rationale, mirroring
how the Go-side gates (`coverage-gate.json` scopes, the global 80
check) are managed. This is the decision issue #780 deferred.

Rejected alternatives:

- **Auto-tightening write-back ratchet** (CI records the newly
  measured value as the threshold on green runs): the gate would
  modify itself without a governance decision. The AGENTS.md
  gatekeeping rule protects threshold values precisely so that gate
  changes carry review; a self-raising gate is still an unreviewed
  gate change. It also requires a CI write path (bot commit or
  mutable artifact), expanding workflow permissions against the
  least-privilege CI discipline, and makes the enforced value
  run-history-dependent rather than a reviewed checked-in integer
  (Constitution VI, Deterministic Execution).
- **Compare-to-main ratchet** (CI measures main and requires the PR
  to meet or exceed it): failure then depends on two measurements,
  doubling the flakiness surface, and obscures failure localization
  (which commit regressed?). This repository's CRAP-load monitoring
  documented exactly this failure class — cross-platform coverage
  divergence between local and CI runners forced an epsilon
  tolerance in ci_crapload.yml. A ratchet pinned near the measured
  value inherits that fragility with no margin to absorb it.

Constitution tie-ins: Autonomous Collaboration (the policy is a
self-describing artifact any worker or CI run observes identically)
and Deterministic Execution (the enforced value is a reviewed
checked-in constant, not run history).

### D2: Floor value — 94, suite-global statements aggregate

- **Anchoring**: 94 sits just below the lowest recorded baseline
  (94.58%), satisfying the issue's "near the recorded baseline"
  scope, and 1.18 points below the current measurement — roughly 44
  of 3699 statements of working room. The silent-regression window
  shrinks by 77%.
- **Rejected values**:
  - 90 (status quo) leaves the 5.18-point / roughly 190-statement
    silent window that issue #780 tracks.
  - 94.58 (the exact baseline) leaves 0.60 points of headroom —
    sub-point precision invites false failures from cross-platform
    V8 instrumentation divergence (macOS local versus ubuntu-latest
    CI) and Node patch-level differences; the CRAP ratchet needed an
    epsilon of 0.5 for exactly this divergence class.
  - 95 leaves 0.18 points of headroom — one mid-sized uncovered
    error-path block flips the gate red.
- **Integer floor**: fractional thresholds imply precision the
  measurement does not have across platforms.

### D3: Suite-global aggregate, not per-file thresholds

Vitest supports per-glob thresholds; rejected. The measured per-file
spread today includes `lib/gathering-evidence.ts` at 88.54% and
`lib/review-diff.ts` at 90.84% statements — a 94 per-file floor
fails immediately and would force exemption machinery or test churn
unrelated to the gate's purpose. The Go-side precedent gates scoped
aggregates (`coverage-gate.json` maps package groups, not individual
files). An aggregate floor also absorbs file-level restructuring
(splits, merges, moves) without gate churn.

### D4: Native Vitest thresholds only — no custom enforcement script

Determinism pack DE-001: mechanical enforcement MUST run through the
executable tool. Vitest's `coverage.thresholds` is the tool-owned
mechanism already wired into `plugin-test`. A custom script (parsing
the coverage summary and comparing with shell arithmetic) would
re-import the documented ratchet guard-class failures — empty
profile files, silently empty computed values, division-by-zero on
renamed scopes — that the replicator's awk/bc ratchet needed three
dedicated guards against. The failure semantics were verified
empirically on this branch (exit 1; the error names the metric, the
measured value, and the threshold). No new dependencies are
introduced; the pinned-dependency policy is untouched.

### D5: No new CI wiring — the existing chain enforces by construction

`make plugin-test` is invoked by `make check` and by the
`plugin-tests` job in ci_local.yml, so raising the checked-in
threshold value enforces locally and in CI with zero workflow edits.
This keeps the CI convention pack's action-pinning surface untouched
and keeps local/CI parity exact (one command, one code path). The
issue's "wire it into CI" is satisfied by the chain that already
exists; only the value changes.

Constitution tie-in: Composability First — the plugin suite remains
independently runnable and gateable via its documented command
(`cd .opencode && npx vitest run --coverage`) without the rest of
the repository, and the repository without the suite is unaffected.

### D6: Provenance comment rewrite in vitest.config.ts

The config's comment currently attributes the 90/85 floors to fanout
D14. The rewrite preserves the D14 attribution for the branches
floor (85, unchanged) and records: the unleash-tier-routing
baselines (94.61% round-1, 94.58% round-2 final recorded), the
2026-10-22 measurement (95.18%), issue #780, and the governance
constraint (the statements floor changes only through a spec-workflow
governance change). Constitution tie-in: Observable Quality — the
gate is self-describing and auditable against recorded evidence.

A config-value drift test (a test file asserting
`thresholds.statements === 94`) was considered and rejected: it
duplicates the value without adding protection — an actor lowering
the threshold can lower the test — and every legitimate governance
raise would require synchronized edits. The AGENTS.md gatekeeping
rule plus Divisor review are the modification control.

### D7: Documentation — CHANGELOG entry and AGENTS.md sentence

CHANGELOG.md gains a Changed entry under Unreleased for the threshold
raise. AGENTS.md Testing Conventions gains one sentence recording the
94% statements floor and its governance control, so agents reading
the repository see that the gate exists and is protected. The AGENTS.md
documentation-issue gate exempts CI-only changes; this change claims
that exemption (quality-gate change, no user-facing surface) while
still recording the CHANGELOG entry per the change-entry rule.

### Constitution cross-reference

The proposal assesses all six constitution principles as PASS. The
decisions map onto them as follows: D1 and D6 — Autonomous
Collaboration (self-describing policy artifact; the decision recorded
where collaborators read it) and Observable Quality (auditable
provenance; machine-actionable failure output); D4 and D5 —
Composability First (independently gateable suite) and Deterministic
Execution (tool-owned enforcement, checked-in value, identical
command locally and in CI); D2 — Testability (the gate itself is
verifiable in both passing and failing modes with executable
commands, recorded in tasks.md).

## Risks / Trade-offs

- **A 1.18-point window remains** (94 floor versus 95.18 measured):
  accepted. The margin absorbs cross-platform divergence and normal
  churn; further tightening is future governance work with fresh
  measurements. The window is now bounded and small (roughly 44
  statements).
- **Coverage can plateau just above the floor**: the conscious cost
  of rejecting a ratchet. Upward pressure remains through review
  expectations and future governance raises; a self-modifying gate
  was rejected as a gatekeeping violation (D1).
- **A static value can go stale as the suite grows**: 94% of a
  growing denominator still bounds uncovered statements
  proportionally, and the provenance trail (this design plus the
  config comment) hands future governance changes a baseline record.
  Deliberate, not auto-corrected.
- **Cross-platform divergence** (macOS local versus ubuntu-latest
  CI): same-commit V8 coverage can differ slightly across platforms
  — documented in this repository's CRAP-load history. The 1.18-point
  margin absorbs it; the pull request's CI run is the ubuntu
  confirmation that the floor holds on the CI platform.
- **Baseline-number reconciliation**: issue #780's shorthand
  attributes 94.58% to "round-3"; the artifact trail records it at
  the round-2 fresh measurement (tasks.md 10.7), with the round-3
  verification (11.8) confirming the suite green at 965 tests and
  94.58% standing as the final recorded value. The floor logic
  depends only on the lowest recorded baseline (94.58%), which both
  readings share; this design and the corrected proposal cite the
  artifact trail.
- **Fanout sibling coupling**: review-council-multi-model-fanout
  (complete, unarchived) spec'd D14 as "Vitest coverage MUST reach
  90 percent statements and 85 percent branches." Raising statements
  from 90 to 94 strengthens rather than contradicts that
  requirement, and the branches floor is preserved verbatim. No
  fanout artifact amendment is needed; archive order is independent
  of this change.
