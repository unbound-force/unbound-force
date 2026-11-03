<!-- capability: review-dispatch
     change: review-dispatch-consumer-bounds
     originating_issue: 941
     Durable lesson proposals, cross-persona proposal
     deduplication, and an artifact-sourced consumer contract.
     Design rationale and rejected alternatives: design.md
     (D1-D6). -->

Requirement numbering: RD-FR-054 through RD-FR-056 are allocated
for this change, from the RD-FR-054..058 range that
`bound-consolidate-dispatch-output` reserved; RD-FR-057 and
RD-FR-058 stay reserved for future review-dispatch changes.

Amendment and archive order: each requirement below amends one or
more requirements of the unarchived
`bound-consolidate-dispatch-output` change (RD-FR-049 through
RD-FR-053). `openspec/specs/review-dispatch/` does not exist until
that change archives, so OpenSpec cannot carry MODIFIED deltas for
them here; the amended clauses are restated in full below. Where a
clause here conflicts with RD-FR-049 through RD-FR-053, this change
governs. This change MUST archive in the same batch as, or after,
`bound-consolidate-dispatch-output`.

## ADDED Requirements

### Requirement: Durable Lesson Proposals

**ID: RD-FR-054** This requirement MUST be satisfied.

The review-dispatch payload MUST accept an optional top-level
`proposals` array. Each entry MUST be a strict object carrying
`agent` (a Divisor agent name), `information` (1-16,384
characters), `tag` (1-128 characters), and an optional `category`
(one of `decision`, `pattern`, `gotcha`, `context`, or
`reference`). The array MUST hold at most 512 entries. The strict
Zod payload schema and `schemas/review-dispatch/v1.0.0.schema.json`
MUST both declare the property; a payload without it MUST remain
valid, so artifacts written before this change keep validating.

When `consolidate_dispatch` finalizes (provenance supplied), the
payload it hands to `finalizeReviewDispatch` MUST carry the
deduplicated proposals of RD-FR-055 with their full `information`
text, and MUST omit `proposals` when there are none. The persisted
artifact MUST serialize `proposals` after `advisories` and before
`plan` and `runs`. `warnings` MUST remain result-level observability
and MUST NOT enter the payload.

This amends RD-FR-050, whose rule that `proposals` never reach the
finalizer payload no longer applies; its rule for `warnings` stands.
The full retrieval result keeps its RD-FR-050 key order, and its
`proposals` value is the same deduplicated array.

The bounded `proposals` arrays of the summary retrieval (RD-FR-049)
and the compact finalization result (RD-FR-051) keep their caps and
`truncated` markers. They are a preview: a command MUST NOT store a
lesson from them. This replaces the RD-FR-049 and RD-FR-051
statements that a dropped or shortened proposal is unrecoverable
from the artifact.

#### Scenario: Full proposal text is persisted

- **GIVEN** a successful dispatch whose runs submitted a lesson
  proposal with 3,000 characters of `information`
- **WHEN** `consolidate_dispatch` finalizes with provenance
- **THEN** the artifact at `artifact_path` carries
  `payload.proposals` with that entry's complete 3,000-character
  `information`, its `agent`, `tag`, and `category`
- **AND** the compact result's preview entry is cut to 512
  characters and carries `truncated: true`

#### Scenario: Proposals precede the plan and runs in the artifact

- **GIVEN** a finalized dispatch with at least one proposal
- **WHEN** the persisted artifact's payload is parsed
- **THEN** `Object.keys(payload)` places `proposals` after
  `advisories` and before `plan` and `runs`

#### Scenario: No proposals leaves the payload unchanged

- **GIVEN** a dispatch whose successful runs submitted no lesson
  proposal
- **WHEN** `consolidate_dispatch` finalizes
- **THEN** the persisted payload carries no `proposals` key

#### Scenario: Legacy artifacts stay valid

- **GIVEN** a review-dispatch payload without `proposals`
- **WHEN** it is validated against the Zod payload schema and
  `schemas/review-dispatch/v1.0.0.schema.json`
- **THEN** both accept it

#### Scenario: Malformed persisted proposal is rejected

- **GIVEN** a finalization payload whose `proposals` entry has an
  unknown key, an empty `tag`, or a `category` outside the five
  values
- **WHEN** it is submitted to `finalize_review_dispatch`
- **THEN** structural validation rejects it and no artifact is
  persisted

### Requirement: Cross-Persona Proposal Deduplication

**ID: RD-FR-055** This requirement MUST be satisfied.

`consolidate_dispatch` MUST deduplicate the collected proposals
once, before it builds the finalizer payload, the full retrieval,
the summary retrieval, or the compact finalization result, so
every output carries the same set in the same order.

Deduplication MUST be deterministic:

1. The token set of a proposal is its `information` after Unicode
   NFKC normalization and lowercasing, split on every run of
   characters that are neither letters nor digits, with a fixed
   list of English stop words removed.
2. Two proposals are near-identical when both token sets carry at
   least 4 tokens and `|A ∩ B| / min(|A|, |B|)` is at least 0.5, or
   when either set carries fewer than 4 tokens and the two
   proposals' token sequences are equal.
3. Proposals are visited in collection order and compared with
   the kept entries in order. On the first near-identical match,
   the entry with the longer `information` (the kept one on a tie)
   occupies the kept slot and the other is dropped.
4. `agent`, `tag`, and `category` MUST NOT affect matching.

Each dropped proposal MUST append one warning
`{type: "duplicate_proposal_merged", agent, kept_agent}`, where
`agent` names the dropped entry's agent and `kept_agent` the kept
entry's. This amends the warning kinds of RD-FR-049 and RD-FR-052
from five to six: the summary and compact projections compose its
`detail` as `agent=<agent> kept=<kept_agent>`, and RD-FR-052
recovery accepts `duplicate_proposal_merged` as a documented kind.

#### Scenario: Two personas submit the same lesson

- **GIVEN** a dispatch in which `divisor-adversary` and
  `divisor-architect` each submitted a lesson saying the plugin
  host does not enforce a tool's declared args shape, worded
  differently and tagged differently
- **WHEN** `consolidate_dispatch` runs
- **THEN** exactly one of the two lessons is returned and
  persisted, the one with the longer `information`
- **AND** one `duplicate_proposal_merged` warning names both agents

#### Scenario: Distinct lessons are kept

- **GIVEN** a dispatch whose proposals share fewer than half of the
  smaller proposal's tokens
- **WHEN** `consolidate_dispatch` runs
- **THEN** every proposal is kept in collection order
- **AND** no `duplicate_proposal_merged` warning is emitted

#### Scenario: Short proposals match only exactly

- **GIVEN** two proposals whose token sets carry fewer than 4
  tokens and differ in one token
- **WHEN** `consolidate_dispatch` runs
- **THEN** both are kept

### Requirement: Artifact-Sourced Consumer Contract

**ID: RD-FR-056** This requirement MUST be satisfied.

This requirement amends RD-FR-053 and the full-mode consumer
paragraph of RD-FR-052, and revisits design D5 of
`bound-consolidate-dispatch-output`.

Every in-repo command that stores dispatch lessons MUST store them
from the persisted artifact at the compact finalization result's
`artifact_path`, reading `payload.proposals` in the same single read
that yields `payload.findings` and `payload.advisories`. It MUST
call `dewey_store_learning` exactly once per entry, using the
entry's `information`, `tag`, and `category`. An absent
`payload.proposals` means no lesson was submitted. It MUST NOT store
a lesson from the summary retrieval, the compact finalization
result, `dispatch_review_runs`, child transcripts, run files, or the
session database. When finalization fails, the command MUST store no
lesson and MUST report the proposals as stranded in the dispatch
store.

No in-repo command may call `consolidate_dispatch` with
`detail: "full"`. The consumer contract by command is:

- `uf.review-council.md`: Section 4 keeps the `detail: "summary"`
  retrieval and the RD-FR-052 recovery path for presentation and
  warnings; lessons are stored after the Section 5 finalization of
  the same iteration, from its artifact.
- `uf.review-pr.md`: Section 5 finalization; lessons from the
  artifact.
- `uf.unleash.md`: Steps 4 and 6 finalize and store lessons from the
  artifact they already read.
- `uf.triage-issue.md`: Section 2.3 performs the Section 4.6
  finalization immediately after `dispatch_review_runs`; lesson
  storage (Section 2.4), verdict resolution (Section 3.1), and the
  `read_assessment_fields` run ids read `payload.runs`,
  `payload.proposals`, findings, and advisories from the artifact.
  Section 4.6 MUST NOT call `consolidate_dispatch` again.
- `uf.address-feedback.md`: each Tier 2 item runs its Section 2.3.8
  finalization immediately after its Section 2.3.4 dispatch;
  consolidation (Section 2.3.5) and lesson storage (Section 2.3.7)
  read that item's artifact.

The `uf.triage-issue.md` and `uf.address-feedback.md` no-provenance
retrievals and their truncation checks are removed. When the
finalization fails closed, no artifact exists and the invocation
(triage) or item (feedback) MUST fail closed as `INCONCLUSIVE`
before any GitHub mutation, fabricating no run record, verdict,
finding, or proposal. The all-skip and forwarded-`correlation_id`
`not_found` semantics are unchanged.

#### Scenario: Review-pr stores full lessons

- **GIVEN** `uf.review-pr.md` finalizes a dispatch whose artifact
  carries seven proposals of 900 characters each
- **WHEN** it stores lessons
- **THEN** it calls `dewey_store_learning` seven times with the
  artifact's complete `information`
- **AND** it never stores the compact result's 512-character preview

#### Scenario: Triage reads runs from the artifact

- **GIVEN** `uf.triage-issue.md` has dispatched its runs
- **WHEN** it continues to verdict resolution
- **THEN** it has already finalized with provenance and reads the
  run-recorded `workflow_verdict` values from the artifact's
  `payload.runs`
- **AND** it makes no no-provenance `consolidate_dispatch` call

#### Scenario: Feedback item reads its own artifact

- **GIVEN** `uf.address-feedback.md` has dispatched a Tier 2 item
- **WHEN** it consolidates the item's recommendation
- **THEN** it reads the runs of the artifact returned by that
  item's own finalization
- **AND** no host-truncated retrieval can fail the item closed

#### Scenario: Failed finalization stores nothing

- **GIVEN** a consumer whose finalization fails closed
- **WHEN** it reaches lesson storage
- **THEN** it stores no lesson and reports the proposals as
  stranded in the dispatch store
- **AND** it fabricates no proposal from a preview array
