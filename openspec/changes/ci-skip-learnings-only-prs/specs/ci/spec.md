<!-- capability: ci
     change: ci-skip-learnings-only-prs
     originating_issue: 1194
     Skip build/test/lint for learnings-only PRs while every
     required check still reports. -->

Requirement numbering: this change allocates CI-FR-001 through
CI-FR-004, the first numbered ids in the `ci` capability (the
live requirements at `openspec/specs/ci/spec.md` are unnumbered
predecessors, per the unnumbered-predecessor exception recorded
by `fix-pinact-gaze-pin`). That change's reservation set
CI-FR-001..CI-FR-020 aside for exactly this kind of change;
CI-FR-005..CI-FR-020 remain reserved for future ci-capability
changes. No other active ci delta allocates numbered ids.
Sibling capabilities' numbered requirements live in their own
deltas (for example review-dispatch's RD-FR ids). Requirement
ids MUST stay inside the allocated range.

## ADDED Requirements

### Requirement: Learnings-Only Change Detection

**ID: CI-FR-001** This requirement MUST be satisfied.

The Local CI (`ci_local.yml`) and CI Checks (`ci_checks.yml`)
workflows MUST each include a `changes` job that, for every
`pull_request` event targeting `main`, computes the files
changed between the pull request's merge base and its head,
and exports a `code` job output. The changed-file list MUST
be computed rename-safe (`--no-renames` semantics): a rename
counts as a deletion of the old path plus an addition of the
new path, and both paths are classified. The output MUST be
`'true'` when any changed file lies outside
`.uf/dewey/learnings/`, lacks a `.md` extension, or is not a
regular file, and `'false'` only when every changed file is
a regular `.md` file under that prefix. A regular file is a
git blob with file mode `100644` or `100755`; symlinks (mode
`120000`), submodules (mode `160000`), and any other
non-regular entry count as code.

The detector MUST process the changed paths NUL-delimited
(`git diff -z` semantics), match the prefix and the `.md`
suffix literally against each raw path, and write the `code`
output through a delimiter-based `GITHUB_OUTPUT` write, so
crafted filenames — embedded newlines, quotes, or shell
metacharacters — cannot inject outputs or alter the
classification. Events that are not `pull_request`, and any
indeterminate state (empty diff, failed diff, missing base),
MUST set `code` to `'true'`.

#### Scenario: Learnings-only pull request

- **GIVEN** a pull request whose changed files are all `.md`
  files under `.uf/dewey/learnings/`
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'false'`

#### Scenario: Non-Markdown file under the prefix fails safe

- **GIVEN** a pull request whose only changed file is
  `.uf/dewey/learnings/payload.bin`
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'true'`

#### Scenario: Mixed pull request

- **GIVEN** a pull request that changes at least one file
  outside `.uf/dewey/learnings/`
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'true'`

#### Scenario: Renamed code file fails safe

- **GIVEN** a pull request whose only change renames
  `src/auth.go` to `.uf/dewey/learnings/auth.md`
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'true'`, because the
  rename-safe diff reports the deleted `src/auth.go` path
  outside the prefix

#### Scenario: Symlink under the prefix fails safe

- **GIVEN** a pull request whose only changed file is a
  symlink `.uf/dewey/learnings/payload.md` (git file mode
  `120000`) whose target lies outside the prefix
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'true'`, because only
  regular files classify as learnings

#### Scenario: Crafted filename cannot alter classification

- **GIVEN** a pull request containing a changed path whose
  name embeds a newline, quote, or shell metacharacter
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be derived from the
  literal path bytes alone, and the `GITHUB_OUTPUT` write
  MUST carry exactly the one `code` assignment

#### Scenario: Push event fails safe

- **GIVEN** a push event to `main`
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'true'`, without
  requiring a pull request base

#### Scenario: Indeterminate diff fails safe

- **GIVEN** a pull request whose diff cannot be computed or
  yields no changed files
- **WHEN** the `changes` job runs
- **THEN** the `code` output MUST be `'true'`

### Requirement: Gating of Expensive CI Jobs

**ID: CI-FR-002** This requirement MUST be satisfied.

The `build-and-test`, `plugin-tests`, and
`council-review-action-tests` jobs in `ci_local.yml` and the
`call_reusable_ci` job in `ci_checks.yml` MUST declare
`needs: [changes]` and MUST run when the `changes` job's
`code` output is `'true'`, and MUST be skipped otherwise.
Whenever these jobs run, they MUST run with the same steps,
commands, thresholds, and pinned action SHAs as before this
change, so that any pull request changing anything other
than `.md` files under `.uf/dewey/learnings/` — alone or
together with learnings — and every push run on `main`
receive the full, unchanged CI.

#### Scenario: Code present runs full CI

- **GIVEN** a pull request with `code` output `'true'`
- **WHEN** the workflows run
- **THEN** every gated job MUST execute with the same steps,
  commands, thresholds, and pinned action SHAs as before this
  change

#### Scenario: Learnings-only PR skips gated jobs

- **GIVEN** a learnings-only pull request as defined by
  CI-FR-001
- **WHEN** the workflows run
- **THEN** `build-and-test`, `plugin-tests`,
  `council-review-action-tests`, and `call_reusable_ci` MUST
  be skipped, and no build, test, or lint step MUST execute

#### Scenario: Push to main runs all jobs

- **GIVEN** a push event to `main`
- **WHEN** the workflows run
- **THEN** every gated job MUST execute

### Requirement: Aggregate Required Check

**ID: CI-FR-003** This requirement MUST be satisfied.

Each affected workflow MUST provide an aggregate job that
declares `if: always()` and `needs` its `changes` job and
every gated job, that MUST fail when any of those jobs
concluded anything other than `success` or `skipped` —
including `failure`, `cancelled`, `timed_out`,
`action_required`, and `neutral` — and that MUST pass
otherwise. The repository's required-checks ruleset for pull
requests targeting `main` MUST require each affected
workflow's aggregate job, and MUST NOT require any check that
a learnings-only pull request would leave unreported. The
ruleset swap MUST be verified by reading the ruleset back
after the PATCH and confirming the required-check names;
when the PATCH fails after the workflow change merges, the
failure MUST be surfaced and the swap retried until that
read-back succeeds, and during the unverified window no pull
request MUST merge on an unreported required check — code
pull requests keep reporting every previous check name, and
learnings-only pull requests wait until the swap completes.
If verification does not succeed, the workflow change MUST
be reverted or the required checks restored to `Build and
Test`, `Plugin Tests`, and `Standardized CI / Run linters`
so no PR waits on an unreported check.

#### Scenario: Gated job failure fails the aggregate

- **GIVEN** a pull request with `code` output `'true'` in
  which any gated job fails
- **WHEN** the aggregate job runs
- **THEN** the aggregate MUST fail

#### Scenario: Cancelled gated job fails the aggregate

- **GIVEN** a workflow run in which any job needed by the
  aggregate was cancelled
- **WHEN** the aggregate job runs
- **THEN** the aggregate MUST fail

#### Scenario: Timed-out gated job fails the aggregate

- **GIVEN** a workflow run in which any job needed by the
  aggregate concluded `timed_out`
- **WHEN** the aggregate job runs
- **THEN** the aggregate MUST fail

#### Scenario: Skipped gated jobs pass the aggregate

- **GIVEN** a learnings-only pull request as defined by
  CI-FR-001
- **WHEN** the aggregate job runs
- **THEN** the aggregate MUST pass while the gated jobs report
  as skipped

#### Scenario: Required checks report on a learnings-only PR

- **GIVEN** a learnings-only pull request
- **WHEN** CI completes
- **THEN** every check required by the ruleset MUST report a
  conclusive passing status, with no required check left
  unreported

#### Scenario: Failed ruleset PATCH after the workflow merge

- **GIVEN** the workflow change has merged and the ruleset
  PATCH failed or is not yet verified by read-back
- **WHEN** pull requests are evaluated during that window
- **THEN** the swap MUST be retried until the read-back
  confirms the two aggregate checks, and no pull request MUST
  merge on an unreported required check

### Requirement: Ungated Workflows

**ID: CI-FR-004** This requirement MUST be satisfied.

The `ci_security.yml`, `ci_dependencies.yml`, and
`ci_crapload.yml` workflows MUST NOT declare a need on any
`changes` job and MUST NOT gate any job on the `changes`
job's outputs; their pull_request triggers, jobs, steps,
thresholds, and pinned action SHAs MUST remain identical to
their state before this change, so every pull request —
learnings-only or not — continues to receive the security,
dependency, and CRAP scans exactly as before.

#### Scenario: Security scans still run on a learnings-only PR

- **GIVEN** a learnings-only pull request as defined by
  CI-FR-001
- **WHEN** CI completes
- **THEN** `ci_security.yml` and `ci_dependencies.yml` MUST
  have run their scans on that pull request

#### Scenario: Ungated workflows carry no changes-job gate

- **GIVEN** the workflow files after this change
- **WHEN** any job in `ci_security.yml`,
  `ci_dependencies.yml`, or `ci_crapload.yml` is inspected
- **THEN** it MUST NOT declare `needs` on the `changes` job
  or gate on its `code` output

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
