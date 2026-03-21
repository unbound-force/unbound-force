# Contract: Workflow Commands

**Spec**: 008-swarm-orchestration
**Date**: 2026-03-20
**Type**: OpenCode command schema

## Command: `/workflow start`

Begin a new hero lifecycle workflow.

### Usage

```
/workflow start [backlog-item-id]
```

### Behavior

1. Detect available heroes by checking agent files in
   `.opencode/agents/` and binaries in PATH.
2. Detect current git branch for workflow scoping.
3. Create a new `WorkflowInstance` with 6 stages
   (define, implement, validate, review, accept, measure).
4. Mark unavailable hero stages as `skipped`.
5. Write workflow state to
   `.unbound-force/workflows/{workflow_id}.json`.
6. Begin the first non-skipped stage.
7. Report available heroes, skipped stages, and next action.

### Output

```text
Workflow started: wf-feat-health-check-20260320T143000

Available heroes:
  ✓ Muti-Mind (define)
  ✓ Cobalt-Crush (implement)
  ✗ Gaze (validate) — not installed
  ✓ The Divisor (review)
  ✓ Muti-Mind (accept)
  ✓ Mx F (measure)

Stage 1/6: define (Muti-Mind)
  Next: Run /speckit.specify to create the feature spec.
```

## Command: `/workflow status`

Check the current workflow state.

### Usage

```
/workflow status [workflow-id]
```

If no workflow-id is provided, shows the most recent active
workflow for the current branch.

### Output

```text
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

## Command: `/workflow list`

List all workflows.

### Usage

```
/workflow list [--status active|completed|all]
```

### Output

```text
Workflows:
  ID                                    Branch              Status     Started
  wf-feat-health-check-20260320T143000  feat/health-check   active     2h ago
  wf-feat-login-20260319T090000         feat/login          completed  1d ago
  wf-feat-signup-20260318T100000        feat/signup         escalated  2d ago
```

## Go API: `orchestration.Orchestrator`

```go
type Orchestrator struct {
    WorkflowDir  string            // .unbound-force/workflows/
    ArtifactDir  string            // .unbound-force/artifacts/
    AgentDir     string            // .opencode/agents/
    GHRunner     sync.GHRunner     // GitHub CLI interface
    Now          func() time.Time  // Clock injection
    Stdout       io.Writer         // Output writer
}

type WorkflowResult struct {
    Workflow     *WorkflowInstance
    StagesRun   int
    Warnings    []string
}
```

## Key Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `Start` | `(branch, backlogItemID string) (*WorkflowResult, error)` | Create workflow, detect heroes, begin first stage |
| `Advance` | `(workflowID string) (*WorkflowResult, error)` | Move to next stage |
| `Skip` | `(workflowID string, stage int, reason string) error` | Skip a stage |
| `Escalate` | `(workflowID, reason string) error` | Escalate workflow to human |
| `Complete` | `(workflowID string) (*WorkflowRecord, error)` | Finalize workflow, produce record |
| `Status` | `(workflowID string) (*WorkflowInstance, error)` | Get current state |
| `List` | `(statusFilter string) ([]WorkflowInstance, error)` | List workflows |
| `DetectHeroes` | `() ([]HeroStatus, error)` | Check which heroes are available |
<!-- scaffolded by unbound vdev -->
