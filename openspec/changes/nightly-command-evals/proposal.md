## Why

Unit and plugin tests check uf's tools in isolation. Nothing checks
what a command does when an agent runs it end to end: whether it
calls the right tools in the right order, stops at its gates, keeps
its guardrails after context compression, and reaches the right
verdict.

A prototype harness (`.eval/` on `eval/review-council-arg-rejection`,
driven by `unbound-force/eval-infra`) runs `/uf.review-council`
through opencode with uf's real plugins, DCP and auto-question, and
checks deterministic assertions on the resulting trajectory. In its
first repeated runs it found a defect no existing test covers: a
Divisor reviewer that hits its output limit is counted as a
successful `APPROVE` vote (#1052). It also found that its own
runtime had silently lost DCP, because a failed plugin build was
recorded as ready.

Run once by hand, the harness only shows how a command behaves at a
single point in time. Run nightly, it catches regressions in
commands, plugins, agents and the runtime within a day of landing,
and builds the pass-rate and cost history needed before evals can
gate merges.

## What Changes

- Land the `.eval/` harness on main: five `/uf.review-council`
  fixtures (three argument-rejection fixtures and two council
  fixtures with a purpose-built mini repo), the isolated opencode
  runtime, and the run scripts.
- `run.sh` fetches the eval driver from `unbound-force/eval-infra`
  at a pinned commit, the same way it already fetches the pinned
  DCP and auto-question plugins, so `EVAL_DRIVER` becomes optional.
- New `ci_eval.yml` workflow: runs every fixture nightly and on
  manual dispatch, one job per fixture, and reports results in the
  job summary and as artifacts.
- `make eval` and `make eval-fast` run the same fixtures locally.
  Neither is part of `make check`.
- AGENTS.md documents the harness, the workflow, and how to add a
  fixture.

## Capabilities

### New Capabilities
- `command-evals`: end-to-end evals of uf commands, run nightly in
  CI and on demand locally (EV-FR-001 through EV-FR-008).

### Modified Capabilities
- None.

### Removed Capabilities
- None.

## Impact

- **New files**: `.eval/**`, `.github/workflows/ci_eval.yml`.
- **Edited files**: `Makefile` (two targets), `AGENTS.md` (CI
  table, project structure, testing conventions).
- **Repository settings**: an `ENMAAS_KEY` Actions secret. The
  workflow needs no other secret and only `contents: read`.
- **Dependency**: the driver changes on eval-infra's
  `feat/trajectory-assertions` branch must be merged into
  `unbound-force/eval-infra` main before the pin can point at an
  org commit.
- **CI cost**: nightly runner time of roughly 2 to 3.5 hours across
  parallel jobs, the longest about 90 minutes. Model tokens go to
  the self-hosted EnMaaS endpoint and carry no per-token charge.
- **Not deployed**: `uf init` does not scaffold `.eval/`. No
  scaffold mirror change, no website documentation impact.
- **Not a gate**: the workflow runs on `schedule` and
  `workflow_dispatch` only, so it cannot block a PR. Making evals
  a required check is a later change.
- **CI replication**: `make check` for the repo changes;
  `make eval-fast` for the harness itself.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The harness exercises commands through their published artifacts
and tool results, and reports through job summaries and
machine-readable artifacts. It adds no runtime coupling to any hero.

### II. Composability First

**Assessment**: PASS

The driver lives in its own repo and is pinned by commit. The
harness depends only on opencode, the pinned plugins, and the
provider key; it runs the same way locally and in CI.

### III. Observable Quality

**Assessment**: PASS

Every run records pass or fail per assertion, wall time, input,
output, reasoning and cache tokens, and the full trajectory of tool
calls, questions and permission asks, kept as CI artifacts.

### IV. Testability

**Assessment**: PASS

Assertions are deterministic checks on the recorded trajectory, with
no judge model. Each fixture runs in a fresh temporary snapshot of
the repo, and GitHub is replaced by a stub that records and refuses
every call. The model endpoint is the one external service, and it
is the thing under test.

### V. Security by Default

**Assessment**: PASS

The provider key comes only from an Actions secret and is never
printed. Workflow permissions are `contents: read`, the workflow
never runs on pull requests from forks, `external_directory` is
denied outside the run's workdir and temp areas, and every
permission ask is rejected rather than left waiting.

### VI. Deterministic Execution

**Assessment**: PASS

The runtime is pinned (opencode version, plugin commits, driver
commit). A run fails before any fixture starts if a pinned plugin
fails to build or a required tool is not offered, so a run can't
silently test a different runtime.
