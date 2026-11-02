<!-- capability: fetch-pr-review-context
     change: fix-uf-finale-ci-poll-gh-schema
     originating_issue: 895 -->

Requirement numbering: FPR-FR-017 is allocated for this
change. FPR-FR-018 through FPR-FR-027 are reserved for future
fetch-pr-review-context changes. The capability's existing bare
ids FR-001..FR-010 belong to the origin fetch-pr-review-context
change and FR-013..FR-016 to extend-fetch-pr-review-context;
they are preserved, not renumbered (MP-002). This change
MODIFIES FR-006 under its existing id to correct the
`ci_pending` derivation. A MODIFIED requirement is restated in
full, so the restatement also composes the `exceeds_threshold`
bullet from the unarchived sibling change
review-dispatch-diff-manifest (RD-FR-032) instead of reverting
it to origin-era text. All other existing requirements are
preserved unchanged. FR-011 and FR-012 were never allocated by
either prior delta and are retired here; they MUST NOT be
reused, and future changes allocate from FPR-FR-018 onward. The
unnumbered "Command file Steps 0-3.5" requirement lives in the
origin delta. The sibling `uf_finale_ci_poll` output-contract
requirement lives in finale-token-reduction's plugin-tools
delta. Requirement ids MUST stay inside the allocated range.

Archive ordering: this change MUST archive after
review-dispatch-diff-manifest. Both changes carry a MODIFIED
FR-006 complete restatement, and OpenSpec archive replaces the
requirement wholesale. Archiving in the opposite order (this
change first) would revert this change's bucket-derived
`ci_pending` eligibility fix, because the sibling's
restatement retains the repealed status-based rule. The
sibling's `exceeds_threshold` manifest-delivery contract
(RD-FR-032) survives either order, because the sibling's own
MODIFIED FR-006 carries it. This mirrors the uf-finale-ci-poll
delta's ordering note against finale-token-reduction.

## ADDED Requirements

### Requirement: ci_checks Sub-Call gh Schema Alignment

**ID: FPR-FR-017** This requirement MUST be satisfied.

The `fetch_pr_review_context` tool's advisory `ci_checks`
sub-call MUST request `gh pr checks <N> --json name,state,bucket`
— only fields verified against gh 2.91.0 — through the
exit-code-tolerant gh runner (`runGhCommandTolerant`, the same
path `fetch_pr_checks` uses). On ANY exit code, stdout that
parses as a JSON array MUST be treated as check-status data;
stdout that does not parse as a JSON array — including valid
JSON that is not an array — without a recognized marker is a
sub-call failure. This makes pending (exit 8) and failing
(exit 1) PRs populate `ci_checks` instead of degrading.

Recognized stderr markers: this advisory sub-call recognizes
exactly one marker — `/no checks reported/i`
(`NO_CHECKS_REPORTED_PATTERN`) — and only when stdout is empty.
That verified checkless-PR shape (exit 1, empty stdout, stderr
matching the pattern) MUST yield `ci_checks: []`,
`ci_pending: false`, and NO `ci_checks degraded` warning. The
PR-not-found phrases (`/not found|no pull request|could not
find|could not resolve to a pullrequest/i`) are NOT recognized
by this sub-call: stdout that does not parse as a JSON array
whose stderr matches them MUST append a `ci_checks degraded`
warning and continue with empty `ci_checks`, because the
preceding `gh pr view` call owns PR-not-found detection
(`pr_not_found`). Every other stdout that does not parse as a
JSON array is likewise a non-fatal degradation.

Each check MUST map into the existing internal
`{name, status, conclusion}` shape with `status` sourced from
`state` and `conclusion` sourced from `bucket`. `ci_pending`
MUST be derived from the observed bucket vocabulary (`true` when
any check buckets as `pending`, else `false`), never from
comparing `status` to the lowercase `"completed"` value gh does
not emit. The published consumer-facing context shape MUST NOT
change: `ci_checks` entries keep
`{name, status, conclusion}`.

The sub-call MUST remain non-fatal on a genuine failure: a
transport error, stdout that does not parse as a JSON array
(empty, malformed, or valid non-array JSON) without a
recognized marker, or a gh argument-validation error MUST append
a `ci_checks degraded` warning and continue with empty
`ci_checks`, exactly as before.

Unknown-bucket divergence: an unrecognized `bucket` value is
classified differently by each CI-check consumer, and each
choice is the conservative one for its job.
`uf_finale_ci_poll` treats it as failed so vocabulary drift
surfaces immediately (FCP-FR-003). `fetch_pr_checks` treats it
as `pending` (fail-closed repair eligibility) and appends a
warning naming the value. This advisory sub-call leaves
`ci_pending` false and surfaces the unknown value verbatim in
`conclusion`. A new consumer MUST choose its unknown-bucket
behavior deliberately and document it; it MUST NOT inherit one
of these defaults implicitly.

#### Scenario: Green PR populates ci_checks without degradation

- **GIVEN** PR #42's checks return
  `[{name: "CI / build", state: "SUCCESS", bucket: "pass"}]`
  from `gh pr checks --json name,state,bucket`
- **WHEN** `fetch_pr_review_context` is called with
  `{pr_number: 42}`
- **THEN** the returned context's `ci_checks` equals
  `[{name: "CI / build", status: "SUCCESS", conclusion: "pass"}]`
- **AND** `ci_pending` is `false`
- **AND** no `ci_checks degraded` warning is appended

#### Scenario: Pending PR on exit 8 populates ci_checks

- **GIVEN** PR #42's checks return
  `[{name: "CI / build", state: "PENDING", bucket: "pending"}]`
  from `gh pr checks --json name,state,bucket` with exit 8
- **WHEN** `fetch_pr_review_context` is called with
  `{pr_number: 42}`
- **THEN** the returned context's `ci_checks` equals
  `[{name: "CI / build", status: "PENDING",
  conclusion: "pending"}]`
- **AND** `ci_pending` is `true`
- **AND** no `ci_checks degraded` warning is appended

#### Scenario: Failing PR on exit 1 populates ci_checks

- **GIVEN** PR #42's checks return
  `[{name: "CI / build", state: "FAILURE", bucket: "fail"}]`
  from `gh pr checks --json name,state,bucket` with exit 1
- **WHEN** `fetch_pr_review_context` is called with
  `{pr_number: 42}`
- **THEN** the returned context's `ci_checks` equals
  `[{name: "CI / build", status: "FAILURE",
  conclusion: "fail"}]`
- **AND** `ci_pending` is `false`
- **AND** no `ci_checks degraded` warning is appended

#### Scenario: Genuine sub-call failure still degrades

- **GIVEN** `gh pr checks` produces stdout that does not parse
  as a JSON array without a recognized marker, or the invocation
  fails
- **WHEN** `fetch_pr_review_context` is called
- **THEN** the tool appends a `ci_checks degraded` warning
  carrying the error message and returns the context with empty
  `ci_checks` rather than failing

#### Scenario: Checkless PR marker yields empty ci_checks

- **GIVEN** PR #42 has no checks and `gh pr checks` exits 1
  with empty stdout and
  `no checks reported on the 'feature/x' branch` on stderr
- **WHEN** `fetch_pr_review_context` is called with
  `{pr_number: 42}`
- **THEN** `ci_checks` is an empty array
- **AND** `ci_pending` is `false`
- **AND** no `ci_checks degraded` warning is appended

#### Scenario: PR-not-found stderr degrades the advisory sub-call

- **GIVEN** `gh pr checks` exits non-zero with empty stdout and
  `Could not resolve to a PullRequest with the number of 42`
  on stderr
- **WHEN** `fetch_pr_review_context` is called with
  `{pr_number: 42}`
- **THEN** the tool appends a `ci_checks degraded` warning
  carrying that stderr and returns the context with empty
  `ci_checks` rather than a `pr_not_found` failure

#### Scenario: Unknown JSON fields are never requested

- **GIVEN** gh 2.91.0 rejects `conclusion` as a
  `gh pr checks --json` field
- **WHEN** the sub-call's gh invocation is inspected
- **THEN** the requested field list is exactly
  `name,state,bucket`

#### Scenario: Unknown bucket surfaces verbatim

- **GIVEN** PR #42's checks return
  `[{name: "CI / build", state: "NEUTRAL",
  bucket: "blocked"}]`
  from `gh pr checks --json name,state,bucket`
- **WHEN** `fetch_pr_review_context` is called with
  `{pr_number: 42}`
- **THEN** the returned context's `ci_checks` equals
  `[{name: "CI / build", status: "NEUTRAL",
  conclusion: "blocked"}]`
- **AND** `ci_pending` is `false`
- **AND** no `ci_checks degraded` warning is appended

## MODIFIED Requirements

### Requirement: FR-006 Pre-computed fields

**ID: FR-006** This requirement MUST be satisfied.

The tool MUST pre-compute these fields from raw PR data:

- `input_context`: formatted string containing kind, PR
  number, base/head refs and SHAs
- `changed_files`: formatted list of changed files with
  additions/deletions counts
- `exceeds_threshold`: true when the diff exceeds 1500 lines
  (`DIFF_THRESHOLD`), computed with the shared line-counting
  function from the review-diff-manifest lib module — the same
  function `build_review_prompts` uses for delivery-mode
  selection — so the field is true if and only if review
  dispatch engages manifest delivery (RD-FR-032)
- `ci_pending`: true when any check's gh `bucket` is
  `pending`, false otherwise

Previously: `ci_pending` was true when any CI check had a
`status` not equal to `completed`. That comparison is repealed
because gh 2.91.0 emits uppercase `state` values (`SUCCESS`,
`PENDING`, `FAILURE`, ...) that never equal the lowercase
literal, so the old rule was always true once any check
existed. FPR-FR-017 supersedes it with the bucket-derived rule.
The `exceeds_threshold` contract is not changed by this change:
its bullet is composed from the unarchived sibling
review-dispatch-diff-manifest (RD-FR-032), which replaced the
origin-era size-advisory wording, so this restatement preserves
that sibling's shared-counting-function and
iff-manifest-delivery semantics rather than reverting them.

#### Scenario: ci_pending derives from the check bucket

- **GIVEN** a PR whose checks include one with
  `bucket: "pending"`
- **WHEN** `fetch_pr_review_context` succeeds
- **THEN** `ci_pending` is `true`
- **AND** a PR whose checks are all `bucket: "pass"` reports
  `ci_pending` false

#### Scenario: Boundary agreement with prompt assembly

- **GIVEN** a PR whose diff contains exactly 1500 lines
  and another containing exactly 1501 lines
- **WHEN** `fetch_pr_review_context` succeeds for both
- **THEN** the 1500-line diff reports `exceeds_threshold`
  false and the 1501-line diff reports it true
- **AND** `build_review_prompts` engages manifest mode for
  exactly the diffs for which `exceeds_threshold` is true
  (both computations use the shared counting function)

## REMOVED Requirements

None.
