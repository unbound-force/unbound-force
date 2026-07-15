# Using Unbound Force

## Start OpenCode

Navigate to your project and start OpenCode:

```bash
opencode
```

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
/review-council [N]
```

Runs 9 AI reviewer personas in parallel. Each focuses
on a different aspect (security, architecture, testing,
operations, intent drift). You receive an **APPROVE** or
**REQUEST CHANGES** verdict with specific findings. The
council auto-detects whether to review code or specs
based on what changed on your branch. When a PR number
is provided (`[N]`), findings can optionally be posted
as a GitHub review.

### Address Review Feedback

After creating a PR and receiving reviewer feedback, use
`/address-feedback` to systematically triage and respond:

```text
/address-feedback [PR_NUMBER]
```

The command runs four phases:

1. **Ingest** — Fetches all unresolved review threads from
   GitHub, groups threaded conversations, and identifies
   reviewer roles
2. **Assess** — Classifies each item as data-driven or
   subjective using project conventions. Simple items are
   assessed directly (Tier 1); complex items are escalated
   to Divisor agents (Tier 2)
3. **Triage** — Presents each item one-by-one with the
   assessment. You choose: Accept, Modify, Reject, or Ask
4. **Execute** — Implements code fixes, runs review-council,
   pushes changes, and posts reply comments

The command is re-entrant: run it multiple times as
reviewers respond. It only shows unresolved threads and
uses a local cache to speed up re-assessment.

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
retrospective. Works with both Speckit (`speckit/NNN-*` branches)
and OpenSpec (`opsx/*` branches). Exits to human
judgment only when it encounters ambiguity, review
failures, or merge conflicts.

### Explore Ideas

```
/opsx-explore
```

Enters a thinking mode for exploring ideas,
investigating problems, and clarifying requirements
before committing to a change. Read-only — no file
modifications.

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

### Persistent CDE with DevPod

Set up a persistent development environment using
DevPod and the devcontainer spec:

```bash
uf sandbox init                          # scaffold devcontainer.json
uf sandbox create --backend devpod       # create workspace
uf sandbox start                         # resume workspace
uf sandbox stop                          # pause workspace
uf sandbox destroy                       # remove workspace
```

DevPod workspaces persist across sessions and run
directly in Podman -- no Kubernetes required. The
gateway proxy auto-starts when a cloud LLM provider
is detected, injecting credentials into the workspace.

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
  severity.md         # severity definitions (tool-owned)
  go.md               # Go conventions (tool-owned)
  go-custom.md        # your Go extensions
  content.md          # writing standards (tool-owned)
  content-custom.md   # your content extensions
  typescript.md       # TypeScript conventions (tool-owned)
  typescript-custom.md # your TypeScript extensions
```

Edit the `*-custom.md` files to add project-specific
rules. Tool-owned files are auto-updated by `uf init`;
custom files are never overwritten.

## CLI Commands

Unbound Force also provides terminal CLI commands
outside of OpenCode:

- `uf init` -- scaffold the framework into your project
- `uf doctor` -- diagnose your development environment
- `uf setup` -- install the full toolchain
- `uf sandbox` -- manage containerized dev sessions
- `uf gateway` -- start/stop the LLM reverse proxy
- `uf config` -- manage .uf/config.yaml

See **[CLI Reference](cli-reference.md)** for full
command documentation with flags and examples.

The `mutimind` CLI provides terminal-based backlog
management. See **[CLI Reference](cli-reference.md#mutimind)**
for details.

## Quick Reference

| Command | Description |
|---------|-------------|
| `/review-council [N]` | Run the 9-persona review council; optionally post to PR |
| `/review-pr` | Review a GitHub PR (post-PR) |
| `/address-feedback` | Triage and address PR review feedback |
| `/opsx-propose` | Create a change proposal with plan and tasks |
| `/opsx-apply` | Implement tasks from an OpenSpec change |
| `/opsx-explore` | Think through ideas (read-only) |
| `/unleash` | Run the full pipeline autonomously |
| `/speckit.specify` | Create a feature specification |
| `/speckit.plan` | Generate implementation plan from spec |
| `/speckit.tasks` | Break plan into ordered task list |
| `/speckit.implement` | Execute tasks from task list |
| `/speckit.constitution` | Create project constitution |
| `/speckit.analyze` | Analyze spec consistency |
| `/speckit.checklist` | Generate quality checklist |
| `/speckit.taskstoissues` | Convert tasks to GitHub issues |
| `/speckit.testreview` | Review test quality for spec |
| `/cobalt-crush` | Invoke developer persona directly |
| `/gaze` | Run quality analysis (Go projects) |
| `/gaze-fix` | Generate tests for weak spots |
| `/finale` | Commit, push, create PR, merge |
| `/agent-brief` | Create or audit AGENTS.md |
| `/org` | Manage work items (cells) |
| `/handoff` | End session with clean handoff |
| `/inbox` | Check agent communication inbox |
| `/forge` | Decompose tasks for parallel execution |
| `/forge-status` | Check parallel execution status |
| `/uf-init` | Run uf init from within OpenCode |
| `/constitution-check` | Check constitution alignment |
| `/workflow-start` | Begin hero lifecycle workflow |
| `/workflow-status` | Check workflow state |
| `/workflow-advance` | Advance workflow to next stage |
| `/workflow-list` | List all workflows |
| `/workflow-seed` | Seed workflow with initial data |

## See Also

- **Backlog management** -- `/muti-mind.init` to set up,
  `/muti-mind.backlog-add` to create items,
  `/muti-mind.prioritize` to rank them
- **Workflow orchestration** -- `/workflow-start`,
  `/workflow-status`, `/workflow-advance` for the
  6-stage hero lifecycle
- **Parallel execution** -- `/forge` to decompose tasks
  and run multiple agents concurrently
- **[AGENTS.md](../AGENTS.md)** -- Full reference for all
  commands, agents, specs, and conventions
