---
status: draft
branch: opsx/ci-crapload-analysis
date: 2026-04-27
---

## Why

The unbound-force CI pipeline (`test.yml`) enforces global coverage
ratchets (80% overall, 90% for `internal/backlog/`) but has no
per-function CRAP score tracking. A developer can introduce a
highly complex, poorly tested function that doesn't move the global
needle enough to fail CI, but creates a maintenance liability.

The complytime organization has solved this with a reusable GitHub
Actions workflow (`complytime/org-infra/.github/workflows/
reusable_crapload_analysis.yml`) that runs Gaze CRAP analysis on
changed Go packages, compares results against a committed baseline,
and reports regressions, improvements, and new function violations
as a PR comment. complyctl consumes this workflow via a thin
`ci_crapload.yml` caller.

Additionally, the existing `test.yml` and `release.yml` workflows
use pinned action SHAs (e.g., `actions/checkout@11bd71901...`) but
there is no dependabot configuration to keep these pins current.
The repository has no `.github/dependabot.yml` at all.

This change brings both capabilities to unbound-force:

- **Per-function CRAP regression tracking**: Detect when a PR
  increases CRAP scores or introduces complex untested functions.
- **Automated dependency management**: Dependabot keeps action
  SHAs and Go module versions current via automated PRs.

## What Changes

Add a `ci_crapload.yml` CI workflow that calls the complytime
reusable CRAP analysis workflow (pinned to a commit SHA managed
by dependabot), create a dependabot configuration for
`github-actions` and `gomod` ecosystems, and generate a
`.gaze/baseline.json` from the current state of the codebase.

## Capabilities

### New Capabilities

- **CRAP Load CI check**: A blocking PR check that runs Gaze
  analysis on changed Go packages, compares against a committed
  baseline, and fails on regressions or new functions exceeding
  the CRAP threshold (default 30). Posts a rich PR comment with
  summary metrics, quadrant distribution, regressions,
  improvements, and new function tables.
- **Dependabot version updates**: Automated PRs for pinned
  GitHub Action SHAs (daily) and Go module versions (weekly,
  capped at 10 open PRs). Uses conventional commit prefixes
  (`ci:` for actions, `chore:` for gomod).
- **CRAP score baseline**: Committed `.gaze/baseline.json`
  capturing per-function CRAP and GazeCRAP scores for all
  functions in the codebase. Enables regression detection
  across PRs.

### Modified Capabilities

- Existing `test.yml` and `release.yml` pinned action SHAs
  will begin receiving dependabot update PRs. No workflow
  logic changes.

### Removed Capabilities

None.

## Impact

### Files Created (new)

- `.github/workflows/ci_crapload.yml` -- Consumer workflow
- `.github/dependabot.yml` -- Dependabot configuration
- `.gaze/baseline.json` -- CRAP score baseline

### Files Modified

None. Existing workflows are unchanged.

### Backward Compatibility

Fully backward compatible. The CRAP check is additive -- it
runs alongside the existing `test.yml` on PRs. Dependabot PRs
are advisory (auto-opened, not auto-merged). The baseline file
is inert until the workflow reads it.

## Documentation Impact

This change adds a new CI check visible to all PR authors. The
PR comment is self-documenting (explains scores, thresholds,
and links to analysis logs). No AGENTS.md or README changes
needed -- CI workflows are not documented in the project
structure section.

This change does not affect user-facing hero capabilities, CLI
commands, or workflows. No website documentation issue required.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The CRAP analysis workflow consumes Gaze output (a hero
artifact) through a well-defined JSON format with provenance
metadata. The reusable workflow is self-contained -- it
produces a Markdown comment body as an uploaded artifact that
the consumer downloads and posts. No synchronous inter-hero
coupling is introduced.

### II. Composability First

**Assessment**: PASS

The CRAP check is fully additive. It runs as a separate CI
job that does not affect the existing `test.yml` pipeline.
Removing the workflow file removes the check entirely --
no residual dependencies. Gaze is installed via `go install`
within the workflow, requiring no pre-deployment.

### III. Observable Quality

**Assessment**: PASS

This change directly strengthens observable quality. The
CRAP analysis produces machine-parseable JSON output (Gaze
report, per-function scores), uploads detailed analysis as
CI artifacts with 30-day retention, and posts human-readable
summaries as PR comments. All quality claims (regression
detection, threshold enforcement) are backed by the automated
baseline comparison.

### IV. Testability

**Assessment**: PASS

This change introduces CI configuration files (YAML), not
Go code. No new Go functions requiring test coverage.
The reusable workflow itself is tested in the
complytime/org-infra repository. The consumer workflow is
verified by tasks 4.1–4.4: YAML syntax validation, pinned
SHA verification, and baseline JSON structure validation.
These verification tasks serve as the testability evidence
for CI configuration artifacts.
