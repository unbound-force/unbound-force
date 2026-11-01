<!--
  capability: review-dispatch
  change: accept-variantless-parent-model
  Archive ordering: this change MUST archive after
  review-council-multi-model-fanout (spec creation plus the
  'Review Dispatch Artifact' base). The ADDED 'Finalization
  Model Validation' requirement establishes the corrected
  contract outright — repo-wide no delta ever ADDEDs that
  requirement, so no predecessor text can materialize for a
  MODIFY to apply against; the same-named MODIFIED entry in
  advisor-host-model-fallback is itself an orphan, a
  pre-existing defect in that change, out of scope here. The
  MODIFIED 'Review Dispatch Artifact' entry supersedes the
  over-strict host-lookup text the fanout delta carries under
  its ADDED requirement of that name (RD-FR-008); that
  supersession is per requirement name and archive-order
  dependent. Strategy and rationale: design.md (D2).
-->

## ADDED Requirements

### Requirement: Finalization Model Validation

`validateRunModel` MUST validate host source (null requested
provenance), advisor source with `requested_model: null`
(validated as host runs), and other sources (non-null
requested, null resolved provenance). A successful host or
host-resolved advisor run MUST carry non-null
`resolved_parent_model` and MAY carry a null
`resolved_parent_variant` when the provider reports none; a
null `resolved_parent_model` MUST be rejected with a message
containing "requires resolved parent model".

#### Scenario: Host-resolved advisor passes validation

- **GIVEN** a finalization payload with a successful run where
  `source: "advisor"` and `requested_model: null`
- **AND** `resolved_parent_model` is non-null
- **AND** `resolved_parent_variant` is non-null because the
  provider reports an active variant
- **AND** the plan entry has `model: null` and `variant: null`
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation succeeds

#### Scenario: Host run passes validation without a variant

- **GIVEN** a finalization payload with a successful run where
  `source: "host"`
- **AND** `resolved_parent_model` is non-null for a variantless
  parent model (`enmaas/rits/zai-org/glm-5-3`) and
  `resolved_parent_variant` is null
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation succeeds

#### Scenario: Host-resolved advisor passes validation without a variant

- **GIVEN** a finalization payload with a successful run where
  `source: "advisor"` and `requested_model: null`
- **AND** `resolved_parent_model` is non-null
  (`enmaas/rits/zai-org/glm-5-3`, observed in the wild)
- **AND** `resolved_parent_variant` is null because the
  provider reports no variant
- **AND** the plan entry has `model: null` and `variant: null`
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation succeeds

#### Scenario: Host-resolved advisor without resolved parent fails

- **GIVEN** a finalization payload with a successful run where
  `source: "advisor"` and `requested_model: null`
- **AND** `resolved_parent_model` is null
- **AND** the plan entry has `model: null` and `variant: null`
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation fails with a message containing
  "requires resolved parent model"

#### Scenario: Host-resolved advisor without resolved parent model fails

- **GIVEN** a successful run with `source: "advisor"` and
  `requested_model: null`
- **AND** `resolved_parent_model` is null
- **WHEN** `finalize_review_dispatch` validates the run
- **THEN** validation fails with a message containing
  "requires resolved parent model"

The variant-or-null sanction matches the writer contract of
`resolveHostModel()` in `.opencode/lib/agent-executor.ts`
(commit 716f2036, absent-variant fallback): resolution
succeeds with no variant when the provider does not populate
one, and the persistence layer writes `host.variant ?? null`
under `NullableVariantSchema`.

This requirement is ADDED, not MODIFIED, because no delta
repo-wide ever ADDEDs 'Finalization Model Validation':
`review-council-multi-model-fanout` ADDEDs RD-FR-001 through
RD-FR-008 only, and the same-named MODIFIED entry in the
unarchived `advisor-host-model-fallback` delta is itself an
orphan — a pre-existing stack-wide defect in that change, out
of scope here, which this entry neither supersedes nor
repairs. No archive ordering can materialize a predecessor
FMV text for a MODIFY to apply against (archive hard-aborts
on an orphan MODIFIED), so this ADDED entry establishes the
corrected requirement outright. The MODIFIED 'Review Dispatch
Artifact' entry in this delta supersedes the over-strict
host-lookup text that the unarchived
`review-council-multi-model-fanout` delta carries under its
ADDED 'Review Dispatch Artifact' requirement (RD-FR-008).
One ordering constraint governs the delta: this change MUST
archive after `review-council-multi-model-fanout` — the
fanout archive creates the capability spec and the RDA
requirement the MODIFIED half replaces. Neither older change
is edited in place.

## MODIFIED Requirements

### Requirement: Review Dispatch Artifact

**ID: RD-FR-008** This requirement MUST be satisfied.

Every dispatch MUST write a version 1.0.0 `review-dispatch` payload.
It MUST use a Hero Interface Contract envelope that identifies
`the-divisor`, producer version, UTC timestamp, artifact and schema
versions, branch, commit, correlation id, and workflow id.

`review-dispatch` MUST be additive execution provenance. Canonical
review decisions MUST use `review-verdict` schema 2.0.0 under
RV-FR-001. Registry documentation, samples, producer and consumer
tables, and compatibility tests MUST describe dual emission and the
major-version consumer migration.

The closed payload MUST require `command`, `mode`, `full`,
`input_context`, `change_profile`, `plan_version`, `plan`, `runs`,
`coverage`,
`findings`, `advisories`, `verdict`, `verdict_reason`, `run_counts`,
`workflow_result`, and `correlation_id`. Unknown properties MUST be
rejected at every object level.

`command` MUST be `review-council`, `triage-issue`,
`address-feedback`, or `speckit-testreview`. `mode` MUST be `code`,
`specs`, `triage`, `feedback`, or `test`.

`input_context` MUST be a closed discriminated object. PR and local
kinds MUST require `kind`, `pr_number`, `base_ref`, `base_sha`,
`head_ref`, and `head_sha`. Refs MUST be 1-255 printable ASCII
characters. SHAs MUST be exactly 40 lowercase hexadecimal characters.
PR kind MUST have an integer PR number from 1 through 999999. Local
kind MUST have null PR number. The resolved SHAs MUST identify the
reviewed diff.

Issue kind MUST require only `kind`, `issue_number`, `issue_url`, and
`content_sha256`. Issue number MUST be from 1 through 999999. URL MUST
be a canonical credential-free HTTPS GitHub issue URL. Content hash
MUST be 64 lowercase hexadecimal characters. Git refs and SHAs MUST be
absent.

`change_profile` MUST be closed and discriminated by `kind`. Both kinds
MUST require boolean `security_sensitive`, unique sorted `categories`,
and tier. Categories MUST use the review-context Protocol 3 enum. Tier
MUST be `lightweight`, `standard`, or `heavy`.

Diff kind MUST require non-negative integer `lines`, `files`, and
`components`. Issue kind MUST instead require non-negative integer
`text_bytes` and `comment_count`, 64-character lowercase
`content_sha256`, and sorted unique `matched_rules`. Fields from the
other profile kind MUST be absent.

`plan_version` MUST equal 1. Each closed plan entry MUST use the
RD-FR-002 fields. Decision MUST be `include` or `skip`. Source MUST be
`explicit`, `advisor`, or `host`. Sequence MUST be a positive integer.
Tier, model, and variant MUST be nullable. Validation errors MUST be
unique bounded strings.

Each planned run MUST appear once in `runs`. A closed run MUST require
`run_id`, `agent`, `source`, `requested_model`, `provider`, `model_id`,
`variant`, `resolved_parent_model`, `resolved_parent_variant`,
`sequence`, `status`, `started_at`, `finished_at`, `usage`, `error`,
`reported_model`, `model_mismatch`, `workflow_verdict`, and `findings`.
Run ids MUST be unique UUIDs. Sequence MUST be positive.

For host source, requested model, provider, model id, and variant MUST
be null. Resolved parent model and variant MUST be null until lookup
succeeds. After successful lookup, resolved parent model MUST be
non-null and MUST form the exact model identity replayed to the child;
resolved parent variant conforms to the variant-or-null schema
contract and MAY be null when the provider reports no active variant.
For explicit source, requested model, provider, and model id MUST be
non-null and form the same model identity; both resolved-parent fields
MUST be null. For advisor source, the run MUST carry either explicit
model provenance — non-null requested model, provider, and model id
forming one identity, with both resolved-parent fields null — or
host-resolved provenance: null requested model, provider, model id,
and variant, with resolved parent model non-null after successful
lookup and resolved parent variant conforming to the variant-or-null
contract. A successful host or host-resolved advisor run MUST report
its child model. A reported model mismatch against the requested or
resolved model identity MUST make the run failed.

`resolved_parent_model` MUST be null or a 3-256 character model string
formed from the resolved provider and model id with one separating
slash. `resolved_parent_variant` MUST be null or satisfy the RD-FR-001
variant constraint. A host availability failure MAY leave either field
null but MUST include its sanitized error.

#### Scenario: Host run carries a null parent variant

- **GIVEN** a successful run with `source: "host"`
- **AND** `resolved_parent_model` is non-null and matches the
  child's reported model
- **AND** `resolved_parent_variant` is null because the provider
  reports no active variant
- **WHEN** the dispatch artifact is finalized
- **THEN** the run passes provenance validation
- **AND** the artifact records the null variant unchanged

#### Scenario: Host-resolved advisor carries a null parent variant

- **GIVEN** a successful run with `source: "advisor"` and
  `requested_model: null`
- **AND** `resolved_parent_model` is non-null and matches the
  child's reported model
- **AND** `resolved_parent_variant` is null
- **WHEN** the dispatch artifact is finalized
- **THEN** the run passes provenance validation

Status MUST be `pending`, `running`, `success`, `failed`, `skipped`,
`budget_skipped`, `limit_skipped`, or `cancelled`. Final artifacts MUST
contain only terminal states. Start and finish times MUST be nullable
UTC RFC3339 values, with finish not earlier than start.

Usage MUST be null or a closed object with nullable non-negative
`cost_usd` and nullable `tokens`. Tokens MUST be null or a closed
object with non-negative integer `input` and `output`. Missing usage
MUST remain null. Error MUST be null or a closed object requiring a
bounded `code`, sanitized bounded `message`, and boolean `retryable`.

A successful run MUST have null error and a non-null native verdict.
Council and test verdicts MUST use `APPROVE`,
`APPROVE WITH ADVISORIES`, or `REQUEST CHANGES`. Triage verdicts MUST
use `VALID`, `INVALID`, or `NEEDS-CLARIFICATION`. Feedback verdicts
MUST use `ACCEPT` or `AUTHOR-DECIDES`. Non-success states MUST have a
null workflow verdict and no findings. Failed runs MUST have an error.

Each run finding MUST be closed and require severity, category,
description, root cause, nullable file, and nullable positive line.
Severity MUST be `CRITICAL`, `HIGH`, `MEDIUM`, or `LOW`.
Consolidated findings MUST add a non-empty unique `run_ids` array.
Advisories MUST require severity, description, and contributing ids.

`coverage` MUST be closed and require `preflight_verdict`,
`checks_total`, and `checks_passed`. The verdict MUST be `PASS`,
`FAIL`, `SOFT_GATE`, or `NOT_RUN`. Counts MUST be non-negative,
and passed MUST NOT exceed total.

`run_counts` MUST require non-negative `total`, `success`, `failed`,
`skipped`, `budget_skipped`, `limit_skipped`, and `cancelled` values.
The JSON Schema MUST enforce fields, types, and bounds. A deterministic
semantic validator MUST require each value to equal the number of runs
in that terminal state and their sum to equal `total`. Final artifacts
MUST NOT contain `pending` or `running` runs. Final verdict MUST be
`APPROVE`,
`APPROVE WITH ADVISORIES`, `REQUEST CHANGES`, `INCONCLUSIVE`, or
`UNAVAILABLE`.

`workflow_result` MUST be a closed discriminated object with `kind`
and `value`. Kind MUST be `council`, `triage`, `feedback`, or
`test-review`. Its value MUST use the corresponding native verdict
union above, plus `INCONCLUSIVE` or `UNAVAILABLE`. Partial and failed
runs MUST remain represented.

The native workflow result MUST remain authoritative. Council and
test-review values MUST map identically to the generic final verdict.
Triage MUST map `VALID` to `APPROVE`, `INVALID` to
`REQUEST CHANGES`, and `NEEDS-CLARIFICATION` to
`APPROVE WITH ADVISORIES`. Feedback MUST map `ACCEPT` to `APPROVE`
and `AUTHOR-DECIDES` to `APPROVE WITH ADVISORIES`. Native
`INCONCLUSIVE` and `UNAVAILABLE` MUST map identically.

`UNAVAILABLE` MUST mean zero runnable or successful assessments
because of provider, model, or runtime availability. `INCONCLUSIVE`
MUST mean policy or plan validation failed, execution limits prevented
assessment,
artifact persistence failed, or another condition prevented calculation
of a valid assessment.

The schema MUST encode structural discriminators, types, required
fields, and nullability. The `review-dispatch` policy
plugin MUST expose `finalize_review_dispatch` to enforce cross-field
mappings, terminal-state rules, count arithmetic, and atomic
persistence. It MUST include valid, missing-field, extra-property,
bad-enum, invalid-state, and same-major compatibility fixtures. Every
fixture MUST run through schema and semantic validation. Positive and
negative fixtures MUST cover every native mapping and count invariant.
Compatibility MUST follow the Hero Interface Contract.

#### Scenario: Malformed matrix model path

- **GIVEN** a model has a trailing slash or repeated slash
- **WHEN** matrix validation runs
- **THEN** matrix validation rejects the model before dispatch
- **AND** no provider request is made

Files MUST use a unique correlation-id name under
`.uf/artifacts/dispatch/`. Directories MUST use mode 0750 and files
0600. Writes MUST use a same-directory temporary file. The writer MUST
flush, close, and chmod it before exclusive atomic publication through
a same-filesystem hard link. Link creation MUST fail with `EEXIST` when
the final target exists and MUST NOT overwrite it. After publication,
the writer MUST remove the temporary name. `EEXIST` MUST generate a new
correlation id and retry. Write failure MUST retain the calculated
assessment for human-only reporting. It MUST change the operation
result to `INCONCLUSIVE`, return a failing status, block automated
progression, and MUST NOT fabricate a review finding.

#### Scenario: Mixed run artifact

- **GIVEN** two runs succeed and one fails
- **WHEN** dispatch finalizes its artifact
- **THEN** all three outcomes and final verdict are present
- **AND** unavailable usage is null rather than fabricated

#### Scenario: Native result maps without loss

- **GIVEN** triage returns `NEEDS-CLARIFICATION`
- **WHEN** dispatch finalizes its artifact
- **THEN** the native result remains `NEEDS-CLARIFICATION`
- **AND** the generic verdict is `APPROVE WITH ADVISORIES`

#### Scenario: Terminal run counts reconcile

- **GIVEN** final runs include every supported terminal state
- **WHEN** dispatch calculates run counts
- **THEN** each distinct count matches its run state
- **AND** their sum equals total
- **AND** no pending or running run is present

#### Scenario: Pull request input is reproducible

- **GIVEN** a review targets a pull request not checked out locally
- **WHEN** dispatch writes its final artifact
- **THEN** input context records PR number and base and head refs
- **AND** both immutable resolved commit SHAs are present

#### Scenario: Exclusive publication collision

- **GIVEN** the chosen target exists or appears during publication
- **WHEN** the writer creates the exclusive hard link
- **THEN** it chooses a new correlation id
- **AND** the existing artifact remains unchanged

#### Scenario: Artifact persistence fails

- **GIVEN** a calculated assessment exists
- **WHEN** its required artifact cannot be written
- **THEN** the operation returns failing `INCONCLUSIVE` status
- **AND** the assessment is non-authoritative human context only

Previously: the host-lookup clause required that, after successful
lookup, both resolved-parent fields be non-null and form the exact
pair and active variant replayed to the child, and every explicit or
advisor run carried non-null requested provenance with null
resolved-parent fields. This change corrects the host-lookup clause
to require only a non-null resolved parent model — the resolved
parent variant is variant-or-null — and recognizes host-resolved
advisor runs (advisor source with a null requested model) under the
host provenance contract, superseding the text the
`review-council-multi-model-fanout` delta ADDED under this
requirement name.

## REMOVED Requirements

None.
