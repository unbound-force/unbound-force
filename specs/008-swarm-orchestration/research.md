# Research: Swarm Orchestration

**Spec**: 008-swarm-orchestration
**Date**: 2026-03-20

## R1: Orchestration Engine Architecture

**Decision**: State machine pattern with `Orchestrator` struct.
The engine manages `WorkflowInstance` objects with explicit
stage transitions. Each transition produces/consumes artifacts
and updates the persistent workflow state.

**Rationale**: A state machine is the natural fit for a
sequential pipeline with conditional transitions (skip on
missing hero, escalate on max iterations, retry on review
rejection). Pure transition functions enable deterministic
testing.

**Alternatives considered**:
- Event-driven (publish/subscribe): More flexible but adds
  complexity. The pipeline is fundamentally sequential.
- Pipeline pattern (chain of responsibility): Simpler but
  doesn't handle branching (skip, retry, escalate).

## R2: Artifact Protocol — Path Convention

**Decision**: Keep the existing artifact path convention
from `internal/artifacts`: `{dir}/{id}-{artifactType}.json`
where `id` includes a timestamp prefix. The spec's proposed
`{artifact_type}/{timestamp}-{hero}.json` pattern would
require breaking the existing Muti-Mind artifact code.

Instead, use `FindArtifacts(dir, type)` for discovery
(already implemented in Spec 007) and add an `ArtifactContext`
struct for branch/commit/backlog-item metadata.

**Rationale**: The existing `WriteArtifact` and `FindArtifacts`
functions work correctly. Changing the path convention would
break Muti-Mind's existing artifact production. The discovery
function abstracts over the naming convention — consumers
don't care about the filename pattern.

**Alternatives considered**:
- Subdirectory-per-type: `{dir}/{type}/{timestamp}-{hero}.json`.
  Cleaner organization but requires migrating existing artifacts
  and changing all callers.

## R3: Artifact Context Population

**Decision**: Add an `ArtifactContext` struct and update
`WriteArtifact` to accept it:

```go
type ArtifactContext struct {
    Branch        string `json:"branch,omitempty"`
    Commit        string `json:"commit,omitempty"`
    BacklogItemID string `json:"backlog_item_id,omitempty"`
    CorrelationID string `json:"correlation_id,omitempty"`
    WorkflowID    string `json:"workflow_id,omitempty"`
}
```

The `Envelope.Context` field (already `json.RawMessage`) will
be populated with this struct. This enables workflow isolation
(FR-010) — artifacts include their branch context.

**Rationale**: The Context field already exists in the Envelope
but is never populated. Adding a typed struct makes it
type-safe for producers and consumers. The `WorkflowID` field
links artifacts to their workflow instance.

## R4: Hero Availability Detection

**Decision**: Detect hero availability by checking for agent
files in `.opencode/agents/`:

| Hero | Agent File | Detection |
|------|-----------|-----------|
| Muti-Mind | `muti-mind-po.md` | file exists |
| Cobalt-Crush | `cobalt-crush-dev.md` | file exists |
| Gaze | (external binary) | `which gaze` succeeds |
| The Divisor | `divisor-guard.md` (any `divisor-*.md`) | file exists |
| Mx F | `mx-f-coach.md` + `which mxf` | agent + binary |

**Rationale**: Agent files are the canonical deployment
indicator. If a hero's agent file isn't present, it's not
deployed. Gaze is special — it's an external binary, not
an embedded agent. The orchestrator checks both.

## R5: Swarm Skills Package Format

**Decision**: Create a project-level Swarm skill at
`.opencode/skill/unbound-force-heroes/SKILL.md`. The skill
file is Markdown with YAML frontmatter:

```yaml
---
name: unbound-force-heroes
description: "Unbound Force hero roles and workflow routing"
tags:
  - heroes
  - workflow
  - routing
---
```

The body describes each hero's role, routing patterns, and
the workflow stage sequence. Loaded by Swarm via
`skills_use({ name: "unbound-force-heroes" })`.

**Rationale**: This is Swarm's documented convention for
project-level skills. The skill teaches the Swarm coordinator
how to route natural language queries to the right hero.

## R6: OpenCode Command Pattern

**Decision**: Three Markdown command files that invoke the
Go orchestration package via bash (calling a helper script
or `go run` during development). Commands:

- `/workflow start [backlog-item-id]` — creates a new
  workflow, detects available heroes, begins first stage
- `/workflow status` — shows current workflow state
- `/workflow list` — lists all workflows with status

The commands read workflow state from
`.unbound-force/workflows/` and display it.

**Rationale**: OpenCode commands are Markdown files that
provide instructions to the LLM. For v1.0.0, the commands
instruct the LLM to read/write workflow JSON files directly.
When a CLI is added later, the commands can delegate to it.

## R7: Workflow Record Artifact

**Decision**: On workflow completion (all stages done or
escalated), produce a `workflow-record` artifact via
`WriteArtifact`. The record captures:

- All stages with timing and status
- All artifacts produced (paths)
- All decisions made (acceptance, review verdicts)
- Total elapsed time
- Outcome (shipped/rejected/abandoned)

**Rationale**: FR-013 and FR-014 require this. Mx F consumes
it for lifecycle metrics. Muti-Mind consumes it for velocity.

## R8: Learning Feedback Model

**Decision**: `LearningFeedback` struct stored alongside
workflow records. Produced by analyzing completed workflow
records for patterns:

- Frequent Divisor findings → convention pack recommendation
- Declining quality trends → testing strategy recommendation
- Velocity patterns → prioritization input

Learning feedback is a proposal, not a mandate (edge case
in spec). Status tracks: proposed → accepted/rejected →
implemented.

**Rationale**: FR-007 requires the feedback loop. The
`coaching-record` artifact from Mx F is the existing
mechanism. The orchestration layer adds cross-hero analysis.

## R9: Embedded Asset Count

**Decision**: No new embedded assets. The `/workflow` commands
and Swarm skills are local-only tooling, not scaffolded by
`unbound init`. Add them to `knownNonEmbeddedFiles`.

**Rationale**: The workflow commands and skills are specific
to this project's development workflow, not something every
`unbound init` target needs. Same treatment as `/cobalt-crush`,
`/gaze`, and the Muti-Mind commands.
<!-- scaffolded by unbound vdev -->
