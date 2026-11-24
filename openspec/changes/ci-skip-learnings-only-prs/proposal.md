## Why

Learnings PRs (`chore(dewey): store learnings from ...` on
branches `chore/dewey-learnings-*`) only add Markdown files under
`.uf/dewey/learnings/`. gh-hud opens them from fleet checkouts,
uf-divisor approves them when every changed file is a learning,
and auto-merge lands them. Nothing in such a PR can affect a
build, a test, or a lint, yet each one runs this repo's full CI
on every push and every update from main.

Because branches must be up to date before merge, the merge train
lands these PRs one at a time: update from main, wait for full
CI, merge, repeat. On 2026-11-24 the gaze repo had 16 open
learnings PRs at roughly 30 minutes of CI each, and the queue
stalled behind cancelled duplicate runs. This repo carries the
most learnings (1,000+ files under `.uf/dewey/learnings/`) and
the most learnings PRs, so it pays the largest share of that
cost.

Originating issue: #1194.

## What Changes

A pull request whose changed files are all regular Markdown
(`.md`) files under `.uf/dewey/learnings/` skips the build,
test, and lint jobs, while every required check still reports
as passing. Any other PR — including one that mixes learnings
with other files, renames a file into or out of the learnings
directory, or places non-Markdown or non-regular files inside
it — runs the full CI exactly as today. Push runs on `main`
are unchanged.

Concretely:

1. Add a cheap `changes` job to `ci_local.yml` and
   `ci_checks.yml` that diffs the PR against its base
   (rename-safe, NUL-delimited, mode-aware) and outputs whether
   any changed file lies outside `.uf/dewey/learnings/`, is not
   a Markdown (`.md`) file, or is not a regular file.
2. Gate the expensive jobs on that output — Build and Test,
   Plugin Tests, and Council Review Pipeline Tests in
   `ci_local.yml`, and the Standardized CI reusable-workflow call
   in `ci_checks.yml` — so they are skipped on learnings-only PRs
   and run otherwise.
3. Add one always-running aggregate job per affected workflow
   (`if: always()`, `needs:` every gated job, failing when any
   of them concluded anything other than success or skipped)
   and make the aggregates the
   required checks. This handles the two cases where a skipped
   job does not report under its required name (matrix jobs and
   reusable-workflow calls) and keeps the required-check list
   stable under later job changes.
4. Update the repository ruleset so the required checks name the
   aggregate jobs instead of the individual job names.

Workflow-level `paths-ignore` is deliberately not used for any
workflow that provides a required check: the check would never
report, and the PR would wait forever on "Expected — Waiting for
status to be reported".

Out of scope: `ci_crapload.yml`, `ci_dependencies.yml`, and
`ci_security.yml` keep running on PRs unchanged. They are not
required checks and do not block the merge train; extending the
skip to them is a possible follow-up.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `ci`: adds requirements for learnings-only PR detection
  (CI-FR-001), gating of the expensive CI jobs (CI-FR-002), an
  aggregate required check (CI-FR-003), and ungated security,
  dependency, and CRAP scans on every pull request (CI-FR-004).

### Removed Capabilities

(none)

## Impact

- `.github/workflows/ci_local.yml`: new `changes` job; gating of
  `build-and-test`, `plugin-tests`, and
  `council-review-action-tests`; new aggregate job.
- `.github/workflows/ci_checks.yml`: new `changes` job; gating of
  the `call_reusable_ci` job; new aggregate job.
- Repository ruleset (GitHub settings, via `gh api`): required
  checks move from `Build and Test`, `Plugin Tests`, and
  `Standardized CI / Run linters` to the two aggregate jobs.
- `ci_security.yml`, `ci_dependencies.yml`, and `ci_crapload.yml`
  stay ungated and untouched (CI-FR-004): security, dependency,
  and CRAP scans keep running on every PR exactly as before.
- No source code, no threshold changes, no coverage or CRAP
  ratchet changes; `push` runs on `main` keep running every job.
- The implementing PR itself changes workflow files, so it runs
  the full CI and exercises the code-present path end to end.
- Documentation: this change is CI-only but alters user-visible
  merge governance (the required-check names), so a
  documentation issue MUST be filed against this repository
  before the implementing PR merges (task 4.5), referencing
  originating issue #1194 and this change by name so the
  governance change is traceable end to end, alongside the
  CHANGELOG entry that task 4.3 adds.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

The change alters CI orchestration only; heroes keep
collaborating through the same artifacts. The `changes` job
emits a machine-readable job output that downstream jobs consume
through GitHub Actions primitives — no new coupling and no
synchronous interaction between heroes.

### II. Composability First

**Assessment**: PASS

Each workflow remains independently functional. Detection uses
only git and GitHub Actions primitives (`needs`, `if`, job
outputs) with no new third-party actions, so the workflows keep
operating standalone.

### III. Observable Quality

**Assessment**: PASS

Every PR still gets a machine-parseable, always-reported
verdict: the aggregate job runs unconditionally and fails when
any gated job concluded anything other than success or
skipped. No check becomes silent;
the skip decision is visible in the `changes` job logs with the
file list that drove it.

### IV. Testability

**Assessment**: PASS

The detection rule (all changed files are regular `.md` files
under `.uf/dewey/learnings/`) is deterministic and verifiable
per scenario — learnings-only PR, mixed PR, code-only PR, push
to main — on real PRs without external services.

### V. Security by Default

**Assessment**: PASS

No new actions or reusable workflows are introduced; existing
SHA-pinned actions are reused, and the `changes` job runs under
the workflows' existing read-only `contents: read` permission.
The fail-safe direction of the detector is full CI: any
indeterminate state (empty diff, diff failure, non-PR event)
classifies as code present.

### VI. Deterministic Execution

**Assessment**: PASS

The detection procedure is a mechanical git diff with a path
prefix check executed by the CI runner itself — no procedural
prose is added to commands, agents, or skills, and nothing needs
re-reasoning per run.
