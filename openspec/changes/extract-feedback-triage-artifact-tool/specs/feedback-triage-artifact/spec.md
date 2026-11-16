<!-- capability: feedback-triage-artifact
     change: extract-feedback-triage-artifact-tool
     requirement ids: FTA-FR-001 to FTA-FR-006 allocated here -->

## ADDED Requirements

### Requirement: One Call Renders and Writes the Artifact

**ID: FTA-FR-001** This requirement MUST be satisfied.

The uf-workflow plugin MUST register a
`write_feedback_triage_artifact` tool whose input carries the PR
identity (`pr_number`, `pr_url`, `branch`), the artifact context
(`commit`, `backlog_item_id`), one triage record per feedback
item, and optionally the `post_feedback_replies` success data.
In one call the tool MUST compute the summary counts, derive the
round, render the standard feedback-triage envelope, validate
the rendered artifact, and write it to
`.uf/artifacts/feedback-triage/pr-<N>-round-<M>.json`.

#### Scenario: Artifact written for a decided triage

- **GIVEN** three decided items (one `accept`, one `file`, one
  `reply`) and the `post_feedback_replies` success data
- **WHEN** `write_feedback_triage_artifact` is called once
- **THEN** the file
  `.uf/artifacts/feedback-triage/pr-<N>-round-<M>.json` MUST
  exist with all three items and the computed summary
- **AND** the result MUST report the path, round, and byte count

#### Scenario: Envelope fields

- **GIVEN** a successful call
- **WHEN** the artifact is read back
- **THEN** the envelope MUST carry `hero: "cobalt-crush"`,
  `version: "1.0.0"`, an ISO 8601 `timestamp`,
  `artifact_type: "feedback-triage"`,
  `schema_version: "1.1.0"`, `context`, and `payload`

#### Scenario: No replies result

- **GIVEN** the author skipped posting, so no
  `post_feedback_replies` result exists
- **WHEN** the tool is called without the replies data
- **THEN** the artifact MUST still be written
- **AND** every item's `filed_issue_number` and
  `filed_issue_url` MUST be null

### Requirement: Round Numbering

**ID: FTA-FR-002** This requirement MUST be satisfied.

The tool MUST scan `.uf/artifacts/feedback-triage/` for files
matching `pr-<N>-round-<M>.json` for the same PR number and set
the new round to the highest matching `M` plus one; when no file
matches, the round MUST be one. Names that do not parse MUST be
ignored. The round MUST NOT be derived from the file count, so
gaps left by deleted artifacts do not reuse a round number.

#### Scenario: Gaps do not reuse a round

- **GIVEN** rounds 1 and 3 exist for PR 42 and round 2 was
  deleted
- **WHEN** the tool writes the next artifact for PR 42
- **THEN** the new round MUST be 4

#### Scenario: First round

- **GIVEN** no artifact exists for PR 42
- **WHEN** the tool writes the artifact
- **THEN** the round MUST be 1

### Requirement: FILE and REPLY Routing From the Replies Result

**ID: FTA-FR-003** This requirement MUST be satisfied.

When the `post_feedback_replies` success data is supplied, the
tool MUST join each triage record to its reply outcome by
comment key (`<comment_kind>:<comment_id>`), failing validation
when the joined decisions disagree. A `file` record's
`filed_issue_number`/`filed_issue_url` MUST come from the joined
reply item, null when its issue failed; every other record MUST
carry null for both. `summary.filed` and `summary.replied` MUST
count the `file` and `reply` decisions, never `rejected`.

#### Scenario: FILE record carries its issue

- **GIVEN** a `file` record whose reply outcome reports issue
  501 at `https://github.com/o/r/issues/501`
- **WHEN** the artifact is rendered
- **THEN** the record MUST carry `filed_issue_number: 501` and
  that URL
- **AND** `summary.filed` MUST count it

#### Scenario: Decision mismatch fails closed

- **GIVEN** a triage record with decision `reply` whose joined
  reply outcome carries decision `file`
- **WHEN** the tool is called
- **THEN** validation MUST fail with nothing written

#### Scenario: Failed issue records nulls

- **GIVEN** a `file` record whose issue creation failed, so the
  reply outcome reports a null issue number
- **WHEN** the artifact is rendered
- **THEN** the record MUST carry null `filed_issue_number` and
  `filed_issue_url`
- **AND** the record MUST still count under `summary.filed`

### Requirement: Validation Before Atomic Write

**ID: FTA-FR-004** This requirement MUST be satisfied.

The rendered payload MUST pass a validation mirroring
`schemas/feedback-triage/v1.1.0.schema.json` before anything is
written; a validation failure MUST fail closed with no file. The
write MUST be atomic — a temp file first, then a rename —
creating files with `0o600` under directories with `0o700`, and
MUST refuse symlinked artifact paths.

#### Scenario: Invalid payload writes nothing

- **GIVEN** a triage record whose `decision` is not one of the
  1.1.0 enum values
- **WHEN** the tool is called
- **THEN** the tool MUST fail with a structured validation error
- **AND** no file MUST exist in the artifact directory

#### Scenario: Validation pinned to the samples

- **GIVEN** the checked-in feedback-triage samples
- **WHEN** they run through the tool's payload validation
- **THEN** the valid 1.1.0 and 1.0.0 samples MUST pass and the
  `invalid-*` samples MUST fail

#### Scenario: Symlink refusal

- **GIVEN** the target artifact path is a symlink
- **WHEN** the tool is called
- **THEN** the call MUST fail without following the symlink

### Requirement: Terminal Progress Cleanup and Structured Result

**ID: FTA-FR-005** This requirement MUST be satisfied.

On a successful write the tool MUST clear the
`uf.address-feedback` progress state through the same path the
`clear_progress` tool uses; a clear failure MUST be reported as
a warning, not a call failure. The result MUST return
`pr_number`, `path`, `round`, `bytes`, the summary counts,
`progress_cleared`, `warnings`, and a rendered `summary` line
the command prints verbatim. Whole-call failures MUST use the
structured error envelope (`status: "error"`, `error: true`,
`code`, `message`, `retryable`).

#### Scenario: Clear failure is a warning

- **GIVEN** the artifact was written and the progress-state
  deletion then fails
- **WHEN** the tool returns
- **THEN** the call MUST succeed
- **AND** `progress_cleared` MUST be false with the failure in
  `warnings`

#### Scenario: Error envelope

- **GIVEN** the artifact directory cannot be created
- **WHEN** the tool is called
- **THEN** the result MUST carry `status: "error"`, `error:
  true`, a `code`, a `message`, and `retryable`

### Requirement: Section 4.7 Calls the Tool

**ID: FTA-FR-006** This requirement MUST be satisfied.

Section 4.7 of `.opencode/commands/uf.address-feedback.md` MUST
instruct one `write_feedback_triage_artifact` call with the
parameter mapping — triage records carrying the Section 4.6
comment identity, and the replies result when it exists —
instead of prose for the envelope render, round scan, FILE/REPLY
routing, or atomic write. The no-success Tier 2 guidance
(`tier2_unavailable`, the `author-decides` compatibility
labeling) MUST stay in prose. On an error envelope the command
MUST STOP.

#### Scenario: Prose template is gone

- **GIVEN** the rewritten Section 4.7
- **WHEN** its text is inspected
- **THEN** it MUST NOT contain the envelope JSON template or the
  hand-written round-scan and atomic-write steps
- **AND** it MUST keep the no-success Tier 2 guidance

#### Scenario: Error envelope stops the command

- **GIVEN** `write_feedback_triage_artifact` returns
  `status: "error"`
- **WHEN** the command reads the envelope
- **THEN** it MUST report the error verbatim and STOP

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
