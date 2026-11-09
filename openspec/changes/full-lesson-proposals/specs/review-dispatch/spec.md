<!-- capability: review-dispatch
     change: full-lesson-proposals
     originating_issue: 969
     Complete lesson proposals in every consolidate_dispatch
     tool result. Design rationale: design.md (D1-D3). -->

Requirement numbering: RD-FR-058 is allocated for this change, the
id `review-dispatch-consumer-bounds` left reserved in the
RD-FR-054..058 range. Requirement ids MUST stay inside the
allocated range.

Amendment and archive order: this requirement amends the proposal
clauses of RD-FR-049 and RD-FR-051 (unarchived
`bound-consolidate-dispatch-output`) and RD-FR-054 (unarchived
`review-dispatch-consumer-bounds`). `openspec/specs/review-dispatch/`
does not exist until those changes archive, so the amended clauses
are restated in full below. Where a clause here conflicts with
them, this change governs. This change MUST archive in the same
batch as, or after, both.

## ADDED Requirements

### Requirement: Complete Lesson Proposals in Tool Results

**ID: RD-FR-058** This requirement MUST be satisfied.

Every `proposals` array in a `consolidate_dispatch` tool result —
the `detail: "summary"` retrieval, the `detail: "full"` retrieval,
and the compact finalization result — MUST carry each entry's
`agent`, `information`, and `tag` exactly as harvested, and its
`category` when valid. No result MAY shorten a proposal field, and
no proposal entry MAY carry a `truncated` key. This replaces the
RD-FR-049 clauses that cut proposal `information` to 512
characters and `tag` and `agent` to 128, and the RD-FR-054 scenario
expecting a 512-character preview entry with `truncated: true`.

A proposal is either present whole or absent. The summary
retrieval and the compact finalization result MUST each stay
strictly under 40,960 bytes of compact UTF-8 JSON. When the entries
do not fit, the builder MUST drop trailing proposals whole and set
`proposals_truncated: true`. The summary retrieval keeps the
RD-FR-049 drop order (findings, advisories, proposals, warnings).
The 10-entry proposal cap and `proposals_truncated` flag of
RD-FR-049 and RD-FR-051 are unchanged, as are the caps and
per-entry `truncated` markers of findings, advisories, warnings,
and `verdict_reason`.

The persisted artifact's `payload.proposals` (RD-FR-054) remains
the lesson source of RD-FR-056; commands MUST still store lessons
from it, never from a tool-result preview.

#### Scenario: A long proposal survives summary retrieval

- **GIVEN** a child run that called `submit_lesson_proposal` with
  an `information` longer than 512 characters
- **WHEN** the parent calls `consolidate_dispatch` with
  `detail: "summary"`
- **THEN** the summary's proposal entry carries the complete
  submitted `information`, `tag`, and `category`
- **AND** the entry carries no `truncated` key
- **AND** the summary carries no `proposals_truncated` flag

#### Scenario: Every retrieval mode agrees on the harvested text

- **GIVEN** the same dispatch
- **WHEN** the parent retrieves the full payload, finalizes with
  provenance, and reads the artifact's `proposals` section through
  `read_dispatch_artifact`
- **THEN** the full retrieval, the compact result, and the artifact
  page each carry a proposal equal to the summary's entry

#### Scenario: Over-budget proposals are dropped whole

- **GIVEN** ten harvested proposals of 16,384 characters each
- **WHEN** the summary retrieval or the compact finalization result
  is built
- **THEN** its compact JSON is under 40,960 bytes
- **AND** every kept proposal carries its complete `information`
- **AND** `proposals_truncated` is `true`
