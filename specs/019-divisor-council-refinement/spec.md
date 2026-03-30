---
spec_id: "019"
title: "Divisor Council Refinement"
status: complete
created: 2026-03-30
branch: 019-divisor-council-refinement
phase: 1
depends_on:
  - "[[specs/005-the-divisor-architecture/spec]]"
---

# Feature Specification: Divisor Council Refinement

**Feature Branch**: `019-divisor-council-refinement`
**Created**: 2026-03-30
**Status**: Draft
**Input**: Refine the Divisor Council agents: delete
legacy reviewer files, de-duplicate cross-persona
responsibilities, harmonize severity definitions,
qualify FR IDs, integrate learning loop, add static
analysis to CI.

## User Scenarios & Testing *(mandatory)*

### User Story 1 -- Zero-Waste Cleanup (Priority: P1)

A developer runs `uf init` in their repo and receives
only the 5 active Divisor persona agents (`divisor-*`).
The 5 legacy `reviewer-*` files are no longer deployed.
The `/review-council` command discovers exactly 5
agents and produces no findings from stale legacy
personas.

**Why this priority**: The legacy files violate the
Zero-Waste Mandate, use a different (weaker) AI model,
contain stale binary name references, and could be
accidentally invoked if the discovery logic is ever
modified. Removing them is the simplest, highest-value
change.

**Independent Test**: Run `uf init` in a temp
directory. Verify no `reviewer-*.md` files exist in
`.opencode/agents/`. Run `/review-council` and verify
only `divisor-*` agents are discovered.

**Acceptance Scenarios**:

1. **Given** a repo with legacy `reviewer-*.md` files,
   **When** `uf init` is run,
   **Then** no `reviewer-*.md` files are deployed (only
   `divisor-*.md` agents are scaffolded).

2. **Given** a repo where `reviewer-*.md` files were
   previously scaffolded,
   **When** `uf init` is run after this change,
   **Then** the legacy files remain on disk but `uf init`
   prints a warning listing the detected legacy files
   and suggesting manual removal (e.g., "Legacy reviewer
   agents detected: reviewer-adversary.md, ... These
   have been superseded by divisor-* agents. Remove
   with: rm .opencode/agents/reviewer-*.md").

3. **Given** `/review-council` is invoked,
   **When** it discovers available reviewer agents,
   **Then** only `divisor-*` agents are found (the
   discovery pattern `divisor-*.md` excludes legacy
   `reviewer-*` files).

---

### User Story 2 -- De-duplicated Review Findings (Priority: P1)

A developer runs `/review-council` on their code and
receives a report with no duplicate findings. Each
review dimension (secrets, test isolation, plan
alignment, dependency health, etc.) is covered by
exactly one Divisor persona with clear ownership
boundaries.

**Why this priority**: Duplicate findings waste
developer time and erode trust in the review system.
When 2 agents both flag "hardcoded secrets" with
different severity levels, the developer gets confused
about which finding to act on.

**Independent Test**: Run `/review-council` on a
codebase with a known hardcoded secret. Verify exactly
one agent (the Adversary) flags it. Verify no other
agent produces a finding about the same secret.

**Acceptance Scenarios**:

1. **Given** code containing a hardcoded API key,
   **When** `/review-council` runs all 5 Divisor agents,
   **Then** only the Adversary flags the secret, and
   the SRE's findings focus on operational concerns
   without duplicating the credential check.

2. **Given** a test file with shared mutable state,
   **When** `/review-council` runs,
   **Then** only the Tester flags the isolation issue,
   and the Adversary's findings focus on security
   concerns without duplicating the test isolation
   check.

3. **Given** implementation that drifts from the
   approved plan,
   **When** `/review-council` runs,
   **Then** only the Guard flags the intent drift, and
   the Architect's findings focus on structural
   patterns without duplicating the plan alignment
   check.

---

### User Story 3 -- Consistent Severity Classification (Priority: P2)

All 5 Divisor agents use the same severity definitions
with domain-specific examples. The `/review-council`
auto-fix policy (fix LOW/MEDIUM, report HIGH/CRITICAL)
operates on a predictable, calibrated baseline. Two
agents assigning different severity levels to the same
class of finding no longer occurs.

**Why this priority**: Without shared severity
definitions, the auto-fix logic in `/review-council`
Spec Review Mode is unpredictable. An issue the
Adversary calls MEDIUM (auto-fixed) might be the same
class of issue the SRE calls HIGH (reported only).

**Independent Test**: Read all 5 Divisor agent files
and verify each contains identical severity definition
references. Run `/review-council` on a spec with a
formatting issue and verify it's consistently classified
as LOW across all agents that might notice it.

**Acceptance Scenarios**:

1. **Given** all 5 Divisor agents have been updated,
   **When** a reviewer reads the severity definitions,
   **Then** all 5 agents reference the same shared
   severity standard with consistent level boundaries.

2. **Given** a formatting inconsistency in a spec,
   **When** two different agents both notice it,
   **Then** both assign the same severity level (LOW).

---

### User Story 4 -- Qualified Requirement References (Priority: P2)

When a Divisor agent references a functional requirement
in its findings, it uses the fully qualified format
"per Spec NNN FR-XXX" instead of just "FR-XXX". This
eliminates ambiguity when the same FR number exists in
multiple specs.

**Why this priority**: FR-020 exists in 7 files across
6 specs with different meanings. Unqualified references
confuse the AI agent and can produce hallucinatory
compliance checks.

**Independent Test**: Run `/review-council` and verify
all findings that reference functional requirements use
the qualified format. Search agent file text for bare
"FR-" patterns without spec qualifiers.

**Acceptance Scenarios**:

1. **Given** a Divisor agent produces a finding
   referencing a functional requirement,
   **When** the finding text is examined,
   **Then** it uses the format "per Spec NNN FR-XXX"
   (e.g., "per Spec 005 FR-020").

2. **Given** the agent instruction text in `divisor-*`
   files,
   **When** searched for "FR-" references,
   **Then** all instances include a spec number
   qualifier.

---

### User Story 5 -- Learning-Informed Reviews (Priority: P3)

The Divisor agents consult prior learnings stored in
semantic memory before beginning their review. If a
previous `/unleash` session discovered that a specific
file has a recurring gotcha (e.g., "scaffold.go
requires initSubTools nil guard for Stdout"), the
reviewing agent references that learning in its
analysis, producing more targeted and context-aware
findings.

**Why this priority**: Without learning loop
integration, the Divisor council is a stateless auditor
that repeats the same generic checks every time. With
it, the council becomes a dynamic system that
accumulates project-specific wisdom across sessions.

**Independent Test**: Store a learning via
`hivemind_store` about a specific file. Run
`/review-council` on changes to that file. Verify the
relevant agent references the stored learning in its
review context.

**Acceptance Scenarios**:

1. **Given** a learning stored in Hivemind about a
   recurring pattern in `scaffold.go`,
   **When** a Divisor agent reviews changes to
   `scaffold.go`,
   **Then** the agent's review context includes the
   stored learning as prior knowledge.

2. **Given** Hivemind is not available,
   **When** a Divisor agent begins its review,
   **Then** the learning lookup step is skipped with
   an informational note and the review proceeds
   without prior learnings.

---

### User Story 6 -- Static Analysis in CI (Priority: P3)

The CI pipeline and `/review-council` Phase 1a include
static analysis tools (`golangci-lint` for comprehensive
linting, `govulncheck` for known vulnerability
detection). The Divisor Adversary and SRE agents
receive concrete tool output to augment their
LLM-based review, rather than relying on manual
inspection for CVE and lint checks.

**Why this priority**: The Adversary and SRE agents
claim to check for dependency CVEs and lint violations,
but no tooling supports these checks. Adding static
analysis tools provides machine-verifiable evidence
that backs the agents' quality claims (Constitution
Principle III: Observable Quality).

**Independent Test**: Run the CI workflow and verify
`golangci-lint` and `govulncheck` produce output. Run
`/review-council` and verify Phase 1a includes these
tools' output.

**Acceptance Scenarios**:

1. **Given** the CI workflow runs on a PR,
   **When** the build and test step executes,
   **Then** `golangci-lint` and `govulncheck` are run
   and their output is visible in the CI logs.

2. **Given** `/review-council` runs in Code Review Mode,
   **When** Phase 1a derives commands from
   `.github/workflows/`,
   **Then** it executes `golangci-lint` and
   `govulncheck` locally alongside the existing build,
   vet, and test commands.

3. **Given** `govulncheck` finds a known vulnerability,
   **When** the Adversary reviews the code,
   **Then** the finding references the specific CVE
   from the tool output (not a generic "check for
   CVEs" instruction).

---

### Edge Cases

- What happens when a repo has both `reviewer-*.md`
  and `divisor-*.md` files from a previous `uf init`?
  The legacy files remain on disk. `uf init` warns
  about their presence and suggests manual removal.
  The `/review-council` discovery only matches
  `divisor-*` pattern, so legacy files are never
  invoked regardless.
- What happens when Hivemind is not available for the
  learning lookup? The step is skipped with an
  informational note. The review proceeds without
  prior learnings (graceful degradation).
- What happens when `golangci-lint` or `govulncheck`
  is not installed? `/review-council` Phase 1a derives
  commands from CI workflow files. If CI doesn't run
  these tools, Phase 1a doesn't either. The Adversary
  and SRE agents revert to LLM-based review without
  tool output.
- What happens when the shared severity convention
  pack is missing? Each agent MUST have a fallback
  severity definition inline. The convention pack is
  the preferred source, but agents MUST NOT fail if it
  is absent.
- What happens when `golangci-lint` produces findings?
  Any non-zero exit code from `golangci-lint` is a gate
  failure -- the review council stops before invoking
  Divisor agents, same as a `go build` failure. The
  project controls what constitutes an error vs. warning
  via `.golangci.yml` configuration, not via the review
  council's interpretation.
- What happens when `govulncheck` finds vulnerabilities?
  Same gate behavior as `golangci-lint` — non-zero exit
  code = gate failure. `govulncheck` exits non-zero
  when vulnerabilities are found in called code paths.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf init` MUST NOT deploy `reviewer-*.md`
  files. Only `divisor-*.md` agents are scaffolded.
- **FR-002**: The scaffold asset set MUST NOT contain
  any `reviewer-*.md` files under
  `internal/scaffold/assets/opencode/agents/`.
- **FR-003**: The expected file count and asset path
  list in scaffold tests MUST reflect the removal of
  5 legacy files.
- **FR-003a**: `uf init` MUST detect previously
  scaffolded `reviewer-*.md` files in
  `.opencode/agents/` and print a warning listing
  them with a suggested removal command. It MUST NOT
  delete the files.
- **FR-004**: Each review dimension (secrets, test
  isolation, plan alignment, dependency health, file
  permissions, hardcoded values, zero-waste,
  constitution alignment) MUST be owned by exactly one
  Divisor persona.
- **FR-005**: The ownership mapping MUST be:
  - Secrets/credentials → Adversary
  - Dependency CVEs/supply chain → Adversary
  - Error handling/resilience → Adversary
  - Test isolation/coverage/depth → Tester
  - Plan alignment/intent drift → Guard
  - Zero-waste mandate → Guard
  - Constitution alignment → Guard
  - File permissions/hardcoded config → SRE
  - Efficiency/performance (O(n²), allocations) → SRE
  - Architectural patterns/conventions → Architect
- **FR-006**: All 5 Divisor agents MUST reference a
  shared severity definitions standard with
  CRITICAL/HIGH/MEDIUM/LOW levels and domain-specific
  examples for each level.
- **FR-007**: The shared severity definitions SHOULD be
  defined in a convention pack file at
  `.opencode/unbound/packs/severity.md` (tool-owned).
- **FR-008**: When referencing functional requirements,
  Divisor agents MUST use the fully qualified format
  "per Spec NNN FR-XXX" to avoid cross-spec ambiguity.
- **FR-009**: All 5 Divisor agents MUST include a
  "Prior Learnings" step at the start of their review
  workflow that searches Hivemind for learnings tagged
  with the repo name and relevant file paths.
- **FR-010**: The Prior Learnings step MUST gracefully
  degrade when Hivemind is not available (skip with
  informational note).
- **FR-011**: The CI workflow (`.github/workflows/
  test.yml`) MUST include `golangci-lint run` and
  `govulncheck ./...` as additional quality checks.
- **FR-012**: `uf setup` MUST install `golangci-lint`
  and `govulncheck` as part of its tool chain setup
  (or document them as prerequisites).
- **FR-013**: `/review-council` Phase 1a MUST execute
  `golangci-lint` and `govulncheck` when they appear
  in the CI workflow files (derived from
  `.github/workflows/`, not hardcoded).

### Key Entities

- **Divisor Persona**: One of 5 review agents with a
  defined ownership domain. Each persona has mutually
  exclusive primary review dimensions.
- **Severity Level**: One of CRITICAL/HIGH/MEDIUM/LOW
  with shared definitions and domain-specific examples
  across all personas.
- **Convention Pack**: A shared configuration file
  (`.opencode/unbound/packs/*.md`) containing standards
  that agents reference during review. Tool-owned and
  auto-updated by `uf init`.
- **Prior Learning**: A semantic memory entry stored
  via `hivemind_store` from a previous session,
  queryable via `hivemind_find` by file path or topic.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `/review-council` on the same
  codebase produces zero duplicate findings across the
  5 Divisor agents (each finding appears in exactly
  one agent's report).
- **SC-002**: All 5 Divisor agents reference the same
  severity standard, and two agents classifying the
  same type of issue assign the same severity level.
- **SC-003**: The scaffold file count decreases by 3
  (net) after removing 4 legacy `reviewer-*.md` files
  and adding 1 `severity.md` pack (52 → 49, adjusted
  for any other changes).
- **SC-004**: When prior learnings exist for a file
  under review, at least one Divisor agent references
  them in its review context.
- **SC-005**: The CI pipeline catches known
  vulnerabilities via `govulncheck` and lint violations
  via `golangci-lint` before code reaches the review
  council.
- **SC-006**: All FR references in Divisor agent files
  use the qualified "per Spec NNN FR-XXX" format with
  zero bare "FR-" patterns.

## Clarifications

### Session 2026-03-30

- Q: Should efficiency checks (O(n²) loops, redundant
  reads, allocations) stay in the Adversary or move to
  the SRE? → A: Move to SRE. Adversary focuses purely
  on security/resilience. SRE handles operational
  concerns including performance/efficiency.
- Q: How should golangci-lint gate behavior work in
  Phase 1a? → A: Non-zero exit code = gate failure,
  same as go build and go test. Project controls
  error vs. warning via .golangci.yml config.
- Q: How should govulncheck gate behavior work? →
  A: Same as golangci-lint — non-zero exit code =
  gate failure. govulncheck exits non-zero when
  vulnerabilities are found in called code, which is
  the desired behavior.
- Q: Should uf init clean up previously-scaffolded
  legacy reviewer-*.md files? → A: Soft cleanup --
  warn about detected legacy files and suggest manual
  removal, but do not delete them.

## Assumptions

- The `/review-council` discovery pattern (`divisor-*`)
  already excludes legacy `reviewer-*` files. No change
  to the discovery logic is needed.
- The `reviewer-*.md` files in existing repos will
  remain on disk after `uf init` runs. `uf init` warns
  about their presence and suggests manual removal.
  Developers can remove them with
  `rm .opencode/agents/reviewer-*.md`. Release notes
  should also mention this migration step.
- `golangci-lint` is available via `go install` or
  Homebrew. `govulncheck` is available via
  `go install golang.org/x/vuln/cmd/govulncheck@latest`.
- The shared severity convention pack follows the same
  ownership model as existing packs: tool-owned
  (`severity.md`) with no user-owned custom stub
  (severity definitions should not be customized
  per-repo).
- The Hivemind learning lookup uses `hivemind_find`
  with a query constructed from the files being
  reviewed. If no results are returned, the review
  proceeds without prior learnings.
- Adding `golangci-lint` to CI may produce new findings
  in the existing codebase that need to be fixed before
  the CI gate passes. This is expected and desirable --
  it raises the quality bar.
