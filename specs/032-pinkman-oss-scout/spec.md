# Feature Specification: Pinkman OSS Scout

**Feature Branch**: `032-pinkman-oss-scout`
**Created**: 2026-04-22
**Status**: Complete
**Input**: User description: "Create an agent that is
called pinkman that will retrieve new or update open
source projects that track with industry trends and are
compatible with the open source licensing within the
unbound-force project. Do not replicate any of the
existing agentic capabilities within unbound-force."

## Clarifications

### Session 2026-04-22

- Q: Should Pinkman check shared dependencies for
  OSI-approved licenses (one level deep)? → A: No --
  list overlaps only, do not check shared dependency
  licenses (defer to future enhancement).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Discover License-Compatible OSS Projects (Priority: P1)

An engineer working on the Unbound Force ecosystem wants
to find new open source projects that could benefit
their work. They invoke Pinkman, which scans public
sources for open source projects relevant to their
domain (AI agent tooling, developer infrastructure,
static analysis, CLI frameworks, etc.) and filters
results to only those with Open Source Initiative (OSI)
approved licenses. For each suggested project, Pinkman
lists that project's direct dependencies and highlights
any dependencies shared across multiple suggested
projects, giving the engineer visibility into supply
chain overlap. The engineer receives a curated list of
projects with summaries, license status, direct
dependencies, overlap indicators, and relevance
signals, enabling informed adoption decisions without
manual research.

**Why this priority**: License compatibility is the
non-negotiable prerequisite for any OSS adoption.
Without this capability, all other Pinkman features are
moot -- an incompatible project cannot be adopted
regardless of how trendy or useful it is.

**Independent Test**: Can be fully tested by invoking
Pinkman with a domain keyword (e.g., "static analysis
Go") and verifying it returns a list of projects that
all carry OSI-approved licenses, with no non-approved
or proprietary projects in the results, and each result
includes a list of its direct dependencies.

**Acceptance Scenarios**:

1. **Given** the Unbound Force project uses Apache 2.0
   licensing, **When** Pinkman is asked to find projects
   for a given domain keyword, **Then** it returns only
   projects whose licenses appear on the OSI-approved
   license list.
2. **Given** a search returns projects with mixed
   licensing, **When** Pinkman filters results, **Then**
   projects with non-OSI-approved licenses are excluded
   from the primary results and flagged separately with
   an explanation of why the license is not OSI-approved.
3. **Given** a project has no detectable license,
   **When** Pinkman encounters it, **Then** it is flagged
   as "license unknown -- manual review required" and
   excluded from the compatible results list.
4. **Given** Pinkman returns multiple suggested projects,
   **When** the results are presented, **Then** each
   project includes a list of its direct dependencies,
   and dependencies that appear in more than one
   suggested project are highlighted as shared
   dependencies with a count of how many projects use
   them.

---

### User Story 2 - Track Industry Trends for OSS Adoption (Priority: P2)

An engineer wants to stay current with industry trends
in open source tooling relevant to the Unbound Force
ecosystem. They invoke Pinkman to identify trending
projects -- those gaining significant community traction
(stars, forks, contributor growth, release velocity) --
within relevant categories. Pinkman surfaces projects
that are both trending and license-compatible, helping
the team proactively adopt valuable tools before they
become industry defaults.

**Why this priority**: Trend tracking transforms Pinkman
from a reactive search tool into a proactive advisor.
Once license compatibility is established (P1), trend
awareness lets the team identify high-value adoption
candidates early, gaining competitive advantage.

**Independent Test**: Can be tested by invoking Pinkman
in trend-scanning mode for a category (e.g., "MCP
servers") and verifying it returns projects sorted by
trend signals with quantitative indicators (growth
rate, community activity metrics).

**Acceptance Scenarios**:

1. **Given** the engineer requests trending projects in a
   category, **When** Pinkman analyzes public project
   data, **Then** it returns projects ranked by trend
   strength with at least three quantitative indicators
   (e.g., star growth rate, fork count trajectory,
   release frequency).
2. **Given** a project is trending but has a non-OSI-approved
   license, **When** Pinkman presents results, **Then**
   that project appears in a separate "trending but
   not OSI-approved" section with the license issue
   explained.
3. **Given** no projects are trending in the requested
   category, **When** Pinkman completes its scan,
   **Then** it reports "no significant trends detected"
   with the date range and sources consulted.

---

### User Story 3 - Monitor Existing Dependencies for Updates (Priority: P3)

An engineer wants to know when existing open source
dependencies used by the Unbound Force ecosystem have
significant updates -- new major versions, security
patches, deprecation notices, or maintainer changes that
could affect project health. They invoke Pinkman to
audit current dependencies and report actionable
findings, such as available updates, end-of-life
warnings, or license changes in newer versions.

**Why this priority**: Monitoring existing dependencies
is essential for ongoing project health but is lower
priority than new discovery (P1) and trend tracking
(P2) because existing dependencies are already known
and managed through standard tooling. Pinkman adds
value here by surfacing higher-level signals (license
changes, maintainer health, deprecation risk) that
standard dependency update tools do not catch.

**Independent Test**: Can be tested by pointing Pinkman
at a project's dependency manifest (e.g., `go.mod`) and
verifying it identifies at least one dependency with an
available update and correctly reports its current vs.
latest version, any license changes, and maintenance
health indicators.

**Acceptance Scenarios**:

1. **Given** a project dependency manifest exists,
   **When** Pinkman audits dependencies, **Then** it
   reports each dependency's current version, latest
   available version, and whether the license has changed
   between versions.
2. **Given** a dependency has changed its license in a
   newer version to a non-OSI-approved license,
   **When** Pinkman reports the update, **Then** it
   prominently warns about the license change and
   recommends staying on the current version or finding
   an alternative.
3. **Given** a dependency shows signs of reduced
   maintenance (no commits in 12+ months, unresolved
   critical issues, archived repository), **When**
   Pinkman audits it, **Then** it flags the dependency
   as "maintenance risk" with the specific indicators.

---

### User Story 4 - Generate Adoption Recommendation Reports (Priority: P3)

After discovery and trend analysis, an engineer wants
Pinkman to produce a structured recommendation report
for a specific project under consideration. The report
summarizes license compatibility, community health,
maintenance signals, trend trajectory, and how the
project relates to existing Unbound Force dependencies.
This report can be shared with the team or stored in
the knowledge graph for future reference.

**Why this priority**: Reports are a presentation layer
on top of discovery (P1) and trend data (P2). They add
organizational value but are not required for the core
scouting function to work.

**Independent Test**: Can be tested by requesting a
report for a known open source project and verifying
the output contains all required sections (license
analysis, community health, trend data, relationship
to existing dependencies) in a structured format.

**Acceptance Scenarios**:

1. **Given** an engineer requests a recommendation report
   for a specific project, **When** Pinkman generates
   the report, **Then** it includes sections for: license
   compatibility verdict (OSI-approved status), community
   health score, maintenance indicators, trend
   trajectory, direct dependency list, shared dependency
   overlap with other evaluated projects, and
   relationship to existing Unbound Force dependencies.
2. **Given** Pinkman generates a report, **When** the
   report is complete, **Then** it is formatted as a
   structured artifact that can be stored in the
   knowledge graph via Dewey.

### Edge Cases

- What happens when a project uses a dual-license model
  (e.g., MIT OR Apache-2.0)? Pinkman MUST evaluate each
  license option and report compatibility based on the
  most favorable compatible option.
- What happens when a project's license is a custom or
  non-standard license? Pinkman MUST flag it as
  "non-standard license -- manual legal review required"
  and exclude it from automatic compatibility verdicts.
- What happens when public data sources are temporarily
  unavailable? Pinkman MUST report which sources failed,
  present partial results from available sources, and
  indicate the results are incomplete.
- What happens when a dependency's repository has been
  transferred to a new owner or organization? Pinkman
  MUST detect the transfer and report the ownership
  change as a risk signal.
- What happens when the user requests scouting for a
  domain with no relevant open source projects? Pinkman
  MUST report "no projects found" with the search
  criteria used, rather than returning unrelated results.
- What happens when two suggested projects share a
  dependency at conflicting versions? Pinkman MUST
  report the version discrepancy as a supply chain
  signal alongside the overlap indicator.
- What happens when a scouted project has no declared
  dependency manifest? Pinkman MUST report "dependencies
  unknown -- no manifest detected" and omit the project
  from overlap analysis.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Pinkman MUST discover open source projects
  relevant to a user-specified domain or keyword query
  from public sources.
- **FR-002**: Pinkman MUST determine the license of each
  discovered project and classify it as OSI-approved or
  not OSI-approved, using the Open Source Initiative's
  published list of approved licenses as the
  authoritative source.
- **FR-003**: Pinkman MUST use the OSI-approved license
  list (maintained by the Open Source Initiative at
  opensource.org) as the sole authority for license
  compatibility. A license is compatible if and only if
  it appears on the current OSI-approved list.
- **FR-004**: Pinkman MUST quantify industry trend
  signals for discovered projects using at least three
  measurable indicators (e.g., star growth rate, fork
  trajectory, release frequency, contributor growth).
- **FR-005**: Pinkman MUST audit a project's dependency
  manifest to identify available updates for existing
  dependencies.
- **FR-006**: Pinkman MUST detect license changes between
  the currently used version and the latest available
  version of a dependency.
- **FR-014**: Pinkman MUST list the direct dependencies
  of each scouted project when a dependency manifest is
  available.
- **FR-015**: Pinkman MUST identify and highlight
  dependencies that are shared across multiple scouted
  projects in a single result set, reporting the
  dependency name and the count of projects that use it.
- **FR-016**: When shared dependencies appear at
  conflicting versions across scouted projects, Pinkman
  MUST report the version discrepancy as a supply chain
  signal.
- **FR-007**: Pinkman MUST flag dependencies with
  maintenance risk indicators (no commits in 12+ months,
  archived repository, unresolved critical issues).
- **FR-008**: Pinkman MUST produce structured
  recommendation reports containing license analysis,
  community health, maintenance signals, trend data,
  and dependency relationships.
- **FR-009**: Pinkman MUST NOT replicate capabilities of
  existing Unbound Force heroes -- specifically: product
  backlog management (Muti-Mind), code implementation
  (Cobalt-Crush), testing and quality validation (Gaze),
  code review (The Divisor), or process coaching (Mx F).
- **FR-010**: Pinkman MUST operate as a standalone agent
  that produces self-describing artifacts, per the
  Autonomous Collaboration principle (Spec 001).
- **FR-011**: Pinkman MUST handle dual-license projects
  by evaluating each license option against the
  OSI-approved list and reporting the most favorable
  approved option.
- **FR-012**: Pinkman MUST gracefully handle unavailable
  data sources by reporting partial results with clear
  indication of which sources were consulted and which
  failed.
- **FR-013**: Pinkman SHOULD integrate with the Dewey
  knowledge layer to store and retrieve scouting
  results, enabling cross-session awareness of
  previously evaluated projects.

### Key Entities

- **Scouted Project**: An open source project discovered
  by Pinkman. Key attributes: name, repository URL,
  description, primary language, license identifier,
  OSI-approved verdict, trend indicators, community
  health metrics, direct dependency list, last updated
  date.
- **Dependency Overlap**: A dependency that appears in
  two or more scouted projects within a single result
  set. Key attributes: dependency name, list of scouted
  projects that use it, version used by each project,
  version conflict flag (true if versions differ).
- **License Compatibility Assessment**: The result of
  evaluating a project's license against the OSI-approved
  list. Key attributes: license SPDX identifier,
  OSI-approved verdict (approved, not approved, unknown,
  manual review required), explanation of the verdict.
- **Dependency Health Report**: The result of auditing
  an existing dependency. Key attributes: dependency
  name, current version, latest version, license change
  detected (boolean), maintenance risk level (healthy,
  warning, critical), specific risk indicators.
- **Adoption Recommendation**: A structured report for
  a specific project. Key attributes: project identity,
  license analysis, community health score, maintenance
  signals, trend trajectory, relationship to existing
  Unbound Force dependencies, overall recommendation.

## Dependencies and Assumptions

### Dependencies

- **Spec 001** (Org Constitution): Pinkman must adhere
  to the four core principles (Autonomous Collaboration,
  Composability First, Observable Quality, Testability).
- **Spec 002** (Hero Interface Contract): If Pinkman is
  implemented as a full hero, it must follow the
  standard hero repo structure and artifact envelope
  format.
- **Spec 009** (Shared Data Model): Recommendation
  reports should use the standard artifact envelope
  format for inter-hero consumption.
- **Spec 014/015** (Dewey Architecture/Integration):
  Pinkman's integration with the knowledge layer
  depends on Dewey being available.

### Assumptions

- The Open Source Initiative (OSI) approved license list
  is the authoritative reference for compatibility
  decisions. SPDX identifiers are used as the standard
  license identifier taxonomy for consistent naming.
- Pinkman uses publicly available data sources (e.g.,
  GitHub, package registries) for project discovery.
  No proprietary or paid data sources are required for
  the initial implementation.
- License compatibility is assessed at the project
  level (root license file). Direct dependencies of
  scouted projects are listed for supply chain
  visibility, but their licenses are not individually
  checked. Transitive dependency license analysis is
  out of scope for the initial version but may be
  added later.
- Trend indicators are derived from publicly available
  metrics. Pinkman does not perform sentiment analysis
  on social media or news articles.
- Pinkman is a non-hero utility agent (similar to the
  onboarding agent) unless the team decides to elevate
  it to full hero status. This keeps the initial scope
  manageable.

## Scope Boundaries

### In Scope

- Discovery of open source projects by domain/keyword
- License compatibility classification using the
  OSI-approved license list
- Direct dependency listing for each scouted project
- Dependency overlap detection across scouted projects
  (shared dependency identification and version
  conflict signals)
- Industry trend quantification using public metrics
- Existing dependency health auditing (from manifest
  files)
- Structured recommendation report generation
- Dewey integration for persistent scouting memory

### Out of Scope

- Automatic adoption or installation of discovered
  projects (that is Cobalt-Crush's domain)
- Legal review or legal advice on license compatibility
  (Pinkman provides technical classification, not legal
  counsel)
- Source code analysis of discovered projects (that is
  Gaze's domain)
- Prioritization of which projects to adopt (that is
  Muti-Mind's domain)
- Review of adopted project integration quality (that
  is The Divisor's domain)
- License checking of scouted projects' dependencies
  (direct dependencies are listed for overlap visibility
  but their individual licenses are not checked)
- Transitive dependency scanning (future enhancement)
- Proprietary or paid data source integration

## Maintenance

### Static Data Review

The agent file (`pinkman.md`) contains two static data
structures that require periodic review:

1. **Fallback License List** (26 SPDX identifiers): A
   hardcoded list of OSI-approved licenses used when the
   OSI website is unreachable. Review annually or when
   OSI approves/revokes license categories.

2. **Compatibility Tier Table**: Maps SPDX identifiers
   to permissive/weak-copyleft/strong-copyleft tiers.
   Review annually or when new copyleft licenses gain
   significant adoption.

Both structures are embedded in `pinkman.md`, which is
**user-owned** (not overwritten by `uf init`). Updates
to the canonical scaffold copy do not propagate to
existing deployments automatically. Users must manually
update their `pinkman.md` or re-scaffold.

A review date marker is embedded as an HTML comment in
the tier table. The first review is due 2027-04-25.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can discover license-compatible open
  source projects for a given domain keyword. First
  results appear within 60 seconds under normal network
  conditions; complete results may take longer depending
  on result count, network conditions, and rate limiting.
- **SC-002**: 100% of projects in Pinkman's "compatible"
  results have licenses verified as OSI-approved --
  zero false positives for license compatibility.
- **SC-003**: Trend analysis results include at least
  three quantitative indicators per project, enabling
  users to compare projects objectively.
- **SC-004**: Dependency health audits correctly identify
  at least 95% of dependencies with available updates
  from a project's manifest file.
- **SC-005**: Recommendation reports contain all required
  sections (license, health, trends, direct dependencies,
  dependency overlap) and can be stored as structured
  artifacts in the knowledge graph.
- **SC-005a**: When multiple projects are scouted in a
  single invocation, 100% of shared dependencies are
  identified and reported with accurate overlap counts.
- **SC-006**: Pinkman produces zero overlap with existing
  hero capabilities -- validated by constitution check
  confirming no functional requirement duplicates an
  existing hero's responsibilities.
<!-- scaffolded by uf vdev -->
