<!-- capability: review-dispatch -->

## ADDED Requirements

### Requirement: Incomplete Final Response SHALL NOT Count as Approval

**ID: RD-FR-059** This requirement MUST be satisfied.

A reviewer run whose final response is incomplete -- final text
whose trimmed length is zero, or a final response terminated by
finish reason `length` -- SHALL NOT cast an approving vote.
Classification depends on whether the run produced accepted
structured submissions:

- No accepted submissions: the run SHALL be recorded `failed`
  with error code `child_empty_response` and `retryable: true`.
- Accepted submissions present: RD-FR-026 precedence applies.
  The run SHALL be classified `success`, its submissions SHALL
  contribute to consolidation, and its `workflow_verdict` SHALL
  be the kind-native non-approving value: `REQUEST CHANGES`
  for council and test-review, `NEEDS-CLARIFICATION` for
  triage, `AUTHOR-DECIDES` for feedback.

#### Scenario: Empty final text at executor boundary

- **GIVEN** a reviewer child whose extracted final text has
  trimmed length zero
- **WHEN** the executor classifies the run status
- **THEN** the executor SHALL return status `failed` with
  error code `child_empty_response` and `retryable: true`
- **AND** the result SHALL preserve `text` and `usage` for
  diagnostics

#### Scenario: Length-terminated empty output at executor

- **GIVEN** the model stream ends with finish reason `length`
  and the extracted final text is empty
- **WHEN** the executor classifies the run
- **THEN** it SHALL return status `failed` with code
  `child_empty_response` (same path as empty text)

#### Scenario: Incomplete success reclassified at dispatch

- **GIVEN** a reviewer run recorded `success` whose final
  response is incomplete (empty text from an executor-bypassing
  caller, or finish reason `length` with truncated text)
- **AND** the run harvested zero accepted structured
  submissions
- **WHEN** dispatch finalization builds the run records
- **THEN** the run SHALL be reclassified `failed` with code
  `child_empty_response`, `retryable: true`, and
  `workflow_verdict: null`
- **AND** the run SHALL be excluded from
  `successfulPersonaVotes` and counted in `run_counts.failed`

#### Scenario: Incomplete response with accepted submissions

- **GIVEN** a reviewer run with an incomplete final response
  (empty text or finish reason `length`) whose structured
  submissions were accepted
- **WHEN** dispatch finalization builds the run records
- **THEN** the run SHALL be classified `success` (restored if
  the executor failed it) and its submissions SHALL contribute
  to consolidation (RD-FR-026 precedence)
- **AND** its `workflow_verdict` SHALL be forced to the
  kind-native non-approving value: `REQUEST CHANGES` for
  council and test-review, `NEEDS-CLARIFICATION` for triage,
  `AUTHOR-DECIDES` for feedback

#### Scenario: Post-submission abort keeps RD-FR-026 rule

- **GIVEN** a reviewer session that aborts after its
  structured submissions were accepted (RD-FR-026 trigger)
- **WHEN** dispatch finalization classifies the run
- **THEN** RD-FR-026 SHALL apply unchanged: the run is
  classified `success`, its submissions contribute, and its
  verdict is derived from its findings
- **AND** this requirement SHALL NOT introduce a new failure
  classification for that case

#### Scenario: Council cannot approve on an incomplete vote

- **GIVEN** a council dispatch where one reviewer run has an
  incomplete final response
- **WHEN** the council verdict is computed
- **THEN** the incomplete run either counts in
  `run_counts.failed` (no submissions) or casts its forced
  non-approving vote (with submissions)
- **AND** no `APPROVE` or `APPROVE WITH ADVISORIES` result
  SHALL rest on the incomplete run's vote

### Requirement: Executor SHALL Surface Response Provenance

**ID: RD-FR-060** This requirement MUST be satisfied.

The executor's `child_empty_response` failure SHALL include
diagnostic provenance: provider, model, finish reason when
present, and `usage` token counts. The executor SHALL
additionally surface the finish reason as a structured field on
`success` results (`null` when unreported) so dispatch can
apply length policy to non-empty runs.

#### Scenario: Diagnostics carried on empty failure

- **GIVEN** an empty reviewer diagnosed at the executor
- **WHEN** the failure is written to the dispatch artifact
- **THEN** the error message SHALL name the condition (empty
  text or length-terminated) and retain `usage` so operators
  can see `tokens.output` and finish reason

#### Scenario: Finish reason surfaced on success

- **GIVEN** a reviewer run that returns non-empty text with a
  reported finish reason
- **WHEN** the executor returns the success result
- **THEN** the result SHALL carry the finish reason as a
  structured field, `null` when unreported, so downstream
  dispatch can enforce length policy on non-empty runs

### Requirement: Exact and Observable child_empty_response Warning

**ID: RD-FR-061** This requirement MUST be satisfied.

The dispatch SHALL emit a warning whose kind is exactly
`child_empty_response` for every run reclassified or
verdict-forced under RD-FR-059. The warning SHALL name the
agent and run id and include finish reason and token counts
when available. For a run whose final text is empty, the
dispatch SHALL emit `child_empty_response` and SHALL NOT emit
`empty_coverage_attestation` for that run, regardless of
delivery mode. The dispatch SHALL NOT emit duplicate
`child_empty_response` warnings for the same run.

#### Scenario: Exact warning kind, no duplication

- **GIVEN** a run whose final text is empty in manifest mode
- **WHEN** the dispatch artifact warnings are built
- **THEN** exactly one warning with kind
  `child_empty_response` SHALL be emitted for that run
- **AND** no `empty_coverage_attestation` warning SHALL be
  emitted for that run

#### Scenario: Operator sees the failure end-to-end

- **GIVEN** a council dispatch where one reviewer returned an
  empty final response with no accepted submissions
- **WHEN** the operator reads the council report
- **THEN** the report SHALL show the `child_empty_response`
  warning naming the reviewer, the `run_counts.failed` total,
  and a non-approving verdict

### Requirement: Bounded Retry for child_empty_response

**ID: RD-FR-062** This requirement MUST be satisfied.

`child_empty_response` SHALL carry `retryable: true` for
operators, but a dispatch operation SHALL NOT automatically
re-execute a run it has already executed: each planned run
executes at most once per dispatch call. Re-execution SHALL
occur only through DR-FR-002 partial re-dispatch or external
orchestration. When every attempt for a reviewer has failed,
consolidation SHALL follow the RD-FR-007 zero-success path: a
non-approving verdict (`INCONCLUSIVE`, or `UNAVAILABLE` when
all failures carry availability-error codes), with the
failures visible in `run_counts.failed` and warnings.

#### Scenario: No automatic re-execution within a dispatch

- **GIVEN** a reviewer fails with `child_empty_response`
  during a dispatch operation
- **WHEN** the dispatch operation completes
- **THEN** the failed run SHALL NOT have been re-executed
  within that operation

#### Scenario: Retry exhaustion is non-approving

- **GIVEN** a reviewer whose model deterministically exhausts
  its output budget, so every DR-FR-002 re-run fails with
  `child_empty_response`
- **WHEN** consolidation runs with zero successful runs
- **THEN** the council verdict SHALL be non-approving per the
  RD-FR-007 zero-success path (`INCONCLUSIVE` or
  `UNAVAILABLE`)
- **AND** `run_counts.failed` and the warnings SHALL expose
  the repeated `child_empty_response` failures
