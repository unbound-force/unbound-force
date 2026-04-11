# Feature Specification: Documentation Curation

**Feature Branch**: `026-documentation-curation`  
**Created**: 2026-04-11  
**Status**: Draft  
**Input**: User description: "Add Divisor Curator agent and Guard documentation completeness checks to ensure documentation stays current, blog opportunities are captured, and tutorial needs are identified — with GH issues filed in the website repo for all content gaps."

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Documentation Gap Detection During Code Review (Priority: P1)

When the review council reviews a code change that
modifies user-facing behavior (CLI commands, agent
capabilities, installation steps, workflows), the
Curator agent detects whether the corresponding
documentation was updated. If AGENTS.md, README.md, or
other in-repo documentation was not updated to reflect
the change, the Curator flags it as a blocking finding.
If a GitHub issue was not filed in the
`unbound-force/website` repo for the website
documentation update, the Curator files the issue and
flags the gap as blocking.

Today, documentation updates are instructions in
AGENTS.md that Cobalt-Crush should follow during
implementation. No reviewer verifies compliance. Changes
regularly ship without documentation updates, creating
drift between the documented state and the actual state.

**Why this priority**: Documentation drift is the most
common quality gap in the project. Every spec in this
session (022-025) required manual AGENTS.md updates that
could have been missed. This is the highest-value fix.

**Independent Test**: Submit a PR that adds a new CLI
command but does not update AGENTS.md or file a website
issue. Verify the Curator flags both gaps as blocking
findings.

**Acceptance Scenarios**:

1. **Given** a PR that adds a new CLI flag, **When** the
   review council runs, **Then** the Curator checks
   whether AGENTS.md was updated and whether a website
   issue was filed. If either is missing, the Curator
   flags a blocking finding.
2. **Given** a PR that modifies an existing agent's
   behavior, **When** the review council runs, **Then**
   the Curator checks whether the agent listing in
   AGENTS.md reflects the change and whether a website
   issue exists for the change. Missing items are
   flagged as blocking.
3. **Given** a PR that is purely internal (refactoring,
   test-only, CI-only), **When** the review council
   runs, **Then** the Curator produces no documentation
   findings (no false positives for internal changes).
4. **Given** a PR where the developer already filed a
   website issue and updated AGENTS.md, **When** the
   review council runs, **Then** the Curator confirms
   compliance and produces no findings.

---

### User Story 2 — Blog Opportunity Identification (Priority: P1)

When a code change introduces a significant new
capability (new hero, new agent, major workflow change,
architectural migration), the Curator identifies it as
a blog opportunity. The Curator files a GitHub issue in
the `unbound-force/website` repo with the `blog` label,
including a suggested topic, angle, key points, and
references to the relevant PR and spec. If no blog issue
is filed for a significant change, the Curator flags it
as a blocking finding.

Today, no agent identifies blog-worthy changes. The
Herald writes blog posts when asked, but nobody
proactively identifies what should be written about.

**Why this priority**: Blog content drives awareness and
adoption. Every major change in this session (Replicator
migration, content agents, .uf/ convention) was
blog-worthy but would not have been identified without
manual judgment.

**Independent Test**: Submit a PR that adds a new Divisor
agent. Verify the Curator files a blog issue in the
website repo with a suggested topic and angle.

**Acceptance Scenarios**:

1. **Given** a PR that adds a new agent persona, **When**
   the review council runs, **Then** the Curator files a
   blog issue in `unbound-force/website` with label
   `blog`, suggested topic, and PR reference.
2. **Given** a PR that is a minor bug fix, **When** the
   review council runs, **Then** the Curator does not
   file a blog issue (no noise for routine changes).
3. **Given** a PR where a blog issue already exists for
   the change, **When** the review council runs, **Then**
   the Curator confirms the existing issue and does not
   create a duplicate.

---

### User Story 3 — Tutorial Opportunity Identification (Priority: P2)

When a code change introduces a new workflow that
engineers need to learn (new slash command, new
multi-step process, new tool integration), the Curator
identifies it as a tutorial opportunity. The Curator
files a GitHub issue in the `unbound-force/website` repo
with the `tutorial` label, including a suggested
structure and target audience.

**Why this priority**: Tutorials are valuable but less
urgent than documentation accuracy and blog coverage.
Engineers can learn workflows from specs and AGENTS.md
in the short term; tutorials provide a better learning
experience long-term.

**Independent Test**: Submit a PR that introduces a new
slash command with a multi-step workflow. Verify the
Curator files a tutorial issue in the website repo.

**Acceptance Scenarios**:

1. **Given** a PR that adds a new slash command (e.g.,
   `/uf sandbox`), **When** the review council runs,
   **Then** the Curator files a tutorial issue in
   `unbound-force/website` with label `tutorial`,
   suggested structure, and target audience.
2. **Given** a PR that modifies internal logic without
   changing the user-facing workflow, **When** the
   review council runs, **Then** the Curator does not
   file a tutorial issue.
3. **Given** a tutorial issue already exists for the
   workflow, **When** the review council runs, **Then**
   the Curator confirms the existing issue and does not
   create a duplicate.

---

### User Story 4 — Guard Documentation Completeness Check (Priority: P2)

The Guard agent's code review audit includes a new
"Documentation Completeness" checklist item that verifies
in-repo documentation was updated when user-facing
behavior changed. This supplements the Curator's
cross-repo issue filing with a focused check on the
current repo's AGENTS.md and README.md.

**Why this priority**: The Guard already reviews every
change for intent drift. Adding documentation
completeness to its existing checklist is a natural
extension with minimal implementation cost.

**Independent Test**: Submit a PR that changes a CLI
command's behavior without updating AGENTS.md. Verify
the Guard flags the missing AGENTS.md update as a
MEDIUM finding.

**Acceptance Scenarios**:

1. **Given** a PR that changes `uf setup` behavior
   without updating AGENTS.md, **When** the Guard
   reviews the change, **Then** the Guard flags the
   missing update as a MEDIUM finding.
2. **Given** a PR that updates both the code and
   AGENTS.md, **When** the Guard reviews, **Then** the
   Guard's documentation completeness check passes.
3. **Given** a test-only PR, **When** the Guard reviews,
   **Then** the documentation completeness check is
   skipped (no user-facing changes).

---

### Edge Cases

- What happens when the Curator cannot access the
  `unbound-force/website` repo (no `gh` token, network
  error)? The Curator SHOULD report the failure as a
  finding with the issue text it would have filed, so
  the developer can file it manually.
- What happens when the Curator files an issue that
  duplicates an existing one? The Curator MUST search
  open issues in the website repo before creating a
  new one. If a matching issue exists, reference it
  instead of creating a duplicate.
- What happens when a PR has both documentation gaps
  AND blog/tutorial opportunities? The Curator MUST
  report all findings — documentation gaps as blocking,
  blog/tutorial opportunities as blocking (issue must
  be filed).
- What happens when the Curator's bash access is used
  for something other than `gh issue create`/`gh issue
  list`? The Curator's agent file MUST include a bash
  access restriction: only `gh issue create` and
  `gh issue list` operations against the
  `unbound-force/website` repository are permitted.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The review council MUST include a Curator
  agent (`divisor-curator.md`) that runs during every
  code review.
- **FR-002**: The Curator MUST detect whether in-repo
  documentation (AGENTS.md, README.md) was updated when
  a PR modifies user-facing behavior.
- **FR-003**: The Curator MUST detect whether a GitHub
  issue was filed in `unbound-force/website` for changes
  requiring website documentation updates.
- **FR-004**: If a documentation gap exists in the
  current repo, the Curator MUST flag it as a MEDIUM
  blocking finding.
- **FR-005**: If a website documentation issue is
  missing, the Curator MUST file the issue via
  `gh issue create` with label `docs` and flag the gap
  as HIGH blocking finding.
- **FR-006**: The Curator MUST identify blog-worthy
  changes (new capabilities, migrations, architectural
  decisions) and file GitHub issues in
  `unbound-force/website` with label `blog`.
- **FR-007**: The Curator MUST identify tutorial-worthy
  changes (new workflows, multi-step processes) and file
  GitHub issues in `unbound-force/website` with label
  `tutorial`.
- **FR-008**: Missing blog or tutorial issues for
  significant changes MUST be flagged as MEDIUM blocking
  findings.
- **FR-009**: The Curator MUST search existing open
  issues in `unbound-force/website` before filing new
  ones to avoid duplicates.
- **FR-010**: The Curator MUST have `bash: true` in its
  frontmatter, with a documented restriction that bash
  is only for `gh issue create` and `gh issue list`.
- **FR-011**: The Guard agent MUST include a
  "Documentation Completeness" checklist item in its
  Code Review audit section.
- **FR-012**: The Guard's documentation completeness
  check MUST flag missing AGENTS.md updates as MEDIUM
  findings when user-facing behavior changed.
- **FR-013**: Internal-only changes (refactoring,
  test-only, CI-only) MUST NOT trigger documentation
  or content findings from either the Curator or Guard.
- **FR-014**: The Curator MUST integrate with Dewey
  (Step 0: Prior Learnings) following the established
  pattern for Divisor agents.
- **FR-015**: All modified agent files MUST have their
  scaffold asset copies synchronized.
- **FR-016**: All existing tests MUST continue to pass
  after the changes.

### Key Entities

- **Documentation Gap**: A mismatch between a code
  change and its corresponding documentation. Detected
  by comparing the PR diff against documentation files.
- **Content Opportunity**: A blog or tutorial topic
  identified from a significant code change. Tracked
  as a GitHub issue in the website repo.
- **Website Issue**: A GitHub issue filed in
  `unbound-force/website` with labels `docs`, `blog`,
  or `tutorial`. Includes what changed, why it matters,
  and suggested content angle.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of PRs that modify user-facing
  behavior trigger the Curator's documentation gap
  check — verified by running the review council on
  a PR that adds a new CLI command.
- **SC-002**: Zero PRs with user-facing changes merge
  without either (a) updated in-repo docs or (b) a
  blocking Curator finding — verified by reviewing the
  Curator's output for changes without doc updates.
- **SC-003**: 100% of significant feature changes
  (new agents, new commands, migrations) have a
  corresponding blog issue filed in the website repo
  — verified by checking the website repo's issue
  tracker after a major PR merges.
- **SC-004**: Zero duplicate issues filed by the Curator
  — verified by checking that the Curator searches
  existing issues before creating new ones.
- **SC-005**: All existing tests pass after the changes
  — verified by running the full test suite.
- **SC-006**: The Curator's bash access is restricted
  to `gh` commands only — verified by reading the
  agent file's documented restriction.

## Dependencies & Assumptions

### Dependencies

- **GitHub CLI (`gh`)**: Must be available in the agent's
  environment for `gh issue create` and `gh issue list`.
  Already installed by `uf setup`.
- **Website repo labels**: The `unbound-force/website`
  repo must have `docs`, `blog`, and `tutorial` labels.
  These may need to be created.

### Assumptions

- The Curator can determine whether a change is
  "user-facing" by inspecting the PR diff for changes
  to files under `cmd/`, `.opencode/agents/`,
  `.opencode/command/`, `internal/scaffold/`, or
  `AGENTS.md`. Internal-only changes (under
  `internal/` excluding scaffold, or test files) are
  not user-facing.
- The review council's automatic discovery of
  `divisor-*.md` files means no changes are needed to
  the `/review-council` command — adding
  `divisor-curator.md` automatically includes the
  Curator in every review.
- The Curator's findings are severity-graded like other
  Divisor agents (MEDIUM/HIGH = blocking). The review
  council's existing fix/re-run loop handles Curator
  findings the same way it handles other agents'
  findings.
- The website repo `unbound-force/website` is
  accessible via `gh issue` commands from the agent's
  environment (same GitHub token).
