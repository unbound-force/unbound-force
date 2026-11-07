## Why

The unleash-tier-routing change recorded the TypeScript plugin
suite's coverage baselines in its design risk notes (94.61%
statements at round-1 verification, tasks.md 9.6; 94.58% at the
round-2 fresh measurement that stood through round-3, tasks.md
10.7 and 11.8) but deliberately deferred configuring a coverage
gate — the AGENTS.md gatekeeping rule classifies
threshold changes as governance decisions that require their own
spec workflow. The round-3 Divisor review council (divisor-testing,
coverage-governance) flagged that without a tracking issue a
deferral is indistinguishable from abandonment; issue #780 was
filed to track the debt.

Since that deferral, the review-council-multi-model-fanout change
landed a floor of 90% statements / 85% branches in
`.opencode/vitest.config.ts` (design D14, commit 7cb5d8e),
enforced natively by the `plugin-test` Make target and the Local
CI `plugin-tests` job. That floor, however, sits far below the
recorded evidence:

- Recorded baselines: 94.61% (round-1) and 94.58% (round-3)
  statements
- Current measurement on this change's branch point (2026-10-22,
  `npx vitest run --coverage`): 95.18% statements (3521/3699),
  87.89% branches, 98.78% functions, 95.39% lines

The result is a 5.18-point silent-regression window: plugin-suite
statements coverage could fall from 95.18% to 90% — roughly 190
uncovered statements — without any gate failing. The deferral
this issue tracks is therefore half-paid: enforcement machinery
exists, but the threshold is not "near the recorded baseline" as
the issue's scope requires.

## What Changes

- Raise the statements coverage threshold in
  `.opencode/vitest.config.ts` from 90 to 94 — just below the
  lowest recorded baseline (94.58%), locking in the
  unleash-tier-routing-era achievement while leaving measured
  working room (1.18 points below current). This cuts the
  silent-regression window by 77% (5.18 → 1.18 points).
- Decide the threshold policy the issue deferred: a **static
  governance-raised floor**, not an auto-tightening ratchet.
  The floor changes only through spec-workflow governance with
  documented rationale — mirroring how the Go-side gates
  (`coverage-gate.json` scopes, the global 80% check) are
  managed. The full reasoning is in design.md D1.
- Update the `vitest.config.ts` provenance comment to record the
  baseline measurements, the current reading, and issue #780,
  while preserving the fanout D14 attribution for the branches
  floor.
- No new CI wiring: the existing native enforcement chain
  (`make plugin-test` → lockless install through the pinned
  `PLUGIN_TEST_NPM` runner with lifecycle scripts disabled +
  `npx vitest run --coverage`
  → vitest threshold failure exits non-zero; the `plugin-tests`
  job in `ci_local.yml` and `make check` both invoke that
  target) already fails on breach. Verification tasks prove both
  the passing and failing modes of the gate.
- The branches threshold stays at 85 (the fanout D14 floor);
  raising it is out of scope for this change (issue #780 scopes
  the work to a statements threshold). The measured 87.89% is
  recorded in design.md as reference for any follow-up.
- Spec delta: the `ci` capability gains a Plugin Suite Coverage
  Gate requirement pinning the floor, the enforcement path, and
  the governance constraint.
- Documentation: a CHANGELOG entry and an AGENTS.md
  testing-conventions sentence documenting the floor, so agents
  reading the repo see that the gate exists and is
  governance-controlled.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci`: adds the Plugin Suite Coverage Gate requirement — the
  TypeScript plugin suite MUST maintain at least 94% statements
  coverage, enforced by the same `plugin-test` Make target CI
  invokes, changeable only through spec-workflow governance.

### Removed Capabilities

None.

## Impact

- **Files changed**:
  `.opencode/vitest.config.ts` (threshold value + provenance
  comment), `CHANGELOG.md`, `AGENTS.md` (testing-conventions
  sentence), plus this change's spec artifacts. No CI workflow
  files, no dependency manifests, no Go code.
- **Behavior**: any future change that drops plugin-suite
  statements coverage below 94% fails `make plugin-test`, `make
  check`, and the Local CI `plugin-tests` job, with vitest
  naming the metric, the actual value, and the threshold.
- **Sibling changes**: `review-council-multi-model-fanout`
  (complete, unarchived) spec'd its D14 floor as "Vitest
  coverage MUST reach 90 percent statements and 85 percent
  branches." Raising statements to 94 strengthens that
  requirement rather than contradicting it, so no fanout
  artifact amendment is needed and archive order is
  independent.
- **Scaffold**: `.opencode/vitest.config.ts` is a live-only
  file with no embedded scaffold mirror (only `package.json` /
  `package-lock.json` are scaffolded), so no drift sync is
  required; the scaffold manifests are untouched.
- **Exemptions claimed**: the AGENTS.md documentation-issue gate
  exempts CI-only changes; this change is a quality-gate/CI
  change with no user-facing surface. The CHANGELOG entry is
  still included per the change-entry rule.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The gate is a checked-in, self-describing policy artifact
(`vitest.config.ts` thresholds plus the provenance comment).
Enforcement requires no synchronous interaction: any worker or
CI run observes the same failure with the same evidence. The
policy decision this change makes (floor vs ratchet) is recorded
in design.md where future collaborators read it rather than
having to rediscover it.

### II. Composability First

**Assessment**: PASS

The change uses Vitest's native threshold mechanism — no new
tooling, no custom scripts, no new dependencies. The plugin
suite remains independently runnable and gateable via its
documented command (`cd .opencode && npx vitest run --coverage`)
without the rest of the repository, and the repo without the
plugin suite is unaffected.

### III. Observable Quality

**Assessment**: PASS

Vitest's threshold failure is machine-actionable output: a
non-zero exit plus an error naming the metric, the measured
value, and the configured threshold. The coverage summary
(statements, branches, functions, lines with numerators and
denominators) is emitted on every run, and this change's
artifacts record the measured provenance (baselines and current
reading) so the floor is auditable against evidence.

### IV. Testability

**Assessment**: PASS

The gate itself is testable in isolation without external
services: the positive case is the existing suite passing at the
raised floor, and the negative case is a scratch verification
that a threshold above the measured actual fails the run with
the expected error. Both are executable commands in tasks.md.

### V. Security by Default

**Assessment**: PASS

No new inputs, secrets, network access, or permissions are
introduced. The install discipline is the lockless install
reconciled by `stop-shipping-plugin-lock` (pinned
`PLUGIN_TEST_NPM` runner, lifecycle scripts disabled, no lock
read or written), and the threshold raises (never lowers) a
quality bar, adding no attack surface.

### VI. Deterministic Execution

**Assessment**: PASS

The work routes through the executable tool (Vitest's native
`coverage.thresholds`) rather than procedural prose or custom
shell arithmetic. Local and CI enforcement are the same command
(`make plugin-test`), and same-commit coverage measurement is
deterministic, so the gate cannot pass locally and fail in CI
for environmental reasons.
