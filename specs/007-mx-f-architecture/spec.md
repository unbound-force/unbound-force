# Feature Specification: Mx F Architecture (Manager)

**Feature Branch**: `007-mx-f-architecture`
**Created**: 2026-02-24
**Status**: Draft
**Input**: User description: "Design the architecture for Mx F (Mx Found), the Manager hero. Mx F is the Flow Facilitator and Continuous Improvement Coach. It includes an AI coaching persona, a full metrics collection and dashboard platform, retrospective facilitation, obstacle tracking, and swarm coordination capabilities."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Metrics Collection Platform (Priority: P1)

A team lead or process owner deploys Mx F to collect and aggregate development metrics from multiple data sources: GitHub (PRs, issues, commits, CI runs), Gaze quality reports, Divisor review verdicts, Muti-Mind backlog state, and speckit artifacts. The collected metrics are stored locally and queryable via CLI.

**Why this priority**: P1 because metrics are the foundation of everything Mx F does. Without data, coaching is opinion, retrospectives are anecdotes, and process improvement is guesswork. The metrics platform must be built first.

**Independent Test**: Can be tested by configuring Mx F to collect metrics from a GitHub repository with at least 10 PRs, running the collection, and verifying the stored data matches the GitHub API output.

**Acceptance Scenarios**:

1. **Given** a configured GitHub repository, **When** `mx-f collect --source github` is run, **Then** Mx F collects: PR count, PR merge time, review turnaround time, CI pass rate, commit frequency, issue open/close rates, and contributor activity.
2. **Given** Gaze quality report artifacts in the project, **When** `mx-f collect --source gaze` is run, **Then** Mx F collects: average CRAP scores, CRAPload counts, contract coverage trends, and over-specification counts over time.
3. **Given** Divisor review-verdict artifacts, **When** `mx-f collect --source divisor` is run, **Then** Mx F collects: review iteration counts, finding categories and frequencies, approval rates, and time-to-approval.
4. **Given** Muti-Mind backlog-item artifacts, **When** `mx-f collect --source muti-mind` is run, **Then** Mx F collects: backlog size, velocity (items completed per sprint), lead time (backlog to done), and acceptance rates.
5. **Given** all sources collected, **When** `mx-f metrics summary` is run, **Then** it produces a consolidated snapshot: velocity, cycle time, defect rate, test quality trend, review efficiency, and overall flow health.

---

### User Story 2 - CLI Metrics Querying and Reporting (Priority: P1)

A team lead queries historical metrics via the Mx F CLI to understand trends, identify bottlenecks, and prepare for retrospectives. Reports are available in both human-readable text and machine-parseable JSON formats.

**Why this priority**: P1 because collection without querying is useless. The ability to ask "How are we doing?" and get an answer is the minimum viable product for Mx F.

**Independent Test**: Can be tested by collecting metrics over a simulated period and running various query commands, verifying the output matches the expected calculations.

**Acceptance Scenarios**:

1. **Given** collected metrics spanning 4 sprints, **When** `mx-f metrics velocity --sprints 4` is run, **Then** it reports velocity for each sprint and the trend (increasing/stable/decreasing).
2. **Given** collected metrics, **When** `mx-f metrics cycle-time --period 30d` is run, **Then** it reports average, median, P90, and P99 cycle times for items completed in the last 30 days.
3. **Given** collected metrics, **When** `mx-f metrics bottlenecks` is run, **Then** it identifies the stage with the longest average wait time (e.g., "Review is the bottleneck: average 2.3 days wait, 4x longer than implementation").
4. **Given** any metrics query, **When** `--format json` is appended, **Then** the output is a JSON object conforming to the `metrics-snapshot` artifact type (Spec 002 artifact envelope).
5. **Given** collected metrics, **When** `mx-f metrics health` is run, **Then** it produces a health dashboard with traffic-light indicators (green/yellow/red) for: velocity, quality, review efficiency, and backlog health.

---

### User Story 3 - AI Coaching and Retrospective Facilitation (Priority: P2)

Mx F provides an AI coaching persona that facilitates retrospectives, root cause analysis, and process improvement conversations. When the team encounters a blocker or failure, Mx F does not prescribe solutions but uses reflective questions (5 Whys, mirroring, probing) to help the team discover the root cause and devise their own path forward.

**Why this priority**: P2 because coaching is Mx F's differentiating capability — what separates it from a dashboard. It depends on the metrics platform (US1) being functional so coaching is grounded in data, not opinion.

**Independent Test**: Can be tested by presenting Mx F with a team scenario (e.g., "Our last three PRs all required 3+ review iterations") and verifying Mx F produces coaching questions that lead toward root cause identification.

**Acceptance Scenarios**:

1. **Given** a team scenario describing a recurring problem, **When** Mx F's coaching agent is consulted, **Then** it produces a series of reflective questions (not solutions) that guide the team toward root cause identification. The questions follow the 5 Whys framework.
2. **Given** metrics showing review iterations have increased over 3 sprints, **When** Mx F facilitates a retrospective, **Then** it presents the data, asks "What changed 3 sprints ago?", and guides the discussion through root cause analysis to an actionable improvement proposal.
3. **Given** a retrospective session, **When** Mx F facilitates it, **Then** it follows a structured format: (1) Data presentation, (2) Pattern identification, (3) Root cause analysis, (4) Improvement proposals, (5) Action items with owners and deadlines.
4. **Given** an action item from a previous retrospective, **When** the next retrospective begins, **Then** Mx F reviews the previous action items and reports their status (completed/in-progress/stale).
5. **Given** a team member asks Mx F "What should we do about slow reviews?", **When** Mx F responds, **Then** it does NOT prescribe a solution. Instead, it asks: "What do you think is causing the slowdown?" and "What have you tried so far?"

---

### User Story 4 - Obstacle Tracking and Impediment Management (Priority: P2)

Mx F tracks impediments that block the team's flow. Impediments are logged with severity, assigned owners, and tracked to resolution. Mx F proactively identifies potential impediments from metrics trends (e.g., CI failure rate spike) before they become blocking.

**Why this priority**: P2 because obstacle removal is a core Scrum Master responsibility. Without tracking, impediments are addressed ad-hoc and recur.

**Independent Test**: Can be tested by logging several impediments, assigning owners, resolving some, and verifying the impediment report accurately reflects the current state.

**Acceptance Scenarios**:

1. **Given** the Mx F CLI, **When** `mx-f impediment add --title "CI pipeline flaky on Linux" --severity high --owner "@dev"` is run, **Then** an impediment is logged with a unique ID, timestamp, and the specified attributes.
2. **Given** logged impediments, **When** `mx-f impediment list` is run, **Then** active impediments are displayed sorted by severity (critical first) with: ID, title, severity, owner, age (days open), and status.
3. **Given** an impediment resolved, **When** `mx-f impediment resolve IMP-003 --resolution "Pinned CI base image to specific version"` is run, **Then** the impediment is marked resolved with the resolution and resolution timestamp.
4. **Given** metrics showing CI failure rate increased from 5% to 25% over 7 days, **When** `mx-f impediment detect` is run, **Then** Mx F creates a draft impediment flagging the CI regression with the supporting data.
5. **Given** an impediment that has been open for more than 14 days, **When** `mx-f impediment list` is run, **Then** it is highlighted as "stale" and Mx F recommends escalation.

---

### User Story 5 - Dashboard and Trend Visualization (Priority: P3)

Mx F provides trend visualizations for key metrics. In the CLI, trends are rendered as text-based charts (ASCII sparklines, bar charts). Optionally, Mx F can generate a static HTML dashboard for sharing with stakeholders.

**Why this priority**: P3 because data visualization is valuable but not essential — tabular metrics reporting (US2) is the MVP. Visualization enhances comprehension for retrospectives and stakeholder communication.

**Independent Test**: Can be tested by collecting metrics over a simulated period and generating a trend chart, verifying the chart accurately represents the data.

**Acceptance Scenarios**:

1. **Given** 4 sprints of velocity data, **When** `mx-f dashboard velocity` is run, **Then** it displays a text-based bar chart showing velocity per sprint with trend line.
2. **Given** 30 days of cycle time data, **When** `mx-f dashboard cycle-time` is run, **Then** it displays a sparkline showing daily average cycle time with min/max/mean annotations.
3. **Given** the `--html` flag, **When** `mx-f dashboard --html --output report.html` is run, **Then** a standalone HTML file is generated with interactive charts using a lightweight charting library (no server required).
4. **Given** a health dashboard, **When** rendered, **Then** it shows traffic-light indicators alongside trend sparklines for each health dimension.

---

### User Story 6 - Swarm Coordination and Process Stewardship (Priority: P3)

Mx F coordinates the swarm's overall process: sprint ceremonies, cross-hero communication, and capacity management. It monitors the flow across the entire pipeline (Muti-Mind -> Cobalt-Crush -> Gaze -> Divisor) and identifies process breakdowns.

**Why this priority**: P3 because swarm coordination is the ultimate vision for Mx F but depends on all other heroes being functional. It is the most complex capability and requires the metrics platform, coaching engine, and obstacle tracking to be in place.

**Independent Test**: Can be tested by simulating a sprint lifecycle (planning, daily standups, review, retrospective) with Mx F facilitating each ceremony and verifying appropriate outputs at each stage.

**Acceptance Scenarios**:

1. **Given** Mx F is configured for a team, **When** `mx-f sprint plan --goal "Implement user authentication"` is run, **Then** it pulls the prioritized backlog from Muti-Mind, calculates team capacity from historical velocity, and suggests a sprint scope.
2. **Given** an active sprint, **When** `mx-f standup` is run, **Then** it reports: items in progress, items blocked (from impediment tracker), CI/test status (from Gaze), review status (from Divisor), and flags any items at risk of missing the sprint goal.
3. **Given** a sprint complete, **When** `mx-f sprint review` is run, **Then** it summarizes: items completed vs. planned, velocity, quality metrics, and acceptance decisions from Muti-Mind.
4. **Given** metrics from The Divisor showing the Architect persona frequently requests "add documentation" findings, **When** Mx F analyzes the pattern, **Then** it recommends a process change: update the Cobalt-Crush convention pack to include mandatory documentation requirements.

---

### Edge Cases

- What happens when no data sources are configured? `mx-f collect` MUST report which sources are unavailable and which are configured, collecting from whatever is available (graceful degradation).
- What happens when GitHub API rate limiting is hit during collection? The collector MUST detect rate limits, report them, and resume collection after the reset window (or use cached data with a staleness warning).
- What happens when no historical metrics exist for trend analysis? Mx F MUST report "insufficient data" and specify how many data points are needed before trends become available.
- What happens when an impediment is logged without an owner? Mx F MUST accept it but flag it as "unassigned" in the impediment list.
- What happens when the coaching agent is asked a technical question (e.g., "How do I fix this bug?")? Mx F MUST redirect: "That's a question for Cobalt-Crush. My focus is process and flow. Can I help you identify what's blocking your progress on this?"
- What happens when metrics from different heroes have inconsistent timestamps? Mx F MUST normalize all timestamps to UTC and report any data source whose clock drift exceeds 5 minutes.
- What happens when two teams use Mx F in the same repository? Mx F MUST support team-scoped metrics via a `--team` flag or configuration. Default is the entire repository.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Mx F MUST provide a metrics collection platform that ingests data from: GitHub API, Gaze quality reports, Divisor review verdicts, Muti-Mind backlog artifacts, and speckit artifacts.
- **FR-002**: Mx F MUST provide a CLI tool (`mx-f`) with subcommands: `collect`, `metrics`, `impediment`, `dashboard`, `sprint`, `standup`, and `retro`.
- **FR-003**: The metrics collection MUST store data locally in a structured format (JSON files or SQLite) in `.mx-f/data/`.
- **FR-004**: Mx F MUST compute at minimum these metrics: velocity, cycle time, lead time, defect rate, review iteration count, CI pass rate, backlog health, and flow efficiency.
- **FR-005**: All metrics queries MUST support `--format json` and `--format text` output (per Org Constitution Principle III).
- **FR-006**: Mx F MUST produce a `metrics-snapshot` artifact type conforming to the inter-hero artifact envelope (Spec 002).
- **FR-007**: Mx F MUST provide an AI coaching persona deployed as an OpenCode agent (`mx-f-coach.md`) installable via `mx-f init`.
- **FR-008**: The coaching persona MUST use reflective questioning techniques (5 Whys, mirroring, probing) instead of prescribing solutions.
- **FR-009**: Mx F MUST facilitate structured retrospectives with: data presentation, pattern identification, root cause analysis, improvement proposals, and action items.
- **FR-010**: Retrospective action items MUST be tracked and reviewed at the start of the next retrospective.
- **FR-011**: Mx F MUST provide impediment tracking: add, list, resolve, detect (from metrics), with severity, owner, and age tracking.
- **FR-012**: Mx F MUST proactively detect potential impediments from metrics trends (e.g., CI failure rate spike, review turnaround increase).
- **FR-013**: Mx F MUST provide text-based trend visualizations (ASCII sparklines, bar charts) for key metrics.
- **FR-014**: Mx F SHOULD provide optional HTML dashboard generation for stakeholder communication.
- **FR-015**: Mx F MUST support sprint lifecycle management: planning (with capacity calculation), daily standup reporting, sprint review summaries.
- **FR-016**: Mx F MUST identify process patterns from cross-hero data (e.g., frequent Divisor findings -> convention pack improvement recommendation).
- **FR-017**: Mx F MUST produce a `coaching-record` artifact type for retrospective outcomes and process improvement records.
- **FR-018**: Mx F MUST conform to the Hero Interface Contract (Spec 002): standard repo structure, hero manifest, speckit integration, OpenCode agent/command standards.
- **FR-019**: Mx F MUST consume artifacts from all other heroes: `quality-report` (Gaze), `review-verdict` (Divisor), `backlog-item` and `acceptance-decision` (Muti-Mind).
- **FR-020**: Mx F MUST support graceful degradation: if a data source is unavailable, it collects from available sources and reports what is missing.
- **FR-021**: Mx F MUST NOT prescribe technical solutions. When asked technical questions, it MUST redirect to the appropriate hero (Cobalt-Crush for coding, Gaze for testing, Divisor for architecture).

### Key Entities

- **Metrics Snapshot**: A point-in-time collection of all metrics. Attributes: timestamp, velocity (items/sprint), cycle_time (hours, avg/median/p90/p99), lead_time (hours), defect_rate (defects/item), review_iterations (avg), ci_pass_rate (%), backlog_health (items total, items ready, items stale), flow_efficiency (%).
- **Health Indicator**: A traffic-light assessment of a metric dimension. Attributes: dimension (velocity/quality/review/backlog/flow), status (green/yellow/red), value, threshold_green, threshold_yellow, trend (improving/stable/declining).
- **Impediment**: A tracked blocker. Attributes: id (IMP-NNN), title, description, severity (critical/high/medium/low), owner, status (open/in-progress/resolved/escalated), created_at, resolved_at, resolution, age_days, source (manual/detected).
- **Retrospective Record**: A structured record of a retrospective session. Attributes: date, participants[], data_presented{}, patterns_identified[], root_causes[], improvement_proposals[], action_items[].
- **Action Item**: A tracked improvement commitment. Attributes: id (AI-NNN), description, owner, deadline, status (pending/in-progress/completed/stale), retrospective_id.
- **Coaching Interaction**: A record of a coaching session. Attributes: topic, questions_asked[], insights_surfaced[], outcome (action_item/escalation/resolved/deferred), timestamp.
- **Sprint State**: Tracking of a sprint lifecycle. Attributes: sprint_name, goal, start_date, end_date, planned_items[], completed_items[], velocity, health_indicators[].

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Metrics collection from GitHub API retrieves all required data points (PR count, merge time, review turnaround, CI pass rate, issue rates) for a repository with 10+ PRs.
- **SC-002**: Metrics collection from Gaze, Divisor, and Muti-Mind artifacts produces valid data when artifacts exist and gracefully reports "no data" when they don't.
- **SC-003**: `mx-f metrics summary` produces a consolidated health dashboard with traffic-light indicators within 5 seconds of query.
- **SC-004**: `mx-f metrics bottlenecks` correctly identifies the slowest stage in a simulated pipeline with known bottleneck placement.
- **SC-005**: The coaching persona produces reflective questions (not solutions) in response to 5 sample team scenarios.
- **SC-006**: Retrospective facilitation produces a structured record with all five sections (data, patterns, root causes, proposals, action items).
- **SC-007**: Impediment tracking round-trips: add, list, resolve, verify resolution recorded.
- **SC-008**: Proactive impediment detection identifies a CI failure rate spike from simulated metrics data.
- **SC-009**: All CLI output supports `--format json` and the JSON validates against the `metrics-snapshot` artifact type schema.
- **SC-010**: Mx F functions with only GitHub as a data source when other heroes are not deployed (standalone capability per Principle II).

## Dependencies

### Prerequisites

- **Spec 001** (Org Constitution): Mx F must align with org principles.
- **Spec 002** (Hero Interface Contract): Mx F must conform to the hero manifest, artifact envelope, and naming conventions.

### Downstream Dependents

- **Spec 008** (Swarm Orchestration): Mx F is the process coordinator in the swarm workflow.
- **Spec 009** (Shared Data Model): Defines the `metrics-snapshot` and `coaching-record` JSON schemas.

### Data Sources (Artifact Consumers)

- **Gaze**: Consumes `quality-report` artifacts for quality metrics.
- **The Divisor**: Consumes `review-verdict` artifacts for review efficiency metrics.
- **Muti-Mind**: Consumes `backlog-item` and `acceptance-decision` artifacts for velocity and acceptance metrics.
- **GitHub API**: Consumes PR, Issue, CI, and commit data for development flow metrics.

```
Data Sources → Mx F Metrics Platform → Outputs
┌──────────────┐
│ GitHub API   │──┐
│ (PRs, CI,    │  │
│  Issues)     │  │
└──────────────┘  │  ┌────────────────┐    ┌─────────────────┐
┌──────────────┐  │  │ Mx F Metrics   │    │ Outputs:        │
│ Gaze         │──┼─>│ Collection     │───>│ - CLI queries   │
│ (quality     │  │  │ & Storage      │    │ - JSON artifacts│
│  reports)    │  │  │ (.mx-f/data/)  │    │ - Text charts   │
└──────────────┘  │  └───────┬────────┘    │ - HTML dashboard│
┌──────────────┐  │          │             │ - Coaching       │
│ The Divisor  │──┤          ▼             │ - Retros        │
│ (review      │  │  ┌────────────────┐    │ - Impediments   │
│  verdicts)   │  │  │ Analysis &     │───>│ - Sprint mgmt   │
└──────────────┘  │  │ Coaching       │    └─────────────────┘
┌──────────────┐  │  │ Engine         │
│ Muti-Mind    │──┘  └────────────────┘
│ (backlog,    │
│  acceptance) │
└──────────────┘
```
