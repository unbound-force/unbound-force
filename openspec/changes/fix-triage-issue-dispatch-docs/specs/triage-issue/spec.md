# triage-issue Issue Triage Command — Dispatch Data-Flow Contract

## ADDED Requirements

### Requirement: Dispatch Output Documentation

**ID: FR-001** This requirement MUST be satisfied.

The command SHALL document `dispatch_review_runs` as returning a
counts-only summary line to the calling model, and SHALL route all
array data (findings, proposals, run records) through the sanctioned
`consolidate_dispatch` retrieval path. The retrieval call is
side-effect-free; on transient failure the command MUST retry it
exactly once before failing closed, and on permanent failure it MUST
fail closed without retry.

#### Scenario: Counts-only dispatch output

- **GIVEN** the command documents the output of `dispatch_review_runs`
- **WHEN** Section 2.3 describes what the calling model receives
- **THEN** it states that the caller-visible output is a one-line
  counts summary (completed, failed, skipped, findings, proposals)
- **AND** it does not claim the caller receives `findings`,
  `proposals`, or `warnings` arrays from `dispatch_review_runs`

#### Scenario: Sanctioned retrieval after dispatch

- **GIVEN** `dispatch_review_runs` has completed
- **WHEN** the command needs run records, findings, or proposals
- **THEN** it instructs exactly one `consolidate_dispatch` call
  WITHOUT provenance
- **AND** that call returns the raw payload: per-run records with
  run-recorded `workflow_verdict`, findings, advisories, run counts,
  and the top-level `proposals` array

#### Scenario: Transient retrieval failure retries once

- **GIVEN** the no-provenance `consolidate_dispatch` retrieval call
  fails transiently (timeout or intermittent tool error)
- **WHEN** the command processes the result
- **THEN** it retries the side-effect-free retrieval call exactly
  once
- **AND** if the retry succeeds, processing continues with the
  retrieved payload
- **AND** if the retry also fails, it stops, reports the failure,
  and fails closed without fabricating run records, verdicts,
  findings, or proposals
- **AND** it does not substitute child transcripts or improvised
  session-database reads for the missing payload

#### Scenario: Permanent retrieval failure fails closed

- **GIVEN** the retrieval call returns a permanent failure (a
  tool-reported error status, or a dispatch session that does not
  exist for a dispatch that executed runs)
- **WHEN** the command processes the result
- **THEN** it does not retry; it stops, reports the failure, and
  fails closed without fabricating run records, verdicts, findings,
  or proposals

#### Scenario: All-skip dispatch is a valid outcome

- **GIVEN** the dispatch summary reports zero completed and zero
  failed runs (every plan entry skipped) and no dispatch session
  exists
- **WHEN** the command processes the dispatch result
- **THEN** it records the all-skip outcome as valid: zero findings,
  zero proposals, zero persona votes, and the verdict outcome the
  no-vote semantics prescribe for zero successful runs
- **AND** it does not treat the missing dispatch session as a
  retrieval failure and does not fail closed

### Requirement: Lesson Proposal Sourcing

**ID: FR-002** This requirement MUST be satisfied.

The command SHALL source the lesson proposals array from the
no-provenance `consolidate_dispatch` payload, not from
`dispatch_review_runs`. Each proposal in that payload carries the
validated shape `{agent, information, tag, category}` as defined by
the review-dispatch plugin's versioned dispatch artifact schema
(each persisted artifact records its `schema_version`); there is no
`reference` field. Per-proposal storage via `dewey_store_learning`
remains unchanged.

#### Scenario: Proposals from consolidate payload

- **GIVEN** the retrieval payload has been fetched
- **WHEN** the command stores lesson proposals in Section 2.4
- **THEN** the proposals are read from the top-level `proposals`
  array of the retrieval payload
- **AND** each is stored via one `dewey_store_learning` call using
  the proposal's `information` and `tag`, with the proposal's
  `category` as the learning category

#### Scenario: Empty proposals array

- **GIVEN** the retrieval payload's `proposals` array is empty
- **WHEN** the command reaches Section 2.4
- **THEN** it records that no lesson proposals were submitted and
  continues without error

### Requirement: Authoritative Verdict Source

**ID: FR-003** This requirement MUST be satisfied.

Verdict consolidation (three-rule majority) SHALL use run-recorded
`workflow_verdict` values from the consolidate payload's `runs`, not
child prose. When child transcript text and run-recorded
`workflow_verdict` disagree, the run-recorded verdict MUST govern.
Category, objectivity, reasoning, and split recommendations SHALL be
obtained exclusively through the review-dispatch plugin's
`read_assessment_fields` tool: given the dispatch's run identifiers,
the tool reads each run's persisted child final text from the
dispatch store, parses the structured response contract fields, and
enforces the untrusted-input validation (Constitution V) as
executable, tested code — `category` a non-empty string of at most
128 characters; `objectivity` a string of at most 64 characters;
`reasoning` a string of at most 4096 characters; split
recommendations an array of at most 16 strings of at most 512
characters each — with every value type-checked against the
structured response contract shape and values that fail validation
skipped as unavailable for that persona. The command MUST NOT
prescribe the field extraction or validation as prose (determinism
pack DE-001) and MUST NOT read session transcripts directly. The
command's existing no-vote semantics are preserved: failed and
non-success terminal runs do not vote, a persona with no successful
run produces no persona vote, and the panel majority counts only
successful persona votes.

#### Scenario: Run-recorded verdict governs

- **GIVEN** a child transcript's textual verdict disagrees with the
  run-recorded `workflow_verdict` for that run
- **WHEN** Phase 3 consolidates persona-level and panel-level
  verdicts
- **THEN** the run-recorded `workflow_verdict` is used

#### Scenario: Assessment fields via the extraction tool

- **GIVEN** Phase 3 needs category, objectivity, reasoning, or split
  recommendations
- **WHEN** those fields are read
- **THEN** they come from a single `read_assessment_fields` call
  with the dispatch's run identifiers
- **AND** each returned value is already length-bounded,
  type-checked, and contract-validated by the tool; values that
  failed validation are returned as unavailable for that persona
- **AND** the command performs no transcript reads and no per-value
  validation of its own

#### Scenario: Missing or unparseable child text

- **GIVEN** a run's persisted child final text is missing, empty, or
  cannot be parsed as the structured response contract
- **WHEN** the extraction tool processes that run
- **THEN** the affected fields are reported as unavailable for that
  persona
- **AND** verdict resolution still proceeds from the run-recorded
  `workflow_verdict` values, which do not depend on child text

#### Scenario: All child text unavailable

- **GIVEN** every run's persisted child final text is missing or
  unparseable
- **WHEN** Phase 3 consolidates assessment fields
- **THEN** assessment fields are reported unavailable panel-wide
- **AND** verdict resolution proceeds from the run-recorded
  `workflow_verdict` values alone
- **AND** the command reports the assessment-field unavailability
  without failing closed

### Requirement: Consolidate Dispatch Dual Mode

**ID: FR-004** This requirement MUST be satisfied.

The command SHALL document `consolidate_dispatch` as dual-mode:
without provenance it is the retrieval path (no finalization side
effects); with provenance it finalizes. The exactly-once-per-
finalization-attempt bound applies only to the provenance-bearing
finalization call; the Section 2.3 retrieval call does not count
against it.

#### Scenario: Retrieval does not finalize

- **GIVEN** `consolidate_dispatch` is called without provenance
- **WHEN** the call completes
- **THEN** no finalization occurs and no finalization artifacts are
  written
- **AND** the Section 4.6 finalization call remains available and
  exactly-once per finalization attempt

#### Scenario: Finalization call unchanged

- **GIVEN** the command finalizes in Section 4.6
- **WHEN** the finalization call is made
- **THEN** it calls `consolidate_dispatch` exactly once with
  provenance `{branch, commit, workflow_id: uf.triage-issue}` as the
  only argument
- **AND** never calls `finalize_review_dispatch` directly

### Requirement: Sibling Command Proposal Sourcing

**ID: FR-005** This requirement MUST be satisfied.

The identical false claims — that `dispatch_review_runs` returns a
`proposals` array to the caller, and that proposals carry a
`reference` field — also appear in the sibling commands
`uf.review-council.md` (Section 4, Store Lesson Proposals),
`uf.address-feedback.md` (Section 2.3.7, Store Lesson Proposals),
and `uf.review-pr.md` (Section 5a, Store Lesson Proposals). Each
sibling section SHALL receive the proposal-sourcing correction
shaped to that command's finalization architecture:

- `uf.review-council.md` Section 4 and `uf.address-feedback.md`
  Section 2.3.7 source proposals from the no-provenance
  `consolidate_dispatch` retrieval payload, and their finalization
  sections (review-council Section 5; address-feedback Section
  2.3.8) add the dual-mode note: the retrieval call is distinct from
  the provenance-bearing exactly-once finalization call and does not
  count against it.
- `uf.review-pr.md` Section 5a adds the same no-provenance
  retrieval step after its `dispatch_review_runs` call and sources
  proposals from that payload; its Section 5b direct
  `finalize_review_dispatch` finalization is unchanged.

In every corrected section, each proposal's `category` field (not a
nonexistent `reference`) is the learning category.

#### Scenario: Sibling sections corrected

- **GIVEN** the shared Store Lesson Proposals section in a sibling
  command
- **WHEN** the section is edited
- **THEN** it sources proposals from the no-provenance
  `consolidate_dispatch` retrieval payload rather than claiming
  `dispatch_review_runs` returns them
- **AND** review-council Section 5 and address-feedback Section
  2.3.8 carry the dual-mode note scoping exactly-once to the
  provenance-bearing finalization call
- **AND** review-pr Section 5a gains the retrieval step without
  changing its Section 5b direct-finalization path
- **AND** it stores each via one `dewey_store_learning` call using
  the proposal's `information`, `tag`, and `category`

#### Scenario: Sibling mirrors stay byte-identical

- **GIVEN** a sibling command file is edited
- **WHEN** the edit completes
- **THEN** its scaffold mirror under
  `internal/scaffold/assets/opencode/commands/` is updated to match
  byte-for-byte

### Requirement: Scaffold Asset Parity

**ID: FR-006** This requirement MUST be satisfied.

The canonical command files and their scaffold mirrors MUST remain
byte-identical after the edits.

#### Scenario: Mirror parity verified

- **GIVEN** `.opencode/commands/uf.triage-issue.md` or an edited
  sibling command file
- **WHEN** the edit phase completes
- **THEN** the corresponding file under
  `internal/scaffold/assets/opencode/commands/` is updated to match
  byte-for-byte
- **AND** `make check` scaffold drift tests pass

### Requirement: Dependency Pin Lockstep and Engine Floor

**ID: FR-007** This requirement MUST be satisfied.

The change SHALL adopt the maintainer-approved `@opencode-ai/plugin`
1.4.10 → 1.19.0 pin with its full lockstep set: the manifests
(`.opencode/package.json`) and their scaffold asset mirrors, the
`TestPluginPackageInputs_ExactVersions`
expectedRuntime map, the pin-governance testdata
(`plugin-package-provenance.json`),
the AGENTS.md Technology Stack, and the README engine floor.
Upstream verification SHALL record each package's actual upstream
license — MIT for `@opencode-ai/plugin` and `@opencode-ai/sdk`;
Apache-2.0 for `@ai-sdk/provider`; AFL-2.1 OR BSD-3-Clause for
`json-schema` — rather than assuming MIT, and SHALL verify the
msgpackr 1 → 2 major transitive bump before adoption. The Node
engine range (ini 7: ^22.22.2 || ^24.15.0 || >=26.0.0) SHALL be
verified locally, declared in the `engines` field of
`.opencode/package.json` and its scaffold asset mirror, and a
single-commit rollback path SHALL remain executable.
The pin lockstep file group SHALL be committed as a dedicated
commit, staged by explicit path, separate from the command-doc and
plugin-tool edits.

#### Scenario: Lockstep parity

- **GIVEN** the pin update is applied
- **WHEN** the lockstep set is updated
- **THEN** manifests, scaffold mirrors, test expectations,
  and testdata all reference 1.19.0 consistently
- **AND** `make plugin-test` and the pin-enforcement test pass

#### Scenario: Engine floor declared and verified

- **GIVEN** ini 7 raises the Node engine floor to a disjunctive
  range (^22.22.2 || ^24.15.0 || >=26.0.0)
- **WHEN** the pin ships
- **THEN** `.opencode/package.json` and its scaffold asset mirror
  declare the range in their `engines` fields
- **AND** the resolved local Node version is recorded as
  reproducible evidence, and AGENTS.md and README document the
  range

#### Scenario: Dedicated pin commit

- **GIVEN** the change's edits span command docs, plugin tool code,
  and the pin lockstep set
- **WHEN** the work is committed
- **THEN** the pin lockstep files are staged by explicit path and
  committed as their own commit, separate from the command-doc and
  plugin-tool commits
