---
name: unbound-force-heroes
description: "Unbound Force hero roles and workflow routing"
tags:
  - heroes
  - workflow
  - routing
---

# Unbound Force Heroes — Swarm Skills Package

This skill teaches the Swarm coordinator about the Unbound Force hero team, their roles, and how to route natural language queries to the appropriate hero.

## Heroes

### Muti-Mind (Product Owner)

- **Agent file**: `.opencode/agents/muti-mind-po.md`
- **Role**: Define work, prioritize backlog, accept increments
- **Routing patterns**:
  - "What should we build next?"
  - "Prioritize the backlog"
  - "Accept this feature"
  - "Review the acceptance criteria"
  - "Create a backlog item for..."
  - Any question about product direction, priorities, or acceptance

### Cobalt-Crush (Developer)

- **Agent file**: `.opencode/agents/cobalt-crush-dev.md`
- **Role**: Implement features from specifications
- **Routing patterns**:
  - "Implement the login feature"
  - "Write code for..."
  - "Fix this bug"
  - "Refactor the..."
  - "Execute the tasks from spec..."
  - Any coding, implementation, or technical task

### Gaze (Tester)

- **Binary**: `gaze` (installed via `brew install unbound-force/tap/gaze`)
- **Role**: Validate code quality, produce quality reports
- **Routing patterns**:
  - "Run quality checks"
  - "Check test coverage"
  - "Analyze code quality"
  - "What's the CRAP score?"
  - Any quality validation or testing question

### The Divisor (Reviewer)

- **Agent files**: `.opencode/agents/divisor-*.md` (5 personas: Guard, Architect, Adversary, SRE, Testing)
- **Role**: Review PRs through multi-persona council
- **Routing patterns**:
  - "Review this PR"
  - "Check this code for issues"
  - "Run the review council"
  - "What do the reviewers think?"
  - Any code review or quality gate question

### Mx F (Manager)

- **Agent file**: `.opencode/agents/mx-f-coach.md`
- **Binary**: `mxf` (installed via `brew install unbound-force/tap/unbound`)
- **Role**: Monitor metrics, coach team, manage sprints
- **Routing patterns**:
  - "How are we doing this sprint?"
  - "Show me the dashboard"
  - "What impediments do we have?"
  - "Run a retrospective"
  - Any process, metrics, or coaching question

## Workflow Stages

The hero lifecycle follows 6 stages in sequence:

```
1. define      (Muti-Mind)     → Backlog item + spec          [human]
2. implement   (Cobalt-Crush)  → Code + tests                 [swarm]
3. validate    (Gaze)          → Quality report                [swarm]
4. review      (The Divisor)   → Review verdict                [swarm]
5. accept      (Muti-Mind)     → Acceptance decision           [human]
6. reflect     (Mx F)          → Metrics snapshot + learning   [swarm]
```

### Stage Transitions

- Each stage produces artifacts consumed by the next stage
- Stages can be **skipped** if the hero is unavailable
- The review → implement loop allows up to 3 iterations
- After 3 iterations, the workflow **escalates** to human review
- At swarm→human boundaries, the workflow **pauses** with `awaiting_human` status

### Reflect Stage

The final stage (owned by Mx F) runs autonomously after the human accepts the increment. It produces a richer output than simple metrics collection:

1. **Metrics snapshot**: Velocity, quality trends, CI health — collected via `mxf collect` and `mxf metrics`
2. **Learning feedback**: Cross-hero pattern analysis via `AnalyzeWorkflows` — detects recurring review findings, quality regressions, and convention pack update opportunities across 3+ completed workflows
3. **Retrospective summary**: Incorporates empirical data from:
   - **Quality report** (from validate stage) — Gaze's CRAP scores, coverage data, testability findings
   - **Review verdict** (from review stage) — The Divisor's multi-persona findings, severity distribution, recurring patterns

The reflect stage explicitly **consumes** artifacts produced by the validate and review stages. When sufficient workflow history exists (3+ completed workflows), it triggers `AnalyzeWorkflows` to produce `LearningFeedback` artifacts with actionable recommendations (e.g., convention pack updates for recurring review findings).

## Execution Modes

Each workflow stage has an execution mode that determines who drives it:

| Stage | Mode | Driver |
|-------|------|--------|
| define | `[human]` | Human operator specifies and clarifies the feature |
| implement | `[swarm]` | Swarm runs Cobalt-Crush autonomously |
| validate | `[swarm]` | Swarm runs Gaze autonomously |
| review | `[swarm]` | Swarm runs The Divisor autonomously |
| accept | `[human]` | Human reviews and makes acceptance decision |
| reflect | `[swarm]` | Swarm runs Mx F autonomously |

### Swarm Delegation Pattern

The workflow supports autonomous swarm delegation with human checkpoints:

1. **Human defines** the feature (define stage) and hands off with `/workflow advance`
2. **Swarm runs autonomously** through implement → validate → review
3. **Workflow pauses** at the swarm→human boundary with `awaiting_human` status
4. **Human resumes** with `/workflow advance`, reviews the increment, and accepts
5. **Swarm runs reflect** autonomously after acceptance — collecting metrics, running learning analysis, and producing a retrospective summary

The complete workflow requires exactly **2 human decision points**: one to hand off after define, and one to accept the increment. Everything else runs autonomously.

### Seed Workflow (Autonomous Define)

The define stage can be configured as `[swarm]` so Muti-Mind
autonomously drafts specifications using Dewey context. This
reduces the workflow to a single human checkpoint (accept).

**Seed command**: Start a workflow with one sentence:
```
/workflow seed "add CSV export to the dashboard"
```

The seed command creates a backlog item and starts a workflow
with `define=swarm`. Muti-Mind queries Dewey for cross-repo
context, drafts the spec, and the swarm continues through
implement → validate → review. The human reviews only at accept.

**Configurable define mode**: Use `/workflow start` with the
`--define-mode` flag for explicit control:
```
/workflow start BI-042 --define-mode=swarm
```

**Optional spec review checkpoint**: For high-stakes features,
add `--spec-review` to pause after the spec is drafted:
```
/workflow seed "migrate credentials to OAuth2" --spec-review
```

| Workflow | Human Checkpoints | Commands |
|----------|:-----------------:|---------|
| **Default** (define=human) | 2 | specify + clarify + advance, then accept |
| **Autonomous** (define=swarm) | 1 | seed, then accept |
| **Autonomous + spec review** | 2 | seed, review spec, then accept |

### Legacy Workflows

Workflows created before execution mode support (without `execution_mode` fields) are treated as all-human for backward compatibility. They advance one stage at a time with no automatic checkpoint pausing.

## Knowledge Retrieval in the Hero Lifecycle

Before delegating to a hero agent at each workflow stage,
Swarm coordinators SHOULD query Dewey for relevant context
that grounds the hero's work in project history. This
step is optional — if Dewey is unavailable, skip it and
proceed with delegation.

### Per-Stage Dewey Queries

| Stage | Dewey Query | Tool | Purpose |
|-------|------------|------|---------|
| define | Feature domain context | `dewey_semantic_search` | Find related specs, prior features, backlog patterns |
| implement | File-specific learnings | `dewey_semantic_search` | Find prior learnings about target files, architectural patterns |
| validate | Quality baselines | `dewey_search` | Find prior quality reports, known CRAP score patterns |
| review | Review history | `dewey_search` | Find recurring review findings, convention violations |
| accept | Acceptance precedents | `dewey_search` | Find prior acceptance decisions for similar features |
| reflect | Process patterns | `dewey_semantic_search` | Find velocity trends, retrospective outcomes |

### Graceful Degradation

If Dewey MCP tools are unavailable (return errors or are
not configured), skip the knowledge retrieval step
entirely. The hero agent will still function — it will
simply lack cross-repo context. Each hero agent has its
own 3-tier degradation pattern (Full Dewey → Graph-only
→ No Dewey) documented in its agent file.

## Escalation Rules

1. **Max iterations**: If the review-implement loop exceeds 3 iterations, escalate to manual review with a summary of unresolved findings.
2. **Acceptance rejection**: If Muti-Mind rejects the increment, create a new backlog item with the rejection rationale.
3. **Hero unavailable**: Skip the stage with a warning. The workflow continues with remaining heroes.
4. **Inter-hero contradiction**: Do not auto-resolve. Surface both perspectives to the human operator with supporting data.
5. **All heroes unavailable**: Report "no heroes available" and provide installation guidance.

## Workflow Commands

- `/workflow start [backlog-item-id] [--define-mode=human|swarm] [--spec-review]` — Begin a new hero lifecycle workflow
- `/workflow seed <description> [--spec-review]` — Seed a new feature with one sentence (define=swarm)
- `/workflow status [workflow-id]` — Check current workflow state
- `/workflow list [--status active|completed|all]` — List all workflows
- `/workflow advance` — Advance to the next stage
