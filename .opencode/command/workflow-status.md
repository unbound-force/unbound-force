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

If no workflow-id is provided, shows the most recent active workflow for the current branch.

## Behavior

When this command is invoked:

1. **If a workflow-id is provided**, read `.unbound-force/workflows/{workflow-id}.json` directly.

2. **If no workflow-id is provided**:
   - Detect the current git branch via `git branch --show-current`
   - Read all JSON files from `.unbound-force/workflows/`
   - Find the most recent active workflow matching the current branch

3. **Parse the workflow JSON** and display:
   - Workflow ID, branch, backlog item, status, start time
   - Iteration count (for review loops)
   - Each stage with status indicator:
     - ✓ = completed
     - ◉ = active
     - ○ = pending
     - ⊘ = skipped
     - ✗ = failed
   - Hero name and elapsed time for each stage
   - List of artifacts produced so far

4. **If no workflow is found**, report "No active workflow found for branch {branch}. Run /workflow start to begin."

## Output Format

```
Workflow: wf-feat-health-check-20260320T143000
Branch: feat/health-check
Backlog Item: BI-042
Status: active
Started: 2026-03-20T14:30:00Z
Iterations: 1

Stages:
  ✓ define      (muti-mind)     completed  15m
  ◉ implement   (cobalt-crush)  active     2h30m
  ○ validate    (gaze)          pending
  ○ review      (divisor)       pending
  ○ accept      (muti-mind)     pending
  ○ measure     (mx-f)          pending

Artifacts produced:
  .unbound-force/artifacts/BI-042-backlog-item.json
  specs/042-health-check/spec.md
```
