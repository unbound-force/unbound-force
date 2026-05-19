<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Feedback-Triage Schema

- [x] 1.1 [P] Create `schemas/feedback-triage/
  v1.0.0.schema.json`. Define the payload schema
  following the same conventions as
  `schemas/review-verdict/v1.0.0.schema.json`:
  - `pr_number` (integer, required)
  - `pr_url` (string, required)
  - `branch` (string, required)
  - `round` (integer, required)
  - `items` (array, required) with per-item object:
    `thread_id`, `reviewer`, `reviewer_role`
    (enum: maintainer/collaborator/contributor/
    external/bot), `file` (string or null),
    `line` (integer or null),
    `classification` (enum: data-driven/subjective),
    `tier` (integer 1 or 2), `evidence` (string
    array), `recommendation`
    (enum: accept/author-decides),
    `decision` (enum: accept/modify/reject/ask),
    `decision_reasoning` (string or null),
    `commit_sha` (string or null),
    `divisor_agents_used` (string array),
    `tier2_unavailable` (boolean, default false),
    `conflict_flag` (boolean, default false)
  - `summary` (object, required): `total_items`,
    `accepted`, `modified`, `rejected`, `asked`,
    `tier1_count`, `tier2_count`,
    `divisor_agents_invoked` (string array)
  Use `additionalProperties: false` on all objects.
  Add enum constraints on `reviewer_role`,
  `classification`, `recommendation`, and `decision`.
  (FR-008)
- [x] 1.2 [P] Create `schemas/feedback-triage/samples/
  sample-feedback-triage.json`. Include a realistic
  sample with 6+ items covering all decision types and
  edge cases: one Tier 1 data-driven ACCEPT, one Tier 1
  subjective REJECT, one Tier 2 security ACCEPT with
  `divisor_agents_used`, one MODIFY with
  `decision_reasoning`, one ASK with
  `decision_reasoning`, one bot-authored item with null
  `file` and `line` (general PR comment). Wrap in the
  standard envelope schema with `hero: "cobalt-crush"`,
  populated `context.branch`, `context.commit`, and
  `context.backlog_item_id` fields. (FR-008)
- [x] 1.3 [P] Create negative test fixtures in
  `schemas/feedback-triage/samples/`: one sample with
  a missing required field (`pr_number`), one with an
  invalid `decision` value (e.g., `"defer"`), one with
  an extra property on an item object. These are used
  by schema validation tests to verify rejection of
  invalid inputs. (FR-008, Test Strategy)
- [x] 1.4 [P] Create `schemas/feedback-triage/README.md`
  following the format of
  `schemas/review-verdict/README.md`. Document producer
  (`/address-feedback` command, `hero: "cobalt-crush"`),
  consumers (Mx F, Muti-Mind, Cobalt-Crush), required
  fields, optional fields, version history, and schema
  evolution note: "Minor versions add optional fields
  and may relax `additionalProperties` constraints.
  Major versions may remove or rename fields." (FR-008)

## 2. Slash Command

- [x] 2.1 Create `.opencode/commands/address-feedback.md`.
  This is the canonical source file for the command.
  This is a UF-custom, tool-owned command (deployed by
  `uf init`). Structure the command in four phases
  following the design document (D1):

  **Phase 1 -- Ingest** (FR-001):
  - Require `gh auth status` check with actionable
    error on failure
  - Accept optional PR number argument; auto-detect
    via `gh pr view --json number` if absent
  - Error if no PR found for current branch
  - Fetch reviews: `gh api repos/{owner}/{repo}/
    pulls/{N}/reviews`
  - Fetch review comments: `gh api repos/{owner}/
    {repo}/pulls/{N}/comments`
  - Fetch issue comments: `gh api repos/{owner}/
    {repo}/issues/{N}/comments`
  - Handle paginated responses to fetch all data
  - On API failure (network, 5xx, 403): report error
    and stop; do not proceed with partial data
  - Determine reviewer authority from
    `author_association` field (D3, FR-004):
    OWNER/MEMBER → maintainer, COLLABORATOR →
    collaborator, CONTRIBUTOR/FIRST_TIMER/
    FIRST_TIME_CONTRIBUTOR → contributor,
    NONE without bot → external,
    NONE with bot → bot
  - Bot detection: login ending in `[bot]`, account
    type `Bot`
  - Filter: skip resolved threads, author's own
    comments, pure approvals with no inline comments
  - Group threaded conversations into discrete items
  - Detect GitHub suggestion blocks
  - Check local cache `.uf/feedback/pr-<N>/state.json`:
    if present and thread unchanged, reuse assessment;
    if thread has new comments, mark stale and
    re-assess; if missing, assess from scratch (FR-007)

  **Phase 2 -- Assess** (FR-002, FR-003, FR-004):
  - Load project context (D10, FR-010): convention
    packs from `.opencode/uf/packs/`, constitution,
    AGENTS.md, spec artifacts for PR branch, linked
    issues from PR description. Use `review-context`
    pack if available, inline discovery otherwise.
  - For each item, classify as DATA-DRIVEN or
    SUBJECTIVE with specific evidence references
  - Apply authority matrix (D3) to produce
    recommendation (ACCEPT or AUTHOR-DECIDES)
  - Detect conflicts between items referencing
    overlapping file/line ranges (FR-009)
  - Tier 1 items: assess directly using loaded context
  - Tier 2 items: delegate to relevant Divisor agent
    via Task tool. Route by domain: security →
    `divisor-adversary`, architecture →
    `divisor-architect`, testing → `divisor-testing`,
    performance → `divisor-sre`, scope/constitution →
    `divisor-guard`. Multiple domains → parallel
    agents. (FR-002)
  - If no Divisor agents available, fall back all items
    to Tier 1
  - Produce suggested approach for ACCEPT items
  - Cache assessment results to
    `.uf/feedback/pr-<N>/state.json` with restrictive
    file permissions (600) (FR-007)

  **Phase 3 -- Triage** (FR-005):
  - Present each item one-by-one with: reviewer
    identity and role, file/line reference, full
    thread content, classification, evidence,
    recommendation, suggested approach, conflict flag
  - Author chooses exactly one decision per item:
    ACCEPT, MODIFY (author provides approach), REJECT
    (author provides reasoning), ASK (author provides
    question)
  - No item may be skipped. All items must receive a
    decision.
  - After all items: display summary table of
    decisions and queued actions. Author confirms
    before execution.

  **Phase 4 -- Execute** (FR-006):
  - Implement all ACCEPT and MODIFY code changes
  - If a code change cannot be applied cleanly: skip
    that item with a report, continue with remaining
  - One commit per fix: conventional commit format
    `fix(<scope>): <description>` referencing PR
    number and reviewer (D6)
  - Run `/review-council` on cumulative changes (D7).
    If passes: continue. If fix loop exhausts: stop
    and report. Do not push until council passes.
  - Before pushing: fetch remote branch state and
    warn if diverged. Author confirms before
    proceeding.
  - Push all commits to PR branch
  - Post reply comments to PR (D9):
    - ACCEPT: "Addressed in `<sha>`" with description
    - REJECT: evidence-based reasoning
    - ASK: author's clarification question
    - Track posted vs. pending in cache for
      crash-recovery idempotency
  - If comment posting fails partway: report partial
    progress, record in cache for retry
  - All comment posting requires author confirmation
  - Offer to resolve threads for accepted items
  - Produce feedback-triage artifact at
    `.uf/artifacts/feedback-triage/
    pr-<N>-round-<M>.json` wrapped in envelope schema
    (D8, FR-008). Determine round from highest
    existing round number + 1. Write atomically
    (temp file then rename).

## 3. Scaffold Asset

- [x] 3.1 Copy `.opencode/commands/address-feedback.md`
  to `internal/scaffold/assets/opencode/commands/
  address-feedback.md` (embedded scaffold asset copy).
  The content MUST be byte-identical except for the
  version marker inserted by the scaffold engine.
- [x] 3.2 Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go` -- add
  `"opencode/commands/address-feedback.md"` to the
  slice in alphabetical order. Update the commands
  count comment to reflect the new total.

## 4. Documentation

- [x] 4.1 [P] Update `AGENTS.md` -- add
  `/address-feedback` to the PR Review Commands table:
  ```
  | `/address-feedback [N]` | Post-PR (GitHub) |
  Triage + address reviewer feedback |
  ```
  Add the `feedback-triage` schema to the Project
  Structure tree under `schemas/`.
- [x] 4.2 [P] Update `docs/usage.md` -- add a
  "Address Review Feedback" section after the existing
  "Review Code" section. Document the command's purpose,
  invocation (`/address-feedback [PR_NUMBER]`), the
  four-phase workflow, decision options (accept, modify,
  reject, ask), and the re-entrant multi-round model.
- [x] 4.3 [P] Update `docs/architecture.md` -- add
  `/address-feedback` to the Review Commands table or
  section. Reference the `feedback-triage` schema and
  its relationship to `review-verdict` (outbound
  review → inbound response). Document Tier 1/Tier 2
  assessment architecture and the authority matrix.
- [x] 4.4 [P] Update `CHANGELOG.md` -- add entry for
  the new `/address-feedback` command and
  `feedback-triage` schema under the appropriate
  version heading.
- [x] 4.5 [P] Add `.uf/feedback/` to `.gitignore` if
  not already covered by an existing `.uf/` pattern.
  The cache directory MUST NOT be committed. (FR-007)

## 5. Verification

- [x] 5.1 Run `make check` -- verify build, test, vet,
  and lint all pass with the new asset files and schema.
  (FR-001 through FR-010)
- [x] 5.2 Verify `TestAssetPaths_MatchExpected` passes
  (confirms `expectedAssetPaths` matches actual
  embedded files including the new command). (FR-001)
- [x] 5.3 Verify `TestEmbeddedAssets_MatchSource` passes
  (confirms live copy and scaffold asset are
  byte-identical after version marker insertion).
- [x] 5.4 Validate `schemas/feedback-triage/
  v1.0.0.schema.json` is well-formed JSON Schema.
  Validate `schemas/feedback-triage/samples/
  sample-feedback-triage.json` against the schema
  (positive test). Validate each negative test fixture
  is rejected by the schema (negative tests). Use Go
  test infrastructure in `internal/schemas/` or
  `ajv-cli`. (FR-008, Test Strategy)
- [x] 5.5 Verify constitution alignment:
  - Principle I (Autonomous Collaboration): artifact
    is self-describing with full provenance metadata
    (FR-008)
  - Principle II (Composability First): command works
    without Divisor agents (Tier 1 fallback, FR-002),
    works without review-context pack (inline
    discovery, FR-010)
  - Principle III (Observable Quality): artifact is
    machine-parseable JSON conforming to registered
    schema with provenance (FR-008)
  - Principle IV (Testability): schema validates
    against positive and negative samples, scaffold
    tests cover new asset (Test Strategy)
- [x] 5.6 Assess Website Documentation Gate -- determine
  whether `/address-feedback` requires a GitHub issue
  in `unbound-force/website` for user-facing
  documentation updates. This is a new user-facing
  command, so an issue MUST be filed via
  `gh issue create --repo unbound-force/website`.
  Cross-reference website issue #94 (Quality Gates
  page) in the issue body. Note blog opportunity:
  "Closing the Review Loop" and tutorial opportunity:
  "Addressing PR Feedback Step-by-Step."
<!-- scaffolded by uf vdev -->

<!-- spec-review: passed -->
<!-- code-review: passed -->
