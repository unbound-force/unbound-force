## ADDED Requirements

### Requirement: Plugin Suite Coverage Gate

The repository MUST enforce a minimum statements coverage
threshold of 94 percent on the `.opencode/` TypeScript plugin
suite through the Vitest `coverage.thresholds` configuration.
The floor MUST sit near the recorded evidence baselines
(94.61 and 94.58 percent recorded across the unleash-tier-routing
verification rounds, 95.18 percent measured 2026-10-22).

The gate MUST be enforced by the `plugin-test` Make target
(a lockless install through the pinned `PLUGIN_TEST_NPM` runner
with lifecycle scripts disabled, followed by `npx vitest run
--coverage`), and Local CI MUST invoke that same target, so
local and CI enforcement are identical. A threshold failure
MUST produce a non-zero exit and name the metric, the measured
value, and the configured threshold.

The branches threshold MUST remain at least 85 percent (the
review-council-multi-model-fanout design D14 floor). The
statements floor MUST NOT be lowered, and the enforcement path
MUST NOT be bypassed or weakened, except through a spec-workflow
governance change with documented rationale per the AGENTS.md
gatekeeping rule.

#### Scenario: Coverage at or above the floor

- **GIVEN** the plugin suite measures at least 94 percent
  statements coverage
- **WHEN** `make plugin-test` runs
- **THEN** the target passes and the coverage summary is
  reported

#### Scenario: Statements coverage regresses below the floor

- **GIVEN** the plugin suite measures below 94 percent
  statements coverage
- **WHEN** `make plugin-test` runs
- **THEN** the target fails with a non-zero exit
- **AND** the failure names statements, the measured value,
  and the 94 percent threshold

#### Scenario: CI parity blocks a regressing pull request

- **GIVEN** a pull request drops plugin-suite statements
  coverage below 94 percent
- **WHEN** the Local CI `plugin-tests` job runs
- **THEN** the job fails because it invokes the same
  `plugin-test` target that enforces the floor

#### Scenario: Threshold lowered without governance

- **GIVEN** a change lowers the statements threshold below 94
  percent or removes the enforcement path
- **AND** the change is not an approved spec-workflow
  governance change with documented rationale
- **WHEN** the change is reviewed
- **THEN** it MUST be rejected per the AGENTS.md gatekeeping
  rule

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
