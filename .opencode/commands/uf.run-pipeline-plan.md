---
description: >
  Execute a prioritized work plan through the full Unbound
  Force pipeline (/opsx-propose -> /uf.unleash -> /uf.finale)
  with fresh subagent isolation per phase, human gate relay,
  and per-item progress tracking.
---
<!-- scaffolded by uf vdev -->

# Command: /uf.run-pipeline-plan

## User Input

```text
$ARGUMENTS
```

## Description

Orchestrate a prioritized work plan through the full
Unbound Force pipeline. For each work item, spawns three
fresh `general` subagents in sequence: `/opsx-propose`
(spec artifacts), `/uf.unleash` (implementation + review),
`/uf.finale` (commit + PR). Human confirmation gates are
relayed through the orchestrator's `question` tool.
Progress is tracked via `todowrite`.

Supports two input modes: a plan file path or space-
separated GitHub issue references (`#NNN` or URLs).

## Usage

```
/uf.run-pipeline-plan sprint-42.md
/uf.run-pipeline-plan #123 #456 #789
/uf.run-pipeline-plan https://github.com/unbound-force/unbound-force/issues/42 #99
```

## Instructions

<protect>

> **SESSION-RESUME GUARD**: If you are resuming this
> command after context compression or a session restart,
> STOP and re-read this entire template before continuing.
> Do NOT infer which items are complete from compressed
> context summaries. Only `todowrite` entries and disk
> state (branches, spec artifacts, PRs) are authoritative
> for resumability.

### Execution Checklist

Maintain this checklist using the `todowrite` tool. One
entry per work item. Mark complete only after the item's
PR is created AND CI is green.

### Phase 0 — Input Resolution and Confirmation

Phase 0 runs in the orchestrator context (no subagents).

1. **Detect input type**: Check if `$ARGUMENTS` is an
   existing file path.

   - **File path**: Parse as a plan file (see Plan File
     Contract below).
   - **Otherwise**: Treat as space-separated GitHub issue
     references (`#NNN` or full GitHub URLs).

2. **Resolve issue metadata** (issue-ref input only):
   For each issue reference, run:

   ```bash
   gh issue view <N> --json title,body,number
   ```

   Derive a kebab-case change name from the issue title
   and a one-line description from the issue body.

3. **Parse plan file** (file input only): Read the plan
   file line by line. Each numbered item MUST carry:
   - One or more issue references (`#NNN`)
   - A kebab-case `Change:` name
   - An optional one-line description
   - Non-mechanical items tagged `[INVESTIGATION]` or
     `[POLICY]`

4. **Flag non-mechanical items**: Present all items
   tagged `[INVESTIGATION]` or `[POLICY]` to the human
   via `question`. The human MUST explicitly resolve
   each flagged item (remove, convert to mechanical,
   or defer) before the pipeline runs. The pipeline
   MUST NOT auto-execute flagged items.

5. **Present work-item table**: Show the full ordered
   list of work items to the human via `question` for
   confirmation. Include issue refs, change names, and
   descriptions. The human MUST confirm before the
   pipeline proceeds.

   For a bare issue list (no grouping information),
   ask via `question` how to group issues into work
   items (default: one PR per issue).

### Plan File Contract

A plan file is a numbered markdown list. Each item
carries:

```
1. #42 Change: add-run-plan-command — Create /uf.run-pipeline-plan slash command
2. #99 Change: fix-scaffold-drift — Sync embedded asset copies
3. #55 [INVESTIGATION] Investigate CI flakiness in TestEmbeddedAssets
```

- Issue refs: `#NNN` (one or more per item)
- Change name: `Change: <kebab-case-name>`
- Description: text after the `—` separator (optional)
- Non-mechanical tag: `[INVESTIGATION]` or `[POLICY]`

### Per-Item Pipeline Protocol

For each confirmed work item, execute three pipeline
phases sequentially, each in a fresh `general` subagent:

1. **Subagent 1**: `/opsx-propose` with input = change
   name + description. Produces spec artifacts
   (proposal.md, design.md, specs/, tasks.md) on an
   `opsx/<change-name>` branch.

2. **Subagent 2**: `/uf.unleash` (OpenSpec mode).
   Implements tasks, runs spec review, code review,
   retrospective, and demo. Step 8 (code review) MUST
   NOT be skipped.

3. **Subagent 3**: `/uf.finale`. Commits changes,
   pushes, creates PR with `Closes #N` in the commit
   message and PR body. PR targets parent repo
   `unbound-force/unbound-force` with head
   `jflowers:<branch>`. PR stays open for human review.

**Disk handoff verification**: Before spawning each
next subagent, verify that the previous subagent's
outputs exist on disk:

- After `/opsx-propose`: verify spec artifacts exist
  (proposal.md, design.md, specs/, tasks.md) and the
  correct `opsx/<name>` branch is checked out.
- After `/uf.unleash`: verify implementation files
  exist and tests pass.
- After `/uf.finale`: verify PR URL is returned.

**Clean tree between items**: After each item completes
(PR created), confirm return to `main` with a clean
working tree (`git status --short` is empty) before
starting the next item.

### Gate Relay Protocol

When a subagent encounters a mandatory human
confirmation gate, it MUST STOP execution and return
a structured value to the orchestrator:

```
GATE:<name>
<exact question text>
<full proposed text / findings verbatim>
<options list>
```

The orchestrator then:
1. Receives the `GATE:` return value
2. Invokes the `question` tool with the exact question,
   full text, and options
3. After the human responds, resumes the SAME subagent
   via `task_id` with the human's answer

**CRITICAL**: Subagents MUST NOT invoke the `question`
tool directly. All human interaction flows through the
orchestrator. This prevents subagents from racing for
user input and ensures consistent gate formatting.

### Guardrails

These rules are non-negotiable:

- **Scaffold drift sync**: Any file under
  `.opencode/commands/` or `openspec/` that changes
  MUST have its `internal/scaffold/assets/...` copy
  synced byte-identical, or
  `TestEmbeddedAssets_MatchSource` fails.
- **Explicit staging**: Stage files explicitly by path.
  `git add -A` and `git add .` are forbidden.
- **Stray file removal**: Remove any untracked
  `internal/scaffold/AGENTS.md` before staging. Never
  commit it.
- **No force-push**: `git push --force` is forbidden.
  Only `git push --force-with-lease` is allowed.
- **PR targeting**: PRs MUST target the parent
  `unbound-force/unbound-force` repo with head
  `jflowers:<branch>`.
- **Closes linkage**: The commit message and PR body
  MUST include `Closes #N` for each issue in the work
  item.
- **No auto-merge**: The PR MUST NOT be merged
  automatically. It stays open for human review.

### Tracking

Maintain `todowrite` with one entry per work item:

```
- [ ] Item 1: #42 add-run-plan-command
- [ ] Item 2: #99 fix-scaffold-drift
- [ ] Item 3: #55 [INVESTIGATION] CI flakiness
```

Mark an entry complete (`[x]`) ONLY after:
1. The item's PR has been created
2. CI checks on the PR are green

Do NOT mark entries complete prematurely. The
`todowrite` tracker is the source of truth for item
completion, not the orchestrator's context.

</protect>
