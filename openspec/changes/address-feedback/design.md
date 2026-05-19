## Context

The PR lifecycle has a gap between PR creation
(`/finale`) and merge. External reviewers provide
feedback that today requires manual, unstructured
effort to address. The `/address-feedback` command
closes this gap with structured ingestion, tiered
assessment, author-driven triage, and batched
execution.

The command sits in the lifecycle as:
```
/unleash → /review-council → /finale → PR
  → reviewers comment → /address-feedback (N rounds)
  → merge
```

## Goals / Non-Goals

### Goals
- Fetch and structure all unresolved PR review
  feedback from GitHub in a single invocation
- Classify each item as data-driven or subjective
  with evidence from project standards
- Present items one-by-one to the author with clear
  recommendations and decision options
- Execute all decisions as a batch: code commits
  (one per fix) + PR reply comments
- Run `/review-council` on code changes before pushing
- Produce a feedback-triage artifact for consumption
  by Mx F, Muti-Mind, and Cobalt-Crush
- Support re-entrant invocation across multiple
  review rounds with idempotent behavior on resolved
  threads
- Cache assessments locally per PR number for speed,
  reconstruct from GitHub when cache is absent

### Non-Goals
- Auto-merging PRs after feedback is addressed
- Replacing `/review-pr` (which is a one-shot review
  tool, not a feedback response tool)
- Implementing a full conversation bot on the PR
- Handling cross-PR or cross-repo review coordination
- Modifying the `/review-council` command itself
- Building a custom GitHub App or webhook integration

## Decisions

### D1: Four-Phase Architecture

The command follows four sequential phases:

```
Phase 1: INGEST     → Fetch + filter + group
Phase 2: ASSESS     → Classify + evidence + recommend
Phase 3: TRIAGE     → Author decides per item
Phase 4: EXECUTE    → Code + review-council + push +
                      comments + artifact
```

Phases are not independently invocable. The command
runs all four in sequence. This keeps the UX simple:
one command, one session, one outcome.

**Rationale**: Separating phases into distinct commands
would add cognitive overhead for the author and
introduce state management complexity between commands.
The local cache handles re-entrancy across invocations.

### D2: Tiered Assessment Engine

Two tiers to balance speed and thoroughness:

**Tier 1 (direct, single-agent)**: The command agent
itself assesses the item using loaded convention packs
and constitution. Handles items where classification
is unambiguous: formatting, naming, typos, clear
convention match, single-file scope, GitHub suggestion
blocks with exact replacement code.

**Tier 2 (Divisor escalation)**: The command delegates
assessment to the relevant Divisor agent via Task tool.
Triggers: security concern, architectural change,
multi-file impact, ambiguous classification, feedback
contradicts a convention pack rule, test strategy or
coverage concern, performance or operational concern.

Routing table for Tier 2:
- Security, input validation, credentials →
  `divisor-adversary`
- Architecture, patterns, DRY →
  `divisor-architect`
- Test coverage, assertions, isolation →
  `divisor-testing`
- Performance, deployment, ops →
  `divisor-sre`
- Scope drift, constitution compliance →
  `divisor-guard`
- Multiple signals → multiple agents (parallel)

**Rationale**: Tier 1 keeps simple items fast (~2-3s).
Tier 2 uses existing specialized expertise for complex
items (~10-20s) without building a monolithic
assessment engine. This aligns with Composability First
(Divisor agents are additive -- if absent, all items
fall back to Tier 1).

### D3: Authority Matrix

Reviewer role detection via GitHub API
(`author_association` field from PR review objects):

| | Data-Driven | Subjective |
|---|---|---|
| Maintainer / Owner | ACCEPT (MUST fix) | AUTHOR-DECIDES (SHOULD consider) |
| Collaborator / Member | ACCEPT (MUST fix) | AUTHOR-DECIDES |
| Contributor | ACCEPT (MUST fix) | AUTHOR-DECIDES |
| External | ACCEPT if validated | AUTHOR-DECIDES |
| Bot / AI-generated | ACCEPT if validated | AUTHOR-DECIDES (informational) |

The `author_association` field maps to:
- `OWNER`, `MEMBER` → maintainer authority
- `COLLABORATOR` → collaborator authority
- `CONTRIBUTOR`, `FIRST_TIMER`,
  `FIRST_TIME_CONTRIBUTOR` → contributor authority
- `NONE` without bot indicators → external authority
- `NONE` with bot indicators → bot authority

Bot detection MUST check: login ending in `[bot]` or
account type `Bot`.

Bot and external review validation: when a bot or
external reviewer flags a data-driven issue, the
assessment engine cross-references the finding against
project convention packs. If the pack confirms the
rule, the recommendation is ACCEPT. If no matching
rule exists, the recommendation is AUTHOR-DECIDES
with a note that the rule is not backed by project
standards.

**Rationale**: Maintainer feedback carries more weight
because they have already applied judgment. Bot and
external feedback needs validation because automated
tools can produce false positives and external
reviewers may not know the project's conventions.

### D4: GitHub as Source of Truth

The command always fetches the full PR state from
GitHub. Local cache under `.uf/feedback/pr-<N>/` is an
accelerator, not the authority.

Cache structure:
```
.uf/feedback/pr-<N>/
├── state.json       # per-thread assessments
└── last-fetched     # ISO 8601 timestamp
```

The `state.json` internal structure is
implementation-defined since the cache is disposable
and not a cross-hero artifact. At minimum it MUST
contain: a map of thread IDs to assessment objects
(classification, tier, evidence, recommendation), the
comment count and last comment ID at assessment time,
the author's decision (once made), and execution
status (committed, pushed, comment-posted) for
crash-recovery idempotency. Cache files MUST be
created with restrictive permissions (600) since
they may contain security-sensitive review content.

Cache keying: PR number (not branch name, not commit
SHA). This handles branch renames, rebases, and force
pushes correctly.

Cache invalidation per thread:
- Thread has new comments since last assessment →
  re-assess
- Code at referenced lines changed (detected via diff
  comparison) → mark stale, re-assess
- Thread resolved on GitHub → skip entirely
- Cache file missing → full reconstruction (correct
  but slower)

**Rationale**: The cache improves round 2+ performance
by avoiding re-assessment of unchanged threads. GitHub
as source of truth means the command works correctly
even if `.uf/feedback/` is deleted, moved to a
different machine, or the author switches branches
between rounds. This also supports the multi-PR
scenario where an author has several PRs under review
simultaneously -- each PR's cache is independent.

### D5: Four Decision Options

Each feedback item gets exactly one decision:
- **ACCEPT**: Queue code change as suggested (or as
  agent recommends)
- **MODIFY**: Queue code change with author's
  alternative approach
- **REJECT**: Queue reply comment with evidence-based
  reasoning
- **ASK**: Queue reply comment with author's
  clarification question

No item may be deferred or skipped. This ensures
complete coverage -- every piece of feedback gets a
response.

ASK items create a natural re-entrancy point: the
reviewer responds, the author runs `/address-feedback`
again, and the thread now has more context for a
definitive decision.

**Rationale**: The original "discuss" option was
replaced with ASK because it produces a concrete
action (a posted comment) rather than a deferred
non-decision. This keeps the PR conversation moving
and eliminates parking-lot items.

### D6: One Commit Per Fix

Each accepted/modified code change gets its own commit
with a conventional commit message:
```
fix(gateway): extract timeout to config constant

Addresses PR #42 review feedback from @alice.
```

**Rationale**: Granular commits let reviewers verify
each fix independently. A single "address all feedback"
commit makes it harder for reviewers to confirm their
specific concern was addressed. The commit message
references the PR number and reviewer for traceability.

### D7: Review-Council Gate Before Push

After all code changes are committed locally, the
command runs `/review-council` on the changes. This
prevents introducing new issues while fixing feedback.

If review-council finds issues:
- Fix loop runs (same behavior as `/unleash`)
- If fixes succeed, continue to push
- If fixes exhaust iterations, stop and report

**Rationale**: Pushing code that passes the author's
triage but fails the project's own review standards
would create another round of feedback. Running the
council locally first catches this before it reaches
reviewers.

### D8: Feedback-Triage Artifact

The command produces a JSON artifact at:
```
.uf/artifacts/feedback-triage/pr-<N>-round-<M>.json
```

The artifact is wrapped in the standard envelope
schema (hero, version, timestamp, artifact_type,
schema_version, context, payload). The payload follows
the `feedback-triage` schema.

Envelope field mapping:
- `hero`: `"cobalt-crush"` (the agent executing the
  command)
- `context.branch`: PR's source branch
- `context.commit`: HEAD after fixes are pushed (or
  pre-fix HEAD if no code changes)
- `context.backlog_item_id`: linked issue or `PR-<N>`

The payload `branch` duplicates `context.branch`
intentionally to support payload-only consumers.

Artifact writes MUST be atomic (write to temp file,
then rename) to prevent truncated JSON. Round number
is derived from highest existing round + 1, not file
count, to handle gaps.

Artifact payload includes:
- PR metadata (number, URL, branch)
- Round number (increments per invocation)
- Per-item records: thread ID, reviewer, role,
  classification, tier, evidence references,
  recommendation, decision, reasoning, commit SHA
  (if code change)
- Summary statistics: totals by decision type, tier
  distribution, agents invoked

**Rationale**: Constitution Principle III (Observable
Quality) requires machine-parseable output with
provenance. The artifact enables Mx F to track review
patterns, Muti-Mind to detect recurring gaps, and
Cobalt-Crush to learn from past triage decisions.

### D9: PR Comment Posting Protocol

After code changes are pushed, the command posts reply
comments to the PR:
- Accepted items: "Addressed in `<commit_sha>`" with
  a brief description of the change
- Rejected items: evidence-based reasoning referencing
  specific convention pack rules or constitution
  principles
- Asked items: the author's clarification question

All comments require author confirmation before
posting (same safety model as `/review-pr`).

Thread resolution: the command offers to resolve
threads for accepted items. Author confirms.

**Rationale**: Posting responses to each thread keeps
the PR conversation clean and shows reviewers that
their feedback was systematically addressed. Evidence-
based rejection comments are more professional and
productive than "I disagree" responses.

### D10: Context Discovery

The command needs project context for assessment:
- Convention packs (`.opencode/uf/packs/`)
- Constitution (`.specify/memory/constitution.md`)
- AGENTS.md coding/testing conventions
- Spec artifacts (if the PR has an associated spec)
- Linked issues (from PR description)

If the `review-context` convention pack (#176) exists,
the command uses it for standardized discovery. If not,
the command inlines the discovery logic (same approach
as `/review-pr` Step 6-7 today).

**Rationale**: Composability First -- the command works
standalone. The review-context pack is additive.

## Risks / Trade-offs

### R1: Token Cost for Tier 2 Escalation

Each Tier 2 item invokes a Divisor agent, consuming
additional tokens. A PR with 10 complex feedback items
could trigger 10 agent invocations.

**Mitigation**: Tier 1 handles the majority of common
feedback (formatting, naming, simple convention
matches). Tier 2 only triggers on genuinely complex
items. The tiered approach is itself the mitigation.

### R2: GitHub API Rate Limits

Fetching reviews, comments, and review threads makes
multiple API calls. Large PRs with extensive discussion
could approach rate limits.

**Mitigation**: The local cache reduces API calls on
subsequent rounds. The command fetches in bulk where
possible (list endpoints rather than per-thread
fetches).

### R3: Review-Council Gate Adds Time

Running `/review-council` on changes before pushing
adds 1-5 minutes to the execution phase.

**Mitigation**: This is the same tradeoff `/unleash`
makes. The alternative -- pushing untested fixes and
getting another round of feedback -- costs more time
overall.

### R4: Stale Review Threads

GitHub marks review comments as "outdated" when the
underlying code changes. The command may present items
whose line references no longer match the current code.

**Mitigation**: The command marks stale items with a
STALE indicator and presents them with a note that line
numbers may have shifted. The underlying concern may
still be valid even if the exact code location changed.

### R5: Conflicting Reviewer Feedback

Two reviewers may give contradictory suggestions on
the same code section.

**Mitigation**: The assessment engine detects conflicts
by comparing feedback items that reference overlapping
file/line ranges. Conflicts are flagged explicitly to
the author with both perspectives presented. The author
must choose one approach and the rejected reviewer gets
a comment explaining the decision.
<!-- scaffolded by uf vdev -->
