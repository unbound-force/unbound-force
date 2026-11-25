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

Each workflow gains a `changes` job that, for `pull_request`
events only, checks out with `fetch-depth: 0` (reusing the
workflows' existing SHA-pinned `actions/checkout` version)
and runs the shared detector (Decision 6), which runs
`git diff -z --no-renames --raw "$BASE_SHA...$HEAD_SHA"`:
NUL-delimited so a crafted filename cannot split or inject,
`--no-renames` so a rename reports both its old and new
paths, `--raw` so each entry carries its file mode. Any
entry that does not both start with `.uf/dewey/learnings/`
and end in `.md`, or whose mode is not a regular file
(`100644`/`100755`), sets the `code` output to `'true'`;
the loop reads each entry with `IFS= read -r -d ''`, and
the job accepts only `true` or `false` from the detector
(anything else becomes `true`) and writes `code` through a
heredoc delimiter, so path bytes never reach `GITHUB_OUTPUT`. Rejected alternatives:

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
checking out or diffing (there is no pull request base). Gated jobs carry a
single uniform gate — `needs: [changes]` plus
`if: needs.changes.outputs.code == 'true'` — with no per-job
event-type duplication.

### Decision 4: One aggregate job per workflow is the required
check

`ci_local.yml` gains `ci-ok` (name: `CI OK`); `ci_checks.yml`
gains an aggregate (name: `Standardized CI OK`). Both declare
`if: always()` and `needs` the `changes` job plus every gated
job. They fail unless the `changes` job concluded `success`
(it has no `if:`, so a skip there means something broke),
and fail when any gated job concluded anything other
than `success` or `skipped` — `failure`, `cancelled`,
`timed_out`, `action_required`, `neutral`, and any conclusion
GitHub introduces later. This handles the
reusable-workflow nesting case — the nested
`Standardized CI / Run linters` check cannot stay required,
because a skipped caller never reports it — is future-proof
for matrix jobs, and frees the ruleset from naming individual
jobs. Two aggregates rather than one because a job cannot
`needs` across workflows.

### Decision 5: Ruleset swap via safe-settings, right after
the workflow change merges

The required-checks ruleset ("safe-settings: unbound-force
required checks") is managed by safe-settings from
`unbound-force/.github` `safe-settings/repos/unbound-force.yml`.
A direct `gh api` PATCH would be reverted by the next sync, so
the swap lives in that config (unbound-force/.github#73).
Merge order: land the workflow change first, then merge
unbound-force/.github#73 and run the "Safe Settings Sync"
workflow for `unbound-force` (dry run, then apply), which
requires `CI OK` and `Standardized CI OK` and removes `Build
and Test`, `Plugin Tests`, and `Standardized CI / Run
linters`. The window between merge and swap is minutes when
the sync runs right away; only a learnings PR opened
inside it would wait on the unreported nested check (code PRs
are unaffected — every old name still reports), and the swap
unblocks it immediately. Requiring the aggregates before the
workflow change merges would hang every PR on checks that do
not exist yet — strictly worse. gh-hud opens learnings PRs
about once a day per checkout, so the collision probability
is low.

### Decision 6: Detection logic in one shared, tested script

The detection logic lives in `scripts/detect-learnings-only.sh`,
called by the `changes` job in both `ci_local.yml` and
`ci_checks.yml`, so the two workflows cannot classify a PR
differently. `scripts/detect-learnings-only_test.sh` builds a
throwaway git repository per scenario (every CI-FR-001
scenario plus nested, executable, deleted, renamed,
submodule, symlink-replacement, bare-`.md`,
lookalike-prefix, and newline-smuggling paths) and runs as a
step of the `changes` job before detection, so a broken
detector fails the job and both aggregates instead of
silently skipping CI. A plain repo script adds no action pin
or dependency (Security by Default). A composite action or Go
CLI (the `cmd/coverage-gate` precedent) was rejected as more
machinery than one diff parse needs. Review originally
accepted the logic inline in both workflows, then flagged the
two copies as a drift risk, which this decision removes.

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
- **Detector changes (Decision 6).** A PR that edits the
  detector script changes a non-learnings file, so it always
  runs full CI. Its own test step runs first, from that PR's
  head, the same trust level as an edit to the workflow
  files themselves.
- **Accepted trade-off.** Learnings-only PRs also skip the
  Council Review Pipeline Tests and the whole Standardized CI
  reusable workflow: MegaLinter (actionlint, shellcheck,
  golangci-lint, Gitleaks), PR title validation, and the
  action pin checks. Accepted because they change no
  executable or lintable content. Two of these losses are
  narrower than they look. Secret scanning still runs through
  the Trivy source scan in `ci_security.yml` (secrets and
  misconfiguration), though it is not a full Gitleaks run, so
  a token pasted into an AI-written learning is caught only
  if Trivy's secret rules match it. Learnings PRs are opened
  by gh-hud with a fixed conventional title, so skipping
  title validation costs little. If a Gitleaks miss shows up
  in practice, add a Gitleaks step that runs regardless of
  the `changes` output.
