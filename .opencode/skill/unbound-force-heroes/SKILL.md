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
1. define      (Muti-Mind)     → Backlog item + spec
2. implement   (Cobalt-Crush)  → Code + tests
3. validate    (Gaze)          → Quality report
4. review      (The Divisor)   → Review verdict
5. accept      (Muti-Mind)     → Acceptance decision
6. measure     (Mx F)          → Metrics snapshot
```

### Stage Transitions

- Each stage produces artifacts consumed by the next stage
- Stages can be **skipped** if the hero is unavailable
- The review → implement loop allows up to 3 iterations
- After 3 iterations, the workflow **escalates** to human review

## Escalation Rules

1. **Max iterations**: If the review-implement loop exceeds 3 iterations, escalate to manual review with a summary of unresolved findings.
2. **Acceptance rejection**: If Muti-Mind rejects the increment, create a new backlog item with the rejection rationale.
3. **Hero unavailable**: Skip the stage with a warning. The workflow continues with remaining heroes.
4. **Inter-hero contradiction**: Do not auto-resolve. Surface both perspectives to the human operator with supporting data.
5. **All heroes unavailable**: Report "no heroes available" and provide installation guidance.

## Workflow Commands

- `/workflow start [backlog-item-id]` — Begin a new hero lifecycle workflow
- `/workflow status [workflow-id]` — Check current workflow state
- `/workflow list [--status active|completed|all]` — List all workflows
- `/workflow advance` — Advance to the next stage
