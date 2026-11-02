<!-- capability: uf-finale-ci-poll
     change: fix-uf-finale-ci-poll-gh-schema
     originating_issue: 895 -->

Requirement numbering: FCP-FR-001 through FCP-FR-005 are
allocated for this change. FCP-FR-006 through FCP-FR-015 are
reserved for future uf-finale-ci-poll changes. The
`uf_finale_ci_poll` output contract, 30-second poll cadence,
`timeoutMinutes` default, and `CI_POLL_TIMEOUT` /
`GH_CLI_ERROR` envelopes are governed by the
`uf_finale_ci_poll tool` requirement in the sibling
capability delta
`openspec/changes/finale-token-reduction/specs/plugin-tools.md`;
this delta refines only the gh wire schema, exit-code
tolerance, no-checks and not-found detection, and bucket
classification. Requirement ids MUST stay inside the
allocated range.

Archive ordering: this change MUST archive after
finale-token-reduction so its plugin-tools
`uf_finale_ci_poll tool` requirement materializes first.
FCP-FR-001..FCP-FR-003 refine that requirement's gh
invocation and classification clauses; FCP-FR-004 preserves
the output contract it governs. The sibling delta is not
rewritten.

## ADDED Requirements

### Requirement: Exit-Code-Tolerant gh Checks Invocation

**ID: FCP-FR-001** This requirement MUST be satisfied.

The `uf_finale_ci_poll` tool's polling loop MUST invoke
`gh pr checks <N> --json name,state,bucket,link` — only fields
verified against gh 2.91.0 — through an argument-array process
spawn (no shell string interpolation). Each invocation MUST run
in the tool's registered working directory — the `cwd` the plugin
binds at registration (`input.worktree || input.directory`) — so
the poll resolves the same repository as the `gh pr view` merge
pre-check and the `.github/workflows/` probe; it MUST NOT inherit
the plugin host process's working directory, which can differ from
the registered one. Each invocation MUST be
bounded by a per-invocation subprocess timeout of at most 30
seconds (`AbortSignal.timeout(30_000)`, matching
review-dispatch's `GH_CLI_TIMEOUT_MS`) so a hung gh process
cannot wedge the poll loop or prevent `CI_POLL_TIMEOUT` from
firing. Verification of the numeric bound: the test suite
verifies the mechanism only — every checks invocation receives
a real `AbortSignal` and an aborted invocation classifies as
`GH_CLI_ERROR` (FCP-FR-005). The 30-second numeric value is
pinned only by the module-local `GH_CLI_TIMEOUT_MS` constant
(`30_000`) in the plugin source the signal derives from; no
test asserts the constant's value. A wall-clock duration
assertion is deliberately not used because
`AbortSignal.timeout`'s deadline is not observable from the
signal object. The invocation MUST NOT throw on a non-zero gh
exit code: on ANY exit code, stdout that parses as a JSON array
MUST be treated as check-status data, and stdout that does not
parse as a JSON array MUST be treated as a fetch failure
returning a `GH_CLI_ERROR` failure envelope that carries gh's
stderr. FCP-FR-002's zero-checks and PR-not-found markers are
matched against stderr and take precedence over this generic
non-array rule; stdout that parses as a JSON array remains
data on any exit code regardless of stderr. The `prNumber`
runtime validation (integer, 1..999999) MUST be preserved
unchanged.

#### Scenario: Pending checks return data on exit 8

- **GIVEN** PR #42 has one check with `state: "PENDING"` and
  `bucket: "pending"` and `gh pr checks` exits 8 with the JSON
  array on stdout
- **WHEN** `uf_finale_ci_poll` is called with `{prNumber: 42}`
- **THEN** the tool treats the stdout as check-status data and
  continues polling rather than returning `GH_CLI_ERROR`

#### Scenario: Failing checks return data on exit 1

- **GIVEN** PR #42 has one check with `state: "FAILURE"` and
  `bucket: "fail"` and `gh pr checks` exits 1 with the JSON
  array on stdout
- **WHEN** `uf_finale_ci_poll` is called with `{prNumber: 42}`
- **THEN** the tool treats the stdout as check-status data and
  returns `{status: "fail", ...}` rather than `GH_CLI_ERROR`

#### Scenario: A hung gh invocation is bounded by a timeout

- **GIVEN** a `gh pr checks` invocation that never exits and
  the per-invocation subprocess timeout (at most 30 seconds)
  elapses
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the invocation is aborted and classified as a
  `GH_CLI_ERROR` failure rather than blocking the poll loop
  indefinitely

#### Scenario: Unparseable stdout fails on any exit code

- **GIVEN** `gh pr checks` exits with any code (including 0)
  with stdout that does not parse as a JSON array and stderr
  that does not match a FCP-FR-002 marker
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns a `GH_CLI_ERROR` failure envelope
  carrying gh's stderr

#### Scenario: Unknown JSON fields are never requested

- **GIVEN** gh 2.91.0 rejects `conclusion` and `detailsUrl` as
  `gh pr checks --json` fields
- **WHEN** the tool's gh invocation is inspected
- **THEN** the requested field list is exactly
  `name,state,bucket,link`

#### Scenario: Checks poll binds to the tool's registered cwd

- **GIVEN** the tool is registered with a working directory that
  differs from the plugin host process's working directory
- **WHEN** the polling loop invokes `gh pr checks`
- **THEN** the invocation is spawned with the tool's registered
  working directory as its `cwd`, so repository resolution
  matches the `gh pr view` pre-check

### Requirement: Checkless-PR and PR-Not-Found Detection

**ID: FCP-FR-002** This requirement MUST be satisfied.

The tool MUST recognize gh's verified zero-checks shape — exit 1
with empty stdout and a stderr matching /no checks reported/i —
and route it to the existing `no-checks-workflows-exist` or
`no-checks-no-workflows` branches exactly as a parsed empty
checks array would. The tool MUST recognize gh's PR-not-found
stderr — gh 2.91.0's only verified unresolvable-PR shape,
matching /could not resolve to a pullrequest/i — and return a
structured failure with code `PR_NOT_FOUND` and `retryable:
false`, naming the PR and carrying gh's matched stderr, instead
of a JSON parse error or a generic `GH_CLI_ERROR`. The marker
MUST NOT match looser not-found phrasing: an unrelated gh
failure whose stderr incidentally contains phrases like
"not found" MUST fall through to the retryable `GH_CLI_ERROR`
path (which echoes stderr), preserving retry semantics for
transient errors.

#### Scenario: Checkless PR with workflows configured

- **GIVEN** PR #42 has no checks, `.github/workflows/` exists,
  and `gh pr checks` exits 1 with empty stdout and
  "no checks reported" on stderr
- **WHEN** `uf_finale_ci_poll` is called with `{prNumber: 42}`
- **THEN** the tool returns
  `{status: "no-checks-workflows-exist", failedChecks: [],
  workflowsConfigured: true}` with the existing skipReason

#### Scenario: Checkless PR without workflows

- **GIVEN** PR #42 has no checks, no `.github/workflows/`
  directory, and `gh pr checks` exits 1 with empty stdout and
  "no checks reported" on stderr
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns
  `{status: "no-checks-no-workflows", workflowsConfigured: false}`

#### Scenario: PR number does not resolve

- **GIVEN** `gh pr checks` exits non-zero with empty stdout and
  "Could not resolve to a PullRequest with the number of 999999"
  on stderr
- **WHEN** `uf_finale_ci_poll` is called with `{prNumber: 999999}`
- **THEN** the tool returns `{error: true, code:
  "PR_NOT_FOUND", retryable: false, message: ...}`
  identifying the unresolved PR and carrying gh's matched
  stderr, not a JSON parse error

#### Scenario: Unrelated not-found stderr stays retryable

- **GIVEN** `gh pr checks` exits non-zero with empty stdout and
  a stderr containing "not found" that does not match the
  verified PR-not-found shape
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns a retryable `GH_CLI_ERROR`
  carrying gh's stderr, not a non-retryable `PR_NOT_FOUND`

### Requirement: Bucket-Based Poll Classification

**ID: FCP-FR-003** This requirement MUST be satisfied.

Poll-completion and pass/fail derivation MUST key off gh's
authoritative `bucket` field, not off `state` string matching.
While any check buckets as `pending`, the tool MUST keep polling
at the existing 30-second interval. When no check buckets as
`pending`: checks bucketing as `fail` or `cancel` MUST populate
`failedChecks` (preserving the existing behavior where a
CANCELLED check counts as failed); checks bucketing as `pass` or
`skipping` MUST NOT populate `failedChecks`. A bucket value
outside `pass`, `fail`, `pending`, `skipping`, and `cancel` MUST
classify fail-closed: the check MUST populate `failedChecks`
(so vocabulary drift surfaces immediately instead of exhausting
the poll timeout).

#### Scenario: Pending bucket keeps polling

- **GIVEN** the first poll returns one `bucket: "pending"` check
  and a later poll returns all checks with `bucket: "pass"`
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool waits through the pending poll and
  ultimately returns `{status: "pass", failedChecks: []}`

#### Scenario: Fail bucket populates failedChecks

- **GIVEN** all checks are terminal and one has
  `bucket: "fail"` with `state: "FAILURE"`
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns `{status: "fail", ...}` with that
  check's name in `failedChecks`

#### Scenario: Cancel bucket counts as failed

- **GIVEN** all checks are terminal and one has
  `bucket: "cancel"` with `state: "CANCELLED"`
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns `{status: "fail", ...}` with that
  check in `failedChecks`

#### Scenario: Skipping bucket passes

- **GIVEN** all checks are terminal with buckets `pass` and
  `skipping`
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns `{status: "pass", failedChecks: []}`

#### Scenario: Unknown bucket classifies fail-closed as failed

- **GIVEN** a terminal check has `state: "NEUTRAL"` and an
  unrecognized `bucket: "blocked"`
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns `{status: "fail", ...}` with that
  check in `failedChecks` and its state value surfaced in the
  entry's conclusion field

### Requirement: Output Contract Preservation

**ID: FCP-FR-004** This requirement MUST be satisfied.

The tool's output contract MUST remain unchanged: `status` (one
of `pass`, `fail`, `conflict`, `no-checks-workflows-exist`,
`no-checks-no-workflows`), `failedChecks`, `mergeable`,
`workflowsConfigured`, and `skipReason` keep their names and
types. Each `failedChecks` entry MUST keep the `{name,
conclusion, url}` shape, with `conclusion` sourced from the
check's `state` and `url` from the check's `link`. The
`gh pr view` merge-conflict pre-check, the `CI_POLL_TIMEOUT`
failure envelope, and the `INVALID_PR_NUMBER` validation MUST be
preserved unchanged. The new `PR_NOT_FOUND` failure envelope
(FCP-FR-002) is the only added failure code.

#### Scenario: failedChecks entry shape

- **GIVEN** a failing check `{name: "CI / test", state:
  "FAILURE", bucket: "fail", link: "https://…/job/1"}`
- **WHEN** `uf_finale_ci_poll` returns `{status: "fail"}`
- **THEN** `failedChecks[0]` equals
  `{name: "CI / test", conclusion: "FAILURE",
  url: "https://…/job/1"}`

#### Scenario: Conflict pre-check unchanged

- **GIVEN** `gh pr view <N> --json mergeable,mergeStateStatus`
  reports `mergeable: "CONFLICTING"`
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns `{status: "conflict", failedChecks:
  [], mergeable: false, workflowsConfigured: true}` without
  invoking `gh pr checks`

### Requirement: Regression Coverage for gh Schema and Exit Codes

**ID: FCP-FR-005** This requirement MUST be satisfied.

The test suite MUST include regression coverage that fails
against the pre-fix implementation: the golden-output-shape
command pin MUST assert the `name,state,bucket,link` field list
and the invocation's `cwd` binding to the tool's registered
working directory (FCP-FR-001),
and unit tests MUST exercise mocked gh process output for exit 8
with a JSON array, exit 1 with a JSON array, exit 1 with empty
stdout plus the no-checks stderr marker, the PR-not-found stderr,
an unrelated not-found stderr classifying as retryable
`GH_CLI_ERROR` (FCP-FR-002), the unknown-bucket classification,
and a simulated hung invocation bounded by the per-invocation
timeout. All gh process
output MUST be mocked — tests MUST NOT require a gh installation,
network access, or a live PR. The `.opencode/` suite's 94 percent
statements coverage floor MUST remain satisfied without modifying
it.

#### Scenario: Golden command pin matches the verified schema

- **GIVEN** the pre-fix implementation requests
  `--json name,state,conclusion,detailsUrl`
- **WHEN** the golden-output-shape suite runs
- **THEN** the test fails, because the pinned command string is
  `gh pr checks 43 --json name,state,bucket,link`

#### Scenario: Exit-code scenarios run without gh installed

- **GIVEN** the gh process is mocked with exit codes 8 and 1 and
  fixture stdout/stderr
- **WHEN** the uf_finale_ci_poll unit tests run
- **THEN** every scenario in FCP-FR-001 through FCP-FR-003 is
  verified with no gh binary invocation

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
