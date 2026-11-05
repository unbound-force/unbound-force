<!-- Requirement allocation: this delta owns GATE-FR-001 through
GATE-FR-012. GATE-FR-013 through GATE-FR-099 are reserved for future
gate-command changes. Other capability requirements are allocated in
their respective OpenSpec deltas. -->

## ADDED Requirements

### Requirement: Gate command existence

**ID:** GATE-FR-001

The `unbound-force` CLI MUST provide a `gate` subcommand
that performs phase-specific constitution compliance checks.

#### Scenario: Gate command is registered

- **GIVEN** the `unbound-force` binary is built
- **WHEN** the user runs `uf gate --help`
- **THEN** help text is displayed showing usage, phase
  values, and format options

### Requirement: Phase flag

**ID:** GATE-FR-002

The `gate` command MUST accept a `--phase` flag with one
of the following values: `specify`, `plan`, `implement`,
`review`, `pr`. The command MUST reject any other value
with a non-zero exit code.

#### Scenario: Valid phase flag

- **GIVEN** a project directory with spec artifacts
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the command executes phase-specific checks
  and exits with code 0 if all pass

#### Scenario: Invalid phase flag

- **GIVEN** any project directory
- **WHEN** the user runs `uf gate --phase deploy`
- **THEN** the command exits with code 1 and an error
  message listing valid phases

#### Scenario: Missing phase flag

- **GIVEN** any project directory
- **WHEN** the user runs `uf gate` without `--phase`
- **THEN** the command exits with code 1 and an error
  message indicating `--phase` is required

### Requirement: JSON output format

**ID:** GATE-FR-003

The `gate` command MUST support `--format json` for
machine-readable output. The default format MUST be
`text`. The JSON output MUST include:

- `version`: schema version string (e.g., "1.0.0")
- `producer`: tool identifier ("uf-gate")
- `timestamp`: ISO 8601 UTC timestamp of the check run
- `branch`: current git branch (empty if unavailable)
- `commit`: checked git commit (empty if unavailable)
- `tool_version`: producing `uf` binary version
- `phase`: the phase checked (string)
- `passed`: overall result (boolean)
- `checks`: array of individual check results, each with
  `name` (string), `passed` (boolean), `message` (string)
- `summary`: object with `total`, `passed`, `failed`
  (integers)

#### Scenario: JSON output on failure

- **GIVEN** a project directory missing a spec artifact
- **WHEN** the user runs `uf gate --phase implement
  --format json`
- **THEN** the command outputs valid JSON to stdout with
  `"passed": false` and exits with code 1

#### Scenario: JSON output on success

- **GIVEN** a project directory with all required artifacts
  for the implement phase
- **WHEN** the user runs `uf gate --phase implement
  --format json`
- **THEN** the command outputs valid JSON to stdout with
  `"passed": true` and exits with code 0

#### Scenario: Invalid format flag

- **GIVEN** any project directory
- **WHEN** the user runs `uf gate --phase specify
  --format xml`
- **THEN** the command exits with code 1 and an error
  message to stderr listing valid formats (text, json)

### Requirement: Specify phase checks

**ID:** GATE-FR-004

When `--phase specify`, the gate MUST check:

- A spec artifact exists (openspec `proposal.md` or
  speckit `spec.md` in a feature directory).

#### Scenario: Specify phase passes with OpenSpec

- **GIVEN** `openspec/changes/my-change/proposal.md` exists
- **WHEN** the user runs `uf gate --phase specify`
- **THEN** the `spec-exists` check passes

#### Scenario: Specify phase passes with Speckit

- **GIVEN** `specs/001-my-feature/spec.md` exists
- **WHEN** the user runs `uf gate --phase specify`
- **THEN** the `spec-exists` check passes

#### Scenario: Specify phase fails

- **GIVEN** no spec artifacts exist in either location
- **WHEN** the user runs `uf gate --phase specify`
- **THEN** the `spec-exists` check fails and the command
  exits with code 1

### Requirement: Plan phase checks

**ID:** GATE-FR-005

When `--phase plan`, the gate MUST check:

- A spec artifact exists (same as specify phase)
- A plan/design artifact exists (openspec `design.md` or
  speckit `plan.md`)

#### Scenario: Plan phase passes

- **GIVEN** `openspec/changes/my-change/proposal.md` and
  `openspec/changes/my-change/design.md` exist
- **WHEN** the user runs `uf gate --phase plan`
- **THEN** both `spec-exists` and `plan-exists` checks pass

#### Scenario: Plan phase fails missing design

- **GIVEN** `openspec/changes/my-change/proposal.md` exists
  but `design.md` does not
- **WHEN** the user runs `uf gate --phase plan`
- **THEN** the `plan-exists` check fails

### Requirement: Implement phase checks

**ID:** GATE-FR-006

When `--phase implement`, the gate MUST check:

- A spec artifact exists
- A plan/design artifact exists
- A tasks artifact exists (openspec `tasks.md` or speckit
  `tasks.md`)
- A coverage strategy is present in the plan or tasks
  artifact (Constitution IV). The check MUST detect a
  coverage strategy by matching at least one of:
  - A heading containing "coverage" (case-insensitive)
    that does not negate it — headings such as "No
    coverage strategy" or "without coverage" MUST NOT
    satisfy the check
  - Body text containing any of: "coverage target",
    "unit test", "integration test", "e2e test"
  - A line matching `coverage.*%` (percentage target)
  - None of the signals above count when the carrying line
    denies the strategy — body text such as "No unit tests
    are planned", or a measured zero-percent line such as
    "current coverage: 0%" — or when the line lies inside
    a fenced code block: documented syntax is not a
    strategy, matching GATE-FR-007's fenced-block exclusion

#### Scenario: Implement phase passes

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  and tasks.md contains a "## Coverage strategy" heading
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** all checks pass and exit code is 0

#### Scenario: Implement phase fails missing coverage

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  but no coverage strategy heading or keywords are present
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `coverage-strategy` check fails

#### Scenario: Coverage strategy detected by keyword

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  and tasks.md contains "coverage target: >= 80%"
  but no coverage heading
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `coverage-strategy` check passes

#### Scenario: Negated coverage heading does not satisfy detection

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  and design.md contains the heading "## No coverage
  strategy is planned" with no other coverage signal
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `coverage-strategy` check fails

#### Scenario: Negated body text does not satisfy detection

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  and design.md contains the sentence "No unit tests are
  planned" with no other coverage signal
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `coverage-strategy` check fails

#### Scenario: Measured zero-percent coverage does not satisfy detection

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  and design.md contains the line "Current coverage: 0%"
  with no other coverage signal
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `coverage-strategy` check fails

#### Scenario: Fenced coverage syntax does not satisfy detection

- **GIVEN** proposal.md, design.md, and tasks.md exist,
  and design.md documents a "## Coverage Strategy" heading
  only inside a fenced code block
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `coverage-strategy` check fails

### Requirement: Review phase checks

**ID:** GATE-FR-007

When `--phase review`, the gate MUST check:

- All task checkboxes in the tasks artifact are marked
  complete (`[x]`). The check scans tasks.md for lines
  matching `- [ ]` outside fenced code blocks — if any
  unchecked boxes remain, the check fails. Fences close
  only on their own marker character at least as long as
  the opening fence (CommonMark), so a line using the
  other marker inside an open fence is content and cannot
  end the fence.
- A review has been executed. The check looks for the
  `<!-- spec-review: passed -->` or
  `<!-- code-review: passed -->` marker in tasks.md.

#### Scenario: Review phase passes

- **GIVEN** tasks.md exists with all checkboxes `[x]`
  and contains `<!-- code-review: passed -->`
- **WHEN** the user runs `uf gate --phase review`
- **THEN** both `tasks-complete` and `review-run`
  checks pass

#### Scenario: Review phase fails incomplete tasks

- **GIVEN** tasks.md exists with at least one `- [ ]`
  checkbox
- **WHEN** the user runs `uf gate --phase review`
- **THEN** the `tasks-complete` check fails with a
  message indicating uncomplete tasks remain

#### Scenario: Mixed-marker fence content does not end a fence

- **GIVEN** tasks.md contains a fenced code block opened
  with three backticks, a line of three tildes inside the
  fence, the closing three backticks, and an unchecked
  `- [ ]` task after the fence
- **WHEN** the user runs `uf gate --phase review`
- **THEN** the `tasks-complete` check fails

#### Scenario: Review phase fails no review marker

- **GIVEN** tasks.md exists with all checkboxes `[x]`
  but no review marker comment
- **WHEN** the user runs `uf gate --phase review`
- **THEN** the `review-run` check fails with a message
  indicating no review has been recorded

#### Scenario: Fenced example does not count as an incomplete task

- **GIVEN** tasks.md contains all checkboxes `[x]` and a
  fenced code block documenting the `- [ ]` syntax
- **WHEN** the user runs `uf gate --phase review`
- **THEN** the `tasks-complete` check passes

### Requirement: PR phase checks

**ID:** GATE-FR-008

When `--phase pr`, the gate MUST check:

- A review pass marker exists. The check looks for
  `<!-- code-review: passed -->` in tasks.md.
- The current git branch is not `main` (no direct
  commits). The check executes
  `git rev-parse --abbrev-ref HEAD` using `exec.Command`
  with explicit argument separation (no shell
  invocation).

#### Scenario: PR phase passes

- **GIVEN** the current git branch is `opsx/uf-gate`
  and tasks.md contains `<!-- code-review: passed -->`
- **WHEN** the user runs `uf gate --phase pr`
- **THEN** both `review-pass` and `not-on-main` checks
  pass

#### Scenario: PR phase fails on main

- **GIVEN** the current git branch is `main`
- **WHEN** the user runs `uf gate --phase pr`
- **THEN** the `not-on-main` check fails and exit code
  is 1

#### Scenario: PR phase fails no review pass

- **GIVEN** the current git branch is `opsx/uf-gate`
  but tasks.md has no `<!-- code-review: passed -->`
  marker
- **WHEN** the user runs `uf gate --phase pr`
- **THEN** the `review-pass` check fails

#### Scenario: Git not available

- **GIVEN** `git` is not in PATH
- **WHEN** the user runs `uf gate --phase pr`
- **THEN** the `not-on-main` check fails with a message
  suggesting `uf doctor` to check tool availability

#### Scenario: Not a git repository

- **GIVEN** `--dir` points to a directory that is not a
  git repository
- **WHEN** the user runs `uf gate --phase pr`
- **THEN** the `not-on-main` check fails with a
  descriptive message (not a raw exec error)

#### Scenario: Detached HEAD state

- **GIVEN** the git repository is in detached HEAD state
- **WHEN** the user runs `uf gate --phase pr`
- **THEN** the `not-on-main` check passes (detached HEAD
  is not `main`)

### Requirement: Exit code semantics

**ID:** GATE-FR-009

The `gate` command MUST exit with code 0 when all checks
for the requested phase pass. It MUST exit with code 1
when any check fails or arguments are invalid (phase,
format, and directory validation per GATE-FR-002,
GATE-FR-003, and GATE-FR-010). It MUST exit with code 2
when an internal error occurs (e.g., filesystem
permission denied, unexpected failures). It MUST always
produce output (text or JSON) before exiting.

Check results (JSON or text) MUST be written to stdout.
Validation errors (invalid phase, invalid format,
non-existent directory) MUST be written to stderr.

#### Scenario: Non-zero exit on failure

- **GIVEN** any failing check
- **WHEN** the gate command runs
- **THEN** it produces output describing the failure and
  exits with code 1

#### Scenario: Exit code 2 on internal error

- **GIVEN** a filesystem permission error prevents
  reading artifacts
- **WHEN** the gate command runs
- **THEN** it writes an error message to stderr and
  exits with code 2

### Requirement: Directory flag

**ID:** GATE-FR-010

The `gate` command MUST accept a `--dir` flag to specify
the project root directory. The default MUST be the
current working directory. The `--dir` value MUST be
validated before any checks execute:

- The path MUST be canonicalized using `filepath.Abs`
  and `filepath.EvalSymlinks`
- The path MUST exist and be a directory (not a file,
  device, or FIFO)
- Non-existent or non-directory paths MUST produce an
  error and exit with code 1

#### Scenario: Custom directory

- **GIVEN** a project at `/tmp/my-project` with spec
  artifacts
- **WHEN** the user runs `uf gate --phase specify
  --dir /tmp/my-project`
- **THEN** checks run against `/tmp/my-project`

#### Scenario: Non-existent directory

- **GIVEN** `/tmp/nonexistent` does not exist
- **WHEN** the user runs `uf gate --phase specify
  --dir /tmp/nonexistent`
- **THEN** the command exits with code 1 and an error
  message indicating the directory does not exist

#### Scenario: Path is a file, not a directory

- **GIVEN** `/tmp/somefile` is a regular file
- **WHEN** the user runs `uf gate --phase specify
  --dir /tmp/somefile`
- **THEN** the command exits with code 1 and an error
  message indicating the path is not a directory

### Requirement: Testability pattern

**ID:** GATE-FR-011

The gate command MUST follow the testable CLI pattern:
a `gateParams` struct with `io.Writer` injection for
both stdout and stderr, and a `runGate(params)` function
separate from cobra wiring. All check functions MUST
accept the project directory and the run's resolved
change context, and return a structured result without
side effects. Git-dependent checks (`checkNotOnMain`)
MUST be tested using `git init` inside `t.TempDir()` to
create isolated git repositories.

#### Scenario: Unit test with temp directory (pass)

- **GIVEN** a test using `t.TempDir()` with scaffolded
  proposal.md and design.md
- **WHEN** `runGate` is called with `phase: "plan"`
  targeting that directory
- **THEN** the function returns without error and output
  contains passing check results

#### Scenario: Unit test with temp directory (fail)

- **GIVEN** a test using `t.TempDir()` with no artifacts
- **WHEN** `runGate` is called with `phase: "implement"`
  targeting that directory
- **THEN** the function returns an error and output
  contains failing check results

### Requirement: Branch-aware artifact resolution

**ID:** GATE-FR-012

Artifact checks MUST resolve the active change from the
current git branch before probing artifacts:

- When the branch maps to a change (`opsx/<name>` under
  `openspec/changes/`, `NNN-<name>` under `specs/`), checks
  MUST validate that change only; artifacts of another
  change MUST NOT satisfy the active change's checks.
- `quick/<name>` branches MUST resolve to no change: no
  spec artifact can satisfy their checks, because the quick
  tier carries no spec artifacts.
- On branches without a change mapping, checks MUST fall
  back to the first alphabetical change directory across
  both layouts, resolved once for the whole run: artifacts
  of different changes MUST NOT be combined to satisfy a
  phase.
- The resolution MUST be computed once per run and shared
  by every check in that run, and the report's `branch`
  provenance field MUST reflect the branch the checks
  resolved against.
- When git reports a detached HEAD (the default checkout
  state on CI runners), the resolver MUST recover the
  branch from CI-provided refs (`GITHUB_HEAD_REF`, then
  `GITHUB_REF_NAME`) before applying the fallback.

#### Scenario: Branch-mapped change is strict

- **GIVEN** `openspec/changes/aaa-other/` holds complete
  implement artifacts and the branch is `opsx/zzz-active`,
  whose change has only `proposal.md`
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the plan-exists, tasks-exists, and
  coverage-strategy checks fail and no check message
  references `aaa-other`

#### Scenario: Quick branch never passes artifact checks

- **GIVEN** `openspec/changes/some-change/proposal.md`
  exists and the branch is `quick/some-change`
- **WHEN** the user runs `uf gate --phase specify`
- **THEN** the spec-exists check fails with an explanation
  that quick branches carry no spec artifacts

#### Scenario: Detached HEAD recovers the CI branch

- **GIVEN** the repository is in a detached HEAD state and
  `GITHUB_HEAD_REF` is set to `opsx/zzz-active`
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** checks resolve artifacts from
  `openspec/changes/zzz-active/` only

#### Scenario: Unmapped branch falls back alphabetically

- **GIVEN** the branch maps to no change and
  `openspec/changes/aaa-first/proposal.md` exists
- **WHEN** the user runs `uf gate --phase specify`
- **THEN** the spec-exists check passes against `aaa-first`

#### Scenario: Unmapped branch does not combine artifacts across changes

- **GIVEN** the branch maps to no change,
  `openspec/changes/aaa-first/` holds `proposal.md` and
  `tasks.md` but no `design.md`, and
  `openspec/changes/bbb-second/design.md` exists
- **WHEN** the user runs `uf gate --phase implement`
- **THEN** the `plan-exists` and `coverage-strategy`
  checks fail against `aaa-first` and no check message
  references `bbb-second`

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
