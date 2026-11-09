<!-- capability: feedback-reply-tools
     change: feedback-reply-tools
     requirement ids: FRT-FR-001 to FRT-FR-005 allocated here -->

## ADDED Requirements

### Requirement: One Call Posts Replies and Resolves Threads

**ID: FRT-FR-001** This requirement MUST be satisfied.

The uf-workflow plugin MUST register a `post_feedback_replies` tool
whose input carries the PR number, the author's resolve decision,
and per item the target comment, the triage decision, the confirmed
reply text, and, for `file` items only, the issue title and body.
In one call the tool MUST file each `file` item's issue before its
reply, post each reply to the review-thread replies endpoint or the
issue-comments endpoint, map review comments to GraphQL thread node
IDs, and, when the author approved resolution, resolve the threads
of `accept`, `modify`, `file`, and `reply` review items whose
replies are posted. Untrusted text MUST reach `gh` only through
temporary files and argument arrays.

#### Scenario: Addressed threads resolve in the same call

- **GIVEN** confirmed replies for an `accept` and a `reject` review
  item and the author chose to resolve addressed threads
- **WHEN** `post_feedback_replies` is called once
- **THEN** both replies MUST be posted
- **AND** the `accept` item's thread MUST be resolved
- **AND** the `reject` item's thread MUST stay open

#### Scenario: FILE reply links its issue

- **GIVEN** a `file` item with an issue title and body
- **WHEN** the tool runs
- **THEN** the issue MUST exist before the reply is posted
- **AND** the reply MUST start with `Filed as #<issue number>: `

### Requirement: Resumable Per-Item Progress

**ID: FRT-FR-002** This requirement MUST be satisfied.

The tool MUST record each filed issue, posted reply, and resolved
thread per item in `.uf/feedback/pr-<N>/state.json` before the next
mutation, preserving the cache's other keys. A call with items
already recorded MUST NOT post or file them again and MUST resolve
approved threads a prior call left open. A failed post MUST stop
further posting in that call. A `file` item whose issue creation
fails on the first try and the retry MUST NOT be posted or
resolved, and later items MUST still run. The tool MUST record the
`reply-comments` checkpoint only when every reply is posted and
every approved thread is resolved.

#### Scenario: Partial failure resumes

- **GIVEN** a call posted item 1 and failed posting item 2 of 3
- **WHEN** the tool is called again with the same items
- **THEN** item 1 MUST NOT be posted again
- **AND** items 2 and 3 MUST be posted and their approved threads
  resolved

#### Scenario: Resolution after an earlier run

- **GIVEN** a run posted every reply but left the threads open
- **WHEN** the tool is called again with resolution approved
- **THEN** it MUST resolve the threads without posting

### Requirement: Issue Filing Tool

**ID: FRT-FR-003** This requirement MUST be satisfied.

The uf-workflow plugin MUST register a `file_feedback_issue` tool
that takes the PR number, the comment that raised the finding, a
title, and a body, and returns the issue number and URL. It MUST
retry `gh issue create` once on failure. When the feedback cache
already records an issue for the same comment, it MUST return that
issue without filing again.

#### Scenario: Second call deduplicates

- **GIVEN** `file_feedback_issue` filed issue 500 for comment 81
- **WHEN** it is called again for comment 81
- **THEN** it MUST return issue 500 with `deduplicated: true`
- **AND** it MUST NOT call `gh issue create`

### Requirement: Structured Results

**ID: FRT-FR-004** This requirement MUST be satisfied.

Both tools MUST validate input with the same schema their `args`
expose. A whole-call failure MUST return `status: "error"`,
`error: true`, `code`, `message`, and `retryable`.
`post_feedback_replies` MUST return per-item reply, issue, and
thread outcomes with any per-item error, counts, whether the
checkpoint was recorded, and a rendered `summary` line the command
prints verbatim.

#### Scenario: Partial run summary

- **GIVEN** 3 of 6 replies posted before a failure
- **WHEN** the tool returns
- **THEN** `complete` MUST be false
- **AND** `summary` MUST name the unposted items and the re-run
  command

### Requirement: Feedback-Triage 1.1.0

**ID: FRT-FR-005** This requirement MUST be satisfied.

`schemas/feedback-triage/v1.1.0.schema.json` MUST add `file` and
`reply` to the decision enum, optional `filed_issue_number` and
`filed_issue_url` item fields, and optional `filed` and `replied`
summary counts. Every 1.0.0 artifact MUST validate under 1.1.0, and
the 1.0.0 schema MUST remain. `/uf.address-feedback` MUST emit
`schema_version: "1.1.0"` and record FILE and REPLY items with the
`file` and `reply` decisions, not `reject` with a prefix.

#### Scenario: 1.0.0 artifact under 1.1.0

- **GIVEN** a valid 1.0.0 feedback-triage payload
- **WHEN** it is validated against the 1.1.0 schema
- **THEN** validation MUST pass

#### Scenario: 1.0.0 reader and a FILE decision

- **GIVEN** a 1.1.0 payload with a `file` decision
- **WHEN** it is validated against the 1.0.0 schema
- **THEN** validation MUST fail

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
