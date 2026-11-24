## Context

Required checks today (repository ruleset): `Build and Test`
and `Plugin Tests` (jobs in `ci_local.yml`) and
`Standardized CI / Run linters`, a check nested inside the
`complytime/org-infra` reusable CI call made by
`ci_checks.yml`'s `call_reusable_ci` job. Also running on PRs:
`ci_crapload.yml`, `ci_dependencies.yml`, `ci_security.yml`,
and `ci_local.yml`'s Council Review Pipeline Tests; the Gaze
AI Report job runs only on push to `main`.

Learnings PRs (`chore/dewey-learnings-*`) add only Markdown
under `.uf/dewey/learnings/`, cost roughly 30 minutes of full
CI each, and serialize through the merge train because
branches must be up to date before merge.

GitHub mechanics that shape this design:

- A job skipped by a job-level `if:` satisfies a required
  check of the same name — except (a) matrix jobs, which skip
  with the matrix expression unexpanded so per-leg names never
  report, and (b) reusable-workflow calls, which report only
  the caller's job name when skipped, so nested checks like
  `Standardized CI / Run linters` never report.
- A workflow-level `paths-ignore` on a workflow that provides
  a required check leaves that check permanently unreported
  ("Expected — Waiting for status to be reported").

Constitution alignment: PASS on all six principles as assessed
in `proposal.md`. Composability First drives Decision 1's
primitives-only detection; Security by Default drives
Decisions 1 and 6; Observable Quality drives Decision 4; the
fail-safe direction (Decision 2) protects every gate the CI
enforces today.

The Gaze AI Report job already restricts itself to push events
and needs `build-and-test` for its coverage artifact; on push,
`build-and-test` always runs (Decision 2), so the artifact is
always produced and that job is untouched.

## Goals / Non-Goals

### Goals

- Learnings-only PRs show every required check passing within
  a couple of minutes, without running build, tests, or
  linters.
- Every other PR — including mixed learnings-plus-code PRs —
  runs the full CI, identical in steps, commands, thresholds,
  and pinned SHAs.
- Push runs on `main` are unchanged.
- The required-check list no longer names individual jobs, so
  later job-list changes need no ruleset edit.

### Non-Goals

- Skipping `ci_crapload.yml`, `ci_dependencies.yml`, or
  `ci_security.yml` on PRs: they are not required checks and
  do not block the merge train; a possible follow-up. CI-FR-004
  pins their present ungated state so any such follow-up is a
  deliberate spec change, not an implementation drift.
- Changing `complytime/org-infra` reusable workflows
  (cross-repo scope).
- Extending the skip prefix beyond `.uf/dewey/learnings/`;
  compiled articles (`.uf/dewey/compiled/`) ride feature PRs
  by the Commit scope policy.
- Any use of `paths-ignore` (Decision 1).

## Decisions

### Decision 1: Detection via a changes job and plain git diff

Each workflow gains a `changes` job that checks out with
`fetch-depth: 0` (reusing the workflows' existing SHA-pinned
`actions/checkout` version) and, for `pull_request` events,
runs `git diff -z --no-renames --raw "$BASE_SHA...$HEAD_SHA"`:
NUL-delimited so a crafted filename cannot split or inject,
`--no-renames` so a rename reports both its old and new
paths, `--raw` so each entry carries its file mode. Any
entry that does not both start with `.uf/dewey/learnings/`
and end in `.md`, or whose mode is not a regular file
(`100644`/`100755`), sets the `code` output to `'true'`;
the loop reads each entry with `IFS= read -r -d ''` and
writes `code` through a heredoc delimiter so path bytes
never reach `GITHUB_OUTPUT` uncontrolled. Rejected alternatives:

- Workflow-level `paths-ignore`: permanently unreports a
  required check (see Context).
- `dorny/paths-filter`: a new third-party action — Security
  by Default's dependency-necessity rule defaults to
  do-not-add, and it introduces a new SHA pin to manage; git
  is already in the runner image.
- Passing the decision into org-infra's reusable CI: a
  cross-repo change; Decision 4 removes the need.

### Decision 2: Fail-safe direction is full CI

`code: 'true'` is the default; `'false'` requires positive
proof that every changed path is a `.md` file under the
prefix. Push events (no PR base exists), empty diffs, failed
diffs, and missing bases all classify as code present. A
detector bug then costs runner minutes, never gate coverage.

### Decision 3: The push short-circuit lives in the changes job

On push events the `changes` job sets `code: 'true'` without
diffing (there is no pull request base). Gated jobs carry a
single uniform gate — `needs: [changes]` plus
`if: needs.changes.outputs.code == 'true'` — with no per-job
event-type duplication.

### Decision 4: One aggregate job per workflow is the required
check

`ci_local.yml` gains `ci-ok` (name: `CI OK`); `ci_checks.yml`
gains an aggregate (name: `Standardized CI OK`). Both declare
`if: always()` and `needs` the `changes` job plus every gated
job, and fail when any needed job concluded anything other
than `success` or `skipped` — `failure`, `cancelled`,
`timed_out`, `action_required`, `neutral`, and any conclusion
GitHub introduces later. This handles the
reusable-workflow nesting case — the nested
`Standardized CI / Run linters` check cannot stay required,
because a skipped caller never reports it — is future-proof
for matrix jobs, and frees the ruleset from naming individual
jobs. Two aggregates rather than one because a job cannot
`needs` across workflows.

### Decision 5: Ruleset swap via gh api, right after the
workflow change merges

Merge order: land the workflow change first, then update the
ruleset — list the repo's rulesets with `gh api`, PATCH the
required-checks rule to require `CI OK` and
`Standardized CI OK`, removing `Build and Test`, `Plugin
Tests`, and `Standardized CI / Run linters`. The window
between merge and swap is minutes; only a learnings PR opened
inside it would wait on the unreported nested check (code PRs
are unaffected — every old name still reports), and the swap
unblocks it immediately. Requiring the aggregates before the
workflow change merges would hang every PR on checks that do
not exist yet — strictly worse. gh-hud opens learnings PRs
about once a day per checkout, so the collision probability
is low.

### Decision 6: Detection logic inline in both workflows

The roughly ten-line detection block is duplicated in
`ci_local.yml` and `ci_checks.yml` rather than extracted to a
shared script, composite action, or Go CLI (the
`cmd/coverage-gate` precedent). Rationale: no new artifact to
test and pin (Security by Default), the logic is one
NUL-delimited diff parse plus a per-path prefix, suffix, and
mode check, and CI-FR-001's scenarios pin both
copies to identical behavior. If review weighs DRY higher,
extraction is a mechanical follow-up.

## Risks / Trade-offs

- **Ordering window (Decision 5).** A learnings PR opened
  between the workflow merge and the ruleset swap waits on an
  unreported check until the swap. Mitigated by swapping
  immediately after merge.
- **GitHub semantics drift.** If GitHub changes how skipped
  jobs satisfy required checks, the aggregates are unaffected:
  they run under `if: always()` and are themselves never
  skipped.
- **Detector misclassification.** The fail-safe direction
  means a bug costs CI minutes, not gate coverage.
  The rename-safe diff reports both paths of a rename, so a
  file renamed into or out of the learnings prefix
  classifies as code — the stricter, fail-safe reading. The
  prefix, `.md` suffix, and regular-file mode matches are
  literal.
- **Inline duplication (Decision 6).** The two detection
  blocks could drift apart; CI-FR-001 covers both workflows
  and review compares the blocks directly.
- **Accepted trade-off.** Learnings-only PRs also skip the
  Council Review Pipeline Tests and the MegaLinter suite;
  accepted because they change no executable or lintable
  content. Security scans (`ci_security.yml`) still run on
  them.
