## ADDED Requirements

### FR-001: Feedback Ingestion

The `/address-feedback` command MUST fetch all review
feedback from a GitHub PR using the `gh` CLI. The
command MUST accept an optional PR number argument. If
no argument is provided, the command MUST auto-detect
the PR number from the current branch using
`gh pr view --json number`.

The command MUST require `gh auth status` to succeed
before proceeding. If auth fails, the command MUST
report an actionable error: "GitHub CLI not
authenticated. Run `gh auth login` to authenticate."

The command MUST fetch:
- PR reviews (APPROVED, CHANGES_REQUESTED, COMMENTED)
- Inline review comments with file and line context
- General issue comments
- Review thread state (resolved / unresolved)
- Reviewer association (`author_association` field)

The command MUST handle paginated GitHub API responses
to ensure all reviews and comments are fetched, not
just the first page.

If a GitHub API call fails (network error, 5xx, 403
rate limit), the command MUST report the error and stop
rather than silently proceeding with partial data. If
the auth token lacks sufficient scopes for read
operations, the command MUST detect and report with
actionable guidance.

The command MUST filter out:
- Already-resolved threads
- The PR author's own top-level comments
- Pure approval reviews with no inline comments

The command MUST group threaded conversations (a
comment and its replies) into a single feedback item.
The assessment MUST consider the latest state of the
thread, not just the opening comment.

The command MUST detect GitHub suggestion blocks
and preserve them as structured data for the triage
phase.

#### Scenario: Fetch feedback for explicit PR number
- **GIVEN** a PR #42 exists with 3 reviews containing
  8 inline comments (2 resolved, 1 author-owned),
  and 2 general issue comments
- **WHEN** the author runs `/address-feedback 42`
- **THEN** the command fetches all reviews and comments
  from PR #42 and presents 7 discrete feedback items
  (8 inline - 2 resolved - 1 author-owned + 2 general)

#### Scenario: Auto-detect PR from current branch
- **GIVEN** the author is on branch `opsx/my-feature`
  which has an open PR
- **WHEN** the author runs `/address-feedback` without
  arguments
- **THEN** the command resolves the PR number via
  `gh pr view` and proceeds with ingestion

#### Scenario: No open PR for current branch
- **GIVEN** the author is on a branch with no open PR
- **WHEN** the author runs `/address-feedback` without
  arguments
- **THEN** the command reports an error: "No open PR
  found for the current branch. Specify a PR number."

#### Scenario: Filter resolved threads
- **GIVEN** a PR has 5 review threads, 2 of which are
  resolved
- **WHEN** the command ingests feedback
- **THEN** only 3 unresolved threads are presented to
  the author

#### Scenario: Group threaded conversations
- **GIVEN** a review comment has 3 replies forming a
  discussion
- **WHEN** the command ingests feedback
- **THEN** the entire thread is presented as a single
  feedback item with the full conversation context

#### Scenario: GitHub API failure
- **GIVEN** the GitHub API returns a 403 rate limit
  error when fetching review comments
- **WHEN** the command attempts to ingest feedback
- **THEN** the command reports "GitHub API error (403):
  rate limit exceeded. Wait and retry." and stops
  without presenting partial data

#### Scenario: gh auth status failure
- **GIVEN** the `gh` CLI is not authenticated
- **WHEN** the author runs `/address-feedback 42`
- **THEN** the command reports "GitHub CLI not
  authenticated. Run `gh auth login`." and stops

### FR-002: Tiered Assessment

The command MUST classify each feedback item into one
of two assessment tiers.

**Tier 1 (direct)**: The command agent MUST assess the
item directly using loaded convention packs and
constitution. Tier 1 MUST be used when ALL of the
following conditions are met:
- Single file affected
- Clear match to a convention pack rule or no rule
  applies (purely subjective)
- No security implications
- No architectural implications
- Reviewer and project standards do not conflict

**Tier 2 (Divisor escalation)**: The command MUST
delegate assessment to the relevant Divisor agent via
the Task tool when ANY of the following conditions
are met:
- Security concern raised
- Architectural change suggested
- Multi-file impact
- Feedback contradicts a convention pack rule
- Ambiguous classification (could be data-driven or
  subjective)
- Test strategy or coverage concern
- Performance or operational concern

The command MUST route Tier 2 items to the appropriate
Divisor agent based on the feedback domain. If multiple
domains apply, the command MUST invoke multiple agents
in parallel.

If no Divisor agents are available (not deployed), all
items MUST fall back to Tier 1 assessment. The
assessment output MUST include a `tier2_unavailable`
indicator when fallback occurs.

#### Scenario: Tier 1 assessment for simple naming
- **GIVEN** a reviewer suggests renaming a private
  variable
- **WHEN** the assessment engine evaluates the item
- **THEN** the item is assessed at Tier 1 with
  classification SUBJECTIVE and the `tier` field set
  to 1 in the assessment output

#### Scenario: Tier 2 escalation for security concern
- **GIVEN** a reviewer flags missing input validation
  on an endpoint
- **WHEN** the assessment engine evaluates the item
- **THEN** the item is escalated to Tier 2 and routed
  to `divisor-adversary` with `divisor_agents_used`
  containing `["adversary"]`

#### Scenario: Tier 2 with multiple domains
- **GIVEN** a reviewer suggests a refactor that changes
  both the architecture and test strategy
- **WHEN** the assessment engine evaluates the item
- **THEN** the item is routed to both
  `divisor-architect` and `divisor-testing` in parallel
  with `divisor_agents_used` containing
  `["architect", "testing"]`

#### Scenario: Fallback when Divisor agents absent
- **GIVEN** no Divisor agents are deployed in the
  project
- **WHEN** the assessment engine encounters an item
  that would normally trigger Tier 2
- **THEN** the item is assessed at Tier 1 with
  `tier2_unavailable` set to true in the assessment
  output

### FR-003: Feedback Classification

For each feedback item, the assessment MUST produce:
- **Classification**: DATA-DRIVEN or SUBJECTIVE
- **Evidence**: specific convention pack section,
  constitution principle, or coding standard reference
  that supports or refutes the feedback
- **Reviewer authority**: maintainer, collaborator,
  contributor, external, or bot (derived from
  `author_association` and bot detection)
- **Recommendation**: one of ACCEPT or AUTHOR-DECIDES
- **Suggested approach**: a concrete description of
  how to implement the fix (if recommendation is
  ACCEPT)
- **Conflict flag**: if another feedback item on the
  same PR provides contradictory guidance

DATA-DRIVEN classification MUST be used when the
feedback is grounded in a verifiable project rule
(convention pack, constitution, coding standard, lint
rule) or identifies a demonstrable defect (logic error,
missing error handling, security vulnerability).

SUBJECTIVE classification MUST be used when the
feedback reflects personal preference, stylistic
choice, or an alternative approach not mandated by
project rules.

#### Scenario: Data-driven classification with evidence
- **GIVEN** a reviewer says "this timeout should be
  configurable, not hardcoded"
- **WHEN** the assessment classifies the item
- **THEN** the classification is DATA-DRIVEN with
  evidence referencing coding-standards.md Section I
  (Single Source of Truth): "Magic numbers MUST NOT
  appear inline"

#### Scenario: Subjective classification
- **GIVEN** a reviewer says "I'd name this
  handleRequest instead of processReq"
- **WHEN** the assessment classifies the item
- **THEN** the classification is SUBJECTIVE with
  evidence field empty and recommendation
  AUTHOR-DECIDES

#### Scenario: Bot feedback validation
- **GIVEN** a bot reviewer flags "function too long"
  but no project convention enforces a maximum function
  length
- **WHEN** the assessment classifies the item
- **THEN** the classification is SUBJECTIVE with
  reviewer_role "bot" and recommendation
  AUTHOR-DECIDES

### FR-004: Authority-Weighted Recommendations

The command MUST apply the authority matrix to produce
recommendations:

| Authority | Data-Driven | Subjective |
|---|---|---|
| Maintainer | ACCEPT (MUST fix) | AUTHOR-DECIDES (SHOULD consider) |
| Collaborator | ACCEPT (MUST fix) | AUTHOR-DECIDES |
| Contributor | ACCEPT (MUST fix) | AUTHOR-DECIDES |
| External | ACCEPT if validated | AUTHOR-DECIDES |
| Bot | ACCEPT if validated | AUTHOR-DECIDES (informational) |

Reviewer authority MUST be determined from the GitHub
API `author_association` field:
- `OWNER`, `MEMBER` → maintainer
- `COLLABORATOR` → collaborator
- `CONTRIBUTOR`, `FIRST_TIMER`,
  `FIRST_TIME_CONTRIBUTOR` → contributor
- `NONE` without bot indicators → external
- `NONE` with bot indicators → bot

Bot detection MUST check: login ending in `[bot]` or
account type `Bot`.

#### Scenario: Maintainer data-driven feedback
- **GIVEN** a maintainer flags a missing error wrap
- **WHEN** the authority matrix is applied
- **THEN** the recommendation is ACCEPT with note
  "MUST fix (maintainer, data-driven)"

#### Scenario: Maintainer subjective feedback
- **GIVEN** a maintainer suggests an alternative
  variable name
- **WHEN** the authority matrix is applied
- **THEN** the recommendation is AUTHOR-DECIDES with
  note "SHOULD consider (maintainer preference)"

#### Scenario: Bot data-driven with valid rule
- **GIVEN** a bot flags a lint violation that matches
  the project's `.golangci.yml` configuration
- **WHEN** the authority matrix is applied
- **THEN** the recommendation is ACCEPT with note
  "Validated against project lint configuration"

#### Scenario: External non-bot reviewer
- **GIVEN** a user with `author_association: NONE` who
  is not a bot provides data-driven feedback citing a
  convention pack rule
- **WHEN** the authority matrix is applied
- **THEN** the reviewer_role is "external" and the
  recommendation is ACCEPT (validated against the
  convention pack)

### FR-005: Author Triage

The command MUST present each feedback item to the
author one-by-one with the full assessment context.
Each presentation MUST include:
- Reviewer identity and role
- File and line reference
- Full thread content (all comments in the thread)
- Classification (DATA-DRIVEN / SUBJECTIVE)
- Evidence references
- Recommendation (ACCEPT / AUTHOR-DECIDES)
- Suggested approach (if applicable)
- Conflict flag (if applicable)

The author MUST choose exactly one decision per item:
- **ACCEPT**: queue code change as recommended
- **MODIFY**: author provides alternative approach,
  queue code change with that approach
- **REJECT**: author provides reasoning, queue reply
  comment
- **ASK**: author provides question, queue reply
  comment asking for clarification

No item MAY be skipped or deferred. Every item MUST
receive a decision before the triage phase completes.

After all items are triaged, the command MUST present
a summary showing all decisions and queued actions.
The author MUST confirm before execution proceeds.

#### Scenario: Accept a recommendation
- **GIVEN** item #1 is assessed as DATA-DRIVEN with
  recommendation ACCEPT
- **WHEN** the author selects ACCEPT
- **THEN** a code change is queued using the suggested
  approach

#### Scenario: Modify with alternative approach
- **GIVEN** item #2 is assessed with recommendation
  ACCEPT
- **WHEN** the author selects MODIFY and provides an
  alternative implementation approach
- **THEN** a code change is queued using the author's
  approach instead, with the commit message noting
  "modified approach" in the description

#### Scenario: Reject with reasoning
- **GIVEN** item #3 is assessed as SUBJECTIVE with
  recommendation AUTHOR-DECIDES
- **WHEN** the author selects REJECT and provides
  reasoning
- **THEN** a reply comment is queued with the author's
  evidence-based reasoning

#### Scenario: Ask for clarification
- **GIVEN** item #4 has ambiguous intent
- **WHEN** the author selects ASK and provides a
  question
- **THEN** a reply comment is queued with the author's
  question to the reviewer

### FR-006: Batch Execution

After triage confirmation, the command MUST execute
all queued actions:

1. Implement all accepted and modified code changes
2. Create one commit per fix with conventional commit
   format referencing the PR number and reviewer
3. Run `/review-council` on the cumulative changes
4. If review-council passes: push all commits to the
   PR branch
5. Post reply comments to the PR for each item:
   - Accepted: "Addressed in `<commit_sha>`" with
     brief description
   - Rejected: evidence-based reasoning
   - Asked: author's clarification question
6. Offer to resolve threads for accepted items

All PR comment posting MUST require author
confirmation before execution.

If `/review-council` fails and the fix loop exhausts
iterations, the command MUST stop and report the
persistent findings. Code MUST NOT be pushed until
review-council passes.

**Error recovery**: If `git push` fails (network
error, branch protection rejection, or diverged
history), the command MUST preserve the local commits
and report which commits are stranded with guidance
to retry. If comment posting fails partway through
(e.g., API error after 3 of 6 comments posted), the
command MUST track which comments were successfully
posted in the cache and report partial progress. On
re-invocation, already-posted comments MUST be
skipped.

Before pushing, the command MUST fetch the remote
branch state and warn if the branch has diverged
(another contributor pushed commits). The author MUST
confirm before force-pushing or rebasing.

If implementing a code change fails (e.g., the
suggestion cannot be applied cleanly), the command
MUST skip that item with a report and continue with
remaining items. Skipped items are not committed and
the reply comment notes the failure.

#### Scenario: Successful execution
- **GIVEN** 4 items accepted, 1 rejected, 1 asked
- **WHEN** execution proceeds after author confirmation
- **THEN** 4 commits are created, review-council
  passes, commits are pushed, 6 reply comments are
  posted (4 "addressed in", 1 rejection reasoning,
  1 clarification question)

#### Scenario: Review-council blocks push
- **GIVEN** code changes introduce a new lint violation
- **WHEN** review-council runs on the changes
- **THEN** the fix loop attempts to resolve it; if
  exhausted, execution stops with a report and no code
  is pushed

#### Scenario: Push fails due to diverged branch
- **GIVEN** another contributor pushed to the PR branch
  between triage and execute
- **WHEN** the command attempts to push
- **THEN** the command detects the divergence, reports
  it to the author, and offers to rebase or abort

#### Scenario: Partial comment posting failure
- **GIVEN** 3 of 6 reply comments are posted and the
  4th fails due to API rate limit
- **WHEN** the command encounters the failure
- **THEN** the command reports which comments were
  posted (items 1-3) and which are pending (items
  4-6), and records progress in the cache for retry

#### Scenario: Code change cannot be applied
- **GIVEN** an accepted suggestion references code
  that has changed since the review was posted
- **WHEN** the command attempts to implement the fix
- **THEN** the command skips that item with a report,
  continues with remaining items, and the reply
  comment notes the fix could not be applied

### FR-007: PR-Scoped Cache

The command MUST maintain a local cache under
`.uf/feedback/pr-<N>/` keyed by PR number. The cache
MUST store:
- Per-thread assessment results
- Comment count at time of assessment
- Last comment ID per thread
- Author decisions (once made)
- Execution status per item (committed, pushed,
  comment-posted) for crash-recovery idempotency
- Last-fetched timestamp

Cache invalidation rules:
- Thread has new comments → re-assess the thread
- Code at referenced lines changed (detected by
  comparing the current file content at the referenced
  line range against the content at assessment time) →
  mark stale, re-assess
- Thread resolved on GitHub → skip entirely
- Cache file missing → full reconstruction from
  GitHub

The cache MUST be treated as disposable. The command
MUST produce correct results even if the cache is
deleted between invocations.

The cache directory MUST be added to `.gitignore`.
Cache files MUST be created with restrictive
permissions (600 for files, 700 for directories)
since they may contain security-sensitive review
content.

On re-invocation after a crash during Phase 4, items
marked as fully executed (comment-posted) in the cache
MUST be skipped to prevent duplicate comments.

Cache directories for merged or closed PRs SHOULD be
pruned periodically. The cache directory can be safely
deleted at any time without affecting correctness.

#### Scenario: Cache hit on unchanged thread
- **GIVEN** round 1 assessed thread T1, round 2 runs
  and T1 has no new comments
- **WHEN** the command processes T1 in round 2
- **THEN** the cached assessment is reused without
  re-invoking Tier 1 or Tier 2 analysis

#### Scenario: Cache miss with full reconstruction
- **GIVEN** the `.uf/feedback/` directory was deleted
- **WHEN** the author runs `/address-feedback 42`
- **THEN** all threads are assessed from scratch and
  results are correct (but slower)

#### Scenario: Multiple PRs cached independently
- **GIVEN** the author has PR #42 and PR #45 under
  review
- **WHEN** the author runs `/address-feedback 42`
  then `/address-feedback 45`
- **THEN** each PR has its own independent cache
  directory and assessments do not interfere

#### Scenario: Stale cache from code changes
- **GIVEN** round 1 assessed an item at line 42, then
  the author pushed a commit changing that line
- **WHEN** round 2 runs and finds the content at line
  42 differs from what was cached
- **THEN** the cached assessment is marked stale and
  the item is re-assessed with the current code context

#### Scenario: Crash recovery with partial execution
- **GIVEN** a previous invocation committed and pushed
  fixes for items 1-3 but crashed before posting reply
  comments
- **WHEN** the author re-runs `/address-feedback 42`
- **THEN** items 1-3 are recognized as partially
  executed (code pushed, comments pending) and only
  the comment-posting step is retried

### FR-008: Feedback-Triage Artifact

The command MUST produce a JSON artifact at
`.uf/artifacts/feedback-triage/pr-<N>-round-<M>.json`
after each invocation. The artifact MUST be wrapped in
the standard envelope schema.

Envelope `context` field mapping:
- `context.branch`: the PR's source branch at artifact
  production time
- `context.commit`: HEAD commit SHA after all fixes
  are applied and pushed (or pre-fix HEAD if no code
  changes were made)
- `context.backlog_item_id`: linked issue number or PR
  number (prefixed with `PR-`)

The envelope `hero` field MUST be set to
`"cobalt-crush"` (the agent executing the command),
consistent with how implementation-time artifacts
identify their producer.

The payload `branch` field is the PR's source branch,
which is the same as `context.branch`. This
redundancy is intentional to allow payload-only
consumers to access branch information without parsing
the envelope.

Artifact writes MUST be atomic: write to a temporary
file, then rename to the final path. This prevents
truncated artifacts from downstream consumer parse
failures.

The round number MUST be derived from the highest
existing round number + 1 (not a count of files),
to handle gaps from manually deleted files.

The artifact payload MUST include:
- `pr_number`: integer
- `pr_url`: string
- `branch`: string
- `round`: integer (increments per invocation on the
  same PR)
- `items`: array of per-item records, each containing:
  - `thread_id`: string
  - `reviewer`: string (GitHub login)
  - `reviewer_role`: string (maintainer / collaborator
    / contributor / external / bot)
  - `file`: string or null (file path, null for
    general PR comments)
  - `line`: integer or null (line number, null for
    general PR comments)
  - `classification`: string (data-driven / subjective)
  - `tier`: integer (1 or 2)
  - `evidence`: array of strings (pack/rule references)
  - `recommendation`: string (accept / author-decides)
  - `decision`: string (accept / modify / reject / ask)
  - `decision_reasoning`: string or null
  - `commit_sha`: string or null (if code change)
  - `divisor_agents_used`: array of strings (if Tier 2)
  - `tier2_unavailable`: boolean (true when Tier 2
    fallback occurred because no Divisor agents were
    available, default false)
  - `conflict_flag`: boolean (true when this item
    conflicts with another item on the same PR,
    default false)
- `summary`: object containing:
  - `total_items`: integer
  - `accepted`: integer
  - `modified`: integer
  - `rejected`: integer
  - `asked`: integer
  - `tier1_count`: integer
  - `tier2_count`: integer
  - `divisor_agents_invoked`: array of strings

#### Scenario: Artifact produced after triage
- **GIVEN** the author completes triage on PR #42
  round 1 with 6 items (3 accepted, 1 modified,
  1 rejected, 1 asked)
- **WHEN** execution completes
- **THEN** a file at `.uf/artifacts/feedback-triage/
  pr-42-round-1.json` is created with full envelope
  provenance, 6 per-item records with decisions, and
  summary showing `total_items: 6, accepted: 3,
  modified: 1, rejected: 1, asked: 1`

#### Scenario: Round number increments
- **GIVEN** `pr-42-round-1.json` already exists
- **WHEN** the author runs `/address-feedback 42` a
  second time
- **THEN** `pr-42-round-2.json` is created with round
  field set to 2

#### Scenario: Round number handles gaps
- **GIVEN** `pr-42-round-1.json` and
  `pr-42-round-3.json` exist (round 2 was deleted)
- **WHEN** the author runs `/address-feedback 42`
- **THEN** `pr-42-round-4.json` is created (highest
  existing round 3 + 1)

### FR-009: Conflict Detection

The command MUST detect when two or more feedback items
on the same PR provide contradictory guidance for the
same code section. Conflict detection MUST compare
feedback items that reference overlapping file and line
ranges.

When a conflict is detected, both items MUST be
flagged to the author with a CONFLICT indicator. The
author MUST choose one approach and the command MUST
queue a reply comment to the non-chosen reviewer
explaining the decision.

#### Scenario: Two reviewers suggest different approaches
- **GIVEN** reviewer @alice suggests "use a mutex" on
  line 42 and reviewer @bob suggests "use channels"
  on line 42
- **WHEN** the assessment engine processes both items
- **THEN** both items are flagged with CONFLICT and
  presented together so the author can decide

#### Scenario: No conflict on different files
- **GIVEN** reviewer @alice comments on `proxy.go:42`
  and reviewer @bob comments on `handler.go:15`
- **WHEN** the assessment engine processes both items
- **THEN** neither item is flagged with CONFLICT

### FR-010: Context Discovery

The command MUST load project context for assessment:
- Convention packs from `.opencode/uf/packs/`
- Constitution from `.specify/memory/constitution.md`
- AGENTS.md coding and testing conventions
- Coding standards from user config if present

The command SHOULD discover the associated spec for
the PR branch (Speckit: `specs/NNN-*/`, OpenSpec:
`openspec/changes/*/`) and load its design decisions
for assessment context.

The command SHOULD parse linked issues from the PR
description (`Fixes #N`, `Closes #N`, `Resolves #N`)
and load their acceptance criteria.

If the `review-context` convention pack (#176) is
available, the command MUST use it for standardized
context discovery. If absent, the command MUST inline
the discovery logic.

#### Scenario: Context loaded with review-context pack
- **GIVEN** the `review-context.md` pack exists
- **WHEN** the command initializes
- **THEN** context discovery follows the pack's
  protocol for spec discovery, issue linking, and
  path classification

#### Scenario: Context loaded without pack
- **GIVEN** no `review-context.md` pack exists
- **WHEN** the command initializes
- **THEN** context discovery is inlined using the same
  logic as `/review-pr` Steps 6-7

## Test Strategy

This change introduces a slash command (agent
instructions) and supporting schema artifacts. The
testable surface is:

### Unit Tests (Go)
- **Schema validation**: validate
  `v1.0.0.schema.json` is well-formed JSON Schema
  using the project's `internal/schemas/` validation
  infrastructure
- **Positive sample validation**: validate
  `sample-feedback-triage.json` against the schema
- **Negative sample validation**: validate that the
  schema rejects invalid inputs (missing required
  fields, unknown `decision` values, extra properties
  on item objects). Create negative test fixtures in
  `schemas/feedback-triage/samples/`
- **Scaffold asset tests**:
  `TestAssetPaths_MatchExpected` (new asset in list),
  `TestEmbeddedAssets_MatchSource` (drift detection),
  `TestRun_CreatesFiles` (file count)
- Coverage target: 100% for schema validation paths,
  100% for scaffold asset test assertions

### Contract Tests
- Schema `additionalProperties: false` enforcement:
  verify that extra fields on payload, item, and
  summary objects cause validation failure
- Enum constraints: verify `classification`,
  `decision`, `recommendation`, `reviewer_role`
  fields reject out-of-enum values
- Required field enforcement: verify each required
  field causes validation failure when absent

### Integration Tests
- Not applicable: the command is agent instructions
  executed within the OpenCode runtime, not compiled
  Go code with integration boundaries

### Coverage Target
- New Go test code: 100% branch coverage for schema
  validation logic
- Schema sample coverage: at least one sample per
  decision type (accept, modify, reject, ask), one
  with null file/line, one with conflict flag, one
  with bot reviewer, one with multiple
  divisor_agents_used

## MODIFIED Requirements

None. This change does not modify existing command
behavior.

## REMOVED Requirements

None. No existing functionality is removed.
<!-- scaffolded by uf vdev -->
