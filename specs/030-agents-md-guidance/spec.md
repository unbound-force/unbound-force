# Feature Specification: AGENTS.md Behavioral Guidance Injection

**Feature Branch**: `030-agents-md-guidance`  
**Created**: 2026-04-15  
**Status**: Draft  
**Issue**: #104  
**Input**: User description: "issue #104 — Add AGENTS.md behavioral guidance injection to /uf-init"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Standardize Critical Behavioral Rules Across Repos (Priority: P1)

When an engineer runs `/uf-init` in any Unbound Force
repository, the command ensures 8 standardized behavioral
guidance sections are present in the repo's `AGENTS.md`.
Today, these rules exist inconsistently — some repos
have gatekeeping protection, others have CI parity gates,
most have neither. After this change, every repo that
runs `/uf-init` gets the same behavioral foundation,
ensuring agents follow the same quality gates, workflow
boundaries, and documentation requirements regardless
of which repo they're working in.

**Why this priority**: Without consistent behavioral
rules, agents in one repo may skip CI checks, write
code during planning phases, or merge without review
— behaviors that are prevented in other repos. The
inconsistency creates a quality gradient across the
ecosystem.

**Independent Test**: Run `/uf-init` in a repo with a
minimal `AGENTS.md` that has none of the 8 guidance
sections. Verify all 8 are injected. Run `/uf-init`
again. Verify none are duplicated (idempotent).

**Acceptance Scenarios**:

1. **Given** a repo's `AGENTS.md` has no behavioral
   guidance sections, **When** the engineer runs
   `/uf-init`, **Then** all 8 standardized sections
   are appended to `AGENTS.md` at appropriate locations.
2. **Given** a repo's `AGENTS.md` already has some of
   the 8 sections (e.g., from manual addition), **When**
   the engineer runs `/uf-init`, **Then** only the
   missing sections are added (existing ones preserved).
3. **Given** a repo's `AGENTS.md` already has all 8
   sections, **When** the engineer runs `/uf-init`,
   **Then** no changes are made to `AGENTS.md`
   (idempotent).
4. **Given** a repo has no `AGENTS.md` at all, **When**
   the engineer runs `/uf-init`, **Then** the AGENTS.md
   guidance step is skipped with a message: "No
   AGENTS.md found — skipping behavioral guidance
   injection."

---

### User Story 2 — Consistent Quality Gates (Priority: P1)

The 3 most critical behavioral rules — gatekeeping
value protection, workflow phase boundaries, and CI
parity gate — are present in every repo after running
`/uf-init`. These rules prevent the most damaging agent
behaviors: modifying quality thresholds, writing code
during planning, and skipping CI checks.

**Why this priority**: These 3 rules are the highest-
impact behavioral controls. Without them, an agent can
silently degrade the project's quality posture.

**Independent Test**: Run `/uf-init` in the Gaze repo
(which has CI parity gate but not gatekeeping protection
or phase boundaries). Verify all 3 are present after.

**Acceptance Scenarios**:

1. **Given** a repo has no gatekeeping protection
   section, **When** `/uf-init` runs, **Then** the
   gatekeeping protection block is added with all 8
   protected value categories and the "stop and report"
   instruction.
2. **Given** a repo has no workflow phase boundaries
   section, **When** `/uf-init` runs, **Then** the
   phase boundary block is added mapping each pipeline
   phase to its allowed output type.
3. **Given** a repo has no CI parity gate, **When**
   `/uf-init` runs, **Then** the CI parity gate block
   is added requiring agents to replicate CI checks
   from `.github/workflows/` before marking tasks
   complete.

---

### User Story 3 — Workflow and Documentation Rules (Priority: P2)

The 3 workflow rules — review council as PR
prerequisite, website documentation sync gate, and
spec-first development with exemptions — are present
in every repo after running `/uf-init`.

**Why this priority**: These rules enforce process
discipline but are less immediately damaging when
missing than the P1 quality gates. An agent that skips
review council produces lower-quality code but doesn't
permanently degrade thresholds.

**Independent Test**: Run `/uf-init` in the Website repo
(which has minimal behavioral guidance). Verify all 3
workflow rules are present after.

**Acceptance Scenarios**:

1. **Given** a repo has no review council PR
   prerequisite, **When** `/uf-init` runs, **Then** the
   5-step workflow (complete → CI → review-council →
   commit → PR) with exemptions is added.
2. **Given** a repo has no website documentation sync
   gate, **When** `/uf-init` runs, **Then** the gate
   requiring website issues for user-facing changes is
   added (with `gh issue create` template and
   exemption list).
3. **Given** a repo has no spec-first development
   section, **When** `/uf-init` runs, **Then** the
   section is added listing what requires a spec, what
   is exempt, and the "when unsure, ask the user" rule.

---

### User Story 4 — Knowledge Retrieval and Core Mission (Priority: P3)

The knowledge retrieval section (with Dewey tool
selection matrix and 3-tier degradation) and core
mission statement (3 strategic framing bullets) are
present in every repo after running `/uf-init`.

**Why this priority**: These are contextual guidance
rather than behavioral gates. Agents work correctly
without them but produce better results when they know
how to use Dewey and understand the project's strategic
framing.

**Independent Test**: Run `/uf-init` in the Replicator
repo (which has knowledge retrieval but no core
mission). Verify the core mission is added and the
existing knowledge retrieval is preserved.

**Acceptance Scenarios**:

1. **Given** a repo has no knowledge retrieval section,
   **When** `/uf-init` runs, **Then** the section is
   added with the Dewey tool selection matrix, fallback
   criteria, and 3-tier degradation pattern.
2. **Given** a repo has no core mission statement,
   **When** `/uf-init` runs, **Then** the 3 strategic
   framing bullets (Strategic Architecture, Outcome
   Orientation, Intent-to-Context) are added.
3. **Given** a repo already has both sections, **When**
   `/uf-init` runs, **Then** neither section is
   modified or duplicated.

---

### Edge Cases

- What happens when `AGENTS.md` uses different section
  headings than expected (e.g., "Constraints" instead of
  "Behavioral Constraints")? The AI agent SHOULD check
  for the specific detection phrases defined in
  `data-model.md` for each block — a primary heading
  match and a semantic fallback phrase. If either is
  found, the block is treated as already present.
- What happens when a guidance section exists but with
  different wording? The AI agent SHOULD check for the
  detection phrases (see `data-model.md` Block
  Definitions) and skip injection if any match is found.
  It SHOULD NOT inject a duplicate section with slightly
  different wording.
- What happens when `AGENTS.md` is very large (1000+
  lines)? The AI agent SHOULD still find the correct
  insertion points — section headings and document
  structure guide placement.
- What happens when the user wants to customize a
  guidance section for their repo? The user can edit
  the injected section after `/uf-init` runs. Running
  `/uf-init` again will not overwrite customized
  sections (the idempotency check detects the section
  heading is present).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `/uf-init` MUST check `AGENTS.md` for
  each of the 8 standardized guidance sections and
  inject any that are missing.
- **FR-002**: Each injection MUST be idempotent — if
  the section heading (or semantic equivalent) already
  exists, the section MUST NOT be duplicated.
- **FR-003**: The 8 guidance sections MUST include:
  1. Core Mission (3 strategic framing bullets)
  2. Gatekeeping Value Protection (8 protected
     categories + stop-and-report)
  3. Workflow Phase Boundaries (phase-to-output mapping)
  4. CI Parity Gate (replicate CI checks from workflow
     files)
  5. Review Council PR Prerequisite (5-step workflow +
     exemptions)
  6. Website Documentation Sync Gate (issue creation +
     exemptions)
  7. Spec-First Development (what requires spec + what
     is exempt + when-unsure rule)
  8. Knowledge Retrieval (Dewey tool matrix + 3-tier
     degradation + fallback criteria)
- **FR-004**: If `AGENTS.md` does not exist in the repo,
  the guidance injection step MUST be skipped with an
  informational message.
- **FR-005**: The injection text for each section MUST
  be defined inline in the `/uf-init` command file
  instructions.
- **FR-006**: The `/uf-init` report summary MUST include
  a "AGENTS.md Guidance" section showing which blocks
  were injected, which were already present, and which
  were skipped.
- **FR-007**: The scaffold asset copy of `uf-init.md`
  MUST be synchronized after modification.
- **FR-008**: All existing tests MUST continue to pass.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `/uf-init` in a repo with a
  minimal `AGENTS.md` (zero guidance sections) results
  in all 8 sections being injected — verified by
  checking for each section heading after running.
- **SC-002**: Running `/uf-init` twice in the same repo
  produces identical `AGENTS.md` content (idempotent) —
  verified by comparing file checksums before and after
  the second run.
- **SC-003**: Running `/uf-init` in a repo that already
  has all 8 sections produces zero modifications —
  verified by checking `git diff` is empty after
  running.
- **SC-004**: All 6 Unbound Force repos have all 8
  guidance sections after running `/uf-init` in each —
  verified by the cross-repo audit table showing YES
  for all cells.
- **SC-005**: All existing tests pass after the changes
  — verified by running the full test suite.

## Dependencies & Assumptions

### Dependencies

- **`/uf-init` command** (existing): The injection
  mechanism. Already handles speckit guardrails, Dewey
  context, and OpenSpec guardrails.

### Assumptions

- Each repo has an `AGENTS.md` file at the repo root.
  If not present, the guidance step is skipped gracefully.
- The AI agent executing `/uf-init` can read `AGENTS.md`,
  understand its structure, and find appropriate
  insertion points for each section based on the
  document's existing headings.
- The guidance text is generic enough to apply to any
  repo in the ecosystem (no repo-specific content in
  the injected sections). Repo-specific customization
  is done by the user editing the injected text after
  `/uf-init` runs.
- The idempotency check uses section heading matching
  (semantic, not exact string) — if a section with the
  same concept exists under a different heading, it is
  treated as "already present."
