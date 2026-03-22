---
name: workflow-advance
description: Advance the current workflow to the next stage
---

# /workflow advance

Advance the current workflow to the next stage, or resume from a human checkpoint.

## Usage

```
/workflow advance [workflow-id]
```

If no workflow-id is provided, advances the most recent in-progress workflow (active or awaiting_human) for the current branch.

## Behavior

When this command is invoked:

1. **Find the in-progress workflow**:
   - If a workflow-id is provided, read `.unbound-force/workflows/{workflow-id}.json`
   - Otherwise, detect the current branch and find the most recent active or awaiting_human workflow

2. **If the workflow is in awaiting_human status** (resuming from checkpoint):
   - Find the next pending human-mode stage
   - Activate it and set workflow status back to "active"
   - Report the resumed stage

3. **If the workflow is active** (normal advance):
   - Validate the current stage can advance (must be active)
   - Complete the current stage (set status to "completed", record timestamps and artifacts)

4. **Find the next non-skipped stage**:
   - Skip stages marked as "skipped" (hero unavailable)
   - If the next stage is "review" and iteration_count >= 3, escalate

5. **Check for human checkpoint**:
   - If the completed stage is swarm-mode and the next stage is human-mode, pause the workflow
   - Set workflow status to "awaiting_human"
   - Report the checkpoint

6. **If no checkpoint**, activate the next stage:
   - Set status to "active"
   - Record started_at timestamp
   - Update current_stage index

7. **If no more stages remain**:
   - Set workflow status to "completed"
   - Generate a workflow-record artifact at `.unbound-force/artifacts/`
   - Report the workflow completion summary

8. **Update the workflow JSON file** with the new state.

9. **Report the result** with the next hero action.

## Output Format

### Stage Advanced

```
Stage advanced: implement → validate

Workflow: wf-feat-health-check-20260320T143000
Stage 3/6: validate (Gaze) [swarm]
  Next: Run Gaze to produce a quality report.
```

### Checkpoint Reached (swarm → human boundary)

```
Workflow paused: wf-feat-health-check-20260320T143000

Swarm completed: implement → validate → review
Awaiting human action at: accept (Muti-Mind) [human]

Run /workflow advance to resume and accept the increment.
```

### Resumed from Checkpoint

```
Workflow resumed: wf-feat-health-check-20260320T143000

Stage 5/6: accept (Muti-Mind) [human]
  Next: Review the increment and make an acceptance decision.
```

### Workflow Completed

```
Workflow completed: wf-feat-health-check-20260320T143000

All stages:
  ✓ define      (muti-mind)     15m   [human]
  ✓ implement   (cobalt-crush)  2h30m [swarm]
  ✓ validate    (gaze)          5m    [swarm]
  ✓ review      (divisor)       20m   [swarm]
  ✓ accept      (muti-mind)     10m   [human]
  ✓ reflect     (mx-f)          5m    [swarm]

Total time: 3h25m
Outcome: shipped
Workflow record: .unbound-force/artifacts/wf-feat-health-check-20260320T143000-workflow-record.json
```

### Escalation

```
Workflow escalated: wf-feat-health-check-20260320T143000
Reason: Maximum review iterations (3) reached.

Unresolved findings require manual review.
```
