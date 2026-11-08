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
| `cobalt-crush-dev` | `/uf.cobalt-crush` | Developer persona |
| `divisor-*` (11 agents) | `/uf.review-council` | Review council personas |
| `gaze-reporter` | `/gaze` | Quality analysis |
| `gaze-test-generator` | `/gaze-fix` | Test generation |

To invoke a subagent directly, type `@` followed by
the agent name in your message.

## Common Workflows

### Review Code

```
/uf.review-council
```

Runs 9 AI reviewer personas in parallel. Each focuses
on a different aspect (security, architecture, testing,
operations, intent drift). You receive an **APPROVE** or
**REQUEST CHANGES** verdict with specific findings. The
council auto-detects whether to review code or specs
based on what changed on your branch.

The unbound-force repository also runs a repo-local
`divisor-linguist` reviewer for instruction prose, which
`uf init` does not deliver. Run its linter on its own with
`make lint-instructions`.

### Address Review Feedback

After creating a PR and receiving reviewer feedback, use
`/uf.address-feedback` to systematically triage and respond:

```text
/uf.address-feedback [PR_NUMBER]
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
   assessment. For a CRITICAL or HIGH item, or one with no
   severity label, you choose: Accept, Modify, Reject, or Ask.
   For a MEDIUM or LOW item you choose: File issue, Reply, or
   Ask; those never change code in this run
4. **Execute** — Implements code fixes, runs the pre-flight
   checks in hard-gate mode, pushes, files issues, posts reply
   comments, and resolves the addressed threads. A failing
   check gets a fix commit and a re-run, at most 3 attempts in
   all; if it still fails, the command stops and does not push

#### Checking out the PR head

Before it reads any saved progress, the command puts the
working tree on the PR head, so it can run from any branch:

- **Identity is keyed on `HEAD`, not the branch name.** When
  `HEAD` equals the PR's head commit, or descends from it
  (unpushed commits from an earlier run, such as a failed
  push), the command continues and never resets or discards
  local commits. A cross-repo PR may check out under a name
  such as `<owner>/<branch>`; that is accepted.
- **Otherwise it runs `gh pr checkout <N>`**, which resolves
  the head repository (fork or same-repo) and records the
  branch's push remote and merge ref.

The command stops before reading progress, committing, or
pushing when:

- a checkout is needed but the tree has uncommitted changes
  to tracked files (commit or stash them, then rerun)
- `gh pr checkout` fails; the command reports the error. The
  usual cause is an untracked file that collides with a
  tracked file on the PR branch: remove, rename, or stash it
  (`git stash -u` includes untracked files), then rerun
- `HEAD` is still unrelated to the PR head after the checkout
- the branch has no configured push remote or merge ref
- the merge ref is the read-only `refs/pull/<N>/head` (a fork
  PR with maintainer edits disabled), which cannot be pushed
- the merge ref is not `refs/heads/<head branch>`, for
  example a local `main` that already contains the merged PR,
  so the PR's commits are never pushed to another branch
- a branch name or push remote contains characters outside a
  conservative set, before any of them reaches a shell command

**Push target.** The push goes to the remote and merge ref
that `gh pr checkout` configured, not to a hard-coded
`origin`, which in a fork-and-upstream clone is usually the
fork. Before pushing, the command fetches that target and
counts divergence with
`git rev-list --left-right --count HEAD...FETCH_HEAD`; a
remote given as a URL updates only `FETCH_HEAD`, so
`git status` cannot see it. When someone else pushed to the
branch, the command asks whether to rebase onto the remote
or stop with your local commits kept. A failed fetch stops
the push.

**Resume.** Progress is saved per checked-out branch and is
read only after the tree is on the PR head. An interrupted
run therefore resumes on the PR branch, including with
unpushed local commits, whatever branch you started from.

This behavior shipped in PR #955 (issue #954). Issue #965
tracks moving the checkout into a tested tool; the behavior
above stays the same.

#### One council per round

The command does not run `/uf.review-council`. The re-review
that follows (`/uf.review-pr`) is the one council for the
round, and when the push is small and touches only code a
council already reviewed, `/uf.review-pr` runs a light check
instead: it confirms the earlier CRITICAL and HIGH findings
are fixed and can approve. `/uf.review-pr` makes that choice
with the `resolve_review_mode` tool, and any input it cannot
read selects the full council.

Dewey learnings the command stores stay untracked. They reach
`main` through a separate `chore/dewey-learnings-*` PR, never
on the PR branch.

The command is re-entrant: run it multiple times as
reviewers respond. It only shows unresolved threads and
uses a local cache to speed up re-assessment.

### Fix Failing CI Checks

When a PR's CI checks are red, use `/uf.fix-ci` to diagnose
and repair them:

```text
/uf.fix-ci [PR_NUMBER]
```

The command fetches each failed check's logs and classifies
the root cause as exactly one of code failure, stale branch,
conflict, or flaky/infra. It then carries the repair
through:

- **Rebase path** — when the branch is conflicted or behind,
  the head branch is rebased onto the base branch. Rebase
  only, never merge: no option ever merges base into head
- **Gated fix loop** — a human gate approves the diagnosis
  and fix plan before any fix executes; `/uf.review-council`
  must approve the fixes before anything is pushed (capped
  at 3 iterations; exhaustion never pushes)
- **Flaky/infra path** — offers rerunning the failed checks
  instead of a code change, then re-polls CI

The command verifies CI ends green and never merges,
approves, or closes the PR.

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
/uf.finale
```

`/uf.finale` commits, syncs with the base branch, pushes,
creates a PR, watches CI, and returns to `main`. When the
branch is behind the base branch, a confirmation gate offers
to rebase onto it (recommended), continue without rebase, or
abort. A base-state check failure (fetch or count, e.g.
offline) degrades to a warning and the finale proceeds.
A push-state count failure likewise degrades to a warning,
with the push gate selecting a safe push mode.

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
/uf.finale                commit, sync with base, push, PR
```

Optional intermediate steps: `/speckit.clarify` (refine
the spec), `/speckit.analyze` (consistency check),
`/speckit.checklist` (quality validation).

### Go Fully Autonomous

```
/uf.unleash [change description]
```

Pass a description (optionally with an issue reference
like `#123`, which is resolved and threaded through to
the PR); tier routing picks quick, OpenSpec, or Speckit
automatically.

Runs the full pipeline autonomously: clarify, plan,
tasks, spec review, implement, code review, and
retrospective. Works with Speckit (`NNN-*` branches),
OpenSpec (`opsx/*` branches), and the quick tier
(`quick/*` branches); invoked from `main` or an
unrecognized branch, it routes to the right tier
automatically. Exits to human judgment only when it
encounters ambiguity, review failures, or merge
conflicts.

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
| Trivial fix, chore, or docs | Quick | `/uf.unleash` |
| Tactical change (< 3 stories) | OpenSpec | `/uf.unleash` or `/opsx-propose` |
| New feature (3+ stories) | Speckit | `/uf.unleash` or `/speckit.specify` |
| "Handle everything" | Any | `/uf.unleash` (routes automatically) |
| Code review | Standalone | `/uf.review-council` |
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
  ci.md               # CI workflow conventions (tool-owned)
  ci-custom.md        # your CI extensions
  python.md           # Python conventions (tool-owned)
  python-custom.md    # your Python extensions
  review-voice.md     # review comment voice (tool-owned)
  review-voice-custom.md # your review-voice extensions
```

Edit the `*-custom.md` files to add project-specific
rules. Tool-owned files are auto-updated by `uf init`;
custom files are never overwritten.

## CLI Commands

Unbound Force also provides terminal CLI commands
outside of OpenCode:

- `uf init` -- scaffold the framework into your project
- `uf doctor` -- diagnose your development environment
- `uf gate` -- enforce constitution compliance at phase boundaries
- `uf setup` -- install the full toolchain
- `uf sandbox` -- manage containerized dev sessions
- `uf gateway` -- start/stop the LLM reverse proxy
- `uf config` -- manage .uf/config.yaml

See **[CLI Reference](cli-reference.md)** for full
command documentation with flags and examples.



## Quick Reference

| Command | Description |
|---------|-------------|
| `/uf.review-council` | Run the 11-persona review council |
| `/uf.review-pr` | Review a GitHub PR (post-PR) |
| `/uf.address-feedback` | Triage and address PR review feedback |
| `/uf.fix-ci` | Diagnose and fix failing PR CI checks |
| `/uf.rebase-pr` | Rebase a conflicted PR onto its base |
| `/opsx-propose` | Create a change proposal with plan and tasks |
| `/opsx-apply` | Implement tasks from an OpenSpec change |
| `/opsx-explore` | Think through ideas (read-only) |
| `/uf.unleash` | Run the full pipeline autonomously |
| `/speckit.specify` | Create a feature specification |
| `/speckit.plan` | Generate implementation plan from spec |
| `/speckit.tasks` | Break plan into ordered task list |
| `/speckit.implement` | Execute tasks from task list |
| `/speckit.constitution` | Create project constitution |
| `/speckit.analyze` | Analyze spec consistency |
| `/speckit.checklist` | Generate quality checklist |
| `/speckit.taskstoissues` | Convert tasks to GitHub issues |
| `/speckit.testreview` | Review test quality for spec |
| `/uf.cobalt-crush` | Invoke developer persona directly |
| `/gaze` | Run quality analysis (Go projects) |
| `/gaze-fix` | Generate tests for weak spots |
| `/uf.finale` | Commit, sync with base, push, create PR |
| `/uf.agent-brief` | Create or audit AGENTS.md |
| `/org` | Manage work items (cells) |
| `/handoff` | End session with clean handoff |
| `/inbox` | Check agent communication inbox |
| `/forge` | Decompose tasks for parallel execution |
| `/forge-status` | Check parallel execution status |
| `/uf.init` | Run uf init from within OpenCode |
| `/uf.constitution-check` | Check constitution alignment |

## See Also

- **Parallel execution** -- `/forge` to decompose tasks
  and run multiple agents concurrently
- **[AGENTS.md](../AGENTS.md)** -- Full reference for all
  commands, agents, specs, and conventions
