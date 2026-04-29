# Using Unbound Force

## Start OpenCode

Navigate to your project and start OpenCode:

```bash
opencode
```

## OpenPackage Installation

You can install the review council and workflow command bundles
without the `uf` binary using OpenPackage (`opkg`):

**Review council** — nine reviewer personas plus `/review-council`
and `/review-pr`, convention packs, and optional Dewey MCP config:

```bash
opkg install @unbound-force/review-council
```

**Workflows** — Speckit pipeline commands, OpenSpec commands, and
`/constitution-check`; depends on `@unbound-force/review-council`:

```bash
opkg install @unbound-force/workflows
```

Regenerated assets live under `packages/` in this repository; maintainers run
`make packages` after changing `.opencode/` sources.

## Modes and Agents

OpenCode has two interaction layers: **primary modes**
you switch between, and **subagents** invoked by
commands.

### Primary Modes (Tab to switch)

| Mode | Purpose |
|------|---------|
| **Build** | Makes changes -- the default mode for development |
| **Plan** | Read-only analysis and planning -- no file modifications |

Press **Tab** to cycle between modes. Use Plan to think
through an approach, then switch to Build to execute.

### Subagents (invoked via slash commands)

Subagents are specialized agents invoked automatically
by slash commands. You rarely need to call them directly.

| Agent | Invoked by | Role |
|-------|-----------|------|
| `cobalt-crush-dev` | `/cobalt-crush` | Developer persona |
| `divisor-*` (9 agents) | `/review-council` | Review council personas |
| `muti-mind-po` | `/muti-mind.*` | Product owner |
| `mx-f-coach` | `@mx-f-coach` | Coaching and retrospectives |
| `gaze-reporter` | `/gaze` | Quality analysis |
| `gaze-test-generator` | `/gaze-fix` | Test generation |

To invoke a subagent directly, type `@` followed by
the agent name in your message.

## Common Workflows

### Review Code

```
/review-council
```

Runs 5+ AI reviewer personas in parallel. Each focuses
on a different aspect (security, architecture, testing,
operations, intent drift). You receive an **APPROVE** or
**REQUEST CHANGES** verdict with specific findings. The
council auto-detects whether to review code or specs
based on what changed on your branch.

### Propose a Change (Small)

For bug fixes, minor enhancements, and tasks under 3
user stories:

```
/opsx-propose <describe what you want to change>
```

This creates a proposal, design, and task list in one
step. Then implement and finalize:

```
/opsx-apply
/finale
```

`/finale` commits, pushes, creates a PR, watches CI, and
merges.

### Build a Feature (Large)

For features with 3+ user stories, use the Speckit
pipeline:

```
/speckit.specify <describe the feature>
        |
        v
/speckit.plan          generate implementation plan
        |
        v
/speckit.tasks         break into ordered task list
        |
        v
/speckit.implement     execute the tasks
        |
        v
/finale                commit, push, PR, merge
```

Optional intermediate steps: `/speckit.clarify` (refine
the spec), `/speckit.analyze` (consistency check),
`/speckit.checklist` (quality validation).

### Go Fully Autonomous

```
/unleash
```

Runs the full pipeline autonomously: clarify, plan,
tasks, spec review, implement, code review, and
retrospective. Works with both Speckit (`NNN-*` branches)
and OpenSpec (`opsx/*` branches). Exits to human
judgment only when it encounters ambiguity, review
failures, or merge conflicts.

### Check Code Quality (Go Projects)

```
/gaze
```

Produces CRAP scores, coverage metrics, side effect
classifications, and overall project health. Then
generate tests for the weakest spots:

```
/gaze-fix
```

## When to Use What

| Situation | Workflow | Start with |
|-----------|----------|------------|
| Bug fix or small task | OpenSpec | `/opsx-propose` |
| New feature (3+ stories) | Speckit | `/speckit.specify` |
| "Handle everything" | Either | `/unleash` |
| Code review | Standalone | `/review-council` |
| Quality check (Go) | Standalone | `/gaze` |

## Customization

Convention packs define coding standards that review
agents enforce. After `uf init`, find them at:

```
.opencode/uf/packs/
  default.md          # language-agnostic (tool-owned)
  default-custom.md   # your project extensions
  go.md               # Go conventions (tool-owned)
  go-custom.md        # your Go extensions
  content.md          # writing standards (tool-owned)
  content-custom.md   # your content extensions
```

Edit the `*-custom.md` files to add project-specific
rules. Tool-owned files are auto-updated by `uf init`;
custom files are never overwritten.

## Quick Reference

| Command | Description |
|---------|-------------|
| `/review-council` | Run the 5-persona review council |
| `/opsx-propose` | Create a change proposal with plan and tasks |
| `/opsx-apply` | Implement tasks from an OpenSpec change |
| `/unleash` | Run the full pipeline autonomously |
| `/speckit.specify` | Create a feature specification |
| `/speckit.plan` | Generate implementation plan from spec |
| `/speckit.tasks` | Break plan into ordered task list |
| `/speckit.implement` | Execute tasks from task list |
| `/cobalt-crush` | Invoke developer persona directly |
| `/gaze` | Run quality analysis (Go projects) |
| `/gaze-fix` | Generate tests for weak spots |
| `/finale` | Commit, push, create PR, merge |
| `/opsx-explore` | Think through ideas (read-only) |

## See Also

- **Backlog management** -- `/muti-mind.init` to set up,
  `/muti-mind.backlog-add` to create items,
  `/muti-mind.prioritize` to rank them
- **Workflow orchestration** -- `/workflow-start`,
  `/workflow-status`, `/workflow-advance` for the
  6-stage hero lifecycle
- **Parallel execution** -- `/forge` to decompose tasks
  and run multiple agents concurrently
- **[AGENTS.md](AGENTS.md)** -- Full reference for all
  commands, agents, specs, and conventions
