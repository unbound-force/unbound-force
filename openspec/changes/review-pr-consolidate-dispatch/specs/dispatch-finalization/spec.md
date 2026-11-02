<!--
  Adds FR-013..FR-016 to the dispatch-finalization
  capability, continuing its established bare-FR id space.
  FR-001..FR-008 come from the dispatch-finalize-determinism
  change; FR-009..FR-012 come from the
  fix-dispatch-verdict-derivation change; this change MUST
  archive after both. FR-017 onward stays reserved for future
  changes of this capability; sibling dispatch requirements
  live in the review-dispatch (RD-FR-NNN) and
  dispatch-recovery (DR-FR-NNN) capabilities. Scaffold asset
  parity is not restated here -- the capability's existing
  FR-005 governs the mirror edits and is extended by FR-016
  below.
-->

## ADDED Requirements

### Requirement: One-Call Finalization for review-pr

FR-013: The `/uf.review-pr` command MUST finalize every
dispatch with exactly one `consolidate_dispatch` call per
finalization attempt, passing the captured `correlation_id` and
`provenance` with `workflow_id` `uf.review-pr`. The provenance
`branch` and `commit` MUST resolve only from the frozen PR
input context (`head_ref` and `head_sha`) captured before child
dispatch -- never from the local checkout or a mutable ref --
because the finalizer validates the provenance commit against
`input_context.head_sha`. A rejected attempt follows the
FR-004 bounded mechanical-retry policy unchanged: at most two
retries of the identical call, only for mechanical rejections,
never altering assessment data. The command MUST NOT assemble,
restate, or recompute the closed finalization payload, and MUST
NOT call `finalize_review_dispatch` directly. If the frozen PR
context lacks `head_ref` or `head_sha`, the command MUST fail
closed with an explicit error naming the missing field.
`head_sha` MUST pass the command's SHA-format validation
(40-character lowercase hex). `head_ref` MUST be present and
SHOULD be pre-validated against the provenance branch schema
(printable ASCII, at most 255 characters -- matching
dispatch-time validation, so a frozen pair accepted at dispatch
time can never fail at finalization). On any missing or
malformed frozen field, the command MUST fail closed and MUST
NOT finalize with a fabricated, defaulted, or omitted value.

#### Scenario: Command finalizes after the council completes

- **GIVEN** a completed `/uf.review-pr` dispatch with at least
  one successful run
- **WHEN** the command reaches its finalization step
- **THEN** it issues a single `consolidate_dispatch` call with
  the captured `correlation_id` and provenance `{branch:
  <head_ref>, commit: <head_sha>, workflow_id: "uf.review-pr"}`
- **AND** no reconstructed runs, findings, run counts, or
  verdicts are passed

#### Scenario: Provenance tracks the frozen PR head

- **GIVEN** a local checkout whose HEAD differs from the PR's
  head SHA
- **WHEN** the command finalizes the dispatch
- **THEN** the provenance commit equals the frozen
  `input_context.head_sha` and internal finalization succeeds
  on the first attempt

#### Scenario: Missing frozen field fails closed

- **GIVEN** a PR input context missing `head_ref` or `head_sha`
- **WHEN** the command resolves finalization provenance
- **THEN** it stops with an explicit error naming the missing
  field and never calls the finalizer with a fabricated or
  omitted value

#### Scenario: Malformed frozen field fails closed

- **GIVEN** a PR input context whose `head_sha` is not a
  40-character lowercase hex string, or whose `head_ref` fails
  the provenance branch schema
- **WHEN** the command resolves finalization provenance
- **THEN** it stops with an explicit error naming the malformed
  field and never calls the finalizer with a fabricated,
  defaulted, or truncated value

#### Scenario: Mechanical rejection retries identically

- **GIVEN** a finalizer rejection classified as mechanical from
  the structured response alone
- **WHEN** the command retries
- **THEN** it re-issues the identical `consolidate_dispatch`
  call at most twice, changes no assessment data, and fails
  closed to `INCONCLUSIVE` when the second retry is rejected

### Requirement: Single-Source Proposals and Warnings

FR-014: The `/uf.review-pr` command MUST read the dispatch's
lesson proposals and consolidation warnings from the same
`consolidate_dispatch` finalization result that persists the
review-dispatch artifact -- the top-level `proposals` and
`warnings` arrays of that response. The command MUST NOT issue
a separate no-provenance retrieval call for them, MUST NOT
source them from `dispatch_review_runs` output, and MUST NOT
store raw child `> learn:` text or child-supplied tags,
categories, or hashes. On the fail-closed path -- exhausted
retries or a non-mechanical failure -- lesson storage is
skipped: the stranded proposals remain in the persisted
dispatch store and the failure is reported, never fabricated.

#### Scenario: Proposals stored from the finalization result

- **GIVEN** a successful finalization whose result carries two
  validated lesson proposals
- **WHEN** the command stores lessons
- **THEN** each proposal is passed to `dewey_store_learning`
  exactly once, using the proposal's `information`, `tag`, and
  `category` from the finalization result

#### Scenario: Fail-closed path does not fabricate lessons

- **GIVEN** a dispatch whose finalization exhausted its bounded
  retries
- **WHEN** the command handles the failure
- **THEN** no lesson is stored from improvised data, the
  stranded proposals are reported as retained in the dispatch
  store, and the operation result is failing `INCONCLUSIVE`

### Requirement: Presentation-Only Cross-Persona Grouping

FR-015: Cross-persona root-cause grouping and compound severity
in `/uf.review-pr`, and the deduplication and grouping prose in
`/uf.review-council` Section 3, MUST apply to presenting the
review comment and terminal report only. They MUST NOT merge,
filter, re-order, or otherwise alter the findings, contributing
run ids, verdicts, run counts, or advisories submitted to or
returned by finalization: the finalized payload's findings MUST
remain exactly the consolidated findings `consolidate_dispatch`
produced, because the finalizer enforces exact correspondence
between run findings and consolidated findings. Command prose
reworded by this change that describes deduplication, grouping,
or compound severity MUST state this presentation-only scope
explicitly.

#### Scenario: Grouped comment leaves finalization data intact

- **GIVEN** successful runs whose findings include an architect
  finding and a guard finding on adjacent lines of the same
  file
- **WHEN** the command composes the GitHub review comment
- **THEN** the comment MAY group the two findings for
  presentation with attribution to both contributing personas
- **AND** the finalized payload's findings keep each finding
  and its contributing run ids exactly as the tool returned
  them

#### Scenario: Council prose cannot reintroduce the rejection

- **GIVEN** `uf.review-council.md` Section 3
- **WHEN** a reader applies its deduplication and grouping
  guidance
- **THEN** the prose explicitly scopes that guidance to
  presentation and names finalization data as excluded

### Requirement: review-pr Mirror Parity and Content Gates

FR-016: Edits to `uf.review-pr.md` and `uf.review-council.md`
MUST be mirrored byte-identically into
`internal/scaffold/assets/opencode/commands/`, per the
capability's existing FR-005 parity rule. Content-gate tests in
`internal/scaffold/scaffold_test.go` MUST assert, for
`uf.review-pr.md`: the one-call `consolidate_dispatch`
finalization contract with `workflow_id` `uf.review-pr`; the
frozen-PR-context provenance source; the never-assemble and
never-call-finalizer-directly prohibitions; the bounded retry
policy; and the presentation-only scope of grouping. The same
tests MUST assert the absence of the manual-finalize prose: no
instruction to call `finalize_review_dispatch`, and no
finalization-path deduplication mandate conflicting with the
finalizer's run-finding correspondence check. For
`uf.review-council.md`, the tests MUST assert the
presentation-only scope of the Section 3 deduplication prose.
All parity and content assertions run under `make check`.

#### Scenario: Content gates pin the new contract

- **GIVEN** the rewritten `uf.review-pr.md` and its mirror
- **WHEN** `make check` runs the scaffold content-gate tests
- **THEN** the one-call contract, frozen-context provenance,
  and bounded-retry assertions pass
- **AND** the absence assertions confirm no manual
  `finalize_review_dispatch` instruction remains

#### Scenario: Mirror drift fails the build

- **GIVEN** a `.opencode/commands/uf.review-pr.md` edit not yet
  mirrored
- **WHEN** `make check` runs the drift-detection tests
- **THEN** the parity failure names the drifted asset and the
  build fails

## MODIFIED Requirements

None. FR-001..FR-012 are unchanged; this change only adds
requirements.

## REMOVED Requirements

None.
