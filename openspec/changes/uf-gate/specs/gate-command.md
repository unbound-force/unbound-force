## ADDED Requirements

### Requirement: FR-001 Gate command existence

The `unbound-force` CLI MUST provide a `gate` subcommand
that performs phase-specific constitution compliance checks.

#### Scenario: Gate command is registered

- **GIVEN** the `unbound-force` binary is built
- **WHEN** the user runs `uf gate --help`
- **THEN** help text is displayed showing usage, phase
  values, and format options

### Requirement: FR-002 Phase flag

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

### Requirement: FR-003 JSON output format

The `gate` command MUST support `--format json` for
machine-readable output. The default format MUST be
`text`. The JSON output MUST include:

- `version`: schema version string (e.g., "1.0.0")
- `producer`: tool identifier ("uf-gate")
- `timestamp`: ISO 8601 UTC timestamp of the check run
- `branch`: current git branch (empty if unavailable)
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

### Requirement: FR-004 Specify phase checks

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

### Requirement: FR-005 Plan phase checks

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

### Requirement: FR-006 Implement phase checks

When `--phase implement`, the gate MUST check:

- A spec artifact exists
- A plan/design artifact exists
- A tasks artifact exists (openspec `tasks.md` or speckit
  `tasks.md`)
- A coverage strategy is present in the plan or tasks
  artifact (Constitution IV). The check MUST detect a
  coverage strategy by matching at least one of:
  - A heading containing "coverage" (case-insensitive)
  - Body text containing any of: "coverage target",
    "unit test", "integration test", "e2e test"
  - A line matching `coverage.*%` (percentage target)

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

### Requirement: FR-007 Review phase checks

When `--phase review`, the gate MUST check:

- All task checkboxes in the tasks artifact are marked
  complete (`[x]`). The check scans tasks.md for lines
  matching `- [ ]` — if any unchecked boxes remain, the
  check fails.
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

#### Scenario: Review phase fails no review marker

- **GIVEN** tasks.md exists with all checkboxes `[x]`
  but no review marker comment
- **WHEN** the user runs `uf gate --phase review`
- **THEN** the `review-run` check fails with a message
  indicating no review has been recorded

### Requirement: FR-008 PR phase checks

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

### Requirement: FR-009 Exit code semantics

The `gate` command MUST exit with code 0 when all checks
for the requested phase pass. It MUST exit with code 1
when any check fails. It MUST exit with code 2 when an
internal error occurs (e.g., invalid arguments beyond
phase/format validation, filesystem permission denied,
unexpected failures). It MUST always produce output
(text or JSON) before exiting.

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

### Requirement: FR-010 Directory flag

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

### Requirement: FR-011 Testability pattern

The gate command MUST follow the testable CLI pattern:
a `gateParams` struct with `io.Writer` injection for
both stdout and stderr, and a `runGate(params)` function
separate from cobra wiring. All check functions MUST
accept a directory path and return a structured result
without side effects. Git-dependent checks (`checkNotOnMain`)
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

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
