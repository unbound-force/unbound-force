---
name: workflow-status
description: Check the current workflow state
---

# /workflow status

Check the current workflow state for the active branch.

## Usage

```
/workflow status [workflow-id]
```

If no workflow-id is provided, shows the most recent in-progress workflow for the current branch (active or awaiting_human).

## Behavior

When this command is invoked:

1. **If a workflow-id is provided**, read `.unbound-force/workflows/{workflow-id}.json` directly.

2. **If no workflow-id is provided**:
   - Detect the current git branch via `git branch --show-current`
   - Read all JSON files from `.unbound-force/workflows/`
   - Find the most recent in-progress workflow (active or awaiting_human) matching the current branch

3. **Parse the workflow JSON** and display:
   - Workflow ID, branch, backlog item, status, start time
   - Iteration count (for review loops)
   - Each stage with status indicator and execution mode:
     - ✓ = completed
     - ◉ = active
     - ○ = pending
     - ⏸ = pending at human checkpoint (awaiting_human status)
     - ⊘ = skipped
     - ✗ = failed
   - `[human]` or `[swarm]` mode indicator per stage
   - Hero name and elapsed time for each stage
   - List of artifacts produced so far

4. **If the workflow is in awaiting_human status**, show:
   - `⏸` indicator on the next pending human-mode stage
   - `← your turn` annotation
   - Resume instruction: "Run /workflow advance to resume."

5. **If no workflow is found**, report "No active workflow found for branch {branch}. Run /workflow start to begin."

## Output Format

### Active Workflow

```
Workflow: wf-feat-health-check-20260320T143000
Branch: feat/health-check
Backlog Item: BI-042
Status: active
Started: 2026-03-20T14:30:00Z
Iterations: 1

Stages:
  ✓ define      (muti-mind)     completed  15m   [human]
  ◉ implement   (cobalt-crush)  active     2h30m [swarm]
  ○ validate    (gaze)          pending          [swarm]
  ○ review      (divisor)       pending          [swarm]
  ○ accept      (muti-mind)     pending          [human]
  ○ reflect     (mx-f)          pending          [swarm]

Artifacts produced:
  .unbound-force/artifacts/BI-042-backlog-item.json
  specs/042-health-check/spec.md
```

### Awaiting Human

```
Workflow: wf-feat-health-check-20260320T143000
Branch: feat/health-check
Status: awaiting_human

Stages:
  ✓ define      (muti-mind)     completed  15m   [human]
  ✓ implement   (cobalt-crush)  completed  2h    [swarm]
  ✓ validate    (gaze)          completed  5m    [swarm]
  ✓ review      (divisor)       completed  20m   [swarm]
  ⏸ accept      (muti-mind)     pending          [human]  ← your turn
  ○ reflect     (mx-f)          pending          [swarm]

Run /workflow advance to resume.
```
