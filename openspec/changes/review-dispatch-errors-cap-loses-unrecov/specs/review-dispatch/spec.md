<!-- capability: review-dispatch
     change: review-dispatch-errors-cap-loses-unrecov
     originating_issue: 1039
     Recoverable finalization errors: whole-or-absent error
     previews plus a persisted failure record readable through
     read_dispatch_artifact. Design rationale: design.md
     (D1-D6). -->

Requirement numbering: RD-FR-066 is allocated for this change.
Requirement ids MUST stay inside the allocated range.

Amendment and archive order: this requirement amends the errors
clauses of RD-FR-051 (unarchived `bound-consolidate-dispatch-
output`) and the section list and path derivation of RD-FR-057
(unarchived `review-dispatch-consumer-bounds`).
`openspec/specs/review-dispatch/` does not exist until those
changes archive, so the amended clauses are restated in full
below. Where a clause here conflicts with them, this change
governs. This change MUST archive in the same batch as, or
after, both.

## ADDED Requirements

### Requirement: Recoverable Finalization Errors

**ID: RD-FR-066** This requirement MUST be satisfied.

Every `errors` array in a compact finalization result — the
`finalize_review_dispatch` tool result and the
`consolidate_dispatch` provenance-bearing result — MUST carry
each error string exactly as the finalizer produced it. No
result MAY shorten an error entry, and no error entry MAY carry
a `truncated` key. This replaces the RD-FR-051 clause that cuts
each error to 256 characters. The 5-entry `errors` cap, the
`errors_truncated` flag, and the `errors` position as the final
key of RD-FR-051 are unchanged; `errors_truncated` MUST be true
only when whole entries were dropped, never for shortening.

The compact finalization result MUST stay strictly under
40,960 bytes of compact UTF-8 JSON (RD-FR-049). When it does
not, the builder MUST drop trailing entries whole — proposals
first, then errors — flagging every drop. Errors drop after
proposals because they are the diagnostic payload of exactly
the paths where nothing else is persisted.

When `finalize_review_dispatch` fails — the input rejected at
schema validation, the payload rejected at semantic validation,
or the artifact persistence failed — the plugin MUST attempt to
write a failure record to
`.uf/artifacts/dispatch/<correlation-id>.failure.json` under
the project root, carrying the complete `errors` array
verbatim. The record MUST use the Hero Interface Contract
envelope shape with `artifact_type` `review-dispatch-failure`
and MUST NOT use `review-dispatch`, so consumers that select
success artifacts (resume detection) keep ignoring it. The
record MUST be published with the same atomic discipline as a
success artifact: exclusive-create temp file, rename-by-link,
0o750 directory, no overwrite — an `EEXIST` collision MUST mint
a fresh record correlation id within the same bounded retry
count. The record's correlation id MUST be the dispatch's when
a usable one exists, and a freshly minted id otherwise; the
record payload MUST carry `dispatch_correlation_id` (the
dispatch correlation when usable, null otherwise) plus the
command and mode when known, `status: "failed"`,
`operation_verdict`, and the errors. Every payload string MUST
be bounded by the record's strict schema with sanity bounds
sized above the natural error-composition ceilings; the bounds
fail closed and MUST never truncate.

A failure-record write failure MUST NOT mask the finalization
failure: the plugin MUST return the finalization result with
`failure_record` null and the record-write error appended to
the result's `errors` (subject to the same preview budget).

The compact finalization result MUST carry a `failure_record`
key — `null` or an object of `{ correlation_id, path }` —
inserted immediately after `artifact_path`, and omitted when
null. This amends the RD-FR-051 insertion order to: `status`,
`operation_verdict`, `correlation_id`, `artifact_path`,
`failure_record` (when present), `review_verdict`, `run_counts`,
the optional observability keys, `calculated_assessment`,
`errors`.

`read_dispatch_artifact` MUST accept section `errors`. For a
correlation id it MUST derive the failure-record path as
`.uf/artifacts/dispatch/<correlation_id>.failure.json` and MUST
NOT read any caller-supplied path. The section MUST return
`status: "ok"` with the correlation id, section, `total`, and
the complete, untruncated errors array when the record exists
and passes the strict failure-record schema, and MUST fail
closed with the RD-FR-057 error codes otherwise: `not_found`
when no record exists (including a dispatch that finalized
successfully), `invalid_artifact` when the file is absent as a
regular non-symlink file of at most 64 MiB, does not parse, has
the wrong `artifact_type`, or fails the payload schema. This
amends the RD-FR-057 section enum to `summary`, `findings`,
`advisories`, `proposals`, `runs`, and `errors`.

Every edited `.opencode/` plugin or lib file and the failure
record's JSON Schema MUST be mirrored byte-identical under
`internal/scaffold/assets/`, and the schema MUST be registered
in `schemas/review-dispatch-failure/v1.0.0.schema.json` with
mirrors under `.uf/schemas/review-dispatch-failure/` and
`internal/scaffold/assets/uf/schemas/review-dispatch-failure/`.

#### Scenario: A composed validation error survives whole

- **GIVEN** a finalization input rejected at schema validation
  whose composed `formatValidationError` string is longer than
  256 characters (three or more Zod issues)
- **WHEN** `finalize_review_dispatch` returns the compact
  result
- **THEN** the `errors` entry carries the complete composed
  string
- **AND** the entry carries no `truncated` key
- **AND** the result carries no `errors_truncated` flag

#### Scenario: Semantic overflow is flagged and recoverable

- **GIVEN** a payload rejected at semantic validation with more
  than five errors
- **WHEN** `consolidate_dispatch` returns the compact result
- **THEN** `errors` carries five complete entries
- **AND** `errors_truncated` is true
- **AND** a `review-dispatch-failure` record exists carrying
  every rejected error verbatim

#### Scenario: The byte budget drops whole errors

- **GIVEN** failure errors whose complete entries push the
  compact JSON to 40,960 bytes or more once proposals are
  exhausted
- **WHEN** the compact finalization result is built
- **THEN** its compact JSON is under 40,960 bytes
- **AND** every kept error entry is complete
- **AND** `errors_truncated` is true
- **AND** the dropped errors remain in the failure record

#### Scenario: A garbage input mints a record correlation

- **GIVEN** a finalization input rejected at schema validation
  that carries no usable correlation id
- **WHEN** the failure record is written
- **THEN** the record is keyed by a freshly minted correlation
  id
- **AND** the compact result's `failure_record` carries that
  id and the record path
- **AND** the record's payload carries `dispatch_correlation_id`
  null

#### Scenario: Record write failure degrades without masking

- **GIVEN** a finalization failure whose failure-record write
  also fails
- **WHEN** the compact result is returned
- **THEN** `failure_record` is null
- **AND** `errors` still carries complete preview entries
- **AND** the record-write error is among them

#### Scenario: The errors section reads the record

- **GIVEN** a failed finalization with a persisted failure
  record
- **WHEN** `read_dispatch_artifact` is called with the record's
  correlation id and `section: "errors"`
- **THEN** it returns `status: "ok"` with `total` and the
  complete errors array, untruncated

#### Scenario: A successful dispatch has no errors section

- **GIVEN** a correlation id with only a success artifact
- **WHEN** `read_dispatch_artifact` is called with
  `section: "errors"`
- **THEN** it returns `status: "error"` with `not_found`

#### Scenario: The failure record never impersonates a success
artifact

- **GIVEN** a failure record written by any failure path
- **WHEN** its envelope is inspected
- **THEN** `artifact_type` is `review-dispatch-failure`
- **AND** a consumer selecting `review-dispatch` envelopes
  skips it

#### Scenario: The result keys keep their contract order

- **GIVEN** a failed finalization with a persisted failure
  record
- **WHEN** the compact result's key order is parsed
- **THEN** `failure_record` appears immediately after
  `artifact_path`
- **AND** `errors` remains the final key

## MODIFIED Requirements

None. The amended RD-FR-051 and RD-FR-057 clauses are restated
in full under RD-FR-066 above, following the
`full-lesson-proposals` amendment precedent.

## REMOVED Requirements

None.
