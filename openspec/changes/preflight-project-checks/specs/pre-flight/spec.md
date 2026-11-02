<!-- capability: pre-flight
     change: preflight-project-checks -->

Requirement numbering: this change allocates PF-FR-005
through PF-FR-021 for the seventeen requirements below.
PF-FR-001 through PF-FR-004 are reserved for the four
soft-gate requirements already drafted against the
pre-flight capability in the unarchived
openspec/changes/ci-causality-soft-gate change (older
bare `### FR-NNN:` heading style); reserving the leading
block keeps the ranges collision-free when both changes
archive into openspec/specs/pre-flight/.

## ADDED Requirements

### Requirement: Declared checks file

**ID: PF-FR-005** This requirement MUST be satisfied.

Pre-flight MUST read project-declared checks from
`.uf/preflight.yaml`, which MUST hold `version: 1` and an
optional `checks` list. Each check MUST have a `name` matching
`^[a-z0-9][a-z0-9-]*$` that is unique in the file, and a `run`
path. Unknown keys MUST be rejected.

#### Scenario: Absent configuration is a no-op

- **GIVEN** a repository without `.uf/preflight.yaml`
- **WHEN** `uf_preflight_declared_checks` runs in any mode
- **THEN** it returns status `ok` with `config_present: false`,
  no checks, verdict `PASS`, and an empty `rendered` string

#### Scenario: Duplicate names rejected

- **GIVEN** two checks with the same `name`
- **WHEN** the tool runs
- **THEN** it returns `INVALID_PREFLIGHT_CONFIG` with a
  duplicate-name issue

### Requirement: Optional check fields

**ID: PF-FR-006** This requirement MUST be satisfied.

A check MAY set the following fields; when a field is omitted the
parser MUST apply its default:
- `args`: literal strings, default empty;
- `requires`: bare binary names, default empty;
- `paths`: diff-scope patterns, default empty;
- `gate`: `required` or `advisory`, default `required`;
- `timeout_seconds`: a positive integer of at most 3600,
  default 600.

#### Scenario: Defaults applied

- **GIVEN** a check that declares only `name` and `run`
- **WHEN** the configuration is parsed
- **THEN** the check is `required`, always in scope, has no
  arguments or prerequisites, and times out after 600 seconds

### Requirement: Run is a single executable path

**ID: PF-FR-007** This requirement MUST be satisfied.

`run` MUST be one repository-relative executable path. It MUST
NOT be absolute, MUST NOT contain a `..` segment, and MUST NOT
contain whitespace or shell metacharacters.

#### Scenario: Shell line rejected

- **GIVEN** a check whose `run` is `make lint && echo ok`
- **WHEN** the tool runs
- **THEN** it returns `INVALID_PREFLIGHT_CONFIG` naming
  `checks.0.run`

#### Scenario: Escaping path rejected

- **GIVEN** a check whose `run` is `../other/script`
- **WHEN** the tool runs
- **THEN** it returns `INVALID_PREFLIGHT_CONFIG`

### Requirement: Invalid configuration fails closed

**ID: PF-FR-008** This requirement MUST be satisfied.

When `.uf/preflight.yaml` is present but is not valid YAML or
does not match the schema, the tool MUST return
`INVALID_PREFLIGHT_CONFIG` listing every issue with its path, MUST
NOT run any declared check, and pre-flight MUST report FAIL.

#### Scenario: Typo does not disable the gate

- **GIVEN** a check with the misspelled key `gatee: advisory`
- **WHEN** pre-flight runs
- **THEN** pre-flight reports FAIL with the validation message
  and no declared check executes

### Requirement: Execution without a shell

**ID: PF-FR-009** This requirement MUST be satisfied.

The tool MUST run each check as `<root>/<run>` from the
repository root, with `args` passed as literal argv and no shell.
The child environment MUST be the process environment plus
`UF_PREFLIGHT_PROJECT_ROOT` set to the repository root.

#### Scenario: Arguments are literal

- **GIVEN** a check with `args: ["$HOME", "a b"]`
- **WHEN** it runs
- **THEN** the executable receives exactly two arguments,
  `$HOME` unexpanded and `a b`

### Requirement: Unavailable checks skip

**ID: PF-FR-010** This requirement MUST be satisfied.

A check MUST be reported `SKIP — not available` (listing what is
missing) when a `requires` binary is not on PATH or the `run`
path is not an executable file. An unavailable check MUST NOT
affect the verdict.

#### Scenario: Missing binary skips

- **GIVEN** a check with `requires: [ripwire]` and no
  `ripwire` on PATH
- **WHEN** the tool runs
- **THEN** the check is `SKIP — not available: ripwire`

#### Scenario: Non-executable run path skips

- **GIVEN** a check whose `run` file lacks the execute bit
- **WHEN** the tool runs
- **THEN** the check is `SKIP — not available` naming the path

### Requirement: Diff-scope filtering

**ID: PF-FR-011** This requirement MUST be satisfied.

A check that declares `paths` MUST be reported
`SKIP — no in-scope files` when no file in the branch diff
matches. A pattern ending in `/` matches by prefix; any other
pattern matches by suffix, with a leading `*` ignored. A check
without `paths` MUST always be in scope.

#### Scenario: Out-of-scope check skips

- **GIVEN** a check with `paths: [scripts/]` and a branch diff
  that touches only `docs/`
- **WHEN** the tool runs
- **THEN** the check is `SKIP — no in-scope files`

### Requirement: Scope fallbacks match Phase 2a

**ID: PF-FR-012** This requirement MUST be satisfied.

When there is no base ref, the base has diverged, the clone is
shallow, `git diff` fails, or the diff is empty on a feature
branch, every check MUST be treated as in scope, and the
returned `scope.fallback_reason` MUST name the cause.

#### Scenario: Shallow clone runs everything

- **GIVEN** a shallow clone and a check with `paths`
- **WHEN** the tool runs
- **THEN** the check runs and `scope.fallback_reason` is
  `shallow clone`

### Requirement: Check status mapping

**ID: PF-FR-013** This requirement MUST be satisfied.

A run that exits 0 within its timeout MUST pass; any other exit,
or a timeout, MUST fail. A `required` check MUST map to `PASS` or
`FAIL`, and an `advisory` check to `ADVISORY PASS` or
`ADVISORY FAIL`. Captured output MUST be truncated to its last
16 KiB. A check killed for exceeding the 64 MiB output budget
MUST fail as an ordinary exit with an output-limit annotation,
never as a timeout.

#### Scenario: Output flood is not a timeout

- **GIVEN** a check that writes past the output budget
- **WHEN** it runs
- **THEN** it fails with `timed_out: false`, exit 1, and output
  ending with an output-limit notice

#### Scenario: Timeout fails

- **GIVEN** a required check that exceeds `timeout_seconds`
- **WHEN** the tool runs
- **THEN** the check is `FAIL` with `timed_out: true`, and its
  output ends with a timed-out notice

### Requirement: Declared checks run in every mode

**ID: PF-FR-014** This requirement MUST be satisfied.

Declared checks MUST run locally in `hard-gate`, `ci-aware` and
`soft-gate`. They MUST NOT be mapped to CI checks.

#### Scenario: ci-aware still runs declared checks

- **GIVEN** `ci-aware` mode and a declared required check
- **WHEN** pre-flight runs
- **THEN** the declared check executes locally

### Requirement: Hard-gate and ci-aware verdict

**ID: PF-FR-015** This requirement MUST be satisfied.

In `hard-gate` and `ci-aware`, the declared verdict MUST be
`FAIL` exactly when at least one `required` check is `FAIL`.
Advisory results MUST NOT affect the verdict in any mode.

#### Scenario: Advisory failure never blocks

- **GIVEN** an advisory check that exits non-zero and no
  required failures
- **WHEN** the tool runs in any mode
- **THEN** the check is `ADVISORY FAIL` and the declared
  verdict is `PASS`

### Requirement: Soft-gate baseline classification

**ID: PF-FR-016** This requirement MUST be satisfied.

In `soft-gate`, when required checks fail, the tool MUST create
a detached git worktree at the merge-base. It MUST re-run each
failing required check with the worktree as cwd, using the
primary checkout's executable and `UF_PREFLIGHT_PROJECT_ROOT`,
and MUST remove the worktree afterwards. A baseline pass MUST
classify the failure `branch-caused`; a baseline failure MUST
classify it `pre-existing`.

#### Scenario: Pre-existing failure

- **GIVEN** a required check that fails on the branch and at
  the merge-base
- **WHEN** the tool runs in `soft-gate`
- **THEN** the check's causality is `pre-existing` and
  `baseline_method` is `worktree`

#### Scenario: Stealth scripts absent from the baseline

- **GIVEN** a check whose `run` is in a git-excluded
  `.uf/scripts/` tree
- **WHEN** the baseline runs in the merge-base worktree
- **THEN** the script runs from the primary checkout, so the
  result reflects the merge-base content rather than the
  script's absence

### Requirement: Soft-gate baseline unavailable

**ID: PF-FR-017** This requirement MUST be satisfied.

When no merge-base is available or the worktree cannot be
created, every failure MUST be classified `unknown`,
`baseline_method` MUST be `unavailable`, and any temporary
directory MUST be removed.

#### Scenario: Worktree creation fails

- **GIVEN** `git worktree add` fails
- **WHEN** the tool runs in `soft-gate` with a required failure
- **THEN** the temp directory is removed and the failure is
  `unknown`

### Requirement: Soft-gate verdict

**ID: PF-FR-018** This requirement MUST be satisfied.

In `soft-gate`, the declared verdict MUST be `FAIL` exactly when
a failure is `branch-caused` or `unknown`.

#### Scenario: Only pre-existing failures pass

- **GIVEN** every required failure is `pre-existing`
- **WHEN** the tool runs in `soft-gate`
- **THEN** the declared verdict is `PASS`

#### Scenario: Branch-caused failure blocks

- **GIVEN** a required check that fails on the branch and
  passes at the merge-base
- **WHEN** the tool runs in `soft-gate`
- **THEN** the declared verdict is `FAIL`

### Requirement: Tool-rendered result rows

**ID: PF-FR-019** This requirement MUST be satisfied.

The tool MUST return a `rendered` string with one Execution
Results row per declared check whose columns match the mode's
table: `| name | command | exit | status |`, plus a trailing
`causality |` column in `soft-gate` only. `—` MUST stand in for
a missing exit code or causality. The command MUST be a code
span that its own backticks cannot terminate, and `|` and
newlines MUST be escaped.

#### Scenario: Columns follow the mode

- **GIVEN** a failing required check
- **WHEN** the tool runs in `hard-gate` or `ci-aware`
- **THEN** its row has four columns and no causality cell

#### Scenario: Backticks stay inside the code span

- **GIVEN** a check whose command contains a backtick
- **WHEN** the row is rendered
- **THEN** the command is fenced with a longer backtick run, so
  the cell renders as one literal code span

#### Scenario: Output shape is pinned

- **GIVEN** the golden output-shape test suite
- **WHEN** the tool returns success or `INVALID_PREFLIGHT_CONFIG`
- **THEN** the envelope keys and row format match the golden
  expectations

### Requirement: Phase 2b is one tool call

**ID: PF-FR-020** This requirement MUST be satisfied.

Pre-flight Phase 2b MUST call `uf_preflight_declared_checks`
once with the consuming command's mode. It MUST insert
`rendered` into the Phase 5 table unchanged, MUST include each
non-PASS check's
`output`, and MUST NOT re-run, CI-map or re-classify declared
checks.

#### Scenario: Results reach divisors unchanged

- **GIVEN** a soft-gate run with a pre-existing failure and an
  unavailable advisory check
- **WHEN** pre-flight assembles `pre_flight_results`
- **THEN** both rows appear exactly as rendered, along with the
  failing check's output

### Requirement: Combined pre-flight verdict

**ID: PF-FR-021** This requirement MUST be satisfied.

The overall pre-flight result MUST be `FAIL` if the declared
verdict or the detected-tool verdict is `FAIL`.

#### Scenario: Declared failure fails pre-flight

- **GIVEN** detected tools pass and the declared verdict is
  `FAIL`
- **WHEN** pre-flight reports
- **THEN** the overall result is `FAIL`
