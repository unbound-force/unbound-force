<!-- capability: command-evals
     change: nightly-command-evals
     End-to-end evals of uf commands, run nightly in CI and
     locally. Design rationale: design.md (D1-D6). -->

Requirement numbering: EV-FR-001 through EV-FR-008 are allocated
for this change. EV-FR-009 through EV-FR-030 are reserved for
future command-evals changes. This is the capability's first delta;
no other range is allocated.

## ADDED Requirements

### Requirement: Pinned Isolated Runtime

**ID: EV-FR-001** This requirement MUST be satisfied.

An eval run MUST load only the opencode configuration declared under
`.eval/opencode/` plus the project snapshot: the model provider, DCP
and auto-question at pinned commits, and uf's own plugins. It MUST
NOT read the host's global opencode configuration. A pinned plugin
whose build fails MUST stop the run before any fixture starts, and
the driver MUST fail at startup when the server does not offer the
`compress` tool.

#### Scenario: Plugin build fails
- **GIVEN** the pinned DCP commit fails to build
- **WHEN** `.eval/run.sh` prepares the runtime
- **THEN** it exits non-zero before starting the driver
- **AND** it does not mark the plugin cache as ready

#### Scenario: Plugin declared but not loaded
- **GIVEN** a runtime where DCP is listed in the config but its
  tools are not registered
- **WHEN** the driver starts the opencode server
- **THEN** it fails with an error naming `compress` as missing

### Requirement: Fixture Isolation

**ID: EV-FR-002** This requirement MUST be satisfied.

Each fixture MUST run in a fresh temporary workdir holding a
snapshot of the repo's tracked and untracked, non-ignored files,
excluding `.eval/`. `external_directory` MUST be denied for every
agent outside that workdir and the temp areas. The `gh` executable
MUST be replaced by a stub that records each call and refuses it.
Every permission ask from any session MUST be rejected and recorded,
so no run waits on a human.

#### Scenario: Reviewer reads outside the workdir
- **GIVEN** a fixture run in progress
- **WHEN** any agent reads a path outside the workdir and temp areas
- **THEN** the read is denied without a permission prompt

#### Scenario: Command tries to post a review
- **GIVEN** a council fixture
- **WHEN** the command calls `gh` to post a review
- **THEN** the stub refuses the call and records its arguments

### Requirement: Unattended Answers

**ID: EV-FR-003** This requirement MUST be satisfied.

Questions a command asks MUST be answered by auto-question from the
fixture's guidance brief. A question still unanswered after 8
minutes MUST be rejected and MUST fail the run.

#### Scenario: Fix-loop limit question
- **GIVEN** the planted-defects fixture reaching the three-iteration
  limit
- **WHEN** the command asks whether to continue or stop
- **THEN** auto-question answers from the brief
- **AND** the trajectory records the question and its answer

### Requirement: Deterministic Assertions

**ID: EV-FR-004** This requirement MUST be satisfied.

Each fixture MUST declare its expectations in `assertions.yaml` as
deterministic checks on the recorded trajectory, stub calls, and
workdir state. No fixture MAY rely on a judge model. Both council
fixtures MUST assert `sessions_completed` for `^divisor-` sessions.

#### Scenario: Reviewer hits its output limit
- **GIVEN** a council fixture run in which a Divisor session ends
  with finish reason `length`
- **WHEN** assertions are evaluated
- **THEN** `sessions_completed` fails and names the session

### Requirement: Nightly Workflow

**ID: EV-FR-005** This requirement MUST be satisfied.

`.github/workflows/ci_eval.yml` MUST run every fixture on a nightly
`schedule` and on `workflow_dispatch`, one job per fixture. It MUST
run only when `github.repository` is `unbound-force/unbound-force`,
MUST declare `permissions: contents: read`, MUST NOT trigger on
`pull_request` or `pull_request_target`, and MUST pin every action
and the eval-infra driver to a full commit SHA. Each job MUST set a
timeout.

#### Scenario: Scheduled run
- **GIVEN** the nightly schedule fires on the upstream repo
- **WHEN** the workflow runs
- **THEN** one job starts per fixture directory under
  `.eval/fixtures/`

#### Scenario: Fork
- **GIVEN** a fork of the repo with the workflow file
- **WHEN** the schedule fires there
- **THEN** every job is skipped

### Requirement: Run Reporting

**ID: EV-FR-006** This requirement MUST be satisfied.

Each job MUST write its fixture's result to the job summary: status,
each assertion's result, wall time, and input, output, reasoning and
cache-read tokens. Each job MUST upload the trajectory JSON, the run
log and `runs.db` as an artifact, and MUST upload them when the
fixture fails. A job MUST fail when any of its assertions fails.

#### Scenario: Assertion fails
- **GIVEN** a fixture whose `sessions_completed` assertion fails
- **WHEN** the job finishes
- **THEN** the job fails, the summary shows the failed assertion
  with its detail, and the artifacts are uploaded

### Requirement: Secret Handling

**ID: EV-FR-007** This requirement MUST be satisfied.

The provider key MUST come only from the `ENMAAS_KEY` Actions secret
in CI or the environment locally. No script MAY print it, and no
artifact or summary MAY contain it.

#### Scenario: Missing key
- **GIVEN** `ENMAAS_KEY` is unset
- **WHEN** `.eval/run.sh` starts
- **THEN** it exits non-zero naming the missing variable, without
  starting opencode

### Requirement: Local Parity

**ID: EV-FR-008** This requirement MUST be satisfied.

`make eval` MUST run every fixture and `make eval-fast` MUST run the
argument-rejection fixtures, both through `.eval/run.sh` with the
same pins the workflow uses. Neither target MAY be a dependency of
`make check`.

#### Scenario: Local fast run
- **GIVEN** a clone with `ENMAAS_KEY` set and the toolchain
  installed
- **WHEN** a maintainer runs `make eval-fast`
- **THEN** the three argument-rejection fixtures run against the
  working tree and report pass or fail
