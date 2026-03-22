# Research: Swarm Delegation Workflow

**Date**: 2026-03-22
**Branch**: `012-swarm-delegation`

## R1: Checkpoint State Machine Design

**Decision**: Add `StatusAwaitingHuman` as a workflow-level
status and use the existing `Advance()` method for both
pausing and resuming.

**Rationale**: The orchestration engine's `Advance()` is
the sole transition API called by both the Swarm coordinator
and human operators. Rather than adding a separate `Resume()`
method, `Advance()` can detect the `awaiting_human` status
and handle it as a resume: activate the pending human-mode
stage and set status back to `active`. This keeps the API
surface minimal and the Swarm coordinator's integration
unchanged -- it simply calls `Advance()` and observes the
resulting status.

**Alternatives considered**:
- Separate `Resume()` method: Rejected because it doubles
  the API surface for no functional benefit. The Swarm
  coordinator would need to learn a new command.
- `Pause()` method: Rejected because pausing is not a user
  action -- it is an automatic consequence of reaching a
  mode boundary. Adding an explicit pause call would require
  the Swarm to know about mode boundaries, violating the
  engine's responsibility to manage transitions.

## R2: Backward Compatibility for ExecutionMode

**Decision**: Use `omitempty` on the `ExecutionMode` JSON
tag. When the field is empty (legacy JSON), treat it as
`"human"` in all transition logic.

**Rationale**: Existing workflow JSON files at
`.unbound-force/workflows/` do not have an `execution_mode`
field on stages. Go's `json.Unmarshal` leaves the field as
its zero value (empty string) for missing keys. By treating
empty as `"human"`, legacy workflows advance one-stage-at-a-
time exactly as before -- no checkpoint logic fires because
every stage-to-stage transition is human-to-human.

**Alternatives considered**:
- Migration script to backfill `execution_mode`: Rejected
  because it adds operational complexity for no behavioral
  change. Legacy workflows already behave as all-human.
- Default to `"swarm"`: Rejected because it would break
  existing workflows by pausing at unexpected points.

## R3: Latest() Discovery of Awaiting Workflows

**Decision**: Update `WorkflowStore.Latest()` to return
workflows in either `StatusActive` or `StatusAwaitingHuman`
status.

**Rationale**: The `Latest()` method is used by
`/workflow advance` and `/workflow status` to find the
current branch's workflow without requiring an explicit
workflow ID. A workflow in `awaiting_human` status is still
"in progress" from the operator's perspective -- they need
to find it to resume it. If `Latest()` only returns
`StatusActive`, an `awaiting_human` workflow would be
invisible and the operator would have to provide the
workflow ID manually.

**Alternatives considered**:
- Separate `LatestAwaitingOrActive()` method: Rejected
  because `Latest()` is the canonical discovery method
  and splitting it forces callers to decide which to use.
- Filter parameter on `Latest()`: Rejected because it
  adds complexity to every call site. The intent is
  always "find the in-progress workflow for this branch."

## R4: Stage Rename Impact Analysis

**Decision**: Rename `StageMeasure` constant and `"measure"`
string to `StageReflect` / `"reflect"`. Update all
references across Go code and Markdown files.

**Rationale**: The reflect stage is a semantic enrichment:
it captures what measure did (metrics collection) plus
learning analysis and retrospective summary. The rename
communicates that the stage is about team reflection, not
just data capture.

**Impact scan** (all occurrences of `StageMeasure` or
`"measure"` as a stage name):

Go files:
- `models.go`: constant definition + `StageOrder()` (2)
- `heroes.go`: `heroSpecs` + `StageHeroMap()` (2)
- `engine_test.go`: 3 assertions
- `heroes_test.go`: 1 assertion

Markdown files:
- `.opencode/skill/unbound-force-heroes/SKILL.md` (1)
- `.opencode/command/workflow-start.md` (3)
- `.opencode/command/workflow-status.md` (1)
- `.opencode/command/workflow-advance.md` (1)

Not changed (completed specs, not living docs):
- `specs/008-swarm-orchestration/*` -- completed spec,
  documents the original design. Not updated.
- `AGENTS.md` -- Recent Changes entry updated to reflect
  the new stage name.

**Alternatives considered**:
- Keep "measure" and just enrich the stage: Rejected
  because "measure" implies passive data collection,
  which misrepresents the retrospective nature of the
  enriched stage.
- Rename to "learn": Rejected because "reflect" better
  captures the retrospective facilitation aspect (Mx F's
  5-phase protocol) versus pure pattern extraction.

## R5: Swarm Coordinator Integration

**Decision**: The orchestration engine communicates the
checkpoint state through the persisted workflow JSON.
No new protocol or signaling mechanism is needed.

**Rationale**: The Swarm coordinator already reads
workflow state via `/workflow status` and advances via
`/workflow advance`. When `Advance()` returns a workflow
with `status: "awaiting_human"`, the Swarm coordinator
simply stops advancing and reports back to the human.
The SKILL.md update teaches the coordinator about this
new status value.

**Alternatives considered**:
- Webhook/callback when checkpoint is reached: Rejected
  because it introduces runtime coupling, violating
  Principle I (Autonomous Collaboration).
- Swarm-specific API for mode queries: Rejected because
  the workflow JSON already contains all needed state.

## R6: FR-014 All-Skipped Swarm Stages

**Decision**: When all swarm-mode stages between two human
checkpoints are skipped (heroes unavailable), the workflow
transitions directly to the next human-mode stage without
entering `awaiting_human`.

**Rationale**: The checkpoint exists to give the swarm time
to complete autonomous work. If no autonomous work will
happen (all swarm heroes unavailable), pausing serves no
purpose. The `Advance()` loop already skips unavailable
stages; this behavior falls out naturally because the
forward scan skips `StatusSkipped` stages and only triggers
the checkpoint logic when it finds a pending human-mode
stage after a completed swarm-mode stage. If the swarm-mode
stages are all skipped, the "last completed stage" is the
human-mode define stage, and advancing from human-to-human
does not trigger a checkpoint.

**Alternatives considered**:
- Always pause at human-mode boundaries regardless: Rejected
  because pausing when nothing happened is confusing to
  the operator.

## R7: Reflect Stage Artifact Consumption

**Decision**: The reflect stage's enriched behavior
(consuming quality report and review verdict, running
learning analysis) is documented in the SKILL.md for the
Swarm coordinator to follow. The Go engine tracks artifact
metadata but does not execute the analysis itself.

**Rationale**: The orchestration engine is a state machine,
not an execution engine. It records which artifacts were
consumed/produced per stage via the `ArtifactsConsumed`
and `ArtifactsProduced` fields. The actual analysis
(collecting metrics, running `AnalyzeWorkflows`,
generating a retrospective summary) is performed by the
Swarm coordinator following the SKILL.md instructions,
using Mx F's agent and the existing coaching/learning
infrastructure.

**Alternatives considered**:
- Have the engine call `AnalyzeWorkflows` directly in
  `Advance()` when completing the reflect stage: Rejected
  because it couples the engine to specific hero logic,
  violating Principle I. The engine manages transitions;
  heroes do work.
