---
name: workflow-advance
description: Advance the current workflow to the next stage
---

# /workflow advance

Advance the current workflow to the next stage.

## Usage

```
/workflow advance [workflow-id]
```

If no workflow-id is provided, advances the most recent active workflow for the current branch.

## Behavior

When this command is invoked:

1. **Find the active workflow**:
   - If a workflow-id is provided, read `.unbound-force/workflows/{workflow-id}.json`
   - Otherwise, detect the current branch and find the most recent active workflow

2. **Validate the current stage can advance**:
   - The current stage must be active (not pending, skipped, or failed)
   - Check if the stage has produced expected artifacts (advisory, not blocking)

3. **Complete the current stage**:
   - Set status to "completed"
   - Record completed_at timestamp
   - Record any artifacts consumed from the previous stage

4. **Find the next non-skipped stage**:
   - Skip stages marked as "skipped" (hero unavailable)
   - If the next stage is "review" and iteration_count >= 3, escalate

5. **Activate the next stage**:
   - Set status to "active"
   - Record started_at timestamp
   - Update current_stage index

6. **If no more stages remain**:
   - Set workflow status to "completed"
   - Generate a workflow-record artifact at `.unbound-force/artifacts/`
   - Report the workflow completion summary

7. **Update the workflow JSON file** with the new state.

8. **Report the new stage** with the next hero action.

## Output Format

### Stage Advanced

```
Stage advanced: implement → validate

Workflow: wf-feat-health-check-20260320T143000
Stage 3/6: validate (Gaze)
  Next: Run Gaze to produce a quality report.
```

### Workflow Completed

```
Workflow completed: wf-feat-health-check-20260320T143000

All stages:
  ✓ define      (muti-mind)     15m
  ✓ implement   (cobalt-crush)  2h30m
  ✓ validate    (gaze)          5m
  ✓ review      (divisor)       20m
  ✓ accept      (muti-mind)     10m
  ✓ measure     (mx-f)          5m

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
