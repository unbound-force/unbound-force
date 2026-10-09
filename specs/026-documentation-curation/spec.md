# Feature Specification: Documentation Curation

**Feature Branch**: `026-documentation-curation`
**Created**: 2026-04-11
**Status**: Draft
**Input**: User description: "Add Divisor Curator agent and Guard
documentation completeness checks to ensure documentation stays
current, blog opportunities are captured, and tutorial needs are
identified, with GitHub issues filed in the website repository for
content gaps."

The original intent remains: documentation impact and content
opportunities must not be lost. The current execution contract makes
Curator a review-capable, scope-selected persona. Child Curator runs
assess and recommend; they never create issues. The parent performs
at most one deduplicated curation action after consolidation and an
explicit human gate.

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Documentation Gap Detection (Priority: P1)

When a code change modifies user-facing behavior, the deterministic
scope classification makes Curator eligible. Curator checks whether
in-repository documentation reflects the change and whether required
website documentation work is already tracked. Missing coverage is a
blocking finding.

Curator MAY be pruned when a change has neither documentation nor
user-facing scope. A child Curator run never files an issue. After
run consolidation, the parent MAY perform one deduplicated curation
action only after the existing human gate approves it.

**Why this priority**: Documentation drift is a frequent quality gap.
The review must preserve documentation impact without granting child
review sessions mutation authority.

**Independent Test**: Submit a change that adds a CLI flag without
updating relevant documentation or tracking website work. Verify
that Curator is included, reports both gaps, and proposes a parent
action without creating an issue itself.

**Acceptance Scenarios**:

1. **Given** a change adds a CLI flag, **When** deterministic scope
   classification marks it user-facing, **Then** Curator checks the
   in-repository update and website issue evidence.
2. **Given** required website work is missing, **When** Curator
   finishes, **Then** it returns a blocking finding and a proposed
   curation action without creating an issue.
3. **Given** Curator proposes an action, **When** the parent
   consolidates all runs, **Then** the parent checks for duplicates
   and asks for human approval before performing at most one action.
4. **Given** a purely internal change, **When** planning classifies
   its scope, **Then** Curator MAY be pruned and no documentation or
   content finding is fabricated.
5. **Given** documentation and tracking are complete, **When**
   Curator reviews the change, **Then** it reports compliance and no
   curation action is proposed.

---

### User Story 2 — Blog Opportunity Identification (Priority: P1)

When a significant user-facing change introduces a capability,
migration, or major workflow, Curator identifies a potential blog
opportunity. Curator provides a suggested topic, angle, key points,
and references. It does not create the website issue directly.

After consolidation, the parent MAY create one deduplicated issue
with the `blog` label if a human approves the action. A missing issue
for an applicable change remains a blocking content finding.

**Why this priority**: Blog content supports awareness and adoption.
The opportunity must be captured without duplicate or unauthorized
child-session mutations.

**Independent Test**: Review a significant new capability. Verify
that Curator proposes a blog issue and the parent creates at most one
issue only after duplicate checking and human approval.

**Acceptance Scenarios**:

1. **Given** a change adds a significant capability, **When**
   Curator reviews it, **Then** Curator proposes a blog topic, angle,
   key points, and source references.
2. **Given** a minor bug fix, **When** Curator reviews it, **Then**
   Curator does not propose a blog action.
3. **Given** a matching blog issue exists, **When** the parent
   checks the proposed action, **Then** it references the existing
   issue and does not create a duplicate.
4. **Given** no matching issue exists, **When** the human approves
   the consolidated action, **Then** the parent MAY create one issue
   labeled `blog`.

---

### User Story 3 — Tutorial Opportunity Identification (Priority: P2)

When a user-facing change introduces a workflow that engineers must
learn, Curator identifies a tutorial opportunity. It proposes a
target audience, structure, and source references. The parent owns
any issue creation after consolidation, deduplication, and approval.

**Why this priority**: Tutorials improve adoption but are less urgent
than documentation accuracy and required website coverage.

**Independent Test**: Review a new multi-step slash command. Verify
that Curator proposes tutorial work and cannot create an issue from
the child review session.

**Acceptance Scenarios**:

1. **Given** a change adds a multi-step command, **When** Curator
   reviews it, **Then** Curator proposes a tutorial structure and
   target audience.
2. **Given** an internal implementation change, **When** planning
   classifies it, **Then** Curator MAY be pruned and no tutorial
   action is proposed.
3. **Given** a matching tutorial issue exists, **When** the parent
   checks the proposal, **Then** no duplicate is created.
4. **Given** the proposal is unique, **When** a human approves it,
   **Then** the parent MAY create one issue labeled `tutorial`.

---

### User Story 4 — Guard Documentation Check (Priority: P2)

Guard's code review audit includes a documentation completeness
check for user-facing behavior. This provides an always-eligible
backstop for in-repository documentation while Curator remains a
scope-selected specialist.

**Why this priority**: Guard already checks intent and cohesion. The
documentation check preserves coverage when Curator is legitimately
pruned or unavailable.

**Independent Test**: Submit a user-facing command change without an
`AGENTS.md` update and verify that Guard reports a MEDIUM finding.

**Acceptance Scenarios**:

1. **Given** a change modifies `uf setup` behavior without updating
   `AGENTS.md`, **When** Guard reviews it, **Then** Guard reports a
   MEDIUM documentation completeness finding.
2. **Given** code and relevant documentation are both updated,
   **When** Guard reviews, **Then** the documentation check passes.
3. **Given** a test-only change, **When** Guard reviews, **Then** the
   documentation check is skipped.

---

### Edge Cases

- If the parent cannot access `unbound-force/website`, it SHOULD
  report the failure and the complete proposed issue for manual use.
- Before any curation action, the parent MUST search existing open
  issues and reference a match instead of creating a duplicate.
- Curator MUST report documentation, blog, and tutorial findings
  together when all apply. Consolidation MUST still permit at most
  one parent action for the review operation.
- Child Curator runs MUST NOT invoke issue creation, issue mutation,
  or any other curation action.
- Multiple model runs for Curator are prohibited by the dispatch
  contract. Curator MUST have at most one assessment run.
- Parent curation failure MUST NOT be hidden or converted into a
  fabricated successful action.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Curator MUST be review-capable with documentation
  scope in the closed reviewer-capabilities manifest. It MUST be
  eligible for documentation or user-facing classifications and MAY
  be pruned outside that scope. It MUST NOT be required for every
  code review.
- **FR-002**: Curator MUST detect whether in-repository
  documentation, including `AGENTS.md` and `README.md`, was updated
  when user-facing behavior changed.
- **FR-003**: Curator MUST detect whether required website
  documentation work already has a corresponding issue.
- **FR-004**: An in-repository documentation gap MUST produce a
  MEDIUM blocking finding.
- **FR-005**: A missing required website documentation issue MUST
  produce a HIGH blocking finding and a proposed `docs` action. Only
  the parent MAY perform that action after consolidation, duplicate
  checking, and an explicit human gate.
- **FR-006**: Curator MUST identify blog-worthy changes and propose
  a website issue labeled `blog`. Only the parent MAY create it under
  the FR-005 action controls.
- **FR-007**: Curator MUST identify tutorial-worthy changes and
  propose a website issue labeled `tutorial`. Only the parent MAY
  create it under the FR-005 action controls.
- **FR-008**: Missing blog or tutorial tracking for applicable
  significant changes MUST produce a MEDIUM blocking finding.
- **FR-009**: The parent MUST search existing open website issues
  before any action. It MUST reference a matching issue and MUST NOT
  create a duplicate.
- **FR-010**: A child Curator run MUST NOT create or mutate issues.
  Parent commands MUST own curation actions and MUST perform at most
  one deduplicated action after consolidation and human approval.
  This supersedes the earlier child-bash restriction with a stricter
  parent-only mutation boundary.
- **FR-011**: Guard MUST include a Documentation Completeness item
  in its code review audit.
- **FR-012**: Guard MUST report missing `AGENTS.md` updates as MEDIUM
  findings when user-facing behavior changed.
- **FR-013**: Internal-only refactoring, test-only changes, and
  CI-only changes MUST NOT produce documentation or content findings
  from Curator or Guard.
- **FR-014**: Curator MUST perform Step 0 prior-learning retrieval
  through Dewey when that service is available.
- **FR-015**: Modified agent files MUST keep scaffold asset copies
  synchronized.
- **FR-016**: Existing tests MUST continue to pass.
- **FR-017**: Curator MUST have at most one included assessment run.
  Dynamic plans MAY contain multiple runs for other included review
  personas, but MUST reject a second Curator run.

### Key Entities

- **Documentation Gap**: A mismatch between user-facing behavior and
  the documentation or tracking that should explain it.
- **Content Opportunity**: A blog or tutorial topic identified from
  a significant user-facing change.
- **Curation Proposal**: A child-produced, non-mutating suggestion
  with issue type, rationale, references, and proposed content.
- **Parent Curation Action**: At most one deduplicated action that a
  parent command MAY perform after consolidation and human approval.
- **Website Issue**: A parent-created issue in
  `unbound-force/website` with `docs`, `blog`, or `tutorial` labels.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every change deterministically classified as
  documentation or user-facing makes Curator eligible. Changes
  outside those scopes MAY prune Curator with a recorded reason.
- **SC-002**: No user-facing change merges without updated relevant
  documentation or a blocking finding.
- **SC-003**: Every applicable significant feature has tracked blog
  work or a blocking finding and parent-owned proposed action.
- **SC-004**: Parent duplicate checks prevent duplicate curation
  issues.
- **SC-005**: All existing tests pass after the changes.
- **SC-006**: Child Curator runs create no issues. Every resulting
  curation action is parent-only, deduplicated, and human-gated.
- **SC-007**: Internal-only changes generate no Curator or Guard
  documentation false positives.

## Dependencies & Assumptions

### Dependencies

- **Reviewer capabilities manifest**: The closed manifest defines
  Curator as a documentation-scoped review persona.
- **Review dispatch planner**: The planner applies deterministic
  scope intersection and the one-run Curator bound.
- **Parent command GitHub access**: The parent needs authenticated
  GitHub access only when a human approves a curation action.
- **Website labels**: `unbound-force/website` SHOULD provide `docs`,
  `blog`, and `tutorial` labels.

### Assumptions

- Deterministic classification identifies documentation and
  user-facing scope before reviewer planning.
- Adversary and Guard remain always eligible. Curator and every
  other non-always review persona require scope intersection.
- Content-only Envoy, Herald, and Scribe never dispatch as reviewers.
- Curator findings use the same severity and consolidation rules as
  other review-capable personas.
- Parent curation actions do not alter the canonical review decision
  or make `review-dispatch` replace `review-verdict`.
