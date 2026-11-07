---
name: unbound-force-heroes
description: "Unbound Force hero roles and routing"
tags:
  - heroes
  - routing
---

# Unbound Force Heroes — Swarm Skills Package

This skill teaches the Swarm coordinator about the Unbound Force hero team, their roles, and how to route natural language queries to the appropriate hero.

## Heroes

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

## Knowledge Retrieval in the Hero Lifecycle

Before delegating to a hero agent, Swarm coordinators SHOULD
query Dewey for relevant context that grounds the hero's work
in project history. This step is optional — if Dewey is
unavailable, skip it and proceed with delegation.

### Per-Stage Dewey Queries

| Stage | Dewey Query | Tool | Purpose |
|-------|------------|------|---------|
| implement | File-specific learnings | `dewey_semantic_search` | Find prior learnings about target files, architectural patterns |
| validate | Quality baselines | `dewey_search` | Find prior quality reports, known CRAP score patterns |
| review | Review history | `dewey_search` | Find recurring review findings, convention violations |

### Graceful Degradation

If Dewey MCP tools are unavailable (return errors or are
not configured), skip the knowledge retrieval step
entirely. The hero agent will still function — it will
simply lack cross-repo context. Each hero agent has its
own 3-tier degradation pattern (Full Dewey → Graph-only
→ No Dewey) documented in its agent file.

## Escalation Rules

1. **Max iterations**: If the review-implement loop exceeds 3 iterations, escalate to manual review with a summary of unresolved findings.
2. **Hero unavailable**: Skip the stage with a warning. The workflow continues with remaining heroes.
3. **Inter-hero contradiction**: Do not auto-resolve. Surface both perspectives to the human operator with supporting data.
4. **All heroes unavailable**: Report "no heroes available" and provide installation guidance.