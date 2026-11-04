<!-- capability: review-dispatch
     change: review-dispatch-consumer-bounds
     originating_issue: 941
     Durable lesson proposals, cross-persona proposal
     deduplication, an artifact-sourced consumer contract, and a
     bounded artifact reader. Design rationale and rejected
     alternatives: design.md (D1-D7). -->

Requirement numbering: RD-FR-054 through RD-FR-057 are allocated
for this change, from the RD-FR-054..058 range that
`bound-consolidate-dispatch-output` reserved; RD-FR-058 stays
reserved for future review-dispatch changes. Elsewhere in the
review-dispatch capability, RD-FR-001..008 is
review-council-multi-model-fanout, RD-FR-025..027
fix-review-dispatch-recovery, RD-FR-028..031
derive-per-run-timeout, RD-FR-032..036
review-dispatch-diff-manifest, RD-FR-037
host-fallback-provenance, RD-FR-038
split-host-resolution-failure-classes, RD-FR-039..048 reserved
there, and RD-FR-049..053 bound-consolidate-dispatch-output; the
review-matrix capability carries RD-FR-009..010 and RD-FR-019..024,
with RD-FR-011..018 reserved there. Requirement ids MUST stay
inside the allocated range.

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
text, and MUST omit `proposals` when there are none. When more than
512 deduplicated proposals remain, it MUST keep the first 512 and
append one `{type: "proposals_overflow", kept_proposals,
dropped_proposals}` warning (summary `detail`
`kept=<n> dropped=<n>`), so the array bound never fails
finalization. The persisted
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

#### Scenario: Proposal overflow is capped, not fatal

- **GIVEN** a dispatch whose successful runs submitted 515 distinct
  proposals
- **WHEN** `consolidate_dispatch` finalizes
- **THEN** finalization succeeds and the artifact carries the first
  512
- **AND** one `proposals_overflow` warning reports 512 kept and 3
  dropped

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
2. Two proposals are never near-identical when exactly one of them
   carries a negation token (`not`, `no`, `never`, `cannot`, `nor`,
   `without`, or the `t` left by an "n't" contraction). Otherwise
   they are near-identical when both token sets carry at least 6
   tokens and the Jaccard similarity `|A ∩ B| / |A ∪ B|` is at least
   0.3, or when either set carries fewer than 6 tokens and the two
   proposals' token sequences are equal. The symmetric measure keeps
   a short lesson whose words mostly recur in a longer one.
3. Proposals are visited in collection order and compared with
   the kept entries in order. On the first near-identical match,
   the entry with the longer `information` (the kept one on a tie)
   occupies the kept slot and the other is dropped.
4. `agent`, `tag`, and `category` MUST NOT affect matching.

Each dropped proposal MUST append one warning
`{type: "duplicate_proposal_merged", agent, kept_agent}`, where
`agent` names the dropped entry's agent and `kept_agent` the kept
entry's. This amends the warning kinds of RD-FR-049 and RD-FR-052
from five to seven: the summary and compact projections compose its
`detail` as `agent=<agent> kept=<kept_agent>`, and RD-FR-052
recovery accepts `duplicate_proposal_merged` and the RD-FR-054
`proposals_overflow` as documented kinds.

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

- **GIVEN** a dispatch whose proposal pairs each have a Jaccard
  similarity below 0.3
- **WHEN** `consolidate_dispatch` runs
- **THEN** every proposal is kept in collection order
- **AND** no `duplicate_proposal_merged` warning is emitted

#### Scenario: Short proposals match only exactly

- **GIVEN** two proposals whose token sets carry fewer than 6
  tokens and differ in one token
- **WHEN** `consolidate_dispatch` runs
- **THEN** both are kept

#### Scenario: A lesson and its negation are both kept

- **GIVEN** "Retry is not safe: finalization is not idempotent" and
  "Retry is safe: finalization is idempotent", worded otherwise
  alike
- **WHEN** `consolidate_dispatch` runs
- **THEN** both are kept and no merge warning is emitted

#### Scenario: A short lesson inside a long one is kept

- **GIVEN** a one-sentence proposal whose words all recur inside a
  much longer proposal
- **WHEN** `consolidate_dispatch` runs
- **THEN** both are kept, because their Jaccard similarity is below
  0.3

### Requirement: Artifact-Sourced Consumer Contract

**ID: RD-FR-056** This requirement MUST be satisfied.

This requirement amends RD-FR-053 and the full-mode consumer
paragraph of RD-FR-052, and revisits design D5 of
`bound-consolidate-dispatch-output`.

Every in-repo command that stores dispatch lessons MUST store them
from the persisted artifact, reading `payload.proposals` through the
RD-FR-057 `read_dispatch_artifact` tool, never with the file-read
tool. Every artifact read in these commands (findings, advisories,
proposals, runs) MUST use that tool. It MUST call
`dewey_store_learning` exactly once per entry, using the entry's
`information`, `tag`, and `category`. An empty `proposals` section
means no lesson was submitted, unless the compact result's
`proposals` preview is non-empty: that means the loaded plugin
predates durable proposals, and the command MUST store nothing and
report the previewed lessons as stranded. It MUST NOT store
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
  `read_assessment_fields` run ids read the `runs`, `proposals`,
  `findings`, and `advisories` sections of the artifact.
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

#### Scenario: An older loaded plugin strands previewed lessons

- **GIVEN** a compact result whose `proposals` preview is non-empty
  and an artifact with no `payload.proposals`
- **WHEN** a consumer reaches lesson storage
- **THEN** it stores nothing and reports the previewed lessons as
  stranded rather than "none submitted"

### Requirement: Bounded Artifact Reader

**ID: RD-FR-057** This requirement MUST be satisfied.

The review-dispatch plugin MUST register a `read_dispatch_artifact`
tool. Its strict input carries a `correlation_id` (UUID), a
`section` (`summary`, `findings`, `advisories`, `proposals`, or
`runs`), and an optional non-negative integer `offset`. The tool
MUST derive the artifact path as
`.uf/artifacts/dispatch/<correlation_id>.json` under the project
root and MUST NOT read any caller-supplied path. It MUST fail
closed with `{status: "error", code, message}` — code
`invalid_input`, `not_found`, or `invalid_artifact` — when the
input fails validation, the file is absent, the file is not a
regular non-symlink file of at most 64 MiB, the content does not
parse, `artifact_type` is not `review-dispatch`, or the payload
fails the strict payload schema.

The `summary` section MUST return the command, mode, verdict,
`verdict_reason`, run counts, coverage, workflow result, and the
entry counts of findings, advisories, proposals, and runs. Every
other section MUST return `{total, offset, entries, next_offset}`
with complete, untruncated entries starting at `offset`, adding
entries while the compact serialization stays under 40,960 bytes
and always returning at least one entry when any remain;
`next_offset` is null on the last page. `runs` entries MUST omit
their `findings` and carry `finding_count` instead, because each
finding appears once in the `findings` section with its `run_ids`.

This replaces reading the artifact with the file-read tool:
`JSON.stringify(envelope, null, 2)` writes each string on one line,
and the host's file-read tool cuts lines at 2,000 characters, so a
long proposal or finding would be stored shortened or fail the
parse.

#### Scenario: Long proposals page intact

- **GIVEN** an artifact carrying seven proposals of 9,000
  characters each
- **WHEN** a consumer calls `read_dispatch_artifact` with
  `section: "proposals"` and follows `next_offset` until null
- **THEN** every page is under 40,960 bytes
- **AND** the seven entries arrive complete and in order

#### Scenario: Runs omit their findings

- **GIVEN** a finalized artifact
- **WHEN** `section: "runs"` is read
- **THEN** each entry carries the run's provenance, terminal state,
  `workflow_verdict`, and `finding_count`, and no `findings`

#### Scenario: Reader fails closed

- **GIVEN** a correlation ID with no artifact, an artifact that does
  not parse, a symlinked artifact, or a payload that fails the
  strict schema
- **WHEN** `read_dispatch_artifact` is called
- **THEN** it returns `status: "error"` with `not_found` or
  `invalid_artifact` and no entries

## REMOVED Requirements

None.
