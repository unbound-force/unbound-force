# Research: Autonomous Define with Dewey

**Date**: 2026-03-26
**Branch**: `016-autonomous-define`

## R1: Configuration Mechanism for Execution Modes

**Decision**: Accept an optional `ExecutionModeOverrides`
map in `NewWorkflow()` that merges with the defaults
from `StageExecutionModeMap()`.

**Rationale**: The current `StageExecutionModeMap()`
returns hardcoded defaults. Rather than replacing it
with a config file, the simplest approach is to accept
overrides at workflow creation time. The caller
(Swarm coordinator or `/workflow start` command) passes
a map of stage→mode overrides. Unspecified stages keep
their defaults.

This approach:
1. Requires no config file changes
2. Is backward compatible (empty overrides = defaults)
3. Is testable (pass overrides in test setup)
4. Allows per-workflow customization (different
   features can have different modes)

**Alternatives considered**:
- Global config file (`.unbound-force/config.yaml`):
  Rejected because per-workflow customization is more
  flexible and avoids shared mutable config state.
- CLI flag on `/workflow start`: Accepted as the
  mechanism for passing overrides from the operator.
  The flag value populates the overrides map.

## R2: Spec Review Checkpoint Implementation

**Decision**: Implement the spec review checkpoint as
a special case in `Advance()` -- when the define stage
completes and the checkpoint is enabled, transition to
`StatusAwaitingHuman` before activating the implement
stage.

**Rationale**: The existing checkpoint logic in
`Advance()` already handles swarm-to-human mode
transitions (Spec 012). The spec review checkpoint is
conceptually identical -- it's a swarm-to-human
transition between define and implement. The difference
is that it's optional and triggered by a flag, not by
the stage's execution mode.

Implementation: Add a `SpecReviewEnabled` boolean to
the workflow record. When true, `Advance()` checks
after completing define: if the next stage (implement)
is in swarm mode, pause with `StatusAwaitingHuman`
instead of proceeding. This is the same behavior as a
mode boundary checkpoint, just triggered differently.

**Alternatives considered**:
- New workflow status (`StatusAwaitingSpecReview`):
  Rejected per spec constraint -- no new status
  constants. Reusing `StatusAwaitingHuman` is cleaner.
- Insert a virtual "review" stage between define and
  implement: Rejected because it changes the 6-stage
  count and breaks stage ordering assumptions.

## R3: Seed Command Design

**Decision**: Implement the seed command as a
`/workflow seed` OpenCode command that combines backlog
item creation with workflow start in one operation.

**Rationale**: The seed command is a convenience
wrapper. Under the hood, it:
1. Creates a backlog item from the seed description
   (using Muti-Mind's `mutimind add --title "<seed>"
   --type story` CLI)
2. Starts a workflow with `define=swarm` mode override
3. Returns the workflow ID

The command does not replace `/workflow start` -- it
coexists as a shortcut for the autonomous define
workflow.

**Alternatives considered**:
- Extend `/workflow start` with a `--seed` flag:
  Acceptable but adds complexity to an existing command.
  A separate command is cleaner for discoverability.
- Standalone CLI command (`uf seed`): Rejected because
  the seed operation is a workflow concern, not a
  standalone tool operation.

## R4: Muti-Mind Autonomous Spec Workflow

**Decision**: Add an "Autonomous Specification" section
to the Muti-Mind agent file describing the step-by-step
workflow for autonomous spec drafting.

**Rationale**: The autonomous spec workflow is an AI
agent instruction, not Go code. Muti-Mind's agent file
(`.opencode/agents/muti-mind-po.md`) already includes
Dewey tool usage and 3-tier degradation. Adding an
"Autonomous Specification" section teaches Muti-Mind
how to:

1. Accept a seed description as input
2. Query Dewey for related specs, issues, and docs
3. Draft a specification using the speckit template
4. Self-clarify by querying Dewey for ambiguities
5. Reference learning feedback from past workflows
6. Produce the spec as an artifact

This is not Go implementation -- it's prompt
engineering in the agent file. The Swarm coordinator
invokes Muti-Mind with the seed description, and
Muti-Mind follows the instructions in its agent file.

**Alternatives considered**:
- Implement spec drafting in Go code: Rejected because
  specification drafting requires LLM reasoning, not
  deterministic code. The agent file is the correct
  layer.
- Use a separate "autonomous-spec" agent: Rejected
  because Muti-Mind is the Product Owner -- spec
  drafting is its core responsibility.

## R5: Backward Compatibility Strategy

**Decision**: The default `StageExecutionModeMap()`
returns `define=human` (unchanged). The overrides map
is empty by default. `SpecReviewEnabled` defaults to
`false`. All existing behavior is preserved.

**Rationale**: The spec explicitly requires backward
compatibility (FR-002, SC-005, SC-006). The override
mechanism ensures that callers who don't pass overrides
get exactly the same workflow as before. Existing tests
don't need modification -- they use the default mode
map.

**Alternatives considered**:
- Make `define=swarm` the default: Rejected per FR-002.
  The default must remain `human` for backward
  compatibility.
