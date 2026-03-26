# Contract: Workflow Seed Command

**Date**: 2026-03-26
**Branch**: `016-autonomous-define`

## `/workflow seed` Command

### Usage

```
/workflow seed <description>
```

### Behavior

1. Creates a backlog item from `<description>`
2. Starts a workflow with `define=swarm` mode override
3. The define stage executes autonomously (Muti-Mind
   drafts the spec using Dewey context)
4. Returns the workflow ID

### Output Format

```
Seeded: wf-feat-csv-export-20260326T143000

Muti-Mind is drafting the specification...

Workflow stages:
  ◉ define      (muti-mind)     active   [swarm]
  ○ implement   (cobalt-crush)  pending  [swarm]
  ○ validate    (gaze)          pending  [swarm]
  ○ review      (divisor)       pending  [swarm]
  ○ accept      (muti-mind)     pending  [human]
  ○ reflect     (mx-f)          pending  [swarm]

The swarm will notify you when the increment is
ready for acceptance.
```

### Error: Empty Description

```
> /workflow seed

Please provide a feature description:
> add CSV export to the dashboard

Seeded: wf-feat-csv-export-20260326T143000
...
```

## `/workflow start` Updates

### New Flag: `--define-mode`

```
/workflow start BI-042 --define-mode=swarm
```

Sets the define stage execution mode. Accepts `human`
(default) or `swarm`.

### New Flag: `--spec-review`

```
/workflow start BI-042 --define-mode=swarm --spec-review
```

Enables the spec review checkpoint between define and
implement. When enabled, the workflow pauses after
Muti-Mind completes the spec.

### Output with Spec Review Checkpoint

When the spec review checkpoint triggers:

```
Specification drafted: wf-feat-csv-export-20260326T143000

Muti-Mind has drafted the specification.
Review at: specs/NNN-feature-name/spec.md

Awaiting spec review. Run /workflow advance to approve.
```

After the human advances:

```
Spec approved. Continuing to implement...

Stage 2/6: implement (Cobalt-Crush) [swarm]
```

## Execution Mode Override API

### NewWorkflow Changes

```go
// NewWorkflow accepts optional execution mode overrides.
// Unspecified stages keep their defaults from
// StageExecutionModeMap(). Invalid override keys (stage
// names not in StageOrder()) or values (not "human" or
// "swarm") return an error.
func (o *Orchestrator) NewWorkflow(
    branch string,                // existing param
    backlogItemID string,         // existing param
    overrides map[string]string,  // NEW: stage→mode overrides
    specReview bool,              // NEW: enable spec review checkpoint
) (*WorkflowInstance, error)      // NEW: returns error for validation
```

Note: `Start()` must also be updated to accept and
forward `overrides` and `specReview` to `NewWorkflow()`.
All 20+ `Start()` call sites in tests pass `nil, false`
for backward compatibility.

### WorkflowInstance Changes (Live State)

```json
{
  "id": "wf-feat-csv-export-20260326T143000",
  "status": "active",
  "spec_review_enabled": true,
  "stages": [
    {
      "stage_name": "define",
      "hero": "muti-mind",
      "execution_mode": "swarm",
      ...
    }
  ]
}
```

New field: `spec_review_enabled` (boolean, defaults to
`false`). Persisted in the workflow **instance** JSON
(at `.unbound-force/workflows/{id}.json`) for
checkpoint logic.

**Note**: This field is on `WorkflowInstance` (live
state), NOT on `WorkflowRecord` (completed workflow
artifact). The `WorkflowRecord` struct and its schema
at `schemas/workflow-record/` are not modified. The
spec review flag is a runtime configuration, not a
permanent artifact attribute.
